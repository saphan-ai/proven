package markov

import "strings"

// The two modes of AGENTS.md section 1.
const (
	ModeArchitect = "ARCHITECT"
	ModeResearch  = "RESEARCH"
)

// markerRules is the marker instruction every step ends with (AGENTS.md section 3).
const markerRules = "End every non-final chunk with [STATE_CHECKPOINT] followed by a concise state block:\n" +
	"COMPLETED: ...\nREMAINING: ...\nDECISIONS: ...\nCONTEXT: ...\n" +
	"End the final chunk with [TASK_COMPLETE]. Exactly one marker per chunk. Never narrate the markers."

// goalSummary is how much of the goal a continuation step repeats (AGENTS.md section 4.4: about 200 characters).
const goalSummary = 200

// Attachment is one named piece of context the first step carries.
type Attachment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// PromptInput is what one step's prompt is built from.
type PromptInput struct {
	Mode          string
	Goal          string
	Step          int
	Carryover     string
	Overlap       string
	MemoryContext string
	Attachments   []Attachment
}

// BuildPrompt builds one step's prompt (AGENTS.md section 4.4). The memory context and the attachments enter the
// first step only (section 5); a continuation step carries the shortened goal, the carried state and, when set,
// the overlap window, and nothing else of the steps before it.
func BuildPrompt(in PromptInput) string {
	var parts []string
	if in.Step == 1 {
		if mc := strings.TrimSpace(in.MemoryContext); mc != "" {
			parts = append(parts, mc, "")
		}
		if in.Mode == ModeResearch {
			parts = append(parts, "You are executing recursive research in bounded chunks. Log every action as '>> action'. "+
				"Produce multi-file style output where useful and keep a research-tree state snapshot.")
		} else {
			parts = append(parts, "You are executing a long build task in bounded chunks. Analyze the goal, produce a plan, and start implementation.")
		}
		parts = append(parts, "", "User Request: "+in.Goal)
		if len(in.Attachments) > 0 {
			parts = append(parts, "")
			for _, a := range in.Attachments {
				parts = append(parts, "[ATTACHMENT "+a.Name+"]\n"+a.Content)
			}
		}
	} else {
		parts = append(parts, "Continuing a chunked "+strings.ToLower(modeOrDefault(in.Mode))+" run. Goal (truncated): "+clipRunes(in.Goal, goalSummary))
		parts = append(parts, "", "[CARRYOVER STATE]\n"+in.Carryover)
		if in.Overlap != "" {
			parts = append(parts, "", "[OVERLAP]\n"+in.Overlap)
		}
		parts = append(parts, "", "Continue exactly where the state leaves off. Do not repeat completed work.")
	}
	parts = append(parts, "", markerRules)
	return strings.Join(parts, "\n")
}

// OverlapSlice is the overlap window: the last overlapTokens*4 characters of the previous step's cleaned output
// (AGENTS.md section 2.1). Zero overlap is an empty window.
func OverlapSlice(previousCleaned string, overlapTokens int) string {
	if overlapTokens <= 0 || previousCleaned == "" {
		return ""
	}
	n := overlapTokens * 4
	if len(previousCleaned) <= n {
		return previousCleaned
	}
	cut := len(previousCleaned) - n
	// The cut moves forward to a rune boundary, so the window is valid text.
	for cut < len(previousCleaned) && previousCleaned[cut]&0xC0 == 0x80 {
		cut++
	}
	return previousCleaned[cut:]
}

// modeOrDefault is the spec's default mode for an empty one.
func modeOrDefault(m string) string {
	if m == "" {
		return ModeArchitect
	}
	return m
}

// clipRunes keeps at most n bytes of s, cut back to a rune boundary.
func clipRunes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && s[n]&0xC0 == 0x80 {
		n--
	}
	return s[:n]
}
