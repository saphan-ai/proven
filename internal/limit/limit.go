// Package limit bounds one adapter call so a review cannot hang on it.
//
// The sample adapters (generator, guardrail, query, test run) and the live test
// runner's stand-in wrap each call in Bound, so a sample review keeps the same
// deadline behaviour as a live one. A call past the bound returns the context's
// error.
package limit

import (
	"context"
	"time"
)

// Adapter is the default bound on one adapter call when the caller did not set
// a shorter deadline. The shell's context still wins when it is sooner.
const Adapter = 5 * time.Second

// Bound returns a context that ends at Adapter or at the parent's deadline,
// whichever is sooner.
//
// The caller must call the returned CancelFunc, as with context.WithTimeout.
func Bound(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, Adapter)
}
