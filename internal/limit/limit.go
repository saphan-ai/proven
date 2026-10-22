// Package limit bounds one adapter call so a review cannot hang on it.
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
func Bound(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, Adapter)
}
