// This file writes one review result as a Markdown pull-request comment
// (ADR-0020 §1, ADR-0027). Where each part sits is the contract in
// docs/OUTPUT_FORMATS.md.

package cli

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/review"
)

// The caps keep one comment under the 65,536 characters GitHub and Gitea accept.
// A cut is always announced in the comment, never silent.
const (
	mdMaxRows    = 20    // rows of one query shown as a table
	mdMaxOutput  = 4000  // bytes of one test, tool or fix-run output shown
	mdMaxComment = 60000 // bytes of a whole comment before the evidence is left out
)

// markdownFormatter writes the review as the comment a person reads on the pull request.
type markdownFormatter struct{}

// Write prints, top to bottom: the verdict line, then either the one-sentence
// clean statement or the findings table with one <details> block per finding,
// then the statements a reader must not miss, then the one-line foot.
//
// The comment never starts with the poster's marker line; the poster
// (lab/workflow/proven.yml, action/entrypoint.sh) puts it in front, because
// only it knows the run number and the head commit.
func (markdownFormatter) Write(w io.Writer, result review.Result) error {
	body := mdBody(result, true)
	// WHY a second pass and not a cut: a body cut at a byte could leave a fence or
	// a <details> open and swallow the rest of the pull request page. Without the
	// evidence the comment is the verdict, the table and the messages, which is small.
	if len(body) > mdMaxComment {
		body = mdBody(result, false)
	}
	_, err := io.WriteString(w, body)
	return err
}

// mdBody renders the whole comment; withEvidence false leaves each finding's proof out and says so.
func mdBody(result review.Result, withEvidence bool) string {
	var b strings.Builder
	b.WriteString(mdVerdict(result))
	b.WriteString("\n\n")
	if len(result.Findings) == 0 {
		b.WriteString(mdChecked(result))
		b.WriteString("\n\n")
	} else {
		mdTable(&b, result.Findings)
		for _, finding := range result.Findings {
			mdFinding(&b, finding, withEvidence)
		}
		if !withEvidence {
			b.WriteString("The evidence is not shown: with it this comment would pass the size limit of the pull request page. It is in the JSON and SARIF files of this run.\n\n")
		}
	}
	mdNotes(&b, result)
	b.WriteString(mdFoot(result))
	b.WriteString("\n")
	return b.String()
}

// mdVerdict is the first line: blocked or clean, with the counts a reader
// decides by. "blocked" is any finding that counts (Result.Findings holds only
// those: an accepted one is not in it). The result carries no count of dropped
// claims, so none is printed; a claim only one model makes is not in the result.
func mdVerdict(result review.Result) string {
	if len(result.Findings) == 0 {
		return "**Proven: clean** — 0 findings" + mdCountsTail(result)
	}
	var errs, warns, infos int
	for _, f := range result.Findings {
		switch f.Severity {
		case review.SeverityError:
			errs++
		case review.SeverityWarning:
			warns++
		default:
			infos++
		}
	}
	line := fmt.Sprintf("**Proven: blocked** — %d %s: %d error, %d warning, %d info",
		len(result.Findings), pluralWord(len(result.Findings), "finding", "findings"), errs, warns, infos)
	return line + mdCountsTail(result)
}

