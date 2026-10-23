package testrun

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DefaultMaxWrittenBytes is the size limit for a test file the Prover writes.
// A ContainerConfig may set a different positive limit. Zero is refused.
const DefaultMaxWrittenBytes = 64 << 10

// Refusal reasons. The trace prefixes each one with "refused: ".
const (
	ReasonAbsolutePath = "absolute path"
	ReasonLeavesRepo   = "path leaves the repository"
	ReasonNotATestFile = "not a test file"
	ReasonOverwrites   = "overwrites a non-test file"
	ReasonTooLarge     = "content over the size limit"
	ReasonEmptyCommand = "empty command"
)

// RefuseWritten reports why a written test must not run. It guards the one
// place where the Prover adds code to the change: the file must be a Go test
// file inside the repository, must not replace a non-test file, and must stay
// under the size limit.
// An empty reason means the request may run. maxBytes <= 0 uses the default.
// The check looks at the path, the command, the size and the two checkouts.
// It does not write a file and it does not start a runtime.
func RefuseWritten(path, content, command, baseDir, changeDir string, maxBytes int) string {
	if strings.TrimSpace(command) == "" {
		return ReasonEmptyCommand
	}
	if filepath.IsAbs(path) {
		return ReasonAbsolutePath
	}
	// A ".." segment could write outside the checkout copy. The ../ refusal
	// test fails if this branch is removed.
	if pathLeaves(path) {
		return ReasonLeavesRepo
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxWrittenBytes
	}
	if len(content) > maxBytes {
		return ReasonTooLarge
	}
	// Both checkouts are checked: a written test must not replace a non-test
	// file on either side, or the base run would no longer test the base.
	if nonTestAt(baseDir, path) || nonTestAt(changeDir, path) {
		return ReasonOverwrites
	}
	if !isGoTestFile(path) {
		return ReasonNotATestFile
	}
	return ""
}

// pathLeaves reports whether path has a ".." segment or cleans to a path above the root.
func pathLeaves(path string) bool {
	if path == "" {
		return false
	}
	slash := filepath.ToSlash(path)
	for _, seg := range strings.Split(slash, "/") {
		if seg == ".." {
			return true
		}
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == ".." || strings.HasPrefix(clean, "../")
}

// isGoTestFile reports whether path names a file ending in _test.go.
func isGoTestFile(path string) bool {
	base := filepath.Base(filepath.ToSlash(path))
	if base == "." || base == ".." || base == string(filepath.Separator) {
		return false
	}
	return strings.HasSuffix(base, "_test.go")
}

// nonTestAt reports whether rel in dir already holds something a written test
// must not replace: a non-test file, or anything that is not a regular file,
// such as a symlink or a directory. A path that does not exist is free.
func nonTestAt(dir, rel string) bool {
	if dir == "" || rel == "" || filepath.IsAbs(rel) || pathLeaves(rel) {
		return false
	}
	dest := filepath.Join(dir, filepath.FromSlash(rel))
	back, err := filepath.Rel(dir, dest)
	if err != nil || pathLeaves(back) {
		return false
	}
	info, err := os.Lstat(dest)
	if err != nil {
		return false
	}
	if !info.Mode().IsRegular() {
		return true
	}
	return !isGoTestFile(rel)
}

// stageWritten copies src and writes the test file into the copy.
// The caller's tree is not modified. The caller removes the returned directory.
func stageWritten(src, rel, content string) (string, error) {
	dst, err := os.MkdirTemp("", "proven-checkout-")
	if err != nil {
		return "", err
	}
	if err := copyTree(src, dst); err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	if err := writeInside(dst, rel, content); err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	return dst, nil
}

// copyTree copies the regular files and directories of src into dst.
// Symlinks and other special files are skipped, so the copy cannot point
// the container at anything outside the checkout.
func copyTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("testrun: checkout %s is not a directory", src)
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return copyFile(path, target)
	})
}

// copyFile copies one file and reports the close error, which can carry a failed write.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// writeInside writes content at rel under root. It checks the path again,
// after cleaning, so it is safe even when called without RefuseWritten.
func writeInside(root, rel, content string) error {
	rel = filepath.Clean(filepath.FromSlash(rel))
	if rel == "." || pathLeaves(rel) || filepath.IsAbs(rel) {
		return fmt.Errorf("testrun: written path %s is outside the checkout", rel)
	}
	dest := filepath.Join(root, rel)
	parent := filepath.Clean(root)
	if dest != parent && !strings.HasPrefix(dest, parent+string(os.PathSeparator)) {
		return fmt.Errorf("testrun: written path %s is outside the checkout", rel)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, []byte(content), 0o644)
}
