// Package reviewer is a worked example of a reviewer adapter.
//
// Table returns a fixed Result from a table keyed by review target. A reader
// copies this package and replaces the table with their own model. What to
// add is in docs/ADAPTERS.md under "Adding a reviewer".
package reviewer

import (
	"context"
	"errors"

	ir "github.com/saphan-ai/proven/internal/review"
)

// The assignment fails the build when internal/review.Reviewer changes,
// so a copy of this example stays bound to the real interface the shell calls.
var _ ir.Reviewer = Table{}

// ErrUnknownTarget is the refusal: the table has no verdict for this target.
var ErrUnknownTarget = errors.New("reviewer: no recorded verdict for this target")

// Row is one target and the verdict the table returns for it.
type Row struct {
	Target string
	Result ir.Result
}

// Table returns a fixed verdict from a table keyed by review target.
type Table struct {
	results map[string]ir.Result
}

// NewTable builds the table from rows. A duplicate target is refused.
func NewTable(rows []Row) (*Table, error) {
	results := make(map[string]ir.Result, len(rows))
	for _, row := range rows {
		if _, ok := results[row.Target]; ok {
			return nil, errors.New("reviewer: two verdicts for target " + row.Target)
		}
		results[row.Target] = row.Result
	}
	return &Table{results: results}, nil
}

// Review returns the table's verdict for the target. An unknown target
// returns ErrUnknownTarget and no result. A canceled context is refused first.
func (t Table) Review(ctx context.Context, req ir.Request) (ir.Result, error) {
	if err := ctx.Err(); err != nil {
		return ir.Result{}, err
	}
	result, ok := t.results[req.Target]
	if !ok {
		return ir.Result{}, ErrUnknownTarget
	}
	return result, nil
}
