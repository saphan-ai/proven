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
	"please ask @someone about #123 and see \\@other and \\#7 now",
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
		for _, live := range liveMentions(prose) {
			bad = append(bad, "live "+live+" (a mention or an issue reference) outside code in: "+line)
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
)

// liveMentions lists each '@' and '#' of prose that a renderer would turn into a
// mention or an issue reference. A backslash does NOT make one inert: the Markdown
// parser reads the escape away, and GitHub links the text that is left. Only the
// writer's zero-width space right after the character does, and the '#' of an
// entity (&#34;, &#8203;) is not a reference at all.
func liveMentions(prose string) []string {
	var live []string
	for i := 0; i < len(prose); i++ {
		c := prose[i]
		if c != '@' && c != '#' {
			continue
		}
		if c == '#' && i > 0 && prose[i-1] == '&' {
			continue
		}
		if strings.HasPrefix(prose[i+1:], zeroWidth) {
			continue
		}
		live = append(live, string(c))
	}
	return live
}

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

// manyFindings is n findings, each with a failing test whose output is outputBytes
// long and a query of rows rows, so a comment of them is far over the size limit.
func manyFindings(n, outputBytes, rows int) []review.Finding {
	var findings []review.Finding
	for i := 0; i < n; i++ {
		var table [][]string
		for r := 0; r < rows; r++ {
			table = append(table, []string{"row-" + strconv.Itoa(r), strings.Repeat("v", 40)})
		}
		findings = append(findings, review.Finding{
			ID: "F" + strconv.Itoa(i), Severity: review.SeverityWarning, Title: "t" + strconv.Itoa(i),
			Message: "m",
			Evidence: []review.Evidence{
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
					SQL: "select 1", Engine: "e", Snapshot: "s", Form: "f", Columns: []string{"a", "b"}, Rows: table,
				}},
				{Kind: review.EvidenceTest, Test: &review.TestEvidence{
					Name: "T", Path: "p", Command: "c", Output: strings.Repeat("x", outputBytes),
				}},
			},
		})
	}
	return findings
}

// TestMarkdownKeepsProofOfEveryFindingWhenLarge proves a result over the size
// limit still shows some proof of EVERY finding (a test and a query each), cut to
// a share and announced, never all of it dropped. The mutated field is the size of
// each finding's proof: 30 findings of 3000 bytes of output and 100 rows each.
func TestMarkdownKeepsProofOfEveryFindingWhenLarge(t *testing.T) {
	body, err := render("md", review.Result{Findings: manyFindings(30, 3000, 100)})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if len(body) > mdMaxComment {
		t.Fatalf("the comment is %d bytes, over %d", len(body), mdMaxComment)
	}
	if got := strings.Count(text, "<details>"); got != 30 {
		t.Fatalf("%d findings are detailed, want 30", got)
	}
	if got := strings.Count(text, "**Failing test**"); got != 30 {
		t.Fatalf("%d findings keep their failing test, want 30: the proof was dropped", got)
	}
	if got := strings.Count(text, "**Query**"); got != 30 {
		t.Fatalf("%d findings keep their query, want 30", got)
	}
	if got := strings.Count(text, "the full evidence is in the JSON and SARIF files of this run"); got != 30 {
		t.Fatalf("%d findings announce their cut, want 30", got)
	}
	if bad := unsafeMarkdown(text, 30); len(bad) != 0 {
		t.Fatalf("unsafe: %v", bad)
	}
}

