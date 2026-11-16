// Package generator defines the model transport used by hypothesis-generating agents.
// Live supports an OpenAI-compatible API, Command supports an operator-configured CLI,
// and Sample replays fixtures. Legacy Nightshift and NVIDIA providers remain optional.
// Unconfigured installations disable model stages and retain deterministic analysis.
package generator

import "context"

// Generator answers one prompt for one agent.
type Generator interface {
	Complete(ctx context.Context, agent, prompt string) (string, error)
}
