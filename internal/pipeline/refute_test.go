package pipeline

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/testrun"
)

// ADR-0006: a refuting test the Skeptic proposes that passes on the change drops
// the finding, and the trace says why. Other findings are not touched.
func TestRefutationThatPassesOnTheChangeDropsTheFinding(t *testing.T) {
	body := `{"refuted":true,"reason":"the test passes","path":"billing/discount_test.go","name":"TestRefute","sql":""}`
	d := reviewWithSkeptic(t, body, &scriptedRun{pass: "TestRefute"})
	if _, ok := kept(d, "F1"); ok {
		t.Fatal("F1 must be dropped")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 must stay")
	}
	if _, ok := kept(d, "F3"); ok {
		t.Fatal("F3 must stay dropped")
	}
	why := "dropped: refuted: test TestRefute passes on the change"
	if !traceSays(t, d.TracePath, "F1", "verdict", why) {
		t.Fatal("the trace does not say why F1 was dropped")
	}
	if !traceSays(t, d.TracePath, "F1", "refutation", "test TestRefute passes on the change") {
		t.Fatal("the refutation does not say why")
	}
}

// ADR-0006: a refutation with no test and no query is only an opinion, so the
// finding is kept.
func TestRefutationWithNoEvidenceLeavesTheFinding(t *testing.T) {
	body := `{"refuted":true,"reason":"the model disagrees","path":"","name":"","command":"","sql":""}`
	d := reviewWithSkeptic(t, body, &scriptedRun{})
	f1, ok := kept(d, "F1")
	if !ok {
		t.Fatal("F1 must stay")
	}
	if f1.Dropped || f1.Unresolved || f1.DropReason != "kept" {
		t.Fatalf("%#v", f1)
	}
	if !traceSays(t, d.TracePath, "F1", "verdict", "kept") {
		t.Fatal("F1 verdict")
	}
}

// A refuting test that cannot run leaves the finding in the review, marked
// unresolved with that reason, neither kept as proved nor dropped.
func TestRefutingTestThatCannotRunIsUnresolved(t *testing.T) {
	body := `{"refuted":true,"reason":"try this","path":"billing/discount_test.go","name":"TestMissing","sql":""}`
	d := reviewWithSkeptic(t, body, &scriptedRun{miss: "TestMissing"})
	f1, ok := kept(d, "F1")
	if !ok {
		t.Fatal("F1 must stay")
	}
	if f1.Dropped || !f1.Unresolved || f1.DropReason != "unresolved: the refuting test could not run" {
		t.Fatalf("%#v", f1)
	}
	if !traceSays(t, d.TracePath, "F1", "verdict", "unresolved: the refuting test could not run") {
		t.Fatal("verdict")
	}
}

// ADR-0006: when the Skeptic and the finder are of the same model family, the
// trace says so by name, and the review still runs.
func TestSameFamilyIsStatedByName(t *testing.T) {
	model := generator.DefaultSkepticModel
	d := reviewReported(t, map[string]string{"*": model})
	want := "skeptic family nemotron is the same as the finder family nemotron (" + model + " and " + model + ")"
	if !traceHas(t, d.TracePath, want) {
		t.Fatalf("missing %s", want)
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("the review must still run")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 must stay")
	}
}

// When the Skeptic and the finder are of different families, no same-family
// sentence is written.
func TestDifferentFamiliesAreNotNamed(t *testing.T) {
	d := reviewReported(t, map[string]string{
		"planner": generator.DefaultSkepticModel,
		"skeptic": "anthropic/claude-opus",
	})
	b, err := os.ReadFile(d.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("is the same as the finder")) {
		t.Fatalf("named a same family\n%s", b)
	}
}

// A sample review, whose models are unknown, never states that the Skeptic and
// the finder share a family.
func TestSampleReviewDoesNotNameAnUnknownFamily(t *testing.T) {
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(d.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("is the same as the finder")) {
		t.Fatal("unknown was reported as a family")
	}
}

func reviewWithSkeptic(t *testing.T, f1Body string, run *scriptedRun) candidate.Draft {
	t.Helper()
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	run.inner = testrun.NewSample(fx.Tests)
	p, err := Build(Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: &answerFor{inner: inner, agent: "skeptic", match: "id: F1\n", body: f1Body},
		TestRun:   run,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Review(context.Background(), Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func reviewReported(t *testing.T, byAgent map[string]string) candidate.Draft {
	t.Helper()
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: &reported{inner: inner, byAgent: byAgent},
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Review(context.Background(), Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func loadDiscount(t *testing.T) *fixture.Fixture {
	t.Helper()
	fx, err := fixture.Load(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	return fx
}

func traceHas(t *testing.T, path, text string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Contains(b, []byte(text))
}

type answerFor struct {
	inner generator.Generator
	agent string
	match string
	body  string
}

func (g *answerFor) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == g.agent && bytes.Contains([]byte(prompt), []byte(g.match)) {
		return g.body, nil
	}
	return g.inner.Complete(ctx, agent, prompt)
}

type reported struct {
	inner   generator.Generator
	byAgent map[string]string
	last    string
}

func (g *reported) Complete(ctx context.Context, agent, prompt string) (string, error) {
	g.last = agent
	return g.inner.Complete(ctx, agent, prompt)
}

func (g *reported) LastUse() (generator.Use, bool) {
	model := g.byAgent[g.last]
	if model == "" {
		model = g.byAgent["*"]
	}
	if model == "" {
		return generator.Use{}, false
	}
	return generator.Use{Model: model, FromReceipt: true}, true
}

type scriptedRun struct {
	inner testrun.Adapter
	pass  string
	miss  string
}

func (s *scriptedRun) Run(ctx context.Context, req testrun.Request) (testrun.Outcome, error) {
	if s.pass != "" && req.Name == s.pass {
		return testrun.Outcome{
			Found: true, Path: req.Path, Name: req.Name,
			FailsOnChange: false, PassesOnBase: true, Output: "change: PASS",
		}, nil
	}
	if s.miss != "" && req.Name == s.miss {
		return testrun.Outcome{Found: false}, nil
	}
	if s.inner == nil {
		return testrun.Outcome{Found: false}, nil
	}
	return s.inner.Run(ctx, req)
}
