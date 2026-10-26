// Package testrun is a worked example of a test-runner and fix-runner adapter.
//
// Runner returns canned test outcomes and canned fix runs. It executes no
// code and starts no container. A reader copies this package and replaces the
// records with their own. What to add is in docs/ADAPTERS.md under
// "Adding a test run".
package testrun

import (
	"context"
	"errors"
	"strconv"

	it "github.com/saphan-ai/proven/internal/testrun"
)

var _ it.Adapter = Runner{}
var _ it.FixRunner = Runner{}

// ErrUnknownFinding is the refusal: the runner records no outcome for this
// finding or attempt.
var ErrUnknownFinding = errors.New("testrun: no recorded outcome for this finding or attempt")

// Recorded is one canned test outcome keyed by finding id.
type Recorded struct {
	FindingID string
	Outcome   it.Outcome
}

// RecordedFix is one canned proof of a fix, keyed by finding id and attempt.
type RecordedFix struct {
	FindingID string
	Attempt   int
	Outcome   it.FixOutcome
}

// Runner returns canned outcomes. It executes no code and starts no container.
type Runner struct {
	tests map[string]it.Outcome
	fixes map[fixKey]it.FixOutcome
}

type fixKey struct {
	id      string
	attempt int
}

// New indexes recorded outcomes by finding id and recorded fix attempts by
// finding id and attempt number. A duplicate is refused.
func New(tests []Recorded, fixes []RecordedFix) (*Runner, error) {
	testsMap := make(map[string]it.Outcome, len(tests))
	for _, row := range tests {
		if _, ok := testsMap[row.FindingID]; ok {
			return nil, errors.New("testrun: two outcomes for finding " + row.FindingID)
		}
		testsMap[row.FindingID] = row.Outcome
	}
	fixesMap := make(map[fixKey]it.FixOutcome, len(fixes))
	for _, row := range fixes {
		key := fixKey{row.FindingID, row.Attempt}
		if _, ok := fixesMap[key]; ok {
			return nil, errors.New("testrun: two runs for attempt " + row.FindingID + ":" + strconv.Itoa(row.Attempt))
		}
		fixesMap[key] = row.Outcome
	}
	return &Runner{tests: testsMap, fixes: fixesMap}, nil
}

// Run returns the recorded outcome for this finding and test name. The test
// name must match: a recording for another test of the same finding is not
// the outcome the prover asked for. Found is false when nothing was recorded.
// It executes no code.
func (r Runner) Run(ctx context.Context, req it.Request) (it.Outcome, error) {
	if err := ctx.Err(); err != nil {
		return it.Outcome{}, err
	}
	out, ok := r.tests[req.FindingID]
	if !ok || (out.Name != "" && out.Name != req.Name) {
		return it.Outcome{Found: false}, nil
	}
	return out, nil
}

// RunFix returns the recorded runs of this attempt. A missing attempt is
// refused by name and no code runs. It executes no code.
func (r Runner) RunFix(ctx context.Context, req it.FixRequest) (it.FixOutcome, error) {
	if err := ctx.Err(); err != nil {
		return it.FixOutcome{}, err
	}
	out, ok := r.fixes[fixKey{req.FindingID, req.Attempt}]
	if !ok {
		return it.FixOutcome{Refused: ErrUnknownFinding.Error()}, nil
	}
	return out, nil
}
