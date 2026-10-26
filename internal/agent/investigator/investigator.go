// Package investigator turns a risk into a query and runs it.
//
// The Investigator gets each risk from the Planner and asks the model to pick
// one query template and give its values (ADR-0032). The model never writes
// a query: Proven checks each value against its format and passes the request
// to the evidence seam. The store builds the form of its own language
// (ADR-0047) over the evidence tables (ADR-0002, ADR-0004). A query and the
// rows it returned are evidence the Judge can keep; the form and the text
// that ran are kept with them so any reader can run the query again.
package investigator

import (
	"context"
	"errors"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/source"
)

// querySpec is the JSON the investigator asks a model to return.
// The task lists the templates of the role's scope and tells the model not to
// read an absent row as "not exploited": a missing KEV row only means the data
// does not say. The model is told to give no SQL. The list is the scope's, not
// the whole registry (ADR-0013), so the model is not offered a template the
// scope would refuse.
func querySpec(scope querytpl.Scope) answer.Spec {
	// The example names kev-by-cve only when the scope allows it; otherwise it
	// shows an empty template, so the example never offers what the scope refuses.
	example := `{"template":"","params":{}}`
	if scope.Allows(querytpl.KEVByCVE) {
		example = `{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`
	}
	return answer.Spec{
		Task: "Pick one query template that would confirm this risk, and give the values of its parameters. " +
			"The templates are: " + scope.Describe() +
			" Do not write SQL: Proven builds the query from the template. Leave template empty when none fits. Do not treat a missing row as not exploited.",
		Shape:   "one JSON object. It has template (string) and params (an object of strings).",
		Example: example,
	}
}

// Prompt asks for the template and values for one risk. scope is the
// templates the Investigator may ask; the nil scope is every template.
func Prompt(id, title, path string, scope querytpl.Scope) string {
	// The title is model text shaped by the change, so it enters as data.
	return answer.Instruct(querySpec(scope), answer.Encode(struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Path  string `json:"path"`
	}{id, title, path}))
}

// row is the query as the model returns it. The prompt does not ask for SQL.
// It is read only so that an answer that carries SQL text can be refused.
type row struct {
	Template string            `json:"template"`
	Params   map[string]string `json:"params"`
	SQL      string            `json:"sql"`
}

// request is the template request in a model's answer. An answer that carries
// SQL text is refused (ADR-0032 point 1), whatever else it holds.
func (r row) request() (querytpl.Request, error) {
	if strings.TrimSpace(r.SQL) != "" {
		return querytpl.Request{}, &querytpl.Refusal{Reason: querytpl.ReasonModelSQL}
	}
	return querytpl.Request{Template: r.Template, Params: r.Params}, nil
}

// Execute runs a request the skeptic named. It does not ask a model. A
// request that fails its format, or that names a template outside scope, is a
// *querytpl.Refusal and no query is run. scope is the Skeptic's evidence scope
// (ADR-0013). The store builds the form it speaks (ADR-0047); this function does not.
func Execute(ctx context.Context, q query.Runner, req querytpl.Request, rel querytpl.Relations, scope querytpl.Scope) (query.Result, error) {
	// The scope is checked before the engine is looked at, so a template the
	// role may not ask is refused the same way whether or not a store is up,
	// and no store is ever asked for it.
	if err := scope.Check(querytpl.RoleSkeptic, req); err != nil {
		return query.Result{}, err
	}
	// The refusal comes first: a request that cannot be built is the same
	// answer whether or not the engine is up. The caller does not keep the
	// text. A store that is up checks the request again before it runs one.
	if q == nil {
		if err := querytpl.Validate(req); err != nil {
			return query.Result{}, err
		}
		return query.Result{}, &service.Down{Service: "query", Reason: "no query engine"}
	}
	return q.Run(ctx, req, rel)
}

// Investigate asks for a template and runs the query Proven builds from it. A
// query with no rows adds no evidence. A refused template adds none either,
// and the candidate's QueryNote says why.
func Investigate(ctx context.Context, c candidate.Candidate, snapshot string, gen generator.Generator, q query.Runner, rel querytpl.Relations, scope querytpl.Scope) (candidate.Candidate, error) {
	var got row
	if err := answer.Read(ctx, gen, "investigator", Prompt(c.ID, c.Title, c.Path, scope), querySpec(scope), &got); err != nil {
		return c, err
	}
	req, err := got.request()
	if err != nil {
		c.QueryNote = "refused: " + err.Error()
		return c, nil
	}
	// No query proposed means no evidence, not a failure.
	if req.Empty() {
		return c, nil
	}
	// ADR-0013: a template outside the Investigator's scope is refused before
	// any store is asked, and the finding goes on without query evidence. The
	// model was told only of the scope, so this is a model that named another.
	if err := scope.Check(querytpl.RoleInvestigator, req); err != nil {
		c.QueryNote = "refused: " + err.Error()
		return c, nil
	}
	// A nil runner panics on Run, and the review would lose the finding
	// instead of naming the query engine down. A request that cannot be
	// built is a note, the same answer whether or not the engine is up.
	if q == nil {
		if err := querytpl.Validate(req); err != nil {
			var refusal *querytpl.Refusal
			if errors.As(err, &refusal) {
				c.QueryNote = "refused: " + refusal.Error()
				return c, nil
			}
			return c, err
		}
		return c, &service.Down{Service: "query", Reason: "no query engine"}
	}
	// ADR-0032, ADR-0047: the seam checks the request and the store builds
	// the form it speaks. A refusal is the finding having no query evidence,
	// named in the trace; the Judge decides as it does for any finding
	// without evidence.
	res, err := q.Run(ctx, req, rel)
	if err != nil {
		var refusal *querytpl.Refusal
		if errors.As(err, &refusal) {
			c.QueryNote = "refused: " + refusal.Error()
			return c, nil
		}
		return c, err
	}
	// ADR-0002: a query with no rows proves nothing, so it adds no evidence.
	if len(res.Rows) == 0 {
		return c, nil
	}
	// The snapshot is the engine's own, when it reports one: it names the data
	// the rows came from more exactly.
	snap := snapshot
	if res.Snapshot != "" {
		snap = res.Snapshot
	}
	// ADR-0046: the adapter that owns the template decides whether these rows
	// are evidence. The built-in rules keep any row; an organisation's rule may
	// keep fewer or none.
	if owner, _, found := source.Lookup(strings.TrimSpace(req.Template)); found {
		kept, ok := owner.Rule(source.Evidence{
			SQL: res.SQL, Form: res.Form, Engine: res.Engine, Snapshot: snap,
			Columns: res.Columns, Rows: res.Rows,
		})
		if !ok {
			return c, nil
		}
		res.Columns, res.Rows = kept.Columns, kept.Rows
	}
	c.Evidence = append(c.Evidence, candidate.Evidence{
		Kind: candidate.KindQuery,
		Query: &candidate.QueryEvidence{
			SQL: res.SQL, Form: res.Form, Engine: res.Engine, Snapshot: snap,
			Columns: res.Columns, Rows: res.Rows,
			// The request is kept so the pipeline can bind the rows to the
			// change. The Investigator never sets Bound: the model chose the
			// query, so Proven decides what it is about.
			Request: req,
		},
	})
	return c, nil
}
