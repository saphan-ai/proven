package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/cli"
)

func TestScanSamplePrintsF1AndF2NotF3(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := cli.Run(
		[]string{"scan", "--sample", filepath.Join(root, "testdata", "fixture", "discount")},
		shellEnv(t), &stdout, &stderr, nil,
	)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	assertDiscountText(t, stdout.String())
}

func TestScanSampleErrorExits1(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := cli.Run(
		[]string{"scan", "--sample", "--error", filepath.Join(root, "testdata", "fixture", "discount")},
		shellEnv(t), &stdout, &stderr, nil,
	)
	if code != 1 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	assertDiscountText(t, stdout.String())
}

func TestScanWithoutSampleIsRefused(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"scan", "somewhere"}, func(k string) string {
		if k == "METERLESS_RUNTIME_TOKEN" {
			return "test-token"
		}
		return ""
	}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("exit %d\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "somewhere/base") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestScanExcludeAllPrintsNoFinding(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := cli.Run(
		[]string{"scan", "--sample", "--exclude=*", filepath.Join(root, "testdata", "fixture", "discount")},
		shellEnv(t), &stdout, &stderr, nil,
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
	code := cli.Run(
		[]string{"scan", "--sample", "--baseline-commit=does-not-exist", filepath.Join(root, "testdata", "fixture", "discount")},
		shellEnv(t), &stdout, &stderr, nil,
	)
	if code != 2 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "unknown flag") || !strings.Contains(stderr.String(), "does-not-exist") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func assertDiscountText(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{
		"finding F1 ERROR Discount is applied twice",
		"evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true",
		"finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list",
		"evidence query local-sample snapshot=sample-2026-10-22",
		"row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true",
		"finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list",
		"SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'",
		"row: CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known",
		"finding osv:evil.example/leftpad ERROR Dependency evil.example/leftpad v1.2.3 is in an OSV affected range",
		"introduced 1.0.0 <= v1.2.3 < fixed 1.2.4",
		"SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'",
		"row: GHSA-px-sample-0001 | CVE-2026-0001 | Go | evil.example/leftpad | 1.0.0 | 1.2.4 |  | HIGH | 2026-01-15",
		"alias CVE-2026-0001 is on the KEV list",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "finding F3") {
		t.Fatalf("F3 was printed\n%s", body)
	}
	if strings.Contains(body, "example.com/unused") {
		t.Fatalf("an unrelated package was queried\n%s", body)
	}
}

// shellEnv returns empty for the variables the shell reads while resolving
// ADR-0020, and fails the test on any other key. The reviewer receives the
// same function and does not read it.
func shellEnv(t *testing.T) func(string) string {
	t.Helper()
	allowed := map[string]struct{}{
		"PROVEN_OUTPUT":           {},
		"PROVEN_OUTPUT_FORMAT":    {},
		"PROVEN_ERROR":            {},
		"PROVEN_BASELINE_COMMIT":  {},
		"SEMGREP_BASELINE_COMMIT": {},
		"SEMGREP_BASELINE_REF":    {},
		"PROVEN_CONFIG":           {},
		"PROVEN_INCLUDE":          {},
		"PROVEN_EXCLUDE":          {},
		"PROVEN_SEVERITY":         {},
		"PROVEN_TIMEOUT":          {},
		"SAMPLE_MODE":             {},
	}
	return func(k string) string {
		if _, ok := allowed[k]; !ok {
			t.Fatalf("unexpected environment key %s", k)
		}
		return ""
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
