// This file writes one review result as one self-contained HTML file a judge can
// open (ADR-0033 point 8, ADR-0027). Where each part sits is the contract in
// docs/OUTPUT_FORMATS.md.

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

// htmlStyle marks a value from the change for HTML text. The sentences shared
// with the Markdown writer (verdictLine, checkedIn, fixSentence, notesIn, footIn)
// take it, so both writers say the same words.
var htmlStyle = inlineStyle{text: htmlText, code: htmlCode, unbounded: true}

// htmlFormatter writes the review as the page a judge opens in a browser.
type htmlFormatter struct{}

// Write prints one HTML document, top to bottom: the verdict and the summary
// table, one card per finding (its message, how we know, its evidence folded
// away until it is opened), the opinions two models agreed on and nobody proved,
// the notes a reader must not miss, and the one-line foot.
//
// WHY the page carries no script, no link element and no address: a judge opens
// it from a file or an e-mail, possibly offline, and a page that fetched anything
// would leak that it was opened and would not render the same twice. The one
// href it holds is an in-page "#finding-N" anchor Proven makes itself. A value
// from the change is escaped as text by one of the helpers below and is never
// part of markup or of an attribute.
func (htmlFormatter) Write(w io.Writer, result review.Result) error {
	_, err := io.WriteString(w, htmlPage(result))
	return err
}

// htmlPage renders the whole document.
func htmlPage(result review.Result) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>Proven review</title>\n<style>\n")
	b.WriteString(htmlCSS)
	b.WriteString("</style>\n</head>\n<body>\n<main>\n")
	b.WriteString("<h1>Proven review</h1>\n")
	htmlVerdict(&b, result)
	if len(result.Findings) == 0 {
		b.WriteString("<p>" + checkedIn(htmlStyle, result) + "</p>\n")
	} else {
		htmlSummary(&b, result.Findings)
		for i, f := range result.Findings {
			htmlCard(&b, i+1, f)
		}
	}
	htmlOpinions(&b, result.Opinions)
	htmlNotes(&b, result)
	b.WriteString("<footer>" + footIn(htmlStyle, result) + "</footer>\n")
	b.WriteString("</main>\n</body>\n</html>\n")
	return b.String()
}

// htmlVerdict is the first block: the verdict and its counts, in the colour of
// what it says. The class comes from the result, never from the wording.
func htmlVerdict(b *strings.Builder, result review.Result) {
	class := "clean"
	switch {
	case len(result.Findings) > 0:
		class = "blocked"
	case len(mdModelsDown(result)) > 0:
		class = "not-reviewed"
	}
	head, tail := verdictLine(result)
	// htmlText folds the leading space of the tail away, so the space between the
	// verdict and its dash is written here.
	fmt.Fprintf(b, "<p class=\"verdict %s\"><strong>%s</strong> %s</p>\n", class, htmlText(head), htmlText(tail))
}

// htmlSummary is the table of findings: severity, finding, file:line. The
// finding is the in-page anchor of its card.
func htmlSummary(b *strings.Builder, findings []review.Finding) {
	b.WriteString("<table class=\"summary\">\n<thead><tr><th>Severity</th><th>Finding</th><th>Location</th></tr></thead>\n<tbody>\n")
	for i, f := range findings {
		fmt.Fprintf(b, "<tr><td>%s</td><td><a href=\"#finding-%d\">%s</a></td><td>%s</td></tr>\n",
			htmlText(mdSeverity(f.Severity)), i+1, htmlText(htmlTitle(f)), htmlCode(mdPlace(f.Location)))
	}
	b.WriteString("</tbody>\n</table>\n")
}

// htmlTitle is a finding's title, marked when the review left it unresolved.
func htmlTitle(f review.Finding) string {
	if f.Unresolved {
		return f.Title + " (unresolved)"
	}
	return f.Title
}

// htmlSeverityClass maps a severity to one of three fixed class names, so the
// attribute never holds a value from the change.
func htmlSeverityClass(sev review.Severity) string {
	switch sev {
	case review.SeverityError:
		return "error"
	case review.SeverityWarning:
		return "warning"
	default:
		return "info"
	}
}

