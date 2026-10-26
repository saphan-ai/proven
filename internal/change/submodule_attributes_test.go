package change

import (
	"os"
	"path/filepath"
	"testing"
)

// submoduleClone is a clone with a checked-out submodule "sub" whose f.txt is
// changed in the working tree, and a filter driver "probe", configured in the
// user's ~/.gitconfig, that leaves a marker when git starts it. The submodule's
// own info/attributes names the filter.
func submoduleClone(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	marker := filepath.Join(t.TempDir(), "driver-ran")
	script := filepath.Join(home, "driver.sh")
	write(t, script, "#!/bin/sh\ntouch '"+marker+"'\nexec cat\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, ".gitconfig"), "[filter \"probe\"]\n\tclean = "+script+"\n[protocol \"file\"]\n\tallow = always\n")
	t.Setenv("HOME", home)
	sub := t.TempDir()
	git(t, sub, "init")
	write(t, filepath.Join(sub, "f.txt"), "one\n")
	git(t, sub, "add", ".")
	git(t, sub, "commit", "-m", "[probe] fixture: sub")
	super := t.TempDir()
	git(t, super, "init")
	write(t, filepath.Join(super, "a.txt"), "a\n")
	git(t, super, "add", ".")
	git(t, super, "commit", "-m", "[probe] fixture: first")
	git(t, super, "submodule", "add", sub, "sub")
	git(t, super, "commit", "-m", "[probe] fixture: add sub")
	write(t, filepath.Join(super, ".git", "modules", "sub", "info", "attributes"), "f.txt filter=probe\n")
	write(t, filepath.Join(super, "sub", "f.txt"), "two\n")
	_ = os.Remove(marker)
	return super, marker
}

// TestLocalSubmoduleStartsNoFilter proves that in local mode reading the working
// tree never enters a checked-out submodule, so a filter its own info/attributes
// names does not start. Online, the same clone starts it, which shows the
// fixture can.
func TestLocalSubmoduleStartsNoFilter(t *testing.T) {
	super, marker := submoduleClone(t)
	defer UseOffline(UseOffline(true))
	if _, err := Between(super, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("local mode started a filter named by the submodule's info/attributes (%v)", err)
	}
	UseOffline(false)
	if _, err := Between(super, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("online, the fixture's filter did not start, so the local check proves nothing: %v", err)
	}
}
