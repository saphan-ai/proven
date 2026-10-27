package change

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// infoAttributesClone is a clone with a changed f.txt and a filter driver,
// configured in the user's ~/.gitconfig, that leaves a marker when git starts
// it. It returns the clone and the marker's path.
func infoAttributesClone(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	marker := filepath.Join(t.TempDir(), "driver-ran")
	script := filepath.Join(home, "driver.sh")
	write(t, script, "#!/bin/sh\ntouch '"+marker+"'\nexec cat\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, ".gitconfig"), "[filter \"probe\"]\n\tclean = "+script+"\n")
	t.Setenv("HOME", home)
	repo := t.TempDir()
	git(t, repo, "init")
	write(t, filepath.Join(repo, "f.txt"), "one\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[probe] fixture: first")
	write(t, filepath.Join(repo, "f.txt"), "two\n")
	return repo, marker
}

// TestLocalInfoAttributesStartNoFilter proves that in local mode a filter named
// by .git/info/attributes never starts: the clone is refused by name and the
// driver leaves no marker. Online, the same clone is read and the driver runs,
// which shows the fixture can start it.
func TestLocalInfoAttributesStartNoFilter(t *testing.T) {
	repo, marker := infoAttributesClone(t)
	write(t, filepath.Join(repo, ".git", "info", "attributes"), "f.txt filter=probe\n")

	defer UseOffline(UseOffline(true))
	_, err := Between(repo, "HEAD", "")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("local mode started the configured filter while reading the working tree (%v)", err)
	}
	if err == nil || !strings.Contains(err.Error(), "info/attributes") || !strings.Contains(err.Error(), "next:") {
		t.Fatalf("local mode read a clone whose info/attributes names a filter: %v", err)
	}

	UseOffline(false)
	if _, err := Between(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("online, the fixture's filter did not start, so the local check proves nothing: %v", err)
	}
}

// TestLocalInfoAttributesOfCommentsOnly proves that an info/attributes file that
// sets nothing, only blank lines and comments, is not refused.
func TestLocalInfoAttributesOfCommentsOnly(t *testing.T) {
	repo, marker := infoAttributesClone(t)
	write(t, filepath.Join(repo, ".git", "info", "attributes"), "# git ls-files -s\n\n   # indented\n")

	defer UseOffline(UseOffline(true))
	if _, err := Between(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("the filter started (%v)", err)
	}
}