// TestMarkdownListsTheRestByRowWhenProofCannotFit proves that when even the
// smallest proof of every finding is too much, the first findings keep their
// <details> and the rest are table rows only, with the count left out said.
// The mutated field is the number of findings: 3000.
func TestMarkdownListsTheRestByRowWhenProofCannotFit(t *testing.T) {
	body, err := render("md", review.Result{Findings: manyFindings(3000, 3000, 100)})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if len(body) > mdMaxComment {
		t.Fatalf("the comment is %d bytes, over %d", len(body), mdMaxComment)
	}
	detailed := strings.Count(text, "<details>")
	if detailed == 0 || detailed >= 3000 {
		t.Fatalf("%d findings are detailed, want some but not all", detailed)
	}
	// Every <details> keeps at least the finding's first piece of proof (its query
	// here); the least level may leave the second, the test, out, and says so.
	if got := strings.Count(text, "**Query**"); got != detailed {
		t.Fatalf("%d findings keep their query for %d <details>: a detailed finding lost its proof", got, detailed)
	}
	if !regexp.MustCompile(`\n\d+ more findings are (only in the table above|not listed here)`).MatchString(text) {
		t.Fatal("the findings left out are not counted")
	}
	if bad := unsafeMarkdown(text, detailed); len(bad) != 0 {
		t.Fatalf("unsafe: %v", bad)
	}
}

// TestMarkdownFitsWhenTooLargeEvenWithoutEvidence proves the bound holds without
// any proof at all: many findings with long messages and titles still give a body
// under the limit (the post is never refused for its size, so the previous run's
// comment is never left standing). The mutated field is the message length.
func TestMarkdownFitsWhenTooLargeEvenWithoutEvidence(t *testing.T) {
	var findings []review.Finding
	for i := 0; i < 400; i++ {
		findings = append(findings, review.Finding{
			ID: "F", Severity: review.SeverityError, Title: strings.Repeat("title ", 500),
			Message: strings.Repeat("message ", 20000),
		})
	}
	failed := []review.ServiceFailure{{Service: "svc", Reason: strings.Repeat("r", 200000)}}
	body, err := render("md", review.Result{Findings: findings, FailedServices: failed})
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > mdMaxComment {
		t.Fatalf("the comment is %d bytes, over %d", len(body), mdMaxComment)
	}
	if !strings.HasPrefix(string(body), "**Proven: blocked** — 400 findings") {
		t.Fatalf("the verdict does not lead the comment\n%.200s", body)
	}
	if bad := unsafeMarkdown(string(body), strings.Count(string(body), "<details>")); len(bad) != 0 {
		t.Fatalf("unsafe: %v", bad)
	}
}

// TestMarkdownMentionsStayInertAfterTheEscape proves a '@user' or '#123' written
// mid-sentence in any field is followed by a zero-width space even where the
// writer also puts a backslash, because the renderer reads the backslash away and
// links what is left. The attack is in hostile (the mentions are its fourth entry).
func TestMarkdownMentionsStayInertAfterTheEscape(t *testing.T) {
	got := mdText("ask @someone about #123 now")
	want := `ask \@` + zeroWidth + `someone about \#` + zeroWidth + `123 now`
	if got != want {
		t.Fatalf("mdText = %q, want %q", got, want)
	}
}

// TestMarkdownMentionCheckBitesWhenTheZeroWidthIsRemoved is the negative control of
// the mention rule (law #27). It mutates exactly the zero-width space after '@'
// and '#' in a body that passes the check, and the same check must turn red.
func TestMarkdownMentionCheckBitesWhenTheZeroWidthIsRemoved(t *testing.T) {
	body, err := render("md", hostileResult())
	if err != nil {
		t.Fatal(err)
	}
	if bad := unsafeMarkdown(string(body), 1); len(bad) != 0 {
		t.Fatalf("the control starts from a body that is already unsafe: %v", bad)
	}
	mutated := strings.NewReplacer(`\@`+zeroWidth, `\@`, `\#`+zeroWidth, `\#`).Replace(string(body))
	if mutated == string(body) {
		t.Fatal("the mutation changed nothing: no escaped mention in the body")
	}
	if bad := unsafeMarkdown(mutated, 1); len(bad) == 0 {
		t.Fatal("a mention without its zero-width space passed the check")
	}
}

