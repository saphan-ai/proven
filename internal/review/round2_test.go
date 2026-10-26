package review

import "testing"

// Refutation rows that name the finding and deny it drop it; the finding's own text is what the
// rows are matched against.
func TestRefutationRowsThatDenyTheFindingDropIt(t *testing.T) {
	f := Finding{ID: "F3", Title: "vulnerable pkg-a", Message: "pkg-a 1.2.0 is in an advisory",
		Evidence:   []Evidence{{Kind: EvidenceQuery, Query: &QueryEvidence{Rows: [][]string{{"pkg-a", "1.2.1"}}, Bound: true}}},
		Refutation: &Refutation{QueryRows: [][]string{{"pkg-a", "no"}}}}
	got, err := Rederive(f)
	if err != nil || got.Reported {
		t.Fatalf("rows naming pkg-a and saying no must drop F3: %+v %v", got, err)
	}
}

// A refuting test that was found and fails on the change does not refute.
func TestARefutingTestThatFailsOnTheChangeDoesNotDrop(t *testing.T) {
	f := Finding{ID: "F1",
		Evidence:   []Evidence{{Kind: EvidenceTest, Test: &TestEvidence{FailsOnChange: true, PassesOnBase: true}}},
		Refutation: &Refutation{TestFound: true, TestFailsOnChange: true}}
	got, err := Rederive(f)
	if err != nil || !got.Reported || got.Unresolved {
		t.Fatalf("a refuting test that fails on the change must leave F1 reported: %+v %v", got, err)
	}
}
