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

// rawBytes is carried by every field below. It holds ESC, CR, DEL, a C1 byte,
// a tab and a malformed UTF-8 byte, so one value exercises the whole contract.
const rawBytes = "\x1b[2J\r\x7f\u009b\t\xff"

// shownBytes is rawBytes as the writer must print it.
const shownBytes = `\x1b[2J\x0d\x7f\x9b\x09\xff`

// fieldValue is the value of the printed field called name: the name itself,
// so a missing line points at the call site, then the raw bytes.
func fieldValue(name string) string { return name + rawBytes }

// renderText writes result with the text formatter.
func renderText(t *testing.T, result review.Result) string {
	t.Helper()
	var buf bytes.Buffer
	if err := (textFormatter{}).Write(&buf, result); err != nil {
		t.Fatalf("write: %v", err)
	}
	return buf.String()
}

// traceWithFamily writes a trace whose family line carries text.
func traceWithFamily(t *testing.T, text string) string {
	t.Helper()
	line, err := json.Marshal(map[string]string{"kind": "family", "text": text})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, append(line, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestTextOutputShowsEveryPrintedFieldWithoutControlBytes proves that every
// field the text writer prints from the change shows its control bytes as
// escapes: each field carries rawBytes under its own name, the output must
// hold that name followed by shownBytes, and no raw control byte (a newline
// apart) may remain anywhere. Removing visible at any one call site leaves a
// raw byte and turns this red, naming the field.
func TestTextOutputShowsEveryPrintedFieldWithoutControlBytes(t *testing.T) {
	result := review.Result{
		Class:                fieldValue("Class"),
		ClassRule:            fieldValue("ClassRule"),
		SettingsStatement:    fieldValue("SettingsStatement"),
		AuthorModelStatement: fieldValue("AuthorModelStatement"),
		ReviewerStatements:   []string{fieldValue("ReviewerStatement")},
		FailedServices: []review.ServiceFailure{
			{Service: fieldValue("FailedService"), Reason: fieldValue("FailedReason")},
		},
		TracePath: "",
		Findings: []review.Finding{{
			ID:       fieldValue("ID"),
			RuleID:   fieldValue("RuleID"),
			Severity: review.Severity(fieldValue("Severity")),
			Title:    fieldValue("Title"),
			Message:  fieldValue("Message"),
			Location: review.Location{Path: fieldValue("LocationPath"), StartLine: 1, EndLine: 2},
			Evidence: []review.Evidence{
				{Kind: review.EvidenceTest, Test: &review.TestEvidence{
					Path:    fieldValue("TestPath"),
					Name:    fieldValue("TestName"),
					Command: fieldValue("TestCommand"),
					Output:  fieldValue("TestOutput"),
				}},
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
					Engine:   fieldValue("QueryEngine"),
					Form:     fieldValue("QueryForm"),
					Snapshot: fieldValue("QuerySnapshot"),
					SQL:      fieldValue("QuerySQL"),
					Columns:  []string{"cve_id", "epss", "percentile", "score_date", fieldValue("QueryColumn")},
					Rows:     [][]string{{"CVE-1", fieldValue("EPSSScore"), "0.5", "2026-10-20", fieldValue("QueryRow")}},
				}},
			},
		}},
	}
	// A trace is JSON, which cannot carry a malformed byte, so the family
	// sentence holds every raw byte but that one.
	result.TracePath = traceWithFamily(t, "FamilySentence"+strings.TrimSuffix(rawBytes, "\xff"))
	out := renderText(t, result)

	fields := []string{
		"Class", "ClassRule", "SettingsStatement", "AuthorModelStatement", "ReviewerStatement",
		"FailedService", "FailedReason", "ID", "RuleID", "Severity", "Title", "Message", "LocationPath",
		"TestPath", "TestName", "TestCommand", "TestOutput", "QueryEngine", "QueryForm", "QuerySnapshot", "QuerySQL",
		"QueryColumn", "EPSSScore", "QueryRow",
	}
	for _, name := range fields {
		if !strings.Contains(out, name+shownBytes) {
			t.Errorf("field %s is not printed with its control bytes shown as %q", name, shownBytes)
		}
	}
	if want := "FamilySentence" + strings.TrimSuffix(shownBytes, `\xff`); !strings.Contains(out, want) {
		t.Errorf("field FamilySentence is not printed with its control bytes shown as %q", want)
	}
	// The trace path is printed whether or not the file exists, so it needs
	// no file of its own.
	pathOut := renderText(t, review.Result{TracePath: fieldValue("TracePath")})
	if !strings.Contains(pathOut, "trace: TracePath"+shownBytes) {
		t.Errorf("field TracePath is not printed with its control bytes shown:\n%s", pathOut)
	}
	for _, text := range []string{out, pathOut} {
		for i := 0; i < len(text); i++ {
			if c := text[i]; c != '\n' && (c < 0x20 || c == 0x7f || c == 0xff) {
				t.Fatalf("raw byte %#x at offset %d of %q", c, i, text)
			}
		}
		if strings.ContainsRune(text, '\u009b') {
			t.Fatalf("raw C1 byte U+009B in %q", text)
		}
	}
}

