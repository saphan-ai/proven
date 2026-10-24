package review

import (
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// Property: a draft from a review run with the guardrail off gives a Result
// that says so; a draft from a review with it on gives a Result that does not.
// Mutated field: the draft's GuardrailStatement.
func TestResultCarriesTheGuardrailOffStatement(t *testing.T) {
	off, err := fill(candidate.Draft{GuardrailStatement: candidate.GuardrailOffStatement})
	if err != nil {
		t.Fatal(err)
	}
	if off.GuardrailStatement != candidate.GuardrailOffStatement {
		t.Fatalf("the Result of a review with the guardrail off carries no statement: %q", off.GuardrailStatement)
	}
	on, err := fill(candidate.Draft{})
	if err != nil {
		t.Fatal(err)
	}
	if on.GuardrailStatement != "" {
		t.Fatalf("negative control: the Result of a review with the guardrail on says %q", on.GuardrailStatement)
	}
}
