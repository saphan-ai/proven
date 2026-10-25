package scout

// This file is the Guard stage (the spec's docs/risk-and-policy.md). Of the five layers, three are
// built: the injection check (known patterns, role drift, out-of-band tool references), the policy
// gate on the registry's roles and surfaces, and a sensitivity layer that redacts e-mail addresses
// and long digit runs. The scope check over session history and the rate limit need state across
// requests, which a one-request decision does not have, so they are named on every risk as not run.

import (
	"regexp"
	"strings"
)

// Risk is the guard stack's one answer: a level and the flags that led there.
type Risk struct {
	Level             string      `json:"level"`
	Flags             []string    `json:"flags"`
	Redactions        []Redaction `json:"redactions"`
	Reason            string      `json:"reason,omitempty"`
	ApprovalsRequired []string    `json:"approvalsRequired,omitempty"`
	Injection         *Injection  `json:"injection,omitempty"`
	// NotRun names the layers this implementation does not run, so a pass is never read as their pass.
	NotRun []string `json:"notRun"`
}

// Injection is the injection layer's finding. Evidence names the patterns that matched, never the
// payload: the spec keeps a detected payload out of the log by default.
type Injection struct {
	Detected bool     `json:"detected"`
	Kind     string   `json:"kind"`
	Score    float64  `json:"score"`
	Evidence []string `json:"evidence"`
}

// Redaction is one span the sensitivity layer replaced, by data class.
type Redaction struct {
	Class string `json:"class"`
	Count int    `json:"count"`
}

// The risk levels, lowest first, and "block".
const (
	LevelLow    = "low"
	LevelMedium = "medium"
	LevelHigh   = "high"
	LevelBlock  = "block"
)

// injectionPattern is one known injection shape and the kind the spec files it under.
type injectionPattern struct {
	name string
	kind string
	re   *regexp.Regexp
}

// injectionPatterns are the known-pattern layer. Each is a sentence shape that asks the system to drop
// its instructions, to take another role, or to hand over what it guards; none matches a word alone,
// because the spec sets a tight precision target and a safety check that blocks plain work gets disabled.
var injectionPatterns = []injectionPattern{
	{"instruction-override", "pattern", regexp.MustCompile(`(?i)\b(ignore|disregard|forget|override)\b[^.\n]{0,30}\b(previous|prior|above|earlier|preceding|all|the)\b[^.\n]{0,20}\b(instructions?|rules|prompts?|directions|guidelines|user)\b`)},
	{"instruction-override", "pattern", regexp.MustCompile(`(?i)\b(ignore|disregard)\s+(previous|prior|above|everything)\b`)},
	{"role-marker", "role-drift", regexp.MustCompile(`(?i)</?\s*(system|assistant|user)\s*>|\[/?INST\]|<<\s*/?SYS\s*>>`)},
	{"role-drift", "role-drift", regexp.MustCompile(`(?i)\b(you are now|from now on you are|pretend (to be|you are)|act as)\b[^.\n]{0,30}\b(system|admin|administrator|root|unrestricted|developer|jailbroken)\b`)},
	{"mode-switch", "role-drift", regexp.MustCompile(`(?i)\b(unrestricted|developer|jailbreak|god)\s+mode\b`)},
	{"secret-request", "pattern", regexp.MustCompile(`(?i)\b(reveal|print|dump|leak|exfiltrate|show me|list)\b[^.\n]{0,30}\b(api keys?|credentials|passwords?|secrets|private keys?|access tokens?)\b`)},
	{"privilege-grant", "pattern", regexp.MustCompile(`(?i)\bgrant\s+(me\s+|us\s+)?(admin|administrator|root|superuser)\b`)},
}

// toolRefRE finds a request to use a named tool, spelled as a tool identifier is: words joined by
// underscores. Whether the tool is out of band is decided against the intent's capabilities.
var toolRefRE = regexp.MustCompile(`(?i)\b(use|call|invoke|run)\s+(the\s+)?([a-z][a-z0-9]*(_[a-z0-9]+)+)\b`)

