package evidence

import (
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestAWrongOutcomeWithNoMarkerIsRefusedAtBuild: redaction_moved lists a finding only for redaction.
// A Result with no redaction marker in any text the rule reads, whose recorded outcome the rule
// contradicts only through that text, is a wrong Result: Build must refuse it, not list it and name
// it. Mutated field: holdsMarker (made to return true, Build writes the pack and names F3).
func TestAWrongOutcomeWithNoMarkerIsRefusedAtBuild(t *testing.T) {
	in, _ := forgeryInput(t)
	f := &in.Result.Findings[2]
	f.Refutation = &review.Refutation{QueryRows: [][]string{{f.Title, "no"}}}
	if o, err := review.Rederive(*f); err != nil || o.Reported {
		t.Fatalf("the sample does not set up a text-only contradiction: %+v, %v", o, err)
	}
	if holdsMarker(*f) {
		t.Fatal("the sample holds a redaction marker, so it proves nothing about holdsMarker")
	}
	if _, err := Build(in); err == nil {
		t.Fatal("Build wrote a pack that records F3 reported while its own text and rows drop it, with no redaction to explain it")
	}
}
