package testrun

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/service"
)

// ProverOffReason is the live adapter's answer while no container runner exists.
const ProverOffReason = "Prover off: no container runner yet"

// Live runs no code. review.Open uses the container adapter instead.
// A test can still inject Live to name the Prover off. It never runs on the host.
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
