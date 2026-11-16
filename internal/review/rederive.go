package review

import (
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/report"
	"github.com/saphan-ai/proven/internal/service"
)

type Outcome = report.Outcome

func Rederive(f Finding) (Outcome, error)               { return report.Rederive(f) }
func TextReadings(f Finding) (Finding, Finding)         { return report.TextReadings(f) }
func CountProofs(findings []Finding) map[ProofLevel]int { return report.CountProofs(findings) }

// toRefutation keeps the facts the Judge's rule reads from the Skeptic's attempt.
func toRefutation(r candidate.Refutation) *Refutation {
	return &Refutation{
		TestFound: r.TestFound, TestFailsOnChange: r.TestFailsOnChange, TestCannotRun: r.TestCannotRun,
		QueryRows: service.CopyRows(r.QueryRows), QueryCannotRun: r.QueryCannotRun,
	}
}
