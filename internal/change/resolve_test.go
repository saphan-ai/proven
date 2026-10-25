package change

import (
	"os"
	"path/filepath"
	"testing"
)

// ResolveCommit names the commit a revision resolves to, and nothing when it
// resolves to none. A revision that starts with a dash is a name, never an
// option: it is the field this test mutates.
func TestResolveCommit(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")
	head := rev(t, dir)

	for name, tc := range map[string]struct{ rev, want string }{
		"empty is the working tree": {"", ""},
		"HEAD":                      {"HEAD", head},
		"the full id":               {head, head},
		"an unknown name":           {"no-such-branch", ""},
		"an option name":            {"--all", ""},
		"an option with a value":    {"--output=" + filepath.Join(dir, "owned"), ""},
	} {
		t.Run(name, func(t *testing.T) {
			if got := ResolveCommit(dir, tc.rev); got != tc.want {
				t.Fatalf("ResolveCommit(%q) = %q, want %q", tc.rev, got, tc.want)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(dir, "owned")); err == nil {
		t.Fatal("a revision that looks like an option wrote a file")
	}
}
