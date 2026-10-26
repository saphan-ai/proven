package markov

import (
	"regexp"
	"strings"
)

// The marker protocol (AGENTS.md section 3). The bracket forms are canonical; the @@@ and --- forms are the
// spec's backward-compatible detection markers.
const (
	// MarkerState ends a step that is not the last, followed by the state to carry.
	MarkerState = "[STATE_CHECKPOINT]"
	// MarkerFinal ends the last step.
	MarkerFinal = "[TASK_COMPLETE]"
)

// Markers is what one step's raw output said through the protocol.
type Markers struct {
	// State is the text after a state marker, trimmed; HasState is false when no state marker was present.
	State    string
	HasState bool
	// Done is true when a final marker is present.
	Done bool
	// Cleaned is the display text with every marker and its state removed.
	Cleaned string
}

// stateTails are the state markers, canonical first; the state is everything after the first one found.
var stateTails = []string{MarkerState, "@@@STATE@@@", "[STATE]", "---STATE---"}

// finalMarkers are the completion markers.
var finalMarkers = []string{MarkerFinal, "@@@FINAL@@@", "[FINAL]", "---FINAL---"}

// ParseMarkers reads the markers of one step's raw output.
func ParseMarkers(raw string) Markers {
	m := Markers{Cleaned: CleanMarkers(raw)}
	for _, mk := range stateTails {
		if i := markerIndex(raw, mk); i >= 0 {
			m.State = strings.TrimSpace(raw[i+len(mk):])
			// A state that runs on into the final marker carries the state only.
			for _, f := range finalMarkers {
				if j := markerIndex(m.State, f); j >= 0 {
					m.State = strings.TrimSpace(m.State[:j])
				}
			}
			m.HasState = true
			break
		}
	}
	for _, f := range finalMarkers {
		if markerIndex(raw, f) >= 0 {
			m.Done = true
			break
		}
	}
	return m
}

// markerIndex finds mk in s. The bare [STATE] and [FINAL] forms are not matched right after an underscore, as
// the reference rules, so a word ending in _ never reads as a marker.
func markerIndex(s, mk string) int {
	from := 0
	for {
		i := strings.Index(s[from:], mk)
		if i < 0 {
			return -1
		}
		i += from
		if (mk == "[STATE]" || mk == "[FINAL]") && i > 0 && s[i-1] == '_' {
			from = i + len(mk)
			continue
		}
		return i
	}
}

var (
	trailingSpace = regexp.MustCompile(`(?m)[ \t]+$`)
	blankRun      = regexp.MustCompile(`\n{3,}`)
)

// CleanMarkers strips every marker, and the state that follows a state marker, from display text (AGENTS.md
// section 3: the model's control text never reaches a reader).
func CleanMarkers(raw string) string {
	out := raw
	for _, mk := range []string{MarkerState, "@@@STATE@@@", "---STATE---"} {
		if i := strings.Index(out, mk); i >= 0 {
			out = out[:i]
		}
	}
	// The bare [STATE] tail is cut only where it starts a line, so prose that mentions it is kept.
	if i := strings.Index(out, "\n[STATE]"); i >= 0 {
		out = out[:i]
	}
	for _, f := range []string{MarkerFinal, "@@@FINAL@@@", "---FINAL---"} {
		out = strings.ReplaceAll(out, f, "")
	}
	if i := strings.Index(out, "\n[FINAL]"); i >= 0 {
		out = out[:i]
	}
	out = trailingSpace.ReplaceAllString(out, "")
	out = blankRun.ReplaceAllString(out, "\n\n")
	return strings.TrimSpace(out)
}
