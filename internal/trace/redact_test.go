package trace

import (
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
)

// TestAppendRedactsEveryTextFieldAndLeavesTheCallersEventAlone proves that a
// Writer with a Redactor writes no configured value in any field of an event,
// and that the event the caller holds is not changed.
// Mutated field: every string field of the event carries the secret.
func TestAppendRedactsEveryTextFieldAndLeavesTheCallersEventAlone(t *testing.T) {
	const secret = "secret-value-7c1e"
	w, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	w.SetRedactor(redact.New(map[string]string{"A_KEY": secret}))
	e := Event{
		FindingID: secret, Kind: secret, Agent: secret, Text: "t " + secret,
		ChangeKey: secret, RuleID: secret, Path: secret, Title: secret, Run: secret,
		Evidence: []Evidence{
			{Kind: secret, Test: &TestEvidence{Path: secret, Name: secret, Command: secret, Output: secret}},
			{Kind: "query", Query: &QueryEvidence{
				SQL: secret, Form: secret, Engine: secret, Snapshot: secret,
				Columns: []string{secret}, Rows: [][]string{{secret, "plain"}},
			}},
			{Kind: "tool", Tool: &ToolEvidence{Tool: secret, Version: secret, Command: secret, Output: secret, BaseOutput: secret}},
		},
		Refutation: &Refutation{Reason: secret},
		Guardrail:  &Check{Side: secret, Verdict: secret, Categories: []string{secret}},
		Model:      &Model{Route: secret, Model: secret, ReceiptID: secret},
	}
	if err := w.Append(e); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(w.Path())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), secret) {
		t.Errorf("the trace carries the secret:\n%s", b)
	}
	if !strings.Contains(string(b), "[redacted: A_KEY]") {
		t.Errorf("the trace does not name the variable:\n%s", b)
	}
	if e.Text != "t "+secret || e.Evidence[1].Query.Rows[0][0] != secret || e.Evidence[2].Tool.Output != secret || e.Guardrail.Categories[0] != secret {
		t.Error("the caller's event was changed")
	}
}

// TestAWriterWithoutARedactorWritesTheTextAsGiven proves that redaction is
// opt-in: sample mode holds no secret and its trace is unchanged.
func TestAWriterWithoutARedactorWritesTheTextAsGiven(t *testing.T) {
	w, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(Event{Kind: "claim", Text: "plain text value-123456"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(w.Path())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "plain text value-123456") {
		t.Errorf("the text changed:\n%s", b)
	}
}
