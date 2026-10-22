package planner_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/service"
)

func TestPlannerReadsAFencedListWithProseAndATrailingComma(t *testing.T) {
	raw := "Risks follow.\n```json\n" +
		`[{"id":"F9","rule_id":"r","title":"t","message":"m","severity":"ERROR","path":"billing/discount.go","start_line":1,"end_line":1,},]` +
		"\n```\nDone.\n"
	g := &script{replies: []string{raw, "unused"}}
	risks, err := planner.Plan(context.Background(), discount(t), g)
	if err != nil {
		t.Fatal(err)
	}
	if len(risks) != 1 || risks[0].ID != "F9" || risks[0].Path != "billing/discount.go" {
		t.Fatalf("%#v", risks)
	}
	if len(g.prompts) != 1 {
		t.Fatalf("calls %d", len(g.prompts))
	}
}

func TestPlannerProseTwiceDegradesByName(t *testing.T) {
	g := &script{replies: []string{"1. this is not json", "1. still not json"}}
	_, err := planner.Plan(context.Background(), discount(t), g)
	d, ok := service.As(err)
	if !ok || d.Service != "planner" || d.Reason != "the model's answer did not follow the schema" {
		t.Fatalf("%v", err)
	}
	if len(g.prompts) != 2 {
		t.Fatalf("calls %d", len(g.prompts))
	}
}

type script struct {
	replies []string
	prompts []string
}

func (s *script) Complete(ctx context.Context, agent, prompt string) (string, error) {
	s.prompts = append(s.prompts, prompt)
	if len(s.replies) == 0 {
		return "", &service.Down{Service: agent, Reason: "no reply"}
	}
	text := s.replies[0]
	s.replies = s.replies[1:]
	return text, nil
}

func discount(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return filepath.Join(dir, "testdata", "fixture", "discount")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
