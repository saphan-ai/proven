// Package manifest reads dependency manifests and names what a change added
// or raised. Go is the first reader. Another ecosystem is another Reader.
package manifest

import (
	"bufio"
	"bytes"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/semver"
)

// Change is one module the change added or raised.
type Change struct {
	Package string
	Version string
	Line    int
	Added   bool
}

// Hit is a Change plus the manifest it came from.
type Hit struct {
	Ecosystem string
	File      string
	Change
}

// Reader reads one manifest kind.
type Reader struct {
	Ecosystem string
	Match     func(path string) bool
	Diff      func(base, change []byte) ([]Change, error)
}

// Go reads go.mod require lines. A replace or an exclude is not a require.
func Go() Reader {
	return Reader{
		Ecosystem: "Go",
		Match: func(path string) bool {
			return path == "go.mod" || strings.HasSuffix(path, "/go.mod")
		},
		Diff: diffGo,
	}
}

// Collect runs each reader on the files it matches.
func Collect(files []change.File, readers []Reader) ([]Hit, error) {
	var out []Hit
	for _, f := range files {
		for _, r := range readers {
			if r.Match == nil || r.Diff == nil || !r.Match(f.Path) {
				continue
			}
			got, err := r.Diff(f.Base, f.Change)
			if err != nil {
				return nil, err
			}
			for _, c := range got {
				out = append(out, Hit{Ecosystem: r.Ecosystem, File: f.Path, Change: c})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Package < out[j].Package
	})
	return out, nil
}

type req struct {
	version string
	line    int
}

func diffGo(base, change []byte) ([]Change, error) {
	before, err := parseRequires(base)
	if err != nil {
		return nil, err
	}
	after, err := parseRequires(change)
	if err != nil {
		return nil, err
	}
	var out []Change
	for path, now := range after {
		old, ok := before[path]
		if !ok {
			out = append(out, Change{Package: path, Version: now.version, Line: now.line, Added: true})
			continue
		}
		if old.version == now.version {
			continue
		}
		cmp, ordered := semver.Compare(now.version, old.version)
		if ordered && cmp <= 0 {
			continue
		}
		out = append(out, Change{Package: path, Version: now.version, Line: now.line, Added: false})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Package < out[j].Package })
	return out, nil
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

func stripComment(s string) string {
	if i := strings.Index(s, "//"); i >= 0 {
		return s[:i]
	}
	return s
}

func moduleVersion(text string) (string, string, bool) {
	fields := strings.Fields(text)
	if len(fields) < 2 || fields[0] == "(" || fields[0] == ")" {
		return "", "", false
	}
	return fields[0], fields[1], true
}