// htmlCard is one finding: severity, title, place, the message, how we know, and
// the evidence folded away until opened.
func htmlCard(b *strings.Builder, n int, f review.Finding) {
	fmt.Fprintf(b, "<section class=\"card %s\" id=\"finding-%d\">\n", htmlSeverityClass(f.Severity), n)
	fmt.Fprintf(b, "<h2><span class=\"severity\">%s</span> %s</h2>\n", htmlText(mdSeverity(f.Severity)), htmlText(htmlTitle(f)))
	fmt.Fprintf(b, "<p class=\"place\">%s</p>\n", htmlCode(mdPlace(f.Location)))
	if f.Message != "" {
		b.WriteString("<p>" + htmlText(f.Message) + "</p>\n")
	}
	b.WriteString("<h3>How we know</h3>\n<ul>\n")
	known := 0
	for _, e := range f.Evidence {
		if sentence := htmlHowWeKnow(e); sentence != "" {
			b.WriteString("<li>" + sentence + "</li>\n")
			known++
		}
	}
	if known == 0 {
		b.WriteString("<li>No test, tool or query is attached to this finding.</li>\n")
	}
	b.WriteString("</ul>\n")
	if len(f.Evidence) > 0 || f.Fix != nil {
		b.WriteString("<details>\n<summary>Evidence and fix</summary>\n")
		for _, e := range f.Evidence {
			htmlEvidence(b, e)
		}
		htmlFix(b, f.Fix)
		b.WriteString("</details>\n")
	}
	b.WriteString("</section>\n")
}

// htmlHowWeKnow is one sentence in words for one piece of evidence, from the
// facts the evidence records and nothing the review did not measure.
func htmlHowWeKnow(e review.Evidence) string {
	switch e.Kind {
	case review.EvidenceTest:
		t := e.Test
		if t == nil {
			return ""
		}
		if t.FailsOnChange && t.PassesOnBase {
			return "A test fails on the change and passes on the base: " + htmlCode(t.Name)
		}
		return fmt.Sprintf("A test was run, %s: fails on the change: %s, passes on the base: %s",
			htmlCode(t.Name), yesNo(t.FailsOnChange), yesNo(t.PassesOnBase))
	case review.EvidenceQuery:
		q := e.Query
		if q == nil {
			return ""
		}
		return fmt.Sprintf("A query to %s returned %d %s (snapshot %s)",
			htmlText(queryEngineName(q.Engine)), len(q.Rows), pluralWord(len(q.Rows), "row", "rows"), htmlText(q.Snapshot))
	case review.EvidenceTool:
		t := e.Tool
		if t == nil {
			return ""
		}
		name := htmlCode(t.Tool) + " " + htmlCode(t.Version)
		if t.PresentOnChange && !t.PresentOnBase {
			return name + " reports it on the change and not on the base"
		}
		return fmt.Sprintf("%s was run: present on the change: %s, present on the base: %s",
			name, yesNo(t.PresentOnChange), yesNo(t.PresentOnBase))
	}
	return ""
}

