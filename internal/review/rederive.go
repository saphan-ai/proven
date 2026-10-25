package review

// This file holds the rule that turns a finding's recorded evidence into its outcome, for the
// reviewer that fills a Result and for `proven verify` that checks a pack (ADR-0035). There is one
// rule: the Judge's own, called here and not copied.

import (
	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
)

// Outcome is what Proven's rule gives for one finding's recorded evidence.
type Outcome struct {
	// Level is the finding's proof level (ProofLevelOf).
	Level ProofLevel
	// Reported is false when the Judge drops the finding: its own evidence does not count, or the
	// recorded refutation holds.
	Reported bool
	// Unresolved is true when the Judge keeps the finding and marks it open.
	Unresolved bool
	// Complete is false when the finding records no refutation. Reported then rests on the
	// finding's own evidence alone and Unresolved is not derived.
	Complete bool
}

// Rederive applies the Judge's rule and ProofLevelOf to what one finding records. It runs no test,
// no query and no model. It returns an error when the evidence has no kind that carries a level.
func Rederive(f Finding) (Outcome, error) {
	level, err := ProofLevelOf(f.Evidence)
	if err != nil {
		return Outcome{}, err
	}
	c := candidate.Candidate{Risk: candidate.Risk{
		ID: f.ID, RuleID: f.RuleID, Title: f.Title, Message: f.Message, Path: f.Location.Path,
	}}
	for _, e := range f.Evidence {
		c.Evidence = append(c.Evidence, candidateEvidence(e))
	}
	if r := f.Refutation; r != nil {
		c.Refutation = candidate.Refutation{
			TestFound: r.TestFound, TestFailsOnChange: r.TestFailsOnChange, TestCannotRun: r.TestCannotRun,
			QueryRows: r.QueryRows, QueryCannotRun: r.QueryCannotRun,
		}
	}
	c = judge.Decide(c)
	return Outcome{Level: level, Reported: !c.Dropped, Unresolved: c.Unresolved, Complete: f.Refutation != nil}, nil
}

// CountProofs counts findings by proof level. Every level is a key, with 0 when none.
func CountProofs(findings []Finding) map[ProofLevel]int {
	counts := map[ProofLevel]int{}
	for _, l := range ProofLevels {
		counts[l] = 0
	}
	for _, f := range findings {
		counts[f.ProofLevel]++
	}
	return counts
}

// candidateEvidence is toEvidence turned around: the Judge reads the pipeline's type.
func candidateEvidence(e Evidence) candidate.Evidence {
	out := candidate.Evidence{Kind: e.Kind}
	if t := e.Test; t != nil {
		out.Test = &candidate.TestEvidence{
			Path: t.Path, Name: t.Name, Command: t.Command, Output: t.Output,
			FailsOnChange: t.FailsOnChange, PassesOnBase: t.PassesOnBase,
		}
	}
	if q := e.Query; q != nil {
		out.Query = &candidate.QueryEvidence{
			SQL: q.SQL, Form: q.Form, Engine: q.Engine, Snapshot: q.Snapshot,
			Columns: q.Columns, Rows: q.Rows, Bound: q.Bound,
		}
	}
	if t := e.Tool; t != nil {
		out.Tool = &candidate.ToolEvidence{
			Tool: t.Tool, Version: t.Version, Command: t.Command, Output: t.Output, BaseOutput: t.BaseOutput,
			PresentOnChange: t.PresentOnChange, PresentOnBase: t.PresentOnBase,
		}
	}
	return out
}

// toRefutation keeps the facts the Judge's rule reads from the Skeptic's attempt.
func toRefutation(r candidate.Refutation) *Refutation {
	return &Refutation{
		TestFound: r.TestFound, TestFailsOnChange: r.TestFailsOnChange, TestCannotRun: r.TestCannotRun,
		QueryRows: copyRows(r.QueryRows), QueryCannotRun: r.QueryCannotRun,
	}
}
