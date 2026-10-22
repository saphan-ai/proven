package generator

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// Environment variables of the live generator. The set is closed: Load reads
// these names and no others. An empty value means the default, except the
// token, which has none.
const (
	EnvToken         = "METERLESS_RUNTIME_TOKEN"
	EnvBaseURL       = "PROVEN_GENERATOR_BASE_URL"
	EnvApp           = "PROVEN_GENERATOR_APP"
	EnvModel         = "PROVEN_GENERATOR_MODEL"
	EnvTimeout       = "PROVEN_GENERATOR_TIMEOUT"
	EnvRequireNVIDIA = "PROVEN_GENERATOR_REQUIRE_NVIDIA"
	EnvSkepticModel  = "PROVEN_SKEPTIC_MODEL"
)

// DefaultSkepticModel is Nemotron through Nightshift (ADR-0006 point 5).
// An empty PROVEN_SKEPTIC_MODEL means this model. A set value is used as it is.
const DefaultSkepticModel = "nvidia/nemotron-3.5-super-vl-120b-a12b"

const (
	defaultBaseURL = "http://127.0.0.1:8088"
	defaultApp     = "proven"
	defaultModel   = "meterless/auto"
	defaultTimeout = 60 * time.Second
	defaultRequire = "all"
)

// Environment lists the live-generator settings in the order the page shows.
// Each Name is the Env constant beside this function.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvToken,
			Default:     "none (required live)",
			Secret:      true,
			Description: "the runtime token sent to Nightshift. Secret: never printed",
		},
		{
			Name:        EnvBaseURL,
			Default:     envdoc.Tick(defaultBaseURL),
			Description: "Nightshift's address",
		},
		{
			Name:        EnvApp,
			Default:     envdoc.Tick(defaultApp),
			Description: "the application name sent with each call",
		},
		{
			Name:        EnvModel,
			Default:     envdoc.Tick(defaultModel),
			Description: "the model asked for; Nightshift's router chooses when `auto`",
		},
		{
			Name:        EnvTimeout,
			Default:     envdoc.Tick(envdoc.SecondSpec(defaultTimeout)),
			Description: "the limit for one model call",
		},
		{
			Name:        EnvRequireNVIDIA,
			Default:     envdoc.Tick(defaultRequire),
			Description: "which agents must be answered by an `nvidia/*` model (by the stream's receipt); `all`, `none`, or a list of agents. A non-NVIDIA answer to such an agent is refused by name",
		},
		{
			Name:        EnvSkepticModel,
			Default:     envdoc.Tick(DefaultSkepticModel),
			Description: "the Skeptic's model (ADR-0006: another family than the finder's)",
		},
	}
}

// agents is the closed set a require-NVIDIA list may name.
var agents = []string{"planner", "investigator", "prover", "skeptic", "judge", "writer"}

// ConfigError names a refused generator setting and the values that are allowed.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := fmt.Sprintf("configuration refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// Settings is one validated live-generator configuration.
// The token is held so the call can send it. It is not a value to print.
type Settings struct {
	BaseURL       string
	App           string
	Model         string
	Token         string
	Timeout       time.Duration
	SkepticModel  string
	requireAll    bool
	requireNone   bool
	requireNVIDIA map[string]bool
}

// ModelFor is the model field one agent sends. The skeptic uses SkepticModel.
// Every other agent uses Model.
func (s Settings) ModelFor(agent string) string {
	if agent == "skeptic" && s.SkepticModel != "" {
		return s.SkepticModel
	}
	return s.Model
}

// requiresNVIDIA reports whether this agent's answer must be an nvidia/* model.
func (s Settings) requiresNVIDIA(agent string) bool {
	if s.requireNone {
		return false
	}
	if s.requireAll || s.requireNVIDIA == nil {
		return true
	}
	return s.requireNVIDIA[agent]
}

// Load reads the live-generator settings. A wrong value is refused.
// An empty lookup uses the defaults, and an empty token is refused.
func Load(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	token := env(EnvToken)
	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, "\r\n") || len(token) > 4096 {
		return Settings{}, refuse(EnvToken, "", "the live generator needs a runtime token", "a non-empty secret in the environment, at most 4096 characters, with no line break")
	}
	base := env(EnvBaseURL)
	if base == "" {
		base = defaultBaseURL
	}
	base, err := checkBase(base)
	if err != nil {
		return Settings{}, err
	}
	app := env(EnvApp)
	if app == "" {
		app = defaultApp
	}
	if err := checkToken(EnvApp, app, 64, "1 to 64 letters, digits, dots, underscores or hyphens", appName); err != nil {
		return Settings{}, err
	}
	model := env(EnvModel)
	if model == "" {
		model = defaultModel
	}
	if err := checkToken(EnvModel, model, 128, "a non-empty model or routing mode without spaces, at most 128 characters", modelName); err != nil {
		return Settings{}, err
	}
	skepticModel := env(EnvSkepticModel)
	if skepticModel == "" {
		skepticModel = DefaultSkepticModel
	}
	if err := checkToken(EnvSkepticModel, skepticModel, 128, "a non-empty model or routing mode without spaces, at most 128 characters", modelName); err != nil {
		return Settings{}, err
	}
	timeout := defaultTimeout
	if raw := env(EnvTimeout); raw != "" {
		d, perr := time.ParseDuration(raw)
		if perr != nil || d <= 0 {
			return Settings{}, refuse(EnvTimeout, raw, "not a Go duration greater than zero", "a Go duration greater than zero, for example 60s")
		}
		timeout = d
	}
	req := env(EnvRequireNVIDIA)
	if req == "" {
		req = defaultRequire
	}
	all, none, set, err := parseRequire(req)
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		BaseURL: base, App: app, Model: model, Token: token, Timeout: timeout,
		SkepticModel: skepticModel,
		requireAll:   all, requireNone: none, requireNVIDIA: set,
	}, nil
}

func refuse(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

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

func checkToken(field, value string, max int, allowed string, ok func(string) bool) error {
	if !utf8.ValidString(value) || len(value) > max || !ok(value) {
		return refuse(field, clip(value), "not an allowed value", allowed)
	}
	return nil
}

func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}

func appName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r == '.' || r == '_' || r == '-' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			continue
		}
		return false
	}
	return true
}

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

func parseRequire(raw string) (all, none bool, set map[string]bool, err error) {
	switch raw {
	case "all":
		return true, false, nil, nil
	case "none":
		return false, true, nil, nil
	}
	known := make(map[string]bool, len(agents))
	for _, a := range agents {
		known[a] = true
	}
	set = map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if !known[name] {
			return false, false, nil, refuse(EnvRequireNVIDIA, raw, "unknown agent", "all, none, or a comma-separated list of planner, investigator, prover, skeptic, judge, writer")
		}
		set[name] = true
	}
	if len(set) == 0 {
		return false, false, nil, refuse(EnvRequireNVIDIA, raw, "empty agent list", "all, none, or a comma-separated list of planner, investigator, prover, skeptic, judge, writer")
	}
	return false, false, set, nil
}

// endpoint joins the API root with a method path such as "/chat/completions".
// A base whose path is /v1 is already the API root.
func endpoint(base, method string) string {
	b := strings.TrimRight(base, "/")
	if strings.HasSuffix(b, "/v1") {
		return b + method
	}
	return b + "/v1" + method
}
