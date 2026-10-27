package change

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// hookClone is a clone whose tracked f.txt is unchanged in content but carries
// a new modification time, so a working-tree diff finds it stat-dirty and may
// refresh the index. The clone's own .git/hooks/post-index-change leaves a
// marker when git starts it.
func hookClone(t *testing.T) (string, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	marker := filepath.Join(t.TempDir(), "hook-ran")
	repo := t.TempDir()
	git(t, repo, "init")
	write(t, filepath.Join(repo, "f.txt"), "one\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[probe] fixture: first")
	hook := filepath.Join(repo, ".git", "hooks", "post-index-change")
	write(t, hook, "#!/bin/sh\ntouch '"+marker+"'\n")
	if err := os.Chmod(hook, 0o700); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(repo, "f.txt"), old, old); err != nil {
		t.Fatal(err)
	}
	return repo, marker
}

// TestLocalWorkingTreeReadStartsNoHook: in local mode, reading the working tree
// starts no hook of the clone. Online, the same clone starts it, which shows the
// fixture can. Mutated field: the mtime of f.txt (content unchanged).
func TestLocalWorkingTreeReadStartsNoHook(t *testing.T) {
	repo, marker := hookClone(t)
	defer UseOffline(UseOffline(true))
	if _, err := Between(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("local mode started the clone's post-index-change hook while reading the working tree (%v)", err)
	}
	repo2, marker2 := hookClone(t)
	UseOffline(false)
	if _, err := Between(repo2, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker2); err != nil {
		t.Fatalf("online, the fixture's hook did not start, so the local check proves nothing: %v", err)
	}
}
