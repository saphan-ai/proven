package query

import "context"

// Result is one query's rows, with the SQL that produced them, the engine
// that ran it, and when they were read. The Judge keeps all of it as
// evidence, so a reader can run the same SQL again (ADR-0002).
// Engine is "zetaris" for a live call and "local-sample" in sample mode.
// Snapshot is when the rows were read; the local engine leaves it empty, and
// the caller fills in the fixture's snapshot. Every cell is text.
type Result struct {
	SQL      string
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
