package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/saphan-ai/proven/internal/review"
)

// TestSARIFLevelFollowsSeverity proves the severity mapping of ADR-0020 §1:
// INFO→note, WARNING→warning, ERROR→error. It also proves one run, driver
// name proven, and version 2.1.0. An empty buffer turns this red.
func TestSARIFLevelFollowsSeverity(t *testing.T) {
	writer, err := FormatterFor("sarif")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = writer.Write(&buf, review.Result{Findings: []review.Finding{
		{ID: "i", Severity: review.SeverityInfo, Message: "low", Location: review.Location{Path: "a.go", StartLine: 1}},
		{ID: "w", Severity: review.SeverityWarning, Message: "mid", Location: review.Location{Path: "a.go", StartLine: 2}},
		{ID: "e", Severity: review.SeverityError, Message: "high", Location: review.Location{Path: "a.go", StartLine: 3}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("sarif writer wrote nothing")
	}
	body := buf.String()
	for _, want := range []string{
		`"name": "proven"`,
		`"version": "2.1.0"`,
		`"level": "note"`,
		`"level": "warning"`,
		`"level": "error"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Count(body, `"runs"`) != 1 {
		t.Fatalf("want one run\n%s", body)
	}
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

// TestSampleSARIFMatchesTheSchema proves four properties of one SARIF log:
// the discount scan --sarif exits 0, the log matches sarif-schema-2.1.0.json,
// removing the required field "version" makes that check fail, and one run's
// --text-output, --json-output and --sarif-output files carry the same findings.
func TestSampleSARIFMatchesTheSchema(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "out.txt")
	jsonPath := filepath.Join(dir, "out.json")
	sarifPath := filepath.Join(dir, "out.sarif")
	stdout, code, stderr := sampleScan(t, []string{
		"scan", "--sample", "--sarif",
		"--text-output=" + textPath,
		"--json-output=" + jsonPath,
		"--sarif-output=" + sarifPath,
		discountTarget(t),
	})
	if code != exitOK {
		t.Fatalf("exit %d\n%s\n%s", code, stderr, stdout)
	}
	if len(stdout) == 0 {
		t.Fatal("sarif writer wrote nothing")
	}
	textBody := readOut(t, textPath)
	jsonBody := readOut(t, jsonPath)
	sarifBody := readOut(t, sarifPath)
	if len(textBody) == 0 || len(jsonBody) == 0 || len(sarifBody) == 0 {
		t.Fatalf("empty file text=%d json=%d sarif=%d", len(textBody), len(jsonBody), len(sarifBody))
	}
	if !strings.Contains(string(textBody), "finding F1 ERROR") {
		t.Fatalf("text file\n%s", textBody)
	}
	jsonIDs := jsonFindingIDs(t, jsonBody)
	sarifIDs := sarifFindingIDs(t, sarifBody)
	stdoutIDs := sarifFindingIDs(t, stdout)
	assertDiscountIDs(t, jsonIDs)
	if strings.Join(jsonIDs, "\n") != strings.Join(sarifIDs, "\n") || strings.Join(jsonIDs, "\n") != strings.Join(stdoutIDs, "\n") {
		t.Fatalf("json %v sarif %v stdout %v", jsonIDs, sarifIDs, stdoutIDs)
	}

	schema := compileSARIFSchema(t)
	if err := validateSARIF(schema, sarifBody); err != nil {
		t.Fatalf("schema: %v\n%s", err, sarifBody)
	}

	// Negative control. Mutated field: version. Nothing else in the log changes.
	var doc map[string]any
	if err := json.Unmarshal(sarifBody, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["version"]; !ok {
		t.Fatal("log has no version to remove")
	}
	delete(doc, "version")
	stripped, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	versionErr := validateSARIF(schema, stripped)
	if versionErr == nil {
		t.Fatal("schema accepted a log with version removed")
	}
	if !strings.Contains(versionErr.Error(), "version") {
		t.Fatalf("rejection did not name version: %v", versionErr)
	}
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

// compileSARIFSchema loads the vendored OASIS SARIF 2.1.0 schema.
// The check does not fetch the URL named in the log, so it needs no network.
func compileSARIFSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	path := filepath.Join(moduleRoot(t), "testdata", "sarif", "sarif-schema-2.1.0.json")
	schema, err := jsonschema.Compile(path)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func validateSARIF(schema *jsonschema.Schema, body []byte) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return err
	}
	return schema.Validate(v)
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
	var doc struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Name string `json:"name"`
				} `json:"driver"`
			} `json:"tool"`
			Results []struct {
				Properties struct {
					ID string `json:"id"`
				} `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Runs) != 1 {
		t.Fatalf("runs %d", len(doc.Runs))
	}
	if doc.Runs[0].Tool.Driver.Name != "proven" {
		t.Fatalf("driver %q", doc.Runs[0].Tool.Driver.Name)
	}
	ids := make([]string, 0, len(doc.Runs[0].Results))
	for _, result := range doc.Runs[0].Results {
		ids = append(ids, result.Properties.ID)
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
