package change

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Between lists the files changed between baseline and head in target's git
// repository. An empty head means the working tree, including untracked files.
// A baseline or a head that does not exist returns an error that names it.
func Between(target, baseline, head string) ([]string, error) {
	if err := ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	if head != "" {
		if err := ensureCommit(target, head, "head"); err != nil {
			return nil, err
		}
	}
	args := []string{"-C", target, "diff", "--name-only", "--relative", "-z", "--end-of-options", baseline}
	if head != "" {
		args = append(args, head)
	}
	names, err := gitZ(args)
	if err != nil {
		return nil, err
	}
	if head == "" {
		extra, err := gitZ([]string{"-C", target, "ls-files", "--others", "--exclude-standard", "--full-name", "-z"})
		if err != nil {
			return nil, err
		}
		extra, err = underTarget(target, extra)
		if err != nil {
			return nil, err
		}
		names = append(names, extra...)
	}
	return uniqueSorted(names), nil
}

// Blob reads one file from rev. ok is false when that path is not in the commit.
// A missing path is not an error. The caller decides what the absence means.
func Blob(target, rev, rel string) ([]byte, bool, error) {
	if err := ensureCommit(target, rev, "baseline"); err != nil {
		return nil, false, err
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || strings.HasPrefix(rel, "/") || strings.Contains(rel, "..") {
		return nil, false, fmt.Errorf("git: path %q leaves the tree", rel)
	}
	cmd := exec.Command("git", "-C", target, "show", "--end-of-options", rev+":"+rel)
	out, err := cmd.Output()
	if err == nil {
		return out, true, nil
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		return nil, false, fmt.Errorf("git: %w", err)
	}
	msg := strings.TrimSpace(string(ee.Stderr))
	if strings.Contains(msg, "does not exist") || strings.Contains(msg, "exists on disk, but not in") {
		return nil, false, nil
	}
	if msg == "" {
		msg = err.Error()
	}
	return nil, false, fmt.Errorf("git: %s", msg)
}

// Messages returns the commit messages between baseline and head, oldest first.
// The baseline commit itself is not included. An empty head means HEAD.
// Proven reads these messages. It does not write a trailer.
func Messages(target, baseline, head string) ([]string, error) {
	if err := ensureCommit(target, baseline, "baseline"); err != nil {
		return nil, err
	}
	end := head
	if head == "" {
		end = "HEAD"
	} else if err := ensureCommit(target, head, "head"); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "-C", target, "log", "--reverse", "--format=%B%x1e", "--end-of-options", baseline+".."+end)
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if ee, ok := err.(*exec.ExitError); ok && msg == "" {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git: %s", msg)
	}
	return splitMessages(out), nil
}

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

func ensureCommit(target, rev, which string) error {
	if rev == "" {
		return fmt.Errorf("%s commit %q does not exist", which, rev)
	}
	cmd := exec.Command("git", "-C", target, "rev-parse", "--verify", "--quiet", "--end-of-options", rev+"^{commit}")
	_, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if _, ok := err.(*exec.Error); ok {
		return fmt.Errorf("git: %w", err)
	}
	return fmt.Errorf("%s commit %q does not exist", which, rev)
}

func gitZ(args []string) ([]string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if ee, ok := err.(*exec.ExitError); ok && msg == "" {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git: %s", msg)
	}
	return splitZ(out), nil
}

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

func underTarget(target string, paths []string) ([]string, error) {
	topRaw, err := gitZText([]string{"-C", target, "rev-parse", "--show-toplevel"})
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

func gitZText(args []string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

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
