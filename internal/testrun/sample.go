package testrun

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
)

// Recorded is one test outcome stored in the fixture. FindingID is the key
// the Prover looks it up by; the other fields are copied into the Outcome.
type Recorded struct {
	FindingID     string
	Path          string
	Name          string
	Command       string
	Output        string
	FailsOnChange bool
	PassesOnBase  bool
}

// RecordedFix is the recorded runs of one attempt to prove a fix (ADR-0042).
// FindingID and Attempt are the key the sample looks it up by.
type RecordedFix struct {
	FindingID string
	Attempt   int
	Runs      []FixRun
}

// Sample returns recorded outcomes. It executes no code.
type Sample struct {
	byID  map[string]Recorded
	fixes map[fixKey]RecordedFix
}

// fixKey names one attempt at one finding.
type fixKey struct {
	id      string
	attempt int
}

// NewSample indexes recorded tests by finding id. Recorded fix attempts, when
// given, are indexed by finding id and attempt number.
func NewSample(rows []Recorded, fixes ...RecordedFix) *Sample {
	s := &Sample{byID: make(map[string]Recorded, len(rows)), fixes: make(map[fixKey]RecordedFix, len(fixes))}
	for _, row := range rows {
		s.byID[row.FindingID] = row
	}
	for _, f := range fixes {
		s.fixes[fixKey{f.FindingID, f.Attempt}] = f
	}
	return s
}

// RunFix returns the recorded runs of this attempt. It executes no code, but it
// does check the patch against the real change tree: a recorded answer whose edits
// do not apply to the trees is refused as it would be live, so a stale fixture
// cannot show a fix that no longer applies (ADR-0042).
func (s *Sample) RunFix(ctx context.Context, req FixRequest) (FixOutcome, error) {
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return FixOutcome{}, err
	}
	if reason := RefuseEdits(req.ChangeDir, req.Edits, 0); reason != "" {
		return FixOutcome{Refused: reason}, nil
	}
	rec, ok := s.fixes[fixKey{req.FindingID, req.Attempt}]
	if !ok {
		// No recording is not a failed fix: nothing was run, so nothing is claimed.
		return FixOutcome{Refused: "the sample records no run of this attempt"}, nil
	}
	// ADR-0042: the replay judges a recorded run as the container judges a live one, by
	// its output and not by the recorded word alone.
	runs := append([]FixRun(nil), rec.Runs...)
	for i, r := range runs {
		if r.Result == ResultPassed && !RecordedPass(r.Command, r.Output) {
			runs[i].Result = ResultNotRun
			runs[i].Output += "\nthe recorded output has no pass line for the run: it is not a pass"
		}
	}
	return FixOutcome{Runs: runs}, nil
}

// Run returns the recorded outcome for this finding and test name.
func (s *Sample) Run(ctx context.Context, req Request) (Outcome, error) {
	// The sample keeps the live deadline behaviour, so a sample review times out the same way.
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	rec, ok := s.byID[req.FindingID]
	// The test name must match too: a recording for another test of the same
	// finding is not evidence for the test the Prover asked about.
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
