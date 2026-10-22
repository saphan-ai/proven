package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanSamplePrintsF1AndF2NotF3(t *testing.T) {
	root := repoRoot(t)
	wd := t.TempDir()
	t.Chdir(wd)
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", filepath.Join(root, "testdata", "fixture", "discount")},
		&stdout, &stderr, os.Getwd, func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	body := stdout.String()
	for _, want := range []string{
		"finding F1 ERROR Discount is applied twice",
		"evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true",
		"finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list",
		"evidence query local-sample snapshot=sample-2026-10-22",
		"row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "finding F3") {
		t.Fatalf("F3 was printed\n%s", body)
	}
}

func TestScanExcludeAllPrintsNoFinding(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", "--exclude=*", filepath.Join(root, "testdata", "fixture", "discount")},
		&stdout, &stderr, os.Getwd, os.Getenv,
	)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "finding ") {
		t.Fatalf("finding printed\n%s", stdout.String())
	}
}

func TestScanMissingBaselineNamesTheCommit(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", "--baseline-commit=does-not-exist", filepath.Join(root, "testdata", "fixture", "discount")},
		&stdout, &stderr, os.Getwd, os.Getenv,
	)
	if code != 2 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "unknown flag") || !strings.Contains(stderr.String(), "does-not-exist") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestScanWithoutSampleIsRefused(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"scan", "somewhere"}, &stdout, &stderr, os.Getwd, os.Getenv)
	if code != 2 {
		t.Fatalf("exit %d", code)
	}
}

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
