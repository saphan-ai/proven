package investigator

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/service"
)

// scriptGen is a generator that always returns text.
type scriptGen struct{ text string }

func (s scriptGen) Complete(context.Context, string, string) (string, error) {
	return s.text, nil
}

// TestNilQueryIsANamedRefusal proves Investigate refuses a nil query adapter
// with the same named failure Execute returns, instead of panicking (F2).
func TestNilQueryIsANamedRefusal(t *testing.T) {
	_, err := Investigate(context.Background(), candidate.Candidate{
		Risk: candidate.Risk{ID: "F1", Title: "Discount is applied twice", Path: "billing/discount.go"},
	}, "", scriptGen{text: `{"sql":"SELECT 1"}`}, nil)
	d, ok := service.As(err)
	if !ok || d.Service != "query" || d.Reason != "no query engine" {
		t.Fatalf("err %v", err)
	}
}
