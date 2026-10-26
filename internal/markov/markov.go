// Package markov is Meterless's Markovian engine in Go: long work cut into bounded steps, with only a small
// carried state passed from one step to the next (ADR-0048).
//
// It follows the Markovian specification (Meterless engines/markovian, AGENTS.md): the chunk configuration with
// its bounds and load-time validation, the marker protocol, the carryover cascade, the two mode prompts, the
// orchestrator loop with its reflection pass, and the efficiency math of docs/efficiency-model.md. Meterless's own
// conformance runner checks it through cmd/meterless-conform and test/conformance/meterless/markovian.
//
// Proven uses the same engine in two places, and only when its switch is on (ADR-0048 points 1 and 2): a change
// above the configured size is reviewed one part per step with a carried state (established, suspected, awaiting
// evidence), and each new attempt of the proved-fix loop carries only what failed and why. The switch is off by
// default; ADR-0048 point 4 keeps it off until its measurement holds.
//
// Every token figure here is the reference estimator ceil(chars/4), labelled "estimated": Proven's generator
// interface returns no provider usage, so nothing is labelled "measured".
package markov

import (
	"fmt"

	"github.com/saphan-ai/proven/internal/service"
)

// RefusalLead is the opening of a Markovian configuration refusal.
// The shared error prints "configuration refused" when Lead is empty, and
// this sentence is the one Markovian refusals already printed.
const RefusalLead = "Markovian config refused"

// Config is the spec's ChunkConfig (AGENTS.md section 2.1). JSON names follow the spec, so a run record reads
// the same in Go and in the conformance runner.
type Config struct {
	// ChunkSize is the target generation budget of one step.
	ChunkSize int `json:"chunkSize"`
	// MaxChunks is the most steps one run may take.
	MaxChunks int `json:"maxChunks"`
	// CarryoverTokens is the budget of the state carried into the next step.
	CarryoverTokens int `json:"carryoverTokens"`
	// OverlapTokens is how much of a step's cleaned output, in tokens, is repeated into the next step's prompt.
	OverlapTokens int `json:"overlapTokens"`
}

// DefaultConfig is the spec's reference default.
func DefaultConfig() Config {
	return Config{ChunkSize: 8000, MaxChunks: 24, CarryoverTokens: 800, OverlapTokens: 0}
}

// bound is one field's allowed range.
type bound struct{ min, max int }

// The spec's enforced bounds (AGENTS.md section 2.1). The spec's "step" values are slider granularity: its own
// default carryoverTokens of 800 is not on a 128 grid from 128, so only the minimum and maximum are enforced,
// as the reference implementation rules.
var (
	boundChunkSize = bound{1000, 128000}
	boundMaxChunks = bound{1, 32}
	boundCarryover = bound{128, 32768}
	boundOverlap   = bound{0, 4096}
)

// FramingTokens and OutputBudget are the spec's engine constants for the load-time check that a step's budget
// holds its framing, its carried state and its answer.
const (
	FramingTokens = 400
	OutputBudget  = 1200
)

// LabelEstimated is the label of a number computed from ceil(chars/4) (AGENTS.md section 14).
const LabelEstimated = "estimated"

// EstimateTokens is the spec's reference estimator: ceil(chars/4), counting bytes as the reference counts
// UTF-16 units. It is an estimate and is labelled so wherever it is reported.
func EstimateTokens(s string) int {
	return (len(s) + 3) / 4
}

// ConfigError is a refused configuration value. It names the field, the value and what is allowed; a value is
// never clamped or silently replaced (AGENTS.md section 2.1, ADR-0013).
type ConfigError = service.ConfigError

// Validate refuses a configuration outside the spec's bounds, and one whose step budget cannot hold the framing,
// the carried state and the answer. The context-window check of the spec is the caller's: this engine does not
// know which model answers.
func (c Config) Validate() error {
	fields := []struct {
		name string
		v    int
		b    bound
	}{
		{"chunkSize", c.ChunkSize, boundChunkSize},
		{"maxChunks", c.MaxChunks, boundMaxChunks},
		{"carryoverTokens", c.CarryoverTokens, boundCarryover},
		{"overlapTokens", c.OverlapTokens, boundOverlap},
	}
	for _, f := range fields {
		if f.v < f.b.min || f.v > f.b.max {
			return &ConfigError{Lead: RefusalLead,
				Field: f.name, Value: fmt.Sprint(f.v), Reason: "outside the spec's bounds",
				Allowed: []string{fmt.Sprintf("a whole number from %d to %d", f.b.min, f.b.max)},
			}
		}
	}
	if need := FramingTokens + c.CarryoverTokens + OutputBudget; c.ChunkSize < need {
		return &ConfigError{Lead: RefusalLead,
			Field: "chunkSize", Value: fmt.Sprint(c.ChunkSize),
			Reason:  fmt.Sprintf("below framing (%d) + carryover (%d) + output budget (%d)", FramingTokens, c.CarryoverTokens, OutputBudget),
			Allowed: []string{fmt.Sprintf("at least %d with carryoverTokens %d", need, c.CarryoverTokens)},
		}
	}
	return nil
}
