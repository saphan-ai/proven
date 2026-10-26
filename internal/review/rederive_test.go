package review

import "testing"

func testEv(fails, passes bool) Evidence {
	return Evidence{Kind: EvidenceTest, Test: &TestEvidence{FailsOnChange: fails, PassesOnBase: passes}}
}

func queryEv(bound bool, rows [][]string) Evidence {
	return Evidence{Kind: EvidenceQuery, Query: &QueryEvidence{Bound: bound, Rows: rows}}
}

func toolEv(onChange, onBase bool) Evidence {
	return Evidence{Kind: EvidenceTool, Tool: &ToolEvidence{PresentOnChange: onChange, PresentOnBase: onBase}}
}

// Rederive gives the Judge's answer for what a finding records, and each row changes exactly the
// one recorded fact its name says.
func TestRederiveFollowsTheJudgesRule(t *testing.T) {
	rows := [][]string{{"pkg", "1.0"}}
	cases := []struct {
		name string
		f    Finding
		want Outcome
	}{
		{"test fails on change and passes on base", Finding{Evidence: []Evidence{testEv(true, true)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByTest, Reported: true, Complete: true}},
		{"test does not fail on the change", Finding{Evidence: []Evidence{testEv(false, true)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByTest, Reported: false, Complete: true}},
		{"test fails on the base too", Finding{Evidence: []Evidence{testEv(true, false)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByTest, Reported: false, Complete: true}},
		{"tool result on the change only", Finding{Evidence: []Evidence{toolEv(true, false)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByTool, Reported: true, Complete: true}},
		{"tool result on both sides", Finding{Evidence: []Evidence{toolEv(true, true)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByTool, Reported: false, Complete: true}},
		{"bound rows", Finding{Evidence: []Evidence{queryEv(true, rows)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByData, Reported: true, Complete: true}},
		{"rows not bound to the change", Finding{Evidence: []Evidence{queryEv(false, rows)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByData, Reported: false, Complete: true}},
		{"no rows", Finding{Evidence: []Evidence{queryEv(true, nil)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByData, Reported: false, Complete: true}},
		{"the strongest kind sets the level", Finding{Evidence: []Evidence{queryEv(true, rows), testEv(true, true)}, Refutation: &Refutation{}},
			Outcome{Level: ProvedByTest, Reported: true, Complete: true}},
		{"a refuting test passes on the change", Finding{Evidence: []Evidence{testEv(true, true)}, Refutation: &Refutation{TestFound: true}},
			Outcome{Level: ProvedByTest, Reported: false, Complete: true}},
		{"the refuting test could not run", Finding{Evidence: []Evidence{testEv(true, true)}, Refutation: &Refutation{TestCannotRun: true}},
			Outcome{Level: ProvedByTest, Reported: true, Unresolved: true, Complete: true}},
		{"no refutation recorded", Finding{Evidence: []Evidence{testEv(true, true)}},
			Outcome{Level: ProvedByTest, Reported: true}},
	}
	for _, c := range cases {
		got, err := Rederive(c.f)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestRederiveRefusesEvidenceOfNoKnownKind(t *testing.T) {
	if _, err := Rederive(Finding{ID: "F1", Evidence: []Evidence{{Kind: "guess"}}}); err == nil {
		t.Fatal("evidence of no known kind has no level of proof")
	}
}

func TestCountProofsHasEveryLevelAsAKey(t *testing.T) {
	got := CountProofs([]Finding{{ProofLevel: ProvedByTest}, {ProofLevel: ProvedByTest}, {ProofLevel: ProvedByData}})
	want := map[ProofLevel]int{ProvedByTest: 2, ProvedByTool: 0, ProvedByData: 1, AgreedNotProven: 0}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for l, n := range want {
		if got[l] != n {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
