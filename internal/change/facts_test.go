package change

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFactsBetweenReadsModeAndShebang proves FactsBetween reports, per changed file, an executable bit set
// on the head side, a shebang line at the head, a plain file as neither, and a rename under both its paths.
// Mutated field, one per file: the mode bit (run.sh), the first line (note.md), and neither (plain.md).
func TestFactsBetweenReadsModeAndShebang(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "run.sh"), "echo old\n")
	write(t, filepath.Join(dir, "note.md"), "old\n")
	write(t, filepath.Join(dir, "plain.md"), "old\n")
	write(t, filepath.Join(dir, "move.md"), "a body long enough for git to see the same file after the move\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: first")
	first := rev(t, dir)
	write(t, filepath.Join(dir, "run.sh"), "echo new\n")
	if err := os.Chmod(filepath.Join(dir, "run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "note.md"), "#!/bin/sh\necho new\n")
	write(t, filepath.Join(dir, "plain.md"), "new\n")
	git(t, dir, "mv", "move.md", "moved.md")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: second")
	second := rev(t, dir)

	got, err := FactsBetween(dir, first, second)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Facts{
		"run.sh":   {Executable: true},
		"note.md":  {Shebang: true},
		"plain.md": {},
		"moved.md": {},
		"move.md":  {},
	}
	if len(got) != len(want) {
		t.Fatalf("%+v", got)
	}
	for p, f := range want {
		if got[p] != f {
			t.Errorf("%s: %+v, want %+v", p, got[p], f)
		}
	}

	// The working tree: an untracked executable file is read from disk.
	write(t, filepath.Join(dir, "fresh.md"), "x\n")
	if err := os.Chmod(filepath.Join(dir, "fresh.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	work, err := FactsBetween(dir, second, "")
	if err != nil {
		t.Fatal(err)
	}
	if f := work["fresh.md"]; !f.Executable || f.Unverified != "" {
		t.Fatalf("%+v", work)
	}
}

// TestTreeFactsReadsPreparedTrees proves TreeFacts reads the mode from either side of a prepared pair and
// refuses to call a file absent from both sides verified.
func TestTreeFactsReadsPreparedTrees(t *testing.T) {
	root := t.TempDir()
	writeTree(t, filepath.Join(root, "base", "a.md"), "old\n")
	writeTree(t, filepath.Join(root, "change", "a.md"), "new\n")
	writeTree(t, filepath.Join(root, "base", "gone.md"), "#!/bin/sh\n")
	if err := os.Chmod(filepath.Join(root, "base", "a.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := TreeFacts(root, []string{"a.md", "gone.md", "nowhere.md"})
	if f := got["a.md"]; !f.Executable || f.Shebang || f.Unverified != "" {
		t.Errorf("a.md %+v", f)
	}
	if f := got["gone.md"]; !f.Shebang || f.Unverified != "" {
		t.Errorf("gone.md %+v", f)
	}
	if f := got["nowhere.md"]; f.Unverified == "" {
		t.Errorf("nowhere.md %+v", f)
	}
}
