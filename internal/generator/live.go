package generator

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/service"
)

// Live is the live generator for this wagon: there is no provider yet.
type Live struct{}

// Complete reports the generator as down. It does not call a network.
func (Live) Complete(ctx context.Context, agent, prompt string) (string, error) {
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", &service.Down{Service: "generator", Reason: "no live generator in this wagon"}
}
