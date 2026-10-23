package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
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
		"review class: security",
		"review rule: go.mod",
		"author model not stated; default reviewer NVIDIA",
		"reviewer NVIDIA; role default; family NVIDIA; author model not stated",
		"special-review model: not configured",
		"finding F1 ERROR Discount is applied twice",
		"evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true",
		"finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list",
		"evidence query local-sample snapshot=sample-2026-10-22",
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
	// ADR-0032: F2's query evidence is the osv-by-package template, built by
	// Proven, so it carries the OSV rows of the package and no free-text query.
	f2 := body[strings.Index(body, "finding F2"):strings.Index(body, "finding F4")]
	for _, want := range []string{
		"SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'",
		"row: GHSA-px-sample-0002 |  | Go | evil.example/leftpad | 2.0.0 | 2.1.0 |  | LOW | 2026-03-01",
	} {
		if !strings.Contains(f2, want) {
			t.Fatalf("F2 section is missing %q\n%s", want, f2)
		}
	}
	const phrase = "EPSS 0.50000 (90th percentile, 2026-10-20)"
	for _, header := range []string{
		"finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list",
		"finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list",
	} {
		block := findingSection(body, header)
		if !strings.Contains(block, phrase) {
			t.Fatalf("%s has no EPSS score\n%s", header, block)
		}
	}
	if strings.Contains(findingSection(body, "finding F1 ERROR Discount is applied twice"), "EPSS ") {
		t.Fatal("F1 named an EPSS score")
	}
	if strings.Contains(body, "example.com/unused") {
		t.Fatalf("an unrelated package was queried\n%s", body)
	}
}

func findingSection(body, header string) string {
	i := strings.Index(body, header)
	if i < 0 {
		return ""
	}
	rest := body[i+len(header):]
	cut := len(rest)
	if n := strings.Index(rest, "\nfinding "); n >= 0 && n < cut {
		cut = n
	}
	if n := strings.Index(rest, "\nfailed services:"); n >= 0 && n < cut {
		cut = n
	}
	return header + rest[:cut]
}

func TestScanOpenFamilyNameIsAccepted(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(cfg, []byte("families: gemini-*=Google\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", "--config=" + cfg, filepath.Join(root, "testdata", "fixture", "discount")},
		&stdout, &stderr, os.Getwd, os.Getenv,
	)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "unknown family") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestScanMalformedFamilyExits7(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(cfg, []byte("families: gemini-*\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", "--config=" + cfg, filepath.Join(root, "testdata", "fixture", "discount")},
		&stdout, &stderr, os.Getwd, os.Getenv,
	)
	if code != 7 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "families") || !strings.Contains(stderr.String(), "pattern=value") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestScanUnknownClassExits7(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(cfg, []byte("path_rules: **/*.go=frontend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", "--config=" + cfg, filepath.Join(root, "testdata", "fixture", "discount")},
		&stdout, &stderr, os.Getwd, os.Getenv,
	)
	if code != 7 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "path_rules") || !strings.Contains(stderr.String(), "frontend") || !strings.Contains(stderr.String(), "unknown class") {
		t.Fatalf("stderr %q", stderr.String())
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

func TestSampleOutputPrintsUnresolved(t *testing.T) {
	var buf bytes.Buffer
	printResult(&buf, review.Result{
		Findings: []review.Finding{{
			ID: "F9", Severity: review.SeverityWarning, Title: "maybe",
			Unresolved: true,
			Location:   review.Location{Path: "a.go", StartLine: 1, EndLine: 1},
			Evidence: []review.Evidence{{
				Kind: review.EvidenceTest,
				Test: &review.TestEvidence{Path: "a_test.go", Name: "TestA", FailsOnChange: true, PassesOnBase: true},
			}},
		}},
		FailedServices: []review.ServiceFailure{},
	})
	if !strings.Contains(buf.String(), "finding F9 WARNING maybe") || !strings.Contains(buf.String(), "  unresolved\n") {
		t.Fatal(buf.String())
	}
}

func TestSampleOutputPrintsTheSameFamilySentence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trace.jsonl")
	sentence := "skeptic family nemotron is the same as the finder family nemotron (nvidia/nemotron-3.5-super-vl-120b-a12b and nvidia/nemotron-3.5-super-vl-120b-a12b)"
	line := "{\"finding_id\":\"review\",\"kind\":\"family\",\"agent\":\"skeptic\",\"text\":\"" + sentence + "\"}\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	printResult(&buf, review.Result{TracePath: path, Findings: []review.Finding{}, FailedServices: []review.ServiceFailure{}})
	if !strings.Contains(buf.String(), sentence) {
		t.Fatal(buf.String())
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
