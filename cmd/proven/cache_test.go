package main

import (
	"fmt"
	"os"
	"testing"
)

// TestMain points the default trace directory at a temp directory. The
// command's default is the user cache, and these tests run the real reviewer.
// Go 1.26 reads XDG_CACHE_HOME for that directory on this OS.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "proven-cmd-cache-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Setenv("XDG_CACHE_HOME", dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
