package review

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/testrun"
)

// scoutOff runs the plan tests with the engine removed: Scout Intent's switch off, as a review runs by
// default. Run as go test ./internal/review -run TestPlan -args -scout-off; the tests that prove a plan
// cuts or adds work must then fail, and the floor test must still pass.
var scoutOff = flag.Bool("scout-off", false, "run the plan tests with Scout Intent switched off")

// planned returns the default settings with Scout Intent on, unless -scout-off is given.
func planned() fileSettings {
	cfg := defaultSettings()
	cfg.ScoutIntent = !*scoutOff
	return cfg
}

// countingQuery is a query adapter that returns no rows and records every query it is asked. An OSV
// query gets the table's columns, as the store answers it, because the check refuses a result without them.
type countingQuery struct{ sql []string }

// Query records sql and returns no rows.
func (q *countingQuery) Query(_ context.Context, sql string) (query.Result, error) {
	q.sql = append(q.sql, sql)
	if strings.Contains(sql, "px_osv") {
		return query.Result{SQL: sql, Columns: []string{"id", "aliases", "ecosystem", "package", "introduced", "fixed", "last_affected", "severity", "published"}}, nil
	}
	return query.Result{SQL: sql}, nil
}

// countingRun is a test runner that runs nothing and counts the runs it is asked for.
type countingRun struct{ runs int }

// Run counts the request.
func (r *countingRun) Run(context.Context, testrun.Request) (testrun.Outcome, error) {
	r.runs++
	return testrun.Outcome{}, nil
}

// planRig is one review's scripted world: the two generators, the query adapter and the test runner,
// so a test reads afterwards which of them the review used.
type planRig struct {
	def, spec *scriptGen
	q         query.Adapter
	run       *countingRun
}

// queries returns the SQL the review asked, whichever counting adapter the rig holds.
func (rig *planRig) queries() []string {
	switch q := rig.q.(type) {
	case *countingQuery:
		return q.sql
	case *advisoryQuery:
		return q.sql
	}
	return nil
}

// newRig scripts a planner that names one risk, so a full plan reaches every agent after it.
func newRig() *planRig {
	return &planRig{
		def:  newScript("NVIDIA", risks(risk("R1", "a risk the planner names")), `{"refuted":false,"reason":"no refutation"}`),
		spec: newScript("qwen-special", risks(risk("S1", "a risk the second planner names")), `{"refuted":false,"reason":"no refutation"}`),
		q:    &countingQuery{},
		run:  &countingRun{},
	}
}

// review runs one live review of target with cfg, through the rig's stubs: no network, no container,
// no model. A special reviewer is configured, so a doubled review calls the second generator.
func (rig *planRig) review(t *testing.T, cfg fileSettings, target string) Result {
	t.Helper()
	return rig.reviewWith(t, cfg, target)
}

