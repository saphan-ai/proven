// Package limit bounds one adapter call so a review cannot hang on it.
//
// The bound is the setting PROVEN_ADAPTER_TIMEOUT (ADR-0013). Empty means
// Adapter. A wrong value is refused when a live review loads its settings,
// and it is never replaced by the default. Sample mode does not read the
// variable: Bound then uses Adapter.
//
// The sample adapters (generator, guardrail, query, test run) and the live test
// runner's stand-in wrap each call in Bound, so a sample review keeps the same
// deadline behaviour as a live one. A call past the bound returns the context's
// error.
package limit

import (
	"context"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/service"
)

const (
	// EnvAdapter is the environment variable that sets the bound on one
	// adapter call. Sample mode does not read it.
	EnvAdapter = "PROVEN_ADAPTER_TIMEOUT"
)

// Adapter is the bound on one adapter call when EnvAdapter is empty, and
// when the caller set no shorter deadline. The shell's context still wins
// when it is sooner.
const Adapter = 5 * time.Second

// ConfigError names a refused adapter-bound setting and the values that are
// allowed (ADR-0013).
type ConfigError = service.ConfigError

// Load reads EnvAdapter. An empty value is Adapter.
// ADR-0013: a value that is set and is not a duration greater than zero is
// refused. It is never replaced by the default.
func Load(env func(string) string) (time.Duration, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvAdapter))
	if raw == "" {
		return Adapter, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 0, &ConfigError{
			Field: EnvAdapter, Value: raw, Reason: "not a Go duration greater than zero",
			Allowed: []string{"a Go duration greater than zero, for example 5s; empty means 5s"},
		}
	}
	return d, nil
}

// Environment lists the adapter-bound setting for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{{
		Name:        EnvAdapter,
		Default:     envdoc.Tick(envdoc.SecondSpec(Adapter)),
		Description: "the limit on one adapter call when the caller set no shorter deadline. Sample mode uses the default and does not read this variable",
	}}
}

// boundKey is the context key of the duration Load returned.
type boundKey struct{}

// WithBound returns a context that carries the loaded adapter bound.
// Bound reads it. Sample mode does not call WithBound, so a sample call
// keeps Adapter and does not read the environment.
func WithBound(ctx context.Context, d time.Duration) context.Context {
	return context.WithValue(ctx, boundKey{}, d)
}

// Bound returns a context that ends at the loaded bound, or at Adapter when
// none was loaded, or at the parent's deadline, whichever is sooner.
//
// The caller must call the returned CancelFunc, as with context.WithTimeout.
func Bound(ctx context.Context) (context.Context, context.CancelFunc) {
	d := Adapter
	// A duration carried by the review is the setting loaded before the
	// review started. Zero means the caller did not load one.
	if v, ok := ctx.Value(boundKey{}).(time.Duration); ok && v > 0 {
		d = v
	}
	return context.WithTimeout(ctx, d)
}
