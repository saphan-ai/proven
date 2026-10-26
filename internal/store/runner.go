package store

import (
	"context"
	"fmt"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// runner is the evidence seam over one store: a template request goes in and
// the form that ran comes back.
type runner struct{ s Store }

// Runner returns the query seam of s, for a caller outside a review that runs
// the dependency check on a store of its own choosing. A nil store is refused
// by name and does not dial.
func Runner(s Store) query.Runner { return runner{s: s} }

// Run asks the store to run the request and keeps the form, the text, the
// engine, the snapshot and the rows it answered with.
func (r runner) Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (query.Result, error) {
	if r.s == nil {
		return query.Result{}, fmt.Errorf("store: no evidence store")
	}
	ev, err := r.s.Run(ctx, req, rel)
	if err != nil {
		return query.Result{}, err
	}
	return query.Result{
		SQL: ev.Query, Form: ev.Form, Engine: ev.Engine, Snapshot: ev.Snapshot,
		Columns: ev.Columns, Rows: ev.Rows,
	}, nil
}
