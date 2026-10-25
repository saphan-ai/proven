package guardrail

import (
	"context"
	"fmt"
	"time"

	"github.com/saphan-ai/proven/internal/limit"
)

// Safe answers safe for every prompt and every answer. It opens no network.
// The sample review uses it, so the fixture's findings stay as they are.
type Safe struct{}

// Check returns a safe verdict. The latency is the time this call took.
func (Safe) Check(ctx context.Context, prompt, answer string) (Verdict, error) {
	start := time.Now()
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Verdict{}, err
	}
	v := Verdict{User: "safe", Latency: time.Since(start)}
	if answer != "" {
		v.Response = "safe"
	}
	return v, nil
}

// Identity names this adapter "sample", the name the pipeline gives the
// guardrail sample mode falls back to. It calls no model.
func (Safe) Identity() (backend, model string) { return "sample", "" }

// Record is one recorded verdict. Answer is empty for a prompt-only check.
type Record struct {
	Prompt     string
	Answer     string
	User       string
	Response   string
	Categories []string
}

// RecordedUnsafePrompt is the measured unsafe-prompt verdict.
// Tests bind it to one prompt from the discount fixture.
// The sample review does not use it, and answers safe instead.
func RecordedUnsafePrompt() Record {
	return Record{
		User:       "unsafe",
		Categories: []string{"Criminal Planning/Confessions", "Other"},
	}
}

// RecordedUnsafeAnswer is the measured verdict for a safe prompt and an unsafe answer.
func RecordedUnsafeAnswer() Record {
	return Record{
		User:       "safe",
		Response:   "unsafe",
		Categories: []string{"Malware", "PII/Privacy"},
	}
}

// Sample answers from recorded verdicts. A prompt it does not have is a refusal,
// not a safe verdict.
type Sample struct {
	rows map[string]Verdict
}

// NewSample indexes verdicts by prompt and answer. A duplicate key is refused.
func NewSample(rows []Record) (*Sample, error) {
	s := &Sample{rows: make(map[string]Verdict, len(rows))}
	for _, row := range rows {
		if row.User != "safe" && row.User != "unsafe" {
			return nil, fmt.Errorf("guardrail: recorded user verdict %q", row.User)
		}
		if row.Answer != "" && row.Response != "safe" && row.Response != "unsafe" {
			return nil, fmt.Errorf("guardrail: recorded response verdict %q", row.Response)
		}
		if row.Answer == "" && row.Response != "" {
			return nil, fmt.Errorf("guardrail: a prompt-only record has a response verdict")
		}
		k := recordKey(row.Prompt, row.Answer)
		if _, ok := s.rows[k]; ok {
			return nil, fmt.Errorf("guardrail: duplicate recorded verdict")
		}
		s.rows[k] = Verdict{
			User:       row.User,
			Response:   row.Response,
			Categories: append([]string(nil), row.Categories...),
		}
	}
	return s, nil
}

// Check returns the recorded verdict. A missing record is a named refusal.
func (s *Sample) Check(ctx context.Context, prompt, answer string) (Verdict, error) {
	start := time.Now()
	// The sample keeps the same time bound as a live check, so a cancelled
	// review stops here too.
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Verdict{}, err
	}
	if s == nil {
		return Verdict{}, Refuse("not configured")
	}
	v, ok := s.rows[recordKey(prompt, answer)]
	if !ok {
		return Verdict{}, Refuse("no recorded verdict")
	}
	v.Latency = time.Since(start)
	// A copy, so a caller cannot change the recorded verdict.
	v.Categories = append([]string(nil), v.Categories...)
	return v, nil
}

// Identity names this adapter "sample-recorded", so a count can tell recorded
// verdicts from the always-safe sample. It calls no model.
func (*Sample) Identity() (backend, model string) { return "sample-recorded", "" }

// recordKey joins prompt and answer with a NUL byte, so two different pairs never share a key.
func recordKey(prompt, answer string) string {
	return prompt + "\x00" + answer
}
