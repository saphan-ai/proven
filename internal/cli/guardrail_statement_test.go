package cli

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/review"
)

// Property: the text output of a review run with the guardrail off says so, and
// the text output of one run with it on does not.
// Mutated field: the Result's GuardrailStatement.
func TestTextSaysWhenTheGuardrailWasOff(t *testing.T) {
	off, err := render("text", review.Result{Class: "backend", ClassRule: "fallback", GuardrailStatement: candidate.GuardrailOffStatement})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(off), candidate.GuardrailOffStatement+"\n") {
		t.Fatalf("the text output does not say the guardrail was off:\n%s", off)
	}
	on, err := render("text", review.Result{Class: "backend", ClassRule: "fallback"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(on), "no guardrail") {
		t.Fatalf("negative control: the text output of a review with the guardrail on mentions it:\n%s", on)
	}
}
