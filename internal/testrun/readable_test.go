package testrun

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMakeReadableGrantsReadAndSearchAndNoWrite proves the grant reaches every
// directory and file under the root, lets others read and search, and adds no
// write bit for group or others.
func TestMakeReadableGrantsReadAndSearchAndNoWrite(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "pkg")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(sub, "a.go")
	tool := filepath.Join(sub, "run.sh")
	if err := os.WriteFile(plain, []byte("package a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := MakeReadable(root); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{
		root:  0o755,
		sub:   0o755,
		plain: 0o644,
		tool:  0o755,
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s mode %o, want %o", path, got, want)
		}
	}
}

// TestStagedCheckoutIsReadableByOthers proves the property the locked-down
// container depends on: the checkout a written test is staged into can be
// read and searched by a user that does not own it (mode bits for others), and
// is not writable by others. The mutated field is the staged root's mode: with
// MakeReadable removed from stageWritten it is 0700 and this test fails.
func TestStagedCheckoutIsReadableByOthers(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "go.mod"), []byte("module m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dst, err := stageWritten(src, "m_test.go", "package m\n")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst)
	for _, path := range []string{dst, filepath.Join(dst, "go.mod"), filepath.Join(dst, "m_test.go")} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		perm := info.Mode().Perm()
		if perm&0o004 == 0 {
			t.Errorf("%s mode %o is not readable by others", path, perm)
		}
		if info.IsDir() && perm&0o001 == 0 {
			t.Errorf("%s mode %o cannot be searched by others", path, perm)
		}
		if perm&0o022 != 0 {
			t.Errorf("%s mode %o is writable by group or others", path, perm)
		}
	}
}
