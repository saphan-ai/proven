package judge_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
)

// A finding with no evidence is dropped.
func TestDropsWhenThereIsNoEvidence(t *testing.T) {
	got := judge.Decide(candidate.Candidate{Risk: candidate.Risk{ID: "F3"}})
	if !got.Dropped || got.DropReason != "dropped: no evidence" {
		t.Fatalf("%#v", got)
	}
}

// A test that fails on the change and passes on the base keeps the finding.
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

// A tool result on the change and not on the base keeps the finding (ADR-0040).
func TestKeepsAToolResultTheBaseDidNotHave(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "R1"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTool,
			Tool: &candidate.ToolEvidence{Tool: "go-race", PresentOnChange: true, PresentOnBase: false},
		}},
	})
	if got.Dropped || got.DropReason != "kept" {
		t.Fatalf("%#v", got)
	}
}

// A tool result that is also on the base is not evidence of the change (ADR-0040).
// Mutated field: ToolEvidence.PresentOnBase.
func TestDropsAToolResultTheBaseAlreadyHad(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "R2"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTool,
			Tool: &candidate.ToolEvidence{Tool: "go-race", PresentOnChange: true, PresentOnBase: true},
		}},
	})
	if !got.Dropped || got.DropReason != "dropped: no evidence" {
		t.Fatalf("%#v", got)
	}
}

// A test that fails on the base too is not evidence.
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

// A query with no rows is not evidence.
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

// Query rows Proven bound to the change keep the finding.
func TestKeepsQueryRowsBoundToTheChange(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{
			ID:      "F2",
			Title:   "Dependency leftpad v1.2.3 is on the known-exploited list",
			Message: "The change requires evil.example/leftpad v1.2.3.",
		},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{Bound: true, Rows: [][]string{
				{"evil.example/leftpad", "1.2.3", "CVE-2026-0001", "true"},
			}},
		}},
	})
	if got.Dropped || got.Unresolved || got.DropReason != "kept" {
		t.Fatalf("%#v", got)
	}
}

// Property: rows that repeat words of the finding's own text are not evidence
// unless Proven bound them to the change. Mutated field: Bound.
func TestDropsQueryRowsThatOnlyRepeatTheFindingsWords(t *testing.T) {
	c := candidate.Candidate{
		Risk: candidate.Risk{
			ID:      "F2",
			Title:   "Dependency leftpad v1.2.3 is on the known-exploited list",
			Message: "The change requires evil.example/leftpad v1.2.3.",
		},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{Rows: [][]string{
				{"evil.example/leftpad", "1.2.3", "CVE-2026-0001", "true"},
			}},
		}},
	}
	if got := judge.Decide(c); !got.Dropped || got.DropReason != "dropped: no evidence" {
		t.Fatalf("%#v", got)
	}
	c.Evidence[0].Query.Bound = true
	if got := judge.Decide(c); got.Dropped {
		t.Fatalf("negative control: the same rows, bound, must keep the finding: %#v", got)
	}
}

// Rows that do not name the finding are not its evidence.
func TestDropsAQueryWhoseRowsSayNothingAboutTheFinding(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{
			ID:    "F2",
			Title: "Dependency leftpad v1.2.3 is on the known-exploited list",
		},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{Rows: [][]string{
				{"unrelated.example/other", "9.9.9"},
			}},
		}},
	})
	if !got.Dropped || got.DropReason != "dropped: no evidence" {
		t.Fatalf("%#v", got)
	}
}

// A refuting test that passes on the change drops the finding.
func TestDropsWhenTheRefutingTestPassesOnTheChange(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "F1"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTest,
			Test: &candidate.TestEvidence{FailsOnChange: true, PassesOnBase: true},
		}},
		Refutation: candidate.Refutation{
			Attempted: true,
			TestFound: true,
			TestName:  "TestRefute",
			Reason:    "test TestRefute passes on the change",
		},
	})
	if !got.Dropped || got.Unresolved || got.DropReason != "dropped: refuted: test TestRefute passes on the change" {
		t.Fatalf("%#v", got)
	}
}

// A refutation without evidence, even one the model calls refuted, does not
// drop a finding that has evidence: the model's flag is not a vote.
func TestKeepsWhenTheRefutationHasNoEvidence(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "F1"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTest,
			Test: &candidate.TestEvidence{FailsOnChange: true, PassesOnBase: true},
		}},
		Refutation: candidate.Refutation{Attempted: true, Refuted: true, Reason: "the model disagrees"},
	})
	if got.Dropped || got.Unresolved || got.DropReason != "kept" {
		t.Fatalf("%#v", got)
	}
}

// A refuting test that cannot run leaves the finding kept and marked unresolved.
func TestUnresolvedWhenTheRefutingTestCannotRun(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{ID: "F1"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTest,
			Test: &candidate.TestEvidence{FailsOnChange: true, PassesOnBase: true},
		}},
		Refutation: candidate.Refutation{
			Attempted:     true,
			TestName:      "TestMissing",
			TestCannotRun: true,
			Reason:        "the refuting test could not run",
		},
	})
	if got.Dropped || !got.Unresolved || got.DropReason != "unresolved: the refuting test could not run" {
		t.Fatalf("%#v", got)
	}
}

// Refuting rows that name the finding and deny it drop the finding.
func TestDropsWhenQueryRowsContradict(t *testing.T) {
	got := judge.Decide(candidate.Candidate{
		Risk: candidate.Risk{
			ID:      "F2",
			Title:   "Dependency leftpad v1.2.3 is on the known-exploited list",
			Message: "The change requires evil.example/leftpad v1.2.3.",
		},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{Rows: [][]string{
				{"evil.example/leftpad", "1.2.3", "true"},
			}},
		}},
		Refutation: candidate.Refutation{
			Attempted: true,
			QueryRows: [][]string{{"evil.example/leftpad", "false"}},
		},
	})
	if !got.Dropped || got.DropReason != "dropped: refuted: query rows contradict the finding" {
		t.Fatalf("%#v", got)
	}
}
