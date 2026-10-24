package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLinkedTreesAreNotAPreparedPair proves a base or change entry that is a symbolic link does not
// count as a prepared tree (ADR-0025). Mutated field: only the type of change, a link to a directory.
func TestLinkedTreesAreNotAPreparedPair(t *testing.T) {
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, "base"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(target, "change"), 0o700); err != nil {
		t.Fatal(err)
	}
	if !hasPreparedTrees(target) {
		t.Fatal("two real directories were not taken for a prepared pair")
	}
	if err := os.Remove(filepath.Join(target, "change")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(target, "change")); err != nil {
		t.Fatal(err)
	}
	if hasPreparedTrees(target) {
		t.Fatal("a symbolic link was taken for a prepared tree")
	}
}
