package guardrail

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/service"
)

// Environment variables of the guardrail. The set is closed: Load reads these
// names and no others. An empty value means the default, except the key,
// which has none when the guardrail is on.
const (
	EnvMode    = "PROVEN_GUARDRAIL_MODE"
	EnvBaseURL = "PROVEN_GUARDRAIL_BASE_URL"
	EnvModel   = "PROVEN_GUARDRAIL_MODEL"
	EnvKey     = "PROVEN_GUARDRAIL_KEY"
	EnvTimeout = "PROVEN_GUARDRAIL_TIMEOUT"
)

// Modes. The default is server: the guardrail is on.
const (
	ModeServer = "server"
	ModeNVIDIA = "nvidia"
	ModeOff    = "off"
)

const (
	// ADR-0005 point 1: the default mode keeps the guardrail on.
	defaultMode = ModeServer
	// defaultServerBase is the listen address measured for our model server.
	// 127.0.0.1:8088 is the generator, a different service.
	defaultServerBase  = "http://127.0.0.1:8081"
	defaultNVIDIABase  = "https://integrate.api.nvidia.com/v1"
	defaultServerModel = "nemotron-3.5-content-safety"
	// defaultNVIDIAModel is the id third parties publish. NVIDIA's sample
	// leaves the id blank, so PROVEN_GUARDRAIL_MODEL replaces it.
	defaultNVIDIAModel = "nvidia/nemotron-3.5-content-safety"
	// defaultTimeout sits above the warm range of a local server and leaves a hosted guardrail room for a slow check
	// on a long review prompt.
	defaultTimeout = 120 * time.Second
)

// Environment lists the guardrail settings in the order the page shows.
// Each Name is the Env constant beside this function. Base URL and model
// have one default per mode, so those cells name both.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvMode,
			Default:     envdoc.Tick(defaultMode),
			Description: "`server` (your own model server), `nvidia` (NVIDIA's hosted API) or `off` (the trace says \"no guardrail\")",
		},
		{
			Name:        EnvBaseURL,
			Default:     envdoc.Tick(defaultServerBase) + " for `server`; " + envdoc.Tick(defaultNVIDIABase) + " for `nvidia`",
			Description: "the guardrail's address",
		},
		{
			Name:        EnvModel,
			Default:     envdoc.Tick(defaultServerModel) + " for `server`; " + envdoc.Tick(defaultNVIDIAModel) + " for `nvidia`",
			Description: "the safety model. The hosted id is not yet confirmed by a live call",
		},
		{
			Name:        EnvKey,
			Default:     "none",
			Secret:      true,
			Description: "the guardrail's key. Secret: never printed. Empty is not off: every model call is refused",
		},
		{
			Name:        EnvTimeout,
			Default:     envdoc.Tick(envdoc.SecondSpec(defaultTimeout)),
			Description: "the limit for one check",
		},
	}
}

// ConfigError names a refused guardrail setting and the values that are allowed.
// Value is empty when the field is the key: the secret is never printed.
type ConfigError = service.ConfigError

// Settings is one validated guardrail configuration.
// Key is held so the call can send it. It is not a value to print.
// When Mode is off, Key is empty even if the environment had one.
type Settings struct {
	Mode    string
	BaseURL string
	Model   string
	Key     string
	Timeout time.Duration
	// Network is the air-gapped policy of ADR-0034. Nil means public.
	Network *network.Policy
}

