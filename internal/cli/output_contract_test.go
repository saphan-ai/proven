package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// jsonBytes renders one result through the JSON writer.
func jsonBytes(t *testing.T, result review.Result) []byte {
	t.Helper()
	writer, err := FormatterFor("json")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writer.Write(&buf, result); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func decodeJSONMap(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("not one JSON object: %v\n%s", err, body)
	}
	return doc
}

// contractResult sets every field of the result the writers read, with a
// value no other field shares, so a field written under the wrong name or
// taken from the wrong source cannot pass by coincidence.
func contractResult(t *testing.T) review.Result {
	return review.Result{
		Class:                "doubled-security",
		ClassRule:            "auth/**",
		Doubled:              true,
		Partial:              true,
		SettingsStatement:    "settings statement text",
		AuthorModelStatement: "author model statement text",
		ReviewerStatements:   []string{"reviewer statement one", "reviewer statement two"},
		AuthorModels:         []review.AuthorModel{{Agent: "author-agent", Model: "author-model", Family: "author-family", Source: "commit"}},
		DefaultReviewer:      "default-reviewer-model",
		SpecialReviewer:      "special-reviewer-model",
		ReviewerModels:       []string{"default-reviewer-model", "special-reviewer-model"},
		TracePath:            writeTrace(t, "family sentence text"),
		Models:               []review.ModelUse{{Agent: "judge", Route: "route-x", Model: "model-y", ReceiptID: "receipt-z", FromReceipt: true}},
		FailedServices:       []review.ServiceFailure{{Service: "nightshift", Reason: "timeout"}},
		// The counts are not a recount of the findings below. A writer that
		// counted the findings itself, instead of copying ProofCounts, would
		// write 0, 1, 0, 1 and turn the contract red.
		ProofCounts: map[review.ProofLevel]int{
			review.ProvedByTest: 4, review.ProvedByTool: 1, review.ProvedByData: 0, review.AgreedNotProven: 9,
		},
		Findings: []review.Finding{
			{
				ID: "F1", RuleID: "rule-one", Title: "title one", Message: "message one",
				Severity: review.SeverityError, Unresolved: true, Models: []string{"model-a", "model-b"},
				// Evidence is a test, whose own level would be proved-by-test.
				// The field says proved-by-tool, so a writer that derives the
				// level from the evidence instead of copying ProofLevel fails.
				ProofLevel: review.ProvedByTool,
				Location:   review.Location{Path: "dir/a.go", StartLine: 5, EndLine: 7},
				Evidence: []review.Evidence{
					{Kind: review.EvidenceTest, Test: &review.TestEvidence{
						Path: "dir/a_test.go", Name: "TestA", Command: "go test ./dir", FailsOnChange: true, PassesOnBase: true, Output: "FAIL",
					}},
					{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
						SQL: "SELECT 1", Engine: "engine-q", Snapshot: "snap-q", Columns: []string{"c1", "c2"}, Rows: [][]string{{"r1", "r2"}},
					}},
				},
			},
			// Every list of this finding is empty. Its level is not the one a query
			// would derive, so the writer has to copy ProofLevel.
			{
				ID: "F2", Severity: review.SeverityInfo, Message: "message two",
				ProofLevel: review.AgreedNotProven,
				Location:   review.Location{Path: "b.go", StartLine: 1, EndLine: 1},
				Evidence:   []review.Evidence{{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{SQL: "SELECT 2", Engine: "e", Snapshot: "s"}}},
			},
		},
	}
}

// wantEvidence is the evidence array of contractResult's first finding as it
// must read back, in the JSON document and under SARIF properties.evidence.
func wantEvidence() []any {
	return []any{
		map[string]any{"kind": "test", "test": map[string]any{
			"path": "dir/a_test.go", "name": "TestA", "command": "go test ./dir",
			"fails_on_change": true, "passes_on_base": true, "output": "FAIL",
		}},
		map[string]any{"kind": "query", "query": map[string]any{
			"sql": "SELECT 1", "engine": "engine-q", "snapshot": "snap-q",
			"columns": []any{"c1", "c2"}, "rows": []any{[]any{"r1", "r2"}},
		}},
	}
}

// wantFailedServices is the failed_services array in both formats.
func wantFailedServices() []any {
	return []any{map[string]any{"service": "nightshift", "reason": "timeout"}}
}