// mdCountsTail adds the counts beyond the severities: findings left unresolved,
// findings an exception accepts, and unproven opinions. Zero counts are left out.
func mdCountsTail(result review.Result) string {
	var parts []string
	unresolved := 0
	for _, f := range result.Findings {
		if f.Unresolved {
			unresolved++
		}
	}
	if unresolved > 0 {
		parts = append(parts, fmt.Sprintf("%d unresolved", unresolved))
	}
	if n := len(result.Accepted); n > 0 {
		parts = append(parts, fmt.Sprintf("%d accepted by exception", n))
	}
	if n := len(result.Opinions); n > 0 {
		parts = append(parts, fmt.Sprintf("%d %s agreed by two models, not proven", n, pluralWord(n, "opinion", "opinions")))
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

// mdChecked is the clean run's one statement: what was checked and nothing else.
// The class and the rule that chose it say how deep the review went; the services
// say whether every outside source answered. The failed ones are listed below it.
func mdChecked(result review.Result) string {
	class := mdCode(result.Class)
	if class == "" {
		class = "unknown"
	}
	reached := "every service answered"
	if len(result.FailedServices) > 0 {
		reached = "not every service answered (listed below)"
	}
	return fmt.Sprintf("Checked as review class %s (rule %s); %s. A finding needs proof to be reported, and none had it.",
		class, mdCode(result.ClassRule), reached)
}

// mdTable is the findings table: severity, finding, file:line, rule.
func mdTable(b *strings.Builder, findings []review.Finding) {
	b.WriteString("| Severity | Finding | Location | Rule |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, f := range findings {
		title := mdText(f.Title)
		if f.Unresolved {
			title += " (unresolved)"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n",
			mdText(string(f.Severity)), title, mdLocation(f.Location), mdCode(f.RuleID))
	}
	b.WriteString("\n")
}

// mdLocation is file:line as one code span; a range prints both lines.
func mdLocation(l review.Location) string {
	switch {
	case l.StartLine <= 0:
		return mdCode(l.Path)
	case l.EndLine > l.StartLine:
		return mdCode(fmt.Sprintf("%s:%d-%d", l.Path, l.StartLine, l.EndLine))
	default:
		return mdCode(fmt.Sprintf("%s:%d", l.Path, l.StartLine))
	}
}

// mdFinding writes one finding's <details> block: the message, the evidence in
// the order the finding carries it, and the fix when the result has one.
func mdFinding(b *strings.Builder, f review.Finding, withEvidence bool) {
	name := f.ID
	if name == "" {
		name = f.RuleID
	}
	// The summary line is raw HTML until the blank line after it, so it takes
	// HTML escaping only: Markdown inside it would print its own punctuation.
	fmt.Fprintf(b, "<details><summary>%s %s — %s</summary>\n\n", htmlText(string(f.Severity)), htmlText(name), htmlText(f.Title))
	if f.Message != "" {
		b.WriteString(mdText(f.Message))
		b.WriteString("\n\n")
	}
	if withEvidence {
		for _, e := range f.Evidence {
			mdEvidence(b, e)
		}
		mdFix(b, f.Fix)
	}
	b.WriteString("</details>\n\n")
}

// mdEvidence writes one piece of proof: a query with its rows, a failing test, or a tool's run.
func mdEvidence(b *strings.Builder, e review.Evidence) {
	switch e.Kind {
	case review.EvidenceQuery:
		q := e.Query
		if q == nil {
			return
		}
		fmt.Fprintf(b, "**Query** (%s, snapshot %s, form %s)\n\n", mdText(q.Engine), mdText(q.Snapshot), mdText(q.Form))
		mdFence(b, "sql", q.SQL)
		mdRows(b, q.Columns, q.Rows)
	case review.EvidenceTest:
		t := e.Test
		if t == nil {
			return
		}
		fmt.Fprintf(b, "**Failing test** %s in %s — fails on the change: %s, passes on the base: %s\n\n",
			mdCode(t.Name), mdCode(t.Path), yesNo(t.FailsOnChange), yesNo(t.PassesOnBase))
		b.WriteString("Command:\n\n")
		mdFence(b, "sh", t.Command)
		b.WriteString("Output:\n\n")
		mdFence(b, "text", mdCut(strings.TrimSpace(t.Output)))
	case review.EvidenceTool:
		t := e.Tool
		if t == nil {
			return
		}
		fmt.Fprintf(b, "**Tool** %s %s — present on the change: %s, present on the base: %s\n\n",
			mdCode(t.Tool), mdCode(t.Version), yesNo(t.PresentOnChange), yesNo(t.PresentOnBase))
		mdFence(b, "sh", t.Command)
		b.WriteString("On the change:\n\n")
		mdFence(b, "text", mdCut(strings.TrimSpace(t.Output)))
		b.WriteString("On the base:\n\n")
		mdFence(b, "text", mdCut(strings.TrimSpace(t.BaseOutput)))
	}
}

// mdRows writes a query's rows as a table. Past mdMaxRows the table stops and says how many rows it left out.
func mdRows(b *strings.Builder, columns []string, rows [][]string) {
	if len(columns) == 0 {
		return
	}
	b.WriteString("|")
	for _, c := range columns {
		b.WriteString(" " + mdCell(c) + " |")
	}
	b.WriteString("\n|")
	for range columns {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")
	shown := rows
	if len(shown) > mdMaxRows {
		shown = shown[:mdMaxRows]
	}
	for _, row := range shown {
		b.WriteString("|")
		for i := range columns {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			b.WriteString(" " + mdCell(cell) + " |")
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if n := len(rows) - len(shown); n > 0 {
		fmt.Fprintf(b, "%d more %s not shown.\n\n", n, pluralWord(n, "row", "rows"))
	}
}

// mdFix writes the fix a finding carries. Like the text writer, it shows a patch
// only when it was proved, and says each fix is a suggestion a person applies.
func mdFix(b *strings.Builder, fix *review.Fix) {
	if fix == nil {
		return
	}
	switch {
	case fix.Dependency != nil:
		d := fix.Dependency
		fmt.Fprintf(b, "**Fix** (named from the advisory table, no test ran): raise %s %s to %s. %s\n\n",
			mdCode(d.Package), mdCode(d.From), mdCode(d.To), mdText(oneLine(fix.Why)))
	case fix.Proved:
		fmt.Fprintf(b, "**Fix** (proved after %d %s; a suggestion for you to apply, Proven changes nothing): %s %s\n\n",
			fix.Attempts, attemptWord(fix.Attempts), mdText(oneLine(fix.What)), mdText(oneLine(fix.Why)))
		for _, edit := range fix.Edits {
			fmt.Fprintf(b, "Edit %s:\n\n", mdCode(edit.Path))
			mdFence(b, "diff", mdDiff(edit))
		}
		mdFixRuns(b, fix.Runs)
	default:
		b.WriteString("**Fix** not proved: " + mdText(oneLine(fix.Note)) + "\n\n")
		mdFixRuns(b, fix.Runs)
	}
}

// mdDiff shows an edit as removed and added lines, the way the text writer does.
func mdDiff(edit review.FixEdit) string {
	var out []string
	for _, line := range strings.Split(strings.TrimRight(edit.Old, "\n"), "\n") {
		out = append(out, "- "+line)
	}
	for _, line := range strings.Split(strings.TrimRight(edit.New, "\n"), "\n") {
		out = append(out, "+ "+line)
	}
	return strings.Join(out, "\n")
}

func mdFixRuns(b *strings.Builder, runs []review.FixRun) {
	for _, run := range runs {
		fmt.Fprintf(b, "- %s: %s, exit %d, %s\n", mdText(run.Label), mdCode(run.Command), run.ExitCode, mdText(run.Result))
	}
	if len(runs) > 0 {
		b.WriteString("\n")
	}
}

// mdNotes prints what the verdict alone does not say: failed services, a partial
// review, a guardrail that was off, exceptions that expired, accepted findings.
func mdNotes(b *strings.Builder, result review.Result) {
	var lines []string
	for _, failed := range result.FailedServices {
		lines = append(lines, fmt.Sprintf("- failed service %s: %s", mdCode(failed.Service), mdText(failed.Reason)))
	}
	if result.Partial {
		lines = append(lines, "- review partial: the deadline was reached")
	}
	if result.GuardrailStatement != "" {
		lines = append(lines, "- "+mdText(result.GuardrailStatement))
	}
	for _, a := range result.Accepted {
		lines = append(lines, fmt.Sprintf("- accepted by exception %s (owner %s, expires %s): %s %s",
			mdCode(a.ExceptionID), mdText(a.Owner), mdText(a.Expires), mdText(a.Finding.Title), mdLocation(a.Finding.Location)))
	}
	for _, e := range result.Expired {
		lines = append(lines, fmt.Sprintf("- exception %s expired on %s: finding %s counts again",
			mdCode(e.ExceptionID), mdText(e.Expires), mdText(e.Finding)))
	}
	if len(lines) == 0 {
		return
	}
	b.WriteString(strings.Join(lines, "\n"))
	b.WriteString("\n\n")
}

// mdFoot is the one line at the bottom: the reviewing models and whether the
// guardrail checked every call. The guardrail words follow GuardrailStatement:
// it is set exactly when the guardrail was switched off for the review.
func mdFoot(result review.Result) string {
	models := result.ReviewerModels
	if len(models) == 0 {
		for _, r := range result.Reviewers {
			models = append(models, r.Model)
		}
	}
	names := "reviewing models not stated"
	if len(models) > 0 {
		escaped := make([]string, 0, len(models))
		for _, m := range models {
			escaped = append(escaped, mdCode(m))
		}
		names = "reviewed by " + strings.Join(escaped, " and ")
	}
	guard := "guardrail checked every call"
	if result.GuardrailStatement != "" {
		guard = "guardrail was off"
	}
	return "<sub>" + names + " · " + guard + "</sub>"
}

// pluralWord picks the singular for one and the plural for any other number.
func pluralWord(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

// mdCut bounds an output at mdMaxOutput bytes on a rune boundary and says so.
func mdCut(s string) string {
	if len(s) <= mdMaxOutput {
		return s
	}
	cut := mdMaxOutput
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + fmt.Sprintf("\n… %d bytes not shown", len(s)-cut)
}

// The escaping rule for Markdown output, written once (ADR-0027). A value that
// came from the change is never written raw; it goes through exactly one of:
//
//   - mdText: prose and table cells. Whitespace runs fold to one space, so a
//     value cannot start a new line, row or block; control and bidirectional
//     characters are shown as escapes (visible); backslash, backtick, '*',
//     '[', ']', '#', '|', '~' and '@' are backslash-escaped, which stops
//     emphasis, links, images, issue references, mentions, strikethrough and
//     table pipes; '_' is escaped unless letters or digits stand on both
//     sides (a snake_case name is not emphasis); '-' and '+' at the start and
//     "1." at the start are escaped, so a value cannot open a list; "://" and
//     "www." are broken, so no bare address becomes a link; then '<', '>',
//     '&', '"' and the apostrophe become HTML entities, so no tag, comment or
//     entity can form. Order matters: the HTML entities are made last,
//     because their own '&', '#' and ';' must not be escaped again.
//   - mdCode: a code span for a path, name or command. The span's backtick
//     fence is one longer than the longest backtick run inside, and '|' is
//     written '\|' because a table splits on it before the span is read.
//   - htmlText: the <summary> line, which is raw HTML: HTML entities, and a
//     zero-width space where a mention, a reference or a link would start.
//   - mdFence: a fenced block, whose fence is longer than any backtick run in
//     the body, so the body cannot close it; no escaping is applied inside.
//
// A new writer path that prints a change value without one of these is a defect.
func mdText(s string) string {
	s = visible(strings.Join(strings.Fields(s), " "))
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		escape := strings.ContainsRune("\\`*[]#|~@", r)
		switch r {
		case '_':
			escape = !(i > 0 && i+1 < len(runes) && isWordRune(runes[i-1]) && isWordRune(runes[i+1]))
		case '-', '+':
			// A list marker only at the start of the line, and the value starts it.
			escape = i == 0
		case ':':
			escape = i+2 < len(runes) && runes[i+1] == '/' && runes[i+2] == '/'
		}
		if escape {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	out := wwwDot.ReplaceAllString(b.String(), `${1}\.`)
	out = listNumber.ReplaceAllString(out, `${1}\.`)
	return htmlEntities(out)
}

// isWordRune reports whether r is a letter or a digit.
func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// wwwDot finds "www." (GitHub links it without a scheme); listNumber finds
// "1." at the start (a value that starts the line would open an ordered list).
var (
	wwwDot     = regexp.MustCompile(`(?i)(www)\.`)
	listNumber = regexp.MustCompile(`^(\d+)\.`)
)

// htmlEntities turns the five characters HTML reads into entities.
func htmlEntities(s string) string {
	return html.EscapeString(s)
}

// htmlText is mdText's counterpart for raw HTML: fold, show controls, entities.
// Backslash escapes do not work inside raw HTML, and GitHub links, mentions and
// issue-references the text nodes of rendered HTML too, after the entities are
// read. So a zero-width space (&#8203;, written as an entity so the file shows it)
// is put where each of those starts: after '@' and '#', after the ':' of "://",
// and before the '.' of "www.". The text reads the same and no longer matches.
func htmlText(s string) string {
	runes := []rune(visible(strings.Join(strings.Fields(s), " ")))
	var b strings.Builder
	for i, r := range runes {
		switch {
		case r == '.' && i >= 3 && strings.EqualFold(string(runes[i-3:i]), "www"):
			b.WriteString(zeroWidth)
		}
		b.WriteString(htmlEntities(string(r)))
		switch {
		case r == '@' || r == '#':
			b.WriteString(zeroWidth)
		case r == ':' && i+2 < len(runes) && runes[i+1] == '/' && runes[i+2] == '/':
			b.WriteString(zeroWidth)
		}
	}
	return b.String()
}

// zeroWidth is the entity of a zero-width space.
const zeroWidth = "&#8203;"

// mdCell is mdText for one table cell. mdText already escapes '|' and folds
// line breaks, so the cell cannot end early or add a row.
func mdCell(s string) string {
	return mdText(s)
}

// mdCode writes s as a code span, empty for an empty value.
func mdCode(s string) string {
	s = visible(strings.Join(strings.Fields(s), " "))
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "|", `\|`)
	fence := strings.Repeat("`", longestRun(s, '`')+1)
	pad := ""
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		pad = " "
	}
	return fence + pad + s + pad + fence
}

// mdFence writes body in a fenced block whose fence no line of the body can close.
// Control characters are shown as escapes line by line; newlines are kept.
func mdFence(b *strings.Builder, info, body string) {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = visible(line)
	}
	text := strings.Join(lines, "\n")
	fence := strings.Repeat("`", max(3, longestRun(text, '`')+1))
	b.WriteString(fence + info + "\n" + text + "\n" + fence + "\n\n")
}

// longestRun is the length of the longest run of c in s.
func longestRun(s string, c byte) int {
	best, run := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			run++
			best = max(best, run)
		} else {
			run = 0
		}
	}
	return best
}
