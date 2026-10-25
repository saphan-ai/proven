package main

import (
	"bytes"
	"github.com/saphan-ai/proven/internal/store"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/cli"
	"github.com/saphan-ai/proven/internal/source"
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
		"Reachability: not measured",
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

// TestOSVPHPComposerReportsTheDependency proves the Packagist adapter
// reports the raised Composer package from composer.lock, inside one
// sample range and outside the other. Dropping Composer and ComposerLock
// from manifest.Readers is what turns this test red.
func TestOSVPHPComposerReportsTheDependency(t *testing.T) {
	body, _ := scanScenario(t, "osv-php-composer")
	for _, want := range []string{
		"review class: security",
		"review rule: composer.json",
		"finding osv:monolog/monolog ERROR Dependency monolog/monolog 2.9.1 is in an OSV affected range",
		"EPSS 0.25000 (91st percentile, 2026-10-24)",
		"composer.lock:5-5",
		"introduced 2.4.0 <= 2.9.1 < fixed 2.10.0",
		"GHSA-php-sample-0002 does not match: 2.9.1 < introduced 3.0.0",
		"alias CVE-2026-9090 is on the KEV list",
		"Reachability: not measured",
		"ecosystem = 'Packagist' AND package = 'monolog/monolog'",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "other/unused") {
		t.Fatalf("an unrelated package was queried\n%s", body)
	}
}

// TestOSVKotlinGradleReportsTheDependency proves the Maven adapter reports
// the raised Gradle package from gradle.lockfile, inside one sample range
// and outside the other. build.gradle.kts writes the same coordinate.
// preferLock keeps the lock file when both files in one directory name the
// package. The catalog in gradle/ does not raise it. Making
// mavenver.Compare report every pair as equal turns this test red. Taking
// GradleKts and GradleLock out of Readers together also turns it red.
// Taking only GradleLock out does not: the literal in build.gradle.kts is
// still a raise.
func TestOSVKotlinGradleReportsTheDependency(t *testing.T) {
	body, _ := scanScenario(t, "osv-kotlin-gradle")
	for _, want := range []string{
		"review class: security",
		"review rule: build.gradle.kts",
		"finding osv:com.example:widget ERROR Dependency com.example:widget 1.4.0 is in an OSV affected range",
		"EPSS 0.27000 (91st percentile, 2026-10-25)",
		"gradle.lockfile:5-5",
		"introduced 1.3.0 <= 1.4.0 < fixed 1.5.0",
		"GHSA-kt-sample-0002 does not match: 1.4.0 < introduced 2.0.0",
		"alias CVE-2026-9291 is on the KEV list",
		"Reachability: not measured",
		"ecosystem = 'Maven' AND package = 'com.example:widget'",
		"fix named from the advisory table, not run by a test: raise com.example:widget 1.4.0 to 1.5.0",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "com.example:unused") || strings.Contains(body, "Dependency com.example:quiet") {
		t.Fatalf("a package that was not raised was queried\n%s", body)
	}
}

// TestOSVRustCargoReportsTheDependency proves the crates.io adapter reports
// the raised Cargo package from Cargo.lock, inside one sample range and
// outside the other. The manifest writes a constraint; the lock file's
// version is the one reported. Dropping Cargo and CargoLock from
// manifest.Readers is what turns this test red.
func TestOSVRustCargoReportsTheDependency(t *testing.T) {
	body, _ := scanScenario(t, "osv-rust-cargo")
	for _, want := range []string{
		"review class: security",
		"review rule: Cargo.toml",
		"finding osv:regex ERROR Dependency regex 1.9.1 is in an OSV affected range",
		"EPSS 0.33000 (91st percentile, 2026-10-24)",
		"Cargo.lock:15-15",
		"introduced 1.7.0 <= 1.9.1 < fixed 1.10.0",
		"GHSA-rs-sample-0002 does not match: 1.9.1 < introduced 1.11.0",
		"alias CVE-2026-9191 is on the KEV list",
		"Reachability: not measured",
		"ecosystem = 'crates.io' AND package = 'regex'",
		"fix named from the advisory table, not run by a test: raise regex 1.9.1 to 1.10.0",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "unused_crate") || strings.Contains(body, "Dependency libc") {
		t.Fatalf("a package that was not raised was queried\n%s", body)
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

// TestMaliciousPackageReportsTheAdvisory proves a raised package with a
// MAL- row is a finding. An empty row list is not evidence: the mutated
// line is the empty-list return in malAdapter.Rule, in
// internal/querytpl/adr0039.go. Making that return accept the empty list
// turns the check below red.
func TestMaliciousPackageReportsTheAdvisory(t *testing.T) {
	body, _ := scanScenario(t, "malicious-package")
	for _, want := range []string{
		"finding F1 ERROR Dependency evil.example/drop was published to do harm",
		"SELECT id, ecosystem, package, summary FROM px_evidence.px_mal_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/drop'",
		"row: MAL-2024-1001 | Go | evil.example/drop | published to take credentials",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	ad, ok := source.ByID("mal")
	if !ok {
		t.Fatal("mal adapter is not registered")
	}
	if _, ruled := ad.Rule(source.Evidence{SQL: "SELECT 1", Rows: nil}); ruled {
		t.Fatal("an empty row list was evidence")
	}
}

// TestProjectHealthReportsNoMaintainers proves a raised package whose
// maintained score is 0 is a finding. The mutated field is that score:
// a maintained, signed, old release is not this finding.
func TestProjectHealthReportsNoMaintainers(t *testing.T) {
	body, _ := scanScenario(t, "project-health")
	for _, want := range []string{
		"finding F1 ERROR Dependency evil.example/young has no maintainers",
		"SELECT ecosystem, package, maintained, signed, release_age_days FROM px_evidence.px_scorecard WHERE ecosystem = 'Go' AND package = 'evil.example/young'",
		"row: Go | evil.example/young | 0 | true | 400",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}

// TestCVSSScoreSitsBesideEPSS proves a CVE the change is affected by shows
// its CVSS row and the EPSS score of the same CVE on one finding. The
// mutated field is the CVE id.
func TestCVSSScoreSitsBesideEPSS(t *testing.T) {
	body, _ := scanScenario(t, "cvss-score")
	for _, want := range []string{
		"finding F1 ERROR CVE-2026-7777 severity is critical",
		"EPSS 0.15000 (80th percentile, 2026-10-25)",
		"SELECT cve_id, cvss_score, severity, source FROM px_evidence.px_cvss_cve WHERE cve_id = 'CVE-2026-7777'",
		"row: CVE-2026-7777 | 9.8 | CRITICAL | nvd",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}

// TestSPDXLicenceReportsADisallowedLicence proves a raised package whose
// licence is outside the permissive set is a finding. The mutated field is
// the licence cell.
func TestSPDXLicenceReportsADisallowedLicence(t *testing.T) {
	body, _ := scanScenario(t, "spdx-licence")
	for _, want := range []string{
		"finding F1 ERROR Dependency evil.example/copyleft has a licence the policy does not allow",
		"SELECT ecosystem, package, licence FROM px_evidence.px_spdx_lic WHERE ecosystem = 'Go' AND package = 'evil.example/copyleft'",
		"row: Go | evil.example/copyleft | GPL-3.0-only",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}

func scanScenario(t *testing.T, name string) (string, []byte) {
	t.Helper()
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	// cli.Run loads the variables ADR-0020 §1 documents (cli.KnownCLIEnv:
	// PROVEN_OUTPUT_FORMAT, PROVEN_OUTPUT, PROVEN_ERROR, PROVEN_BASELINE_COMMIT,
	// SEMGREP_BASELINE_COMMIT, SEMGREP_BASELINE_REF, PROVEN_CONFIG, PROVEN_INCLUDE,
	// PROVEN_EXCLUDE, PROVEN_SEVERITY, PROVEN_TIMEOUT, SAMPLE_MODE, and
	// GITHUB_EVENT_PATH, GITHUB_BASE_REF, GITHUB_SHA). An empty answer keeps the
	// default. Any other key, including an adapter secret, fails the test.
	// W19 failed on every key; that predicate cannot survive the loader.
	code := cli.Run(
		[]string{"scan", "--sample", filepath.Join(root, "testdata", "fixture", name)},
		func(k string) string {
			// ADR-0045: the OTEL_* family is the one sample mode reads besides the shell's own keys.
			// ADR-0047 point 7: the evidence-record database and link base are read in sample mode too.
			if cli.KnownCLIEnv(k) || strings.HasPrefix(k, "OTEL_") || k == store.EnvRecordsDB || k == store.EnvEvidenceLink {
				return ""
			}
			t.Fatalf("sample mode read %s", k)
			return ""
		}, &stdout, &stderr, nil,
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
