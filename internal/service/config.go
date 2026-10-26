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
	// Sentence names which refusal line Error prints. Empty is the seam
	// line. SentenceTelemetry and SentenceRawQuote keep the lines a direct
	// caller of Setup and LoadTable already read. A copy made for the shell
	// leaves Sentence empty, so that line does not change.
	Sentence string
}

// SentenceTelemetry is the refusal line telemetry.Setup prints for a direct
// caller. The allowed list stays on the error and off this line.
const SentenceTelemetry = "telemetry"

// SentenceRawQuote is the refusal line osv.LoadTable prints for a direct
// caller. The value is wrapped in quotes, and a quote inside it is not escaped.
const SentenceRawQuote = "raw-quote"

// Error returns the refusal on one line. An empty Sentence prints the seam
// line: the opening, the field, the quoted value, the reason, and the allowed
// values. SentenceTelemetry and SentenceRawQuote print the lines a direct
// caller of Setup and LoadTable already read.
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	switch e.Sentence {
	case SentenceTelemetry:
		// The package line never grew an allowed list, so Allowed stays off it.
		return fmt.Sprintf("telemetry: %s=%q: %s", e.Field, e.Value, e.Reason)
	case SentenceRawQuote:
		msg := "configuration refused: " + e.Field + "=" + `"` + e.Value + `"`
		if e.Reason != "" {
			msg += ": " + e.Reason
		}
		if len(e.Allowed) > 0 {
			msg += "; allowed: " + strings.Join(e.Allowed, ", ")
		}
		return msg
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
