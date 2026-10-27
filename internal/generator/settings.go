package generator

import (
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/service"
)

// Environment variables of the live generator. The set is closed: Load reads
// these names and no others. An empty value means the default, except the
// token, which has none. ADR-0013: no value the generator runs with is
// hardcoded; each comes from one of these settings or its named default.
const (
	EnvMode           = "PROVEN_GENERATOR_MODE"
	EnvKey            = "PROVEN_GENERATOR_KEY"
	EnvToken          = "METERLESS_RUNTIME_TOKEN"
	EnvBaseURL        = "PROVEN_GENERATOR_BASE_URL"
	EnvNVIDIABaseURL  = "PROVEN_NVIDIA_BASE_URL"
	EnvApp            = "PROVEN_GENERATOR_APP"
	EnvModel          = "PROVEN_GENERATOR_MODEL"
	EnvTimeout        = "PROVEN_GENERATOR_TIMEOUT"
	EnvRequireNVIDIA  = "PROVEN_GENERATOR_REQUIRE_NVIDIA"
	EnvSkepticModel   = "PROVEN_SKEPTIC_MODEL"
	EnvThinkingBudget = "PROVEN_GENERATOR_THINKING_BUDGET"
)

// DefaultSkepticModel is Nemotron through Nightshift (ADR-0006 point 5).
// An empty PROVEN_SKEPTIC_MODEL means this model. A set value is used as it is.
const DefaultSkepticModel = "nvidia/nemotron-3.5-super-vl-120b-a12b"

// The two providers PROVEN_GENERATOR_MODE chooses between (ADR-0031 point 1).
const (
	// ModeNightshift calls Meterless Nightshift and is the default (ADR-0019, ADR-0031 point 2).
	ModeNightshift = "nightshift"
	// ModeNVIDIA calls NVIDIA's hosted API directly (ADR-0031 point 3).
	ModeNVIDIA = "nvidia"
)

const (
	// The defaults below apply when a setting is empty. defaultRequire "all" means
	// every agent must be answered by an NVIDIA model (ADR-0019 point 4).
	defaultMode    = ModeNightshift
	defaultBaseURL = "http://127.0.0.1:8088"
	defaultApp     = "proven"
	defaultModel   = "meterless/auto"
	defaultTimeout = 60 * time.Second
	defaultRequire = "all"

	// The defaults of nvidia mode (ADR-0031 point 3). The base is the hosted
	// API's OpenAI-compatible root. The model is one id from the list NVIDIA
	// publishes at build.nvidia.com. The earlier default was retired (HTTP 410,
	// measured 2026-10-25), so the finder now asks for the Skeptic's model; the
	// trace notes the shared family (ADR-0006). The API names this model's
	// answer with an alias, accepted by modelAliases. A setting replaces either.
	defaultNVIDIABase  = "https://integrate.api.nvidia.com/v1"
	defaultNVIDIAModel = "nvidia/nemotron-3.5-super-vl-120b-a12b"
)

