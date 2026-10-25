package change

import (
	"path/filepath"
	"testing"
)

// TestStatusBetweenNamesHowEachFileChanged proves StatusBetween gives each changed file its status: an
// added, a modified, a deleted and a renamed file between two commits, and an untracked file in the
// working tree as added. An unchanged file is not listed.
func TestStatusBetweenNamesHowEachFileChanged(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "keep.txt"), "same\n")
	write(t, filepath.Join(dir, "edit.txt"), "old\n")
	write(t, filepath.Join(dir, "gone.txt"), "bye\n")
	write(t, filepath.Join(dir, "move.txt"), "a body long enough for git to see the same file after the move\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: first")
	first := rev(t, dir)
	write(t, filepath.Join(dir, "edit.txt"), "new\n")
	write(t, filepath.Join(dir, "new.txt"), "hello\n")
	git(t, dir, "rm", "-q", "gone.txt")
	git(t, dir, "mv", "move.txt", "moved.txt")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: second")
	second := rev(t, dir)

	got, err := StatusBetween(dir, first, second)
	if err != nil {
		t.Fatal(err)
	}
	want := []Status{
		{Status: "M", Path: "edit.txt"},
		{Status: "D", Path: "gone.txt"},
		{Status: "R", Path: "moved.txt", From: "move.txt"},
		{Status: "A", Path: "new.txt"},
	}
	if len(got) != len(want) {
		t.Fatalf("%+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%d: %+v, want %+v", i, got[i], want[i])
		}
	}

	write(t, filepath.Join(dir, "untracked.txt"), "x\n")
	work, err := StatusBetween(dir, second, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(work) != 1 || work[0] != (Status{Status: "A", Path: "untracked.txt"}) {
		t.Fatalf("%+v", work)
	}
}

// TestStatusBetweenNamesAMissingBaseline proves a wrong baseline is refused by name, as Between refuses it.
func TestStatusBetweenNamesAMissingBaseline(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")
	if _, err := StatusBetween(dir, "no-such-commit", ""); err == nil {
		t.Fatal("a missing baseline was accepted")
	}
}
