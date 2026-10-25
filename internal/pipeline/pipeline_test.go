package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/investigator"
	"github.com/saphan-ai/proven/internal/agent/prover"
	"github.com/saphan-ai/proven/internal/agent/skeptic"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Mutated field: PROVEN_TEST_MEMORY. The live fallback must load it.
func TestLiveContainerSettingIsRefused(t *testing.T) {
	p, err := Build(Options{
		LogDir:   t.TempDir(),
		GuardOff: true,
		Redactor: redact.New(nil),
		Env: func(k string) string {
			if k == testrun.EnvMemory {
				return "lots"
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Review(context.Background(), Request{Target: fixturePath(t)})
	var cfg *testrun.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != testrun.EnvMemory {
		t.Fatalf("%v", err)
	}
	if !strings.Contains(err.Error(), "allowed") {
		t.Fatalf("%v", err)
	}
}

// With the test runner unplugged, the Prover is named as off with its reason,
// a finding that needed a test is not reported as proved, and a finding with
// query evidence stays.
func TestUnplugLiveTestRunNamesProverOff(t *testing.T) {
	p := buildSample(t, t.TempDir(), testrun.Live{})
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if d.Partial {
		t.Fatal("partial")
	}
	if _, ok := kept(d, "F1"); ok {
		t.Fatal("F1 must not be reported as proved")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 still has query evidence")
	}
	found := false
	for _, f := range d.Failed {
		if f.Service == "prover" && f.Reason == testrun.ProverOffReason {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", d.Failed)
	}
}

// Property: query rows for a package the change does not require are not
// evidence about the change, even when the Investigator's recorded request
// names that package. Mutated fixture field: tables.sql. An OSV row is added for
// package billing/note, the package F3's recorded template request names; the
// change requires no such package, so F3 stays dropped.
func TestNegativeControl_F3RecordedQueryRow(t *testing.T) {
	dir := copyFixture(t)
	sql := filepath.Join(dir, "tables.sql")
	f, err := os.OpenFile(sql, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-px-note-0001', 'CVE-2026-0099', 'Go', 'billing/note', '0', '1.0.0', '', 'LOW', '2026-04-01');\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	orig, err := os.ReadFile(filepath.Join(fixturePath(t), "tables.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(orig, []byte("billing/note")) {
		t.Fatal("the committed tables.sql was mutated")
	}
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := kept(d, "F3"); ok {
		t.Fatal("F3 must stay dropped: its rows are about a package the change does not require")
	}
	// The same review keeps the finding whose rows are about the change's own package.
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 must stay: its rows are about the package the change requires")
	}
}

// Mutated field: the skeptic's answer. Every skeptic prompt, including the one
// repair, returns prose. The review must finish, name skeptic, keep F1 and F2,
// and drop F3 for lack of evidence.
// Negative control: remove the schema degrade (the service.Down returned after
// the second decode) and this test turns red.
func TestProseTwiceDegradesByNameAndDropsUnevidenced(t *testing.T) {
	fx, err := fixture.Load(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	wrap := &proseTwice{inner: inner, agent: "skeptic", text: "1. this is not json"}
	p, err := Build(Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: wrap,
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
		t.Fatalf("review ended: %v", err)
	}
	if d.Partial {
		t.Fatal("partial")
	}
	if wrap.calls != 8 {
		t.Fatalf("skeptic calls %d, want 8 (one repair each for F1, F2, F3 and F4)", wrap.calls)
	}
	named := false
	for _, f := range d.Failed {
		if f.Service == "skeptic" && f.Reason == "the model's answer did not follow the schema" {
			named = true
		}
	}
	if !named || len(d.Failed) != 1 {
		t.Fatalf("failed %#v", d.Failed)
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("F1 must stay")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 must stay")
	}
	if _, ok := kept(d, "F3"); ok {
		t.Fatal("F3 must stay dropped")
	}
	if !traceSays(t, d.TracePath, "F3", "verdict", "dropped: no evidence") {
		t.Fatal("F3 was not dropped for lack of evidence")
	}
}

// proseTwice is a generator whose one agent answers with the same prose twice.
type proseTwice struct {
	inner generator.Generator
	agent string
	text  string
	calls int
}

func (g *proseTwice) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == g.agent {
		g.calls++
		return g.text, nil
	}
	return g.inner.Complete(ctx, agent, prompt)
}

func traceSays(t *testing.T, path, id, kind, text string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(line) == 0 {
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
		if e.FindingID == id && e.Kind == kind && e.Text == text {
			return true
		}
	}
	return false
}

// Mutated field: the package column of the OSV sample rows for leftpad. It
// changes to a package the change does not require, so F2's template request
// (osv-by-package for evil.example/leftpad) returns no row.
func TestNegativeControl_PackageTheAdvisoryDoesNotName(t *testing.T) {
	dir := copyFixture(t)
	path := filepath.Join(dir, "tables.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const named = "'Go', 'evil.example/leftpad'"
	if bytes.Count(b, []byte(named)) != 2 {
		t.Fatalf("package occurrences %d", bytes.Count(b, []byte(named)))
	}
	next := bytes.ReplaceAll(b, []byte(named), []byte("'Go', 'evil.example/other'"))
	if err := os.WriteFile(path, next, 0o644); err != nil {
		t.Fatal(err)
	}
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := kept(d, "F2"); ok {
		t.Fatal("F2 must be absent when the OSV table has no row for its package")
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("F1 must stay")
	}
	// F4 is keyed by cve_id, which this mutation does not change. But the
	// change is no longer affected by any advisory row, so the CVE is not about
	// the change and F4 is dropped too.
	if _, ok := kept(d, "F4"); ok {
		t.Fatal("F4 must be absent when no advisory row names a package the change requires")
	}
}

// Mutated field: cve_id on the KEV sample row that was CVE-2026-0001. After the
// change the CVE is absent from the KEV view that F5's kev-by-cve template
// request reads, and the review must not read that as "not exploited".
func TestNegativeControl_CVEAbsentFromKEVIsNotCalledNotExploited(t *testing.T) {
	dir := copyFixture(t)
	const title = "KEV entry for CVE-2026-0001"
	const answer = `{"template":"kev-by-cve","params":{"cve_id":"CVE-2026-0001"}}`
	if err := addQueryFinding(dir, title, answer); err != nil {
		t.Fatal(err)
	}
	p := buildSample(t, t.TempDir(), nil)
	before, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	f5, ok := kept(before, "F5")
	if !ok || len(f5.Evidence) != 2 || f5.Evidence[0].Query == nil || len(f5.Evidence[0].Query.Rows) != 1 {
		t.Fatal("the KEV row must keep F5 while the CVE is in the KEV view")
	}
	if f5.Evidence[1].Query == nil || len(f5.Evidence[1].Query.Rows) != 1 || f5.Evidence[1].Query.Rows[0][1] != "0.50000" {
		t.Fatalf("EPSS row %#v", f5.Evidence[1])
	}
	if f5.Evidence[0].Query.Rows[0][0] != "CVE-2026-0001" {
		t.Fatalf("row %#v", f5.Evidence[0].Query.Rows)
	}
	tables := filepath.Join(dir, "tables.sql")
	b, err := os.ReadFile(tables)
	if err != nil {
		t.Fatal(err)
	}
	from := []byte("INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001'")
	to := []byte("INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-0000-0000'")
	if bytes.Count(b, from) != 1 {
		t.Fatal("the KEV row for CVE-2026-0001 was not found once")
	}
	if err := os.WriteFile(tables, bytes.Replace(b, from, to, 1), 0o644); err != nil {
		t.Fatal(err)
	}
	p = buildSample(t, t.TempDir(), nil)
	after, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := kept(after, "F5"); ok {
		t.Fatal("F5 must be absent when the CVE is missing from the KEV view")
	}
	if !traceSays(t, after.TracePath, "F5", "verdict", "dropped: no evidence") {
		t.Fatal("F5 was not dropped for lack of evidence")
	}
	raw, err := os.ReadFile(after.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(bytes.ToLower(raw), []byte("not exploited")) {
		t.Fatal("the review read a missing KEV row as not exploited")
	}
	for _, c := range after.Findings {
		if strings.Contains(strings.ToLower(c.Message), "not exploited") || strings.Contains(strings.ToLower(c.Title), "not exploited") {
			t.Fatalf("finding %s says not exploited", c.ID)
		}
	}
}

// addQueryFinding adds a finding F5 to the fixture in dir, with the recorded
// Investigator answer given as JSON (a template and its values).
func addQueryFinding(dir, title, answer string) error {
	path := filepath.Join(dir, "fixture.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc struct {
		Snapshot        string          `json:"snapshot"`
		Hash            string          `json:"planner_prompt_sha256"`
		PlannerResponse json.RawMessage `json:"planner_response"`
		Responses       []struct {
			Agent    string `json:"agent"`
			Prompt   string `json:"prompt"`
			Response string `json:"response"`
		} `json:"responses"`
		RecordedTests json.RawMessage `json:"recorded_tests"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	var risks []map[string]any
	if err := json.Unmarshal(doc.PlannerResponse, &risks); err != nil {
		return err
	}
	risks = append(risks, map[string]any{
		"id": "F5", "rule_id": "kev-epss-join", "title": title,
		"message":    "The KEV view returned a row for the CVE.",
		"severity":   "ERROR",
		"path":       "go.mod",
		"start_line": 5,
		"end_line":   5,
	})
	rawRisks, err := json.Marshal(risks)
	if err != nil {
		return err
	}
	doc.PlannerResponse = rawRisks
	doc.Responses = append(doc.Responses,
		struct {
			Agent    string `json:"agent"`
			Prompt   string `json:"prompt"`
			Response string `json:"response"`
		}{Agent: "investigator", Prompt: investigator.Prompt("F5", title, "go.mod", nil), Response: answer},
		struct {
			Agent    string `json:"agent"`
			Prompt   string `json:"prompt"`
			Response string `json:"response"`
		}{Agent: "prover", Prompt: prover.Prompt("F5", title), Response: `{"path":"","name":"","command":""}`},
		struct {
			Agent    string `json:"agent"`
			Prompt   string `json:"prompt"`
			Response string `json:"response"`
		}{Agent: "skeptic", Prompt: skeptic.Prompt("F5", title, nil), Response: `{"refuted":false,"reason":"no refutation"}`},
	)
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// Mutated fixture field: recorded_tests. The entry whose finding_id is F1 is removed.
func TestNegativeControl_RemoveF1RecordedTest(t *testing.T) {
	dir := copyFixture(t)
	path := filepath.Join(dir, "fixture.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	raw, ok := doc["recorded_tests"].([]any)
	if !ok || len(raw) != 1 {
		t.Fatalf("recorded_tests %#v", doc["recorded_tests"])
	}
	doc["recorded_tests"] = []any{}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	rewritten, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(rewritten, []byte("TestDiscounted")) {
		t.Fatal("the prover's recorded answer must stay; only recorded_tests changes")
	}
	if bytes.Contains(rewritten, []byte(`"finding_id": "F1"`)) {
		t.Fatal("recorded_tests still has F1")
	}
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := kept(d, "F1"); ok {
		t.Fatal("F1 must disappear when its recorded test is removed")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 must stay")
	}
}

// dirProbe records the last request it was asked to run.
type dirProbe struct {
	req testrun.Request
}

func (d *dirProbe) Run(ctx context.Context, req testrun.Request) (testrun.Outcome, error) {
	d.req = req
	return testrun.Outcome{}, nil
}

// checkoutRun fills the base and change directories only when the request left
// them empty, and keeps the ones the caller set.
func TestCheckoutRunFillsEmptyDirectories(t *testing.T) {
	probe := &dirProbe{}
	a := checkoutRun{inner: probe, base: "/tmp/base", change: "/tmp/change"}
	if _, err := a.Run(context.Background(), testrun.Request{Path: "billing/discount_test.go", Name: "TestDiscounted"}); err != nil {
		t.Fatal(err)
	}
	if probe.req.BaseDir != "/tmp/base" || probe.req.ChangeDir != "/tmp/change" {
		t.Fatalf("%#v", probe.req)
	}
	if _, err := a.Run(context.Background(), testrun.Request{Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: "/already", ChangeDir: "/set"}); err != nil {
		t.Fatal(err)
	}
	if probe.req.BaseDir != "/already" || probe.req.ChangeDir != "/set" {
		t.Fatalf("%#v", probe.req)
	}
}

// Mutated field: Options.MaxWrittenBytes. A body over that limit is not run.
func TestWrittenLimitFromOptionsRefusesBeforeTheRun(t *testing.T) {
	fx, err := fixture.Load(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	body := "package billing\n\nimport \"testing\"\n\nfunc TestExtra(t *testing.T) {\n\tif Discounted(100, 20) != 80 {\n\t\tt.Fatal(Discounted(100, 20))\n\t}\n}\n"
	raw, err := json.Marshal(map[string]string{
		"path": "billing/extra_test.go", "content": body, "name": "TestExtra",
	})
	if err != nil {
		t.Fatal(err)
	}
	run := &outcomeRun{out: testrun.Outcome{Found: true, PassesOnBase: true, Path: "billing/extra_test.go", Name: "TestExtra"}}
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), Generator: &proverSwap{inner: inner, response: string(raw)},
		TestRun: run, MaxWrittenBytes: 8,
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
	if run.calls != 0 {
		t.Fatalf("calls %d", run.calls)
	}
	if !traceSays(t, d.TracePath, "F1", "evidence", "refused: "+testrun.ReasonTooLarge) {
		b, readErr := os.ReadFile(d.TracePath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		t.Fatalf("trace:\n%s", b)
	}
}

// A finding whose evidence names a CVE gets that CVE's EPSS row as one more
// query, a finding with no CVE stays unscored, and the findings come out in
// the Writer's order.
func TestKEVFindingGetsItsEPSSRow(t *testing.T) {
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	f4, ok := kept(d, "F4")
	if !ok {
		t.Fatal("F4 missing")
	}
	q := epssQuery(f4)
	if q == nil || q.SQL != "SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'" {
		t.Fatalf("F4 EPSS %#v", q)
	}
	if len(q.Rows) != 1 || q.Rows[0][0] != "CVE-2026-0001" || q.Rows[0][1] != "0.50000" || q.Rows[0][2] != "0.90000" || q.Rows[0][3] != "2026-10-20" {
		t.Fatalf("row %#v", q.Rows)
	}
	f2, ok := kept(d, "F2")
	if !ok || epssQuery(f2) == nil || epssQuery(f2).Rows[0][1] != "0.50000" {
		t.Fatal("F2 has no EPSS row")
	}
	f1, ok := kept(d, "F1")
	if !ok || epssQuery(f1) != nil {
		t.Fatal("F1 must stay unscored")
	}
	var got []string
	for _, c := range d.Findings {
		got = append(got, c.ID)
	}
	want := []string{"F2", "F4", "osv:evil.example/leftpad", "F1"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order %s", strings.Join(got, ","))
	}
}

// Mutated field: the EPSS row of F4's CVE is removed from tables.sql. F4 stays,
// bound to the change through its KEV row, and no score is attached.
func TestCVEWithNoEPSSRowAttachesNothing(t *testing.T) {
	dir := copyFixture(t)
	sql := filepath.Join(dir, "tables.sql")
	b, err := os.ReadFile(sql)
	if err != nil {
		t.Fatal(err)
	}
	const row = "INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-0001', '0.50000', '0.90000', '2026-10-20');\n"
	if bytes.Count(b, []byte(row)) != 1 {
		t.Fatal("the EPSS row of CVE-2026-0001 is not in tables.sql once")
	}
	if err := os.WriteFile(sql, bytes.Replace(b, []byte(row), nil, 1), 0o644); err != nil {
		t.Fatal(err)
	}
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	f4, ok := kept(d, "F4")
	if !ok {
		t.Fatal("F4 should stay: its KEV row is about the change")
	}
	if len(f4.Evidence) != 1 || epssQuery(f4) != nil {
		t.Fatalf("F4 evidence %#v", f4.Evidence)
	}
	if !traceContains(t, d.TracePath, "F4", "no EPSS row") {
		t.Fatal("trace did not say no EPSS row")
	}
}

func epssQuery(c candidate.Candidate) *candidate.QueryEvidence {
	for _, e := range c.Evidence {
		if e.Query == nil {
			continue
		}
		cols := map[string]bool{}
		for _, name := range e.Query.Columns {
			cols[strings.ToLower(name)] = true
		}
		if cols["cve_id"] && cols["epss"] && cols["percentile"] && cols["score_date"] {
			return e.Query
		}
	}
	return nil
}

func traceContains(t *testing.T, path, id, substr string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var e struct {
			FindingID string `json:"finding_id"`
			Text      string `json:"text"`
		}
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		if e.FindingID == id && strings.Contains(e.Text, substr) {
			return true
		}
	}
	return false
}

// buildSample builds a sample pipeline that fails the test if it reads any
// setting: sample mode must run with no environment.
func buildSample(t *testing.T, logDir string, run testrun.Adapter) *Pipeline {
	t.Helper()
	p, err := Build(Options{
		Sample:  true,
		LogDir:  logDir,
		TestRun: run,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// A written test that passes on the change and on the base is not evidence.
// The trace names that reason. The runner is the only thing that runs.
func TestWrittenTestThatPassesOnBothIsNamedInTheTrace(t *testing.T) {
	fx, err := fixture.Load(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	body := "package billing\n\nimport \"testing\"\n\nfunc TestExtra(t *testing.T) {\n\tif Discounted(100, 20) != 80 {\n\t\tt.Fatal(Discounted(100, 20))\n\t}\n}\n"
	raw, err := json.Marshal(map[string]string{
		"path": "billing/extra_test.go", "content": body, "name": "TestExtra",
	})
	if err != nil {
		t.Fatal(err)
	}
	run := &outcomeRun{out: testrun.Outcome{Found: true, PassesOnBase: true, Path: "billing/extra_test.go", Name: "TestExtra"}}
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), Generator: &proverSwap{inner: inner, response: string(raw)}, TestRun: run,
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
	if run.calls != 1 || run.last.Content != body {
		t.Fatalf("calls %d content %q", run.calls, run.last.Content)
	}
	if !traceSays(t, d.TracePath, "F1", "evidence", "written test passed on the change") {
		b, readErr := os.ReadFile(d.TracePath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		t.Fatalf("trace:\n%s", b)
	}
	for _, c := range d.Findings {
		if c.ID != "F1" {
			continue
		}
		for _, e := range c.Evidence {
			if e.Test != nil {
				t.Fatal("F1 kept a test that passed on both")
			}
		}
	}
}

type proverSwap struct {
	inner    generator.Generator
	response string
}

func (g *proverSwap) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "prover" && strings.Contains(prompt, findingMarker("F1")) {
		return g.response, nil
	}
	return g.inner.Complete(ctx, agent, prompt)
}

type outcomeRun struct {
	calls int
	last  testrun.Request
	out   testrun.Outcome
}

func (o *outcomeRun) Run(_ context.Context, req testrun.Request) (testrun.Outcome, error) {
	o.calls++
	o.last = req
	return o.out, nil
}

func kept(d candidate.Draft, id string) (candidate.Candidate, bool) {
	for _, c := range d.Findings {
		if c.ID == id {
			return c, true
		}
	}
	return candidate.Candidate{}, false
}

// fixturePath is the discount sample fixture under the module root.
func fixturePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "fixture", "discount")
}

// repoRoot walks up from the working directory to the proven module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

// copyFixture copies the discount fixture to a temporary directory, so a test
// can change one field of it.
func copyFixture(t *testing.T) string {
	t.Helper()
	src := fixturePath(t)
	dst := filepath.Join(t.TempDir(), "fixture")
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}