// TestMarkdownStatesWhatWasChecked proves the comment carries the statements the
// text writer prints, names a failed service whether the run is clean or not, and
// that a clean comment says what was checked in plain words. The mutated fields
// are the statements and the failed services.
func TestMarkdownStatesWhatWasChecked(t *testing.T) {
	statements := review.Result{
		PolicyStatement:      "policy: org.yaml",
		SettingsStatement:    "the change edits proven.yaml",
		AuthorModelStatement: "author model not stated",
		ReviewerStatements:   []string{"reviewer model-a: family x"},
		FailedServices:       []review.ServiceFailure{{Service: "zetaris", Reason: "ZETARIS_REST_URL is empty"}},
		Class:                "standard", ClassRule: "fallback",
		SkippedPaths: []string{"link"},
	}
	for name, result := range map[string]review.Result{
		"clean":    statements,
		"findings": withFinding(statements),
	} {
		body, err := render("md", result)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"policy: org.yaml", "the change edits proven.yaml", "author model not stated",
			"reviewer model-a: family x", "failed service `zetaris`: ZETARIS_REST_URL is empty"} {
			if !strings.Contains(string(body), want) {
				t.Errorf("%s: the comment does not say %q\n%s", name, want, body)
			}
		}
	}
	clean, _ := render("md", statements)
	for _, want := range []string{"review class `standard` (rule `fallback`)", "1 tracked path was skipped", "these services did not answer: `zetaris`"} {
		if !strings.Contains(string(clean), want) {
			t.Errorf("the clean sentence does not say %q\n%s", want, clean)
		}
	}
	answered, _ := render("md", review.Result{Class: "standard", ClassRule: "fallback"})
	if !strings.Contains(string(answered), "every service answered") {
		t.Errorf("the clean sentence does not say every service answered\n%s", answered)
	}
}

func withFinding(r review.Result) review.Result {
	r.Findings = []review.Finding{{ID: "F1", Severity: review.SeverityError, Title: "t"}}
	return r
}

