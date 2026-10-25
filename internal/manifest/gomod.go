// Package manifest reads dependency manifests and lock files and names the
// packages a change added or raised. Go (go.mod and go.sum), npm
// (package.json and package-lock.json), PyPI (requirements.txt and
// poetry.lock), Packagist (composer.json and composer.lock), Rust
// (Cargo.toml and Cargo.lock) and Kotlin (libs.versions.toml,
// build.gradle.kts and gradle.lockfile) each have a manifest reader and a
// lock reader. Kotlin has two manifest readers, one per file. Each reader
// is an Adapter: the
// files, the diff, the OSV ecosystem name, and the version comparison
// (ADR-0038).
//
// It feeds the dependency check of the review, which runs with no model:
// osv.Check calls Collect, queries the OSV table once per package it gets
// back, and hands a matching range to the Judge as query evidence
// (ADR-0002, ADR-0004). A reader never decides that a version is affected.
// It only says what changed and on which line, so the finding can point
// at that line. The format each reader accepts is written down in
// docs/OSV_TABLE.md.
package manifest

import (
	"bufio"
	"bytes"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/pep503"
	"github.com/saphan-ai/proven/internal/semver"
)

// Change is one module the change added or raised.
// Package is the module path or the package name. Go and npm keep the
// spelling in the file. A PyPI name is the PEP 503 form, which is the name
// the comparison, the lock pairing and the OSV query share.
// Version is the version on the change side, as written. It is not
// rewritten, so the trace shows what the author wrote.
// Line is the 1-based line on the change side that the finding points at.
// Added is true when the base did not name the package at all, and false
// when the base named it and this version is a raise.
type Change struct {
	Package string
	Version string
	Line    int
	Added   bool
}

// Hit is a Change plus the manifest it came from.
// Ecosystem is the OSV ecosystem name the adapter declares. The query
// filters the table's ecosystem column on it. File is the slash path of
// the manifest under the change root.
// Lock is set when the hit came from a lock file.
type Hit struct {
	Ecosystem string
	File      string
	Lock      bool
	Change
}

// Reader reads one manifest kind. It is the adapter of one ecosystem
// (ADR-0038): which files, how a base and a change differ, the OSV
// ecosystem name, and how two versions compare. Adapter is the same type.
// Notes, when set, names lines of the change side that were not dependencies.
// Each string is one sentence the OSV check writes to the trace.
// Lock is set for a lock file. preferLock then keeps that hit and drops the
// manifest hit for the same package in the same directory.
// NoteErr is set when a Diff error is one trace note and not a failed check.
// A lock file that cannot be parsed then contributes no finding.
// Match reports whether a slash path under the change root is a file of
// this kind; it matches the exact file name, never a suffix like .bak.
// Diff names what the change side adds or raises over the base side. A
// file that exists on one side only reaches it as empty bytes on the other.
// Compare orders two versions by this ecosystem's rules. The bool is false
// when either string is not a version this adapter compares. It is the same
// order Diff uses, so a raise and an OSV range cannot disagree.
// Explain says why a value is not such a version. kind is the bound name
// (introduced, fixed, last_affected) or empty for the package version.
// The bool is true when the value parses. A nil Explain means this
// ecosystem is not compared.
type Reader struct {
	Ecosystem string
	Match     func(path string) bool
	Diff      func(base, change []byte) ([]Change, error)
	Compare   func(a, b string) (int, bool)
	Explain   func(kind, value string) (string, bool)
	Notes     func(change []byte) []string
	Lock      bool
	NoteErr   bool
}

// Go reads go.mod require lines. A replace or an exclude is not a require.
func Go() Reader {
	return Reader{
		Ecosystem: "Go",
		Match: func(path string) bool {
			return path == "go.mod" || strings.HasSuffix(path, "/go.mod")
		},
		Diff:    diffGo,
		Compare: semver.Compare,
		Explain: semver.Explain,
	}
}

