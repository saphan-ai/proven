package query

import (
	"context"
	"fmt"
)

// Sample runs SQL on the local one-file store, opened on a fixture's
// tables file (ADR-0047). It opens no network. It is the query adapter of
// sample mode (ADR-0002, point 4). The engine name stays local-sample so a
// sample review reads as it did before the store had a file.
type Sample struct {
	file *File
}

// OpenSample loads a SQL file of sample tables as the local store.
// The file may hold only CREATE TABLE and INSERT statements. The file is
// not rewritten: sample mode only queries it.
func OpenSample(path string) (*Sample, error) {
	// A fixture has its own snapshot date. The store leaves Snapshot empty
	// so the caller can fill that date, as it always has.
	f, err := openFile(path, "local-sample", nil)
	if err != nil {
		return nil, err
	}
	return &Sample{file: f}, nil
}

// Query runs the statement on the loaded tables.
func (s *Sample) Query(ctx context.Context, sql string) (Result, error) {
	if s == nil || s.file == nil {
		return Result{}, fmt.Errorf("query: no local store")
	}
	return s.file.Query(ctx, sql)
}