// TestMarkdownPipeIsEscapedOnlyInATable proves '|' is written '\|' in a table cell
// and as itself in a code span elsewhere, where the backslash would print.
// The mutated field is a pipe in the location (a cell) and in a fix command (prose).
func TestMarkdownPipeIsEscapedOnlyInATable(t *testing.T) {
	result := review.Result{Findings: []review.Finding{{
		ID: "F1", Severity: review.SeverityError, Title: "t",
		Location: review.Location{Path: "a|b.go", StartLine: 1},
		Fix:      &review.Fix{Note: "n", Runs: []review.FixRun{{Label: "run", Command: "go test | tee", Result: "r"}}},
	}}}
	body, err := render("md", result)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "`a\\|b.go:1`") {
		t.Errorf("the table cell does not escape the pipe\n%s", text)
	}
	if !strings.Contains(text, "`go test | tee`") || strings.Contains(text, `test \|`) {
		t.Errorf("the pipe outside a table is escaped or lost\n%s", text)
	}
	if bad := unsafeMarkdown(text, 1); len(bad) != 0 {
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

// TestMarkdownNeverTradesAllProofForALongerTable proves a body that fits only
// because its table left no room for any <details> is not taken while a shorter
// table keeps some proof. The mutated field is the number of findings whose rows
// nearly fill the budget: 175 findings with 300-byte titles and 2000-byte messages
// (review r4 of round 2 measured 0 <details> at 175 and 180).
func TestMarkdownNeverTradesAllProofForALongerTable(t *testing.T) {
	for _, n := range []int{175, 180} {
		var findings []review.Finding
		for i := 0; i < n; i++ {
			findings = append(findings, review.Finding{
				ID: "F" + strconv.Itoa(i), Severity: review.SeverityError, Title: strings.Repeat("t", mdMaxTitle),
				Message: strings.Repeat("m", mdMaxMessage),
				Evidence: []review.Evidence{{Kind: review.EvidenceTest, Test: &review.TestEvidence{
					Name: "T", Path: "p", Command: "c", Output: "fails",
				}}},
			})
		}
		body, err := render("md", review.Result{Findings: findings})
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if len(body) > mdMaxComment {
			t.Fatalf("n=%d: the comment is %d bytes, over %d", n, len(body), mdMaxComment)
		}
		if strings.Count(text, "<details>") == 0 || strings.Count(text, "**Failing test**") == 0 {
			t.Fatalf("n=%d: the comment keeps no proof at all", n)
		}
	}
}

// TestMarkdownKeepsTheLeastProofOfAFindingWithManyPieces proves one finding whose
// every level is too large, because it holds many pieces of evidence and many fix
// edits rather than long ones, still keeps its first piece of proof, says how many
// pieces were left out, and fits. The mutated field is the number of pieces: 400
// tests and a proved fix of 200 edits (review r5 of round 2).
func TestMarkdownKeepsTheLeastProofOfAFindingWithManyPieces(t *testing.T) {
	var evidence []review.Evidence
	for i := 0; i < 400; i++ {
		evidence = append(evidence, review.Evidence{Kind: review.EvidenceTest, Test: &review.TestEvidence{
			Name: "T" + strconv.Itoa(i), Path: "p", Command: "c", Output: strings.Repeat("x", 1000),
		}})
	}
	var edits []review.FixEdit
	for i := 0; i < 200; i++ {
		edits = append(edits, review.FixEdit{Path: "p" + strconv.Itoa(i), Old: strings.Repeat("o", 500), New: "n"})
	}
	f := review.Finding{ID: "F1", Severity: review.SeverityError, Title: "many", Message: "m", Evidence: evidence,
		Fix: &review.Fix{Proved: true, Attempts: 1, What: "w", Why: "y", Edits: edits}}
	body, err := render("md", review.Result{Findings: []review.Finding{f}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if len(body) > mdMaxComment {
		t.Fatalf("the comment is %d bytes, over %d", len(body), mdMaxComment)
	}
	if strings.Count(text, "<details>") != 1 || !strings.Contains(text, "**Failing test** `T0`") {
		t.Fatalf("the finding lost its proof\n%.600s", text)
	}
	if !strings.Contains(text, "_cut: 399 more pieces of evidence, the fix not shown") {
		t.Fatalf("the pieces left out are not counted\n%.2000s", text)
	}
	if len(f.Evidence) != 400 || len(f.Fix.Edits) != 200 {
		t.Fatal("the writer changed the finding it rendered")
	}
}

// TestMarkdownLeastProofCountsTheColumnsItLeavesOut proves the least proof of a
// query announces the columns past the fourth and the bytes it shortened, so no
// part of the proof goes unsaid (review r7 of round 3). The mutated field is the
// column count: 6 columns name "2 more query columns"; 4 name none.
func TestMarkdownLeastProofCountsTheColumnsItLeavesOut(t *testing.T) {
	finding := func(cols int) review.Finding {
		var names, row []string
		for i := 0; i < cols; i++ {
			names = append(names, "c"+strconv.Itoa(i))
			row = append(row, strings.Repeat("v", 70000))
		}
		return review.Finding{ID: "F1", Severity: review.SeverityError, Title: "wide", Message: "m",
			Evidence: []review.Evidence{{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
				SQL: "select 1", Engine: "e", Snapshot: "s", Form: "sql", Columns: names, Rows: [][]string{row},
			}}}}
	}
	for _, tc := range []struct {
		cols int
		want bool
	}{{6, true}, {4, false}} {
		body, err := render("md", review.Result{Findings: []review.Finding{finding(tc.cols)}})
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if len(body) > mdMaxComment {
			t.Fatalf("cols=%d: the comment is %d bytes, over %d", tc.cols, len(body), mdMaxComment)
		}
		if got := strings.Contains(text, "2 more query columns"); got != tc.want {
			t.Errorf("cols=%d: the cut line names 2 more query columns: %v, want %v\n%.1500s", tc.cols, got, tc.want, text)
		}
		if !regexp.MustCompile(`_cut: \d+ more rows? and [1-9]\d* bytes? not shown`).MatchString(text) {
			t.Errorf("cols=%d: the shortened cells are not counted in bytes\n%.1500s", tc.cols, text)
		}
	}
}