// Load reads the guardrail settings. A wrong value is refused.
// Mode off does not keep the key. An empty lookup uses the defaults.
func Load(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	mode := env(EnvMode)
	if mode == "" {
		mode = defaultMode
	}
	switch mode {
	case ModeServer, ModeNVIDIA, ModeOff:
	default:
		return Settings{}, refuse(EnvMode, clip(mode), "not a guardrail mode", "server, nvidia, or off")
	}
	// ADR-0034: the air-gapped mode refuses NVIDIA's hosted API by name.
	policy, perr := network.Load(env)
	if perr != nil {
		return Settings{}, fromNetwork(perr)
	}
	if mode == ModeNVIDIA {
		if err := policy.Refuse(EnvMode, mode, "the hosted guardrail is outside the organisation's network"); err != nil {
			return Settings{}, fromNetwork(err)
		}
	}
	timeout := defaultTimeout
	if raw := env(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return Settings{}, refuse(EnvTimeout, raw, "not a Go duration greater than zero", "a Go duration greater than zero, for example 5s")
		}
		timeout = d
	}
	// With the guardrail off the other settings are still checked, so a typo is
	// caught before it is switched back on; the key is then dropped, not kept.
	if mode == ModeOff {
		if raw := env(EnvBaseURL); raw != "" {
			checked, err := checkBase(raw)
			if err != nil {
				return Settings{}, err
			}
			if err := policy.CheckURL(EnvBaseURL, checked); err != nil {
				return Settings{}, fromNetwork(err)
			}
		}
		if raw := env(EnvModel); raw != "" {
			if err := checkModel(raw); err != nil {
				return Settings{}, err
			}
		}
		if raw := env(EnvKey); raw != "" {
			// Mode off drops the key. A broken non-empty key is still refused,
			// so turning the guardrail back on later does not hide a bad value.
			// An empty key is not an error here: off is the switch (ADR-0005 point 3).
			if err := refuseBrokenKey(raw); err != nil {
				return Settings{}, err
			}
		}
		return Settings{Mode: ModeOff, Timeout: timeout}, nil
	}
	base := env(EnvBaseURL)
	if base == "" {
		if mode == ModeNVIDIA {
			base = defaultNVIDIABase
		} else {
			base = defaultServerBase
		}
	}
	base, err := checkBase(base)
	if err != nil {
		return Settings{}, err
	}
	if err := policy.CheckURL(EnvBaseURL, base); err != nil {
		return Settings{}, fromNetwork(err)
	}
	model := env(EnvModel)
	if model == "" {
		if mode == ModeNVIDIA {
			model = defaultNVIDIAModel
		} else {
			model = defaultServerModel
		}
	}
	if err := checkModel(model); err != nil {
		return Settings{}, err
	}
	key := env(EnvKey)
	if err := checkKey(key); err != nil {
		return Settings{}, err
	}
	return Settings{Mode: mode, BaseURL: base, Model: model, Key: key, Timeout: timeout, Network: policy}, nil
}

// fromNetwork maps a refused network setting onto this package's ConfigError,
// so the shell reports it like any other refused setting. Any other error passes through.
func fromNetwork(err error) error {
	if ne, ok := network.As(err); ok {
		return &ConfigError{Field: ne.Field, Value: ne.Value, Reason: ne.Reason, Allowed: ne.Allowed}
	}
	return err
}

// refuse builds the refusal for one setting with one allowed form.
func refuse(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

// checkKey refuses a malformed key and names an empty one as down.
// The error never carries the value: it is a secret. A line break could
// split the Authorization header, so a non-empty broken key is still a
// refusal (ADR-0013 point 4a). An empty key is not mode off (ADR-0005 point 5).
func checkKey(key string) error {
	if err := refuseBrokenKey(key); err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return &service.Down{Service: Service, Reason: EnvKey + " is empty"}
	}
	return nil
}

// refuseBrokenKey refuses a key that is set but cannot be sent.
// An empty key is not this case. The refusal's value is empty, so the secret
// is never printed.
func refuseBrokenKey(key string) error {
	// A key too short to be redacted is refused too: the redactor has no
	// length floor, so the loader holds the floor (service.MinSecretBytes).
	if strings.ContainsAny(key, "\r\n") || len(key) > 4096 || service.TooShortSecret(key) {
		return refuse(EnvKey, "", "the guardrail needs a key", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	return nil
}

// checkBase accepts only a plain http or https root. User info would carry a
// credential in the URL, and a query or fragment would change every endpoint
// built from it. The result has no trailing slash.
func checkBase(raw string) (string, error) {
	allowed := "an absolute http or https URL with no user info, query or fragment, and a path of empty or /v1"
	if len(raw) > 2048 || strings.ContainsAny(raw, " \t\r\n") {
		return "", refuse(EnvBaseURL, clip(raw), "not a base URL", allowed)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return "", refuse(EnvBaseURL, clip(raw), "not a base URL", allowed)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", refuse(EnvBaseURL, clip(raw), "not a base URL", allowed)
	}
	path := strings.TrimRight(u.Path, "/")
	if path != "" && path != "/v1" {
		return "", refuse(EnvBaseURL, clip(raw), "the path is not empty or /v1", allowed)
	}
	u.Path = path
	return strings.TrimRight(u.String(), "/"), nil
}

// checkModel refuses a model id that is not valid UTF-8, too long, or has a
// character modelName does not allow.
func checkModel(value string) error {
	if !utf8.ValidString(value) || len(value) > 128 || !modelName(value) {
		return refuse(EnvModel, clip(value), "not an allowed value", "a non-empty model id without spaces, at most 128 characters")
	}
	return nil
}

// clip keeps a refused value short enough to print in an error.
func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}

// modelName allows letters, digits and . _ - / : as in "nvidia/model:tag".
func modelName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r == '.' || r == '_' || r == '-' || r == '/' || r == ':' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			continue
		}
		return false
	}
	return true
}
