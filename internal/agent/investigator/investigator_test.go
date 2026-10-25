package investigator

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/source"
)

// scriptGen is a generator that always returns text.
type scriptGen struct{ text string }

func (s scriptGen) Complete(context.Context, string, string) (string, error) {
	return s.text, nil
}

// recorder is a query adapter that keeps each statement it is asked to run and
// answers every one with the same one-row result.
type recorder struct {
	sql  []string
	rows [][]string
}

func (r *recorder) Query(_ context.Context, sql string) (query.Result, error) {
	r.sql = append(r.sql, sql)
	return query.Result{SQL: sql, Engine: "mem", Columns: []string{"cve_id"}, Rows: r.rows}, nil
}

var risk = candidate.Candidate{
	Risk: candidate.Risk{ID: "F1", Title: "Dependency is on the KEV list", Path: "go.mod"},
}

// TestNilQueryIsANamedRefusal proves Investigate refuses a nil query adapter
// with the same named failure Execute returns, instead of panicking (F2). The
// template is a good one, so the refusal is the engine's absence.
func TestNilQueryIsANamedRefusal(t *testing.T) {
	_, err := Investigate(context.Background(), risk, "",
		scriptGen{text: `{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`}, nil, querytpl.Relations{})
	d, ok := service.As(err)
	if !ok || d.Service != "query" || d.Reason != "no query engine" {
		t.Fatalf("err %v", err)
	}
}

// TestInvestigatorRunsTheQueryProvenBuilds proves ADR-0032 point 3: the model
// gives a template and a value, and the SQL that runs, and is kept as the
// evidence, is the text Proven built from them with the review's view name.
func TestInvestigatorRunsTheQueryProvenBuilds(t *testing.T) {
	rec := &recorder{rows: [][]string{{"CVE-2021-44228"}}}
	got, err := Investigate(context.Background(), risk, "snap",
		scriptGen{text: `{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`},
		rec, querytpl.Relations{KEVView: "my_schema.my_kev"})
	if err != nil {
		t.Fatal(err)
	}
	const want = "SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM my_schema.my_kev WHERE cve_id = 'CVE-2021-44228'"
	if len(rec.sql) != 1 || rec.sql[0] != want {
		t.Fatalf("sent %q, want %q", rec.sql, want)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Query.SQL != want || got.QueryNote != "" {
		t.Fatalf("%#v", got)
	}
}

// TestInvestigatorSendsNothingForAnAnswerThatIsNotATemplate proves ADR-0032
// points 1 and 2: an answer that carries SQL text, a template that is not on
// the list, and a value outside its format each send no query, add no
// evidence, and leave a QueryNote that names the reason and does not repeat
// the model's text. The mutated field is the model's answer.
func TestInvestigatorSendsNothingForAnAnswerThatIsNotATemplate(t *testing.T) {
	cases := map[string]struct {
		answer string
		reason string
	}{
		"sql text": {`{"sql":"SELECT * FROM px_evidence.px_kev_vulns"}`, querytpl.ReasonModelSQL},
		"sql beside a template": {`{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"},"sql":"DELETE FROM t"}`,
			querytpl.ReasonModelSQL},
		"sql as the template name": {`{"template":"SELECT 1","params":{}}`, querytpl.ReasonUnknownTemplate},
		"unknown template":         {`{"template":"drop-everything","params":{}}`, querytpl.ReasonUnknownTemplate},
		"quote in a value":         {`{"template":"kev-by-cve","params":{"cve_id":"CVE-2026-0001' OR cve_id = 'CVE-2026-0002"}}`, querytpl.ReasonBadValue},
		"backslash in a value":     {`{"template":"osv-by-package","params":{"ecosystem":"npm","package":"a\\"}}`, querytpl.ReasonBadValue},
		"control byte in a value":  {`{"template":"epss-by-cve","params":{"cve_id":"CVE-2021-44228\u001b"}}`, querytpl.ReasonBadValue},
		"missing value":            {`{"template":"kev-by-cve","params":{}}`, querytpl.ReasonMissingParam},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			rec := &recorder{rows: [][]string{{"x"}}}
			got, err := Investigate(context.Background(), risk, "", scriptGen{text: c.answer}, rec, querytpl.Relations{})
			if err != nil {
				t.Fatal(err)
			}
			if len(rec.sql) != 0 || len(got.Evidence) != 0 {
				t.Fatalf("sent %q, evidence %#v", rec.sql, got.Evidence)
			}
			if !strings.HasPrefix(got.QueryNote, "refused: "+c.reason) {
				t.Fatalf("note %q, want reason %q", got.QueryNote, c.reason)
			}
			if strings.Contains(got.QueryNote, "SELECT") || strings.Contains(got.QueryNote, "DELETE") || strings.Contains(got.QueryNote, "OR cve_id") {
				t.Fatalf("the note repeats the model's text: %q", got.QueryNote)
			}
		})
	}
}

