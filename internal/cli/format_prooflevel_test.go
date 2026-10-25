package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSampleJSONAndSARIFCarryProofLevel proves the sample review's JSON and
// SARIF carry each finding's level of proof and the same totals (ADR-0044).
// F1 is proved-by-test. The dependency finding, the one whose id starts with
// osv:, is proved-by-data. Omitting proof_level or proof_counts from either
// writer turns this red. The mutated field of a writer that drops the level
// is findings[].proof_level (JSON) and properties.proof_level (SARIF).
func TestSampleJSONAndSARIFCarryProofLevel(t *testing.T) {
	jsonBody, code, stderr := sampleScan(t, []string{"scan", "--sample", "--json", discountTarget(t)})
	if code != exitOK {
		t.Fatalf("json exit %d\n%s\n%s", code, stderr, jsonBody)
	}
	sarifBody, code, stderr := sampleScan(t, []string{"scan", "--sample", "--sarif", discountTarget(t)})
	if code != exitOK {
		t.Fatalf("sarif exit %d\n%s\n%s", code, stderr, sarifBody)
	}

	jsonDoc := decodeJSONMap(t, jsonBody)
	jsonLevels := findingProofLevels(t, "JSON", jsonFindings(t, jsonDoc))
	depID := dependencyFindingID(t, jsonFindings(t, jsonDoc))
	assertProofLevel(t, "JSON", jsonLevels, "F1", "proved-by-test")
	assertProofLevel(t, "JSON", jsonLevels, depID, "proved-by-data")

	sarifDoc := decodeSARIF(t, sarifBody)
	sarifLevels := sarifProofLevels(t, sarifDoc)
	assertProofLevel(t, "SARIF", sarifLevels, "F1", "proved-by-test")
	assertProofLevel(t, "SARIF", sarifLevels, depID, "proved-by-data")
	if !jsonEqual(jsonLevels, sarifLevels) {
		t.Errorf("proof levels differ\nJSON  %v\nSARIF %v", jsonLevels, sarifLevels)
	}

	// The sample review's totals, the same ones the core records (ADR-0044):
	// one finding proved by a test, three proved by data, none by a tool,
	// and the agreed level kept at 0 because opinions are not findings.
	wantCounts := map[string]any{
		"proved-by-test":                  float64(1),
		"proved-by-tool":                  float64(0),
		"proved-by-data":                  float64(3),
		"agreed-by-two-models-not-proven": float64(0),
	}
	jsonCounts, jsonOK := proofCountsObject(t, "JSON", jsonDoc["proof_counts"])
	sarifCounts, sarifOK := proofCountsObject(t, "SARIF", sarifDoc.Runs[0].Properties["proof_counts"])
	if jsonOK && !jsonEqual(jsonCounts, wantCounts) {
		t.Errorf("JSON proof_counts %#v; want %#v", jsonCounts, wantCounts)
	}
	if sarifOK && !jsonEqual(sarifCounts, wantCounts) {
		t.Errorf("SARIF proof_counts %#v; want %#v", sarifCounts, wantCounts)
	}
	if jsonOK && sarifOK && !jsonEqual(jsonCounts, sarifCounts) {
		t.Errorf("proof_counts differ\nJSON  %#v\nSARIF %#v", jsonCounts, sarifCounts)
	}
}

// findingProofLevels reads proof_level by finding id. A missing key is
// recorded as absent, so the assertion can name the field rather than an
// empty string a writer chose to send.
func findingProofLevels(t *testing.T, format string, findings []any) map[string]any {
	t.Helper()
	out := map[string]any{}
	for _, item := range findings {
		finding, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("%s finding is %T", format, item)
		}
		id, _ := finding["id"].(string)
		level, present := finding["proof_level"]
		if !present {
			t.Errorf("%s finding %s has no proof_level", format, id)
			continue
		}
		out[id] = level
	}
	return out
}

func jsonFindings(t *testing.T, doc map[string]any) []any {
	t.Helper()
	findings, ok := doc["findings"].([]any)
	if !ok {
		t.Fatalf("JSON findings is %T", doc["findings"])
	}
	return findings
}

func sarifProofLevels(t *testing.T, doc sarifDoc) map[string]any {
	t.Helper()
	findings := make([]any, 0, len(doc.Runs[0].Results))
	for _, result := range doc.Runs[0].Results {
		findings = append(findings, result.Properties)
	}
	return findingProofLevels(t, "SARIF", findings)
}

// dependencyFindingID is the sample's one dependency finding.
// The OSV check names that finding with an id prefix of osv: (internal/osv).
// The discount fixture's F2 also has "Dependency" in its title; that is a
// planner finding, not the dependency finding this test guards.
func dependencyFindingID(t *testing.T, findings []any) string {
	t.Helper()
	var id string
	for _, item := range findings {
		finding := item.(map[string]any)
		one, _ := finding["id"].(string)
		if !strings.HasPrefix(one, "osv:") {
			continue
		}
		if id != "" {
			t.Fatalf("two dependency findings: %s and %s", id, one)
		}
		id = one
	}
	if id == "" {
		raw, _ := json.Marshal(findings)
		t.Fatalf("no dependency finding\n%s", raw)
	}
	return id
}

func assertProofLevel(t *testing.T, format string, levels map[string]any, id, want string) {
	t.Helper()
	got, ok := levels[id]
	if !ok {
		t.Errorf("%s %s: no proof_level; want %s", format, id, want)
		return
	}
	if got != want {
		t.Errorf("%s %s: proof_level %#v; want %s", format, id, got, want)
	}
}

// proofCountsObject requires every level the core keeps, including a zero.
// A missing key is the failure: a writer that drops the zero level stays red.
func proofCountsObject(t *testing.T, format string, v any) (map[string]any, bool) {
	t.Helper()
	counts, ok := v.(map[string]any)
	if !ok {
		t.Errorf("%s proof_counts is %T (%#v), want an object", format, v, v)
		return nil, false
	}
	for _, level := range []string{
		"proved-by-test",
		"proved-by-tool",
		"proved-by-data",
		"agreed-by-two-models-not-proven",
	} {
		if _, present := counts[level]; !present {
			t.Errorf("%s proof_counts missing %s", format, level)
		}
	}
	return counts, true
}
