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

func TestRecordedDemoWritesAllFormatsWithoutAnAssumedVendor(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	dir := t.TempDir()
	t.Chdir(dir)
	var out, errout bytes.Buffer
	args := []string{"scan", "--sample", "--html-output", filepath.Join(dir, "r.html"), "--json-output", filepath.Join(dir, "r.json"), "--sarif-output", filepath.Join(dir, "r.sarif"), "--md-output", filepath.Join(dir, "r.md"), target}
	if code := Run(args, envOf(map[string]string{"PROVEN_LOG_DIR": filepath.Join(dir, "logs")}), &out, &errout, nil); code != 0 {
		t.Fatalf("%d %s", code, errout.String())
	}
	for _, format := range []string{"html", "json", "sarif", "md"} {
		body, err := os.ReadFile(filepath.Join(dir, "r."+format))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(body, []byte("NVIDIA")) || bytes.Contains(body, []byte("default reviewer")) {
			t.Fatalf("%s has a fictitious model identity", format)
		}
		if (format == "html" || format == "md") && !bytes.Contains(body, []byte("Recorded demo")) {
			t.Fatalf("%s did not identify the replay", format)
		}
	}
	var doc struct {
		AnalysisMode    string   `json:"analysis_mode"`
		DefaultReviewer string   `json:"default_reviewer"`
		ReviewerModels  []string `json:"reviewer_models"`
		Guardrail       struct {
			State string `json:"state"`
		} `json:"guardrail"`
	}
	body, _ := os.ReadFile(filepath.Join(dir, "r.json"))
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.AnalysisMode != "recorded" || doc.DefaultReviewer != "" || len(doc.ReviewerModels) != 0 || doc.Guardrail.State != "replayed" {
		t.Fatalf("%+v", doc)
	}
}
func TestDeterministicModeOverridesConfiguredReviewerIdentity(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	os.MkdirAll(filepath.Join(target, "base"), 0700)
	os.MkdirAll(filepath.Join(target, "change"), 0700)
	os.WriteFile(filepath.Join(target, "change", "README.md"), []byte("# Example\n"), 0600)
	config := filepath.Join(dir, "proven.yaml")
	os.WriteFile(config, []byte("default_reviewer: NVIDIA\nspecial_reviewer: other-model\n"), 0600)
	r, err := review.Open(review.Options{ConfigPath: config, LogDir: filepath.Join(dir, "logs"), Env: envOf(map[string]string{"HOME": dir, "PROVEN_GENERATOR_MODE": "off", "PROVEN_NETWORK": "local", "PROVEN_EVIDENCE_FILE": filepath.Join(dir, "evidence.sqlite")})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Review(t.Context(), review.Request{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if result.DefaultReviewer != "" || result.SpecialReviewer != "" || len(result.ReviewerModels) != 0 || len(result.Models) != 0 || result.AnalysisMode != "deterministic" {
		t.Fatalf("%+v", result)
	}
	for _, f := range []string{"text", "json", "sarif", "md", "html"} {
		b, err := render(f, result)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "NVIDIA") || strings.Contains(string(b), "default reviewer") {
			t.Fatalf("%s retained a fictitious vendor", f)
		}
	}
}
