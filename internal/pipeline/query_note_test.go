package pipeline

import (
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// TestEvidenceTextNamesTheInvestigatorsRefusal proves ADR-0032 point 2 reaches
// the trace: a finding whose query was refused says so in its evidence line,
// alone or beside the Prover's note, and a finding with neither still says
// "none". The mutated field is the candidate's QueryNote.
func TestEvidenceTextNamesTheInvestigatorsRefusal(t *testing.T) {
	refused := "refused: query template is not on the list"
	cases := []struct {
		name string
		c    candidate.Candidate
		want string
	}{
		{"query refusal alone", candidate.Candidate{QueryNote: refused}, refused},
		{"both notes", candidate.Candidate{QueryNote: refused, ProverNote: "the test could not run"}, refused + "; the test could not run"},
		{"prover note alone", candidate.Candidate{ProverNote: "the test could not run"}, "the test could not run"},
		{"neither", candidate.Candidate{}, "none"},
	}
	for _, c := range cases {
		if got := evidenceText(c.c); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}
