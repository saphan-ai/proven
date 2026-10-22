// Package generator is the one model interface. Agents call it; they do not call a model.
package generator

import "context"

// Generator answers one prompt for one agent.
type Generator interface {
	Complete(ctx context.Context, agent, prompt string) (string, error)
}
