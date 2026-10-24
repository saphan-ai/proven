package generator

import (
	"context"

	"github.com/saphan-ai/proven/internal/service"
)

// Absent is a generator that does not dial.
// Complete names the provider down. ADR-0013 point 4a: an empty token or key
// is this, not a refused configuration and not a call.
type Absent struct {
	// Service names the provider that is down; empty means nightshift.
	Service string
	Reason  string
}

// Complete returns the provider's failure and does not use the network.
func (a Absent) Complete(context.Context, string, string) (string, error) {
	reason := a.Reason
	if reason == "" {
		reason = "not configured"
	}
	svc := a.Service
	if svc == "" {
		svc = "nightshift"
	}
	return "", &service.Down{Service: svc, Reason: reason}
}
