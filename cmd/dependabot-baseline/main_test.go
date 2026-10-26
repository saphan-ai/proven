package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The repository, the alerts and the advisory below are synthetic: made up
// for this test, naming no real package or advisory.

const lockFile = `{"name":"f","version":"0.0.0","lockfileVersion":3,"requires":true,"packages":{"":{"name":"f","version":"0.0.0","dependencies":{"fixture-a":"1.2.0"}},"node_modules/fixture-a":{"version":"1.2.0"}}}
`

const advisory = `{"id":"GHSA-fix-aaaa-0001","aliases":["CVE-2099-0001"],"published":"2099-01-01T00:00:00Z",
"affected":[{"package":{"ecosystem":"npm","name":"fixture-a"},"ranges":[{"type":"SEMVER","events":[{"introduced":"1.0.0"},{"fixed":"1.3.0"}]}]}]}`

type fixture struct {
	repo, alerts, out, work, advisories string
}

// newFixture makes a git repository with one lock file, an alerts file pinned
// to its commit, and a directory holding one advisory.
func newFixture(t *testing.T) fixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH; CI's go test job has it")
	}
	root := t.TempDir()
	fx := fixture{
		repo: filepath.Join(root, "repo"), alerts: filepath.Join(root, "alerts.json"), out: filepath.Join(root, "out"),
		work: filepath.Join(root, "work"), advisories: filepath.Join(root, "advisories"),
	}
	for path, body := range map[string]string{
		filepath.Join(fx.repo, "web", "package-lock.json"): lockFile,
		filepath.Join(fx.advisories, "a.json"):             advisory,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git := func(args ...string) []byte {
		out, err := exec.Command("git", append([]string{"-C", fx.repo}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return out
	}
	git("init", "-q")
	git("add", ".")
	git("-c", "user.name=t", "-c", "user.email=t@example.org", "commit", "-q", "-m", "[baseline-test] repo: a fixture commit")
	commit := strings.TrimSpace(string(git("rev-parse", "HEAD")))
	ref := map[string]any{
		"repository": "example/fixture", "pinned_commit": commit,
		"alerts": []map[string]any{{
			"number": 1, "ecosystem": "npm", "package": "fixture-a", "manifest": "web/package-lock.json",
			"ghsa": "GHSA-fix-aaaa-0001", "cve": "CVE-2099-0001", "vulnerable_range": ">= 1.0.0, < 1.3.0", "patched": "1.3.0",
		}},
	}
	b, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fx.alerts, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return fx
}

func (fx fixture) args(extra ...string) []string {
	return append([]string{"-repo", fx.repo, "-alerts", fx.alerts, "-out", fx.out, "-work", fx.work}, extra...)
}

func envOf(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func (fx fixture) wrote() bool {
	_, err := os.Stat(filepath.Join(fx.out, "REPORT.md"))
	return err == nil
}

func TestRunRefusesWhenTheDefaultStoreIsNotConfigured(t *testing.T) {
	fx := newFixture(t)
	var stdout, stderr bytes.Buffer
	if code := run(fx.args(), envOf(nil), &stdout, &stderr); code != 2 {
		t.Fatalf("exit %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Zetaris store is not configured") || !strings.Contains(stderr.String(), "-manifest-only") {
		t.Errorf("the refusal does not say what was measured and what to do: %s", stderr.String())
	}
	if fx.wrote() {
		t.Error("a refused run wrote a report")
	}
}

// Mutated field: -evidence-label (and, in the second case, -snapshot-date).
// Without them a local run has no verdict line that says which snapshot it rests on.
func TestLocalStoreRefusesWithoutALabelAndADate(t *testing.T) {
	fx := newFixture(t)
	env := envOf(map[string]string{"PROVEN_EVIDENCE_STORE": "local", "PROVEN_EVIDENCE_FILE": filepath.Join(fx.work, "evidence.sql")})
	for name, extra := range map[string][]string{
		"no label":   {"-snapshot-date", "2099-01-06", "-advisories", fx.advisories},
		"no date":    {"-evidence-label", "fixture snapshot", "-advisories", fx.advisories},
		"bad date":   {"-evidence-label", "fixture snapshot", "-snapshot-date", "06/01/2099", "-advisories", fx.advisories},
		"label only": {"-evidence-label", "fixture snapshot"},
	} {
		var stderr bytes.Buffer
		if code := run(fx.args(extra...), env, &bytes.Buffer{}, &stderr); code != 2 {
			t.Errorf("%s: exit %d, want 2", name, code)
		}
		if fx.wrote() {
			t.Errorf("%s: a refused run wrote a report", name)
		}
	}
}

func TestManifestOnlyRunWritesBothFiles(t *testing.T) {
	fx := newFixture(t)
	var stdout, stderr bytes.Buffer
	if code := run(fx.args("-manifest-only"), envOf(nil), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	report, err := os.ReadFile(filepath.Join(fx.out, "REPORT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(report), "**not checked**") || !strings.Contains(string(report), "evidence: none") {
		t.Errorf("the manifest-only report does not say it asked no store:\n%s", report)
	}
	if _, err := os.Stat(filepath.Join(fx.out, "proven-findings.json")); err != nil {
		t.Error(err)
	}
}

// A local run builds its evidence file from the advisories, names its
// snapshot in every verdict line, and gives the same bytes twice.
func TestLocalRunIsRepeatableAndNamesItsEvidence(t *testing.T) {
	fx := newFixture(t)
	env := envOf(map[string]string{"PROVEN_EVIDENCE_STORE": "local", "PROVEN_EVIDENCE_FILE": filepath.Join(fx.work, "evidence.sql")})
	args := fx.args("-evidence-label", "local OSV snapshot 2099-01-06", "-snapshot-date", "2099-01-06", "-advisories", fx.advisories)
	var first [2][]byte
	for round := 0; round < 2; round++ {
		var stderr bytes.Buffer
		if code := run(args, env, &bytes.Buffer{}, &stderr); code != 0 {
			t.Fatalf("round %d: exit %d: %s", round, code, stderr.String())
		}
		for i, name := range []string{"REPORT.md", "proven-findings.json"} {
			got, err := os.ReadFile(filepath.Join(fx.out, name))
			if err != nil {
				t.Fatal(err)
			}
			if round == 0 {
				first[i] = got
			} else if !bytes.Equal(first[i], got) {
				t.Errorf("%s differs between two runs", name)
			}
		}
	}
	report := string(first[0])
	if !strings.Contains(report, "**found**") || !strings.Contains(report, "ev-") {
		t.Errorf("the fixture alert is not found with an evidence id:\n%s", report)
	}
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, "**found**") && !strings.Contains(line, "local OSV snapshot 2099-01-06") {
			t.Errorf("a verdict line does not carry its evidence: %s", line)
		}
	}
}
