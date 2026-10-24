package source

import (
	"fmt"
	"sync"
)

// builtins is the fixed list querytpl registers. added is every adapter
// registered afterwards, which is how a test or an organisation extends
// the list without editing the built-ins (ADR-0046).
var (
	mu       sync.Mutex
	builtins []Adapter
	added    []Adapter
)

// SetBuiltins replaces the built-in list.
// It does not remove adapters added with Register. querytpl calls it once,
// from init, because the query text lives in that package and this one
// cannot import it. A duplicate id or template name is refused and the
// previous list is left as it was.
func SetBuiltins(list []Adapter) error {
	mu.Lock()
	defer mu.Unlock()
	// Validate before swapping, so a bad list leaves the previous one.
	if err := checkNew(added, list); err != nil {
		return err
	}
	builtins = append([]Adapter(nil), list...)
	return nil
}

// Register adds one adapter after the built-ins.
// This is the door an organisation uses, and the door a test uses. A
// duplicate id or template name is refused. The built-in list is unchanged.
func Register(a Adapter) error {
	mu.Lock()
	defer mu.Unlock()
	if err := checkNew(append(append([]Adapter(nil), builtins...), added...), []Adapter{a}); err != nil {
		return err
	}
	added = append(added, a)
	return nil
}

// Unregister removes an adapter that Register added.
// A built-in id is left in place: the fixed list is SetBuiltins, and taking
// OSV off it would change every review. An id that is not in the added list
// is a no-op, so a test can clean up either way.
func Unregister(id string) {
	mu.Lock()
	defer mu.Unlock()
	// A new slice, not a reslice of added: the removed adapter must not stay
	// reachable in the old array.
	next := make([]Adapter, 0, len(added))
	for _, a := range added {
		if a != nil && a.ID() == id {
			continue
		}
		next = append(next, a)
	}
	added = next
}

// All returns the built-in adapters, then any added later, in that order.
// The order is fixed so a review of the same change asks the same questions.
func All() []Adapter {
	mu.Lock()
	defer mu.Unlock()
	return allLocked()
}

// ByID returns the adapter with that id.
// The bool is false when none has it.
func ByID(id string) (Adapter, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, a := range allLocked() {
		if a != nil && a.ID() == id {
			return a, true
		}
	}
	return nil, false
}

// Lookup returns the template of that name and the adapter that owns it.
// The bool is false when no adapter does. The first owner wins; Register
// refuses a second template of the same name, so a review cannot see two.
func Lookup(name string) (Adapter, Spec, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, a := range allLocked() {
		if a == nil {
			continue
		}
		for _, spec := range a.Templates() {
			if spec.Template.Name == name {
				return a, spec, true
			}
		}
	}
	return nil, Spec{}, false
}

// Build runs the builder of the named template.
// The bool is false when no adapter owns the name, and then err is nil:
// the caller refuses an unknown template with its own reason. The builder
// is called without the registry lock so it can look a template up itself.
func Build(name string, params map[string]string, rel Rel) (string, bool, error) {
	_, spec, ok := Lookup(name)
	if !ok {
		return "", false, nil
	}
	if spec.Build == nil {
		return "", true, fmt.Errorf("source: template %s has no builder", name)
	}
	sql, err := spec.Build(params, rel)
	return sql, true, err
}

func allLocked() []Adapter {
	out := make([]Adapter, 0, len(builtins)+len(added))
	out = append(out, builtins...)
	out = append(out, added...)
	return out
}

// checkNew refuses an adapter that has no id, no load, no template, or a
// name already used. have is the list already registered. The new list is
// checked against itself and against have.
func checkNew(have, list []Adapter) error {
	ids := map[string]struct{}{}
	names := map[string]struct{}{}
	for _, a := range have {
		if a == nil {
			continue
		}
		ids[a.ID()] = struct{}{}
		for _, spec := range a.Templates() {
			names[spec.Template.Name] = struct{}{}
		}
	}
	for _, a := range list {
		if a == nil || a.ID() == "" {
			return fmt.Errorf("source: an adapter needs an id")
		}
		if _, ok := ids[a.ID()]; ok {
			return fmt.Errorf("source: adapter %s is already registered", a.ID())
		}
		ids[a.ID()] = struct{}{}
		load := a.Load()
		if load.Origin == "" || load.Refresh == "" || load.Table == "" {
			return fmt.Errorf("source: adapter %s needs an origin, a refresh and a table", a.ID())
		}
		specs := a.Templates()
		if len(specs) == 0 {
			return fmt.Errorf("source: adapter %s has no template", a.ID())
		}
		for _, spec := range specs {
			if spec.Template.Name == "" {
				return fmt.Errorf("source: adapter %s has a template with no name", a.ID())
			}
			if _, ok := names[spec.Template.Name]; ok {
				return fmt.Errorf("source: template %s is already registered", spec.Template.Name)
			}
			if spec.Build == nil {
				return fmt.Errorf("source: template %s has no builder", spec.Template.Name)
			}
			names[spec.Template.Name] = struct{}{}
		}
	}
	return nil
}
