package generator_test

import (
	"context"
	"errors"
	"testing"

	gen "github.com/saphan-ai/proven/examples/generator"
	igen "github.com/saphan-ai/proven/internal/generator"
)

// The assignment proves Table compiles against the real interface:
// generator.Generator and the type a reader copies agree on one method set.
var _ igen.Generator = gen.Table{}

// TestTableAnswersItsAgent proves one answer: a known agent gets its recorded
// answer and no error, and the answer is the one the table was built with.
func TestTableAnswersItsAgent(t *testing.T) {
	table, err := gen.NewTable([]gen.Row{
		{Agent: "planner", Answer: "the risks"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := table.Complete(context.Background(), "planner", "any prompt at all")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "the risks" {
		t.Fatalf("got %q, want %q", got, "the risks")
	}
}

// TestTableRefusesAnUnknownAgent proves the refusal: an unknown agent returns
// ErrUnknownAgent and no answer.
func TestTableRefusesAnUnknownAgent(t *testing.T) {
	table, err := gen.NewTable([]gen.Row{
		{Agent: "planner", Answer: "the risks"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := table.Complete(context.Background(), "fixer", "any prompt at all")
	if got != "" {
		t.Fatalf("got %q, want no answer", got)
	}
	if !errors.Is(err, gen.ErrUnknownAgent) {
		t.Fatalf("got %v, want %v", err, gen.ErrUnknownAgent)
	}
}

// TestTableRefusesADuplicateAgent proves the builder refuses two rows for one agent.
func TestTableRefusesADuplicateAgent(t *testing.T) {
	_, err := gen.NewTable([]gen.Row{
		{Agent: "planner", Answer: "one"},
		{Agent: "planner", Answer: "two"},
	})
	if err == nil {
		t.Fatal("a duplicate agent must be refused")
	}
}
