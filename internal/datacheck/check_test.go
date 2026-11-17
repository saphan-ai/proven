package datacheck

import (
	"context"
	"fmt"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/report"
)

type fakeQuery struct {
	fail  string
	calls int
}

func (q *fakeQuery) Run(_ context.Context, r querytpl.Request, _ querytpl.Relations) (query.Result, error) {
	q.calls++
	if r.Params[querytpl.ParamPackage] != "example-package" && r.Params[querytpl.ParamCVEID] != "CVE-2026-4242" {
		return query.Result{}, fmt.Errorf("not bound to this change")
	}
	if r.Template == q.fail {
		return query.Result{}, fmt.Errorf("not available")
	}
	result := query.Result{SQL: "a fixed registered query", Form: "sql", Engine: "fixture", Snapshot: "fixture-content-id"}
	switch r.Template {
	case querytpl.MalByPackage:
		result.Columns = []string{"id", "ecosystem", "package", "summary"}
		result.Rows = [][]string{{"MAL-example", "npm", "example-package", "Package-level fixture"}}
	case querytpl.ScorecardByPackage:
		result.Columns = []string{"ecosystem", "package", "maintained", "signed", "release_age_days"}
		result.Rows = [][]string{{"npm", "example-package", "0", "false", "3"}}
	case querytpl.SPDXByPackage:
		result.Columns = []string{"ecosystem", "package", "licence"}
		result.Rows = [][]string{{"npm", "example-package", "GPL-3.0-only"}}
	case querytpl.CVSSByCVE:
		result.Columns = []string{"cve_id", "cvss_score", "severity", "source"}
		result.Rows = [][]string{{"CVE-2026-4242", "8.1", "HIGH", "fixture"}}
	default:
		return query.Result{}, fmt.Errorf("unexpected template")
	}
	return result, nil
}
func TestAllBindableSourceRulesRunWithoutAPlannerAndRemainObservations(t *testing.T) {
	files := []change.File{{Path: "package.json", Base: []byte(`{"dependencies":{}}`), Change: []byte(`{"dependencies":{"example-package":"1.2.3"}}`)}}
	matches := []osv.Match{{Ecosystem: "npm", Package: "example-package", Version: "1.2.3", Advisories: []osv.Advisory{{Affected: true, Row: osv.Row{ID: "GHSA-example", Aliases: "CVE-2026-4242"}}}}}
	for _, fail := range []string{"", querytpl.SPDXByPackage} {
		q := &fakeQuery{fail: fail}
		out := Check(context.Background(), files, matches, q, querytpl.Relations{})
		if len(out) != 4 || q.calls != 4 {
			t.Fatalf("%+v calls=%d", out, q.calls)
		}
		for _, o := range out {
			if o.Template == fail {
				if o.Complete || o.Issue == "" {
					t.Fatal("lost failed source")
				}
			} else if !o.Complete || !o.Matched || o.Snapshot == "" || len(o.Rows) == 0 {
				t.Fatalf("%+v", o)
			}
		}
		r := report.Result{SourceChecks: out}
		if len(r.Findings) != 0 || len(r.ProofCounts) != 0 {
			t.Fatal("observations became code defects")
		}
		if (r.CompletionStatus() == "complete") != (fail == "") {
			t.Fatal("source failure did not affect completion")
		}
	}
}