// TestInvestigatorProposingNoQueryIsNotARefusal proves that an empty template
// means no query and no note, so the trace does not call a quiet answer a
// refusal.
func TestInvestigatorProposingNoQueryIsNotARefusal(t *testing.T) {
	rec := &recorder{}
	got, err := Investigate(context.Background(), risk, "", scriptGen{text: `{"template":"","params":{}}`}, rec, querytpl.Relations{})
	if err != nil || len(rec.sql) != 0 || got.QueryNote != "" || len(got.Evidence) != 0 {
		t.Fatalf("%v %q %#v", err, rec.sql, got)
	}
}

// TestPromptNamesEveryTemplateAndAsksForNoSQL proves the prompt is written
// from the closed list: every template name is in it, and it tells the model
// not to write SQL.
func TestPromptNamesEveryTemplateAndAsksForNoSQL(t *testing.T) {
	p := Prompt("F1", "title", "go.mod")
	for _, tpl := range querytpl.Catalogue() {
		if !strings.Contains(p, tpl.Name) {
			t.Errorf("the prompt does not name template %s", tpl.Name)
		}
	}
	if !strings.Contains(p, "Do not write SQL") {
		t.Error("the prompt does not forbid SQL")
	}
}

// TestExecuteRefusesBeforeItSends proves that the Skeptic's request is checked
// again here: a request outside its format is a *querytpl.Refusal and nothing
// is sent.
func TestExecuteRefusesBeforeItSends(t *testing.T) {
	rec := &recorder{}
	_, err := Execute(context.Background(), rec,
		querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "x' OR '1'='1"}}, querytpl.Relations{})
	if err == nil || len(rec.sql) != 0 {
		t.Fatalf("err %v, sent %q", err, rec.sql)
	}
}

// refusingAdapter is a registered source whose builder refuses with a plain
// error, as an organisation's builder would.
type refusingAdapter struct{}

func (refusingAdapter) ID() string { return "refusing-probe" }
func (refusingAdapter) Load() source.Load {
	return source.Load{Origin: "test", Refresh: "never", Table: "t"}
}
func (refusingAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{Name: "refusing-by-name", Answer: "a row", Params: []source.Param{{Name: "n", Format: "any", Match: regexp.MustCompile(`^[a-z]+$`)}}},
		Build: func(map[string]string, source.Rel) (string, error) {
			return "", errors.New("refusing-by-name: n is outside its format")
		},
	}}
}
func (refusingAdapter) Rule(in source.Evidence) (source.Evidence, bool) { return in, len(in.Rows) > 0 }

// TestRefusalByARegisteredBuilderIsAQueryNote proves a value a registered
// builder refuses becomes the candidate's QueryNote and adds no evidence; the
// review does not fail. The mutated field is the builder's error.
func TestRefusalByARegisteredBuilderIsAQueryNote(t *testing.T) {
	if err := source.Register(refusingAdapter{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Unregister("refusing-probe") })
	rec := &recorder{rows: [][]string{{"x"}}}
	got, err := Investigate(context.Background(), risk, "",
		scriptGen{text: `{"template":"refusing-by-name","params":{"n":"x"}}`}, rec, querytpl.Relations{})
	if err != nil {
		t.Fatalf("the review failed on a refused value: %v", err)
	}
	if !strings.HasPrefix(got.QueryNote, "refused: ") || len(got.Evidence) != 0 || len(rec.sql) != 0 {
		t.Fatalf("note %q evidence %d sent %d", got.QueryNote, len(got.Evidence), len(rec.sql))
	}
}

// ruleDropsAdapter is a registered source whose rule keeps no evidence even
// when the query returns rows.
type ruleDropsAdapter struct{}

func (ruleDropsAdapter) ID() string { return "rule-drops-probe" }
func (ruleDropsAdapter) Load() source.Load {
	return source.Load{Origin: "test", Refresh: "never", Table: "t"}
}
func (ruleDropsAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{Name: "rule-drops-by-name", Answer: "a row", Params: []source.Param{{Name: "n", Format: "any", Match: regexp.MustCompile(`^[a-z]+$`)}}},
		Build:    func(map[string]string, source.Rel) (string, error) { return "SELECT 1", nil },
	}}
}
func (ruleDropsAdapter) Rule(source.Evidence) (source.Evidence, bool) {
	return source.Evidence{}, false
}

// TestOwningAdaptersRuleDecidesEvidence proves the rule of the adapter that
// owns the template decides whether rows become evidence: a rule that returns
// false on a non-empty row set leaves the candidate with no evidence. The
// mutated field is the rule's verdict.
func TestOwningAdaptersRuleDecidesEvidence(t *testing.T) {
	if err := source.Register(ruleDropsAdapter{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Unregister("rule-drops-probe") })
	rec := &recorder{rows: [][]string{{"x"}}}
	got, err := Investigate(context.Background(), risk, "",
		scriptGen{text: `{"template":"rule-drops-by-name","params":{"n":"x"}}`}, rec, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.sql) != 1 || len(got.Evidence) != 0 {
		t.Fatalf("sent %d, evidence %d; the rule was not consulted", len(rec.sql), len(got.Evidence))
	}
}
