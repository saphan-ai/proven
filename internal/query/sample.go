package query

import (
	"context"
	"os"

	"github.com/saphan-ai/proven/internal/limit"
)

// Sample runs SQL on the local engine. It opens no network.
type Sample struct {
	db *DB
}

// OpenSample loads a SQL file of sample tables.
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
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	cols, rows, err := s.db.Query(sql)
	if err != nil {
		return Result{}, err
	}
	if rows == nil {
		rows = [][]string{}
	}
	return Result{SQL: sql, Engine: "local-sample", Columns: cols, Rows: rows}, nil
}
