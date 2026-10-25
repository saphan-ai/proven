package query

import (
	"context"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// Result is one query's rows. SQL is the text that ran, Form names that
// text's language, Engine names the store, and Snapshot is when the rows
// were read. The Judge keeps all of it, so a reader can run the same query
// again (ADR-0002, ADR-0047).
// Form is "sql" for Zetaris and the local store, and "mango" for CouchDB.
// Engine is "zetaris" for a live Zetaris call, "local" for the configured
// local file, "couchdb" for CouchDB, and "local-sample" when sample mode opens a fixture file.
// A fixture leaves Snapshot empty and the caller fills the fixture's date.
// Zetaris and a configured local store set Snapshot to the UTC time of the
// call. Every cell is text.
type Result struct {
	SQL      string
	Form     string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
}

// Adapter runs one SQL statement. Live, the local file and Sample implement
// it. A store that speaks SQL calls Query with the sql form it built. An
// empty API key does not build a live client: Absent names Zetaris down and
// does not dial (ADR-0013 point 4a).
type Adapter interface {
	Query(ctx context.Context, sql string) (Result, error)
}

// Runner is the review's evidence seam (ADR-0047, ADR-0032).
// It takes a template request and the review's relation names, and returns
// the evidence of the form that ran. SQL is the text of that form. The
// Investigator, the dependency check and the EPSS step take a Runner, so a
// live review and a sample review call the same seam and neither builds SQL.
type Runner interface {
	Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (Result, error)
}
