package change

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Property: a tracked file whose parent directory has been replaced by a symbolic link in the
// working tree is not copied from behind the link; it is left out and named as skipped (ADR-0025).
func TestExportWorktreeDoesNotCopyThroughALinkedParent(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	writeTree(t, filepath.Join(dir, "sub", "a.txt"), "tracked")
	git(t, dir, "add", "sub/a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")

	// Mutated field: only the parent, now a link to a directory holding a file of the same name.
	outside := t.TempDir()
	write(t, filepath.Join(outside, "a.txt"), "outside-content")
	if err := os.RemoveAll(filepath.Join(dir, "sub")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "sub")); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	skipped, err := ExportWorktree(dir, dst)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(dst, "sub", "a.txt")); err == nil {
		t.Fatal("a file was copied through a symbolic link in its parent directory")
	}
	if !slices.Contains(skipped, "sub/a.txt") {
		t.Fatalf("the path behind the link was not named as skipped: %v", skipped)
	}
}
