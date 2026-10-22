package testrun

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/service"
)

// ProverOffReason is the live adapter's answer while no container runner exists.
const ProverOffReason = "Prover off: no container runner yet"

// Live runs no code. A container runner is a later wagon, and this adapter
// does not fall back to the host.
type Live struct{}

// Run reports the Prover as off.
func (Live) Run(ctx context.Context, req Request) (Outcome, error) {
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	return Outcome{}, &service.Down{Service: "prover", Reason: ProverOffReason}
}
