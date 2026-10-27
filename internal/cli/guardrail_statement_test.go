package cli

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/review"
)

// checkedResult is a review whose guardrail made 2 prompt checks and 1 answer
// check, blocked 1, and could not make 0 (or unanswered, when it is given).
func checkedResult(unanswered int) review.Result {
	return review.Result{Class: "backend", ClassRule: "fallback", Guardrail: review.GuardrailCount{
		Prompts: 2, Answers: 1, Blocked: 1, Unanswered: unanswered,
		Backend: "nvidia", Model: "nvidia/test-guard-model",
	}}
}

// Property: the one guardrail line has the same words in the text output and in
// the Markdown foot, with the counts of the Result and the model id and backend
// that answered (ADR-0005).
// Mutated fields: the counts of the Result's Guardrail.
func TestGuardrailLineIsTheSameInTextAndMarkdown(t *testing.T) {
	const want = "Guardrail: nvidia/test-guard-model (nvidia), checked 2 prompts and 1 answers, 1 blocked"
	for _, format := range []string{"text", "md"} {
		out, err := render(format, checkedResult(0))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), want) {
			t.Errorf("%s output lacks the line %q:\n%s", format, want, out)
		}
	}
	// A check that could not be made is named, never folded into the counts.
	const wantU = want + ", 3 not checked"
	for _, format := range []string{"text", "md"} {
		out, err := render(format, checkedResult(3))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), wantU) {
			t.Errorf("%s output lacks the line %q:\n%s", format, wantU, out)
		}
	}
}

// Property: a review run with the guardrail off says so in every format, with
// the same words, and carries zero counts in JSON.
// Mutated field: the Result's GuardrailStatement.
func TestEveryFormatSaysWhenTheGuardrailWasOff(t *testing.T) {
	const want = "Guardrail: off for this review"
	off := review.Result{Class: "backend", ClassRule: "fallback", GuardrailStatement: candidate.GuardrailOffStatement}
	for _, format := range []string{"text", "md", "json", "sarif"} {
		out, err := render(format, off)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), want) {
			t.Errorf("%s output does not say the guardrail was off:\n%s", format, out)
		}
	}
	doc := decodeJSONMap(t, jsonBytes(t, off))["guardrail"].(map[string]any)
	if doc["state"] != "off" || doc["prompts"] != float64(0) || doc["answers"] != float64(0) || doc["blocked"] != float64(0) {
		t.Errorf("guardrail object of a review with the guardrail off = %v", doc)
	}
}

// Property: a review that made no model call says there was nothing to check,
// and the same Result with checks does not say it (negative control).
// Mutated field: the Result's Guardrail counts.
func TestGuardrailLineSaysWhenThereWasNoModelCall(t *testing.T) {
	const want = "Guardrail: no model call to check"
	none, err := render("text", review.Result{Class: "backend", ClassRule: "fallback"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(none), want+"\n") {
		t.Fatalf("a review with no check does not say so:\n%s", none)
	}
	some, err := render("text", checkedResult(0))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(some), want) {
		t.Fatalf("negative control: a review with checks says there was nothing to check:\n%s", some)
	}
}