// TestJSONDocumentIsTheContract proves the whole JSON document, decoded and
// compared with a literal: every key of docs/OUTPUT_FORMATS.md, at every
// level, with its value. The comparison is exact, so a renamed, dropped or
// extra field turns it red (unplug JM1: rename one json tag).
func TestJSONDocumentIsTheContract(t *testing.T) {
	result := contractResult(t)
	got := decodeJSONMap(t, jsonBytes(t, result))
	want := map[string]any{
		"class": "doubled-security", "class_rule": "auth/**", "doubled": true, "partial": true,
		"settings_statement":     "settings statement text",
		"author_model_statement": "author model statement text",
		"reviewer_statements":    []any{"reviewer statement one", "reviewer statement two"},
		"author_models": []any{map[string]any{
			"agent": "author-agent", "model": "author-model", "family": "author-family", "source": "commit",
		}},
		"default_reviewer": "default-reviewer-model",
		"special_reviewer": "special-reviewer-model",
		"reviewer_models":  []any{"default-reviewer-model", "special-reviewer-model"},
		"trace_path":       result.TracePath,
		"family_sentence":  "family sentence text",
		"models": []any{map[string]any{
			"agent": "judge", "route": "route-x", "model": "model-y", "receipt_id": "receipt-z", "from_receipt": true,
		}},
		"failed_services": wantFailedServices(),
		"proof_counts": map[string]any{
			"proved-by-test": float64(4), "proved-by-tool": float64(1),
			"proved-by-data": float64(0), "agreed-by-two-models-not-proven": float64(9),
		},
		"findings": []any{
			map[string]any{
				"id": "F1", "proof_level": "proved-by-tool", "rule_id": "rule-one", "title": "title one", "message": "message one",
				"severity": "ERROR", "unresolved": true, "models": []any{"model-a", "model-b"},
				"location": map[string]any{"path": "dir/a.go", "start_line": float64(5), "end_line": float64(7)},
				"evidence": wantEvidence(),
			},
			map[string]any{
				"id": "F2", "proof_level": "agreed-by-two-models-not-proven", "rule_id": "", "title": "", "message": "message two",
				"severity": "INFO", "unresolved": false, "models": []any{},
				"location": map[string]any{"path": "b.go", "start_line": float64(1), "end_line": float64(1)},
				"evidence": []any{map[string]any{"kind": "query", "query": map[string]any{
					"sql": "SELECT 2", "engine": "e", "snapshot": "s", "columns": []any{}, "rows": []any{},
				}}},
			},
		},
	}
	if !jsonEqual(got, want) {
		gotB, _ := json.MarshalIndent(got, "", "  ")
		wantB, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("document differs from the contract\n--- got\n%s\n--- want\n%s", gotB, wantB)
	}
	// Every name in the document is also in the docs, so the table cannot
	// fall behind the writer. A name counts as documented when the docs hold
	// it as `name` or as the end of a dotted path `a.b.name`; the check does
	// not tell which object a leaf such as "path" belongs to.
	assertNamesDocumented(t, want)
}

// TestJSONEmptyListsAreArraysNotNull proves the "[], not null" line of
// docs/OUTPUT_FORMATS.md for every list the document has, from a result that
// leaves each of them nil. A writer that forgot stringSlice on one list would
// write null and turn that key red.
func TestJSONEmptyListsAreArraysNotNull(t *testing.T) {
	doc := decodeJSONMap(t, jsonBytes(t, review.Result{
		Findings: []review.Finding{{ID: "F", Evidence: []review.Evidence{{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{}}}}},
	}))
	for _, key := range []string{"reviewer_statements", "author_models", "reviewer_models", "models", "failed_services"} {
		assertEmptyArray(t, key, doc[key])
	}
	finding := doc["findings"].([]any)[0].(map[string]any)
	assertEmptyArray(t, "findings[].models", finding["models"])
	query := finding["evidence"].([]any)[0].(map[string]any)["query"].(map[string]any)
	assertEmptyArray(t, "findings[].evidence[].query.columns", query["columns"])
	assertEmptyArray(t, "findings[].evidence[].query.rows", query["rows"])

	empty := decodeJSONMap(t, jsonBytes(t, review.Result{}))
	assertEmptyArray(t, "findings", empty["findings"])
}