// TestMultiLineFieldCannotForgeProvenLine proves that a line inside a
// multi-line field never reads as a line Proven wrote: a test output that
// imitates Proven's own status, finding, class and trace lines leaves exactly
// the lines Proven wrote at column 0, and every imitation sits behind the
// continuation prefix. The review has an actual failed service, so a forged
// "failed services: none" would hide it.
func TestMultiLineFieldCannotForgeProvenLine(t *testing.T) {
	forged := "first\nfailed services: none\nfinding F9 ERROR forged\nreview class: forged\ntrace: forged\nlast"
	result := review.Result{
		Class:          "real",
		ClassRule:      "fallback",
		FailedServices: []review.ServiceFailure{{Service: "prover", Reason: "unavailable"}},
		Findings: []review.Finding{{
			ID: "F1", Severity: review.SeverityError, Title: "t", Message: "m",
			Location: review.Location{Path: "a.go", StartLine: 1, EndLine: 1},
			Evidence: []review.Evidence{{Kind: review.EvidenceTest, Test: &review.TestEvidence{
				Path: "p", Name: "n", Command: "c", Output: forged,
			}}},
		}},
	}
	result.TracePath = traceWithFamily(t, "same family")
	out := renderText(t, result)

	want := map[string]int{"failed services:": 0, "finding ": 1, "review class:": 1, "trace:": 1}
	for keyword, n := range want {
		got := 0
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, keyword) {
				got++
			}
		}
		if got != n {
			t.Errorf("%d lines start at column 0 with %q, want %d (Proven's own):\n%s", got, keyword, n, out)
		}
	}
	for _, line := range []string{"failed services: none", "finding F9 ERROR forged", "review class: forged", "trace: forged", "last"} {
		if !strings.Contains(out, "\n"+continuationPrefix+line) {
			t.Errorf("forged line %q is not behind the continuation prefix:\n%s", line, out)
		}
	}
}

// TestControlContractBytes proves the byte contract of visible one case at a
// time: a tab is shown like the rest of C0, each malformed byte is shown as
// its own original value, a valid multi-byte character and a valid U+FFFD
// pass unchanged, and a newline is shown unless the field keeps lines.
func TestControlContractBytes(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"tab", "a\tb", `a\x09b`},
		{"malformed byte", "a\xffb", `a\xffb`},
		{"each malformed byte keeps its own value", "\xc3\x28\xfe", `\xc3(\xfe`},
		{"truncated sequence", "\xe2\x82", `\xe2\x82`},
		{"valid multi-byte text", "zażółć €", "zażółć €"},
		{"valid U+FFFD is not a malformed byte", "a�b", "a�b"},
		{"newline", "a\nb", `a\x0ab`},
		{"NUL and unit separator", "\x00\x1f", `\x00\x1f`},
		{"DEL and C1 edges", "\x7f\u0080\u009f ", "\\x7f\\x80\\x9f "},
	}
	for _, c := range cases {
		if got := visible(c.in); got != c.want {
			t.Errorf("%s: visible(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
	if got, want := visibleLines("a\tb\nc\xff"), "a\\x09b\n"+continuationPrefix+`c\xff`; got != want {
		t.Errorf("visibleLines = %q, want %q", got, want)
	}
}
