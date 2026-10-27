// This file writes one review result as a Markdown pull-request comment
// (ADR-0020 §1, ADR-0027). Where each part sits is the contract in
// docs/OUTPUT_FORMATS.md.

package cli

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/review"
)

// mdVerdict is the first line: blocked, clean, or not reviewed when the model
// did not answer (mdModelsDown), with the counts a reader
// decides by. "blocked" is any finding that counts (Result.Findings holds only
// those: an accepted one is not in it). The result carries no count of dropped
// claims, so none is printed; a claim only one model makes is not in the result.
func mdVerdict(result review.Result) string {
	if len(result.Findings) == 0 {
		if down := mdModelsDown(result); len(down) > 0 {
			// Some calls answered and later ones did not: the review is incomplete, not
			// absent. Result.Models holds every model call that answered (ADR-0019).
			if len(result.Models) > 0 {
				return "**Proven: incomplete** — " + strings.Join(down, " and ") + " did not answer every call, so 0 findings proves nothing" + mdCountsTail(result)
			}
			return "**Proven: not reviewed** — " + strings.Join(down, " and ") + " did not answer, so 0 findings proves nothing" + mdCountsTail(result)
		}
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

// mdChecked is the clean run's statement of what was checked: the review class
// and the rule that chose it (how deep the review went), the paths a live review
// did not copy, and whether every outside service answered, by name when one did
// not. The Result carries no list of the files or the commit range reviewed, so
// none is printed; the foot names the models.
func mdChecked(result review.Result) string {
	class := mdText(result.Class)
	if class == "" {
		class = "an unclassified"
	} else {
		class = "a " + class
	}
	var parts []string
	if n := len(result.SkippedPaths); n > 0 {
		parts = append(parts, fmt.Sprintf("%d tracked %s skipped because %s symlinks",
			n, pluralWord(n, "path was", "paths were"), pluralWord(n, "it is a", "they are")))
	}
	if len(result.FailedServices) == 0 {
		parts = append(parts, "every service answered")
	} else {
		names := make([]string, 0, len(result.FailedServices))
		for _, failed := range result.FailedServices {
			names = append(names, mdServiceName(failed.Service))
		}
		parts = append(parts, "these services did not answer: "+strings.Join(names, ", "))
	}
	closing := "A finding needs proof to be reported, and none had it."
	if len(mdModelsDown(result)) > 0 {
		closing = "No model read the change, so no finding could be raised; run the review again when the model answers."
		if len(result.Models) > 0 {
			closing = "The model answered some calls and not others, so the review stopped before it could raise a finding; run it again when the model answers."
		}
	}
	return fmt.Sprintf("Checked as %s change%s; %s. %s",
		class, mdClassReason(result.ClassRule), strings.Join(parts, "; "), closing)
}

// mdClassReason says in words why the review chose its class: a path rule names
// its pattern, the fallback says no rule matched, and an empty rule says nothing.
// The owner's rule of 27 Oct: no raw field name or value on a page a judge reads.
func mdClassReason(rule string) string {
	switch {
	case rule == "":
		return ""
	case rule == "fallback":
		return " (no path rule matched)"
	case strings.HasPrefix(rule, "scout-intent"):
		return " (raised by the plan Proven made for this change)"
	default:
		return " (path rule " + mdCode(mdShort(rule, mdMaxTitle)) + ")"
	}
}

// mdServiceNames are the outside services a review calls, as a reader knows them.
var mdServiceNames = map[string]string{
	"nightshift": "the model reviewer",
	"nvidia":     "the NVIDIA model",
	"zetaris":    "the Zetaris dependency database",
	"guardrail":  "the guardrail",
	"prover":     "the Prover, which runs tests",
	"query":      "the evidence store",
	"local":      "the local evidence file",
}

// mdServiceName is a failed service as a reader knows it; an unknown one keeps its name.
func mdServiceName(service string) string {
	if name, ok := mdServiceNames[service]; ok {
		return name
	}
	return mdCode(mdShort(service, mdMaxTitle))
}

// mdUnsetSetting is a reason that only names an empty setting, such as "ZETARIS_REST_URL is empty".
var mdUnsetSetting = regexp.MustCompile(`^[A-Z][A-Z0-9_]* is empty$`)

// mdServiceReason is why a service did not answer, in words: a setting left empty
// says the service is not configured for this run instead of naming the variable.
func mdServiceReason(reason string) string {
	if mdUnsetSetting.MatchString(reason) {
		return "not configured for this run"
	}
	return mdText(mdShort(reason, mdMaxNote))
}

// mdSeverity is a severity as a word with a capital: Error, Warning, Info.
func mdSeverity(sev review.Severity) string {
	word := strings.ToLower(string(sev))
	if word == "" {
		return ""
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

// mdModelsDown names the model services that did not answer: the generator in
// either mode. With the model down no candidate is ever raised, so a result with
// no findings is not a clean review, and the comment must not call it one.
func mdModelsDown(result review.Result) []string {
	var down []string
	for _, failed := range result.FailedServices {
		if failed.Service == generator.ModeNightshift || failed.Service == generator.ModeNVIDIA {
			down = append(down, mdServiceName(failed.Service))
		}
	}
	return down
}

// mdTable is the findings table: severity, finding, file:line. The rule id is a
// name for machines; it stays in JSON and SARIF.
func mdTable(b *strings.Builder, findings []review.Finding) {
	b.WriteString("| Severity | Finding | Location |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, f := range findings {
		title := mdText(mdShort(f.Title, mdMaxTitle))
		if f.Unresolved {
			title += " (unresolved)"
		}
		fmt.Fprintf(b, "| %s | %s | %s |\n",
			mdSeverity(f.Severity), title, mdLocation(f.Location, true))
	}
	b.WriteString("\n")
}

// mdPlace is file:line as plain text, for a heading.
func mdPlace(l review.Location) string {
	switch {
	case l.StartLine <= 0:
		return l.Path
	case l.EndLine > l.StartLine:
		return fmt.Sprintf("%s:%d-%d", l.Path, l.StartLine, l.EndLine)
	default:
		return fmt.Sprintf("%s:%d", l.Path, l.StartLine)
	}
}

// mdShort bounds a raw value at max bytes and marks the cut with an ellipsis.
func mdShort(s string, max int) string {
	kept, cut := cutRaw(s, max)
	if cut > 0 {
		return kept + "…"
	}
	return kept
}

// mdLocation is file:line as one code span; a range prints both lines.
func mdLocation(l review.Location, cell bool) string {
	switch {
	case l.StartLine <= 0:
		return mdCodeIn(l.Path, cell)
	case l.EndLine > l.StartLine:
		return mdCodeIn(fmt.Sprintf("%s:%d-%d", l.Path, l.StartLine, l.EndLine), cell)
	default:
		return mdCodeIn(fmt.Sprintf("%s:%d", l.Path, l.StartLine), cell)
	}
}

// mdBlock renders one finding's <details> block at one proof level: the message,
// the evidence in the order the finding carries it, and the fix when the result has
// one. What the level left out is announced in the block, with the count.
func mdBlock(f review.Finding, lim mdLimits) string {
	var b strings.Builder
	// The summary line is raw HTML until the blank line after it, so it takes
	// HTML escaping only: Markdown inside it would print its own punctuation.
	// It names the finding by its title and place; the internal id stays in JSON.
	fmt.Fprintf(&b, "<details><summary>%s — %s — %s</summary>\n\n", htmlText(mdSeverity(f.Severity)), htmlText(mdShort(f.Title, mdMaxTitle)), htmlText(mdShort(mdPlace(f.Location), mdMaxTitle)))
	if f.Message != "" {
		message, cut := cutRaw(f.Message, mdMaxMessage)
		lim.cutBytes += cut
		b.WriteString(mdText(message))
		b.WriteString("\n\n")
	}
	for _, e := range f.Evidence {
		mdEvidence(&b, e, &lim)
	}
	mdFix(&b, f.Fix, &lim)
	var least []string
	if lim.cutEntries > 0 {
		least = append(least, fmt.Sprintf("%d more %s of evidence", lim.cutEntries, pluralWord(lim.cutEntries, "piece", "pieces")))
	}
	if lim.cutColumns > 0 {
		least = append(least, fmt.Sprintf("%d more query %s", lim.cutColumns, pluralWord(lim.cutColumns, "column", "columns")))
	}
	if lim.cutFix {
		least = append(least, "the fix")
	}
	if len(least) > 0 {
		fmt.Fprintf(&b, "_cut: %s not shown; the full evidence is in the JSON and SARIF files of this run._\n\n", strings.Join(least, ", "))
	}
	if lim.cutRows > 0 || lim.cutBytes > 0 {
		fmt.Fprintf(&b, "_cut: %d more %s and %d %s not shown; the full evidence is in the JSON and SARIF files of this run._\n\n",
			lim.cutRows, pluralWord(lim.cutRows, "row", "rows"), lim.cutBytes, pluralWord(lim.cutBytes, "byte", "bytes"))
	}
	b.WriteString("</details>\n\n")
	return b.String()
}

// mdEvidence writes one piece of proof: a query with its rows, a failing test, or a tool's run.
func mdEvidence(b *strings.Builder, e review.Evidence, lim *mdLimits) {
	switch e.Kind {
	case review.EvidenceQuery:
		q := e.Query
		if q == nil {
			return
		}
		fmt.Fprintf(b, "**Query** (%s, snapshot %s, form %s)\n\n", mdText(q.Engine), mdText(q.Snapshot), mdText(q.Form))
		mdFence(b, "sql", mdCut(q.SQL, lim))
		mdRows(b, q.Columns, q.Rows, lim)
	case review.EvidenceTest:
		t := e.Test
		if t == nil {
			return
		}
		fmt.Fprintf(b, "**Failing test** %s in %s — fails on the change: %s, passes on the base: %s\n\n",
			mdCode(t.Name), mdCode(t.Path), yesNo(t.FailsOnChange), yesNo(t.PassesOnBase))
		b.WriteString("Command:\n\n")
		mdFence(b, "sh", mdCut(t.Command, lim))
		b.WriteString("Output:\n\n")
		mdFence(b, "text", mdCut(strings.TrimSpace(t.Output), lim))
	case review.EvidenceTool:
		t := e.Tool
		if t == nil {
			return
		}
		fmt.Fprintf(b, "**Tool** %s %s — present on the change: %s, present on the base: %s\n\n",
			mdCode(t.Tool), mdCode(t.Version), yesNo(t.PresentOnChange), yesNo(t.PresentOnBase))
		mdFence(b, "sh", mdCut(t.Command, lim))
		b.WriteString("On the change:\n\n")
		mdFence(b, "text", mdCut(strings.TrimSpace(t.Output), lim))
		b.WriteString("On the base:\n\n")
		mdFence(b, "text", mdCut(strings.TrimSpace(t.BaseOutput), lim))
	}
}

// mdRows writes a query's rows as a table. Past lim.rows the table stops; the
// rows it left out are counted in lim and announced by the finding's block.
func mdRows(b *strings.Builder, columns []string, rows [][]string, lim *mdLimits) {
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
	if len(shown) > lim.rows {
		shown = shown[:lim.rows]
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
	lim.cutRows += len(rows) - len(shown)
}

// mdFix writes the fix a finding carries. Like the text writer, it shows a patch
// only when it was proved, and says each fix is a suggestion a person applies.
func mdFix(b *strings.Builder, fix *review.Fix, lim *mdLimits) {
	if fix == nil {
		return
	}
	switch {
	case fix.Dependency != nil:
		d := fix.Dependency
		fmt.Fprintf(b, "**Fix** (named from the advisory table, no test ran): raise %s %s to %s. %s\n\n",
			mdCode(d.Package), mdCode(d.From), mdCode(d.To), mdText(mdShort(oneLine(fix.Why), mdMaxMessage)))
	case fix.Proved:
		fmt.Fprintf(b, "**Fix** (proved after %d %s; a suggestion for you to apply, Proven changes nothing): %s %s\n\n",
			fix.Attempts, attemptWord(fix.Attempts), mdText(mdShort(oneLine(fix.What), mdMaxMessage)), mdText(mdShort(oneLine(fix.Why), mdMaxMessage)))
		for _, edit := range fix.Edits {
			fmt.Fprintf(b, "Edit %s:\n\n", mdCode(edit.Path))
			mdFence(b, "diff", mdCut(mdDiff(edit), lim))
		}
		mdFixRuns(b, fix.Runs, lim)
	default:
		b.WriteString("**Fix** not proved: " + mdText(mdShort(oneLine(fix.Note), mdMaxMessage)) + "\n\n")
		mdFixRuns(b, fix.Runs, lim)
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

func mdFixRuns(b *strings.Builder, runs []review.FixRun, lim *mdLimits) {
	if len(runs) > lim.rows {
		lim.cutRows += len(runs) - lim.rows
		runs = runs[:lim.rows]
	}
	for _, run := range runs {
		fmt.Fprintf(b, "- %s: %s, exit %d, %s\n", mdText(mdShort(run.Label, mdMaxTitle)), mdCode(mdShort(run.Command, mdMaxNote)), run.ExitCode, mdText(mdShort(run.Result, mdMaxNote)))
	}
	if len(runs) > 0 {
		b.WriteString("\n")
	}
}

// mdNotes prints what the verdict alone does not say, as the text writer does:
// the policy, the settings and author-model statements, the reviewers, every
// failed service (clean or not), a partial review, a guardrail that was off,
// accepted findings and exceptions that expired. Each line is bounded, and a
// long list of exceptions is cut with its count.
// mdPeople are the author and reviewer lines, read from the structured fields so
// a reader sees words, not the review's internal role names (ADR-0037 point 5).
func mdPeople(result review.Result) []string {
	var lines []string
	if len(result.AuthorModels) == 0 {
		lines = append(lines, "- Author model: unknown, the change does not declare it")
	} else {
		names := make([]string, 0, len(result.AuthorModels))
		for _, a := range result.AuthorModels {
			names = append(names, fmt.Sprintf("%s (%s)", mdCode(mdShort(a.Model, mdMaxTitle)), mdText(mdFamily(a.Family))))
		}
		lines = append(lines, "- Author model: "+strings.Join(names, ", "))
	}
	for _, r := range result.Reviewers {
		role := "Reviewer"
		if r.Role == review.RoleSpecial {
			role = "Second reviewer, for " + mdText(result.Class) + " changes"
		}
		line := fmt.Sprintf("- %s: %s (%s)", role, mdCode(mdShort(r.Model, mdMaxTitle)), mdText(mdFamily(r.Family)))
		if r.IsAuthorModel == "yes" {
			line += ", the same family as the author"
		}
		lines = append(lines, line)
	}
	if result.SpecialReviewer == "" && result.Doubled {
		lines = append(lines, "- Second reviewer: not configured")
	}
	return lines
}

// mdFamily is a model family without the word "family".
func mdFamily(f string) string {
	return strings.TrimPrefix(f, "family ")
}

func mdNotes(b *strings.Builder, result review.Result) {
	var lines []string
	statement := func(s string) {
		if s != "" {
			lines = append(lines, "- "+mdText(mdShort(oneLine(s), mdMaxNote)))
		}
	}
	statement(result.PolicyStatement)
	statement(result.SettingsStatement)
	if len(result.Reviewers) > 0 {
		lines = append(lines, mdPeople(result)...)
	} else {
		statement(result.AuthorModelStatement)
		for _, line := range result.ReviewerStatements {
			statement(line)
		}
	}
	for _, failed := range result.FailedServices {
		lines = append(lines, fmt.Sprintf("- Did not answer: %s (%s)", mdServiceName(failed.Service), mdServiceReason(failed.Reason)))
	}
	if result.Partial {
		lines = append(lines, "- review partial: the deadline was reached")
	}
	for i, a := range result.Accepted {
		if i == mdMaxNotes {
			lines = append(lines, fmt.Sprintf("- %d more accepted by exception, not listed", len(result.Accepted)-i))
			break
		}
		lines = append(lines, fmt.Sprintf("- accepted by exception %s (owner %s, expires %s): %s %s",
			mdCode(a.ExceptionID), mdText(mdShort(a.Owner, mdMaxTitle)), mdText(mdShort(a.Expires, mdMaxTitle)),
			mdText(mdShort(a.Finding.Title, mdMaxTitle)), mdLocation(a.Finding.Location, false)))
	}
	for i, e := range result.Expired {
		if i == mdMaxNotes {
			lines = append(lines, fmt.Sprintf("- %d more exceptions expired, not listed", len(result.Expired)-i))
			break
		}
		lines = append(lines, fmt.Sprintf("- exception %s expired on %s: finding %s counts again",
			mdCode(e.ExceptionID), mdText(mdShort(e.Expires, mdMaxTitle)), mdText(mdShort(e.Finding, mdMaxTitle))))
	}
	if len(lines) == 0 {
		return
	}
	b.WriteString(strings.Join(lines, "\n"))
	b.WriteString("\n\n")
}

// mdFoot is the one line at the bottom: the reviewing models and the guardrail
// line, the same words the text output prints (review.Result.GuardrailLine),
// built from the checks the pipeline counted.
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
			escaped = append(escaped, mdCode(mdShort(m, mdMaxTitle)))
		}
		names = "reviewed by " + strings.Join(escaped, " and ")
	}
	return "<sub>" + names + " · " + mdText(result.GuardrailLine()) + "</sub>"
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

// mdCut bounds an output at lim.out bytes on a rune boundary, says so at the
// cut, and counts the bytes in lim for the finding's block to announce.
func mdCut(s string, lim *mdLimits) string {
	kept, cut := cutRaw(s, lim.out)
	if cut == 0 {
		return s
	}
	lim.cutBytes += cut
	return kept + fmt.Sprintf("\n… %d bytes not shown", cut)
}

// The escaping rule for Markdown output, written once (ADR-0027). A value that
// came from the change is never written raw; it goes through exactly one of:
//
//   - mdText: prose and table cells. Whitespace runs fold to one space, so a
//     value cannot start a new line, row or block; control and bidirectional
//     characters are shown as escapes (visible); backslash, backtick, '*',
//     '[', ']', '#', '|', '~' and '@' are backslash-escaped, which stops
//     emphasis, links, images, issue references, mentions, strikethrough and
//     table pipes (a zero-width space follows each '@' and '#', because the
//     renderer reads the escape away and then links mentions and references);
//     '_' is escaped unless letters or digits stand on both
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
	// WHY a zero-width space after the escaped '@' and '#': the Markdown parser
	// consumes the backslash, and GitHub then links a mention or an issue reference
	// in the rendered text, so "\@user" would still notify. Made after the entities
	// because the entity's own '&' and '#' must not be escaped.
	return mentionBreak.Replace(htmlEntities(out))
}

// mentionBreak puts a zero-width space after each escaped '@' and '#' of mdText.
// Every '@' and '#' of the text is escaped before this runs; the '#' of an entity
// (&#34;) is not, so it is not matched.
var mentionBreak = strings.NewReplacer(`\@`, `\@`+zeroWidth, `\#`, `\#`+zeroWidth)

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
func mdCode(s string) string { return mdCodeIn(s, false) }

// mdCodeIn writes s as a code span. In a table cell ('|' splits a row before the
// span is read) a pipe is written '\|'; anywhere else that backslash would print.
func mdCodeIn(s string, cell bool) string {
	s = visible(strings.Join(strings.Fields(s), " "))
	if s == "" {
		return ""
	}
	if cell {
		s = strings.ReplaceAll(s, "|", `\|`)
	}
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