func assertEmptyArray(t *testing.T, name string, v any) {
	t.Helper()
	list, ok := v.([]any)
	if !ok || len(list) != 0 {
		t.Errorf("%s = %#v, want an empty array", name, v)
	}
}

// TestSARIFPropertiesAreTheContract proves where SARIF puts what the JSON
// document has: the run's properties and each result's properties, decoded and
// compared with literals, and the evidence array equal to the JSON's own. The
// comparison is exact, so dropping properties.evidence, unresolved or
// failed_services turns it red (unplug SM1).
func TestSARIFPropertiesAreTheContract(t *testing.T) {
	result := contractResult(t)
	doc := decodeSARIF(t, sarifBytes(t, result))
	run := doc.Runs[0]

	wantRun := map[string]any{
		"class": "doubled-security", "class_rule": "auth/**", "partial": true,
		"failed_services":        wantFailedServices(),
		"settings_statement":     "settings statement text",
		"author_model_statement": "author model statement text",
		"reviewer_statements":    []any{"reviewer statement one", "reviewer statement two"},
		"trace_path":             result.TracePath,
		"family_sentence":        "family sentence text",
		"proof_counts": map[string]any{
			"proved-by-test": float64(4), "proved-by-tool": float64(1),
			"proved-by-data": float64(0), "agreed-by-two-models-not-proven": float64(9),
		},
	}
	if !jsonEqual(run.Properties, wantRun) {
		gotB, _ := json.MarshalIndent(run.Properties, "", "  ")
		wantB, _ := json.MarshalIndent(wantRun, "", "  ")
		t.Fatalf("run properties differ\n--- got\n%s\n--- want\n%s", gotB, wantB)
	}

	if len(run.Results) != 2 {
		t.Fatalf("results %d, want 2", len(run.Results))
	}
	wantResults := []map[string]any{
		{"id": "F1", "proof_level": "proved-by-tool", "rule_id": "rule-one", "title": "title one", "severity": "ERROR", "unresolved": true, "evidence": wantEvidence()},
		{"id": "F2", "proof_level": "agreed-by-two-models-not-proven", "rule_id": "", "title": "", "severity": "INFO", "unresolved": false, "evidence": []any{
			map[string]any{"kind": "query", "query": map[string]any{
				"sql": "SELECT 2", "engine": "e", "snapshot": "s", "columns": []any{}, "rows": []any{},
			}},
		}},
	}
	for i, want := range wantResults {
		if !jsonEqual(run.Results[i].Properties, want) {
			gotB, _ := json.MarshalIndent(run.Results[i].Properties, "", "  ")
			wantB, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("result %d properties differ\n--- got\n%s\n--- want\n%s", i, gotB, wantB)
		}
	}

	// The SARIF evidence is the JSON evidence, not a second rendering of it.
	jsonDoc := decodeJSONMap(t, jsonBytes(t, result))
	jsonEvidence := jsonDoc["findings"].([]any)[0].(map[string]any)["evidence"]
	if !jsonEqual(run.Results[0].Properties["evidence"], jsonEvidence) {
		t.Errorf("SARIF evidence differs from the JSON evidence")
	}

	assertNamesDocumented(t, map[string]any{"run": wantRun, "result": wantResults[0]})
}

// assertNamesDocumented fails for every key under v that docs/OUTPUT_FORMATS.md
// does not name in backticks. The first level of v is skipped when its key is
// a grouping label of the test and not a name of the format.
func assertNamesDocumented(t *testing.T, v any) {
	t.Helper()
	docs, err := os.ReadFile(filepath.Join(moduleRoot(t), "docs", "OUTPUT_FORMATS.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(docs)
	names := map[string]bool{}
	collectNames(v, names)
	delete(names, "run")
	delete(names, "result")
	var missing []string
	for name := range names {
		if !strings.Contains(text, "`"+name+"`") && !strings.Contains(text, "."+name+"`") {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("docs/OUTPUT_FORMATS.md does not name: %s", strings.Join(missing, ", "))
	}
}

func collectNames(v any, into map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			into[k] = true
			collectNames(child, into)
		}
	case []any:
		for _, child := range x {
			collectNames(child, into)
		}
	}
}
