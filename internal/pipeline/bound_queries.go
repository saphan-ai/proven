package pipeline

import (
	"context"

	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// A model's built-in advisory query must describe the change before any
// database is contacted. The deterministic dependency scan runs separately.
type boundQueries struct {
	query.Runner
	facts osv.Facts
}

func (b boundQueries) Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (query.Result, error) {
	switch req.Template {
	case querytpl.OSVByPackage, querytpl.OSVByAdvisory, querytpl.KEVByCVE, querytpl.EPSSByCVE,
		querytpl.MalByPackage, querytpl.ScorecardByPackage, querytpl.CVSSByCVE, querytpl.SPDXByPackage:
		if !b.facts.Binds(req) {
			return query.Result{}, &querytpl.Refusal{Reason: "query is not about a dependency or advisory affected by the change"}
		}
	}
	if b.Runner == nil {
		return query.Result{}, &service.Down{Service: "query", Reason: "no query engine"}
	}
	return b.Runner.Run(ctx, req, rel)
}
