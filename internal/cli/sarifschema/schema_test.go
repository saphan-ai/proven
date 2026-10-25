package sarifschema

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/saphan-ai/proven/internal/cli"
	"github.com/saphan-ai/proven/internal/review"
)

// exitOK is the exit code of a review that ran (ADR-0020 §2).
const exitOK = 0

func emptyEnv(string) string { return "" }

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

// shapes are the edge shapes of the writer's input. internal/cli's
// TestSARIFShapes checks the same inputs for structure; the two tables are
// separate because an external test package cannot read that package's test
// variables, so a shape added there is added here too.
var shapes = []struct {
	name   string
	result review.Result
}{
	{"no evidence", review.Result{Findings: []review.Finding{
		{ID: "F1", Severity: review.SeverityError, Message: "m", Location: review.Location{Path: "a.go", StartLine: 1, EndLine: 2}},
	}}},
	{"unresolved", review.Result{Findings: []review.Finding{
		{ID: "F1", Severity: review.SeverityWarning, Unresolved: true, Message: "m", Location: review.Location{Path: "a.go", StartLine: 1}},
	}}},
	{"failed service", review.Result{
		FailedServices: []review.ServiceFailure{{Service: "nightshift", Reason: "timeout"}},
	}},
	{"empty path", review.Result{Findings: []review.Finding{
		{ID: "F1", Severity: review.SeverityInfo, Message: "m", Location: review.Location{Path: "", StartLine: 3}},
	}}},
	{"line 0", review.Result{Findings: []review.Finding{
		{ID: "F1", Severity: review.SeverityInfo, Message: "m", Location: review.Location{Path: "a.go", StartLine: 0, EndLine: 0}},
	}}},
	{"end line before start line", review.Result{Findings: []review.Finding{
		{ID: "F1", Severity: review.SeverityInfo, Message: "m", Location: review.Location{Path: "a.go", StartLine: 9, EndLine: 4}},
	}}},
	{"unknown severity", review.Result{Findings: []review.Finding{
		{ID: "F1", Severity: review.Severity("CRITICAL"), Message: "m", Location: review.Location{Path: "a.go", StartLine: 1}},
	}}},
	{"empty result", review.Result{}},
}

// TestSARIFShapesMatchTheSchema proves every edge shape validates against
// sarif-schema-2.1.0.json: no evidence, unresolved, a failed service, an empty
// path, line 0, an end line before the start line, an unknown severity and
// the empty result. A writer that emitted startLine 0 or a result without a
// message would turn its row red.
func TestSARIFShapesMatchTheSchema(t *testing.T) {
	schema := compileSARIFSchema(t)
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			body := sarifBytes(t, shape.result)
			if err := validateSARIF(schema, body); err != nil {
				t.Fatalf("schema: %v\n%s", err, body)
			}
		})
	}
}

// sarifBytes renders one result through the SARIF writer.
func sarifBytes(t *testing.T, result review.Result) []byte {
	t.Helper()
	writer, err := cli.FormatterFor("sarif")
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

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// sampleScan runs one proven command in a fresh directory.
// The environment is empty but for PROVEN_LOG_DIR, a temporary directory, so
// the host's variables stay out of the sample path and the trace is not
// written to the user's real cache directory.
func sampleScan(t *testing.T, args []string) (stdout []byte, code int, stderr string) {
	t.Helper()
	t.Chdir(t.TempDir())
	logs := t.TempDir()
	env := func(k string) string {
		if k == cli.EnvLogDir {
			return logs
		}
		return ""
	}
	var out, errBuf bytes.Buffer
	code = cli.Run(args, env, &out, &errBuf, nil)
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
	var doc struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Name string `json:"name"`
				} `json:"driver"`
			} `json:"tool"`
			Results []struct {
				Properties map[string]any `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Runs) != 1 {
		t.Fatalf("want one run, got %d\n%s", len(doc.Runs), body)
	}
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
