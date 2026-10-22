// Package investigator turns a risk into a query and runs it.
package investigator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/query"
)

// Prompt asks for the SQL for one risk.
func Prompt(id, title, path string) string {
	return "query " + id + "\n" + title + "\n" + path
}

type row struct {
	SQL string `json:"sql"`
}

// Investigate asks for SQL and runs it. A query with no rows adds no evidence.
func Investigate(ctx context.Context, c candidate.Candidate, snapshot string, gen generator.Generator, q query.Adapter) (candidate.Candidate, error) {
	raw, err := gen.Complete(ctx, "investigator", Prompt(c.ID, c.Title, c.Path))
	if err != nil {
		return c, err
	}
	var got row
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		return c, fmt.Errorf("investigator: response is not a query: %w", err)
	}
	if got.SQL == "" {
		return c, nil
	}
	res, err := q.Query(ctx, got.SQL)
	if err != nil {
		return c, err
	}
	if len(res.Rows) == 0 {
		return c, nil
	}
	c.Evidence = append(c.Evidence, candidate.Evidence{
		Kind: candidate.KindQuery,
		Query: &candidate.QueryEvidence{
			SQL: res.SQL, Engine: res.Engine, Snapshot: snapshot,
			Columns: res.Columns, Rows: res.Rows,
		},
	})
	return c, nil
}
