package guardrail_test

import (
	"context"
	"testing"

	eg "github.com/saphan-ai/proven/examples/guardrail"
	ig "github.com/saphan-ai/proven/internal/guardrail"
)

// The assignment proves List compiles against the real interface:
// examples/guardrail.List and internal/guardrail.Adapter agree on one
// method set, and List also names its backend through Identity.
var _ ig.Adapter = eg.List{}
var _ ig.Identity = eg.List{}

// TestListAnswersSafe proves one answer: a prompt and an answer that name no
// barred word return a safe verdict, no error, and nothing is unsafe.
func TestListAnswersSafe(t *testing.T) {
	list, err := eg.NewList([]eg.Row{{Word: "forbidden", Category: "word list"}})
	if err != nil {
		t.Fatalf("NewList: %v", err)
	}
	v, err := list.Check(context.Background(), "please review the change", "it looks fine")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if v.Unsafe() {
		t.Fatalf("got an unsafe verdict, want a safe one: %+v", v)
	}
	if v.User != "safe" || v.Response != "safe" {
		t.Fatalf("got %q / %q, want safe / safe", v.User, v.Response)
	}
}

// TestListRefusesABarredWord proves the refusal: a barred word in the prompt
// makes that side unsafe, with the category the row names, so Wrap refuses
// the step by name instead of answering it.
func TestListRefusesABarredWord(t *testing.T) {
	list, err := eg.NewList([]eg.Row{{Word: "forbidden", Category: "word list"}})
	if err != nil {
		t.Fatalf("NewList: %v", err)
	}
	v, err := list.Check(context.Background(), "the forbidden value", "all good")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !v.Unsafe() || v.User != "unsafe" {
		t.Fatalf("got %+v, want an unsafe user verdict", v)
	}
	if len(v.Categories) != 1 || v.Categories[0] != "word list" {
		t.Fatalf("got categories %v, want [word list]", v.Categories)
	}
}

// TestListRefusesACanceledContext proves the interface's error path: a
// canceled context returns an error and no verdict.
func TestListRefusesACanceledContext(t *testing.T) {
	list, err := eg.NewList([]eg.Row{{Word: "forbidden", Category: "word list"}})
	if err != nil {
		t.Fatalf("NewList: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = list.Check(ctx, "hello", "world")
	if err == nil {
		t.Fatal("a canceled context must be refused")
	}
}

// TestListRefusesADuplicateWord proves the builder refuses two rows for one word.
func TestListRefusesADuplicateWord(t *testing.T) {
	_, err := eg.NewList([]eg.Row{
		{Word: "forbidden", Category: "one"},
		{Word: "forbidden", Category: "two"},
	})
	if err == nil {
		t.Fatal("a duplicate word must be refused")
	}
}
