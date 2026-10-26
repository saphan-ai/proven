package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEveryOSCacheRuleIsATemporaryDirectory proves the suite redirected every
// cache root os.UserCacheDir reads. Go 1.26.10 os/file.go: Unix uses
// $XDG_CACHE_HOME or else $HOME/.cache, Darwin uses $HOME/Library/Caches,
// and Windows uses %LocalAppData%. A rule that still names the real home
// writes the real user cache on that OS and the suite stays green.
// Mutated field: HOME, XDG_CACHE_HOME, and LocalAppData.
func TestEveryOSCacheRuleIsATemporaryDirectory(t *testing.T) {
	temp := os.TempDir()
	if !strings.HasSuffix(temp, string(filepath.Separator)) {
		temp += string(filepath.Separator)
	}
	for _, goos := range []string{"linux", "darwin", "windows"} {
		dir, err := hostCacheDir(goos)
		if err != nil {
			t.Errorf("%s: %v", goos, err)
			continue
		}
		if !strings.HasPrefix(dir, temp) {
			t.Errorf("%s cache rule %s is not under the temp directory %s", goos, dir, os.TempDir())
		}
	}
}

// hostCacheDir is os.UserCacheDir for one GOOS, copied from Go 1.26.10
// os/file.go so a linux test can see the Darwin and Windows rules.
func hostCacheDir(goos string) (string, error) {
	switch goos {
	case "windows":
		dir := os.Getenv("LocalAppData")
		if dir == "" {
			return "", fmt.Errorf("%%LocalAppData%% is not defined")
		}
		return dir, nil
	case "darwin", "ios":
		dir := os.Getenv("HOME")
		if dir == "" {
			return "", fmt.Errorf("$HOME is not defined")
		}
		return dir + "/Library/Caches", nil
	default:
		dir := os.Getenv("XDG_CACHE_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", fmt.Errorf("neither $XDG_CACHE_HOME nor $HOME are defined")
			}
			dir += "/.cache"
		} else if !filepath.IsAbs(dir) {
			return "", fmt.Errorf("path in $XDG_CACHE_HOME is relative")
		}
		return dir, nil
	}
}

// TestMain points every cache root os.UserCacheDir reads at a temp directory.
// Go 1.26.10 os/file.go: Unix reads $XDG_CACHE_HOME or else $HOME/.cache,
// Darwin reads $HOME/Library/Caches, and Windows reads %LocalAppData%.
// The real user cache is read before those variables change. A proven
// directory that appears there fails the suite.
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
	home, err := os.MkdirTemp("", "proven-cmd-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	xdg, err := os.MkdirTemp("", "proven-cmd-xdg-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	local, err := os.MkdirTemp("", "proven-cmd-local-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Setenv("HOME", home); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Setenv("XDG_CACHE_HOME", xdg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Setenv("LocalAppData", local); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	after, snapErr := cacheProvenSnapshot(real)
	os.RemoveAll(home)
	os.RemoveAll(xdg)
	os.RemoveAll(local)
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
