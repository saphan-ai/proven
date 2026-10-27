package change

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// WorkTree reports whether target is inside a git working tree.
// A missing git binary is an error. A directory that is not a repository is false.
func (p Policy) WorkTree(target string) (bool, error) {
	cmd := p.GitCommand("-C", target, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)) == "true", nil
	}
	// git itself could not run. A non-repository exits non-zero, which is an
	// answer, not a failure.
	if _, ok := err.(*exec.Error); ok {
		return false, fmt.Errorf("git: %w", err)
	}
	return false, nil
}

// TopLevel returns the top level of the git working tree that holds target, as git prints it. A target
// that is not in a working tree is "", false. A missing git binary is an error.
func (p Policy) TopLevel(target string) (string, bool, error) {
	in, err := p.WorkTree(target)
	if err != nil || !in {
		return "", false, err
	}
	top, err := p.gitZText([]string{"-C", target, "rev-parse", "--show-toplevel"})
	if err != nil {
		return "", false, err
	}
	return top, top != "", nil
}

// gitArchive returns the tar of rev's tracked files. It reads the commit and
// does not run the change. The git directory is not in the archive. The
// stream is bounded by the archive limit.
func (p Policy) gitArchive(target, rev string) ([]byte, error) {
	cmd := p.GitCommand("-C", target, "archive", "--format=tar", "--end-of-options", rev)
	return p.outputBounded(cmd, p.Bounds().Archive, "the archive of the commit", EnvMaxArchiveBytes)
}

// maxFailureBytes is the most of a command's standard error kept for a failure message.
const maxFailureBytes = 512

// clippedBuffer keeps the first max bytes written to it and discards the rest, so what a repository
// makes git print on standard error is never held whole (ADR-0013).
type clippedBuffer struct {
	buf bytes.Buffer
	max int
}

// Write keeps what fits and reports every byte as written, so the command never blocks or fails on it.
func (c *clippedBuffer) Write(p []byte) (int, error) {
	if room := c.max - c.buf.Len(); room > 0 {
		if len(p) < room {
			room = len(p)
		}
		c.buf.Write(p[:room])
	}
	return len(p), nil
}

// String returns what was kept.
func (c *clippedBuffer) String() string { return c.buf.String() }

// outputBounded runs cmd and returns its standard output. A repository can make
// git print as much as it likes, so the read stops one byte past max and the
// command is ended: the output is never held whole. The refusal names what was
// read and the bound, and never the output.
func (p Policy) outputBounded(cmd *exec.Cmd, max int64, what, env string) ([]byte, error) {
	stderr := clippedBuffer{max: maxFailureBytes}
	cmd.Stderr = &stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("git: %w", err)
	}
	out, readErr := io.ReadAll(io.LimitReader(pipe, max+1))
	if readErr == nil && int64(len(out)) > max {
		// Past the bound: end git, so it does not keep writing into a pipe
		// nobody reads, and report the bound.
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, OverBound(what, max, env)
	}
	waitErr := cmd.Wait()
	if readErr != nil {
		return nil, readErr
	}
	if waitErr != nil {
		// ADR-0013: standard output is repository content, so git's failure message is its
		// standard error alone, and that is clipped.
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return nil, &gitFailure{msg: p.offlineMessage(cmd, msg)}
	}
	return out, nil
}

// Between lists the files changed between baseline and head in target's git
// repository. An empty head means the working tree, including untracked files.
// A baseline or a head that does not exist returns an error that names it.
func (p Policy) Between(target, baseline, head string) ([]string, error) {
	// Check both commits first, so a wrong name is reported as that name and
	// not as git's own diff error.
	if err := p.ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	if head != "" {
		if err := p.ensureCommit(target, head, "head"); err != nil {
			return nil, err
		}
	}
	// --relative keeps the paths relative to target when it is a subdirectory,
	// -z keeps unusual file names intact, and --end-of-options stops a
	// revision that starts with "-" from being read as a flag.
	args := []string{"-C", target, "diff", "--name-only", "--relative", "-z", "--end-of-options", baseline}
	if head != "" {
		args = append(args, head)
	}
	names, err := p.gitZ(args)
	if err != nil {
		return nil, err
	}
	// git diff does not list untracked files, but a new file in the working
	// tree is part of the change. ls-files names them from the repository
	// root, so underTarget keeps only those inside target.
	if head == "" {
		extra, err := p.gitZ([]string{"-C", target, "ls-files", "--others", "--exclude-standard", "--full-name", "-z"})
		if err != nil {
			return nil, err
		}
		extra, err = p.underTarget(target, extra)
		if err != nil {
			return nil, err
		}
		names = append(names, extra...)
	}
	return uniqueSorted(names), nil
}