// Environment lists the live-generator settings in the order the page shows.
// Each Name is the Env constant beside this function.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvMode,
			Default:     envdoc.Tick(defaultMode),
			Description: "`nightshift` (Meterless Nightshift) or `nvidia` (NVIDIA's hosted API, ADR-0031)",
		},
		{
			Name:        EnvToken,
			Default:     "none (required in `nightshift` mode)",
			Secret:      true,
			Description: "the runtime token sent to Nightshift. Secret: never printed. Empty: Nightshift is down and the review names it. Not read in `nvidia` mode",
		},
		{
			Name:        EnvKey,
			Default:     "none (required in `nvidia` mode)",
			Secret:      true,
			Description: "the NVIDIA API key, from build.nvidia.com. Secret: never printed. Empty: the hosted API is down and the review names it. Not read in `nightshift` mode",
		},
		{
			Name:        EnvBaseURL,
			Default:     envdoc.Tick(defaultBaseURL),
			Description: "Nightshift's address, or the root of another OpenAI-compatible API. A path is empty or ends in a version, such as `/v1` or GLM's `/api/paas/v4`, and that path is then the API root. Read in `nightshift` mode only; `nvidia` mode never reads it",
		},
		{
			Name:        EnvNVIDIABaseURL,
			Default:     envdoc.Tick(defaultNVIDIABase),
			Description: "the hosted API's address. Read in `nvidia` mode only. It must be `https://`: any other value is refused by name, so the key is never sent over plain http. The answer's `model` field is trusted as the answering model only from this address (ADR-0031)",
		},
		{
			Name:        EnvApp,
			Default:     envdoc.Tick(defaultApp),
			Description: "the application name sent with each call to Nightshift. Not sent in `nvidia` mode",
		},
		{
			Name:        EnvModel,
			Default:     envdoc.Tick(defaultModel) + " for `nightshift`; " + envdoc.Tick(defaultNVIDIAModel) + " for `nvidia`",
			Description: "the model asked for; Nightshift's router chooses when `auto`. The hosted default answers under the alias `private/nvidia/nemotron-3.5-super-120b-a12b`, accepted for this model only",
		},
		{
			Name:        EnvTimeout,
			Default:     envdoc.Tick(envdoc.SecondSpec(defaultTimeout)),
			Description: "the limit for one model call",
		},
		{
			Name:        EnvRequireNVIDIA,
			Default:     envdoc.Tick(defaultRequire),
			Description: "which agents must be answered by an `nvidia/*` model (by Nightshift's receipt, or in `nvidia` mode by the `model` field of the hosted API's answer); `all`, `none`, or a list of agents. A non-NVIDIA answer to such an agent is refused by name",
		},
		{
			Name:        EnvSkepticModel,
			Default:     envdoc.Tick(DefaultSkepticModel),
			Description: "the Skeptic's model (ADR-0006: another family than the finder's). In `nvidia` mode the finder's default is this same model, so the two share a family unless one is set, and the trace states the shared family",
		},
		{
			Name:        EnvThinkingBudget,
			Default:     "empty (nothing is sent)",
			Description: "the most tokens a reasoning model may think before it answers, sent as `thinking_budget` on every call: for a model that always thinks, such as GLM-5.3 on Alibaba Cloud Model Studio, which refuses to turn thinking off. Empty sends nothing, so a provider that does not know the field never sees it",
		},
	}
}

// agents is the closed set a require-NVIDIA list may name.
// The names are the six agents of the review and the Fixer (ADR-0042), as they call the generator.
var agents = []string{"planner", "investigator", "prover", "fixer", "skeptic", "judge", "writer"}

// ConfigError names a refused generator setting and the values that are allowed.
// A refused token is reported with an empty value, so the secret is never printed.
type ConfigError = service.ConfigError

// Settings is one validated live-generator configuration.
// The token is held so the call can send it. It is not a value to print.
type Settings struct {
	// Mode is ModeNightshift or ModeNVIDIA.
	Mode    string
	BaseURL string
	App     string
	Model   string
	// Token is the Nightshift runtime token or, in nvidia mode, the NVIDIA key.
	Token        string
	Timeout      time.Duration
	SkepticModel string
	// ThinkingBudget is PROVEN_GENERATOR_THINKING_BUDGET; zero sends no thinking_budget field.
	ThinkingBudget int
	plain          bool
	// Network is the air-gapped policy of ADR-0034. Nil means public.
	Network       *network.Policy
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

// BaseHost is the host (and port, if any) of the address this generator calls.
// It names where an answer came from, without the path.
func (s Settings) BaseHost() string {
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return ""
	}
	return u.Host
}

// requiresNVIDIA reports whether this agent's answer must be an nvidia/* model.
func (s Settings) requiresNVIDIA(agent string) bool {
	if s.requireNone {
		return false
	}
	// A Settings not built by Load has no list: it fails closed and requires NVIDIA.
	if s.requireAll || s.requireNVIDIA == nil {
		return true
	}
	return s.requireNVIDIA[agent]
}

