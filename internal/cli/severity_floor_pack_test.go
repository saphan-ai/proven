package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/review"
)

// A review's own pack must verify whatever severity floor the scan applied. The floor removes
// findings from the result after the review counted them; the proof counts the pack records must
// be the counts of the findings it records. A fix recounts after the filter (or filters before the
// counts are made), so verify accepts the pack and re-derives every finding it holds.
func TestSeverityFloorPackVerifies(t *testing.T) {
	proved := func(id string, sev review.Severity) review.Finding {
		return review.Finding{
			ID: id, RuleID: id + ".rule", Title: id, Message: id, Severity: sev,
			Location:   review.Location{Path: "a.go", StartLine: 1, EndLine: 1},
			ProofLevel: review.ProvedByTest,
			Evidence: []review.Evidence{{Kind: review.EvidenceTest, Test: &review.TestEvidence{
				Path: "a_test.go", Name: "T" + id, Command: "go test", FailsOnChange: true, PassesOnBase: true}}},
			Refutation: &review.Refutation{},
		}
	}
	findings := []review.Finding{proved("F1", review.SeverityWarning), proved("F2", review.SeverityError)}
	rev := &captureReviewer{res: review.Result{Findings: findings, ProofCounts: review.CountProofs(findings)}}

	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	priv, pub := packKeys(t)
	pack := filepath.Join(t.TempDir(), "review.pack")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--severity=ERROR", "--evidence-pack", pack, target},
		packEnv(map[string]string{evidence.EnvPackKey: priv}), &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("scan: exit %d\n%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"verify", pack, "--public-key", pub}, packEnv(nil), &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("verify of the review's own pack under --severity=ERROR: exit %d\n%s%s", code, stdout.String(), stderr.String())
	}
}
