package cli

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// updateGolden rewrites the golden files: go test ./internal/cli -run TestMarkdownGolden -update-golden.
var updateGolden = flag.Bool("update-golden", false, "rewrite the Markdown golden files")

// goldenPath names one of the two comments a person sees on a pull request, under
// testdata/golden/pr-comment. They are the contract: a change to the writer that
// moves a character shows up there as a diff to read, and -update-golden writes
// the new one for the review to read like code.
func goldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(moduleRoot(t), "testdata", "golden", "pr-comment", name)
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := goldenPath(t, name)
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s: %v (-update-golden writes it)", name, err)
	}
	if string(want) != got {
		t.Fatalf("golden %s differs\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

// TestMarkdownGoldenRunWithFindings proves the comment of a run with findings:
// the verdict line, the table, a <details> block per finding with its test and
// its query rows, and the foot. The input is the real sample run on the
// discount fixture, so the golden is what the lab's pull request shows.
// The property fails when any part of that layout moves.
func TestMarkdownGoldenRunWithFindings(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	env := envOf(map[string]string{"PROVEN_LOG_DIR": t.TempDir()})
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", "--md", target}, env, &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("exit %d\nstderr: %s\nstdout: %s", code, stderr.String(), stdout.String())
	}
	assertGolden(t, "findings.md", stdout.String())
}

// TestMarkdownGoldenCleanRun proves a clean run is a short comment that says
// what was checked and carries no table and no <details>. The mutated field is
// the findings list: empty here, so any table or block turns the golden red.
func TestMarkdownGoldenCleanRun(t *testing.T) {
	body, err := render("md", review.Result{
		Class:          "standard",
		ClassRule:      "fallback",
		ReviewerModels: []string{"model-a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "<details>") || strings.Contains(string(body), "| Severity") {
		t.Fatalf("a clean comment carries a table or a block\n%s", body)
	}
	assertGolden(t, "clean.md", string(body))
}

// hostile is a value from a change that tries, in one string, every way out of
// the place the writer puts it. Each entry is a payload on its own line of the
// attack so a failure names the one that got through.
var hostile = []string{
	"</details>",
	"<img src=x onerror=alert(1)>",
	"<!-- comment",
	"| extra | cell |",
	"```",
	"[click](http://evil.example/x)",
	"![pixel](http://evil.example/p.png)",
	"http://evil.example/bare",
	"www.evil.example",
	"@everyone",
	"# heading",
	"&lt;b&gt;",
}

// hostileResult puts the whole attack in every field of every kind the writer
// prints from the change: title, message, path, rule, model, query, rows, test
// output, tool output, fix text.
func hostileResult() review.Result {
	attack := strings.Join(hostile, "\n")
	return review.Result{
		Class:          attack,
		ClassRule:      attack,
		ReviewerModels: []string{attack},
		FailedServices: []review.ServiceFailure{{Service: attack, Reason: attack}},
		Findings: []review.Finding{{
			ID: "F1", RuleID: attack, Title: attack, Message: attack,
			Severity: review.SeverityError,
			Location: review.Location{Path: attack, StartLine: 3},
			Evidence: []review.Evidence{
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
					SQL: attack, Engine: attack, Snapshot: attack, Form: attack,
					Columns: []string{attack, "b"}, Rows: [][]string{{attack, attack}},
				}},
				{Kind: review.EvidenceTest, Test: &review.TestEvidence{
					Path: attack, Name: attack, Command: attack, Output: attack,
				}},
				{Kind: review.EvidenceTool, Tool: &review.ToolEvidence{
					Tool: attack, Version: attack, Command: attack, Output: attack, BaseOutput: attack,
				}},
			},
			Fix: &review.Fix{Proved: true, Attempts: 1, What: attack, Why: attack,
				Edits: []review.FixEdit{{Path: attack, Old: attack, New: attack}}},
		}},
	}
}

// unsafeMarkdown reports every way the body breaks the rules the writer
// promises (docs/OUTPUT_FORMATS.md, ADR-0027), one line each. It reads the body
// as a Markdown renderer would: fences first, then code spans, then the rest.
// It is a separate reader from the writer, so a writer that skips one escape
// cannot also skip the check.
func unsafeMarkdown(body string, findings int) []string {
	var bad []string
	opened, closed, tableWidth := 0, 0, 0
	fenceOpen := ""
	for _, line := range strings.Split(body, "\n") {
		if fenceOpen != "" {
			if strings.HasPrefix(line, fenceOpen) && strings.Trim(line, "`") == "" {
				fenceOpen = ""
			}
			continue
		}
		if strings.HasPrefix(line, "```") {
			fenceOpen = line[:len(line)-len(strings.TrimLeft(line, "`"))]
			continue
		}
		prose := stripCodeSpans(line)
		// A table is one width from its header to its last row: a value that
		// closed a cell would make a row wider than its header.
		if strings.HasPrefix(line, "|") {
			if width := tableCells(line); tableWidth != 0 && width != tableWidth {
				bad = append(bad, "table row of width "+itoa(width)+" in a table of width "+itoa(tableWidth)+": "+line)
			} else {
				tableWidth = width
			}
		} else {
			tableWidth = 0
		}
		// Only a tag outside code counts: inside a fence or a span it is text.
		opened += strings.Count(prose, "<details>")
		closed += strings.Count(prose, "</details>")
		// Outside code, '<' may only start the writer's own tags.
		for _, at := range regexp.MustCompile(`<[^>]*>?`).FindAllString(prose, -1) {
			if !allowedTag.MatchString(at) {
				bad = append(bad, "raw tag outside code: "+at)
			}
		}
		if unescapedAt.MatchString(prose) {
			bad = append(bad, "live @ (a mention) outside code in: "+line)
		}
		for _, marker := range []string{"http://", "https://", "www."} {
			if strings.Contains(prose, marker) {
				bad = append(bad, "live "+marker+" outside code in: "+line)
			}
		}
		// A <details> line is an HTML block: Markdown link syntax is inert in it, so only
		// the rules a renderer applies to HTML text (above) bind there.
		rawHTML := strings.HasPrefix(line, "<details>")
		// "](" is a link only when the bracket is not backslash-escaped.
		if !rawHTML && (unescapedLink.MatchString(prose) || strings.Contains(prose, "![")) {
			bad = append(bad, "live link or image outside code in: "+line)
		}
		if strings.HasPrefix(prose, "#") {
			bad = append(bad, "a line starts a heading: "+line)
		}
	}
	if fenceOpen != "" {
		bad = append(bad, "a fence is never closed")
	}
	if opened != findings || closed != findings {
		bad = append(bad, "<details> opened "+itoa(opened)+" and closed "+itoa(closed)+" times for "+itoa(findings)+" findings")
	}
	return bad
}

var (
	allowedTag    = regexp.MustCompile(`^</?(details|summary|sub)>$`)
	unescapedLink = regexp.MustCompile(`(^|[^\\])\]\(`)
	// '@' is a mention unless a backslash comes before it or the writer's zero-width space after it.
	unescapedAt = regexp.MustCompile(`(^|[^\\])@($|[^&])`)
)

func itoa(n int) string { return strconv.Itoa(n) }

// stripCodeSpans removes `...` spans (any backtick run length) from a line.
func stripCodeSpans(line string) string {
	var out strings.Builder
	for i := 0; i < len(line); {
		if line[i] == '\\' && i+1 < len(line) {
			// An escaped character is text, an escaped backtick included.
			out.WriteString(line[i : i+2])
			i += 2
			continue
		}
		if line[i] != '`' {
			out.WriteByte(line[i])
			i++
			continue
		}
		j := i
		for j < len(line) && line[j] == '`' {
			j++
		}
		run := line[i:j]
		end := strings.Index(line[j:], run)
		if end < 0 {
			out.WriteString(run)
			i = j
			continue
		}
		i = j + end + len(run)
	}
	return out.String()
}

// tableCells counts the cells of a table row, not counting an escaped pipe.
func tableCells(line string) int {
	n := 0
	for i := 1; i < len(line); i++ {
		switch {
		case line[i] == '\\':
			i++
		case line[i] == '|':
			n++
		}
	}
	return n
}

// TestMarkdownEscapesEveryValueFromTheChange proves a change cannot close a
// table cell, add a table row, open a tag, end a <details>, break out of a
// fence, or inject a link, an image, a heading or a mention, through any field
// the writer prints. The attack sits in every field at once.
func TestMarkdownEscapesEveryValueFromTheChange(t *testing.T) {
	body, err := render("md", hostileResult())
	if err != nil {
		t.Fatal(err)
	}
	if bad := unsafeMarkdown(string(body), 1); len(bad) != 0 {
		t.Fatalf("the comment is not safe:\n%s\n--- body\n%s", strings.Join(bad, "\n"), body)
	}
	// The findings table keeps its shape: header, rule and one row for the one
	// finding, so the attack added no row; four cells each, so it closed no cell.
	// (unsafeMarkdown holds every other table, the query rows, to one width.)
	table, _, _ := strings.Cut(string(body), "<details>")
	rows := 0
	for _, line := range strings.Split(table, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		rows++
		if cells := tableCells(line); cells != 4 {
			t.Fatalf("table line has %d cells, want 4: %s", cells, line)
		}
	}
	if rows != 3 {
		t.Fatalf("the findings table has %d lines, want 3\n%s", rows, body)
	}
}

// TestMarkdownEscapingCheckBitesOnOneUnescapedField is the negative control of
// the escaping test (law #27). It mutates exactly one field: the finding's
// title, which the table prints through mdText. With that one value written
// raw, the same check that passes on the real body must turn red. A check that
// stays green here would prove nothing about the escaping above.
func TestMarkdownEscapingCheckBitesOnOneUnescapedField(t *testing.T) {
	result := hostileResult()
	body, err := render("md", result)
	if err != nil {
		t.Fatal(err)
	}
	if bad := unsafeMarkdown(string(body), 1); len(bad) != 0 {
		t.Fatalf("the control starts from a body that is already unsafe: %v", bad)
	}
	escaped := mdText(result.Findings[0].Title)
	raw := strings.Join(strings.Fields(result.Findings[0].Title), " ")
	mutated := strings.Replace(string(body), "| "+escaped+" |", "| "+raw+" |", 1)
	if mutated == string(body) {
		t.Fatal("the mutation changed nothing: the title cell was not found")
	}
	if bad := unsafeMarkdown(mutated, 1); len(bad) == 0 {
		t.Fatal("one unescaped title passed the check")
	}
}

// TestMarkdownFenceCannotBeClosedByItsBody proves a body holding a run of
// backticks longer than three still sits inside its own fence: the fence is
// longer than any run inside. The mutated field is the test output.
func TestMarkdownFenceCannotBeClosedByItsBody(t *testing.T) {
	result := review.Result{Findings: []review.Finding{{
		ID: "F1", Severity: review.SeverityError, Title: "t",
		Evidence: []review.Evidence{{Kind: review.EvidenceTest, Test: &review.TestEvidence{
			Name: "T", Path: "p_test.go", Command: "go test", Output: "a\n`````\n</details>\nb",
		}}},
	}}}
	body, err := render("md", result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "``````text\na\n`````\n</details>\nb\n``````\n") {
		t.Fatalf("the fence is not longer than the run inside it\n%s", body)
	}
	if bad := unsafeMarkdown(string(body), 1); len(bad) != 0 {
		t.Fatalf("unsafe: %v\n%s", bad, body)
	}
}

// TestMarkdownShowsControlBytesAsEscapes proves a terminal escape from the
// change is shown as text, as the text writer does, in a cell and in a fence.
func TestMarkdownShowsControlBytesAsEscapes(t *testing.T) {
	result := review.Result{Findings: []review.Finding{{
		ID: "F1", Severity: review.SeverityError, Title: "red\x1b[31m",
		Evidence: []review.Evidence{{Kind: review.EvidenceTest, Test: &review.TestEvidence{
			Name: "T", Path: "p", Command: "c", Output: "boom\x1b[2J",
		}}},
	}}}
	body, err := render("md", result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(string(body), 0x1b) {
		t.Fatalf("an ESC byte reached the comment\n%q", body)
	}
}

// TestMarkdownLeavesEvidenceOutWhenTooLarge proves a comment past the size
// limit is rewritten without the evidence, closed and announced, instead of
// being cut. The mutated field is the size of a test's output, repeated.
func TestMarkdownLeavesEvidenceOutWhenTooLarge(t *testing.T) {
	var findings []review.Finding
	for i := 0; i < 30; i++ {
		findings = append(findings, review.Finding{
			ID: "F", Severity: review.SeverityWarning, Title: "t",
			Evidence: []review.Evidence{{Kind: review.EvidenceTest, Test: &review.TestEvidence{
				Name: "T", Path: "p", Command: "c", Output: strings.Repeat("x", 3000),
			}}},
		})
	}
	body, err := render("md", review.Result{Findings: findings})
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > mdMaxComment {
		t.Fatalf("the comment is %d bytes, over %d", len(body), mdMaxComment)
	}
	if !strings.Contains(string(body), "The evidence is not shown") {
		t.Fatal("the cut is not announced")
	}
	if bad := unsafeMarkdown(string(body), 30); len(bad) != 0 {
		t.Fatalf("unsafe: %v", bad)
	}
}

// TestMarkdownFlagsAndFormat proves --md is a format beside --text/--json/--sarif
// and --md-output writes the same body as --md to a file, refusing a second
// format flag. PROVEN_OUTPUT_FORMAT=md picks it too.
func TestMarkdownFlagsAndFormat(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	dir := t.TempDir()
	t.Chdir(dir)
	logs := t.TempDir()
	out := filepath.Join(dir, "comment.md")
	env := envOf(map[string]string{"PROVEN_LOG_DIR": logs})
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"scan", "--sample", "--json", "--md-output", out, target}, env, &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	file, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := Run([]string{"scan", "--sample", "--md", target}, env, &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if !bytes.Equal(file, stdout.Bytes()) {
		t.Fatal("--md-output and --md wrote different bodies")
	}
	stdout.Reset()
	env = envOf(map[string]string{"PROVEN_LOG_DIR": logs, "PROVEN_OUTPUT_FORMAT": "md"})
	if code := Run([]string{"scan", "--sample", target}, env, &stdout, &stderr, nil); code != exitOK || !bytes.Equal(file, stdout.Bytes()) {
		t.Fatalf("PROVEN_OUTPUT_FORMAT=md: exit %d", code)
	}
	stderr.Reset()
	if code := Run([]string{"scan", "--sample", "--md", "--json", target}, env, &stdout, &stderr, nil); code == exitOK || !strings.Contains(stderr.String(), "--md") {
		t.Fatalf("two formats: exit %d, stderr %s", code, stderr.String())
	}
}
