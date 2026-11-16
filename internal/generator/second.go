package generator

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/service"
)

// Base settings of the optional second reviewer. Call controls also accept
// PROVEN_SECOND_REVIEWER_ overrides for a different provider or model.
const (
	EnvSecondBaseURL        = "PROVEN_SECOND_REVIEWER_BASE_URL"
	EnvSecondKey            = "PROVEN_SECOND_REVIEWER_KEY"
	EnvSecondModel          = "PROVEN_SECOND_REVIEWER_MODEL"
	EnvSecondTimeout        = "PROVEN_SECOND_REVIEWER_TIMEOUT"
	EnvSecondThinkingBudget = "PROVEN_SECOND_REVIEWER_THINKING_BUDGET"
	EnvSecondMode           = "PROVEN_SECOND_REVIEWER_MODE"
	EnvSecondCommand        = "PROVEN_SECOND_REVIEWER_COMMAND"
)

// SecondService is the name a second-reviewer failure carries in the review's list of failed services.
const SecondService = "second-reviewer"

// SecondEnvironment lists the second reviewer's settings in the order the page shows.
// Each Name is the Env constant beside this function.
func SecondEnvironment() []envdoc.Setting {
	rows := []envdoc.Setting{
		{Name: EnvSecondMode, Default: "openai; command when COMMAND is set", Description: "openai, command or off; the second reviewer may use an API or its own operator-configured CLI"},
		{Name: EnvSecondCommand, Default: "none", Description: "JSON argv for the second reviewer's CLI; prompt on stdin, final answer on stdout; {model} and {agent} are substituted; no shell is started"},
		{
			Name:        EnvSecondKey,
			Default:     "none (optional for local endpoints)",
			Secret:      true,
			Description: "Secret: optional API key, never printed; ignored in command/off mode. An API reviewer without a key needs both endpoint and model",
		},
		{
			Name:        EnvSecondBaseURL,
			Default:     "none (required for an enabled API reviewer)",
			Description: "the address of any OpenAI-compatible chat API, for example `https://api.openai.com/v1` or a local server",
		},
		{
			Name:        EnvSecondModel,
			Default:     "none (required when enabled)",
			Description: "the model asked for; the review names it as the second reviewer",
		},
	}
	rows = append(rows, envdoc.Setting{Name: EnvSecondTimeout, Default: "inherits generator timeout", Description: "positive total attempt timeout for this reviewer"},
		envdoc.Setting{Name: EnvSecondThinkingBudget, Default: "inherits generator thinking budget", Description: "thinking_budget for this reviewer; zero omits the field"})
	for _, row := range CallEnvironment() {
		primary := row.Name
		row.Name = strings.Replace(primary, "PROVEN_GENERATOR_", "PROVEN_SECOND_REVIEWER_", 1)
		row.Default = "inherits " + envdoc.Tick(primary)
		rows = append(rows, row)
	}
	return rows
}

// SecondSettings is one validated second-reviewer configuration. The key is held so the call can
// send it. It is not a value to print.
type SecondSettings struct {
	Mode    string
	Command []string
	Options *CallOptions
	BaseURL string
	Key     string
	Model   string
	Timeout time.Duration
	// ThinkingBudget is the generator's own PROVEN_GENERATOR_THINKING_BUDGET; zero sends nothing.
	ThinkingBudget int
	// Network is the air-gapped policy of . Nil means public.
	Network *network.Policy
}

