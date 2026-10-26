package review

import (
	"encoding/json"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// TestSeamEvidenceKeepsNilRowsNil proves the review copies keep a nil row
// set nil, so "no rows" stays distinct from an empty answer and JSON is null.
// Turning nil into an empty list makes these cases fail.
func TestSeamEvidenceKeepsNilRowsNil(t *testing.T) {
	t.Run("evidence", func(t *testing.T) {
		ev := toEvidence(candidate.Evidence{
			Kind:  candidate.KindQuery,
			Query: &candidate.QueryEvidence{},
		})
		if ev.Query == nil || ev.Query.Rows != nil {
			t.Fatalf("rows %#v", ev.Query)
		}
		body, err := json.Marshal(ev.Query.Rows)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "null" {
			t.Fatalf("json %s want null", body)
		}
	})
	t.Run("refutation", func(t *testing.T) {
		got := toRefutation(candidate.Refutation{})
		if got == nil || got.QueryRows != nil {
			t.Fatalf("rows %#v", got)
		}
		body, err := json.Marshal(got.QueryRows)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "null" {
			t.Fatalf("json %s want null", body)
		}
	})
	t.Run("empty answer", func(t *testing.T) {
		ev := toEvidence(candidate.Evidence{
			Kind:  candidate.KindQuery,
			Query: &candidate.QueryEvidence{Rows: [][]string{}},
		})
		if ev.Query == nil || ev.Query.Rows == nil || len(ev.Query.Rows) != 0 {
			t.Fatalf("rows %#v", ev.Query)
		}
		body, err := json.Marshal(ev.Query.Rows)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "[]" {
			t.Fatalf("json %s want []", body)
		}
	})
}
