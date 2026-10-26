// Package scout is Meterless's Scout Intent in Go, built to the specification Meterless publishes and
// checked with its eval harness (ADR-0049 point 4). It decides before anything acts: Sense scores every
// registered intent, Interpret binds parameters and decides whether to ask, Guard runs the risk and
// policy stack, Route turns the intent's capabilities into a tool plan, and Recommend signs the
// execution contract that is the only thing a downstream engine accepts.
//
// Proven uses it on a change rather than a prompt: a change registry (change.go) names the six kinds
// of change ADR-0049 point 1 lists, and the class it picks chooses the review plan (internal/review).
// Scout never runs a model: Stage 2 of the spec's scoring is optional and is not built, so every
// decision here is the deterministic Stage-1 scan, the same on every run.
package scout

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Options tune an Engine. The zero value of each field is the spec's default.
type Options struct {
	Bands Bands
	// TieWindow is how close the top two scores must be to ask which was meant (spec: 0.05).
	TieWindow float64
	// HighRiskProceed is the confidence a high-risk intent needs to go ahead without asking (spec: 0.95).
	HighRiskProceed float64
	// SecondaryFloor is the score a second intent of a multi-intent prompt needs (spec: 0.55).
	SecondaryFloor float64
	// TTL is how long a contract is valid (spec: five minutes).
	TTL time.Duration
	// PolicyPack names the packs stamped on the contract; the guard layers are the same for every pack.
	PolicyPack []string
	// SigningKey signs contracts; empty means a random key for this process.
	SigningKey []byte
	// Now is the clock; nil means time.Now.
	Now func() time.Time
}

// Engine is one Scout over one registry.
type Engine struct {
	reg    *Registry
	opts   Options
	text   *textIndex
	signer *Signer
}

