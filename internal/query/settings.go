package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Environment variables of the live Zetaris adapter. The set is closed.
// An empty row limit or timeout means the default. The URL, the key and the
// organisation id have no default.
const (
	EnvRESTURL  = "ZETARIS_REST_URL"
	EnvAPIKey   = "ZETARIS_API_KEY"
	EnvOrgID    = "ZETARIS_ORG_ID"
	EnvRowLimit = "ZETARIS_ROW_LIMIT"
	EnvTimeout  = "ZETARIS_TIMEOUT"
)

const (
	defaultRowLimit = 100
	defaultTimeout  = 60 * time.Second
)

// ConfigError names a refused Zetaris setting and the values that are allowed.
// The API key's Value is always empty: a refusal does not print the secret.
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

// Settings is one validated live-Zetaris configuration.
// APIKey is held so the call can send it. It is not a value to print.
type Settings struct {
	BaseURL  string
	APIKey   string
	OrgID    string
	RowLimit int
	Timeout  time.Duration
}

// Load reads the live Zetaris settings. A wrong value is refused before a review runs.
func Load(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	base := strings.TrimSpace(env(EnvRESTURL))
	if base == "" {
		return Settings{}, refuse(EnvRESTURL, "", "the live query adapter needs a base URL", "an absolute http or https URL with no user info, query or fragment")
	}
	base, err := checkBase(base)
	if err != nil {
		return Settings{}, err
	}
	key := env(EnvAPIKey)
	if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "\r\n") || len(key) > 4096 {
		return Settings{}, refuse(EnvAPIKey, "", "the live query adapter needs an API key", "a non-empty secret in the environment, at most 4096 characters, with no line break")
	}
	org := strings.TrimSpace(env(EnvOrgID))
	if !orgID(org) {
		return Settings{}, refuse(EnvOrgID, clip(org), "not an organisation id", "a non-empty string of digits, at most 18 characters")
	}
	limit := defaultRowLimit
	if raw := strings.TrimSpace(env(EnvRowLimit)); raw != "" {
		n, nerr := strconv.Atoi(raw)
		if nerr != nil || n < 1 || n > 10000 {
			return Settings{}, refuse(EnvRowLimit, clip(raw), "not a row limit", "an integer from 1 to 10000; empty means 100")
		}
		limit = n
	}
	timeout := defaultTimeout
	if raw := strings.TrimSpace(env(EnvTimeout)); raw != "" {
		d, perr := time.ParseDuration(raw)
		if perr != nil || d <= 0 {
			return Settings{}, refuse(EnvTimeout, clip(raw), "not a Go duration greater than zero", "a Go duration greater than zero, for example 60s; empty means 60s")
		}
		timeout = d
	}
	return Settings{BaseURL: base, APIKey: key, OrgID: org, RowLimit: limit, Timeout: timeout}, nil
}

func refuse(field, value, reason, allowed string) *ConfigError {
	if field == EnvAPIKey {
		value = ""
	}
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

func checkBase(raw string) (string, error) {
	allowed := "an absolute http or https URL with no user info, query or fragment"
	if len(raw) > 2048 || strings.ContainsAny(raw, " \t\r\n") || !utf8.ValidString(raw) {
		return "", refuse(EnvRESTURL, clip(raw), "not a base URL", allowed)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return "", refuse(EnvRESTURL, clip(raw), "not a base URL", allowed)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", refuse(EnvRESTURL, clip(raw), "not a base URL", allowed)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	return strings.TrimRight(u.String(), "/"), nil
}

func orgID(s string) bool {
	if s == "" || len(s) > 18 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
