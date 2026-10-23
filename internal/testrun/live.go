package testrun

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/service"
)

// ProverOffReason is the reason Live reports. It was the live answer before
// the container adapter existed; Live keeps it so a test can name the Prover off.
const ProverOffReason = "Prover off: no container runner yet"

// Live runs no code. review.Open uses the container adapter instead.
// A test can still inject Live to name the Prover off. It never runs on the host.
type Live struct{}

// Run reports the Prover as off.
func (Live) Run(ctx context.Context, req Request) (Outcome, error) {
	// The same bound as the sample adapters, so a test sees the same deadline behaviour.
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	// ADR-0025 point 6: with no container the Prover runs no code, and says why.
	return Outcome{}, &service.Down{Service: "prover", Reason: ProverOffReason}
}