// htmlEvidence writes one piece of proof inside the card's <details>: a query
// with its rows as a table, or a test's or tool's command and output.
func htmlEvidence(b *strings.Builder, e review.Evidence) {
	switch e.Kind {
	case review.EvidenceQuery:
		q := e.Query
		if q == nil {
			return
		}
		fmt.Fprintf(b, "<h4>Query</h4>\n<p>(%s)</p>\n", htmlText(querySource(q)))
		b.WriteString("<pre>" + htmlPre(q.SQL) + "</pre>\n")
		htmlRows(b, q.Columns, q.Rows)
	case review.EvidenceTest:
		t := e.Test
		if t == nil {
			return
		}
		fmt.Fprintf(b, "<h4>Failing test</h4>\n<p>%s in %s — fails on the change: %s, passes on the base: %s</p>\n",
			htmlCode(t.Name), htmlCode(t.Path), yesNo(t.FailsOnChange), yesNo(t.PassesOnBase))
		b.WriteString("<p>Command:</p>\n<pre>" + htmlPre(t.Command) + "</pre>\n")
		b.WriteString("<p>Output:</p>\n<pre>" + htmlPre(strings.TrimSpace(t.Output)) + "</pre>\n")
	case review.EvidenceTool:
		t := e.Tool
		if t == nil {
			return
		}
		fmt.Fprintf(b, "<h4>Tool</h4>\n<p>%s %s — present on the change: %s, present on the base: %s</p>\n",
			htmlCode(t.Tool), htmlCode(t.Version), yesNo(t.PresentOnChange), yesNo(t.PresentOnBase))
		b.WriteString("<pre>" + htmlPre(t.Command) + "</pre>\n")
		b.WriteString("<p>On the change:</p>\n<pre>" + htmlPre(strings.TrimSpace(t.Output)) + "</pre>\n")
		b.WriteString("<p>On the base:</p>\n<pre>" + htmlPre(strings.TrimSpace(t.BaseOutput)) + "</pre>\n")
	}
}

