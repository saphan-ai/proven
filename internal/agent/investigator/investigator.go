// Package investigator turns a risk into a query and runs it.
//
// The Investigator gets each risk from the Planner and asks the model for one
// SQL query over the evidence tables behind Zetaris (ADR-0002, ADR-0004). A
// query and the rows it returned are evidence the Judge can keep; the SQL is
// kept with them so any reader can run it again.
package investigator

import (
	"context"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

// querySpec is the JSON the investigator asks a model to return.
// The task tells the model not to read an absent row as "not exploited": a
// missing KEV row only means the data does not say.
var querySpec = answer.Spec{
	Task:    "Write one SQL query that would confirm this risk. For a CVE query px_evidence.px_kev_vulns. For an EPSS score query px_evidence.px_epss_cve_rows. For an advisory id query px_evidence.px_osv_id. Do not treat a missing row as not exploited.",
	Shape:   "one JSON object. It has sql (string).",
	Example: `{"sql":"SELECT 1"}`,
}

// Prompt asks for the SQL for one risk.
func Prompt(id, title, path string) string {
	return answer.Instruct(querySpec, "id: "+id+"\ntitle: "+title+"\npath: "+path+"\n")
}

// row is the query as the model returns it.
type row struct {
	SQL string `json:"sql"`
}

// Execute runs a query the skeptic named. It does not ask a model.
func Execute(ctx context.Context, q query.Adapter, sql string) (query.Result, error) {
	if q == nil {
		return query.Result{}, &service.Down{Service: "query", Reason: "no query engine"}
	}
	return q.Query(ctx, sql)
}

// Investigate asks for SQL and runs it. A query with no rows adds no evidence.
func Investigate(ctx context.Context, c candidate.Candidate, snapshot string, gen generator.Generator, q query.Adapter) (candidate.Candidate, error) {
	var got row
	if err := answer.Read(ctx, gen, "investigator", Prompt(c.ID, c.Title, c.Path), querySpec, &got); err != nil {
		return c, err
	}
	// No query proposed means no evidence, not a failure.
	if got.SQL == "" {
		return c, nil
	}
	res, err := q.Query(ctx, got.SQL)
	if err != nil {
		return c, err
	}
	// ADR-0002: a query with no rows proves nothing, so it adds no evidence.
	if len(res.Rows) == 0 {
		return c, nil
	}
	// The engine's own snapshot, when it reports one, names the data the rows came from more exactly.
	snap := snapshot
	if res.Snapshot != "" {
		snap = res.Snapshot
	}
	c.Evidence = append(c.Evidence, candidate.Evidence{
		Kind: candidate.KindQuery,
		Query: &candidate.QueryEvidence{
			SQL: res.SQL, Engine: res.Engine, Snapshot: snap,
			Columns: res.Columns, Rows: res.Rows,
		},
	})
	return c, nil
}