// Load reads the live-generator settings. A wrong non-empty value is refused.
// An empty token or key is not a refusal: the service is down (ADR-0013 point 4a).
// An empty lookup uses the defaults for every setting that has one.
func Load(env func(string) string) (Settings, error) {
	// A nil lookup is an empty environment, so the defaults apply and the token is empty.
	if env == nil {
		env = func(string) string { return "" }
	}
	mode := env(EnvMode)
	if mode == "" {
		mode = defaultMode
	}
	if mode != ModeNightshift && mode != ModeNVIDIA {
		return Settings{}, refuse(EnvMode, clip(mode), "not a generator mode", "nightshift or nvidia")
	}
	// Each mode reads its own secret and ignores the other, so a key left in the
	// environment for one provider is never sent to the other (ADR-0031 point 3).
	secretEnv, svc, baseDefault, modelDefault := EnvToken, "nightshift", defaultBaseURL, defaultModel
	if mode == ModeNVIDIA {
		secretEnv, svc, baseDefault, modelDefault = EnvKey, "nvidia", defaultNVIDIABase, defaultNVIDIAModel
	}
	token := env(secretEnv)
	// A line break or an over-long token is a wrong value, refused before any
	// review. The value is not echoed: it is a secret. An empty token is
	// handled after the other settings, so a bad URL is still refused.
	// A secret too short to be redacted is refused too: the redactor has no
	// length floor, so the loader holds the floor (service.MinSecretBytes).
	if strings.ContainsAny(token, "\r\n") || len(token) > 4096 || service.TooShortSecret(token) {
		return Settings{}, refuse(secretEnv, "", "the live generator needs a secret", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	// ADR-0031 point 3: each mode reads its own address, as it reads its own secret.
	// The key must never follow a base meant for Nightshift, because in nvidia mode
	// the answer's own model field is trusted as the answering model. A base left
	// set for Nightshift is therefore not read here.
	baseEnv, needHTTPS := EnvBaseURL, false
	if mode == ModeNVIDIA {
		baseEnv, needHTTPS = EnvNVIDIABaseURL, true
	}
	base := env(baseEnv)
	if base == "" {
		base = baseDefault
	}
	base, err := checkBase(baseEnv, base, needHTTPS)
	if err != nil {
		return Settings{}, err
	}
	// ADR-0034: in the air-gapped mode a hosted address is refused here, by name.
	// The hosted generator mode is refused by name even when its address is local,
	// the same way the guardrail's nvidia mode is.
	policy, err := network.Load(env)
	if err != nil {
		return Settings{}, fromNetwork(err)
	}
	if mode == ModeNVIDIA {
		if err := policy.Refuse(EnvMode, mode, "the hosted generator is outside the organisation's network"); err != nil {
			return Settings{}, fromNetwork(err)
		}
	}
	if err := policy.CheckURL(baseEnv, base); err != nil {
		return Settings{}, fromNetwork(err)
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
		model = modelDefault
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
	budget := 0
	if raw := env(EnvThinkingBudget); raw != "" {
		n, perr := strconv.Atoi(raw)
		if perr != nil || n < 1 || n > 1000000 {
			return Settings{}, refuse(EnvThinkingBudget, clip(raw), "not a token count", "a whole number from 1 to 1000000, or empty")
		}
		budget = n
	}
	req := env(EnvRequireNVIDIA)
	if req == "" {
		req = defaultRequire
	}
	all, none, set, err := parseRequire(req)
	if err != nil {
		return Settings{}, err
	}
	// ADR-0013 point 4a: an empty token or key means the service is down. The
	// review goes on and names it. The reason names the variable and not the value.
	if strings.TrimSpace(token) == "" {
		return Settings{}, &service.Down{Service: svc, Reason: secretEnv + " is empty"}
	}
	return Settings{
		Mode: mode, BaseURL: base, App: app, Model: model, Token: token, Timeout: timeout,
		SkepticModel: skepticModel, Network: policy, ThinkingBudget: budget,
		requireAll: all, requireNone: none, requireNVIDIA: set,
	}, nil
}

// serviceName is the name a failed call carries in the review's failed services.
func (s Settings) serviceName() string {
	if s.Mode == ModeNVIDIA {
		return "nvidia"
	}
	return "nightshift"
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

// checkBase accepts only a plain http or https root. User info would carry a
// credential in the URL, and a query or fragment would change every endpoint
// built from it. The result has no trailing slash. field names the setting in a
// refusal. httpsOnly is set for the NVIDIA key's address: a key must not travel
// over plain http.
func checkBase(field, raw string, httpsOnly bool) (string, error) {
	allowed := "an absolute http or https URL with no user info, query or fragment, and a path that is empty or ends in a version such as /v1 or /api/paas/v4"
	if httpsOnly {
		allowed = "an absolute https URL with no user info, query or fragment, and a path that is empty or ends in a version such as /v1 or /api/paas/v4"
	}
	if len(raw) > 2048 || strings.ContainsAny(raw, " \t\r\n") {
		return "", refuse(field, clip(raw), "not a base URL", allowed)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return "", refuse(field, clip(raw), "not a base URL", allowed)
	}
	if httpsOnly && u.Scheme != "https" {
		return "", refuse(field, clip(raw), "the NVIDIA key is not sent over plain http", allowed)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", refuse(field, clip(raw), "not a base URL", allowed)
	}
	path := strings.TrimRight(u.Path, "/")
	if path != "" && !versioned(path) {
		return "", refuse(field, clip(raw), "the path is not empty and does not end in a version", allowed)
	}
	u.Path = path
	return strings.TrimRight(u.String(), "/"), nil
}

// checkToken refuses a value that is not valid UTF-8, is longer than max, or has a
// character ok does not allow.
func checkToken(field, value string, max int, allowed string, ok func(string) bool) error {
	if !utf8.ValidString(value) || len(value) > max || !ok(value) {
		return refuse(field, clip(value), "not an allowed value", allowed)
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

// appName allows letters, digits, dots, underscores and hyphens, because the
// value is sent as a header.
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

// modelName allows the app-name characters plus "/" and ":", as in "nvidia/model:tag".
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

// parseRequire reads PROVEN_GENERATOR_REQUIRE_NVIDIA: "all", "none", or a
// comma-separated list of agent names. An unknown name is refused, not ignored,
// so a typo cannot quietly drop an agent from the NVIDIA check.
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
			return false, false, nil, refuse(EnvRequireNVIDIA, raw, "unknown agent", "all, none, or a comma-separated list of planner, investigator, prover, fixer, skeptic, judge, writer")
		}
		set[name] = true
	}
	if len(set) == 0 {
		return false, false, nil, refuse(EnvRequireNVIDIA, raw, "empty agent list", "all, none, or a comma-separated list of planner, investigator, prover, fixer, skeptic, judge, writer")
	}
	return false, false, set, nil
}

// endpoint joins the API root with a method path such as "/chat/completions".
// A base whose path ends in a version (/v1, or GLM's /api/paas/v4) is already the API root.
func endpoint(base, method string) string {
	b := strings.TrimRight(base, "/")
	if u, err := url.Parse(b); err == nil && versioned(u.Path) {
		return b + method
	}
	return b + "/v1" + method
}

// versioned reports whether path's last segment is an API version: v and one or more digits.
func versioned(path string) bool {
	last := path[strings.LastIndex(path, "/")+1:]
	if len(last) < 2 || last[0] != 'v' {
		return false
	}
	for _, r := range last[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// checkBaseFor checks the address in the setting named field, so the second reviewer's address
// is refused under its own name and never under Nightshift's (ADR-0037).
func checkBaseFor(field, raw string) (string, error) {
	return checkBase(field, raw, false)
}