// htmlRows writes a query's rows as a table. A row shorter than the columns
// gets empty cells; a cell past the last column is not shown, as in Markdown.
func htmlRows(b *strings.Builder, columns []string, rows [][]string) {
	if len(columns) == 0 {
		return
	}
	b.WriteString("<div class=\"rows\"><table>\n<thead><tr>")
	for _, c := range columns {
		b.WriteString("<th>" + htmlText(c) + "</th>")
	}
	b.WriteString("</tr></thead>\n<tbody>\n")
	for _, row := range rows {
		b.WriteString("<tr>")
		for i := range columns {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			b.WriteString("<td>" + htmlText(cell) + "</td>")
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</tbody>\n</table></div>\n")
}

// htmlFix writes the fix a finding carries, with the sentence the Markdown
// writer uses, then each edit as removed and added lines and each run of the fix.
func htmlFix(b *strings.Builder, fix *review.Fix) {
	if fix == nil {
		return
	}
	b.WriteString("<h4>Fix</h4>\n<p>" + strings.TrimSpace(fixSentence(htmlStyle, fix)) + "</p>\n")
	if fix.Dependency != nil {
		return
	}
	if fix.Proved {
		for _, edit := range fix.Edits {
			b.WriteString("<p>Edit " + htmlCode(edit.Path) + ":</p>\n<pre>" + htmlPre(mdDiff(edit)) + "</pre>\n")
		}
	}
	if len(fix.Runs) > 0 {
		b.WriteString("<ul>\n")
		for _, run := range fix.Runs {
			fmt.Fprintf(b, "<li>%s: %s, exit %d, %s</li>\n", htmlText(run.Label), htmlCode(run.Command), run.ExitCode, htmlText(run.Result))
		}
		b.WriteString("</ul>\n")
	}
}

// htmlOpinions lists the claims two models made at the same place that no test,
// tool or data proved, one line each, and says why a claim one model makes is
// not here. The result holds no list of dropped claims, so none is printed.
func htmlOpinions(b *strings.Builder, opinions []review.Opinion) {
	b.WriteString("<section class=\"opinions\">\n<h2>Agreed by two models, not proven</h2>\n")
	if len(opinions) == 0 {
		b.WriteString("<p>No claim was agreed by two models without proof.</p>\n")
	} else {
		b.WriteString("<p>Two models raised each of these at the same place, and nothing proved it. They do not count as findings.</p>\n<ul>\n")
		for _, o := range opinions {
			b.WriteString("<li>" + htmlText(o.Title) + " — " + htmlCode(mdPlace(o.Location)) + "</li>\n")
		}
		b.WriteString("</ul>\n")
	}
	b.WriteString("<p>A claim only one model makes is not kept.</p>\n</section>\n")
}

// htmlNotes is the reviewer, author and failed-service lines, and the other
// statements a reader must not miss, as the Markdown notes say them.
func htmlNotes(b *strings.Builder, result review.Result) {
	lines := notesIn(htmlStyle, result)
	if len(lines) == 0 {
		return
	}
	b.WriteString("<ul class=\"notes\">\n")
	for _, line := range lines {
		b.WriteString("<li>" + strings.TrimPrefix(line, "- ") + "</li>\n")
	}
	b.WriteString("</ul>\n")
}

// The escaping rule for HTML output, written once (ADR-0027). A value that came
// from the change, a model or a service is never written raw; it goes through
// exactly one of:
//
//   - htmlText: prose, table cells and headings, by the rule written above htmlText
//     in format_md.go: whitespace folds, control characters show as escapes, and
//     '<', '>', '&' and both quotes become entities. A zero-width space follows the
//     ':' of "://", so no address stays an address in the file.
//   - htmlCode: a <code> element around htmlText, for a path, name or command.
//   - htmlPre: the same entities for a <pre> body, which keeps its line breaks.
//
// No value is ever put in an attribute: the only attributes the page has are
// fixed words and the "#finding-N" anchors Proven numbers itself.
// A new writer path that prints a change value without one of these is a defect.

// htmlCode writes s as a <code> element, empty for an empty value.
func htmlCode(s string) string {
	text := htmlText(s)
	if text == "" {
		return ""
	}
	return "<code>" + text + "</code>"
}

// htmlPre escapes a multi-line value for the inside of <pre>: each line is made
// visible and escaped as htmlText does, and the line breaks stay.
func htmlPre(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = htmlRunes([]rune(visible(line)))
	}
	return strings.Join(lines, "\n")
}

// htmlCSS is the page's one stylesheet. It names no font file and no address.
const htmlCSS = `body { margin: 0; font: 16px/1.5 system-ui, sans-serif; color: #1b1f24; background: #f6f8fa; }
main { max-width: 60rem; margin: 0 auto; padding: 1.5rem 1rem 3rem; }
h1 { font-size: 1.6rem; margin: 0 0 1rem; }
h2 { font-size: 1.2rem; margin: 0 0 .5rem; }
h3 { font-size: 1rem; margin: 1rem 0 .25rem; }
h4 { font-size: .95rem; margin: 1rem 0 .25rem; }
code, pre { font: .9rem/1.4 ui-monospace, Menlo, Consolas, monospace; }
pre { background: #eaeef2; padding: .75rem; overflow-x: auto; white-space: pre-wrap; overflow-wrap: anywhere; border-radius: 4px; }
.verdict { font-size: 1.1rem; padding: .75rem 1rem; border-left: 6px solid #6e7781; background: #fff; border-radius: 4px; }
.verdict.blocked { border-color: #cf222e; }
.verdict.clean { border-color: #1a7f37; }
.verdict.not-reviewed { border-color: #9a6700; }
table { border-collapse: collapse; background: #fff; }
th, td { border: 1px solid #d0d7de; padding: .3rem .6rem; text-align: left; vertical-align: top; }
.summary { width: 100%; margin: 1rem 0; }
.summary td code { overflow-wrap: anywhere; word-break: break-all; }
.rows { overflow-x: auto; }
.card { background: #fff; border: 1px solid #d0d7de; border-left: 6px solid #0969da; border-radius: 4px; padding: 1rem; margin: 1rem 0; }
.card.error { border-left-color: #cf222e; }
.card.warning { border-left-color: #bf8700; }
.severity { font-size: .8rem; text-transform: uppercase; letter-spacing: .04em; margin-right: .4rem; }
.place { margin: 0 0 .5rem; }
summary { cursor: pointer; font-weight: 600; }
.opinions { margin: 1.5rem 0; }
.notes { color: #424a53; }
footer { margin-top: 2rem; font-size: .85rem; color: #57606a; }
@media (prefers-color-scheme: dark) {
body { color: #e6edf3; background: #0d1117; }
.verdict, table, .card { background: #161b22; border-color: #30363d; }
th, td { border-color: #30363d; }
pre { background: #21262d; }
.notes, footer { color: #8b949e; }
}
`
