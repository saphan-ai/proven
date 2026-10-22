package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/saphan-ai/proven/internal/testrun"
)

func TestSampleModeOpensNoNetwork(t *testing.T) {
	p := buildSample(t, t.TempDir(), nil)
	if p.Guard() == nil || !p.Guard().Armed() {
		t.Fatal("sample mode did not arm a refusing client")
	}
	if _, err := p.Review(context.Background(), Request{Target: fixturePath(t)}); err != nil {
		t.Fatal(err)
	}
	if p.Guard().Dials() != 0 {
		t.Fatalf("dials %d", p.Guard().Dials())
	}
}

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

// Mutated fixture field: tables.sql. An INSERT is added for package billing/note,
// the row F3's recorded query selects.
func TestNegativeControl_F3RecordedQueryRow(t *testing.T) {
	dir := copyFixture(t)
	sql := filepath.Join(dir, "tables.sql")
	f, err := os.OpenFile(sql, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("INSERT INTO advisories VALUES ('billing/note', 'change', 'CVE-2026-0099', 'false');\n"); err != nil {
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
	f3, ok := kept(d, "F3")
	if !ok {
		t.Fatal("F3 should appear once its query returns a row")
	}
	if len(f3.Evidence) != 1 || f3.Evidence[0].Query == nil || len(f3.Evidence[0].Query.Rows) != 1 {
		t.Fatalf("%#v", f3.Evidence)
	}
	if f3.Evidence[0].Query.Rows[0][0] != "billing/note" {
		t.Fatalf("row %#v", f3.Evidence[0].Query.Rows)
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

// Mutated field: the version column of the advisories sample row.
// It changes from 1.2.3, the version the advisory covers, to 9.9.9.
func TestNegativeControl_VersionTheAdvisoryDoesNotCover(t *testing.T) {
	dir := copyFixture(t)
	path := filepath.Join(dir, "tables.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const covered = "'1.2.3'"
	if bytes.Count(b, []byte(covered)) != 1 {
		t.Fatalf("version occurrences %d", bytes.Count(b, []byte(covered)))
	}
	next := bytes.Replace(b, []byte(covered), []byte("'9.9.9'"), 1)
	if err := os.WriteFile(path, next, 0o644); err != nil {
		t.Fatal(err)
	}
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := kept(d, "F2"); ok {
		t.Fatal("F2 must be absent when the advisory does not cover version 1.2.3")
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("F1 must stay")
	}
	if _, ok := kept(d, "F4"); !ok {
		t.Fatal("F4 is keyed by cve_id, which this mutation does not change")
	}
}

// Mutated field: cve_id on the EPSS sample row that was CVE-2026-0001.
// After the change the CVE is absent from the EPSS side of the inner join.
func TestNegativeControl_CVEAbsentFromEPSSIsNotCalledNotExploited(t *testing.T) {
	dir := copyFixture(t)
	const title = "Join of KEV and EPSS for CVE-2026-0001"
	const sqlText = "SELECT k.cve_id, k.vendor_project, k.product, k.date_added, e.epss, e.percentile FROM px_evidence.px_kev_vulns k JOIN px_evidence.px_epss_cve_rows e ON k.cve_id = e.cve_id WHERE k.cve_id = 'CVE-2026-0001'"
	if err := addJoinFinding(dir, title, sqlText); err != nil {
		t.Fatal(err)
	}
	p := buildSample(t, t.TempDir(), nil)
	before, err := p.Review(context.Background(), Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	f5, ok := kept(before, "F5")
	if !ok || len(f5.Evidence) != 1 || f5.Evidence[0].Query == nil || len(f5.Evidence[0].Query.Rows) != 1 {
		t.Fatal("the join must keep F5 while the CVE is on both sides")
	}
	if f5.Evidence[0].Query.Rows[0][0] != "CVE-2026-0001" {
		t.Fatalf("row %#v", f5.Evidence[0].Query.Rows)
	}
	tables := filepath.Join(dir, "tables.sql")
	b, err := os.ReadFile(tables)
	if err != nil {
		t.Fatal(err)
	}
	from := []byte("INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-0001'")
	to := []byte("INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-0000-0000'")
	if bytes.Count(b, from) != 1 {
		t.Fatal("the EPSS row for CVE-2026-0001 was not found once")
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
		t.Fatal("F5 must be absent when the CVE is missing from EPSS")
	}
	if !traceSays(t, after.TracePath, "F5", "verdict", "dropped: no evidence") {
		t.Fatal("F5 was not dropped for lack of evidence")
	}
	raw, err := os.ReadFile(after.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(bytes.ToLower(raw), []byte("not exploited")) {
		t.Fatal("the review read a missing join row as not exploited")
	}
	for _, c := range after.Findings {
		if strings.Contains(strings.ToLower(c.Message), "not exploited") || strings.Contains(strings.ToLower(c.Title), "not exploited") {
			t.Fatalf("finding %s says not exploited", c.ID)
		}
	}
}

func addJoinFinding(dir, title, sqlText string) error {
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
		"message":    "The inner join returned a row for the CVE.",
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
		}{Agent: "investigator", Prompt: investigator.Prompt("F5", title, "go.mod"), Response: `{"sql":"` + sqlText + `"}`},
		struct {
			Agent    string `json:"agent"`
			Prompt   string `json:"prompt"`
			Response string `json:"response"`
		}{Agent: "prover", Prompt: prover.Prompt("F5", title), Response: `{"path":"","name":"","command":""}`},
		struct {
			Agent    string `json:"agent"`
			Prompt   string `json:"prompt"`
			Response string `json:"response"`
		}{Agent: "skeptic", Prompt: skeptic.Prompt("F5", title), Response: `{"refuted":false,"reason":"no refutation"}`},
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

type dirProbe struct {
	req testrun.Request
}

func (d *dirProbe) Run(ctx context.Context, req testrun.Request) (testrun.Outcome, error) {
	d.req = req
	return testrun.Outcome{}, nil
}

func TestCheckoutRunFillsEmptyDirectories(t *testing.T) {
	probe := &dirProbe{}
	a := checkoutRun{inner: probe, base: "/tmp/base", change: "/tmp/change"}
	if _, err := a.Run(context.Background(), testrun.Request{Command: "go test ./billing"}); err != nil {
		t.Fatal(err)
	}
	if probe.req.BaseDir != "/tmp/base" || probe.req.ChangeDir != "/tmp/change" {
		t.Fatalf("%#v", probe.req)
	}
	if _, err := a.Run(context.Background(), testrun.Request{Command: "go test", BaseDir: "/already", ChangeDir: "/set"}); err != nil {
		t.Fatal(err)
	}
	if probe.req.BaseDir != "/already" || probe.req.ChangeDir != "/set" {
		t.Fatalf("%#v", probe.req)
	}
}

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

func kept(d candidate.Draft, id string) (candidate.Candidate, bool) {
	for _, c := range d.Findings {
		if c.ID == id {
			return c, true
		}
	}
	return candidate.Candidate{}, false
}

func fixturePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "fixture", "discount")
}

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
