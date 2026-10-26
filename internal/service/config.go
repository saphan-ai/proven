package service

import (
	"fmt"
	"strings"
)

// ConfigError names a refused setting and the values that are allowed.
// Field is the setting, Value is what was refused, Reason says why, and
// Allowed lists what would be accepted. A nil receiver still prints, so a
// typed nil that reaches an error check never panics.
type ConfigError struct {
	Field   string
	Value   string
	Reason  string
	Allowed []string
	// Lead is the opening of the refusal sentence. Empty means
	// "configuration refused". A Markovian refusal sets it so that sentence
	// stays the one it already printed.
	Lead string
}

// Error returns one line: the opening, the refused field and value, the
// reason, and the allowed values, so a user can fix the setting without
// reading the code.
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	lead := e.Lead
	if lead == "" {
		lead = "configuration refused"
	}
	msg := fmt.Sprintf("%s: %s=%q", lead, e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}