// Status is one changed file and how it changed: A added, M modified, D deleted, R renamed. From is
// the old path of a rename.
type Status struct {
	Status string
	Path   string
	From   string
}

// StatusBetween lists the files changed between baseline and head with their status, for the readers
// that need to know a file was added or moved, not only that it changed (ADR-0049). An empty head means
// the working tree, and an untracked file there is added. Between is left as it is: the review classes
// read paths only (ADR-0014).
func (p Policy) StatusBetween(target, baseline, head string) ([]Status, error) {
	if err := p.ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	if head != "" {
		if err := p.ensureCommit(target, head, "head"); err != nil {
			return nil, err
		}
	}
	args := []string{"-C", target, "diff", "--name-status", "--find-renames", "--relative", "-z", "--end-of-options", baseline}
	if head != "" {
		args = append(args, head)
	}
	fields, err := p.gitZ(args)
	if err != nil {
		return nil, err
	}
	var out []Status
	// With -z, each entry is the status, then the path, and a rename or copy adds the new path after the old.
	for i := 0; i < len(fields); {
		st := fields[i]
		if st == "" {
			i++
			continue
		}
		letter := st[:1]
		switch letter {
		case "R", "C":
			if i+2 >= len(fields) {
				return nil, fmt.Errorf("git: a rename entry without its two paths")
			}
			s := Status{Status: "R", From: fields[i+1], Path: fields[i+2]}
			// A copy leaves the old file in place: the new path is an added file.
			if letter == "C" {
				s = Status{Status: "A", Path: fields[i+2]}
			}
			out = append(out, s)
			i += 3
		default:
			if i+1 >= len(fields) {
				return nil, fmt.Errorf("git: a status entry without its path")
			}
			// T (type changed) and U (unmerged) are changes to an existing path, read as modified.
			if !strings.Contains("AMD", letter) {
				letter = "M"
			}
			out = append(out, Status{Status: letter, Path: fields[i+1]})
			i += 2
		}
	}
	if head == "" {
		extra, err := p.gitZ([]string{"-C", target, "ls-files", "--others", "--exclude-standard", "--full-name", "-z"})
		if err != nil {
			return nil, err
		}
		extra, err = p.underTarget(target, extra)
		if err != nil {
			return nil, err
		}
		for _, p := range extra {
			out = append(out, Status{Status: "A", Path: p})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// Blob reads one file from rev. ok is false when that path is not in the commit.
// A missing path is not an error. The caller decides what the absence means.
func (p Policy) Blob(target, rev, rel string) ([]byte, bool, error) {
	if err := p.ensureCommit(target, rev, "baseline"); err != nil {
		return nil, false, err
	}
	rel = filepath.ToSlash(rel)
	// Refuse a path that could read outside target's tree.
	if rel == "" || strings.HasPrefix(rel, "/") || strings.Contains(rel, "..") {
		return nil, false, fmt.Errorf("git: path %q leaves the tree", rel)
	}
	// "./" makes the path relative to target, not to the repository root (git rev:path rule).
	cmd := p.GitCommand("-C", target, "show", "--end-of-options", rev+":./"+rel)
	out, err := p.outputBounded(cmd, p.Bounds().File, "the file "+rel+" in the commit", EnvMaxFileBytes)
	if err == nil {
		return out, true, nil
	}
	var gf *gitFailure
	if !errors.As(err, &gf) {
		return nil, false, err
	}
	// These are git's words for a path absent from the commit. That is an
	// answer, not a failure, so it returns ok false with no error.
	if strings.Contains(gf.msg, "does not exist") || strings.Contains(gf.msg, "exists on disk, but not in") {
		return nil, false, nil
	}
	return nil, false, err
}

// gitFailure is git exiting non-zero. It is a type of its own so a caller can
// tell git's answer from a refusal made by this package.
type gitFailure struct{ msg string }

// Error reports git's own message.
func (e *gitFailure) Error() string { return "git: " + e.msg }

// Messages returns the commit messages between baseline and head, oldest first.
// The baseline commit itself is not included. An empty head means HEAD.
// Proven reads these messages. It does not write a trailer.
func (p Policy) Messages(target, baseline, head string) ([]string, error) {
	if err := p.ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	end := head
	if head == "" {
		end = "HEAD"
	} else if err := p.ensureCommit(target, head, "head"); err != nil {
		return nil, err
	}
	// %x1e ends each message with an ASCII record separator, so a message
	// with blank lines in it is still split in the right place.
	cmd := p.GitCommand("-C", target, "log", "--reverse", "--format=%B%x1e", "--end-of-options", baseline+".."+end)
	// The log is repository content: a commit message can be as long as its
	// author likes, so the read is bounded like a file.
	out, err := p.outputBounded(cmd, p.Bounds().File, "the commit log", EnvMaxFileBytes)
	if err != nil {
		return nil, err
	}
	return splitMessages(out), nil
}

// splitMessages splits git log output on the record separator and drops the
// empty pieces.
func splitMessages(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	parts := bytes.Split(b, []byte{0x1e})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(string(p))
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ensureCommit checks that rev names a commit in target's repository. which is
// "baseline" or "head", so the error says which input was wrong.
func (p Policy) ensureCommit(target, rev, which string) error {
	if rev == "" {
		return fmt.Errorf("%s commit %q does not exist", which, rev)
	}
	cmd := p.GitCommand("-C", target, "rev-parse", "--verify", "--quiet", "--end-of-options", rev+"^{commit}")
	_, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	// git itself could not run: report that, not a missing commit.
	if _, ok := err.(*exec.Error); ok {
		return fmt.Errorf("git: %w", err)
	}
	return fmt.Errorf("%s commit %q does not exist", which, rev)
}

// ResolveCommit returns the full commit id rev names in target's repository.
// It is empty when rev is empty or git cannot resolve it: the caller keeps the
// name it was given and records no resolution.
func (p Policy) ResolveCommit(target, rev string) string {
	if rev == "" {
		return ""
	}
	out, err := p.GitCommand("-C", target, "rev-parse", "--verify", "--quiet", "--end-of-options", rev+"^{commit}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitZ runs git with args and splits its NUL-separated output into slash
// paths. A failure carries git's own message.
//
// These commands compare or list the working tree, and git converts a file it
// compares (the clean filter) and asks the file system monitor about the tree.
// In local mode they run with no attribute and no monitor that could name a
// program Proven did not choose (withoutAttributes).
func (p Policy) gitZ(args []string) ([]string, error) {
	args, err := p.withoutAttributes(args)
	if err != nil {
		return nil, err
	}
	cmd := p.GitCommand(args...)
	// A list of path names is repository content too, so it has the archive bound.
	out, err := p.outputBounded(cmd, p.Bounds().Archive, "the list of file names", EnvMaxArchiveBytes)
	if err != nil {
		return nil, err
	}
	return splitZ(out), nil
}

// splitZ splits NUL-separated git output into slash paths.
func splitZ(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	parts := bytes.Split(b, []byte{0})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		out = append(out, filepath.ToSlash(string(p)))
	}
	return out
}

// underTarget keeps the repository-root paths that lie inside target and makes
// them relative to target, to match what git diff --relative returns.
func (p Policy) underTarget(target string, paths []string) ([]string, error) {
	topRaw, err := p.gitZText([]string{"-C", target, "rev-parse", "--show-toplevel"})
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}
	// Compare real paths: on macOS the temporary directory is a symlink (/var -> /private/var).
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		abs = real
	}
	if real, err := filepath.EvalSymlinks(topRaw); err == nil {
		topRaw = real
	}
	rel, err := filepath.Rel(topRaw, abs)
	if err != nil {
		return nil, err
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return paths, nil
	}
	prefix := rel + "/"
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if strings.HasPrefix(p, prefix) {
			out = append(out, strings.TrimPrefix(p, prefix))
		}
	}
	return out, nil
}

// gitZText runs git with args and returns its trimmed output as one string.
func (p Policy) gitZText(args []string) (string, error) {
	cmd := p.GitCommand(args...)
	out, err := cmd.Output()
	if err != nil {
		detail := err.Error()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.TrimSpace(string(exitErr.Stderr)) != "" {
			detail = strings.TrimSpace(string(exitErr.Stderr))
		}
		return "", fmt.Errorf("git: %s", p.offlineMessage(cmd, detail))
	}
	return strings.TrimSpace(string(out)), nil
}

// uniqueSorted sorts paths and drops repeats, so each file is listed once and
// in the same order on every run.
func uniqueSorted(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	sort.Strings(paths)
	out := make([]string, 0, len(paths))
	prev := ""
	for i, p := range paths {
		if i > 0 && p == prev {
			continue
		}
		out = append(out, p)
		prev = p
	}
	return out
}
