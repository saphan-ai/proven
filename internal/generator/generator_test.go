package generator_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

// The sample answers only the exact agent and prompt it recorded; another
// prompt or another agent gets no answer.
func TestSampleMatchesAgentAndPrompt(t *testing.T) {
	s, err := generator.NewSample([]generator.Response{
		{Agent: "planner", Prompt: "name the risks\n", Response: "[]"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Complete(context.Background(), "planner", "name the risks\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "[]" {
		t.Fatalf("got %q", got)
	}
	if _, err := s.Complete(context.Background(), "planner", "other"); err == nil {
		t.Fatal("a different prompt must not match")
	}
	if _, err := s.Complete(context.Background(), "skeptic", "name the risks\n"); err == nil {
		t.Fatal("a different agent must not match")
	}
}

// A live generator with no settings fails as a named Nightshift outage,
// not with a panic or a silent empty answer.
func TestUnconfiguredLiveGeneratorIsDown(t *testing.T) {
	live := &generator.Live{}
	_, err := live.Complete(context.Background(), "planner", "name the risks\n")
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "not configured" {
		t.Fatalf("got %v", err)
	}
}
