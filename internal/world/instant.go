package world

import (
	"errors"
	"fmt"
	"time"
)

// ErrTimestamp is returned, wrapped, when a caller's timestamp is not an
// RFC 3339 instant. The World Model refuses it by name rather than comparing
// it as text, because text order is not time order once offsets differ.
var ErrTimestamp = errors.New("world: timestamp is not an RFC 3339 instant")

// ParseInstant parses an RFC 3339 timestamp, with or without fractional
// seconds and with any offset, and returns its canonical form: UTC, with
// millisecond digits (or nine digits when the instant has finer detail).
// Two spellings of one instant give one canonical string, so an edge's
// content address and its stored bounds do not depend on the writer's offset.
func ParseInstant(s string) (string, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrTimestamp, s)
	}
	return canonicalInstant(t), nil
}

// canonicalInstant renders t in the one stored form.
func canonicalInstant(t time.Time) string {
	t = t.UTC()
	if t.Nanosecond()%int(time.Millisecond) == 0 {
		return t.Format("2006-01-02T15:04:05.000Z")
	}
	return t.Format("2006-01-02T15:04:05.000000000Z")
}

// instantOf parses a stored or caller timestamp for comparison. An empty
// string is not an instant; the caller decides what empty means.
func instantOf(s string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339Nano, s)
	return t, err == nil
}
