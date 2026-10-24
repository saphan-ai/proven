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
	"github.com/saphan-ai/proven/internal/testrun"
)

// The skeptic refutes F1 with evidence: a refuting test that passes on the
// change. The first review drops F1 as refuted. The second review of the same
// fixture does not report F1 either, and the trace names the H-MEM record that
// suppressed it. F2 is not refuted and stays.
func TestRefutedFindingIsAbsentOnTheNextReview(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	first := reviewWithRefutingTest(t, filepath.Join(root, "logs1"), state)
	if _, ok := kept(first, "F1"); ok {
		t.Fatal("F1 must be dropped on the review whose refutation carries evidence")
	}
	if _, ok := kept(first, "F2"); !ok {
		t.Fatal("F2 must stay")
	}
	if id := memoryID(t, first.TracePath, "F1"); id != "" {
		t.Fatalf("the first review suppressed F1 as %s", id)
	}

	second := reviewWithRefutingTest(t, filepath.Join(root, "logs2"), state)
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
// (a refuting test that passes on the change) stay as they were. With the
// store cleared the third review drops F1 on its own judgement and no memory
// record suppresses it.
func TestNegativeControl_ClearHMEMStore(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	first := reviewWithRefutingTest(t, filepath.Join(root, "logs1"), state)
	if _, ok := kept(first, "F1"); ok {
		t.Fatal("F1 must be dropped on the first review")
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
	third := reviewWithRefutingTest(t, filepath.Join(root, "logs2"), state)
	if id := memoryID(t, third.TracePath, "F1"); id != "" {
		t.Fatalf("a cleared store still suppressed F1 as %s", id)
	}
	if !traceSays(t, third.TracePath, "F1", "verdict", "dropped: refuted: test TestRefute passes on the change") {
		t.Fatal("after the store is cleared, F1 must be judged again on its own evidence")
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

// findingMarker is the text of an agent prompt that names one finding by id.
func findingMarker(id string) string { return `{"id":"` + id + `",` }

// reviewWithRefutingTest runs a sample review whose Skeptic refutes F1 with a
// test that passes on the change, so the refutation carries evidence.
func reviewWithRefutingTest(t *testing.T, logDir, state string) candidate.Draft {
	t.Helper()
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{
		Sample:   true,
		LogDir:   logDir,
		StateDir: state,
		Generator: &answerFor{
			inner: inner, agent: "skeptic", match: findingMarker("F1"),
			body: `{"refuted":true,"reason":"the test passes","path":"billing/discount_test.go","name":"TestRefute"}`,
		},
		TestRun: &scriptedRun{inner: testrun.NewSample(fx.Tests), pass: "TestRefute"},
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fx.Root})
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
	if agent == "skeptic" && strings.Contains(prompt, findingMarker(g.id)) {
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
