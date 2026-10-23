// Package service is the shared shape of an adapter that failed.
//
// Every adapter (Nightshift generator, Zetaris query, guardrail, test runner)
// returns a *Down when its outside service fails. The pipeline catches it with
// As, records the service in the review's failed services and carries on with
// what it has: a failed service degrades the review, it does not stop it.
package service

import "errors"

// Down is a failed outside service. The review continues and names it.
// Service is the short name the review prints, for example "zetaris";
// Reason says why the call failed.
type Down struct {
	Service string
	Reason  string
}

// Error returns "service: reason". A nil *Down still prints, so a typed nil
// that reaches an error check never panics.
func (d *Down) Error() string {
	if d == nil {
		return "service down"
	}
	return d.Service + ": " + d.Reason
}

// As reports a Down error.
// It unwraps err, so a Down wrapped with %w by a caller is still found.
func As(err error) (*Down, bool) {
	var d *Down
	if errors.As(err, &d) {
		return d, true
	}
	return nil, false
}
