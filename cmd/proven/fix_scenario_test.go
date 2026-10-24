package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/cli"
)

// scanWith runs `proven scan --sample` on a fixture with extra arguments and returns stdout and the trace.
// Sample mode reads no setting, so any variable the loader does not know fails the test.
func scanWith(t *testing.T, name string, extra ...string) (string, []byte) {
	t.Helper()
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	args := append([]string{"scan", "--sample"}, extra...)
	args = append(args, filepath.Join(root, "testdata", "fixture", name))
	code := cli.Run(args, func(k string) string {
		if cli.KnownCLIEnv(k) {
			return ""
		}
		t.Fatalf("sample mode read %s", k)
		return ""
	}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, stdout.String(), stderr.String())
	}
	body := stdout.String()
	i := strings.LastIndex(body, "trace: ")
	if i < 0 {
		t.Fatalf("no trace path\n%s", body)
	}
	trace, err := os.ReadFile(strings.TrimSpace(body[i+len("trace: "):]))
	if err != nil {
		t.Fatal(err)
	}
	return body, trace
}

// TestFixScenarioShowsOneProvedFixAndOneFailedAttempt proves ADR-0042 end to end in sample mode: F1's
// patch is shown as a fix with its explanation and the commands of its proof; F2's first patch fails the
// proving test, its retry fixes that test and breaks a neighbour, so the review says the fix is not
// proved after two attempts and shows neither patch.
func TestFixScenarioShowsOneProvedFixAndOneFailedAttempt(t *testing.T) {
	body, trace := scanWith(t, "fix")
	f1, f2, ok := strings.Cut(body, "finding F2 ")
	if !ok {
		t.Fatalf("F2 was not printed\n%s", body)
	}
	for _, want := range []string{
		"fix proved after 1 attempt: a suggestion for you to apply; Proven changes nothing",
		"fix changes: Subtract the discount once.",
		"fix closes the problem because: Discounted took the percent off twice",
		"fix edit billing/discount.go",
		"  - return cents - cents*percent/100 - cents*percent/100",
		"  + return cents - cents*percent/100",
		"fix run base: go test -json -run '^TestDiscounted$' ./billing exit=0 passed",
		"fix run change: go test -json -run '^TestDiscounted$' ./billing exit=1 failed",
		"fix run change+patch: go test -json -run '^TestDiscounted$' ./billing exit=0 passed",
		"fix run package on change+patch: go test -json ./billing exit=0 passed",
	} {
		if !strings.Contains(f1, want) {
			t.Fatalf("F1 is missing %q\n%s", want, f1)
		}
	}
	if !strings.Contains(f2, "fix not proved after 2 attempts: the patch breaks another test of the package") {
		t.Fatalf("F2 does not say the attempt failed\n%s", f2)
	}
	// ADR-0042 point 3: a patch that was not proved is not shown as a fix.
	for _, hidden := range []string{"fix edit", "return 50", "return 500", "fix proved"} {
		if strings.Contains(f2, hidden) {
			t.Fatalf("F2 shows %q although its fix was not proved\n%s", hidden, f2)
		}
	}
	// The record of the last attempt stays, so a reader sees which run failed and its exit code.
	if !strings.Contains(f2, "fix run package on change+patch: go test -json ./shipping exit=1 failed") {
		t.Fatalf("F2 does not show the failed package run\n%s", f2)
	}
	for _, want := range []string{"fix proved after 1 attempt(s)", "fix not proved after 2 attempts", "run change+patch: go test"} {
		if !bytes.Contains(trace, []byte(want)) {
			t.Fatalf("the trace is missing %q", want)
		}
	}
}

// TestFixesOffProposesNoFix proves the setting that turns proposed fixes off (ADR-0042): the same
// scenario with `fixes: off` prints the findings and no fix line of any kind.
func TestFixesOffProposesNoFix(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(cfg, []byte("fixes: off\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, _ := scanWith(t, "fix", "--config", cfg)
	if !strings.Contains(body, "finding F1") || !strings.Contains(body, "finding F2") {
		t.Fatalf("the findings are gone\n%s", body)
	}
	if strings.Contains(body, "fix ") {
		t.Fatalf("a fix was shown with fixes off\n%s", body)
	}
}

// TestOneAttemptSettingEndsTheFailedFixAtOnce proves fix_attempts bounds the retries: with one
// attempt, F2's first patch is the only try, and the review says so.
func TestOneAttemptSettingEndsTheFailedFixAtOnce(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(cfg, []byte("fix_attempts: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, _ := scanWith(t, "fix", "--config", cfg)
	if !strings.Contains(body, "fix not proved after 1 attempt: the proving test does not pass on the change with the patch applied") {
		t.Fatalf("%s", body)
	}
}

// TestADependencyFindingNamesTheSmallestFixedVersion proves ADR-0042 point 5 end to end: the OSV
// scenario's finding names the fixed version of the advisory row that holds the version in use,
// shows that row, and says no test ran.
func TestADependencyFindingNamesTheSmallestFixedVersion(t *testing.T) {
	body, _ := scanScenario(t, "osv-go-module")
	for _, want := range []string{
		"fix named from the advisory table, not run by a test: raise example.com/widget v1.4.0 to 1.5.0",
		"fix advisory GHSA-w19-sample-0001 columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published",
		"fix advisory row: GHSA-w19-sample-0001 | CVE-2026-4242 | Go | example.com/widget | 1.3.0 | 1.5.0 |  | HIGH | 2026-04-01",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}
