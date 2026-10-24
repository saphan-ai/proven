// Package store is the evidence-store adapter (ADR-0047).
//
// One interface, two stores this wagon. Zetaris sends a SELECT over its REST
// API, as it does today. The local store is one file and no server, and it
// runs the same SQL on the engine sample mode already uses, so it works with
// no network (ADR-0034). CouchDB is named and refused: it is not built yet.
// PROVEN_EVIDENCE_STORE chooses the store. The default is zetaris. Sample
// mode does not read the setting; it opens the fixture file as the local store.
//
// A source registers a table, then runs a query Proven built from a template
// (ADR-0032, ADR-0046). The result is the columns, the rows, the snapshot
// date, and the form that ran. The sql form is filled. The mango form is
// empty until CouchDB.
package store

import (
	"context"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// Store registers a source table and runs a template query (ADR-0047).
// Register loads the table into whichever store is configured. Run builds
// the form that store speaks and returns the columns, the rows and the
// snapshot date. The evidence names the form that ran.
type Store interface {
	Register(ctx context.Context, table Table) error
	Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (Evidence, error)
}

// Table is one source table and the snapshot date of this load.
// Rows are text, one cell per column, in column order.
type Table struct {
	Name     string
	Columns  []string
	Rows     [][]string
	Snapshot string
}

// Evidence is what a review keeps from one query.
// Form is the query language that ran. Query is the text of that form.
// Engine is zetaris or local. Snapshot is the UTC time of the read.
type Evidence struct {
	Form     string
	Query    string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
}
