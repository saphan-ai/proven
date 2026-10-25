package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The tests of ADR-0036 through the command line: PROVEN_POLICY names the policy, an accepted finding is
// printed with its owner and does not make --error fail, an expired exception counts again and is
// named, and a weaker proven.yaml exits 7.

// policyException accepts one finding of the discount fixture: rule and path from the sample review.
func policyException(id, rule, where, expires string) string {
	p := "exception." + id + "."
	return p + "rule: " + rule + "\n" + p + "location: " + where + "\n" + p + "reason: reviewed by the platform team\n" +
		p + "owner: platform-owner\n" + p + "expires: " + expires + "\n"
}

// runWithPolicy scans the discount fixture in sample mode under the policy body.
func runWithPolicy(t *testing.T, policyBody string, extra ...string) (int, string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(path, []byte(policyBody), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	args := append([]string{"scan", "--sample", "--error"}, extra...)
	var stdout, stderr bytes.Buffer
	code := Run(append(args, target), envOf(map[string]string{"PROVEN_POLICY": path}), &stdout, &stderr, nil)
	return code, stdout.String(), stderr.String()
}

// TestAnAcceptedFindingIsPrintedWithItsOwnerAndDoesNotFailTheRun accepts all four findings of the
// fixture: the text names each with reason, owner and expiry, and --error exits 0.
func TestAnAcceptedFindingIsPrintedWithItsOwnerAndDoesNotFailTheRun(t *testing.T) {
	body := policyException("a", "discount-applied-twice", "billing/discount.go", "2999-12-31") +
		policyException("b", "known-exploited-dependency", "go.mod:5", "2999-12-31") +
		policyException("c", "known-exploited-cve", "go.mod", "2999-12-31") +
		policyException("d", "osv-affected-package", "go.mod", "2999-12-31")
	code, stdout, stderr := runWithPolicy(t, body)
	if code != exitOK {
		t.Fatalf("exit %d\nstderr: %s\nstdout: %s", code, stderr, stdout)
	}
	for _, want := range []string{
		"organisation policy: ",
		"accepted finding F1 ERROR",
		"accepted by exception a: reason: reviewed by the platform team; owner: platform-owner; expires: 2999-12-31",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout lacks %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "\nfinding ") {
		t.Errorf("an accepted finding is also printed as a finding:\n%s", stdout)
	}
}

// TestAnExpiredExceptionCountsAgainAndIsNamed: the same exception dated in the past fails --error and
// the text says it expired, with its date.
func TestAnExpiredExceptionCountsAgainAndIsNamed(t *testing.T) {
	code, stdout, stderr := runWithPolicy(t, policyException("a", "discount-applied-twice", "billing/discount.go", "2000-01-01"))
	if code != exitFindings {
		t.Fatalf("exit %d\nstderr: %s\nstdout: %s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "exception a expired on 2000-01-01: finding F1 counts again") || !strings.Contains(stdout, "finding F1 ERROR") {
		t.Errorf("stdout:\n%s", stdout)
	}
	if strings.Contains(stdout, "accepted finding") {
		t.Errorf("an expired exception still accepts:\n%s", stdout)
	}
}

// TestAWeakerProvenYamlExitsSevenNamingThePolicy: the repository turns off a class the policy doubles.
func TestAWeakerProvenYamlExitsSevenNamingThePolicy(t *testing.T) {
	local := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(local, []byte("doubled_classes: off\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runWithPolicy(t, "doubled_classes: security\n", "--config", local)
	if code != exitConfig {
		t.Fatalf("exit %d\nstderr: %s", code, stderr)
	}
	for _, want := range []string{"doubled_classes=", "weaker than the organisation policy", "policy.yaml", `"security"`} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q: %s", want, stderr)
		}
	}
}

// TestAMissingPolicyExitsSeven: a policy that is named and cannot be read stops the run on load.
func TestAMissingPolicyExitsSeven(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	missing := filepath.Join(t.TempDir(), "absent.yaml")
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{"PROVEN_POLICY": missing}), &stdout, &stderr, nil)
	if code != exitConfig || !strings.Contains(stderr.String(), "PROVEN_POLICY") {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
}
