// Package writer fills the review draft from the judge's candidates.
package writer

import "github.com/saphan-ai/proven/internal/candidate"

// Input is everything the writer needs.
type Input struct {
	Candidates []candidate.Candidate
	Failed     []candidate.Failure
	Models     []candidate.ModelUse
	Partial    bool
	TracePath  string
}

// Write keeps the findings the judge did not drop and copies the rest of the draft.
func Write(in Input) candidate.Draft {
	d := candidate.Draft{
		Failed: in.Failed, Models: in.Models, Partial: in.Partial, TracePath: in.TracePath,
	}
	for _, c := range in.Candidates {
		if c.Dropped {
			continue
		}
		d.Findings = append(d.Findings, c)
	}
	return d
}
