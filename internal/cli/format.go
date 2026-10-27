package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/review"
)

// Formatter writes one review result in one format.
// text, JSON, SARIF and Markdown each have a writer in this build (ADR-0020 §1).
type Formatter interface {
	Write(w io.Writer, result review.Result) error
}

// FormatterFor returns the writer for format, one of text, json, sarif or md.
// ADR-0020 §1: the same four names are the standard-output format and the --*-output files.
func FormatterFor(format string) (Formatter, error) {
	switch format {
	case "text":
		return textFormatter{}, nil
	case "json":
		return jsonFormatter{}, nil
	case "sarif":
		return sarifFormatter{}, nil
	case "md":
		return markdownFormatter{}, nil
	default:
		// A name the flag parser did not offer still has to refuse, or a caller
		// inside this package could write an empty file under an unknown format.
		return nil, fmt.Errorf("format %q has no writer in this build", format)
	}
}

// render writes result in format to memory, so a failure is known before
// any output is written.
func render(format string, result review.Result) ([]byte, error) {
	writer, err := FormatterFor(format)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := writer.Write(&buf, result); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// textFormatter is the human-readable output: the review class, then each
// finding with the evidence that proves it, then the services that failed.
type textFormatter struct{}

// Write prints the review in a fixed order a reader can scan top to bottom.
// "failed services: none" is printed rather than left out, so a reader can
// tell a review that reached everything from one whose failures went unsaid.
func (textFormatter) Write(w io.Writer, result review.Result) error {
	if _, err := fmt.Fprintf(w, "review class: %s\n", visible(result.Class)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "review rule: %s\n", visible(result.ClassRule)); err != nil {
		return err
	}
	if result.PolicyStatement != "" {
		if _, err := fmt.Fprintln(w, visible(result.PolicyStatement)); err != nil {
			return err
		}
	}
	if result.SettingsStatement != "" {
		if _, err := fmt.Fprintln(w, visible(result.SettingsStatement)); err != nil {
			return err
		}
	}
	if result.AuthorModelStatement != "" {
		if _, err := fmt.Fprintln(w, visible(result.AuthorModelStatement)); err != nil {
			return err
		}
	}
	for _, line := range result.ReviewerStatements {
		if _, err := fmt.Fprintln(w, visible(line)); err != nil {
			return err
		}
	}
	for _, finding := range result.Findings {
		if err := writeFinding(w, finding); err != nil {
			return err
		}
	}
	// ADR-0041: the opinion section is its own block, after the findings, so a
	// line that starts with "finding " is never an opinion.
	if err := writeOpinions(w, result); err != nil {
		return err
	}
	for _, accepted := range result.Accepted {
		if err := writeAccepted(w, accepted); err != nil {
			return err
		}
	}
	for _, expired := range result.Expired {
		if _, err := fmt.Fprintf(w, "exception %s expired on %s: finding %s counts again (rule %s at %s, owner %s, reason: %s)\n",
			visible(expired.ExceptionID), visible(expired.Expires), visible(expired.Finding), visible(expired.Rule), visible(expired.Location), visible(expired.Owner), visible(expired.Reason)); err != nil {
			return err
		}
	}
	if len(result.FailedServices) == 0 {
		if _, err := fmt.Fprintln(w, "failed services: none"); err != nil {
			return err
		}
	} else {
		for _, failed := range result.FailedServices {
			if _, err := fmt.Fprintf(w, "failed service: %s: %s\n", visible(failed.Service), visible(failed.Reason)); err != nil {
				return err
			}
		}
	}
	if result.Partial {
		if _, err := fmt.Fprintln(w, "review partial: the deadline was reached"); err != nil {
			return err
		}
	}
	// The guardrail line is in the footer of every output, off or on, so a
	// reader of the end of the review sees whether NVIDIA's guardrail checked it.
	if _, err := fmt.Fprintln(w, visible(result.GuardrailLine())); err != nil {
		return err
	}
	if sentence := familySentence(result.TracePath); sentence != "" {
		if _, err := fmt.Fprintln(w, visible(sentence)); err != nil {
			return err
		}
	}
	if result.TracePath != "" {
		if _, err := fmt.Fprintf(w, "trace: %s\n", visible(result.TracePath)); err != nil {
			return err
		}
	}
	return nil
}

// familySentence is the review's same-family line, when the trace has one.
// It is read back from the trace rather than passed in the result, so the
// output says exactly what the run recorded; an unreadable trace prints no
// line instead of failing the review.
func familySentence(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Kind == "family" && e.Text != "" {
			return e.Text
		}
	}
	return ""
}