// reviewWith is review; the measurement calls it with settings it chose itself, not planned()'s.
func (rig *planRig) reviewWith(t *testing.T, cfg fileSettings, target string) Result {
	t.Helper()
	cfg.SpecialReviewer = "qwen-special"
	p, err := pipeline.Build(pipeline.Options{
		LogDir: t.TempDir(), Generator: rig.def, Special: rig.spec, LiveQuery: rig.q, TestRun: rig.run,
		GuardOff: true, Redactor: redact.FromEnv(func(string) string { return "" }),
		Env: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	r := &runner{p: p, cfg: cfg, explicit: true}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := r.Review(ctx, Request{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// calls counts every model call the review made, on both generators.
func (rig *planRig) calls() int { return len(rig.def.calls) + len(rig.spec.calls) }

// trees writes a prepared base/ and change/ pair. A file with an empty base is added; every other file
// differs between the sides.
func trees(t *testing.T, files map[string][2]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, sides := range files {
		for i, side := range []string{"base", "change"} {
			if sides[i] == "" {
				continue
			}
			p := filepath.Join(dir, side, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(sides[i]), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

// planLines returns the text of every plan line in a trace.
func planLines(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		var e struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if json.Unmarshal([]byte(line), &e) == nil && e.Kind == "plan" {
			out = append(out, e.Text)
		}
	}
	return out
}

// docsOnly, dependencyOnly and sensitive are the three changes whose plans ADR-0049 point 2 names.
func docsOnly(t *testing.T) string {
	return trees(t, map[string][2]string{
		"README.md":       {"# Old\n", "# New\n"},
		"docs/INSTALL.md": {"old\n", "new\n"},
	})
}

func dependencyOnly(t *testing.T) string {
	return trees(t, map[string][2]string{
		"go.mod": {"module example.com/app\n\ngo 1.22\n", "module example.com/app\n\ngo 1.22\n\nrequire example.com/lib v1.2.0\n"},
		"go.sum": {"", "example.com/lib v1.2.0 h1:abc=\n"},
	})
}

func sensitive(t *testing.T) string {
	return trees(t, map[string][2]string{
		"internal/session/store.go": {"package session\n", "package session\n\nvar TTL = 0\n"},
	})
}

// TestPlanDocsOnlyRunsNoModel proves ADR-0049 point 2: a documentation-only change runs no model call,
// on either reviewer, and still states its class. With the engine removed (-scout-off) the full plan
// runs and the planner is called, so this test fails.
func TestPlanDocsOnlyRunsNoModel(t *testing.T) {
	rig := newRig()
	res := rig.review(t, planned(), docsOnly(t))
	if n := rig.calls(); n != 0 {
		t.Fatalf("a documentation-only change made %d model calls: %+v", n, rig.def.calls)
	}
	if res.Class != "docs" {
		t.Fatalf("class %s", res.Class)
	}
}

// TestPlanDependencyOnlySkipsTheProver proves ADR-0049 point 2: a dependency-only change runs the
// dependency evidence (the OSV table is queried for the package the change adds) and no Prover call.
// With the engine removed the full plan runs the Prover, so this test fails.
func TestPlanDependencyOnlySkipsTheProver(t *testing.T) {
	rig := newRig()
	res := rig.review(t, planned(), dependencyOnly(t))
	if called(rig.def, "prover") || called(rig.spec, "prover") || rig.run.runs != 0 {
		t.Fatalf("the Prover ran on a dependency-only change: %d runs", rig.run.runs)
	}
	osvQueried := false
	for _, s := range rig.queries() {
		if strings.Contains(s, "px_osv") && strings.Contains(s, "example.com/lib") {
			osvQueried = true
		}
	}
	if !osvQueried {
		t.Fatalf("the dependency evidence did not run: %q", rig.queries())
	}
	if !called(rig.def, "planner") {
		t.Fatal("the planner did not run")
	}
	// The configured rules call go.mod security; the plan keeps that class (ADR-0049 point 3).
	if res.Class != "security" || !res.Doubled {
		t.Fatalf("class %s doubled %t", res.Class, res.Doubled)
	}
}

// TestPlanSecuritySensitiveIsDoubled proves ADR-0049 point 2: a change to session code, which no
// configured rule names, is planned as security-sensitive and doubled, so the second reviewer runs.
// With the engine removed the rules give it the fallback class, undoubled, so this test fails.
func TestPlanSecuritySensitiveIsDoubled(t *testing.T) {
	rig := newRig()
	res := rig.review(t, planned(), sensitive(t))
	if !res.Doubled || res.Class != "security" || len(rig.spec.calls) == 0 {
		t.Fatalf("class %s (%s) doubled %t, second reviewer calls %d", res.Class, res.ClassRule, res.Doubled, len(rig.spec.calls))
	}
	if res.ClassRule != "scout-intent: change.security" {
		t.Fatalf("rule %s", res.ClassRule)
	}
}

// TestPlanNeverLowersTheConfiguredClass proves ADR-0049 point 3: a documentation change that the
// configured rules review as backend keeps class backend and its full plan, though Scout plans it as
// documentation. It holds with the engine on and with it removed.
func TestPlanNeverLowersTheConfiguredClass(t *testing.T) {
	cfg := planned()
	cfg.PathRules = append([]pathRule{{Pattern: "docs/api/**", Class: "backend"}}, cfg.PathRules...)
	rig := newRig()
	res := rig.review(t, cfg, trees(t, map[string][2]string{"docs/api/endpoints.md": {"old\n", "new\n"}}))
	if res.Class != "backend" || res.ClassRule != "docs/api/**" {
		t.Fatalf("class %s (%s)", res.Class, res.ClassRule)
	}
	if !called(rig.def, "planner") || !called(rig.def, "prover") {
		t.Fatalf("the configured class did not run in full: %+v", rig.def.calls)
	}
	if !*scoutOff {
		if lines := planLines(t, res.TracePath); len(lines) == 0 || !strings.Contains(lines[0], "outranks the plan's class docs") {
			t.Fatalf("the trace does not say the floor held: %q", lines)
		}
	}
}

// TestPlanAmbiguousChangeRunsTheConfiguredClass proves that an ambiguous change, documentation beside
// code, is never planned as a confident class: it runs the class its rules choose in full, and the trace
// says it was ambiguous.
func TestPlanAmbiguousChangeRunsTheConfiguredClass(t *testing.T) {
	if *scoutOff {
		t.Skip("the ambiguity is Scout's; with it off there is nothing to read")
	}
	rig := newRig()
	res := rig.review(t, planned(), trees(t, map[string][2]string{
		"docs/guide.md":     {"old\n", "new\n"},
		"internal/x/run.go": {"package x\n", "package x\n\nvar N = 1\n"},
	}))
	if res.Class != "docs" && res.Class != "backend" {
		t.Fatalf("class %s", res.Class)
	}
	if !called(rig.def, "planner") || !called(rig.def, "prover") {
		t.Fatalf("an ambiguous change was cut: %+v", rig.def.calls)
	}
	lines := planLines(t, res.TracePath)
	if len(lines) == 0 || !strings.Contains(lines[0], "ambiguous, so no plan is chosen") {
		t.Fatalf("trace %q", lines)
	}
}

// TestPlanSwitchOffRunsEverythingAndSaysSo is the unplug test (law #41): with the switch off, a
// documentation-only change runs the full plan, and the trace names that no plan was chosen.
func TestPlanSwitchOffRunsEverythingAndSaysSo(t *testing.T) {
	cfg := defaultSettings()
	cfg.ScoutIntent = false
	rig := newRig()
	res := rig.review(t, cfg, docsOnly(t))
	if !called(rig.def, "planner") || !called(rig.def, "prover") || !called(rig.def, "skeptic") {
		t.Fatalf("the full plan did not run: %+v", rig.def.calls)
	}
	lines := planLines(t, res.TracePath)
	if len(lines) != 1 || !strings.Contains(lines[0], "no plan chosen (scout_intent is off)") {
		t.Fatalf("trace %q", lines)
	}
}

// TestPlanTraceNamesClassSignalsAndPlan proves the trace line of a chosen plan names the class, the
// signals that led there, the ambiguity and the plan.
func TestPlanTraceNamesClassSignalsAndPlan(t *testing.T) {
	if *scoutOff {
		t.Skip("no plan is chosen with the engine off")
	}
	rig := newRig()
	res := rig.review(t, planned(), docsOnly(t))
	lines := planLines(t, res.TracePath)
	if len(lines) == 0 {
		t.Fatal("no plan line")
	}
	for _, want := range []string{"class change.docs", "signals: M README.md matched **/*.md", "ambiguity: none", "agents none", "doubled no"} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("%q lacks %q", lines[0], want)
		}
	}
}

// TestScoutIntentIsOffByDefault proves the switch lands off (the owner's decision on ADR-0049), in the
// defaults and in an empty proven.yaml, and that a file can switch it on.
func TestScoutIntentIsOffByDefault(t *testing.T) {
	if defaultSettings().ScoutIntent {
		t.Fatal("on by default")
	}
	cfg, err := parseConfig([]byte("timeout: 0s\n"))
	if err != nil || cfg.ScoutIntent {
		t.Fatalf("%v %t", err, cfg.ScoutIntent)
	}
	cfg, err = parseConfig([]byte("scout_intent: on\n"))
	if err != nil || !cfg.ScoutIntent {
		t.Fatalf("%v %t", err, cfg.ScoutIntent)
	}
}

// TestPlanSettingsAreCheckedOnLoad proves a plan setting is read whole and refused by name when one field
// is wrong. Mutated field, one per case: doubled set to off, an unknown agent, a missing field, an unknown class.
func TestPlanSettingsAreCheckedOnLoad(t *testing.T) {
	cfg, err := parseConfig([]byte("plan_config: class=backend, agents=planner+skeptic, tools=none, sources=osv, doubled=on\n"))
	if err != nil {
		t.Fatal(err)
	}
	got := cfg.Plans["change.config"]
	if got.Class != "backend" || !same(got.Agents, []string{"planner", "skeptic"}) || len(got.Tools) != 0 || !same(got.Sources, []string{"osv"}) || !got.DoubledOn {
		t.Fatalf("%+v", got)
	}
	if !same(defaultSettings().Plans["change.config"].Agents, pipeline.PlanAgents) {
		t.Fatal("setting one plan changed the shared defaults")
	}
	for name, val := range map[string]string{
		"doubled off":   "class=docs, agents=none, tools=none, sources=none, doubled=off",
		"unknown agent": "class=docs, agents=planner+oracle, tools=none, sources=none, doubled=class",
		"missing field": "class=docs, agents=none, tools=none, doubled=class",
		"unknown class": "class=frontend, agents=none, tools=none, sources=none, doubled=class",
	} {
		_, err := parseConfig([]byte("plan_docs: " + val + "\n"))
		var ce *ConfigError
		if err == nil || !asConfigError(err, &ce) || ce.Field != "plan_docs" {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestFloorTakesThePolicyWhenItIsHigher proves the one floor function takes a policy minimum when one
// is given and it outranks the configured class, and keeps the configured class otherwise (ADR-0049
// point 3, with ADR-0036's policy fed in when it lands).
func TestFloorTakesThePolicyWhenItIsHigher(t *testing.T) {
	cfg := defaultSettings()
	files := []string{"internal/x/run.go"}
	if c, r := floorClass(files, cfg, ""); c != "backend" || r != "fallback" {
		t.Fatalf("%s %s", c, r)
	}
	if c, r := floorClass(files, cfg, "security"); c != "security" || r != "policy" {
		t.Fatalf("%s %s", c, r)
	}
	if c, _ := floorClass(files, cfg, "docs"); c != "backend" {
		t.Fatalf("a lower policy lowered the class to %s", c)
	}
}