// Collect runs each reader on the files it matches.
// The strings are the readers' skip notes, each prefixed with the file path.
func Collect(files []change.File, readers []Reader) ([]Hit, []string, error) {
	var out []Hit
	var notes []string
	for _, f := range files {
		for _, r := range readers {
			// One file can match more than one reader, so every reader is tried.
			if r.Match == nil || r.Diff == nil || !r.Match(f.Path) {
				continue
			}
			got, err := r.Diff(f.Base, f.Change)
			if err != nil {
				// A lock file is often generated and large. One that this reader cannot
				// parse is a trace note, not a failed review: the manifest beside it is
				// still checked.
				if r.NoteErr {
					notes = append(notes, f.Path+": "+err.Error())
					continue
				}
				// A manifest the author edits by hand must parse. Its error stops
				// the check, so a broken file is never read as "no dependency".
				return nil, nil, err
			}
			for _, c := range got {
				out = append(out, Hit{Ecosystem: r.Ecosystem, File: f.Path, Lock: r.Lock, Change: c})
			}
			if r.Notes != nil {
				// Notes come from the change side only: they explain why a line
				// the author wrote was not checked.
				for _, n := range r.Notes(f.Change) {
					notes = append(notes, f.Path+": "+n)
				}
			}
		}
	}
	// A fixed order, so the queries, the trace and the review come out the
	// same on every run of the same change.
	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		if out[i].Package != out[j].Package {
			return out[i].Package < out[j].Package
		}
		return out[i].Version < out[j].Version
	})
	// The lock file wins only after every file is read, because a manifest
	// and its lock file can come in either order.
	return preferLock(out), notes, nil
}

// preferLock drops a manifest hit when a lock file in the same directory
// already names that package. The lock hit carries the exact version.
// PyPI names match after PEP 503 normalisation. Go, npm and crate names do not.
func preferLock(hits []Hit) []Hit {
	type key struct {
		dir string
		eco string
		pkg string
	}
	locked := map[key]struct{}{}
	for _, h := range hits {
		if !h.Lock {
			continue
		}
		locked[key{dirOf(h.File), h.Ecosystem, foldName(h.Ecosystem, h.Package)}] = struct{}{}
	}
	if len(locked) == 0 {
		return hits
	}
	out := make([]Hit, 0, len(hits))
	for _, h := range hits {
		if !h.Lock {
			k := key{dirOf(h.File), h.Ecosystem, foldName(h.Ecosystem, h.Package)}
			if _, ok := locked[k]; ok {
				continue
			}
		}
		out = append(out, h)
	}
	return out
}

// dirOf returns the directory part of a slash path, or "" at the root.
// A manifest and a lock file pair up only inside one directory.
func dirOf(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[:i]
	}
	return ""
}

// foldName gives the name preferLock compares. A PyPI name is the PEP 503
// form, so Flask and flask are one package, and zope_interface and
// zope.interface are one package. Go module paths, npm names and crate names
// are compared as written: case folding would merge modules that are not the same.
func foldName(ecosystem, pkg string) string {
	if ecosystem == "PyPI" {
		return pep503.Normalise(pkg)
	}
	return pkg
}

// req is one named requirement in a manifest: its version and its line.
type req struct {
	version string
	line    int
}

// diffGo compares the require lines of two go.mod files.
func diffGo(base, change []byte) ([]Change, error) {
	before, err := parseRequires(base)
	if err != nil {
		return nil, err
	}
	after, err := parseRequires(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, semver.Compare), nil
}

// parseRequires reads require directives. The last line wins when a module
// is required twice. A replace block is skipped: it does not change the
// required version this reader reports.
func parseRequires(b []byte) (map[string]req, error) {
	out := map[string]req{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	inRequire := false
	skipping := false
	for sc.Scan() {
		line++
		text := strings.TrimSpace(stripComment(sc.Text()))
		if text == "" {
			continue
		}
		// Inside a replace, exclude or retract block every line is skipped
		// up to its closing parenthesis (docs/OSV_TABLE.md: only require
		// lines are read).
		if skipping {
			if text == ")" {
				skipping = false
			}
			continue
		}
		if inRequire {
			if text == ")" {
				inRequire = false
				continue
			}
			path, ver, ok := moduleVersion(text)
			if ok {
				out[path] = req{version: ver, line: line}
			}
			continue
		}
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "require":
			// "require (" opens a block; "require path version" is one line.
			if len(fields) == 2 && fields[1] == "(" {
				inRequire = true
				continue
			}
			if len(fields) >= 3 {
				out[fields[1]] = req{version: fields[2], line: line}
			}
		case "replace", "exclude", "retract":
			if len(fields) >= 2 && fields[1] == "(" {
				skipping = true
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// stripComment drops a // comment, so "// indirect" is not read as a field.
func stripComment(s string) string {
	if i := strings.Index(s, "//"); i >= 0 {
		return s[:i]
	}
	return s
}

// moduleVersion splits one line of a require block into path and version.
func moduleVersion(text string) (string, string, bool) {
	fields := strings.Fields(text)
	if len(fields) < 2 || fields[0] == "(" || fields[0] == ")" {
		return "", "", false
	}
	return fields[0], fields[1], true
}
