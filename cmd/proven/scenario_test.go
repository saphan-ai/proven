package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefutedScenarioDropsTheFinding(t *testing.T) {
	body, trace := scanScenario(t, "refuted")
	if strings.Contains(body, "finding F1") {
		t.Fatalf("F1 was printed\n%s", body)
	}
	const why = "dropped: refuted: test TestTotal passes on the change"
	if !bytes.Contains(trace, []byte(why)) {
		t.Fatalf("trace missing %s\n%s", why, trace)
	}
	if !strings.Contains(body, "failed services: none") {
		t.Fatalf("stdout\n%s", body)
	}
}

func TestUnsafePromptStopsAndNamesCategories(t *testing.T) {
	body, trace := scanScenario(t, "unsafe-prompt")
	const reason = "failed service: guardrail: unsafe: Criminal Planning/Confessions, Other"
	if !strings.Contains(body, reason) {
		t.Fatalf("stdout\n%s", body)
	}
	if strings.Contains(body, "\nfinding ") {
		t.Fatalf("a finding was printed\n%s", body)
	}
	for _, cat := range []string{"Criminal Planning/Confessions", "Other"} {
		if !bytes.Contains(trace, []byte(cat)) {
			t.Fatalf("trace missing %s\n%s", cat, trace)
		}
	}
	if !bytes.Contains(trace, []byte("categories=")) {
		t.Fatalf("trace does not name the categories\n%s", trace)
	}
}

func TestOSVGoModuleShowsTheRangeAndTheEPSSScore(t *testing.T) {
	body, _ := scanScenario(t, "osv-go-module")
	for _, want := range []string{
		"finding osv:example.com/widget ERROR Dependency example.com/widget v1.4.0 is in an OSV affected range",
		"EPSS 0.42000 (91st percentile, 2026-10-21)",
		"go.mod:5-5",
		"introduced 1.3.0 <= v1.4.0 < fixed 1.5.0",
		"GHSA-w19-sample-0002 does not match: v1.4.0 < introduced 2.0.0",
		"alias CVE-2026-4242 is on the KEV list",
		"SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'example.com/widget'",
		"row: GHSA-w19-sample-0001 | CVE-2026-4242 | Go | example.com/widget | 1.3.0 | 1.5.0 |  | HIGH | 2026-04-01",
		"row: GHSA-w19-sample-0002 |  | Go | example.com/widget | 2.0.0 | 2.1.0 |  | LOW | 2026-05-01",
		"SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-4242'",
		"row: CVE-2026-4242 | 0.42000 | 0.91000 | 2026-10-21",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "example.com/other") {
		t.Fatalf("an unrelated package was queried\n%s", body)
	}
}

func TestUnknownAuthorSaysTheDefaultReviewer(t *testing.T) {
	body, _ := scanScenario(t, "unknown-author")
	for _, want := range []string{
		"author model not stated; default reviewer NVIDIA",
		"reviewer NVIDIA; role default; family NVIDIA; author model not stated",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "from commit") || strings.Contains(body, "from description") || strings.Contains(body, "\nfinding ") {
		t.Fatalf("stdout\n%s", body)
	}
}

func scanScenario(t *testing.T, name string) (string, []byte) {
	t.Helper()
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"scan", "--sample", filepath.Join(root, "testdata", "fixture", name)},
		&stdout, &stderr, os.Getwd, func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	body := stdout.String()
	const mark = "trace: "
	i := strings.LastIndex(body, mark)
	if i < 0 {
		t.Fatalf("no trace path\n%s", body)
	}
	path := strings.TrimSpace(body[i+len(mark):])
	trace, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body, trace
}
