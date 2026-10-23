// Package generator is the one model interface. Agents call it; they do not call a model.
//
// The model-calling agents (Planner, Investigator, Prover and Skeptic) send
// their prompts through a Generator; the Judge and the Writer call no model. Live speaks to Meterless Nightshift (ADR-0019) and reads the
// answering model from its receipt; Sample answers from recorded responses
// with no network, for sample mode and tests (ADR-0001). An empty token does
// not build a live client: the service is down and Absent names it (ADR-0013
// point 4a).
package generator

import "context"

// Generator answers one prompt for one agent.
type Generator interface {
	Complete(ctx context.Context, agent, prompt string) (string, error)
}
