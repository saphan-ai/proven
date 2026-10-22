package testrun

import (
	"os"
	"path/filepath"
	"testing"
)

// A checkout may hold a symlinked directory that points outside it. A written test aimed
// through that directory must land in the copy and never in the symlink's target.
func TestWrittenTestThroughASymlinkedDirStaysInTheCopy(t *testing.T) {
	outside := t.TempDir()
	src := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(src, "pkg")); err != nil {
		t.Skipf("symlink: %v", err)
	}
	dst, err := stageWritten(src, "pkg/escape_test.go", "package pkg\n")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst)
	if _, err := os.Lstat(filepath.Join(outside, "escape_test.go")); err == nil {
		t.Fatal("written test reached the symlink's target outside the checkout")
	}
	info, err := os.Lstat(filepath.Join(dst, "pkg", "escape_test.go"))
	if err != nil {
		t.Fatalf("written test not in the copy: %v", err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("written test is not a regular file: %v", info.Mode())
	}
	if info, err := os.Lstat(filepath.Join(dst, "pkg")); err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("the copy kept the symlink")
	}
}
