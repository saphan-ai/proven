package pipeline

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/testrun"
)

// askedRunner is an evidence seam that keeps the name of each template it is
// asked to run, so a test can say that a store was never asked.
type askedRunner struct{ asked []string }

func (r *askedRunner) Run(_ context.Context, req querytpl.Request, _ querytpl.Relations) (query.Result, error) {
	r.asked = append(r.asked, req.Template)
	return query.Result{}, nil
}

// skepticAsksForEPSS is a candidate whose Skeptic proposed the EPSS template.
func skepticAsksForEPSS() candidate.Candidate {
	var c candidate.Candidate
	c.Refutation.Query = querytpl.Request{
		Template: querytpl.EPSSByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2026-0001"},
	}
	return c
}

// TestSkepticRequestOutsideItsScopeAsksNoStore proves ADR-0013's evidence
// scope for the Skeptic: a template the Skeptic's scope leaves out is refused
// with a named reason, the store is never asked, the proposal is cleared, and
// the refutation run returns no error, so the review goes on. With the nil
// scope the same request reaches the store, which is the behaviour before
// scopes existed. Mutated field: the Skeptic's scope.
func TestSkepticRequestOutsideItsScopeAsksNoStore(t *testing.T) {
	p := &Pipeline{}

	open := &askedRunner{}
	c := skepticAsksForEPSS()
	if err := p.runRefutation(context.Background(), &c, open, testrun.Off{}, nil); err != nil {
		t.Fatal(err)
	}
	if len(open.asked) != 1 || open.asked[0] != querytpl.EPSSByCVE {
		t.Fatalf("with no scope the store was asked %v", open.asked)
	}

	closed := &askedRunner{}
	c = skepticAsksForEPSS()
	err := p.runRefutation(context.Background(), &c, closed, testrun.Off{}, querytpl.Scope{querytpl.KEVByCVE})
	if err != nil {
		t.Fatalf("a refusal failed the review: %v", err)
	}
	if len(closed.asked) != 0 {
		t.Fatalf("the store was asked for %v outside the scope", closed.asked)
	}
	want := "refused: " + querytpl.ReasonOutOfScope + ": epss-by-cve for the skeptic"
	if c.Refutation.Reason != want || !c.Refutation.Query.Empty() || c.Refutation.QueryCannotRun {
		t.Fatalf("reason %q, query %#v, cannot-run %t", c.Refutation.Reason, c.Refutation.Query, c.Refutation.QueryCannotRun)
	}
}

// reviewScoped runs the sample review with the Skeptic answering every finding
// with body and held to the scope given. An empty match is used because a
// narrower scope changes the prompt the fixture keys its answers on. The
// Investigator keeps every template for the same reason.
func reviewScoped(t *testing.T, body string, skeptic querytpl.Scope) candidate.Draft {
	t.Helper()
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: &answerFor{inner: inner, agent: "skeptic", match: "", body: body},
		TestRun:   testrun.NewSample(fx.Tests),
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{
		Target: fx.Root, SkepticScope: skeptic,
	})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestSkepticIsNarrowerThanTheInvestigator is the sample scenario of ADR-0013's
// evidence scope: the Investigator may ask every template and the Skeptic one.
// A Skeptic that asks for another template is refused by
// name in the trace and the finding stays; a Skeptic that asks for its own
// template is run. Mutated field: the Skeptic's template, with the scope held.
func TestSkepticIsNarrowerThanTheInvestigator(t *testing.T) {
	narrow := querytpl.Scope{querytpl.KEVByCVE}

	outside := `{"refuted":true,"reason":"look at EPSS","path":"","name":"","template":"epss-by-cve","params":{"cve_id":"CVE-2026-0001"}}`
	d := reviewScoped(t, outside, narrow)
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("a refused request dropped the finding")
	}
	if !traceHas(t, d.TracePath, "refused: "+querytpl.ReasonOutOfScope+": epss-by-cve for the skeptic") {
		t.Fatal("the trace does not name the out-of-scope refusal")
	}

	inside := `{"refuted":true,"reason":"look at KEV","path":"","name":"","template":"kev-by-cve","params":{"cve_id":"CVE-2026-0001"}}`
	d = reviewScoped(t, inside, narrow)
	if traceHas(t, d.TracePath, querytpl.ReasonOutOfScope) {
		t.Fatal("a template in the Skeptic's scope was refused")
	}
	if !traceSays(t, d.TracePath, "F1", "refutation", "look at KEV") {
		t.Fatal("the refutation inside the scope did not run")
	}
}