// controlRE finds the invisible and direction-changing characters that hide text from a reader, both
// as the characters themselves and as their written escapes.
var controlRE = regexp.MustCompile(`[\x{200b}-\x{200f}\x{202a}-\x{202e}\x{2066}-\x{2069}\x{feff}]|(?i)\\u(200[b-f]|202[a-e]|206[6-9]|feff)`)

// sensitive data classes the redaction layer replaces.
var (
	emailRE  = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	digitsRE = regexp.MustCompile(`\b\d{12,19}\b`)
)

// guard runs the stack for one prompt and the chosen intent. in may be nil when no intent was chosen.
func guard(prompt string, in *Intent, user User, surface string) Risk {
	r := Risk{Level: LevelLow, Flags: []string{}, Redactions: []Redaction{}, NotRun: []string{"scope", "rate-limit"}}
	// Spec: the intent's risk class is the floor; risk is re-evaluated per request but never below it.
	if in != nil {
		r.Level = in.RiskClass
	}
	if inj := injection(prompt, in); inj.Detected {
		r.Injection = &inj
		r.Level = LevelBlock
		r.Flags = append(r.Flags, "injection:"+inj.Kind)
		r.Reason = "prompt injection: " + strings.Join(inj.Evidence, ", ")
		return r
	}
	if in != nil && in.ID != GeneralIntent {
		// Spec: surfaces and roles are enforced at Guard; a call outside them blocks.
		if !allows(in.Roles, user.Role) {
			r.Level = LevelBlock
			r.Flags = append(r.Flags, "policy:role")
			r.Reason = "role " + user.Role + " may not invoke " + in.ID
			return r
		}
		if !allows(in.Surfaces, surface) {
			r.Level = LevelBlock
			r.Flags = append(r.Flags, "policy:surface")
			r.Reason = "surface " + surface + " may not invoke " + in.ID
			return r
		}
	}
	if n := len(emailRE.FindAllString(prompt, -1)); n > 0 {
		r.Redactions = append(r.Redactions, Redaction{Class: "email", Count: n})
	}
	if n := len(digitsRE.FindAllString(prompt, -1)); n > 0 {
		r.Redactions = append(r.Redactions, Redaction{Class: "long-number", Count: n})
	}
	// Spec: a warning lifts the level by one band. Sensitive data in the prompt is that warning.
	if len(r.Redactions) > 0 {
		r.Flags = append(r.Flags, "sensitive:redacted")
		r.Level = lift(r.Level)
	}
	// Spec: a high-risk intent at medium or high level carries an approval requirement.
	if in != nil && in.RiskClass == LevelHigh && (r.Level == LevelMedium || r.Level == LevelHigh) {
		r.ApprovalsRequired = []string{"user-confirm"}
	}
	return r
}

// injection runs the injection layer: known patterns, hidden control characters and out-of-band tool
// references. The score is the share of the layer's checks that fired, so two signs score higher than one.
func injection(prompt string, in *Intent) Injection {
	var ev []string
	kind := ""
	for _, p := range injectionPatterns {
		if p.re.MatchString(prompt) && !containsString(ev, p.name) {
			ev = append(ev, p.name)
			if kind == "" {
				kind = p.kind
			}
		}
	}
	if controlRE.MatchString(prompt) {
		ev = append(ev, "hidden-control-character")
		if kind == "" {
			kind = "pattern"
		}
	}
	for _, m := range toolRefRE.FindAllStringSubmatch(prompt, -1) {
		tool := strings.ToLower(m[3])
		if in == nil || !containsString(in.Capabilities, tool) {
			ev = append(ev, "out-of-band-tool:"+tool)
			if kind == "" {
				kind = "tool-ref"
			}
		}
	}
	if len(ev) == 0 {
		return Injection{Evidence: []string{}}
	}
	checks := float64(len(injectionPatterns) + 2)
	score := float64(len(ev)) / checks
	if score > 1 {
		score = 1
	}
	return Injection{Detected: true, Kind: kind, Score: round3(score), Evidence: ev}
}

// lift raises a level by one band; high stays high.
func lift(level string) string {
	switch level {
	case LevelLow:
		return LevelMedium
	case LevelMedium:
		return LevelHigh
	}
	return level
}

func containsString(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
