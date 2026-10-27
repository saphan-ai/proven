package review

import "fmt"

// GuardrailCount is what the guardrail did in one review: the prompts and the
// answers it checked, how many of those it blocked (an unsafe verdict that
// stopped a call), and how many checks it could not make. Backend and Model
// name the guardrail that answered. It has the fields of candidate.GuardrailCount.
type GuardrailCount struct {
	Prompts, Answers, Blocked, Unanswered int
	Backend, Model                        string
}

// The states of a review's guardrail, as JSON and SARIF name them.
const (
	GuardrailChecked = "checked"
	GuardrailOff     = "off"
	GuardrailNoCall  = "no_call"
)

// GuardrailState is GuardrailOff when the review ran without the guardrail,
// GuardrailNoCall when it ran with it and made no check (a sample run with no
// model call), and GuardrailChecked otherwise.
func (r Result) GuardrailState() string {
	switch {
	case r.GuardrailStatement != "":
		return GuardrailOff
	case r.Guardrail.Prompts+r.Guardrail.Answers+r.Guardrail.Unanswered == 0:
		return GuardrailNoCall
	}
	return GuardrailChecked
}

// GuardrailLine is the one line every output carries about NVIDIA's guardrail
// (ADR-0005): the same words in text, Markdown, JSON and SARIF, built from the
// counts the pipeline measured and never from a typed number. The model id and
// the backend are a service's words, so a format escapes the whole line for
// its own medium before it prints it.
func (r Result) GuardrailLine() string {
	switch r.GuardrailState() {
	case GuardrailOff:
		return "Guardrail: off for this review"
	case GuardrailNoCall:
		return "Guardrail: no model call to check"
	}
	c := r.Guardrail
	who := c.Model
	if who == "" {
		who = "model not named"
	}
	if c.Backend != "" {
		who += " (" + c.Backend + ")"
	}
	line := fmt.Sprintf("Guardrail: %s, checked %d prompts and %d answers, %d blocked", who, c.Prompts, c.Answers, c.Blocked)
	if c.Unanswered > 0 {
		line += fmt.Sprintf(", %d not checked", c.Unanswered)
	}
	return line
}
