package review

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/pipeline"
)

// rowSQL is the discount fixture query that returns one advisory row.
const rowSQL = "SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'"

// defaultOnlyTitle marks a risk only the default run names; the special run must never see it.
const defaultOnlyTitle = "DEFAULT-ONLY-MARKER-7f3a"

// TestDoubledSecurityRunsBothGenerators proves ADR-0009 point 3 end to end: a doubled security review
// runs the planner and the skeptic a second time on the special model, as a fresh agent that never sees
// the first run's risks or answers. A risk both runs found is one finding naming both models, the default
// first; a risk one run found names that run; a risk with no evidence is dropped whichever run found it;
// and the trace records which run made each claim.
func TestDoubledSecurityRunsBothGenerators(t *testing.T) {
	def := newScript("NVIDIA", risks(
		risk("B1", "seen by both"),
		risk("D1", defaultOnlyTitle),
		risk("N1", "no evidence from either"),
	), `{"refuted":false,"reason":"default-saw-it"}`)
	def.sql = map[string]string{"B1": rowSQL, "D1": rowSQL, "S1": rowSQL}
	spec := newScript("qwen-special", risks(
		risk("B1", "seen by both"),
		risk("S1", "special-only-marker"),
		risk("N1", "no evidence from either"),
		risk("N2", "special without evidence"),
	), `{"refuted":false,"reason":"special-fresh"}`)
	cfg := defaultSettings()
	cfg.SpecialReviewer = "qwen-special"
	res := reviewScripted(t, cfg, def, spec)

	if res.Class != "security" || !res.Doubled {
		t.Fatalf("class %s doubled %t", res.Class, res.Doubled)
	}
	if !called(spec, "planner") || !called(spec, "skeptic") {
		t.Fatalf("special model was not called: %#v", spec.calls)
	}
	for _, c := range spec.calls {
		if c.agent != "planner" && c.agent != "skeptic" {
			t.Fatalf("special generator used by %s", c.agent)
		}
		if strings.Contains(c.prompt, defaultOnlyTitle) || strings.Contains(c.prompt, "default-saw-it") {
			t.Fatalf("special run saw the first run: %s", c.agent)
		}
	}
	if !modelUsed(res, "planner", "qwen-special") {
		t.Fatalf("models %#v", res.Models)
	}
	b1, ok := finding(res, "B1")
	if !ok || !same(b1.Models, []string{"NVIDIA", "qwen-special"}) {
		t.Fatalf("B1 %#v", b1.Models)
	}
	if len(b1.Evidence) != 2 || b1.Evidence[0].Query == nil || len(b1.Evidence[0].Query.Rows) != 1 || b1.Evidence[0].Query.SQL != rowSQL {
		t.Fatalf("B1 evidence %#v", b1.Evidence)
	}
	d1, ok := finding(res, "D1")
	if !ok || !same(d1.Models, []string{"NVIDIA"}) {
		t.Fatalf("D1 %#v", d1.Models)
	}
	s1, ok := finding(res, "S1")
	if !ok || !same(s1.Models, []string{"qwen-special"}) {
		t.Fatalf("S1 %#v", s1.Models)
	}
	if _, ok := finding(res, "N1"); ok {
		t.Fatal("N1 has no evidence and was kept")
	}
	if _, ok := finding(res, "N2"); ok {
		t.Fatal("N2 has no evidence and was kept")
	}
	runs := claimRuns(t, res.TracePath, "B1")
	if !same(runs, []string{"default", "special"}) {
		t.Fatalf("B1 claims %#v", runs)
	}
	if !same(claimRuns(t, res.TracePath, "D1"), []string{"default"}) {
		t.Fatal("D1 claim run")
	}
	if !same(claimRuns(t, res.TracePath, "S1"), []string{"special"}) {
		t.Fatal("S1 claim run")
	}
}

// TestNonDoubledClassRunsOnce proves that a class that is not doubled never calls the special model,
// even when one is configured.
func TestNonDoubledClassRunsOnce(t *testing.T) {
	def := newScript("NVIDIA", risks(risk("D1", defaultOnlyTitle)), `{"refuted":false,"reason":"no refutation"}`)
	def.sql = map[string]string{"D1": rowSQL}
	spec := newScript("qwen-special", risks(risk("S1", "special-only-marker")), `{"refuted":false,"reason":"no refutation"}`)
	cfg := defaultSettings()
	cfg.SpecialReviewer = "qwen-special"
	cfg.PathRules = []pathRule{{Pattern: "**/*.md", Class: "docs"}}
	res := reviewScripted(t, cfg, def, spec)
	if res.Class != "backend" || res.Doubled {
		t.Fatalf("class %s doubled %t", res.Class, res.Doubled)
	}
	if len(spec.calls) != 0 {
		t.Fatalf("special model was called: %#v", spec.calls)
	}
	if !called(def, "planner") {
		t.Fatal("default planner did not run")
	}
	d1, ok := finding(res, "D1")
	if !ok || !same(d1.Models, []string{"NVIDIA"}) {
		t.Fatalf("D1 %#v ok %v", d1.Models, ok)
	}
	if _, ok := finding(res, "S1"); ok {
		t.Fatal("special-only finding appeared on a single run")
	}
}

