package query

import (
	"context"

	"github.com/saphan-ai/proven/internal/service"
)

// Absent is a query adapter that does not dial.
// Query names Zetaris down, so a review has no query evidence from it and
// still names the service (ADR-0013 point 4a).
type Absent struct {
	Reason string
}

// Query returns the zetaris failure and does not use the network.
func (a Absent) Query(context.Context, string) (Result, error) {
	reason := a.Reason
	if reason == "" {
		reason = "not configured"
	}
	return Result{}, &service.Down{Service: "zetaris", Reason: reason}
}
