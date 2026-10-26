package osv

import (
	"encoding/json"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
)

// TestQueryEvidencePrintsNoRowsAsAnEmptyList proves an OSV query answer
// with no rows becomes an empty list, so its JSON is [] and not null.
// Keeping nil as nil makes the nil case fail.
func TestQueryEvidencePrintsNoRowsAsAnEmptyList(t *testing.T) {
	cases := []struct {
		name string
		rows [][]string
		want string
	}{
		{name: "nil answer", rows: nil, want: "[]"},
		{name: "empty answer", rows: [][]string{}, want: "[]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := queryEvidence(query.Result{Rows: c.rows}, "snap")
			if ev.Query == nil {
				t.Fatal("query evidence missing")
			}
			if c.rows == nil && ev.Query.Rows == nil {
				t.Fatal("nil rows stayed nil")
			}
			if ev.Query.Rows == nil || len(ev.Query.Rows) != len(c.rows) {
				t.Fatalf("rows %#v", ev.Query.Rows)
			}
			body, err := json.Marshal(ev.Query.Rows)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != c.want {
				t.Fatalf("json %s want %s", body, c.want)
			}
		})
	}
}
