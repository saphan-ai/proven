package change

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLocalGitRunsNoFilterDriver: in PROVEN_NETWORK=local, a git command Proven starts must not run a
// filter driver (git-lfs is one, and its smudge fetches over HTTPS outside the policy). The driver here
// is defined in the user's ~/.gitconfig, where `git lfs install` puts git-lfs, and only writes a marker.
// Expected of a fix: ExportCommit in local mode returns the stored blob and the marker is never written.
// Mutated field: the filter attribute of f.txt (a driver from ~/.gitconfig).
func TestLocalGitRunsNoFilterDriver(t *testing.T) {
	home := t.TempDir()
	marker := filepath.Join(t.TempDir(), "smudge-ran")
	script := filepath.Join(home, "smudge.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntouch '"+marker+"'\nexec cat\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	gitconfig := "[filter \"probe\"]\n\tclean = cat\n\tsmudge = " + script + "\n"
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(gitconfig), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	repo := t.TempDir()
	git(t, repo, "init")
	write(t, filepath.Join(repo, ".gitattributes"), "f.txt filter=probe\n")
	write(t, filepath.Join(repo, "f.txt"), "hello\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[probe] fixture: one")
	os.Remove(marker)

	defer UseOffline(UseOffline(true))
	dst := t.TempDir()
	if _, err := ExportCommit(repo, "HEAD", dst); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("local mode: git archive ran the filter driver from ~/.gitconfig (a git-lfs smudge would connect)")
	}
	if got, err := os.ReadFile(filepath.Join(dst, "f.txt")); err != nil || string(got) != "hello\n" {
		t.Fatalf("the exported file is not the stored blob: %q %v", got, err)
	}
	// Control: the same driver does run through GitCommand when git writes a working-tree file,
	// so a pass above is not a driver that never runs.
	os.Remove(filepath.Join(repo, "f.txt"))
	if out, err := GitCommand("-C", repo, "checkout", "--", "f.txt").CombinedOutput(); err != nil {
		t.Fatalf("checkout: %v %s", err, out)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("control: the filter driver did not run on checkout, so the archive result proves nothing")
	}
}
