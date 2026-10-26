package gateway

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

// The gateway's own settings. The two credentials and the Zetaris address are
// the variables Proven already reads (ADR-0013: one name for one thing), so
// the owner sets them once for both.
const (
	EnvTokensFile      = "JUDGES_GATEWAY_TOKENS_FILE"
	EnvPort            = "JUDGES_GATEWAY_PORT"
	EnvLogFile         = "JUDGES_GATEWAY_LOG"
	EnvModels          = "JUDGES_GATEWAY_MODELS"
	EnvUpstreamTimeout = "JUDGES_GATEWAY_UPSTREAM_TIMEOUT"
	EnvMaxModelReply   = "JUDGES_GATEWAY_MAX_RESPONSE_BYTES"
	EnvNightshiftToken = "METERLESS_RUNTIME_TOKEN"
	EnvNightshiftURL   = "PROVEN_GENERATOR_BASE_URL"
	EnvNightshiftApp   = "PROVEN_GENERATOR_APP"
)

// The defaults of the optional settings.
const (
	DefaultPort            = 8089
	DefaultModels          = "meterless/auto"
	DefaultNightshiftURL   = "http://127.0.0.1:8088"
	DefaultNightshiftApp   = "proven"
	DefaultUpstreamTimeout = 120 * time.Second

	// MaxModelBody is the largest chat request: a change's diff and the
	// prompt around it. MaxEvidenceBody is the largest template request.
	MaxModelBody    = 4 << 20
	MaxEvidenceBody = 16 << 10

	// MaxModelReply is the most bytes the gateway relays from the model
	// service in one call; the setting may be MinModelReply to
	// MaxModelReplySetting.
	MaxModelReply        = 4 << 20
	MinModelReply        = 1 << 10
	MaxModelReplySetting = 64 << 20
)

// Settings is a loaded gateway configuration, ready to serve.
type Settings struct {
	Config   Config
	Port     int
	LogFile  string
	TokensAt string
}

// ConfigError names a refused setting and what is allowed. It never holds the
// value of a credential.
type ConfigError struct {
	Field   string
	Reason  string
	Allowed string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration refused: %s: %s; allowed: %s", e.Field, e.Reason, e.Allowed)
}

// Load reads the environment and builds the gateway's configuration. A wrong
// value is refused by name before the gateway listens. A secret is never part
// of an error.
func Load(env func(string) string) (Settings, error) {
	var s Settings
	s.TokensAt = strings.TrimSpace(env(EnvTokensFile))
	if s.TokensAt == "" {
		return s, &ConfigError{EnvTokensFile, "the token settings file is not named", "a path to the file `judges-gateway token new` writes"}
	}
	store, err := NewStore(s.TokensAt)
	if err != nil {
		return s, &ConfigError{EnvTokensFile, err.Error(), "a readable token settings file in the settings format"}
	}

	s.Port = DefaultPort
	if raw := strings.TrimSpace(env(EnvPort)); raw != "" {
		n, perr := strconv.Atoi(raw)
		if perr != nil || n < 1 || n > 65535 {
			return s, &ConfigError{EnvPort, "not a port", "an integer from 1 to 65535; empty means " + strconv.Itoa(DefaultPort)}
		}
		s.Port = n
	}
	s.LogFile = strings.TrimSpace(env(EnvLogFile))

	token := env(EnvNightshiftToken)
	if token == "" || strings.ContainsAny(token, "\r\n") || len(token) > 4096 || service.TooShortSecret(token) {
		return s, &ConfigError{EnvNightshiftToken, "the model service token is missing or not usable; the value is not printed", "a secret in the environment, 8 to 4096 characters, no line break"}
	}
	base, err := checkOrigin(strings.TrimSpace(env(EnvNightshiftURL)))
	if err != nil {
		return s, err
	}
	app := strings.TrimSpace(env(EnvNightshiftApp))
	if app == "" {
		app = DefaultNightshiftApp
	}
	if strings.ContainsAny(app, "\r\n") || len(app) > 64 {
		return s, &ConfigError{EnvNightshiftApp, "not an application name", "at most 64 characters, no line break"}
	}

	models := strings.TrimSpace(env(EnvModels))
	if models == "" {
		models = DefaultModels
	}
	var list []string
	for _, m := range strings.Split(models, ",") {
		m = strings.TrimSpace(m)
		if m == "" || strings.ContainsAny(m, " \r\n\t\"") || len(m) > 128 {
			return s, &ConfigError{EnvModels, "not a list of model names", "comma-separated names with no space; empty means " + DefaultModels}
		}
		list = append(list, m)
	}

	timeout := DefaultUpstreamTimeout
	if raw := strings.TrimSpace(env(EnvUpstreamTimeout)); raw != "" {
		d, perr := time.ParseDuration(raw)
		if perr != nil || d <= 0 {
			return s, &ConfigError{EnvUpstreamTimeout, "not a Go duration greater than zero", "for example 120s; empty means 120s"}
		}
		timeout = d
	}

	reply := int64(MaxModelReply)
	if raw := strings.TrimSpace(env(EnvMaxModelReply)); raw != "" {
		n, perr := strconv.ParseInt(raw, 10, 64)
		if perr != nil || n < MinModelReply || n > MaxModelReplySetting {
			return s, &ConfigError{EnvMaxModelReply, "not a byte count in range", "an integer from " + strconv.Itoa(MinModelReply) + " to " + strconv.Itoa(MaxModelReplySetting) + "; empty means " + strconv.Itoa(MaxModelReply)}
		}
		reply = n
	}

	zs, err := query.Load(env)
	if err != nil {
		return s, err
	}

	s.Config = Config{
		Tokens:          store,
		Limiter:         NewLimiter(),
		NightshiftURL:   base,
		NightshiftToken: token,
		NightshiftApp:   app,
		Models:          list,
		Upstream: &http.Client{
			Timeout: timeout,
			// A redirect would carry our token to another host.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		Zetaris:         query.NewLive(zs, nil),
		MaxModelBody:    MaxModelBody,
		MaxEvidenceBody: MaxEvidenceBody,
		MaxModelReply:   reply,
	}
	return s, nil
}

// checkOrigin accepts the origin of the model service: http or https, a host,
// and a path that is empty or /v1 (the same rule as the generator's).
func checkOrigin(raw string) (string, error) {
	allowed := "an http or https origin with no user info, query or fragment; a path is empty or /v1; empty means " + DefaultNightshiftURL
	if raw == "" {
		raw = DefaultNightshiftURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", &ConfigError{EnvNightshiftURL, "not an origin", allowed}
	}
	switch strings.TrimRight(u.Path, "/") {
	case "", "/v1":
	default:
		return "", &ConfigError{EnvNightshiftURL, "the path is not empty or /v1", allowed}
	}
	return u.Scheme + "://" + u.Host, nil
}
