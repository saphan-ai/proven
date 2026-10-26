package scout

// This file holds the intent registry (the spec's docs/intent-registry.md): its record shape, the two
// loaders, and the checks a registry must pass before any prompt is scored against it.

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// GeneralIntent is the id of the safe fallback intent. The spec sends a prompt in the Reject band to it
// and never to a guessed intent.
const GeneralIntent = "GENERAL"

// The two kinds of input a registry can describe. A text registry scores a prompt by its words, as the
// spec's examples do. A change registry scores a list of changed files, one per line, by their paths:
// it is how Proven gives Scout a change to classify (ADR-0049 point 1).
const (
	InputText   = "text"
	InputChange = "change"
)

// Registry is one loaded intent registry.
type Registry struct {
	// Version is the registry version stamped on every contract, so a downstream engine sees drift.
	Version int `json:"version"`
	// Input is InputText or InputChange. Empty reads as InputText, the spec's own kind.
	Input string `json:"input,omitempty"`
	// Bands, when set, are this registry's band edges. The spec makes bands configurable per surface;
	// a change registry tightens them, because a review plan may only be cut for a change that is
	// plainly one kind.
	Bands   *Bands   `json:"bands,omitempty"`
	Note    string   `json:"note,omitempty"`
	Intents []Intent `json:"intents"`
}

// Intent is one record of the registry, in the spec's shape. Triggers is an addition of this
// implementation: weighted phrases or path patterns that Stage 1 scans for besides the examples.
type Intent struct {
	ID           string          `json:"id"`
	Category     string          `json:"category"`
	Description  string          `json:"description"`
	Parameters   json.RawMessage `json:"parameters,omitempty"`
	Capabilities []string        `json:"capabilities"`
	RiskClass    string          `json:"riskClass"`
	Surfaces     []string        `json:"surfaces"`
	Roles        []string        `json:"roles"`
	Examples     []string        `json:"examples"`
	Version      int             `json:"version"`
	Triggers     []Trigger       `json:"triggers,omitempty"`
}

// Trigger is one weighted signal of an intent. Phrase is matched in a text prompt; Path is a pattern
// matched against one changed file, with ** for any number of directories. Status, when set, limits a
// path trigger to files with one of those statuses (A added, M modified, D deleted, R renamed).
type Trigger struct {
	Phrase string  `json:"phrase,omitempty"`
	Path   string  `json:"path,omitempty"`
	Status string  `json:"status,omitempty"`
	Weight float64 `json:"weight"`
}

// idGrammar is the spec's category.name grammar for an intent id.
var idGrammar = regexp.MustCompile(`^[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$`)

// LoadRegistry reads one registry file and checks it.
func LoadRegistry(path string) (*Registry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseRegistry(b)
}

// ParseRegistry decodes a registry and checks it. A registry that fails a check is refused whole: an
// intent that cannot be classified or evaluated must not be half-loaded (spec: examples are mandatory).
func ParseRegistry(b []byte) (*Registry, error) {
	var r Registry
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("scout: registry: %w", err)
	}
	if err := r.validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

// validate applies the spec's registry rules and this implementation's trigger rules.
func (r *Registry) validate() error {
	if r.Input == "" {
		r.Input = InputText
	}
	if r.Input != InputText && r.Input != InputChange {
		return fmt.Errorf("scout: registry: input %q is not %s or %s", r.Input, InputText, InputChange)
	}
	if b := r.Bands; b != nil && !(0 <= b.Low && b.Low <= b.Medium && b.Medium <= b.High && b.High <= 1) {
		return fmt.Errorf("scout: registry: bands must rise from low to high within 0 and 1")
	}
	if len(r.Intents) == 0 {
		return fmt.Errorf("scout: registry: no intents")
	}
	seen := map[string]bool{}
	general := false
	for i := range r.Intents {
		in := &r.Intents[i]
		if in.ID == GeneralIntent {
			general = true
		} else if !idGrammar.MatchString(in.ID) {
			return fmt.Errorf("scout: registry: intent id %q is not category.name", in.ID)
		}
		if seen[in.ID] {
			return fmt.Errorf("scout: registry: intent %q is listed twice", in.ID)
		}
		seen[in.ID] = true
		// Spec: an intent without examples cannot be classified or evaluated.
		if len(in.Examples) == 0 {
			return fmt.Errorf("scout: registry: intent %q has no examples", in.ID)
		}
		switch in.RiskClass {
		case "low", "medium", "high":
		default:
			return fmt.Errorf("scout: registry: intent %q risk class %q is not low, medium or high", in.ID, in.RiskClass)
		}
		for _, t := range in.Triggers {
			if err := t.check(r.Input); err != nil {
				return fmt.Errorf("scout: registry: intent %q: %w", in.ID, err)
			}
		}
	}
	// Spec: a prompt in the Reject band falls back to GENERAL, so a registry without it has nowhere to send one.
	if !general {
		return fmt.Errorf("scout: registry: no %s intent", GeneralIntent)
	}
	return nil
}

// check refuses a trigger that the registry's input kind could never match, so a typo is a load error
// rather than a signal that silently scores nothing.
func (t Trigger) check(input string) error {
	if t.Weight <= 0 {
		return fmt.Errorf("trigger weight %v is not above zero", t.Weight)
	}
	switch input {
	case InputText:
		if t.Phrase == "" || t.Path != "" {
			return fmt.Errorf("a text registry's trigger names a phrase and no path")
		}
	case InputChange:
		if t.Path == "" || t.Phrase != "" {
			return fmt.Errorf("a change registry's trigger names a path and no phrase")
		}
		if err := validGlob(t.Path); err != nil {
			return fmt.Errorf("trigger path %q: %w", t.Path, err)
		}
		for _, s := range strings.Split(t.Status, "") {
			if !strings.Contains("AMDR", s) {
				return fmt.Errorf("trigger status %q is not made of A, M, D and R", t.Status)
			}
		}
	}
	return nil
}

// intent returns the intent with id.
func (r *Registry) intent(id string) (*Intent, bool) {
	for i := range r.Intents {
		if r.Intents[i].ID == id {
			return &r.Intents[i], true
		}
	}
	return nil, false
}

// allows reports whether list admits value; "*" admits every value.
func allows(list []string, value string) bool {
	for _, v := range list {
		if v == "*" || v == value {
			return true
		}
	}
	return false
}
