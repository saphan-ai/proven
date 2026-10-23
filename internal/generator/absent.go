package generator

import (
	"context"

	"github.com/saphan-ai/proven/internal/service"
)

// Absent is a generator that does not dial.
// Complete names Nightshift down. ADR-0013 point 4a: an empty token is this,
// not a refused configuration and not a call.
type Absent struct {
	Reason string
}

// Complete returns the nightshift failure and does not use the network.
func (a Absent) Complete(context.Context, string, string) (string, error) {
	reason := a.Reason
	if reason == "" {
		reason = "not configured"
	}
	return "", &service.Down{Service: "nightshift", Reason: reason}
}
