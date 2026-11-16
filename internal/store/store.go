// Package store provides evidence adapters. SQLite is the standalone default;
// PostgreSQL, CouchDB, local fixture files and Zetaris use the same template seam.
// A store registers source tables and returns rows with their source identity.
package store

import (
	"context"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// Store registers a source table and runs a template query .
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
// Engine is zetaris, local or couchdb. Snapshot is the UTC time of the read.
type Evidence struct {
	Form     string
	Query    string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
}
