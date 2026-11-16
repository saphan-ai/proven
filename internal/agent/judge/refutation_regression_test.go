package judge_test

import (
	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
	"testing"
)

func TestUnrelatedDenialCannotSuppressDifferentialProof(t *testing.T) {
	c := candidate.Candidate{Risk: candidate.Risk{Title: "discount regression"}, Evidence: []candidate.Evidence{{Kind: candidate.KindTest, Test: &candidate.TestEvidence{FailsOnChange: true, PassesOnBase: true}}}}
	for _, rows := range [][][]string{{{"discount", "unreviewed"}, {"unrelated-component", "clean"}}, {{"discount", "clean"}}} {
		c.Refutation.QueryRows = rows
		if got := judge.Decide(c); got.Dropped {
			t.Fatalf("observed proof suppressed: %s", got.DropReason)
		}
	}
}

func TestQueryRefutationMustReferToTheSameRow(t *testing.T) {
	c := candidate.Candidate{Risk: candidate.Risk{Title: "leftpad issue"}, Evidence: []candidate.Evidence{{Kind: candidate.KindQuery, Query: &candidate.QueryEvidence{Bound: true, Rows: [][]string{{"leftpad", "true"}}}}}}
	c.Refutation.QueryRows = [][]string{{"leftpad", "pending"}, {"other", "clean"}}
	if got := judge.Decide(c); got.Dropped {
		t.Fatal(got.DropReason)
	}
	c.Refutation.QueryRows = [][]string{{"leftpad", "false"}}
	if got := judge.Decide(c); !got.Dropped {
		t.Fatal("same-row refutation ignored")
	}
}
