package judge_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
)

func TestDropsWhenThereIsNoEvidence(t *testing.T) {
	got := judge.Decide(candidate.Candidate{Risk: candidate.Risk{ID: "F3"}})
	if !got.Dropped || got.DropReason != "dropped: no evidence" {
		t.Fatalf("%#v", got)
	}
}

func TestKeepsATestThatFailsOnTheChange(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "F1"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTest,
			Test: &candidate.TestEvidence{FailsOnChange: true, PassesOnBase: true},
		}},
	})
	if got.Dropped || got.DropReason != "kept" {
		t.Fatalf("%#v", got)
	}
}

func TestDropsATestThatDoesNotPassOnTheBase(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "F4"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTest,
			Test: &candidate.TestEvidence{FailsOnChange: true, PassesOnBase: false},
		}},
	})
	if !got.Dropped || got.DropReason != "dropped: no evidence" {
		t.Fatalf("%#v", got)
	}
}

func TestDropsAQueryWithNoRows(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "F3"},
		Evidence: []candidate.Evidence{{
			Kind:  candidate.KindQuery,
			Query: &candidate.QueryEvidence{Rows: nil},
		}},
	})
	if !got.Dropped {
		t.Fatal("empty query is not evidence")
	}
}