// New returns an Engine over reg with opts, the spec's defaults filling every zero field.
func New(reg *Registry, opts Options) (*Engine, error) {
	if reg == nil {
		return nil, fmt.Errorf("scout: no registry")
	}
	if opts.Bands == (Bands{}) {
		opts.Bands = DefaultBands()
		if reg.Bands != nil {
			opts.Bands = *reg.Bands
		}
	}
	if opts.TieWindow == 0 {
		opts.TieWindow = 0.05
	}
	if opts.HighRiskProceed == 0 {
		opts.HighRiskProceed = 0.95
	}
	if opts.SecondaryFloor == 0 {
		opts.SecondaryFloor = 0.55
	}
	if opts.TTL == 0 {
		opts.TTL = 5 * time.Minute
	}
	if len(opts.PolicyPack) == 0 {
		opts.PolicyPack = []string{"default"}
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	signer, err := NewSigner(opts.SigningKey)
	if err != nil {
		return nil, err
	}
	e := &Engine{reg: reg, opts: opts, signer: signer}
	if reg.Input == InputText {
		e.text = newTextIndex(reg)
	}
	return e, nil
}

// Registry returns the engine's registry.
func (e *Engine) Registry() *Registry { return e.reg }

// Signer returns the engine's contract signer, so a caller can verify what it issued.
func (e *Engine) Signer() *Signer { return e.signer }

// User is who asks: an id and a role the registry's roles are checked against.
type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// Input is one request: a prompt (for a change registry, the changed files, one per line), who asks
// and on which surface.
type Input struct {
	Prompt  string `json:"prompt"`
	User    User   `json:"user"`
	Surface string `json:"surface"`
}

// Clarification is a structured question back to the surface, with at most four options, each naming
// an intent. Trigger is which of the spec's four rules asked it.
type Clarification struct {
	Question string   `json:"question"`
	Options  []Option `json:"options"`
	Trigger  string   `json:"trigger"`
}

// Option is one answer a clarification offers.
type Option struct {
	Label  string `json:"label"`
	Intent string `json:"intent"`
}

// Ambiguity says whether the decision is unsure and why. A decision with a clarification, or one that
// fell back to GENERAL, is ambiguous: no confident intent was chosen.
type Ambiguity struct {
	Ambiguous bool     `json:"ambiguous"`
	Reasons   []string `json:"reasons"`
}

// Decision is everything Scout decided for one Input. ExecutionContract is nil when nothing may act:
// the request was blocked, or Scout asked a question it needs answered first.
type Decision struct {
	TraceID           string         `json:"traceId"`
	Candidates        []Candidate    `json:"candidates"`
	Intent            string         `json:"intent"`
	Intents           []string       `json:"intents,omitempty"`
	Confidence        float64        `json:"confidence"`
	Band              string         `json:"band"`
	Signals           []Signal       `json:"signals"`
	Unexplained       []string       `json:"unexplained"`
	Ambiguity         Ambiguity      `json:"ambiguity"`
	Clarification     *Clarification `json:"clarification,omitempty"`
	Parameters        map[string]any `json:"parameters"`
	Risk              Risk           `json:"risk"`
	ToolPlan          []ToolStep     `json:"toolPlan"`
	ExecutionContract *Contract      `json:"executionContract,omitempty"`
}

// Decide runs the five stages on one input.
func (e *Engine) Decide(in Input) (Decision, error) {
	now := e.opts.Now()
	trace, err := newULID(now)
	if err != nil {
		return Decision{}, err
	}
	if in.Surface == "" {
		in.Surface = "chat"
	}
	d := Decision{TraceID: trace, Signals: []Signal{}, Unexplained: []string{}, Parameters: map[string]any{}, ToolPlan: []ToolStep{}}
	d.Ambiguity.Reasons = []string{}

	// Sense.
	s := e.sense(in.Prompt)
	d.Candidates = s.candidates
	if d.Candidates == nil {
		d.Candidates = []Candidate{}
	}
	if s.unexplained != nil {
		d.Unexplained = s.unexplained
	}
	top, second := Candidate{Intent: GeneralIntent}, Candidate{}
	if len(s.candidates) > 0 {
		top = s.candidates[0]
	}
	if len(s.candidates) > 1 {
		second = s.candidates[1]
	}
	d.Confidence = top.Score
	d.Band = e.opts.Bands.Of(top.Score)
	// Nothing matched at all: that is the Reject band whatever the edges are, because a registry whose
	// Low edge is zero would otherwise ask which of no intents was meant.
	if len(s.candidates) == 0 {
		d.Band = BandReject
	}
	d.Intent = top.Intent
	for _, sig := range s.signals {
		if sig.Intent == top.Intent {
			d.Signals = append(d.Signals, sig)
		}
	}

	// Interpret: multi-intent first, then the four clarification triggers.
	if e.reg.Input == InputText {
		if parts := e.multi(in.Prompt); len(parts) > 1 {
			d.Intents = make([]string, 0, len(parts))
			for _, c := range parts {
				d.Intents = append(d.Intents, c.Intent)
			}
			d.Candidates = append(parts, without(d.Candidates, parts)...)
			top = parts[0]
			d.Intent, d.Confidence, d.Band = top.Intent, top.Score, e.opts.Bands.Of(top.Score)
		}
	}
	chosen, _ := e.reg.intent(d.Intent)
	tied := second.Intent != "" && top.Score-second.Score < e.opts.TieWindow
	switch {
	case d.Band == BandLow:
		d.Clarification = e.ask("low-band", "Which of these did you mean?", d.Candidates)
	case tied && (d.Band == BandMedium || d.Band == BandReject):
		// Spec rule 2 asks on a tie in the Medium band. A tie below it asks too: falling back to GENERAL
		// would drop an intent the prompt named, and the spec never wants a guess in its place.
		d.Clarification = e.ask("tie", "Did you mean "+top.Intent+" or "+second.Intent+"?", d.Candidates)
	case e.reg.Input == InputText && unresolvedReference(in.Prompt):
		d.Clarification = e.ask("params", "Which earlier item do you mean?", d.Candidates)
	case chosen != nil && chosen.RiskClass == LevelHigh && d.Confidence < e.opts.HighRiskProceed && d.Band != BandReject:
		d.Clarification = &Clarification{Question: "Confirm: " + chosen.Description, Options: []Option{{Label: "Yes, go ahead", Intent: chosen.ID}}, Trigger: "high-risk"}
	}
	if d.Band == BandReject && d.Clarification == nil {
		// Spec: below the Reject edge the decision falls back to the safe GENERAL intent, never a guess.
		d.Intent = GeneralIntent
		// GENERAL is listed when nothing else scored, so the candidates always hold the intent chosen.
		if !hasIntent(d.Candidates, GeneralIntent) && len(d.Candidates) == 0 {
			d.Candidates = []Candidate{{Intent: GeneralIntent, Score: 0}}
		}
		d.Ambiguity.Ambiguous = true
		d.Ambiguity.Reasons = append(d.Ambiguity.Reasons, fmt.Sprintf("no intent clears the reject edge %.2f (top %s %.3f)", e.opts.Bands.Low, top.Intent, top.Score))
		chosen, _ = e.reg.intent(GeneralIntent)
	}
	if d.Clarification != nil && d.Clarification.Trigger != "high-risk" {
		d.Ambiguity.Ambiguous = true
		d.Ambiguity.Reasons = append(d.Ambiguity.Reasons, "clarification: "+d.Clarification.Trigger)
	}
	if len(s.unexplained) > 0 && e.reg.Input == InputChange {
		d.Ambiguity.Reasons = append(d.Ambiguity.Reasons, fmt.Sprintf("%d file(s) match no pattern", len(s.unexplained)))
	}
	if chosen != nil {
		d.Parameters = bindParameters(in.Prompt, chosen)
	}

	// Guard.
	d.Risk = guard(in.Prompt, chosen, in.User, in.Surface)

	// Route: the intent's capabilities, in the order the registry declares them.
	if chosen != nil && d.Risk.Level != LevelBlock {
		for _, c := range chosen.Capabilities {
			d.ToolPlan = append(d.ToolPlan, ToolStep{Tool: c, Capability: c, Input: map[string]any{}})
		}
	}

	// Recommend: a contract only when something may act on it. A blocked request has none; a question
	// that must be answered first has none, except the high-risk confirmation, whose contract carries the
	// approval it waits for (spec: high-risk contracts carry approval records).
	if d.Risk.Level == LevelBlock || (d.Clarification != nil && d.Clarification.Trigger != "high-risk") {
		return d, nil
	}
	if d.Clarification != nil && len(d.Risk.ApprovalsRequired) == 0 {
		d.Risk.ApprovalsRequired = []string{"user-confirm"}
	}
	c, err := e.contract(now, trace, d, chosen, in)
	if err != nil {
		return Decision{}, err
	}
	d.ExecutionContract = c
	return d, nil
}

// sense runs the Stage-1 scan of the registry's input kind.
func (e *Engine) sense(prompt string) sensed {
	if e.reg.Input == InputChange {
		return scanChange(e.reg, prompt)
	}
	return e.text.scan(prompt)
}

// partSplit splits a prompt into the parts a multi-intent request joins: commas and "and".
var partSplit = regexp.MustCompile(`(?i)\s*,\s*(?:and\s+)?|\s+and\s+`)

// multi returns one candidate per part when a prompt holds two or more intents. Spec: two candidates
// both in the High band on disjoint spans make a plan of both; later parts join at the Medium floor.
func (e *Engine) multi(prompt string) []Candidate {
	parts := partSplit.Split(prompt, -1)
	if len(parts) < 2 {
		return nil
	}
	var out []Candidate
	highs := 0
	for _, p := range parts {
		s := e.text.scan(p)
		if len(s.candidates) == 0 {
			continue
		}
		c := s.candidates[0]
		if c.Score < e.opts.SecondaryFloor || hasIntent(out, c.Intent) {
			continue
		}
		if c.Score >= e.opts.Bands.High {
			highs++
		}
		c.Span = strings.TrimSpace(p)
		out = append(out, c)
	}
	if highs < 2 || len(out) < 2 {
		return nil
	}
	return out
}

// referenceRE finds a reference to something the prompt does not name: "the thing we talked about".
// Spec rule 3: a reference that cannot be bound asks rather than guesses.
var referenceRE = regexp.MustCompile(`(?i)\b(the|that|this) (thing|stuff|one|matter)\b|\b(we|you|i) (talked|spoke|discussed|mentioned)\b|\bas discussed\b|\bthe usual\b`)

// unresolvedReference reports whether the prompt leans on a reference with nothing to bind it to.
func unresolvedReference(prompt string) bool { return referenceRE.MatchString(prompt) }

// ask builds a clarification offering the top candidates, at most four (spec).
func (e *Engine) ask(trigger, question string, cands []Candidate) *Clarification {
	c := &Clarification{Question: question, Trigger: trigger, Options: []Option{}}
	for _, cand := range cands {
		if len(c.Options) == 4 {
			break
		}
		c.Options = append(c.Options, Option{Label: cand.Intent, Intent: cand.Intent})
	}
	return c
}

// contract builds and signs the execution contract.
func (e *Engine) contract(now time.Time, trace string, d Decision, in *Intent, req Input) (*Contract, error) {
	id, err := newULID(now)
	if err != nil {
		return nil, err
	}
	alts := []Candidate{}
	for _, c := range d.Candidates {
		if c.Intent != d.Intent {
			alts = append(alts, c)
		}
	}
	ci := ContractIntent{
		Primary: IntentRef{ID: d.Intent}, Parameters: d.Parameters, References: map[string]any{},
		Confidence: d.Confidence, Alternates: alts,
	}
	for _, s := range d.Intents[min(1, len(d.Intents)):] {
		ci.Secondary = append(ci.Secondary, IntentRef{ID: s})
	}
	caps := []string{}
	if in != nil {
		caps = append(caps, in.Capabilities...)
	}
	c := &Contract{
		ContractID: id, TraceID: trace, Intent: ci, Risk: d.Risk, ToolPlan: d.ToolPlan,
		// No model profiles are configured: Scout itself runs no model, and the plan it feeds names its own.
		ModelProfile:      "none",
		Scope:             Scope{User: req.User.ID, Surface: req.Surface, Capabilities: caps},
		IssuedAt:          now.UTC().Format(time.RFC3339Nano),
		ExpiresAt:         now.Add(e.opts.TTL).UTC().Format(time.RFC3339Nano),
		RegistryVersion:   e.reg.Version,
		PolicyPackVersion: strings.Join(e.opts.PolicyPack, "+") + "@1",
	}
	if err := e.signer.Sign(c); err != nil {
		return nil, err
	}
	return c, nil
}

// numberRE finds a count or a value in a prompt: "5", "10", "$100k", "280k".
var numberRE = regexp.MustCompile(`(?i)\$?(\d+(?:\.\d+)?)\s*([km])?\b`)

// bindParameters binds what Stage 1 can bind without a model: an integer parameter from the first
// number, a minimum value from "over $N", and an enum parameter from a value the prompt names.
// Anything else stays unbound; a parameter is never guessed.
func bindParameters(prompt string, in *Intent) map[string]any {
	out := map[string]any{}
	if len(in.Parameters) == 0 {
		return out
	}
	var schema struct {
		Properties map[string]struct {
			Type string   `json:"type"`
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(in.Parameters, &schema); err != nil {
		return out
	}
	low := strings.ToLower(prompt)
	ws := map[string]bool{}
	for _, w := range words(prompt) {
		ws[w] = true
	}
	for name, p := range schema.Properties {
		switch {
		case len(p.Enum) > 0:
			for _, v := range p.Enum {
				if ws[stem(strings.ToLower(v))] {
					out[name] = v
					break
				}
			}
		case p.Type == "integer":
			if m := numberRE.FindStringSubmatch(prompt); m != nil && m[2] == "" && !strings.HasPrefix(m[0], "$") {
				if n, err := strconv.Atoi(m[1]); err == nil {
					out[name] = n
				}
			}
		case p.Type == "number" && strings.Contains(low, "over"):
			if m := numberRE.FindStringSubmatch(prompt[strings.Index(low, "over"):]); m != nil {
				if v, err := strconv.ParseFloat(m[1], 64); err == nil {
					switch strings.ToLower(m[2]) {
					case "k":
						v *= 1000
					case "m":
						v *= 1000000
					}
					out[name] = v
				}
			}
		}
	}
	return out
}

// without returns cands less the intents already in parts.
func without(cands, parts []Candidate) []Candidate {
	var out []Candidate
	for _, c := range cands {
		if !hasIntent(parts, c.Intent) {
			out = append(out, c)
		}
	}
	return out
}

// hasIntent reports whether cands names intent.
func hasIntent(cands []Candidate, intent string) bool {
	for _, c := range cands {
		if c.Intent == intent {
			return true
		}
	}
	return false
}
