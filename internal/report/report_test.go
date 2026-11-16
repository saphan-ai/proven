package report

import "testing"

// The result and evidence rules can be tested without configuring the application.
func TestRecordedProofIsIndependentOfModelAndDatabase(t *testing.T) {
	f := Finding{Title: "discount regression", Evidence: []Evidence{{Kind: EvidenceTest, Test: &TestEvidence{FailsOnChange: true, PassesOnBase: true}}}, Refutation: &Refutation{QueryRows: [][]string{{"discount", "clean"}}}}
	out, err := Rederive(f)
	if err != nil || !out.Reported || out.Level != ProvedByTest || out.Unresolved {
		t.Fatalf("%+v %v", out, err)
	}
	f.Evidence[0].Test.PassesOnBase = false
	out, err = Rederive(f)
	if err != nil || out.Reported {
		t.Fatalf("non-differential test counted: %+v %v", out, err)
	}
}
