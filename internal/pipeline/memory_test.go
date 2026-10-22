package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
)

// The skeptic refutes F1, which has test evidence, on the first review.
// The second review of the same fixture does not report F1, and the trace
// names the H-MEM record. F2 is not refuted and stays.
func TestRefutedFindingIsAbsentOnTheNextReview(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	gen := refuteGenerator(t, "F1")
	first := reviewWith(t, filepath.Join(root, "logs1"), state, gen)
	if _, ok := kept(first, "F1"); !ok {
		t.Fatal("F1 must be reported on the review that refutes it")
	}
	if _, ok := kept(first, "F2"); !ok {
		t.Fatal("F2 must stay")
	}
	if id := memoryID(t, first.TracePath, "F1"); id != "" {
		t.Fatalf("the first review suppressed F1 as %s", id)
	}

	second := reviewWith(t, filepath.Join(root, "logs2"), state, gen)
	if _, ok := kept(second, "F1"); ok {
		t.Fatal("F1 returned without the store being cleared")
	}
	if _, ok := kept(second, "F2"); !ok {
		t.Fatal("F2 must stay on the second review")
	}
	id := memoryID(t, second.TracePath, "F1")
	if !strings.HasPrefix(id, "mem_") {
		t.Fatalf("trace id %q", id)
	}
}

// Mutated field: the H-MEM state file. The world log and the skeptic's answer
// stay as they were. Clearing the store brings F1 back.
func TestNegativeControl_ClearHMEMStore(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	gen := refuteGenerator(t, "F1")
	if _, ok := kept(reviewWith(t, filepath.Join(root, "logs1"), state, gen), "F1"); !ok {
		t.Fatal("F1 must be reported on the first review")
	}
	path := filepath.Join(state, "hmem", "state.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(before), "mem_") {
		t.Fatalf("store did not hold a refutation: %s", before)
	}
	empty := []byte("{\"version\":1,\"seq\":0,\"records\":[]}\n")
	if err := os.WriteFile(path, empty, 0o644); err != nil {
		t.Fatal(err)
	}
	worldPath := filepath.Join(state, "world", "state.json")
	worldBefore, err := os.ReadFile(worldPath)
	if err != nil {
		t.Fatal(err)
	}
	third := reviewWith(t, filepath.Join(root, "logs2"), state, gen)
	if _, ok := kept(third, "F1"); !ok {
		t.Fatal("F1 must return after the H-MEM store is cleared")
	}
	worldAfter, err := os.ReadFile(worldPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(worldBefore) == "" || len(worldAfter) == 0 {
		t.Fatal("world state was removed")
	}
}

func reviewWith(t *testing.T, logDir, state string, gen generator.Generator) candidate.Draft {
	t.Helper()
	p, err := Build(Options{
		Sample:    true,
		LogDir:    logDir,
		StateDir:  state,
		Generator: gen,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func refuteGenerator(t *testing.T, id string) generator.Generator {
	t.Helper()
	fx, err := fixture.Load(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	return &refuteOne{
		inner: inner,
		id:    id,
		body:  `{"refuted":true,"reason":"the helper applies the percent once"}`,
	}
}

type refuteOne struct {
	inner generator.Generator
	id    string
	body  string
}

func (g *refuteOne) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "skeptic" && strings.Contains(prompt, "id: "+g.id+"\n") {
		return g.body, nil
	}
	return g.inner.Complete(ctx, agent, prompt)
}

// memoryID is the H-MEM record named in the trace. An absent event returns "".
func memoryID(t *testing.T, path, id string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e struct {
			FindingID string `json:"finding_id"`
			Kind      string `json:"kind"`
			Text      string `json:"text"`
		}
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		if e.FindingID == id && e.Kind == "memory" && strings.HasPrefix(e.Text, "matched ") {
			return strings.TrimPrefix(e.Text, "matched ")
		}
	}
	return ""
}
