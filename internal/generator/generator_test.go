package generator_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

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

func TestLiveGeneratorIsDown(t *testing.T) {
	_, err := generator.Live{}.Complete(context.Background(), "planner", "name the risks\n")
	d, ok := service.As(err)
	if !ok || d.Service != "generator" {
		t.Fatalf("got %v", err)
	}
}
