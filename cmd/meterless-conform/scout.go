package main

// This file answers Scout Intent's calls (ADR-0049 point 4). A decision needs no session: each call
// names its registry, and the engine is built from it for that one call, so the recency term of the
// spec's formula, which needs a session, is always zero here.

import (
	"encoding/json"
	"strings"

	"github.com/saphan-ai/proven/internal/scout"
)

// scoutOp runs one Scout operation.
func scoutOp(req request) map[string]any {
	switch req.Op {
	case "decide", "classify":
	// The spec's replan reads the parent contract, which a one-process call does not keep; it is
	// refused by name rather than answered with a fresh decision that would claim a lineage it lacks.
	case "replan":
		return fail(req.ID, "scout: replan needs the parent contract, which this command does not keep")
	default:
		return fail(req.ID, "scout: unknown op "+req.Op)
	}
	reg, err := scout.LoadRegistry(req.Registry)
	if err != nil {
		return fail(req.ID, err.Error())
	}
	eng, err := scout.New(reg, scout.Options{PolicyPack: policyPacks(req.PolicyPack)})
	if err != nil {
		return fail(req.ID, err.Error())
	}
	d, err := eng.Decide(scout.Input{Prompt: req.Prompt, User: req.User, Surface: req.Surface})
	if err != nil {
		return fail(req.ID, err.Error())
	}
	if req.Op == "classify" {
		return ok(req.ID, map[string]any{"candidates": d.Candidates})
	}
	return ok(req.ID, map[string]any{"decision": d})
}

// policyPacks reads the spec's policyPack option, a name or a list of names.
func policyPacks(raw json.RawMessage) []string {
	var one string
	if json.Unmarshal(raw, &one) == nil && strings.TrimSpace(one) != "" {
		return []string{one}
	}
	var many []string
	if json.Unmarshal(raw, &many) == nil {
		return many
	}
	return nil
}
