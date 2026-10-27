package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/saphan-ai/proven/internal/markov"
)

// markovOp runs one Markovian operation (ADR-0048 point 3). The runner's generator is a TypeScript function, so
// the shim cannot hand it to Go: it asks "start" for the first prompt, calls the generator, and hands the answer
// to "feed" with the machine the last answer carried, until the run is done. Every decision of the loop is the
// Go machine's; the shim only carries text both ways.
func markovOp(req request) map[string]any {
	switch req.Op {
	case "config":
		cfg, err := chunkConfig(req.ChunkConfig)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"config": cfg})
	case "start":
		cfg, err := chunkConfig(req.ChunkConfig)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		// The spec's default is a run with its reflection pass; only an explicit false turns it off.
		reflect := req.Reflect == nil || *req.Reflect
		m, err := markov.Start(cfg, markov.Options{
			Goal: req.Goal, Mode: req.Mode, MemoryContext: req.MemoryContext,
			Attachments: req.Attachments, Reflect: reflect,
		}, req.At)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		return machineAnswer(req.ID, m)
	case "feed":
		var m markov.Machine
		if err := json.Unmarshal(req.Machine, &m); err != nil {
			return fail(req.ID, "markov: the machine does not decode: "+err.Error())
		}
		var genErr error
		if req.Failure != nil {
			genErr = errors.New(*req.Failure)
		}
		m.Feed(req.Text, genErr, req.Aborted, req.At)
		return machineAnswer(req.ID, &m)
	case "modelTotals":
		return ok(req.ID, map[string]any{"totals": markov.ModelTotals(req.N, markov.DocDefaults)})
	case "runEfficiency":
		// The configuration is the run's own snapshot and is taken as given, never validated against the
		// current settings (AGENTS.md section 2.3).
		var cfg markov.Config
		if len(bytes.TrimSpace(req.ChunkConfig)) > 0 {
			if err := json.Unmarshal(req.ChunkConfig, &cfg); err != nil {
				return fail(req.ID, "markov: the config does not decode: "+err.Error())
			}
		}
		return ok(req.ID, map[string]any{"efficiency": markov.RunEfficiency(req.Chunks, cfg)})
	default:
		return fail(req.ID, "markov: unknown op "+req.Op)
	}
}

// machineAnswer is the machine, the prompt it asks for next, and the finished run once there is none.
func machineAnswer(id int, m *markov.Machine) map[string]any {
	prompt, system, more := m.Next()
	fields := map[string]any{"machine": m, "done": !more, "step": m.Step()}
	if more {
		fields["prompt"] = prompt
		fields["system"] = system
	} else {
		fields["run"] = m.Record
	}
	return ok(id, fields)
}

// chunkConfig reads a partial chunk configuration over the spec's defaults and validates it. A key the spec does
// not name, or a value that is not a whole number, is refused by name rather than ignored.
func chunkConfig(raw json.RawMessage) (markov.Config, error) {
	cfg := markov.DefaultConfig()
	if len(bytes.TrimSpace(raw)) == 0 || string(bytes.TrimSpace(raw)) == "null" {
		return cfg, cfg.Validate()
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return markov.Config{}, fmt.Errorf("Markovian config refused: not an object: %v", err)
	}
	targets := map[string]*int{
		"chunkSize": &cfg.ChunkSize, "maxChunks": &cfg.MaxChunks,
		"carryoverTokens": &cfg.CarryoverTokens, "overlapTokens": &cfg.OverlapTokens,
	}
	for k, v := range fields {
		dst, known := targets[k]
		if !known {
			return markov.Config{}, &markov.ConfigError{Lead: markov.RefusalLead, Field: k, Value: string(v), Reason: "unknown field",
				Allowed: []string{"chunkSize", "maxChunks", "carryoverTokens", "overlapTokens"}}
		}
		if err := json.Unmarshal(v, dst); err != nil {
			return markov.Config{}, &markov.ConfigError{Lead: markov.RefusalLead, Field: k, Value: string(v), Reason: "not a whole number",
				Allowed: []string{"a whole number"}}
		}
	}
	return cfg, cfg.Validate()
}
