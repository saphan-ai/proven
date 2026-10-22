package pipeline

import (
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

func risk(id, rule, title, path string, start, end int) candidate.Risk {
	return candidate.Risk{ID: id, RuleID: rule, Title: title, Path: path, StartLine: start, EndLine: end}
}

func TestDoubledRunsKeepDistinctRisksThatShareAnID(t *testing.T) {
	base := []candidate.Candidate{{
		Risk: risk("F1", "discount-applied-twice", "Discount is applied twice", "billing/discount.go", 5, 5),
		Runs: []string{runDefault}, Models: []string{"A"},
	}}
	extra := []candidate.Candidate{{
		Risk: risk("F1", "sql-injection", "SQL injection in report", "report/query.go", 12, 14),
		Runs: []string{runSpecial}, Models: []string{"B"},
	}}
	got, ids := mergeCandidates(base, extra)
	if len(got) != 2 {
		t.Fatalf("candidates %d, want 2 (a different risk with the same id was merged)", len(got))
	}
	if got[0].ID == got[1].ID {
		t.Fatalf("both candidates have id %q", got[0].ID)
	}
	if got[1].Title != "SQL injection in report" || len(got[1].Models) != 1 || got[1].Models[0] != "B" {
		t.Fatalf("special risk = %+v", got[1])
	}
	if len(got[0].Models) != 1 || got[0].Models[0] != "A" {
		t.Fatalf("default risk models = %v, want [A]", got[0].Models)
	}
	if ids[0] != got[1].ID {
		t.Fatalf("final id of the special risk = %q, candidate id %q", ids[0], got[1].ID)
	}
}

func TestDoubledRunsJoinTheSameRiskUnderOneFinding(t *testing.T) {
	base := []candidate.Candidate{{
		Risk: risk("F1", "discount-applied-twice", "Discount is applied twice", "billing/discount.go", 5, 5),
		Runs: []string{runDefault}, Models: []string{"A"},
	}}
	extra := []candidate.Candidate{{
		Risk: risk("F7", "discount-applied-twice", "Discount applied two times", "billing/discount.go", 5, 5),
		Runs: []string{runSpecial}, Models: []string{"B"},
	}}
	got, ids := mergeCandidates(base, extra)
	if len(got) != 1 {
		t.Fatalf("candidates %d, want 1 (the same risk from both runs)", len(got))
	}
	if len(got[0].Models) != 2 || got[0].Models[1] != "B" {
		t.Fatalf("models = %v, want [A B]", got[0].Models)
	}
	if ids[0] != "F1" {
		t.Fatalf("final id = %q, want F1", ids[0])
	}
}
