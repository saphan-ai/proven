package markov

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// The environment settings of bounded steps (ADR-0013, ADR-0048). Sample mode reads them too (ADR-0053): the lab's review is a sample review.
const (
	// EnvSwitch turns the engine on or off. ADR-0053: on by default, and switched off by setting it to off.
	EnvSwitch = "PROVEN_MARKOVIAN"
	// EnvThreshold is the size of a change, in estimated tokens of the Planner's whole prompt, above which the
	// review runs in bounded steps.
	EnvThreshold = "PROVEN_MARKOVIAN_THRESHOLD"
	// EnvUnit is what one step reads: one file, or one hunk.
	EnvUnit = "PROVEN_MARKOVIAN_STEP_UNIT"
	// EnvCarry bounds the state carried from one step to the next, in estimated tokens.
	EnvCarry = "PROVEN_MARKOVIAN_CARRY_TOKENS"
)

// The step units.
const (
	UnitFile = "file"
	UnitHunk = "hunk"
)

// Defaults. The threshold is the spec's default chunkSize, the budget of one step, so a change that would not fit
// one step's budget is the change that is cut. The carried-state bound is the spec's default carryoverTokens.
const (
	DefaultThreshold = 8000
	DefaultCarry     = 800
)

// maxThreshold keeps a typing slip from reading as "never": above it, the setting is refused, and off is the way
// to say never.
const maxThreshold = 10_000_000

// Settings is the loaded configuration of bounded steps.
type Settings struct {
	// On is the switch. The zero value is off; DefaultSettings and an unset environment are on (ADR-0053).
	On bool
	// Threshold is in estimated tokens; only a change strictly above it is cut.
	Threshold int
	// Unit is UnitFile or UnitHunk.
	Unit string
	// Carry bounds the carried state, in estimated tokens.
	Carry int
}

// DefaultSettings is what an unset environment loads: on (ADR-0053), with the documented defaults behind the switch.
func DefaultSettings() Settings {
	return Settings{On: true, Threshold: DefaultThreshold, Unit: UnitFile, Carry: DefaultCarry}
}

// allowedThreshold and allowedCarry are the value lists a refusal names.
var (
	allowedThreshold = fmt.Sprintf("a whole number of estimated tokens from 1 to %d; empty means %d", maxThreshold, DefaultThreshold)
	allowedCarry     = fmt.Sprintf("a whole number of estimated tokens from %d to %d; empty means %d", boundCarryover.min, boundCarryover.max, DefaultCarry)
)

// Load reads the four settings. An empty value is its default; a value that is set and not allowed is refused by
// name with the values allowed, and never replaced by the default (ADR-0013).
func Load(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	s := DefaultSettings()
	switch v := strings.TrimSpace(env(EnvSwitch)); v {
	case "":
	case "on":
		s.On = true
	case "off":
		s.On = false
	default:
		return Settings{}, &ConfigError{Field: EnvSwitch, Value: v, Reason: "not on or off", Allowed: []string{"on", "off; empty means on"}}
	}
	if v := strings.TrimSpace(env(EnvThreshold)); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxThreshold {
			return Settings{}, &ConfigError{Field: EnvThreshold, Value: v, Reason: "not a size in range", Allowed: []string{allowedThreshold}}
		}
		s.Threshold = n
	}
	switch v := strings.TrimSpace(env(EnvUnit)); v {
	case "":
	case UnitFile, UnitHunk:
		s.Unit = v
	default:
		return Settings{}, &ConfigError{Field: EnvUnit, Value: v, Reason: "unknown step unit", Allowed: []string{UnitFile, UnitHunk + "; empty means " + UnitFile}}
	}
	if v := strings.TrimSpace(env(EnvCarry)); v != "" {
		// The spec's carryoverTokens bounds apply: the carried state is the spec's carryover.
		n, err := strconv.Atoi(v)
		if err != nil || n < boundCarryover.min || n > boundCarryover.max {
			return Settings{}, &ConfigError{Field: EnvCarry, Value: v, Reason: "not a bound in range", Allowed: []string{allowedCarry}}
		}
		s.Carry = n
	}
	return s, nil
}

// Environment lists the settings for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvSwitch,
			Default:     envdoc.Tick("on"),
			Description: "`on` reviews a change above the threshold in bounded steps (ADR-0048) and gives each new attempt of a proposed fix only what failed and why; `off` reviews every change whole, as before. On by default (ADR-0053); its benefit is not measured yet",
		},
		{
			Name:        EnvThreshold,
			Default:     envdoc.Tick(strconv.Itoa(DefaultThreshold)),
			Description: "the size, in estimated tokens (characters / 4) of the Planner's whole prompt, above which a change is reviewed in bounded steps; at or below it the review is one whole step",
		},
		{
			Name:        EnvUnit,
			Default:     envdoc.Tick(UnitFile),
			Description: "what one step reads: `file` (one changed file, its base and its change) or `hunk` (one changed region of a file)",
		},
		{
			Name:        EnvCarry,
			Default:     envdoc.Tick(strconv.Itoa(DefaultCarry)),
			Description: fmt.Sprintf("the bound, in estimated tokens, on the state carried from one step to the next (what is established, what is suspected, what awaits evidence); from %d to %d. When it cannot hold every claim that awaits evidence, the review stops and names itself incomplete; it never drops such a claim", boundCarryover.min, boundCarryover.max),
		},
	}
}
