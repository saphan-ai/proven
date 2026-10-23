package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// sarifBytes renders one result through the SARIF writer.
func sarifBytes(t *testing.T, result review.Result) []byte {
	t.Helper()
	writer, err := FormatterFor("sarif")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writer.Write(&buf, result); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("sarif writer wrote nothing")
	}
	return buf.Bytes()
}

// sarifDoc is the part of a SARIF log these tests read back.
type sarifDoc struct {
	Version string `json:"version"`
	Runs    []struct {
		Tool struct {
			Driver struct {
				Name string `json:"name"`
			} `json:"driver"`
		} `json:"tool"`
		Results []struct {
			RuleID  string `json:"ruleId"`
			Level   string `json:"level"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
					Region *struct {
						StartLine int `json:"startLine"`
						EndLine   int `json:"endLine"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
			Properties map[string]any `json:"properties"`
		} `json:"results"`
		Properties map[string]any `json:"properties"`
	} `json:"runs"`
}

func decodeSARIF(t *testing.T, body []byte) sarifDoc {
	t.Helper()
	var doc sarifDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Runs) != 1 {
		t.Fatalf("want one run, got %d\n%s", len(doc.Runs), body)
	}
	return doc
}

// TestSARIFLevelFollowsSeverity proves the severity mapping of ADR-0020 §1,
// result by result: INFO→note, WARNING→warning, ERROR→error, and a severity
// the writer does not know→note. The level is read back by each result's
// ruleId, so swapping the WARNING and ERROR arms of sarifLevel turns the
// "w" and "e" rows red; a test that only looked for each level somewhere in
// the file would stay green. It also proves one run, driver name proven, and
// version 2.1.0.
func TestSARIFLevelFollowsSeverity(t *testing.T) {
	doc := decodeSARIF(t, sarifBytes(t, review.Result{Findings: []review.Finding{
		{ID: "i", Severity: review.SeverityInfo, Message: "low", Location: review.Location{Path: "a.go", StartLine: 1}},
		{ID: "w", Severity: review.SeverityWarning, Message: "mid", Location: review.Location{Path: "a.go", StartLine: 2}},
		{ID: "e", Severity: review.SeverityError, Message: "high", Location: review.Location{Path: "a.go", StartLine: 3}},
		{ID: "u", Severity: review.Severity("CRITICAL"), Message: "odd", Location: review.Location{Path: "a.go", StartLine: 4}},
	}}))
	if doc.Version != "2.1.0" || doc.Runs[0].Tool.Driver.Name != "proven" {
		t.Fatalf("version %q driver %q", doc.Version, doc.Runs[0].Tool.Driver.Name)
	}
	want := map[string]string{"i": "note", "w": "warning", "e": "error", "u": "note"}
	got := map[string]string{}
	for _, result := range doc.Runs[0].Results {
		got[result.RuleID] = result.Level
	}
	if len(got) != len(want) {
		t.Fatalf("results %v, want %v", got, want)
	}
	for id, level := range want {
		if got[id] != level {
			t.Errorf("result %s: level %q, want %q", id, got[id], level)
		}
	}
}

// sarifShape is one edge of the writer's input. The same table drives the
// structure check here; the schema check in internal/cli/sarifschema keeps a
// copy of the inputs, so a shape added here is added there too.
type sarifShape struct {
	name   string
	result review.Result
	// check reads the decoded log and states what this shape must look like.
	check func(t *testing.T, doc sarifDoc)
}

