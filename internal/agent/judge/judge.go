// Package judge keeps a finding only when the evidence settles it.
//
// The Judge is the last step before the Writer. It reads each candidate's
// evidence and the Skeptic's refutation and decides with no model: a test that
// fails on the change and passes on the base, a query whose rows name the
// finding (ADR-0002), or a tool that reports a result on the change and not
// on the base (ADR-0040). A disagreement is settled by evidence, never by a vote
// (ADR-0006 point 3).
package judge

import (
	"strings"
	"unicode"

	"github.com/saphan-ai/proven/internal/candidate"
)

// Decide keeps a finding whose own evidence holds and whose refutation has none.
// It drops a finding whose evidence is contradicted by evidence.
// It keeps, and marks Unresolved, a finding whose evidence is present but an
// instrument the skeptic named could not run. It never votes and it never
// reads a confidence.
func Decide(c candidate.Candidate) candidate.Candidate {
	// A refutation backed by evidence wins over the finding's own evidence
	// (ADR-0006 point 4), so it is checked first.
	if why := refuteWhy(c); why != "" {
		c.Dropped = true
		c.Unresolved = false
		c.Refutation.Reason = why
		c.DropReason = "dropped: refuted: " + why
		return c
	}
	// Evidence the Skeptic could not test is kept but marked, so the review says it is open.
	if unsettled(c) {
		c.Dropped = false
		c.Unresolved = true
		why := c.Refutation.Reason
		if why == "" {
			why = "the evidence did not settle it"
		}
		c.DropReason = "unresolved: " + why
		return c
	}
	if counts(c) {
		c.Dropped = false
		c.Unresolved = false
		c.DropReason = "kept"
		return c
	}
	c.Dropped = true
	c.Unresolved = false
	// ADR-0002 point 2: a finding with no evidence is dropped, not softened.
	c.DropReason = "dropped: no evidence"
	return c
}

// Evidenced reports whether Decide would count this candidate's own evidence.
func Evidenced(c candidate.Candidate) bool {
	return counts(c)
}

// counts reports whether the candidate has evidence of its own: a test with
// both outcomes right, query rows Proven bound to the change, or a tool result
// the base did not already have (ADR-0040).
func counts(c candidate.Candidate) bool {
	for _, e := range c.Evidence {
		if e.Kind == candidate.KindTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase {
			return true
		}
		// Rows are evidence only when Proven tied them to the change's own
		// package, version or advisory. A word a cell shares with the finding's
		// text, which a model wrote, binds nothing.
		if e.Kind == candidate.KindQuery && e.Query != nil && e.Query.Bound && len(e.Query.Rows) > 0 {
			return true
		}
		// A result on both sides is the base, not the change, so it is not evidence.
		if e.Kind == candidate.KindTool && e.Tool != nil && e.Tool.PresentOnChange && !e.Tool.PresentOnBase {
			return true
		}
	}
	return false
}

// unsettled reports evidence whose refuting test or query could not run.
func unsettled(c candidate.Candidate) bool {
	if !counts(c) {
		return false
	}
	return c.Refutation.TestCannotRun || c.Refutation.QueryCannotRun
}

// Refutes reports whether the candidate's refutation carries evidence: a
// refuting test found and passing on the change, or query rows that contradict
// the finding. The Skeptic's own word is not evidence (ADR-0006 point 4).
func Refutes(c candidate.Candidate) bool {
	return refuteWhy(c) != ""
}

// refuteWhy is the reason a refutation holds, or empty when it does not.
// A refuting test holds when it was found and does not fail on the change.
func refuteWhy(c candidate.Candidate) string {
	r := c.Refutation
	if r.TestFound && !r.TestFailsOnChange {
		// Keep a reason that already says this, so the trace does not repeat itself.
		if strings.Contains(r.Reason, "passes on the change") {
			return r.Reason
		}
		name := r.TestName
		if name == "" {
			name = "the refuting test"
		}
		return "test " + name + " passes on the change"
	}
	if QueryContradicts(c, r.QueryRows) {
		return "query rows contradict the finding"
	}
	return ""
}

// QueryContradicts reports whether rows are about the finding and deny it.
// A row that does not mention the finding does not contradict it.
func QueryContradicts(c candidate.Candidate, rows [][]string) bool {
	if !queryRelates(c, rows) {
		return false
	}
	for _, row := range rows {
		for _, cell := range row {
			if denial(cell) {
				return true
			}
		}
	}
	return false
}

// queryRelates reports whether any cell names the finding: the whole cell, or
// one of its words, appears in the finding's id, rule, title, message or path.
// Short cells and words are skipped, so "1" or "id" matches nothing by chance.
func queryRelates(c candidate.Candidate, rows [][]string) bool {
	if len(rows) == 0 {
		return false
	}
	hay := strings.ToLower(c.ID + " " + c.RuleID + " " + c.Title + " " + c.Message + " " + c.Path)
	for _, row := range rows {
		for _, cell := range row {
			cell = strings.TrimSpace(cell)
			if len(cell) < 3 {
				continue
			}
			low := strings.ToLower(cell)
			if strings.Contains(hay, low) {
				return true
			}
			for _, tok := range tokens(low) {
				if len(tok) < 4 {
					continue
				}
				if strings.Contains(hay, tok) {
					return true
				}
			}
		}
	}
	return false
}

// denial reports a cell that says no.
func denial(cell string) bool {
	switch strings.ToLower(strings.TrimSpace(cell)) {
	case "false", "no", "absent", "clean", "pass", "passed":
		return true
	default:
		return false
	}
}

// tokens splits s into words of letters and digits.
func tokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}
