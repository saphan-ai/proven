package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// This file holds every test that needs the vendored SARIF schema and the
// jsonschema module, so the rest of the package tests do not depend on it.

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

// TestSARIFShapesMatchTheSchema proves every edge shape of sarifShapes
// validates against sarif-schema-2.1.0.json: no evidence, unresolved, a failed
// service, an empty path, line 0, an end line before the start line, an
// unknown severity and the empty result. A writer that emitted startLine 0 or
// a result without a message would turn its row red.
func TestSARIFShapesMatchTheSchema(t *testing.T) {
	schema := compileSARIFSchema(t)
	for _, shape := range sarifShapes {
		t.Run(shape.name, func(t *testing.T) {
			body := sarifBytes(t, shape.result)
			if err := validateSARIF(schema, body); err != nil {
				t.Fatalf("schema: %v\n%s", err, body)
			}
		})
	}
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
