package planner_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/service"
)

// The Planner reads its risks from a fenced list wrapped in prose with a trailing comma.
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

// Two unreadable replies degrade the Planner by name after exactly two calls.
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

// script is a generator that returns its replies in order and records each prompt.
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

// discount is the path of the discount fixture change in the module's test data.
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

// A live review exports a real repository whole, so its Planner is asked about the files that changed and never the rest:
// an unchanged file's content does not reach the prompt, while an edited, an added and a deleted file do.
// Mutated field: the unchanged file's two sides are equal.
func TestThePlannerIsNotSentAnUnchangedFile(t *testing.T) {
	root := t.TempDir()
	write := func(side, name, body string) {
		t.Helper()
		p := filepath.Join(root, side, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("base", "same.go", "package same // UNCHANGED-MARKER\n")
	write("change", "same.go", "package same // UNCHANGED-MARKER\n")
	write("base", "edit.go", "package edit // EDIT-BEFORE\n")
	write("change", "edit.go", "package edit // EDIT-AFTER\n")
	write("change", "added.go", "package added // ADDED-MARKER\n")
	write("base", "gone.go", "package gone // DELETED-MARKER\n")
	// A sample review's recorded answers are keyed by the prompt over every file, so without the mark the
	// unchanged file is still sent.
	g := &script{replies: []string{"[]", "[]"}}
	if _, err := planner.Plan(context.Background(), root, g); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains([]byte(g.prompts[0]), []byte("UNCHANGED-MARKER")) {
		t.Error("an unmarked (sample) review lost the unchanged file")
	}
	if _, err := planner.Plan(change.WithChangedOnly(context.Background(), nil), root, g); err != nil {
		t.Fatal(err)
	}
	if len(g.prompts) != 2 {
		t.Fatalf("calls %d", len(g.prompts))
	}
	p := g.prompts[1]
	if bytes.Contains([]byte(p), []byte("UNCHANGED-MARKER")) {
		t.Error("the unchanged file reached the prompt")
	}
	for _, m := range []string{"EDIT-AFTER", "ADDED-MARKER", "DELETED-MARKER"} {
		if !bytes.Contains([]byte(p), []byte(m)) {
			t.Errorf("%s is missing from the prompt", m)
		}
	}
}

// A change of mode only reads the same bytes on both sides, so the Planner is sent the file git names and not the
// rest of the exported repository; with nothing named and nothing different, the model is not called at all.
// Mutated field: the paths git names as changed.
func TestAModeOnlyChangeSendsOnlyTheFileGitNames(t *testing.T) {
	root := t.TempDir()
	for _, side := range []string{"base", "change"} {
		for name, body := range map[string]string{"run.sh": "echo MODE-MARKER\n", "other.go": "package other // UNRELATED-MARKER\n"} {
			p := filepath.Join(root, side, name)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	g := &script{replies: []string{"[]"}}
	if _, err := planner.Plan(change.WithChangedOnly(context.Background(), []string{"run.sh"}), root, g); err != nil {
		t.Fatal(err)
	}
	if len(g.prompts) != 1 || !bytes.Contains([]byte(g.prompts[0]), []byte("MODE-MARKER")) || bytes.Contains([]byte(g.prompts[0]), []byte("UNRELATED-MARKER")) {
		t.Fatalf("prompts %d; the prompt must carry run.sh and not other.go", len(g.prompts))
	}
	g = &script{}
	risks, err := planner.Plan(change.WithChangedOnly(context.Background(), nil), root, g)
	if err != nil || len(risks) != 0 || len(g.prompts) != 0 {
		t.Fatalf("risks %v err %v calls %d; nothing changed, so no call", risks, err, len(g.prompts))
	}
}
