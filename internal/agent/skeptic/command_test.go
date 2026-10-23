package skeptic

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/testrun"
)

// textGen is a generator that always answers text.
type textGen struct{ text string }

func (g textGen) Complete(context.Context, string, string) (string, error) { return g.text, nil }

// TestAnswerThatCarriesACommandProposesNoTest proves that a refutation whose
// answer carries a command proposes no test to run: the test is dropped and the
// reason names the refusal. An answer with the file and the name alone keeps them.
func TestAnswerThatCarriesACommandProposesNoTest(t *testing.T) {
	risk := candidate.Candidate{Risk: candidate.Risk{ID: "F1", Title: "Discount is applied twice"}}
	carried, err := Attempt(context.Background(), risk, textGen{
		`{"refuted":true,"reason":"it passes","path":"billing/a_test.go","name":"TestA","command":"make proof","sql":""}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := carried.Refutation; got.TestPath != "" || got.TestName != "" || got.Reason != "refused: "+testrun.ReasonModelCommand {
		t.Fatalf("%#v", got)
	}
	plain, err := Attempt(context.Background(), risk, textGen{
		`{"refuted":true,"reason":"it passes","path":"billing/a_test.go","name":"TestA","sql":""}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := plain.Refutation; got.TestPath != "billing/a_test.go" || got.TestName != "TestA" || got.Reason != "it passes" {
		t.Fatalf("%#v", got)
	}
}