// writeOpinions prints the section "Agreed by two models, not proven" (ADR-0041).
// The section is always there. With one reviewer it says why it is empty: a
// claim only one model makes stays dropped. Each opinion prints both reviewers'
// reasons as they wrote them. The lines are indented, so they are not findings.
func writeOpinions(w io.Writer, result review.Result) error {
	if _, err := fmt.Fprintln(w, "Agreed by two models, not proven"); err != nil {
		return err
	}
	if len(result.Opinions) == 0 {
		if len(result.Reviewers) < 2 {
			_, err := fmt.Fprintln(w, "  none: one reviewer; a claim only one model makes stays dropped")
			return err
		}
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	for _, op := range result.Opinions {
		if _, err := fmt.Fprintf(w, "  opinion %s %s:%d-%d\n", visible(op.ID), visible(op.Location.Path), op.Location.StartLine, op.Location.EndLine); err != nil {
			return err
		}
		for _, reason := range op.Reasons {
			if _, err := fmt.Fprintf(w, "  %s %s: %s\n", visible(reason.Role), visible(reason.Model), visible(reason.Reason)); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeFinding prints one finding: its name, severity and title, its EPSS
// lines, an "unresolved" mark when its evidence did not settle it, where it
// is, why it matters, its rule, and the evidence behind it.
func writeFinding(w io.Writer, finding review.Finding) error {
	name := finding.ID
	if name == "" {
		name = finding.RuleID
	}
	if _, err := fmt.Fprintf(w, "finding %s %s %s\n", visible(name), visible(string(finding.Severity)), visible(finding.Title)); err != nil {
		return err
	}
	for _, e := range finding.Evidence {
		if e.Kind != review.EvidenceQuery || e.Query == nil {
			continue
		}
		for _, line := range epss.Phrases(e.Query.Columns, e.Query.Rows) {
			if _, err := fmt.Fprintf(w, "  %s\n", visible(line)); err != nil {
				return err
			}
		}
	}
	if finding.Unresolved {
		if _, err := fmt.Fprintln(w, "  unresolved"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "  %s:%d-%d\n", visible(finding.Location.Path), finding.Location.StartLine, finding.Location.EndLine); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s\n", visible(finding.Message)); err != nil {
		return err
	}
	// The rule line is printed only when it adds something the name line
	// does not already say.
	if finding.RuleID != "" && finding.RuleID != name {
		if _, err := fmt.Fprintf(w, "  rule %s\n", visible(finding.RuleID)); err != nil {
			return err
		}
	}
	for _, evidence := range finding.Evidence {
		if err := writeEvidence(w, evidence); err != nil {
			return err
		}
	}
	return writeFix(w, finding.Fix)
}

// writeFix prints the fix under a finding (ADR-0042). A proved patch shows its edits, its
// explanation and the commands of its proof; a patch that was not proved shows only that the
// attempt failed and why, never the patch. A dependency fix shows the version and the advisory row.
// Every fix says it is a suggestion: a person applies it, Proven writes nothing to the change.
func writeFix(w io.Writer, fix *review.Fix) error {
	if fix == nil {
		return nil
	}
	switch {
	case fix.Dependency != nil:
		return writeDependencyFix(w, fix)
	case fix.Proved:
		return writeProvedFix(w, fix)
	default:
		if _, err := fmt.Fprintf(w, "  %s\n", fix.Note); err != nil {
			return err
		}
		return writeFixRuns(w, fix.Runs)
	}
}

// writeProvedFix prints a patch that the same test proved, with its explanation (point 7).
func writeProvedFix(w io.Writer, fix *review.Fix) error {
	if _, err := fmt.Fprintf(w, "  fix proved after %d %s: a suggestion for you to apply; Proven changes nothing\n", fix.Attempts, attemptWord(fix.Attempts)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  fix changes: %s\n  fix closes the problem because: %s\n", oneLine(fix.What), oneLine(fix.Why)); err != nil {
		return err
	}
	for _, edit := range fix.Edits {
		if _, err := fmt.Fprintf(w, "  fix edit %s\n", edit.Path); err != nil {
			return err
		}
		for _, line := range strings.Split(strings.TrimRight(edit.Old, "\n"), "\n") {
			if _, err := fmt.Fprintf(w, "  - %s\n", line); err != nil {
				return err
			}
		}
		for _, line := range strings.Split(strings.TrimRight(edit.New, "\n"), "\n") {
			if _, err := fmt.Fprintf(w, "  + %s\n", line); err != nil {
				return err
			}
		}
	}
	return writeFixRuns(w, fix.Runs)
}

// writeFixRuns prints the validation steps: each run with its command, exit code and result.
func writeFixRuns(w io.Writer, runs []review.FixRun) error {
	for _, run := range runs {
		if _, err := fmt.Fprintf(w, "  fix run %s: %s exit=%d %s\n", run.Label, run.Command, run.ExitCode, run.Result); err != nil {
			return err
		}
	}
	return nil
}

// writeDependencyFix prints the smallest version that leaves the affected range, with the
// advisory row that shows it (point 5). It says no test ran, so it is not read as a proved fix.
func writeDependencyFix(w io.Writer, fix *review.Fix) error {
	dep := fix.Dependency
	if _, err := fmt.Fprintf(w, "  fix named from the advisory table, not run by a test: raise %s %s to %s\n", dep.Package, dep.From, dep.To); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  fix closes the problem because: %s\n", oneLine(fix.Why)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  fix advisory %s columns: %s\n  fix advisory row: %s\n", dep.AdvisoryID, strings.Join(dep.AdvisoryColumns, ", "), strings.Join(dep.AdvisoryRow, " | ")); err != nil {
		return err
	}
	return nil
}

// attemptWord is "attempt" for one and "attempts" for any other number.
func attemptWord(n int) string {
	if n == 1 {
		return "attempt"
	}
	return "attempts"
}

// oneLine folds the line breaks of a model's sentence, so it stays on one line of the review.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// writeAccepted prints a finding an exception of the organisation policy accepts (ADR-0036): where it
// is and which rule found it, then the exception's reason, owner and expiry. It is shown, never hidden,
// and it is not counted: it is not in result.Findings.
func writeAccepted(w io.Writer, accepted review.AcceptedFinding) error {
	finding := accepted.Finding
	name := finding.ID
	if name == "" {
		name = finding.RuleID
	}
	if _, err := fmt.Fprintf(w, "accepted finding %s %s %s\n", visible(name), visible(string(finding.Severity)), visible(finding.Title)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s:%d-%d\n", visible(finding.Location.Path), finding.Location.StartLine, finding.Location.EndLine); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  rule %s\n", visible(finding.RuleID)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "  accepted by exception %s: reason: %s; owner: %s; expires: %s\n",
		visible(accepted.ExceptionID), visible(accepted.Reason), visible(accepted.Owner), visible(accepted.Expires))
	return err
}

// writeEvidence prints the proof under a finding: the test the Prover wrote,
// with whether it fails on the change and passes on the base, the query
// that was run and the rows it returned, or a tool's run on both sides.
// A reader can rerun any of them.
func writeEvidence(w io.Writer, evidence review.Evidence) error {
	switch evidence.Kind {
	case review.EvidenceTest:
		if evidence.Test == nil {
			return nil
		}
		test := evidence.Test
		if _, err := fmt.Fprintf(w, "  evidence test %s %s fails_on_change=%t passes_on_base=%t\n", visible(test.Path), visible(test.Name), test.FailsOnChange, test.PassesOnBase); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  command: %s\n", visible(test.Command)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  output: %s\n", visibleLines(strings.TrimSpace(test.Output))); err != nil {
			return err
		}
	case review.EvidenceQuery:
		if evidence.Query == nil {
			return nil
		}
		query := evidence.Query
		// ADR-0047: the line names the form that ran. The SQL line below is
		// that form's text. Existing readers match the engine and the snapshot
		// as a prefix of this line.
		if _, err := fmt.Fprintf(w, "  evidence query %s snapshot=%s form=%s\n", visible(query.Engine), visible(query.Snapshot), visible(query.Form)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  SQL: %s\n", visible(query.SQL)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  columns: %s\n", visible(strings.Join(query.Columns, ", "))); err != nil {
			return err
		}
		for _, row := range query.Rows {
			if _, err := fmt.Fprintf(w, "  row: %s\n", visible(strings.Join(row, " | "))); err != nil {
				return err
			}
		}
	case review.EvidenceTool:
		if evidence.Tool == nil {
			return nil
		}
		// ADR-0040: the reader sees the tool, the command and both sides.
		// The bools say whether the result was new; the outputs are the difference.
		tool := evidence.Tool
		if _, err := fmt.Fprintf(w, "  evidence tool %s version=%s present_on_change=%t present_on_base=%t\n", tool.Tool, tool.Version, tool.PresentOnChange, tool.PresentOnBase); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  command: %s\n", tool.Command); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  change: %s\n", strings.TrimSpace(tool.Output)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  base: %s\n", strings.TrimSpace(tool.BaseOutput)); err != nil {
			return err
		}
	}
	return nil
}

// continuationPrefix starts every line after the first of a multi-line field.
// Every line Proven writes itself starts at column 0 or with exactly two
// spaces, and none starts with four spaces and a bar, so a line carrying this
// prefix is always read as the continuation of the field above it.
const continuationPrefix = "    | "

// visible returns s with every byte that could act on a terminal shown as a
// visible escape. The byte contract, which ADR-0027 asks to be written once:
//
//   - C0 (U+0000 to U+001F), tab and newline included, DEL (U+007F) and C1
//     (U+0080 to U+009F) are shown as \xNN of the code point, e.g. \x1b, \x09.
//   - A bidirectional control or an invisible mark (evidence.IsBidi) is shown
//     as a backslash, u and four hex digits of the code point.
//   - A byte that is not part of valid UTF-8 is shown as \xNN of the original
//     byte, so what the change held is what the reader sees, not U+FFFD.
//   - Everything else is written unchanged.
//
// Every text the writer prints that can come from the change goes through it,
// so the change cannot move the cursor, recolour or rewrite the reader's
// terminal. This is the text writer's contract. encoding/json escapes C0 and
// leaves DEL and C1 raw, so the JSON and SARIF writers escape that range
// after encoding (writeEncoded).
func visible(s string) string {
	return escapeControls(s, false)
}

// visibleLines is visible for a field that is several lines long, such as a
// test's output. Its newlines are kept, and each one is followed by
// continuationPrefix: otherwise a line inside the field, such as
// "failed services: none", would be indistinguishable from one Proven wrote.
func visibleLines(s string) string {
	return escapeControls(s, true)
}

func escapeControls(s string, keepNewline bool) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case r == '\n' && keepNewline:
			b.WriteString("\n" + continuationPrefix)
		case isControl(r):
			fmt.Fprintf(&b, `\x%02x`, r)
		case evidence.IsBidi(r):
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}