// TestDoubledWithoutSpecialModelRunsOnce proves that a doubled class with no special model configured
// runs once and says so, rather than inventing a second reviewer.
func TestDoubledWithoutSpecialModelRunsOnce(t *testing.T) {
	def := newScript("NVIDIA", risks(risk("D1", defaultOnlyTitle)), `{"refuted":false,"reason":"no refutation"}`)
	def.sql = map[string]string{"D1": rowSQL}
	spec := newScript("qwen-special", risks(risk("S1", "special-only-marker")), `{"refuted":false,"reason":"no refutation"}`)
	cfg := defaultSettings()
	res := reviewScripted(t, cfg, def, spec)
	if res.Class != "security" || !res.Doubled {
		t.Fatalf("class %s doubled %t", res.Class, res.Doubled)
	}
	joined := strings.Join(res.ReviewerStatements, "\n")
	if !strings.Contains(joined, "special-review model: not configured") {
		t.Fatal(joined)
	}
	if len(spec.calls) != 0 {
		t.Fatalf("special model was called: %#v", spec.calls)
	}
	if _, ok := finding(res, "D1"); !ok {
		t.Fatal("D1 missing")
	}
	if _, ok := finding(res, "S1"); ok {
		t.Fatal("S1 appeared without a special model")
	}
}

// reviewScripted runs one review with the explicit settings cfg and the two scripted generators. Its
// environment lookup fails the test, because a sample review must read no environment.
func reviewScripted(t *testing.T, cfg fileSettings, def, spec *scriptGen) Result {
	t.Helper()
	p, err := pipeline.Build(pipeline.Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: def,
		Special:   spec,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := &runner{p: p, cfg: cfg, explicit: true}
	res, err := r.Review(context.Background(), Request{Target: discountFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// scriptCall is one call a scripted generator received.
type scriptCall struct {
	agent  string
	prompt string
}

// scriptGen is a generator that answers from a script and records every call, so a test can see
// which model each agent used and what it was shown.
type scriptGen struct {
	model   string
	planner string
	skeptic string
	sql     map[string]string
	calls   []scriptCall
}

// newScript returns a scripted generator for model, with the planner's and the skeptic's answers.
func newScript(model, planner, skeptic string) *scriptGen {
	return &scriptGen{model: model, planner: planner, skeptic: skeptic}
}

// Complete answers by agent: the planner's risks, the investigator's SQL for the risk named in the
// prompt, no test from the prover, and the skeptic's answer.
func (g *scriptGen) Complete(ctx context.Context, agent, prompt string) (string, error) {
	g.calls = append(g.calls, scriptCall{agent: agent, prompt: prompt})
	switch agent {
	case "planner":
		return g.planner, nil
	case "investigator":
		sql := ""
		if g.sql != nil {
			sql = g.sql[promptID(prompt)]
		}
		b, err := json.Marshal(map[string]string{"sql": sql})
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "prover":
		return `{"path":"","name":"","command":""}`, nil
	case "skeptic":
		return g.skeptic, nil
	default:
		return `{}`, nil
	}
}

// LastUse names the script's model as if a receipt had named it.
func (g *scriptGen) LastUse() (generator.Use, bool) {
	return generator.Use{Model: g.model, FromReceipt: true}, true
}

// riskRow is one risk in the planner's JSON answer.
type riskRow struct {
	ID        string `json:"id"`
	RuleID    string `json:"rule_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// risk returns a planner risk on the discount fixture.
func risk(id, title string) riskRow {
	return riskRow{
		// The message names the package the sample row is about: the Judge keeps a query only when its rows
		// relate to the finding.
		ID: id, RuleID: id, Title: title, Message: title + " (evil.example/leftpad 1.2.3)", Severity: "ERROR",
		Path: "billing/discount.go", StartLine: 1, EndLine: 1,
	}
}

// risks encodes rows as the planner's JSON answer.
func risks(rows ...riskRow) string {
	b, err := json.Marshal(rows)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// promptID returns the risk id an agent's prompt carries on its "id: " line.
func promptID(prompt string) string {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.HasPrefix(line, "id: ") {
			return strings.TrimPrefix(line, "id: ")
		}
	}
	return ""
}

// called reports whether agent called the generator g.
func called(g *scriptGen, agent string) bool {
	for _, c := range g.calls {
		if c.agent == agent {
			return true
		}
	}
	return false
}

// modelUsed reports whether the result names model, from a receipt, for agent.
func modelUsed(res Result, agent, model string) bool {
	for _, m := range res.Models {
		if m.Agent == agent && m.Model == model && m.FromReceipt {
			return true
		}
	}
	return false
}

// finding returns the finding with id.
func finding(res Result, id string) (Finding, bool) {
	for _, f := range res.Findings {
		if f.ID == id {
			return f, true
		}
	}
	return Finding{}, false
}

// same reports whether two string lists are equal, in order.
func same(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// claimRuns returns, in trace order, the run ("default" or "special") of each claim made for id.
func claimRuns(t *testing.T, path, id string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var runs []string
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" {
			continue
		}
		var e struct {
			FindingID string `json:"finding_id"`
			Kind      string `json:"kind"`
			Run       string `json:"run"`
		}
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		if e.FindingID == id && e.Kind == "claim" {
			runs = append(runs, e.Run)
		}
	}
	return runs
}

// discountFixture returns the discount fixture under the module's testdata directory.
func discountFixture(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(b), "module github.com/saphan-ai/proven") {
			return filepath.Join(dir, "testdata", "fixture", "discount")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
