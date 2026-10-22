// Package judge keeps a finding only when it has evidence.
package judge

import "github.com/saphan-ai/proven/internal/candidate"

// Decide drops a finding that has neither a qualifying test nor a query that returned rows.
// The drop reason for that case is exactly "dropped: no evidence".
func Decide(c candidate.Candidate) candidate.Candidate {
	if counts(c.Evidence) {
		c.Dropped = false
		c.DropReason = "kept"
		return c
	}
	c.Dropped = true
	c.DropReason = "dropped: no evidence"
	return c
}

// Evidenced reports whether Decide would keep a candidate with this evidence.
func Evidenced(ev []candidate.Evidence) bool {
	return counts(ev)
}

func counts(ev []candidate.Evidence) bool {
	for _, e := range ev {
		if e.Kind == candidate.KindTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase {
			return true
		}
		if e.Kind == candidate.KindQuery && e.Query != nil && len(e.Query.Rows) > 0 {
			return true
		}
	}
	return false
}
