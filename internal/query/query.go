package query

import "context"

// Result is one query's rows.
type Result struct {
	SQL      string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
}

// Adapter runs one SQL statement.
type Adapter interface {
	Query(ctx context.Context, sql string) (Result, error)
}
