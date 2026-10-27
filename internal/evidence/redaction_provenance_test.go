package evidence

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
)

// TestRedactionBeforeBuildFailsClosed records a known limit (N1, owed): the pipeline redacts a
// finding's text after the Judge decided, and the pack does not record which findings it redacted.
// So when a redaction moves the rule's outcome, Build refuses the pack rather than guess, and an
// honest pack is not written. It must never be written with an outcome the evidence contradicts.
// The repair records redaction provenance at the moment of replacement; until then this is the
// fail-closed behaviour. Mutated field: the message and the denying row, each redacted.
func TestRedactionBeforeBuildFailsClosed(t *testing.T) {
	in, _ := forgeryInput(t)
	secretA, secretB := strings.Repeat("A", 32), strings.Repeat("B", 32)
	rd := redact.New(map[string]string{"TOKEN_A": secretA, "TOKEN_B": secretB})
	f := &in.Result.Findings[2]
	f.Message = "pkg-a 1.2.0 is in an advisory; the lock file pins " + secretA
	f.Refutation = &review.Refutation{QueryRows: [][]string{{secretB, "no"}}}
	if judged, err := review.Rederive(*f); err != nil || !judged.Reported {
		t.Fatalf("the Judge's reading before redaction should keep F3 reported, got %+v, %v", judged, err)
	}
	f.Message = rd.Text(f.Message)
	f.Refutation.QueryRows = [][]string{{rd.Text(secretB), "no"}}
	in.Redactor = rd
	if _, err := Build(in); err == nil {
		t.Fatal("Build wrote a pack whose recorded outcome its own redacted evidence contradicts")
	}
}

// TestLiteralMarkerTextIsNotRedaction: text that only looks like a redaction marker, which no
// configured redactor produced, never excuses an outcome the evidence contradicts: Build refuses
// the pack, or Verify refuses it. Mutated field: the finding's message (marker text, no redactor).
func TestLiteralMarkerTextIsNotRedaction(t *testing.T) {
	in, pub := forgeryInput(t)
	f := &in.Result.Findings[2]
	f.Message = "pkg-a [redacted: UNCONFIGURED]"
	f.Refutation = &review.Refutation{QueryRows: [][]string{{f.Title, "no"}}}
	if o, err := review.Rederive(*f); err != nil || o.Reported {
		t.Fatalf("precondition: the recorded outcome must contradict the visible evidence: %+v %v", o, err)
	}
	pack, err := Build(in)
	if err != nil {
		return
	}
	if rep, err := Verify(pack, pub); err == nil {
		t.Fatalf("an outcome the evidence contradicts was accepted behind marker text: %d re-derived, %v named", rep.Rederived, rep.NotRederived)
	}
}