// LoadSecond reads and validates the second reviewer's configuration. A key,
// or both endpoint and model, enable it. Call controls inherit the primary
// configuration unless a PROVEN_SECOND_REVIEWER_ override is set.
func LoadSecond(env func(string) string) (SecondSettings, bool, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	mode := env(EnvSecondMode)
	if mode == ModeOff {
		return SecondSettings{}, false, nil
	}
	if mode == ModeCommand || (mode == "" && env(EnvSecondCommand) != "") {
		return loadSecondCommand(env)
	}
	if mode != "" && mode != ModeOpenAI {
		return SecondSettings{}, false, refuse(EnvSecondMode, "", "not a second-reviewer mode", "openai, command or off")
	}
	key := env(EnvSecondKey)
	if strings.ContainsAny(key, "\r\n") || len(key) > 4096 || service.TooShortSecret(key) {
		return SecondSettings{}, false, refuse(EnvSecondKey, "", "not a usable key", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	connected := mode == ModeOpenAI || strings.TrimSpace(key) != "" || (env(EnvSecondBaseURL) != "" && env(EnvSecondModel) != "")
	base := env(EnvSecondBaseURL)
	if base == "" {
		if connected {
			return SecondSettings{}, false, refuse(EnvSecondBaseURL, "", "a second reviewer needs an address", "an absolute http or https URL")
		}
	} else {
		var err error
		if base, err = checkBaseFor(EnvSecondBaseURL, base); err != nil {
			return SecondSettings{}, false, err
		}
	}
	// in the air-gapped mode a second reviewer's address off the organisation's network is
	// refused here, by name, even without a key, so a hosted address cannot wait there to be switched on.
	policy, err := network.Load(env)
	if err != nil {
		return SecondSettings{}, false, fromNetwork(err)
	}
	if base != "" {
		if err := policy.CheckURL(EnvSecondBaseURL, base); err != nil {
			return SecondSettings{}, false, fromNetwork(err)
		}
	}
	model := env(EnvSecondModel)
	if model == "" {
		if connected {
			return SecondSettings{}, false, refuse(EnvSecondModel, "", "a second reviewer needs a model", "a non-empty model name without spaces, at most 128 characters")
		}
	} else if err := checkToken(EnvSecondModel, model, 128, "a non-empty model name without spaces, at most 128 characters", modelName); err != nil {
		return SecondSettings{}, false, err
	}
	timeout := defaultTimeout
	if raw := env(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return SecondSettings{}, false, refuse(EnvTimeout, raw, "not a Go duration greater than zero", "a Go duration greater than zero, for example 60s")
		}
		timeout = d
	}
	budget, err := thinkingBudget(env)
	if err != nil {
		return SecondSettings{}, false, err
	}
	lookup := func(k string) string {
		if v := env(strings.Replace(k, "PROVEN_GENERATOR_", "PROVEN_SECOND_REVIEWER_", 1)); v != "" {
			return v
		}
		return env(k)
	}
	options, err := loadCallOptions(lookup)
	if err != nil {
		var cfg *ConfigError
		if errors.As(err, &cfg) {
			copy := *cfg
			if alt := strings.Replace(cfg.Field, "PROVEN_GENERATOR_", "PROVEN_SECOND_REVIEWER_", 1); env(alt) != "" {
				copy.Field = alt
			}
			return SecondSettings{}, false, &copy
		}
		return SecondSettings{}, false, err
	}
	if raw := env(EnvSecondTimeout); raw != "" {
		timeout, err = time.ParseDuration(raw)
		if err != nil || timeout <= 0 {
			return SecondSettings{}, false, refuse(EnvSecondTimeout, "", "invalid timeout", "a positive Go duration")
		}
	}
	if raw := env(EnvSecondThinkingBudget); raw != "" {
		if raw == "0" {
			raw = ""
		}
		budget, err = thinkingBudget(func(string) string { return raw })
		if err != nil {
			return SecondSettings{}, false, refuse(EnvSecondThinkingBudget, "", "invalid thinking budget", "a non-negative integer")
		}
	}
	if !connected {
		return SecondSettings{}, false, nil
	}
	return SecondSettings{Mode: ModeOpenAI, Options: options, BaseURL: base, Key: key, Model: model, Timeout: timeout, ThinkingBudget: budget, Network: policy}, true, nil
}

// Reuse the primary command loader, while keeping credentials, argv and model
// independent. Only call controls inherit; an absent second argv/model cannot
// silently select the primary CLI. Local-network refusal applies to both roles.
func loadSecondCommand(env func(string) string) (SecondSettings, bool, error) {
	lookup := func(name string) string {
		switch name {
		case EnvMode:
			return ModeCommand
		case EnvCommand:
			return env(EnvSecondCommand)
		case EnvModel:
			return env(EnvSecondModel)
		case EnvSkepticModel:
			return ""
		}
		if strings.HasPrefix(name, "PROVEN_GENERATOR_") {
			if value := env(strings.Replace(name, "PROVEN_GENERATOR_", "PROVEN_SECOND_REVIEWER_", 1)); value != "" {
				return value
			}
		}
		return env(name)
	}
	cfg, err := Load(lookup)
	if err != nil {
		var ce *ConfigError
		if errors.As(err, &ce) {
			copy := *ce
			alt := strings.Replace(ce.Field, "PROVEN_GENERATOR_", "PROVEN_SECOND_REVIEWER_", 1)
			if ce.Field == EnvMode || ce.Field == EnvCommand || ce.Field == EnvModel || env(alt) != "" {
				copy.Field = alt
			}
			return SecondSettings{}, false, &copy
		}
		return SecondSettings{}, false, err
	}
	return SecondSettings{Mode: ModeCommand, Command: cfg.Command, Model: cfg.Model, Timeout: cfg.Timeout, Options: cfg.Options, Network: cfg.Network}, true, nil
}

// Second wraps an independently configured API or command model. Failures are
// attributed to this role; command answers retain unknown responding identity.
type Second struct {
	live    *Live
	command *Command
	model   string
}

// Complete sends one prompt and attributes transport failures to the second reviewer.
func (s *Second) Complete(ctx context.Context, agent, prompt string) (string, error) {
	var text string
	var err error
	if s.command != nil {
		text, err = s.command.Complete(ctx, agent, prompt)
	} else {
		text, err = s.live.Complete(ctx, agent, prompt)
	}
	var d *service.Down
	if errors.As(err, &d) {
		return "", &service.Down{Service: SecondService, Reason: d.Reason}
	}
	return text, err
}

// LastUse preserves command metadata; the API route retains its established
// configured-model label, which is not a cryptographic identity receipt.
func (s *Second) LastUse() (Use, bool) {
	if s.command != nil {
		return s.command.LastUse()
	}
	use, ok := s.live.LastUse()
	if !ok {
		return Use{}, false
	}
	use.Model = s.model
	use.FromReceipt = false
	return use, true
}
