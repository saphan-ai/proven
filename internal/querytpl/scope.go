package querytpl

import (
	"strings"

	"github.com/saphan-ai/proven/internal/source"
)

// The roles that can ask a store a question. The Planner, the Prover, the
// Judge, the Writer and the Fixer have no query path, so they have no scope.
const (
	RoleInvestigator = "investigator"
	RoleSkeptic      = "skeptic"
)

// ReasonOutOfScope is the refusal reason for a template that is on the list
// but that the asking role's evidence scope does not allow (ADR-0013).
const ReasonOutOfScope = "query template is outside the evidence scope of this role"

// Scope is the set of template names one role may ask. The nil Scope allows
// every template, which is what a review with no setting gets. A Scope that is
// set holds only names on the list: the settings loader refuses any other name
// before a review starts.
type Scope []string

// Allows reports whether the role may ask the template of that name.
func (s Scope) Allows(name string) bool {
	if s == nil {
		return true
	}
	for _, n := range s {
		if n == name {
			return true
		}
	}
	return false
}

// Catalogue is the templates of the registry that the scope allows, in the
// registry's order, so a prompt reads the same whatever order the setting
// named them in.
func (s Scope) Catalogue() []Template {
	var out []Template
	for _, t := range Catalogue() {
		if s.Allows(t.Name) {
			out = append(out, t)
		}
	}
	return out
}

// Describe is the scope's templates as a prompt gives them to a model. A role
// is told only of the templates it may ask, so it cannot propose one the
// scope would refuse. The nil Scope describes every template.
func (s Scope) Describe() string {
	return describe(s.Catalogue())
}

// Check refuses a request that names a template outside the scope. It is the
// check every store call passes first, so a store is never asked for a
// template the role may not ask. A name that is not on the list at all is
// refused as unknown, not as out of scope, and without repeating the name,
// which is model text. role names the asker in the refusal.
func (s Scope) Check(role string, req Request) error {
	if s == nil {
		return nil
	}
	name := strings.TrimSpace(req.Template)
	if s.Allows(name) {
		return nil
	}
	if _, _, found := source.Lookup(name); !found {
		return &Refusal{Reason: ReasonUnknownTemplate}
	}
	return &Refusal{Reason: ReasonOutOfScope, Detail: name + " for the " + role}
}

// ParseScope reads a comma-separated list of template names, as proven.yaml
// writes one. A name that is not on the list, a name twice and an empty entry
// are each refused by name; the second result is the entry at fault and the
// third says why.
func ParseScope(list string) (Scope, string, string) {
	var out Scope
	seen := map[string]bool{}
	for _, part := range strings.Split(list, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			return nil, name, "empty entry"
		}
		if _, _, found := source.Lookup(name); !found {
			return nil, name, "unknown template"
		}
		if seen[name] {
			return nil, name, "duplicate template"
		}
		seen[name] = true
		out = append(out, name)
	}
	return out, "", ""
}

// Names are the template names on the list, in the registry's order. A refusal
// of a setting lists them as the allowed values.
func Names() []string {
	var out []string
	for _, t := range Catalogue() {
		out = append(out, t.Name)
	}
	return out
}
