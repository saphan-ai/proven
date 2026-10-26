package review

import (
	"testing"
	"time"
)

func TestAcceptedExceptionDoesNotCountInProofTotals(t *testing.T) {
	finding := Finding{ID: "F1", RuleID: "rule-one", Location: Location{Path: "a.go", StartLine: 2}}
	level := ProvedByTest
	result := Result{Findings: []Finding{finding}, ProofCounts: map[ProofLevel]int{level: 1}}
	policy := &policy{exceptions: []exception{{ID: "approved", Rule: "rule-one", Path: "a.go", Reason: "triaged", Owner: "team", Expires: "2999-12-31"}}}
	policy.applyExceptions(&result, time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC))
	if len(result.Accepted) != 1 || len(result.Findings) != 0 {
		t.Fatalf("accepted=%d active=%d", len(result.Accepted), len(result.Findings))
	}
	if result.ProofCounts[level] != 0 {
		t.Fatalf("accepted finding still counts in proof totals: %d", result.ProofCounts[level])
	}
}
