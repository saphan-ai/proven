package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/review"
)

// Formatter writes one review result in one format.
// Text is implemented. JSON and SARIF use this same seam and refuse until they are built.
type Formatter interface {
	Write(w io.Writer, result review.Result) error
}

// FormatterFor returns the writer for format, one of text, json or sarif.
func FormatterFor(format string) (Formatter, error) {
	switch format {
	case "text":
		return textFormatter{}, nil
	case "json", "sarif":
		return nil, fmt.Errorf("format %q has no writer in this build", format)
	default:
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
	if _, err := fmt.Fprintf(w, "review class: %s\n", result.Class); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "review rule: %s\n", result.ClassRule); err != nil {
		return err
	}
	if result.SettingsStatement != "" {
		if _, err := fmt.Fprintln(w, result.SettingsStatement); err != nil {
			return err
		}
	}
	if result.AuthorModelStatement != "" {
		if _, err := fmt.Fprintln(w, result.AuthorModelStatement); err != nil {
			return err
		}
	}
	for _, line := range result.ReviewerStatements {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	for _, finding := range result.Findings {
		if err := writeFinding(w, finding); err != nil {
			return err
		}
	}
	if len(result.FailedServices) == 0 {
		if _, err := fmt.Fprintln(w, "failed services: none"); err != nil {
			return err
		}
	} else {
		for _, failed := range result.FailedServices {
			if _, err := fmt.Fprintf(w, "failed service: %s: %s\n", failed.Service, failed.Reason); err != nil {
				return err
			}
		}
	}
	if result.Partial {
		if _, err := fmt.Fprintln(w, "review partial: the deadline was reached"); err != nil {
			return err
		}
	}
	if sentence := familySentence(result.TracePath); sentence != "" {
		if _, err := fmt.Fprintln(w, sentence); err != nil {
			return err
		}
	}
	if result.TracePath != "" {
		if _, err := fmt.Fprintf(w, "trace: %s\n", result.TracePath); err != nil {
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

// writeFinding prints one finding: its name, severity and title, its EPSS
// lines, an "unresolved" mark when its evidence did not settle it, where it
// is, why it matters, its rule, and the evidence behind it.
func writeFinding(w io.Writer, finding review.Finding) error {
	name := finding.ID
	if name == "" {
		name = finding.RuleID
	}
	if _, err := fmt.Fprintf(w, "finding %s %s %s\n", name, finding.Severity, finding.Title); err != nil {
		return err
	}
	for _, e := range finding.Evidence {
		if e.Kind != review.EvidenceQuery || e.Query == nil {
			continue
		}
		for _, line := range epss.Phrases(e.Query.Columns, e.Query.Rows) {
			if _, err := fmt.Fprintf(w, "  %s\n", line); err != nil {
				return err
			}
		}
	}
	if finding.Unresolved {
		if _, err := fmt.Fprintln(w, "  unresolved"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "  %s:%d-%d\n", finding.Location.Path, finding.Location.StartLine, finding.Location.EndLine); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s\n", finding.Message); err != nil {
		return err
	}
	// The rule line is printed only when it adds something the name line
	// does not already say.
	if finding.RuleID != "" && finding.RuleID != name {
		if _, err := fmt.Fprintf(w, "  rule %s\n", finding.RuleID); err != nil {
			return err
		}
	}
	for _, evidence := range finding.Evidence {
		if err := writeEvidence(w, evidence); err != nil {
			return err
		}
	}
	return nil
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
		if _, err := fmt.Fprintf(w, "  evidence test %s %s fails_on_change=%t passes_on_base=%t\n", test.Path, test.Name, test.FailsOnChange, test.PassesOnBase); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  command: %s\n", test.Command); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  output: %s\n", strings.TrimSpace(test.Output)); err != nil {
			return err
		}
	case review.EvidenceQuery:
		if evidence.Query == nil {
			return nil
		}
		query := evidence.Query
		if _, err := fmt.Fprintf(w, "  evidence query %s snapshot=%s\n", query.Engine, query.Snapshot); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  SQL: %s\n", query.SQL); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  columns: %s\n", strings.Join(query.Columns, ", ")); err != nil {
			return err
		}
		for _, row := range query.Rows {
			if _, err := fmt.Fprintf(w, "  row: %s\n", strings.Join(row, " | ")); err != nil {
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
