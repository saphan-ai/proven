package query

import "context"

// Result is one query's rows. SQL is the text that ran, Form names that
// text's language, Engine names the store, and Snapshot is when the rows
// were read. The Judge keeps all of it, so a reader can run the same query
// again (ADR-0002, ADR-0047).
// Form is "sql" for Zetaris and the local store. "mango" is reserved for
// CouchDB and is not produced yet.
// Engine is "zetaris" for a live Zetaris call, "local" for the configured
// local file, and "local-sample" when sample mode opens a fixture file.
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

// Adapter runs one SQL statement. Live and Sample implement it. The
// Investigator, the EPSS step and the OSV step take an Adapter, so the same
// code runs live and in sample mode (ADR-0001, point 2). An empty API key
// does not build a live client: Absent names Zetaris down and does not dial
// (ADR-0013 point 4a).
type Adapter interface {
	Query(ctx context.Context, sql string) (Result, error)
}