var sarifShapes = []sarifShape{
	{
		name: "no evidence",
		result: review.Result{Findings: []review.Finding{
			{ID: "F1", Severity: review.SeverityError, Message: "m", Location: review.Location{Path: "a.go", StartLine: 1, EndLine: 2}},
		}},
		check: func(t *testing.T, doc sarifDoc) {
			ev, ok := doc.Runs[0].Results[0].Properties["evidence"].([]any)
			if !ok || len(ev) != 0 {
				t.Errorf("evidence %#v, want an empty array", doc.Runs[0].Results[0].Properties["evidence"])
			}
		},
	},
	{
		name: "unresolved",
		result: review.Result{Findings: []review.Finding{
			{ID: "F1", Severity: review.SeverityWarning, Unresolved: true, Message: "m", Location: review.Location{Path: "a.go", StartLine: 1}},
		}},
		check: func(t *testing.T, doc sarifDoc) {
			if got := doc.Runs[0].Results[0].Properties["unresolved"]; got != true {
				t.Errorf("unresolved %#v, want true", got)
			}
		},
	},
	{
		name: "failed service",
		result: review.Result{
			FailedServices: []review.ServiceFailure{{Service: "nightshift", Reason: "timeout"}},
		},
		check: func(t *testing.T, doc sarifDoc) {
			failed, ok := doc.Runs[0].Properties["failed_services"].([]any)
			if !ok || len(failed) != 1 {
				t.Fatalf("failed_services %#v", doc.Runs[0].Properties["failed_services"])
			}
			one, _ := failed[0].(map[string]any)
			if one["service"] != "nightshift" || one["reason"] != "timeout" {
				t.Errorf("failed service %v", one)
			}
		},
	},
	{
		name: "empty path",
		result: review.Result{Findings: []review.Finding{
			{ID: "F1", Severity: review.SeverityInfo, Message: "m", Location: review.Location{Path: "", StartLine: 3}},
		}},
		check: func(t *testing.T, doc sarifDoc) {
			if n := len(doc.Runs[0].Results[0].Locations); n != 0 {
				t.Errorf("%d locations for a finding with no path, want none", n)
			}
		},
	},
	{
		name: "line 0",
		result: review.Result{Findings: []review.Finding{
			{ID: "F1", Severity: review.SeverityInfo, Message: "m", Location: review.Location{Path: "a.go", StartLine: 0, EndLine: 0}},
		}},
		check: func(t *testing.T, doc sarifDoc) {
			locs := doc.Runs[0].Results[0].Locations
			if len(locs) != 1 || locs[0].PhysicalLocation.Region != nil {
				t.Errorf("locations %+v, want one without a region", locs)
			}
		},
	},
	{
		name: "end line before start line",
		result: review.Result{Findings: []review.Finding{
			{ID: "F1", Severity: review.SeverityInfo, Message: "m", Location: review.Location{Path: "a.go", StartLine: 9, EndLine: 4}},
		}},
		check: func(t *testing.T, doc sarifDoc) {
			locs := doc.Runs[0].Results[0].Locations
			if len(locs) != 1 || locs[0].PhysicalLocation.Region == nil {
				t.Fatalf("locations %+v, want one with a region", locs)
			}
			if r := locs[0].PhysicalLocation.Region; r.StartLine != 9 || r.EndLine != 0 {
				t.Errorf("region %+v, want startLine 9 and no endLine", r)
			}
		},
	},
	{
		name: "unknown severity",
		result: review.Result{Findings: []review.Finding{
			{ID: "F1", Severity: review.Severity("CRITICAL"), Message: "m", Location: review.Location{Path: "a.go", StartLine: 1}},
		}},
		check: func(t *testing.T, doc sarifDoc) {
			if got := doc.Runs[0].Results[0].Level; got != "note" {
				t.Errorf("level %q, want note", got)
			}
		},
	},
	{
		name:   "empty result",
		result: review.Result{},
		check: func(t *testing.T, doc sarifDoc) {
			if len(doc.Runs[0].Results) != 0 {
				t.Errorf("results %d, want none", len(doc.Runs[0].Results))
			}
			if failed, ok := doc.Runs[0].Properties["failed_services"].([]any); !ok || len(failed) != 0 {
				t.Errorf("failed_services %#v, want an empty array", doc.Runs[0].Properties["failed_services"])
			}
		},
	},
}

// TestSARIFShapes proves each edge shape of the input is written as the
// contract says: no evidence is an empty array, a finding without a path has
// no location, line 0 has no region, an end line before the start line is
// dropped, an unknown severity is a note, and the empty result is still a log
// with an empty results array. The schema half is TestSARIFShapesMatchTheSchema.
func TestSARIFShapes(t *testing.T) {
	for _, shape := range sarifShapes {
		t.Run(shape.name, func(t *testing.T) {
			body := sarifBytes(t, shape.result)
			// results must be [], never null: a consumer ranges it.
			if !bytes.Contains(body, []byte(`"results"`)) || bytes.Contains(body, []byte(`"results": null`)) {
				t.Fatalf("results missing or null\n%s", body)
			}
			shape.check(t, decodeSARIF(t, body))
		})
	}
}

