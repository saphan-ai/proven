package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestUnknownFormatHasNoWriter proves a format outside text, json and sarif
// still refuses. ADR-0020 §1 names those three. The mutated field is the
// format name: "xml" must fail, and the error must name it.
func TestUnknownFormatHasNoWriter(t *testing.T) {
	if _, err := FormatterFor("xml"); err == nil || !strings.Contains(err.Error(), "xml") {
		t.Fatalf("err %v", err)
	}
}

// TestJSONWriterEmitsFindings proves the JSON writer carries every finding,
// its severity, its evidence, its unresolved mark and the failed services.
// The field names are the contract in docs/OUTPUT_FORMATS.md.
// An empty buffer is a failure: a writer that writes nothing turns this red.
func TestJSONWriterEmitsFindings(t *testing.T) {
	writer, err := FormatterFor("json")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = writer.Write(&buf, review.Result{
		Findings: []review.Finding{{
			ID:         "F1",
			RuleID:     "discount-applied-twice",
			Title:      "Discount is applied twice",
			Message:    "the discount runs on the discounted total",
			Severity:   review.SeverityError,
			Unresolved: true,
			Location:   review.Location{Path: "billing/discount.go", StartLine: 5, EndLine: 5},
			Evidence: []review.Evidence{
				{
					Kind: review.EvidenceTest,
					Test: &review.TestEvidence{
						Path: "billing/discount_test.go", Name: "TestDiscounted",
						Command:       "go test ./billing -run TestDiscounted",
						FailsOnChange: true, PassesOnBase: true, Output: "FAIL",
					},
				},
				{
					Kind: review.EvidenceQuery,
					Query: &review.QueryEvidence{
						SQL: "SELECT id FROM t", Engine: "local-sample", Snapshot: "sample-2026-10-22",
						Columns: []string{"id"}, Rows: [][]string{{"row-1"}},
					},
				},
			},
		}},
		FailedServices: []review.ServiceFailure{{Service: "nightshift", Reason: "timeout"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("json writer wrote nothing")
	}
	body := buf.String()
	for _, want := range []string{
		`"id": "F1"`,
		`"severity": "ERROR"`,
		`"unresolved": true`,
		`"kind": "test"`,
		`"name": "TestDiscounted"`,
		`"fails_on_change": true`,
		`"kind": "query"`,
		`"rows"`,
		`"row-1"`,
		`"service": "nightshift"`,
		`"reason": "timeout"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}
