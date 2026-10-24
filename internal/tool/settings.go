package tool

import (
	"fmt"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
)

const (
	// EnvTimeout is the limit for one side of a tool run, base or change.
	// Sample mode does not read it. Empty means DefaultTimeout.
	EnvTimeout = "PROVEN_TOOL_TIMEOUT"
)

// DefaultTimeout is the limit for one side when EnvTimeout is empty.
// It matches the Prover's own default for one test side: a race run is
// not given a longer leash than a test unless a setting says so.
const DefaultTimeout = 2 * time.Minute

// ConfigError names a refused tool setting and the values that are allowed (ADR-0013).
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the setting, the refused value, why, and what is allowed.
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := fmt.Sprintf("configuration refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// Load reads EnvTimeout. An empty value is DefaultTimeout.
// A value that is set and is not a duration greater than zero is refused
// and is never replaced by the default (ADR-0013).
func Load(env func(string) string) (time.Duration, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvTimeout))
	if raw == "" {
		return DefaultTimeout, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 0, &ConfigError{
			Field:   EnvTimeout,
			Value:   raw,
			Reason:  "not a Go duration greater than zero",
			Allowed: []string{"a Go duration greater than zero, for example 120s; empty means 120s"},
		}
	}
	return d, nil
}

// Environment lists the tool settings for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{{
		Name:        EnvTimeout,
		Default:     envdoc.Tick(envdoc.SecondSpec(DefaultTimeout)),
		Description: "the limit for one side of a deterministic tool run, base or change. A run past it is named and is not evidence. Sample mode uses the default and does not read this variable",
	}}
}
