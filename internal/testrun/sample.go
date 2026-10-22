package testrun

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
)

// Recorded is one test outcome stored in the fixture.
type Recorded struct {
	FindingID     string
	Path          string
	Name          string
	Command       string
	Output        string
	FailsOnChange bool
	PassesOnBase  bool
}

// Sample returns recorded outcomes. It executes no code.
type Sample struct {
	byID map[string]Recorded
}

// NewSample indexes recorded tests by finding id.
func NewSample(rows []Recorded) *Sample {
	s := &Sample{byID: make(map[string]Recorded, len(rows))}
	for _, row := range rows {
		s.byID[row.FindingID] = row
	}
	return s
}

// Run returns the recorded outcome for this finding and test name.
func (s *Sample) Run(ctx context.Context, req Request) (Outcome, error) {
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	rec, ok := s.byID[req.FindingID]
	if !ok || rec.Name != req.Name {
		return Outcome{Found: false}, nil
	}
	return Outcome{
		Found:         true,
		Path:          rec.Path,
		Name:          rec.Name,
		Command:       rec.Command,
		Output:        rec.Output,
		FailsOnChange: rec.FailsOnChange,
		PassesOnBase:  rec.PassesOnBase,
	}, nil
}