// TestArtifactURIEscapesAColonInTheFirstSegment proves RFC 3986 §4.2: a
// relative reference whose first segment holds ":" would read as a scheme, so
// the colon is escaped there and only there.
func TestArtifactURIEscapesAColonInTheFirstSegment(t *testing.T) {
	for _, tc := range []struct{ path, want string }{
		{"a:b/c.go", "a%3Ab/c.go"},
		{"a:b", "a%3Ab"},
		{"dir/a:b.go", "dir/a:b.go"},
		{"dir/sp ace.go", "dir/sp%20ace.go"},
		{"", ""},
	} {
		if got := artifactURI(tc.path); got != tc.want {
			t.Errorf("artifactURI(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// writeTrace writes a trace holding one family line, the way the reviewer
// records it, and returns its path.
func writeTrace(t *testing.T, sentence string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	line, err := json.Marshal(map[string]string{"kind": "family", "text": sentence})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// textOnlyResult sets every line the text output prints beside the findings.
func textOnlyResult(t *testing.T) review.Result {
	return review.Result{
		SettingsStatement:    "settings changed by this change",
		AuthorModelStatement: "author model not stated",
		ReviewerStatements:   []string{"reviewer one: model-a", "reviewer two not configured"},
		TracePath:            writeTrace(t, "reviewer and author are the same family"),
	}
}

// TestSARIFCarriesEveryTextOnlyLine proves that each line the text output
// prints beside the findings is in the SARIF run properties, by name and
// value, so a SARIF-only consumer loses none of them. Each key is its own
// subtest: dropping any one of them turns exactly that subtest red.
func TestSARIFCarriesEveryTextOnlyLine(t *testing.T) {
	result := textOnlyResult(t)
	props := decodeSARIF(t, sarifBytes(t, result)).Runs[0].Properties
	for key, want := range map[string]any{
		"settings_statement":     result.SettingsStatement,
		"author_model_statement": result.AuthorModelStatement,
		"reviewer_statements":    []any{"reviewer one: model-a", "reviewer two not configured"},
		"trace_path":             result.TracePath,
		"family_sentence":        "reviewer and author are the same family",
	} {
		t.Run(key, func(t *testing.T) {
			got, ok := props[key]
			if !ok {
				t.Fatalf("run properties have no %q: %v", key, props)
			}
			if !jsonEqual(got, want) {
				t.Errorf("%s = %#v, want %#v", key, got, want)
			}
		})
	}
}

// TestJSONCarriesEveryTextOnlyLine is the same property for the JSON
// document, which already named four of the lines and now names the family
// sentence as well.
func TestJSONCarriesEveryTextOnlyLine(t *testing.T) {
	result := textOnlyResult(t)
	doc := decodeJSONMap(t, jsonBytes(t, result))
	for key, want := range map[string]any{
		"settings_statement":     result.SettingsStatement,
		"author_model_statement": result.AuthorModelStatement,
		"reviewer_statements":    []any{"reviewer one: model-a", "reviewer two not configured"},
		"trace_path":             result.TracePath,
		"family_sentence":        "reviewer and author are the same family",
	} {
		t.Run(key, func(t *testing.T) {
			got, ok := doc[key]
			if !ok {
				t.Fatalf("document has no %q", key)
			}
			if !jsonEqual(got, want) {
				t.Errorf("%s = %#v, want %#v", key, got, want)
			}
		})
	}
}

// jsonEqual compares two decoded JSON values by their encoding.
func jsonEqual(a, b any) bool {
	ab, errA := json.Marshal(a)
	bb, errB := json.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(ab, bb)
}

// TestSampleJSONExitsZero proves scan --sample --json exits 0 and names the
// discount findings the sample keeps (docs/OUTPUT_FORMATS.md).
func TestSampleJSONExitsZero(t *testing.T) {
	body, code, stderr := sampleScan(t, []string{"scan", "--sample", "--json", discountTarget(t)})
	if code != exitOK {
		t.Fatalf("exit %d\n%s\n%s", code, stderr, body)
	}
	if len(body) == 0 {
		t.Fatal("json writer wrote nothing")
	}
	ids := jsonFindingIDs(t, body)
	assertDiscountIDs(t, ids)
}

// sampleScan runs one proven command in a fresh directory.
// The empty environment keeps the host's variables out of the sample path.
func sampleScan(t *testing.T, args []string) (stdout []byte, code int, stderr string) {
	t.Helper()
	t.Chdir(t.TempDir())
	var out, errBuf bytes.Buffer
	code = Run(args, emptyEnv, &out, &errBuf, nil)
	return out.Bytes(), code, errBuf.String()
}

func discountTarget(t *testing.T) string {
	t.Helper()
	return filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
}

func readOut(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func jsonFindingIDs(t *testing.T, body []byte) []string {
	t.Helper()
	var doc struct {
		Findings []struct {
			ID string `json:"id"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(doc.Findings))
	for _, finding := range doc.Findings {
		ids = append(ids, finding.ID)
	}
	return ids
}

// sarifFindingIDs reads each result's properties.id.
// That is where the writer puts the finding id (docs/OUTPUT_FORMATS.md).
func sarifFindingIDs(t *testing.T, body []byte) []string {
	t.Helper()
	doc := decodeSARIF(t, body)
	if doc.Runs[0].Tool.Driver.Name != "proven" {
		t.Fatalf("driver %q", doc.Runs[0].Tool.Driver.Name)
	}
	ids := make([]string, 0, len(doc.Runs[0].Results))
	for _, result := range doc.Runs[0].Results {
		id, _ := result.Properties["id"].(string)
		ids = append(ids, id)
	}
	return ids
}

// assertDiscountIDs proves the sample's reported findings.
// F3 is dropped for lack of evidence, so a writer that emits it fails here.
func assertDiscountIDs(t *testing.T, ids []string) {
	t.Helper()
	joined := strings.Join(ids, "\n")
	for _, want := range []string{"F1", "F2", "F4", "osv:evil.example/leftpad"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %s in %v", want, ids)
		}
	}
	if strings.Contains(joined, "F3") {
		t.Fatalf("F3 was reported: %v", ids)
	}
}
