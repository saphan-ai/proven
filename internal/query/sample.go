package query

import (
	"context"
	"os"

	"github.com/saphan-ai/proven/internal/limit"
)

// Sample runs SQL on the local engine. It opens no network.
// It is the query adapter of sample mode (ADR-0002, point 4).
type Sample struct {
	db *DB
}

// OpenSample loads a SQL file of sample tables.
// The file may hold only CREATE TABLE and INSERT statements.
func OpenSample(path string) (*Sample, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	db, err := LoadSQL(string(b))
	if err != nil {
		return nil, err
	}
	return &Sample{db: db}, nil
}

// Query runs the statement on the loaded tables.
func (s *Sample) Query(ctx context.Context, sql string) (Result, error) {
	// ADR-0001, point 3: every adapter call is bounded, the local one too.
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	cols, rows, err := s.db.Query(sql)
	if err != nil {
		return Result{}, err
	}
	// No match is an empty list, not nil, so the evidence prints [] and not null.
	if rows == nil {
		rows = [][]string{}
	}
	return Result{SQL: sql, Engine: "local-sample", Columns: cols, Rows: rows}, nil
}
