package sarifschema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain points every cache root os.UserCacheDir reads at a temp directory, as cmd/proven's does: the
// tests here run cli.Run, which writes its trace under the user cache, and a write to the real one changes
// the state internal/cli's own TestMain compares, so the two packages failed together under the full suite.
// Go 1.26.10 os/file.go: Unix reads $XDG_CACHE_HOME or else $HOME/.cache, Darwin $HOME/Library/Caches,
// Windows %LocalAppData%. A proven directory that changes in the real cache fails this suite.
func TestMain(m *testing.M) {
	real, err := os.UserCacheDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	before, err := cacheProvenSnapshot(real)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var temps []string
	for _, name := range []string{"HOME", "XDG_CACHE_HOME", "LocalAppData"} {
		dir, err := os.MkdirTemp("", "proven-sarifschema-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		temps = append(temps, dir)
		if err := os.Setenv(name, dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	code := m.Run()
	after, snapErr := cacheProvenSnapshot(real)
	for _, dir := range temps {
		os.RemoveAll(dir)
	}
	if snapErr != nil {
		fmt.Fprintln(os.Stderr, snapErr)
		os.Exit(1)
	}
	if before != after {
		fmt.Fprintf(os.Stderr, "assertOSCacheUntouched: real user cache changed\n")
		os.Exit(1)
	}
	os.Exit(code)
}

// cacheProvenSnapshot is cmd/proven's: the names, sizes and times under root/proven, or absent.
func cacheProvenSnapshot(root string) (string, error) {
	dir := filepath.Join(root, "proven")
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return "absent", nil
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "not-a-directory", nil
	}
	var b strings.Builder
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		st, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s %d %s\n", rel, st.Size(), st.Mode())
		return nil
	})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
