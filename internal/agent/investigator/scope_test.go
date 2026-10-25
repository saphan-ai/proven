package investigator

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// TestPromptListsOnlyTheScopesTemplates proves the Investigator is told of the
// templates its scope allows and of no other, and that the nil scope tells it
// of every template. Mutated field: the scope.
func TestPromptListsOnlyTheScopesTemplates(t *testing.T) {
	narrow := Prompt("F1", "title", "go.mod", querytpl.Scope{querytpl.KEVByCVE})
	if !strings.Contains(narrow, querytpl.KEVByCVE+" (") {
		t.Fatal("the allowed template is not in the prompt")
	}
	if strings.Contains(narrow, querytpl.EPSSByCVE) || strings.Contains(narrow, querytpl.OSVByPackage) {
		t.Fatalf("the prompt names a template outside the scope:\n%s", narrow)
	}
	whole := Prompt("F1", "title", "go.mod", nil)
	for _, tpl := range querytpl.Catalogue() {
		if !strings.Contains(whole, tpl.Name+" (") {
			t.Errorf("the unscoped prompt lacks %s", tpl.Name)
		}
	}
}

// TestInvestigateRefusesOutOfScopeBeforeAnyStore proves a model that names a
// template outside the Investigator's scope asks no store, the review goes on
// with no error, and the candidate's note names the refusal. Mutated field:
// the scope, with the model's answer held the same.
func TestInvestigateRefusesOutOfScopeBeforeAnyStore(t *testing.T) {
	rec := &recorder{rows: [][]string{{"CVE-2021-44228"}}}
	answer := `{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`
	got, err := Investigate(context.Background(), risk, "", scriptGen{text: answer}, rec,
		querytpl.Relations{}, querytpl.Scope{querytpl.EPSSByCVE})
	if err != nil {
		t.Fatalf("a refusal failed the review: %v", err)
	}
	if len(rec.sql) != 0 {
		t.Fatalf("the store was asked: %q", rec.sql)
	}
	want := "refused: " + querytpl.ReasonOutOfScope + ": kev-by-cve for the investigator"
	if got.QueryNote != want || len(got.Evidence) != 0 {
		t.Fatalf("note %q, evidence %d", got.QueryNote, len(got.Evidence))
	}
	// The same answer inside the scope does run.
	if _, err := Investigate(context.Background(), risk, "", scriptGen{text: answer}, rec,
		querytpl.Relations{}, querytpl.Scope{querytpl.KEVByCVE}); err != nil || len(rec.sql) != 1 {
		t.Fatalf("a template in scope did not run: %v, %d", err, len(rec.sql))
	}
}

// TestExecuteRefusesOutOfScopeBeforeAnyStore proves the Skeptic's request is
// held to the Skeptic's scope where it is run: the store is not asked, with or
// without a store up, and the refusal names the template and the role.
// Mutated field: the scope.
func TestExecuteRefusesOutOfScopeBeforeAnyStore(t *testing.T) {
	rec := &recorder{rows: [][]string{{"x"}}}
	req := querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}
	for name, q := range map[string]query.Runner{"store up": rec, "no store": nil} {
		_, err := Execute(context.Background(), q, req, querytpl.Relations{}, querytpl.Scope{querytpl.EPSSByCVE})
		ref, ok := err.(*querytpl.Refusal)
		if !ok || ref.Reason != querytpl.ReasonOutOfScope || ref.Detail != "kev-by-cve for the skeptic" {
			t.Errorf("%s: %v", name, err)
		}
		if _, down := service.As(err); down {
			t.Errorf("%s: the refusal was reported as a service down", name)
		}
	}
	if len(rec.sql) != 0 {
		t.Fatalf("the store was asked: %q", rec.sql)
	}
}
