package pipeline

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// TestTraceEvidenceKeepsNilRowsNil proves the pipeline trace keeps a nil row
// set nil, so the trace JSON says null and not []. Turning nil into an empty
// list makes the nil case fail.
func TestTraceEvidenceKeepsNilRowsNil(t *testing.T) {
	t.Run("nil answer", func(t *testing.T) {
		out := traceEvidence([]candidate.Evidence{{
			Kind:  candidate.KindQuery,
			Query: &candidate.QueryEvidence{},
		}})
		if len(out) != 1 || out[0].Query == nil || out[0].Query.Rows != nil {
			t.Fatalf("rows %#v", out)
		}
		body, err := json.Marshal(out[0].Query)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"rows":null`) {
			t.Fatalf("json %s", body)
		}
	})
	t.Run("empty answer", func(t *testing.T) {
		out := traceEvidence([]candidate.Evidence{{
			Kind:  candidate.KindQuery,
			Query: &candidate.QueryEvidence{Rows: [][]string{}},
		}})
		if len(out) != 1 || out[0].Query == nil || out[0].Query.Rows == nil {
			t.Fatalf("rows %#v", out)
		}
		body, err := json.Marshal(out[0].Query)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"rows":[]`) {
			t.Fatalf("json %s", body)
		}
	})
}
