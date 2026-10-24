//go:build !windows

package change

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// Property: a working-tree export sets each file's mode explicitly and does not
// follow the process umask, so a file lands readable by others (0644, or 0755 when
// the source is executable) even under a strict umask. Mutated field: the umask,
// 077 instead of 022; the export's explicit Chmod is what the test pins.
func TestExportWorktreeSetsFileModesWhateverTheUmask(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	writeTree(t, filepath.Join(dir, "plain.txt"), "tracked")
	writeTree(t, filepath.Join(dir, "run.sh"), "tracked")
	if err := os.Chmod(filepath.Join(dir, "plain.txt"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(dir, "run.sh"), 0o775); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "plain.txt", "run.sh")
	git(t, dir, "commit", "-m", "[change] git: first")

	old := syscall.Umask(0o077)
	defer syscall.Umask(old)
	dst := t.TempDir()
	if _, err := ExportWorktree(dir, dst); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]os.FileMode{"plain.txt": 0o644, "run.sh": 0o755} {
		info, err := os.Lstat(filepath.Join(dst, name))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s: mode %o, want %o", name, got, want)
		}
	}
}
