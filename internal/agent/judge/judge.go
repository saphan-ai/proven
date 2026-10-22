// Package judge keeps a finding only when the evidence settles it.
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
	if why := refuteWhy(c); why != "" {
		c.Dropped = true
		c.Unresolved = false
		c.Refutation.Reason = why
		c.DropReason = "dropped: refuted: " + why
		return c
	}
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
	c.DropReason = "dropped: no evidence"
	return c
}

func counts(c candidate.Candidate) bool {
	for _, e := range c.Evidence {
		if e.Kind == candidate.KindTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase {
			return true
		}
		if e.Kind == candidate.KindQuery && e.Query != nil && queryRelates(c, e.Query.Rows) {
			return true
		}
	}
	return false
}

func unsettled(c candidate.Candidate) bool {
	if !counts(c) {
		return false
	}
	return c.Refutation.TestCannotRun || c.Refutation.QueryCannotRun
}

func refuteWhy(c candidate.Candidate) string {
	r := c.Refutation
	if r.TestFound && !r.TestFailsOnChange {
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

func denial(cell string) bool {
	switch strings.ToLower(strings.TrimSpace(cell)) {
	case "false", "no", "absent", "clean", "pass", "passed":
		return true
	default:
		return false
	}
}

func tokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}
