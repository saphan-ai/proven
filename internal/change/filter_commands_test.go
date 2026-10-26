package change

import (
	"os"
	"path/filepath"
	"testing"
)

// filterRepo makes a repository whose f.txt names a filter driver defined in the user's ~/.gitconfig.
// The driver writes marker when it runs, for any of clean, smudge or process. HOME is the test's own.
func filterRepo(t *testing.T) (repo, marker string) {
	t.Helper()
	home := t.TempDir()
	marker = filepath.Join(t.TempDir(), "driver-ran")
	script := filepath.Join(home, "driver.sh")
	write(t, script, "#!/bin/sh\ntouch '"+marker+"'\nexec cat\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, ".gitconfig"), "[filter \"probe\"]\n\tclean = "+script+"\n\tsmudge = "+script+"\n[diff \"probe\"]\n\ttextconv = "+script+"\n")
	t.Setenv("HOME", home)
	repo = t.TempDir()
	git(t, repo, "init")
	write(t, filepath.Join(repo, ".gitattributes"), "f.txt filter=probe diff=probe\n")
	write(t, filepath.Join(repo, "f.txt"), "hello\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[probe] fixture: one")
	os.Remove(marker)
	return repo, marker
}

// TestLocalWorkingTreeReadsRunNoFilterDriver: the commands that look at the working tree (the diff of a
// working-tree review, the untracked listing) compare file content, which git converts with the clean
// filter. In local mode none of them starts a driver. Mutated field: the content of f.txt in the working
// tree, so git has a reason to filter it.
func TestLocalWorkingTreeReadsRunNoFilterDriver(t *testing.T) {
	repo, marker := filterRepo(t)
	write(t, filepath.Join(repo, "f.txt"), "changed\n")
	write(t, filepath.Join(repo, "new.txt"), "new\n")
	defer UseOffline(UseOffline(true))
	names, err := Between(repo, "HEAD", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("positive control: Between named %v, want f.txt and new.txt", names)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("local mode: a git command of Between ran the filter driver from ~/.gitconfig")
	}
	if _, err := ExportWorktree(repo, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("local mode: a git command of ExportWorktree ran the filter driver from ~/.gitconfig")
	}
	// The readers of a commit run no driver either: a blob is shown as stored and a log has no content.
	if _, ok, err := Blob(repo, "HEAD", "f.txt"); err != nil || !ok {
		t.Fatalf("Blob: %v %v", err, ok)
	}
	if _, err := Messages(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("local mode: a reader of a commit ran the textconv or filter driver from ~/.gitconfig")
	}
	// Control: git does run the textconv driver for a diff, so a pass above is not a driver that cannot run.
	diff := GitCommand("-C", repo, "diff", "HEAD", "--", "f.txt")
	diff.Dir = repo
	diff.Env = append(diff.Env, "TMPDIR="+t.TempDir()) // git writes the converted text to a temporary file
	if out, err := diff.CombinedOutput(); err != nil {
		t.Fatalf("diff: %v %s", err, out)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("control: git diff did not run the driver, so the readers' result proves nothing")
	}
	os.Remove(marker)
	// Control: online, the same diff runs the driver, so a pass above is not a driver that cannot run.
	UseOffline(false)
	if _, err := Between(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("control: online, Between did not run the filter driver, so the local result proves nothing")
	}
}
