package generator

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
)

const (
	ModeOpenAI  = "openai"
	ModeCommand = "command"
	ModeOff     = "off"
	EnvCommand  = "PROVEN_GENERATOR_COMMAND"
)

// Load selects a model transport. No configuration means deterministic analysis.
// Existing Nightshift tokens retain their old behavior during migration.
func Load(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	o, err := loadCallOptions(env)
	if err != nil {
		return Settings{}, err
	}
	s, err := loadProvider(env)
	s.Options = o
	return s, err
}

func loadProvider(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	mode := env(EnvMode)
	if mode == "" {
		switch {
		case env(EnvCommand) != "":
			mode = ModeCommand
		case env(EnvToken) != "":
			mode = ModeNightshift
		case env(EnvBaseURL) != "" || env(EnvModel) != "" || env(EnvKey) != "":
			mode = ModeOpenAI
		default:
			mode = ModeOff
		}
	}
	if mode == ModeNightshift || mode == ModeNVIDIA {
		return loadLegacy(env)
	}
	if mode != ModeOpenAI && mode != ModeCommand && mode != ModeOff {
		return Settings{}, refuse(EnvMode, clip(mode), "not a generator mode", "openai, command, off, nightshift, nvidia")
	}
	s := Settings{Mode: mode, Timeout: defaultTimeout, requireNone: true}
	policy, err := network.Load(env)
	if err != nil {
		return Settings{}, fromNetwork(err)
	}
	s.Network = policy
	if raw := env(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return Settings{}, refuse(EnvTimeout, clip(raw), "not a positive duration", "a Go duration greater than zero")
		}
		s.Timeout = d
	}
	if mode == ModeOff {
		return s, nil
	}
	if raw := env(EnvBaseURL); mode == ModeOpenAI && raw != "" {
		if _, err := checkBase(EnvBaseURL, raw, false); err != nil {
			return Settings{}, err
		}
	}
	s.Model, s.SkepticModel = env(EnvModel), env(EnvSkepticModel)
	if err := checkToken(EnvModel, s.Model, 128, "a non-empty model id", modelName); err != nil {
		return Settings{}, err
	}
	if s.SkepticModel == "" {
		s.SkepticModel = s.Model
	}
	if err := checkToken(EnvSkepticModel, s.SkepticModel, 128, "a non-empty model id", modelName); err != nil {
		return Settings{}, err
	}
	if mode == ModeCommand {
		if policy.Local() {
			return Settings{}, refuse(EnvMode, mode, "external CLI network access cannot be constrained by the HTTP network policy", "use openai with an allowed local endpoint")
		}
		if err := json.Unmarshal([]byte(env(EnvCommand)), &s.Command); err != nil || len(s.Command) == 0 || len(s.Command) > 128 {
			return Settings{}, refuse(EnvCommand, "", "not an argument array", "a JSON array containing the executable and its arguments")
		}
		for _, arg := range s.Command {
			if strings.ContainsRune(arg, 0) || len(arg) > 16384 {
				return Settings{}, refuse(EnvCommand, "", "invalid argument", "arguments without NUL bytes, at most 16384 bytes each")
			}
		}
		if strings.TrimSpace(s.Command[0]) == "" {
			return Settings{}, refuse(EnvCommand, "", "empty executable", "an executable name or path")
		}
		return s, nil
	}
	base := env(EnvBaseURL)
	if base == "" {
		return Settings{}, refuse(EnvBaseURL, "", "an API endpoint is required", "your OpenAI-compatible API root")
	}
	s.BaseURL, err = checkBase(EnvBaseURL, base, false)
	if err != nil {
		return Settings{}, err
	}
	s.Network, err = network.Load(env)
	if err != nil {
		return Settings{}, fromNetwork(err)
	}
	if err := s.Network.CheckURL(EnvBaseURL, s.BaseURL); err != nil {
		return Settings{}, fromNetwork(err)
	}
	s.Token = env(EnvKey)
	if strings.ContainsAny(s.Token, "\r\n") || len(s.Token) > 4096 {
		return Settings{}, refuse(EnvKey, "", "invalid API key", "a single-line key of at most 4096 bytes, or empty")
	}
	s.ThinkingBudget, err = thinkingBudget(env)
	return s, err
}

func ProviderEnvironment() []envdoc.Setting {
	return append([]envdoc.Setting{{Name: EnvCommand, Default: "none", Description: "operator-configured JSON argv for a local CLI; prompt on stdin, answer on stdout; {model} and {agent} are substituted in arguments; no shell is started"}}, CallEnvironment()...)
}
