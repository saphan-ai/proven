package fixture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/investigator"
	"github.com/saphan-ai/proven/internal/agent/prover"
	"github.com/saphan-ai/proven/internal/agent/skeptic"
)

func TestRecordedPromptsFollowTheAgents(t *testing.T) {
	f, err := Load(discount(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"investigator\x00" + investigator.Prompt("F1", "Discount is applied twice", "billing/discount.go"):                   "",
		"investigator\x00" + investigator.Prompt("F2", "Dependency leftpad v1.2.3 is on the known-exploited list", "go.mod"): "",
		"investigator\x00" + investigator.Prompt("F3", "note may include a payment token", "billing/discount.go"):            "",
		"investigator\x00" + investigator.Prompt("F4", "CVE-2026-0001 for leftpad v1.2.3 is on the KEV list", "go.mod"):      "",
		"prover\x00" + prover.Prompt("F1", "Discount is applied twice"):                                                      "",
		"prover\x00" + prover.Prompt("F2", "Dependency leftpad v1.2.3 is on the known-exploited list"):                       "",
		"prover\x00" + prover.Prompt("F3", "note may include a payment token"):                                               "",
		"prover\x00" + prover.Prompt("F4", "CVE-2026-0001 for leftpad v1.2.3 is on the KEV list"):                            "",
		"skeptic\x00" + skeptic.Prompt("F1", "Discount is applied twice"):                                                    "",
		"skeptic\x00" + skeptic.Prompt("F2", "Dependency leftpad v1.2.3 is on the known-exploited list"):                     "",
		"skeptic\x00" + skeptic.Prompt("F3", "note may include a payment token"):                                             "",
		"skeptic\x00" + skeptic.Prompt("F4", "CVE-2026-0001 for leftpad v1.2.3 is on the KEV list"):                          "",
	}
	got := map[string]bool{}
	for _, a := range f.Answers {
		if a.Agent == "planner" {
			continue
		}
		got[a.Agent+"\x00"+a.Prompt] = true
	}
	if len(got) != len(want) {
		t.Fatalf("recorded prompts %d, want %d", len(got), len(want))
	}
	for k := range want {
		if !got[k] {
			t.Fatalf("missing recorded prompt %q", k)
		}
	}
}

func discount(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return filepath.Join(dir, "testdata", "fixture", "discount")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
