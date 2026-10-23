package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunWritesTheSampleCSV proves that the command writes the sample
// advisories to the named file, aliases joined, and nothing to stdout.
func TestRunWritesTheSampleCSV(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.csv")
	var stdout, stderr bytes.Buffer
	code := run([]string{filepath.Join(repoRoot(t), "testdata", "osv"), out}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d %s", code, stderr.String())
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "example.com/widget") || !strings.Contains(string(b), "CVE-2024-1111 GHSA-zz") {
		t.Fatalf("%s", b)
	}
	if stdout.Len() != 0 {
		t.Fatal(stdout.String())
	}
}

// TestRunUsage proves that a call with no arguments exits with status 2 and
// prints the usage line.
func TestRunUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("%d %s", code, stderr.String())
	}
}

// repoRoot finds the module root, where testdata/ lives.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
