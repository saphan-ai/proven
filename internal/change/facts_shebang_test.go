package change

import (
	"path/filepath"
	"testing"
)

// TestFactsBetweenKeepsABaselineShebang proves ADR-0049's rule that shebang evidence is aggregated across
// both sides of a change, as the executable bit is: a script that loses its "#!" line, in place or across
// a rename, still reports Shebang, so a caller never takes the edit for prose. Mutated field, one per
// case: the first line of the head side (removed in place), the same with a rename, and a file whose
// head side keeps its shebang (the head side still counts).
func TestFactsBetweenKeepsABaselineShebang(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	script := "#!/bin/sh\necho a body long enough for git to pair the renamed file with its old path\n"
	write(t, filepath.Join(dir, "tool.md"), script)
	write(t, filepath.Join(dir, "old.md"), script)
	write(t, filepath.Join(dir, "kept.md"), script)
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: first")
	first := rev(t, dir)
	write(t, filepath.Join(dir, "tool.md"), "plain text now\n")
	write(t, filepath.Join(dir, "kept.md"), script+"echo more\n")
	git(t, dir, "mv", "old.md", "new.md")
	write(t, filepath.Join(dir, "new.md"), "echo a body long enough for git to pair the renamed file with its old path\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: second")
	second := rev(t, dir)

	got, err := FactsBetween(dir, first, second)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"tool.md", "kept.md"} {
		if f := got[p]; !f.Shebang || f.Unverified != "" {
			t.Errorf("%s: %+v, want Shebang", p, f)
		}
	}
	// A rename is keyed under both paths, so a caller checking either one sees the old side's shebang.
	for _, p := range []string{"old.md", "new.md"} {
		if f, ok := got[p]; !ok || !f.Shebang || f.Unverified != "" {
			t.Errorf("%s: %+v (listed %t), want Shebang", p, f, ok)
		}
	}
}

// TestTreeFactsKeepsABaseShebang proves the same aggregation for a prepared pair: a file whose base side
// starts with "#!" and whose change side does not still reports Shebang.
func TestTreeFactsKeepsABaseShebang(t *testing.T) {
	root := t.TempDir()
	writeTree(t, filepath.Join(root, "base", "tool.md"), "#!/bin/sh\necho old\n")
	writeTree(t, filepath.Join(root, "change", "tool.md"), "plain text now\n")
	writeTree(t, filepath.Join(root, "base", "plain.md"), "old\n")
	writeTree(t, filepath.Join(root, "change", "plain.md"), "new\n")
	got := TreeFacts(root, []string{"tool.md", "plain.md"})
	if f := got["tool.md"]; !f.Shebang || f.Unverified != "" {
		t.Errorf("tool.md %+v, want Shebang", f)
	}
	if f := got["plain.md"]; f.Shebang {
		t.Errorf("plain.md %+v, want no Shebang", f)
	}
}