func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}

// writeEncoded writes v as one indented JSON document, then escapes DEL and C1.
// HTML escaping stays off: evidence and messages are text, and the default
// encoder would turn "<" into an escape a person would have to undo.
// The indent is two spaces, so a person can read the file.
func writeEncoded(w io.Writer, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := w.Write(escapeDELAndC1(buf.Bytes()))
	return err
}

// escapeDELAndC1 rewrites U+007F and U+0080 to U+009F as \u00XX.
// WHY after the encoder: encoding/json escapes C0 (U+0000 to U+001F) and writes
// this range as raw UTF-8. A terminal that prints the file treats C1 as a
// control sequence (U+009B starts one). The escape is still JSON and decodes
// to the same rune. A document with none of these runes is unchanged.
func escapeDELAndC1(b []byte) []byte {
	var out bytes.Buffer
	out.Grow(len(b))
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			fmt.Fprintf(&out, `\u%04x`, r)
			i += size
			continue
		}
		out.Write(b[i : i+size])
		i += size
	}
	return out.Bytes()
}

// filterFindings keeps the findings at or above min.
func filterFindings(in []review.Finding, min review.Severity) []review.Finding {
	if len(in) == 0 {
		return nil
	}
	floor := severityRank(min)
	out := make([]review.Finding, 0, len(in))
	for _, finding := range in {
		if severityRank(finding.Severity) >= floor {
			out = append(out, finding)
		}
	}
	return out
}

// severityRank orders the severities; an unknown one ranks with INFO, so
// it is never hidden by a floor.
func severityRank(s review.Severity) int {
	switch s {
	case review.SeverityWarning:
		return 1
	case review.SeverityError:
		return 2
	default:
		return 0
	}
}
