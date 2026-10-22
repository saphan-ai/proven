package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"

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

type textFormatter struct{}

func (textFormatter) Write(w io.Writer, result review.Result) error {
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
	if result.TracePath != "" {
		if _, err := fmt.Fprintf(w, "trace: %s\n", result.TracePath); err != nil {
			return err
		}
	}
	return nil
}

func writeFinding(w io.Writer, finding review.Finding) error {
	name := finding.ID
	if name == "" {
		name = finding.RuleID
	}
	if _, err := fmt.Fprintf(w, "finding %s %s %s\n", name, finding.Severity, finding.Title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s:%d-%d\n", finding.Location.Path, finding.Location.StartLine, finding.Location.EndLine); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s\n", finding.Message); err != nil {
		return err
	}
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
	}
	return nil
}

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
