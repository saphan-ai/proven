package markov

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// The run statuses of AGENTS.md section 2.3.
const (
	StatusCompleted = "completed"
	StatusMaxChunks = "max-chunks"
	StatusErrored   = "errored"
	StatusAborted   = "aborted"
)

// The machine's phases: it asks for a step, then for the reflection, then has nothing more to ask.
const (
	phaseStep     = "step"
	phaseReflect  = "reflect"
	phaseFinished = "finished"
)

// storedPrompt is how much of the goal a run record keeps (AGENTS.md section 2.3: truncated for storage).
const storedPrompt = 500

// reflectionOutput bounds the output the reflection prompt carries (AGENTS.md section 7: truncated if huge).
const reflectionOutput = 4000

// reflectionSystem is the system prompt of the reflection pass.
const reflectionSystem = "You are the reflection pass. Be brief and concrete."

// Run is the spec's run record, MarkovianRun (AGENTS.md section 2.3).
type Run struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Prompt    string `json:"prompt"`
	Mode      string `json:"mode"`
	// ChunkConfig is the mandatory snapshot: the run's efficiency is computed against it, never the current one.
	ChunkConfig       Config  `json:"chunkConfig"`
	Chunks            []Chunk `json:"chunks"`
	TotalTokensUsed   int     `json:"totalTokensUsed"`
	TotalTokensSaved  int     `json:"totalTokensSaved"`
	EfficiencyPercent float64 `json:"efficiencyPercent"`
	DurationMs        int64   `json:"durationMs"`
	Status            string  `json:"status"`
	// TokenLabel says where every number came from (AGENTS.md section 14).
	TokenLabel string `json:"tokenLabel"`
	// Output is the steps' cleaned content, joined.
	Output     string `json:"output"`
	Reflection string `json:"reflection,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Options are one run's inputs. Reflect off skips the reflection pass; the spec's default is on.
type Options struct {
	Goal          string       `json:"goal"`
	Mode          string       `json:"mode"`
	MemoryContext string       `json:"memoryContext,omitempty"`
	Attachments   []Attachment `json:"attachments,omitempty"`
	Reflect       bool         `json:"reflect"`
}

// Machine is the orchestrator loop of AGENTS.md section 4.5 held as data: Next names the prompt the generator
// must answer, Feed takes the answer. Its fields are exported and JSON-encoded so the conformance command, which
// lives one process per request, can carry a run across calls; Run drives the same machine in one process.
type Machine struct {
	Config          Config  `json:"config"`
	Options         Options `json:"options"`
	Record          Run     `json:"run"`
	Carryover       string  `json:"carryover"`
	PreviousCleaned string  `json:"previousCleaned"`
	// Savings is the running estimate of section 6.1; the record's total is recomputed from the chunks at the end.
	Savings int    `json:"savings"`
	Done    bool   `json:"done"`
	Phase   string `json:"phase"`
	Start   int64  `json:"start"`
	// Levels is the cascade level that produced each carried state, one per step.
	Levels []string `json:"levels"`
}

// Start checks the configuration and the goal and returns a machine at its first step. now is the clock in Unix
// milliseconds; the caller owns the clock so a run can be replayed.
func Start(cfg Config, opts Options, now int64) (*Machine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.Goal) == "" {
		return nil, errors.New("Markovian: a run needs a goal")
	}
	opts.Mode = modeOrDefault(opts.Mode)
	if opts.Mode != ModeArchitect && opts.Mode != ModeResearch {
		return nil, &ConfigError{Lead: RefusalLead, Field: "mode", Value: opts.Mode, Reason: "unknown mode", Allowed: []string{ModeArchitect, ModeResearch}}
	}
	return &Machine{
		Config:  cfg,
		Options: opts,
		Record: Run{
			ID:          fmt.Sprintf("run-%d", now),
			Timestamp:   now,
			Prompt:      clipRunes(opts.Goal, storedPrompt),
			Mode:        opts.Mode,
			ChunkConfig: cfg,
			Chunks:      []Chunk{},
			Status:      StatusCompleted,
			TokenLabel:  LabelEstimated,
		},
		Phase: phaseStep,
		Start: now,
	}, nil
}

// Step is the 1-based index of the step Next asks for.
func (m *Machine) Step() int {
	return len(m.Record.Chunks) + 1
}

// Next is the prompt and the system prompt the generator answers next, with ok false when the run is finished.
func (m *Machine) Next() (prompt, system string, ok bool) {
	switch m.Phase {
	case phaseStep:
		step := m.Step()
		in := PromptInput{
			Mode: m.Options.Mode, Goal: m.Options.Goal, Step: step, Carryover: m.Carryover,
			Overlap:       OverlapSlice(m.PreviousCleaned, m.Config.OverlapTokens),
			MemoryContext: m.Options.MemoryContext,
		}
		// Attachments enter the first step only (AGENTS.md section 5).
		if step == 1 {
			in.Attachments = m.Options.Attachments
		}
		return BuildPrompt(in), "", true
	case phaseReflect:
		return m.reflectionPrompt(), reflectionSystem, true
	default:
		return "", "", false
	}
}

// Feed takes the answer to the prompt Next named. genErr is the generator's failure, if any; aborted says the
// caller's abort signal fired. A step that fails ends the run as errored with its partial output kept; a
// reflection that fails leaves the run completed (AGENTS.md sections 7 and 11).
func (m *Machine) Feed(text string, genErr error, aborted bool, now int64) {
	switch m.Phase {
	case phaseStep:
		m.feedStep(text, genErr, aborted, now)
	case phaseReflect:
		// An abort before the reflection leaves the run as it stood: nothing was generated to keep.
		if aborted {
			m.finish(now)
			return
		}
		if genErr != nil {
			m.Record.Error = "reflection failed (run preserved): " + genErr.Error()
		} else {
			m.Record.Reflection = CleanMarkers(text)
		}
		m.finish(now)
	}
}

// feedStep accounts one step and decides whether the run goes on.
func (m *Machine) feedStep(text string, genErr error, aborted bool, now int64) {
	// The abort is checked before the step is accepted: an aborted generator call produced nothing to keep.
	if aborted {
		m.Record.Status = StatusAborted
		m.close(now)
		return
	}
	if genErr != nil {
		m.Record.Status = StatusErrored
		m.Record.Error = genErr.Error()
		m.close(now)
		return
	}
	step := m.Step()
	parsed := ParseMarkers(text)
	tokens := EstimateTokens(parsed.Cleaned)
	// Section 6.1: the append-history step pays every earlier step at its full budget; this one pays the state.
	standard := (step-1)*m.Config.ChunkSize + tokens
	markovian := m.Config.CarryoverTokens + tokens
	if standard > markovian {
		m.Savings += standard - markovian
	}
	// The chunk is kept before the state is compressed: its cost is paid, so it survives anything after.
	m.Record.Chunks = append(m.Record.Chunks, Chunk{ID: step, Tokens: tokens, Carryover: m.Carryover, Content: parsed.Cleaned})
	m.PreviousCleaned = parsed.Cleaned
	carry, level := NextCarryover(parsed.State, parsed.HasState, m.Carryover, text, m.Config.CarryoverTokens, nil)
	m.Carryover = carry
	m.Levels = append(m.Levels, level)
	if parsed.Done {
		m.Done = true
	}
	if m.Done || len(m.Record.Chunks) >= m.Config.MaxChunks {
		if !m.Done {
			m.Record.Status = StatusMaxChunks
		}
		m.close(now)
	}
}

// close ends the step phase: it totals the run against its own snapshot and moves to the reflection when the run
// completed with at least one step and reflection is on, or to the end otherwise.
func (m *Machine) close(now int64) {
	parts := make([]string, 0, len(m.Record.Chunks))
	for _, c := range m.Record.Chunks {
		parts = append(parts, c.Content)
	}
	m.Record.Output = strings.Join(parts, "\n\n")
	e := RunEfficiency(m.Record.Chunks, m.Record.ChunkConfig)
	m.Record.TotalTokensUsed = e.TotalTokensUsed
	m.Record.TotalTokensSaved = e.TokensSaved
	m.Record.EfficiencyPercent = e.EfficiencyPercent
	if m.Options.Reflect && m.Record.Status == StatusCompleted && len(m.Record.Chunks) > 0 {
		m.Phase = phaseReflect
		return
	}
	m.finish(now)
}

// finish stamps the duration and ends the run.
func (m *Machine) finish(now int64) {
	m.Record.DurationMs = now - m.Start
	m.Phase = phaseFinished
}

// reflectionPrompt is the post-chain synthesis prompt of AGENTS.md section 7.
func (m *Machine) reflectionPrompt() string {
	blocks := strings.Count(m.Record.Output, "```") / 2
	return "Reflect on this completed run.\nGoal: " + clipRunes(m.Options.Goal, goalSummary) +
		"\nFinal state: " + m.Carryover +
		"\nOutput (truncated): " + clipRunes(m.Record.Output, reflectionOutput) +
		fmt.Sprintf("\nCode blocks: %d\n", blocks) +
		"Verdict: ratify if the goal is met, refine with concrete gaps otherwise."
}

// Generator answers one prompt with a system prompt; it is the spec's provider-agnostic generator.
type Generator func(ctx context.Context, prompt, system string) (string, error)

// Execute drives a machine to its end in one process. A context that ends before a step aborts the run; one that
// ends before the reflection ends the run without it.
func Execute(ctx context.Context, cfg Config, opts Options, gen Generator, clock func() int64) (Run, error) {
	m, err := Start(cfg, opts, clock())
	if err != nil {
		return Run{}, err
	}
	for {
		prompt, system, ok := m.Next()
		if !ok {
			return m.Record, nil
		}
		if ctx.Err() != nil {
			if m.Phase == phaseReflect {
				m.finish(clock())
				continue
			}
			m.Feed("", nil, true, clock())
			continue
		}
		text, gerr := gen(ctx, prompt, system)
		m.Feed(text, gerr, false, clock())
	}
}
