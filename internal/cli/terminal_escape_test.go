package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestTextOutputCarriesNoControlByteFromTheChange proves the property that
// output to a terminal carries no control byte taken from the change: every
// field the text writer prints from the change shows ESC, CR, DEL and a C1
// byte as a visible escape, and the one newline an output field may hold is
// kept. A raw byte would let the change repaint or rewrite the reader's
// terminal.
func TestTextOutputCarriesNoControlByteFromTheChange(t *testing.T) {
	const raw = "x\x1b[2Jy\rz\x7f\u009bw"
	const shown = `x\x1b[2Jy\x0dz\x7f\x9bw`
	result := review.Result{
		FailedServices: []review.ServiceFailure{{Service: "prover", Reason: "r" + raw}},
		Findings: []review.Finding{{
			ID: "F" + raw, RuleID: "rule" + raw, Severity: review.SeverityError,
			Title: "title" + raw, Message: "message" + raw,
			Location: review.Location{Path: "pkg" + raw + "/a.go", StartLine: 1, EndLine: 2},
			Evidence: []review.Evidence{
				{Kind: review.EvidenceTest, Test: &review.TestEvidence{
					Path: "t" + raw, Name: "name" + raw, Command: "cmd" + raw,
					Output: "line one" + raw + "\nline two\x1b[0m",
				}},
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
					SQL: "sql" + raw, Engine: "e" + raw, Snapshot: "s" + raw,
					Columns: []string{"c" + raw}, Rows: [][]string{{"v" + raw, "w"}},
				}},
			},
		}},
	}
	rev := &captureReviewer{res: result}
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitFindings && code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	out := stdout.String()
	t.Logf("raw output: %q", out)
	for i, r := range out {
		if r == '\n' {
			continue
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			t.Fatalf("raw control byte %U at offset %d in %q", r, i, out)
		}
	}
	for _, want := range []string{"title" + shown, "pkg" + shown + "/a.go", "line two\\x1b[0m"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output lacks %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, shown+"\nline two") {
		t.Fatalf("the newline inside the output field was not kept:\n%s", out)
	}
}
