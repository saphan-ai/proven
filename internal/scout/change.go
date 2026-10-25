package scout

// This file is Proven's change registry (ADR-0049 point 1): the six kinds of change, as a registry
// Scout scores like any other, and the one call a review makes to classify its changed files.

import (
	_ "embed"
	"fmt"
)

// The six kinds of change ADR-0049 point 1 names, as the registry's intent ids.
const (
	KindDependency = "change.dependency"
	KindRefactor   = "change.refactor"
	KindFeature    = "change.feature"
	KindSecurity   = "change.security"
	KindDocs       = "change.docs"
	KindConfig     = "change.config"
)

// Kinds lists the six kinds in ADR-0049's order.
func Kinds() []string {
	return []string{KindDependency, KindRefactor, KindFeature, KindSecurity, KindDocs, KindConfig}
}

//go:embed changes.json
var changesJSON []byte

// ChangesJSON returns the change registry as shipped, so the eval harness scores the very file the
// review uses.
func ChangesJSON() []byte { return append([]byte(nil), changesJSON...) }

// ChangeRegistry returns the change registry with security added: each pattern a review's own
// configuration files under the security class (ADR-0014) becomes a security signal too, so Scout
// never calls a file ordinary that the team's rules call sensitive. A pattern already present is not
// added twice; dependency manifests are left to the dependency kind, which already names them.
func ChangeRegistry(security []string) (*Registry, error) {
	reg, err := ParseRegistry(changesJSON)
	if err != nil {
		return nil, fmt.Errorf("scout: the shipped change registry: %w", err)
	}
	sec, _ := reg.intent(KindSecurity)
	dep, _ := reg.intent(KindDependency)
	have := map[string]bool{}
	for _, t := range append(append([]Trigger(nil), sec.Triggers...), dep.Triggers...) {
		have[t.Path] = true
	}
	weight := sec.Triggers[0].Weight
	for _, p := range security {
		if have[p] {
			continue
		}
		if err := validGlob(p); err != nil {
			return nil, fmt.Errorf("scout: security pattern %q: %w", p, err)
		}
		have[p] = true
		sec.Triggers = append(sec.Triggers, Trigger{Path: p, Weight: weight})
	}
	return reg, nil
}
