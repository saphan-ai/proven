package pipeline

import (
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// The switches of the two stores (ADR-0053). Both stores are on by default; each is switched off by setting
// its variable to off. Sample mode reads them too, because the lab's pull-request review is a sample review.
const (
	// EnvHMem turns H-MEM, the store of refuted findings, on or off.
	EnvHMem = "PROVEN_HMEM"
	// EnvWorldModel turns the World Model, the folded record of what each review found, on or off.
	EnvWorldModel = "PROVEN_WORLD_MODEL"
)

// StoreSettings is the loaded switches of the two stores. The fields say "off" so that the zero value is
// what the default is: both on (ADR-0053). A zero value that meant off is how Markovian ended up off in
// sample mode.
type StoreSettings struct {
	// HMemOff set means the review neither opens, reads nor writes H-MEM. A refuted finding from an earlier
	// review is not suppressed, and the fold records no refutation.
	HMemOff bool
	// WorldOff set means the review neither opens nor writes the World Model: there is no fold.
	WorldOff bool
}

// StoreConfigError is a refused store switch. It names the field, the value and what is allowed; a value is
// never replaced by the default (ADR-0013).
type StoreConfigError struct {
	Field   string
	Value   string
	Reason  string
	Allowed []string
}

// Error names the field, the refused value, why, and what is allowed.
func (e *StoreConfigError) Error() string {
	msg := fmt.Sprintf("store switch refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// LoadStores reads the two switches. An empty value is on; a value that is set and is neither on nor off is
// refused by name and never replaced by the default.
func LoadStores(env func(string) string) (StoreSettings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	var s StoreSettings
	for _, f := range []struct {
		name string
		off  *bool
	}{
		{EnvHMem, &s.HMemOff},
		{EnvWorldModel, &s.WorldOff},
	} {
		switch v := strings.TrimSpace(env(f.name)); v {
		case "", "on":
		case "off":
			*f.off = true
		default:
			return StoreSettings{}, &StoreConfigError{Field: f.name, Value: v, Reason: "not on or off", Allowed: []string{"on", "off; empty means on"}}
		}
	}
	return s, nil
}

// Environment lists the two switches for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvHMem,
			Default:     envdoc.Tick("on"),
			Description: "`on` opens H-MEM, so a finding an earlier review of the same change refuted is dropped and a landed refutation is remembered; `off` neither opens, reads nor writes it. On by default (ADR-0053); its benefit is not measured yet",
		},
		{
			Name:        EnvWorldModel,
			Default:     envdoc.Tick("on"),
			Description: "`on` opens the World Model and folds each finished review's trace into it; `off` neither opens nor writes it. With H-MEM off the World Model still folds and records no refutation. On by default (ADR-0053); its benefit is not measured yet",
		},
	}
}
