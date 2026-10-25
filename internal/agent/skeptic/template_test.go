package skeptic

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestRefutingQueryIsATemplateAndValues proves ADR-0032 point 1 for the
// Skeptic: a good template and values are kept as the refuting query, and the
// prompt names every template and asks for no SQL.
func TestRefutingQueryIsATemplateAndValues(t *testing.T) {
	risk := candidate.Candidate{Risk: candidate.Risk{ID: "F1", Title: "x"}}
	got, err := Attempt(context.Background(), risk, textGen{
		`{"refuted":false,"reason":"check KEV","path":"","name":"","template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	q := got.Refutation.Query
	if q.Template != "kev-by-cve" || q.Params["cve_id"] != "CVE-2021-44228" || got.Refutation.Reason != "check KEV" {
		t.Fatalf("%#v", got.Refutation)
	}
	p := Prompt("F1", "x", nil)
	for _, tpl := range querytpl.Catalogue() {
		if !strings.Contains(p, tpl.Name) {
			t.Errorf("the prompt does not name template %s", tpl.Name)
		}
	}
	if !strings.Contains(p, "Do not add a command or SQL") {
		t.Error("the prompt does not forbid SQL")
	}
}

// TestAnswerThatIsNotATemplateProposesNoQuery proves ADR-0032 points 1 and 2:
// an answer that carries SQL text, a template that is not on the list and a
// value outside its format each propose no query, and the reason names the
// refusal without repeating the model's text. The test part of the answer is
// kept. The mutated field is the query part of the answer.
func TestAnswerThatIsNotATemplateProposesNoQuery(t *testing.T) {
	risk := candidate.Candidate{Risk: candidate.Risk{ID: "F1", Title: "x"}}
	cases := map[string]struct{ answer, reason string }{
		"sql text":         {`"sql":"SELECT * FROM t"`, querytpl.ReasonModelSQL},
		"unknown template": {`"template":"SELECT 1","params":{}`, querytpl.ReasonUnknownTemplate},
		"quote in a value": {`"template":"epss-by-cve","params":{"cve_id":"CVE-2021-44228' OR '1'='1"}`, querytpl.ReasonBadValue},
		"backslash":        {`"template":"osv-by-package","params":{"ecosystem":"Go","package":"a\\"}`, querytpl.ReasonBadValue},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := Attempt(context.Background(), risk, textGen{
				`{"refuted":true,"reason":"r","path":"a_test.go","name":"TestA",` + c.answer + `}`,
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := got.Refutation
			if !r.Query.Empty() {
				t.Fatalf("a query was kept: %#v", r.Query)
			}
			if !strings.HasPrefix(r.Reason, "refused: "+c.reason) || strings.Contains(r.Reason, "SELECT") || strings.Contains(r.Reason, "OR '1'") {
				t.Fatalf("reason %q, want %q and no model text", r.Reason, c.reason)
			}
			if r.TestPath != "a_test.go" || r.TestName != "TestA" {
				t.Fatalf("the test part was dropped: %#v", r)
			}
		})
	}
}
