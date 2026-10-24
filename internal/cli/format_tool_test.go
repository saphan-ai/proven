package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestWriteEvidencePrintsAToolRun proves a tool's evidence is printed with the
// tool, the command and both sides, so a result is not dropped on the way to
// the reader (ADR-0040).
func TestWriteEvidencePrintsAToolRun(t *testing.T) {
	var buf bytes.Buffer
	err := writeEvidence(&buf, review.Evidence{
		Kind: review.EvidenceTool,
		Tool: &review.ToolEvidence{
			Tool:            "go-race",
			Version:         "go1.26",
			Command:         "go test -race ./counter",
			Output:          "WARNING: DATA RACE",
			BaseOutput:      "ok",
			PresentOnChange: true,
			PresentOnBase:   false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"evidence tool go-race version=go1.26 present_on_change=true present_on_base=false",
		"command: go test -race ./counter",
		"change: WARNING: DATA RACE",
		"base: ok",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in\n%s", want, got)
		}
	}
}
