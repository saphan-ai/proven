package markov

import (
	"regexp"
	"strings"
)

// The cascade levels of AGENTS.md section 4.3, in their strict order. LevelModel, the spec's model-based
// compression, runs only when the caller gives a compressor.
const (
	LevelMarker     = "marker"
	LevelModel      = "llm"
	LevelHeuristic  = "heuristic"
	LevelTruncation = "truncation"
)

// minMarkerState is the shortest state a model's own checkpoint may carry and still be accepted (AGENTS.md
// section 4.3, level 1): shorter text is a fragment, not a state.
const minMarkerState = 24

var (
	// markerArtifacts are runs of bracket, @ and dash characters a model leaves behind around a marker.
	markerArtifacts = regexp.MustCompile(`[<>\[\]@-]{3,}`)
	// keyPhrase captures the sentences the heuristic level keeps.
	keyPhrase = regexp.MustCompile(`(?i)(?:^|\n|\.\s+)((?:Therefore|Key insight|Status|Decision|Decided|Next|Remaining|Completed)[^.\n]*)`)
	spaces    = regexp.MustCompile(`\s+`)
)

// Compressor is the spec's injected compressorFn: one model call that compresses the state.
type Compressor func(prompt, system string) (string, error)

// NextCarryover runs the compression cascade after one step and returns the state to carry and the level that
// produced it. Each level falls through on failure, so the chain survives an unstable model.
func NextCarryover(markerState string, hasState bool, previous, chunkText string, budget int, compress Compressor) (string, string) {
	// Level 1: the model's own checkpoint, accepted only when long enough and within the budget.
	if hasState {
		s := strings.TrimSpace(markerArtifacts.ReplaceAllString(markerState, ""))
		if len(s) >= minMarkerState && EstimateTokens(s) <= budget {
			return s, LevelMarker
		}
	}
	// Level 2: model compression over the previous state and the tail of this step.
	if compress != nil {
		tail := chunkText
		if len(tail) > 2000 {
			tail = tail[len(tail)-2000:]
		}
		prev := previous
		if prev == "" {
			prev = "(none)"
		}
		prompt := "Previous state:\n" + prev + "\n\nLatest chunk (tail):\n" + tail + "\n\n" +
			"Compress into 3-5 critical points as a labeled block:\nCOMPLETED: ...\nREMAINING: ...\nDECISIONS: ...\nCONTEXT: ..."
		if out, err := compress(prompt, "You compress reasoning state. Be terse and lossless on decisions."); err == nil {
			out = strings.TrimSpace(out)
			if len(out) >= minMarkerState && EstimateTokens(out) <= budget {
				return out, LevelModel
			}
		}
	}
	// Level 3: the key phrases of this step, at most five.
	var phrases []string
	for _, m := range keyPhrase.FindAllStringSubmatch(chunkText, -1) {
		phrases = append(phrases, strings.TrimSpace(m[1]))
		if len(phrases) == 5 {
			break
		}
	}
	if len(phrases) > 0 {
		if s := strings.Join(phrases, "; "); EstimateTokens(s) <= budget {
			return s, LevelHeuristic
		}
	}
	// Level 4: the last words of the step, as many as the budget holds.
	src := chunkText
	if strings.TrimSpace(src) == "" {
		src = previous
	}
	words := strings.Fields(spaces.ReplaceAllString(src, " "))
	tail := ""
	for i := len(words) - 1; i >= 0; i-- {
		next := words[i]
		if tail != "" {
			next += " " + tail
		}
		if EstimateTokens(next) > budget {
			break
		}
		tail = next
	}
	return tail, LevelTruncation
}
