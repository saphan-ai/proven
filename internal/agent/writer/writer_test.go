package writer

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// Order within one severity: highest EPSS first, then findings with no
// score in the order they already had. A severity keeps the place where
// it first appeared. A higher score does not move a finding into another
// severity.
func TestOrderWithinSeverity(t *testing.T) {
	in := []candidate.Candidate{
		scored("A", "WARNING", "0.10"),
		plain("B", "ERROR"),
		scored("C", "ERROR", "0.20"),
		scored("D", "ERROR", "0.90"),
		plain("E", "WARNING"),
		scored("F", "ERROR", "0.20"),
	}
	got := ids(Write(Input{Candidates: in}).Findings)
	want := "A,E,D,C,F,B"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	// The input order is unchanged.
	if ids(in) != "A,B,C,D,E,F" {
		t.Fatalf("input moved: %s", ids(in))
	}
}

// Findings with equal EPSS scores keep the order they came in.
func TestEqualScoresKeepTheirOrder(t *testing.T) {
	in := []candidate.Candidate{
		plain("F1", "ERROR"),
		scored("F2", "ERROR", "0.50000"),
		scored("F4", "ERROR", "0.50000"),
		scored("osv", "ERROR", "0.50000"),
	}
	if got := ids(Write(Input{Candidates: in}).Findings); got != "F2,F4,osv,F1" {
		t.Fatalf("%s", got)
	}
}

// A finding the Judge dropped does not reach the draft.
func TestDroppedFindingsAreNotOrderedIn(t *testing.T) {
	in := []candidate.Candidate{
		scored("F2", "ERROR", "0.2"),
		{Risk: candidate.Risk{ID: "F3", Severity: "ERROR"}, Dropped: true},
	}
	got := Write(Input{Candidates: in}).Findings
	if len(got) != 1 || got[0].ID != "F2" {
		t.Fatalf("%#v", got)
	}
}

// scored is a finding with query rows that carry an EPSS score.
func scored(id, sev, score string) candidate.Candidate {
	return candidate.Candidate{
		Risk: candidate.Risk{ID: id, Severity: sev},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{
				Columns: []string{"cve_id", "epss", "percentile", "score_date"},
				Rows:    [][]string{{"CVE-2026-0001", score, "0.50", "2026-10-20"}},
			},
		}},
	}
}

// plain is a finding with no EPSS score.
func plain(id, sev string) candidate.Candidate {
	return candidate.Candidate{Risk: candidate.Risk{ID: id, Severity: sev}}
}

// ids lists the findings' ids in order.
func ids(in []candidate.Candidate) string {
	parts := make([]string, len(in))
	for i, c := range in {
		parts[i] = c.ID
	}
	return strings.Join(parts, ",")
}
