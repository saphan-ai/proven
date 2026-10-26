package change

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Facts is what a reader needs to know about one changed file beyond its name before it may treat the
// file as prose that is never built or executed (ADR-0049): whether the file carries an executable bit,
// whether it starts with a shebang line, and whether any of that could not be established. A file whose
// facts could not be read is Unverified, and a caller treats it as not prose.
type Facts struct {
	// Executable is true when the file is executable on either side of the change, or is not a regular file.
	Executable bool
	// Shebang is true when the file's content starts with "#!" on either side of the change.
	Shebang bool
	// Unverified is the reason the facts could not be established, or empty.
	Unverified string
}

// symlinkMode is git's mode for a symbolic link; a link is never prose because what it names is unknown.
const symlinkMode = "120000"

// errTruncatedRaw is the refusal of a raw diff entry that lacks its path.
var errTruncatedRaw = errors.New("git: a raw diff entry without its path")

// FactsBetween reads the facts of every file changed between baseline and head, keyed by path. An empty
// head means the working tree. A rename's facts are keyed under both its old and its new path. Modes come
// from git's own raw diff; the shebang is read from the file's content on both sides, as the executable
// bit is (at baseline unless the file is new, at head unless it is gone, and from disk for the working tree).
func (p Policy) FactsBetween(target, baseline, head string) (map[string]Facts, error) {
	if err := p.ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	if head != "" {
		if err := p.ensureCommit(target, head, "head"); err != nil {
			return nil, err
		}
	}
	args := []string{"-C", target, "diff", "--raw", "--find-renames", "--relative", "-z", "--end-of-options", baseline}
	if head != "" {
		args = append(args, head)
	}
	fields, err := p.gitZ(args)
	if err != nil {
		return nil, err
	}
	out := map[string]Facts{}
	// With -z, a raw entry is ":oldmode newmode oldsha newsha status", then the path; a rename or copy
	// names two paths after the status.
	for i := 0; i < len(fields); {
		meta := strings.Fields(strings.TrimPrefix(fields[i], ":"))
		if !strings.HasPrefix(fields[i], ":") || len(meta) < 5 {
			i++
			continue
		}
		letter := meta[4][:1]
		paths := []string{fields[i+1]}
		next := i + 2
		if letter == "R" || letter == "C" {
			if i+2 >= len(fields) {
				return nil, errTruncatedRaw
			}
			paths = []string{fields[i+1], fields[i+2]}
			next = i + 3
		} else if i+1 >= len(fields) {
			return nil, errTruncatedRaw
		}
		f := Facts{Executable: executableMode(meta[0]) || executableMode(meta[1])}
		if meta[0] == symlinkMode || meta[1] == symlinkMode {
			f.Unverified = "a symbolic link"
		}
		// A shebang on either side of the change counts, as an executable bit does: a script that loses its
		// header is still a script change, and a model must review it. Each side that exists is read: the
		// baseline unless the file is new, the head unless the file is gone (a rename has both paths).
		if f.Unverified == "" && letter != "A" {
			f.Shebang, f.Unverified = p.gitShebang(target, baseline, paths[0])
		}
		if f.Unverified == "" && letter != "D" {
			var sb bool
			sb, f.Unverified = p.gitShebang(target, head, paths[len(paths)-1])
			f.Shebang = f.Shebang || sb
		}
		for _, p := range paths {
			out[p] = f
		}
		i = next
	}
	if head == "" {
		// An untracked file is part of the change and has no raw entry: stat it on disk.
		extra, err := p.gitZ([]string{"-C", target, "ls-files", "--others", "--exclude-standard", "--full-name", "-z"})
		if err != nil {
			return nil, err
		}
		extra, err = p.underTarget(target, extra)
		if err != nil {
			return nil, err
		}
		for _, p := range extra {
			out[p] = diskFacts(filepath.Join(target, filepath.FromSlash(p)))
		}
	}
	return out, nil
}

// TreeFacts reads the facts of paths in a prepared base/ and change/ pair under root, keyed by path. A
// file is executable when it is on either side, and has a shebang when either side starts with one.
func TreeFacts(root string, paths []string) map[string]Facts {
	out := make(map[string]Facts, len(paths))
	for _, p := range paths {
		rel := filepath.FromSlash(p)
		var f Facts
		last := ""
		for _, side := range []string{"base", "change"} {
			df := diskFacts(filepath.Join(root, side, rel))
			if df.Unverified != "" && !os.IsNotExist(statErr(filepath.Join(root, side, rel))) {
				f.Unverified = df.Unverified
			}
			if df.Unverified == "" {
				f.Executable = f.Executable || df.Executable
				f.Shebang, last = f.Shebang || df.Shebang, side
			}
		}
		if last == "" && f.Unverified == "" {
			f.Unverified = "the file is on neither side"
		}
		out[p] = f
	}
	return out
}

// statErr returns the error Lstat gives for path, or nil.
func statErr(path string) error {
	_, err := os.Lstat(path)
	return err
}

// diskFacts reads one file from disk. Anything that is not a regular file is Executable and Unverified,
// because what a link or a device stands for is not known.
func diskFacts(path string) Facts {
	info, err := os.Lstat(path)
	if err != nil {
		return Facts{Unverified: "cannot be read: " + err.Error()}
	}
	if !info.Mode().IsRegular() {
		return Facts{Executable: true, Unverified: "not a regular file"}
	}
	f := Facts{Executable: info.Mode().Perm()&0o111 != 0}
	fh, err := os.Open(path)
	if err != nil {
		f.Unverified = "cannot be read: " + err.Error()
		return f
	}
	defer fh.Close()
	buf := make([]byte, 2)
	n, err := io.ReadFull(fh, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		f.Unverified = "cannot be read: " + err.Error()
		return f
	}
	f.Shebang = n == 2 && string(buf) == "#!"
	return f
}

// gitShebang reports whether the file at rev:name starts with "#!". An empty rev reads the working tree.
// A file that cannot be read is reported as unverified, never as free of a shebang.
func (p Policy) gitShebang(target, rev, name string) (bool, string) {
	if rev == "" {
		df := diskFacts(filepath.Join(target, filepath.FromSlash(name)))
		return df.Shebang, df.Unverified
	}
	// "./" makes the path relative to target, as --relative made the listed names.
	cmd := p.GitCommand("-C", target, "cat-file", "blob", rev+":./"+name)
	b, err := p.outputBounded(cmd, p.Bounds().File, "the file "+name, EnvMaxFileBytes)
	if err != nil {
		return false, "cannot be read at " + rev + ": " + err.Error()
	}
	return len(b) >= 2 && b[0] == '#' && b[1] == '!', ""
}

// executableMode reports whether a git mode carries an executable bit. The all-zero mode of an absent
// side is not executable.
func executableMode(mode string) bool {
	return len(mode) >= 3 && mode != "000000" && strings.ContainsAny(mode[len(mode)-3:], "1357")
}
