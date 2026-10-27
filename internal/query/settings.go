package query

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

// Environment variables of the live Zetaris adapter. The set is closed.
// An empty row limit or timeout means the default. The URL, the key and the
// organisation id have no default.
// ADR-0013 and ADR-0015, point 3: every value of the adapter is a setting,
// validated when it loads. The key comes from the environment only.
const (
	EnvRESTURL  = "ZETARIS_REST_URL"
	EnvAPIKey   = "ZETARIS_API_KEY"
	EnvOrgID    = "ZETARIS_ORG_ID"
	EnvRowLimit = "ZETARIS_ROW_LIMIT"
	EnvTimeout  = "ZETARIS_TIMEOUT"
)

// The defaults of the two optional settings. Environment prints them from
// these constants, so the page and the code cannot disagree.
const (
	defaultRowLimit = 100
	defaultTimeout  = 60 * time.Second
)

// Environment lists the live Zetaris settings in the order the page shows.
// Each Name is the Env constant beside this function. The evidence view is
// not here: livecheck owns it.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvRESTURL,
			Default:     "none (required live)",
			Description: "Zetaris's REST address; an absolute `http` or `https` URL with no user info, query or fragment",
		},
		{
			Name:        EnvAPIKey,
			Default:     "none (required live)",
			Secret:      true,
			Description: "Proven's Zetaris key. Secret: never printed. Empty: Zetaris is down and the review names it",
		},
		{
			Name:        EnvOrgID,
			Default:     "none (required live)",
			Description: "the organisation id sent with each query",
		},
		{
			Name:        EnvRowLimit,
			Default:     envdoc.Tick(strconv.Itoa(defaultRowLimit)),
			Description: "the most rows one query returns",
		},
		{
			Name:        EnvTimeout,
			Default:     envdoc.Tick(envdoc.SecondSpec(defaultTimeout)),
			Description: "the limit for one query",
		},
	}
}

// ConfigError names a refused Zetaris setting and the values that are allowed.
// The Value of the API key and of the organisation id is always empty: a
// refusal names the variable and the rule, and does not print either.
type ConfigError = service.ConfigError

// Settings is one validated live-Zetaris configuration.
// APIKey is held so the call can send it. It is not a value to print.
type Settings struct {
	BaseURL  string
	APIKey   string
	OrgID    string
	RowLimit int
	Timeout  time.Duration
	// Network is the air-gapped policy of ADR-0034. Nil means public.
	Network *network.Policy
}

// Load reads the live Zetaris settings. A wrong value is refused before a review runs.
// ADR-0013: a value that is set wrongly is refused by name. Only an empty
// optional value takes its default.
func Load(env func(string) string) (Settings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	// ADR-0013 point 4a: an empty URL, key or organisation id means Zetaris is
	// down, not that the review is refused. A value that is set but wrong is
	// still refused. The first empty one is named after every set value has
	// been checked, so a bad value is never hidden behind an empty one.
	policy, perr := network.Load(env)
	if perr != nil {
		return Settings{}, fromNetwork(perr)
	}
	var missing string
	base := strings.TrimSpace(env(EnvRESTURL))
	if base == "" {
		missing = EnvRESTURL
	} else {
		var err error
		base, err = checkBase(base)
		if err != nil {
			return Settings{}, err
		}
		// ADR-0034: in the air-gapped mode a hosted address is refused here, by name.
		if err := policy.CheckURL(EnvRESTURL, base); err != nil {
			return Settings{}, fromNetwork(err)
		}
	}
	// A line break or an over-long key is a wrong value, refused before any
	// review. The value is not echoed: it is a secret. An empty key is a down
	// service, decided after the other settings so a bad URL is still refused.
	key := env(EnvAPIKey)
	// A key too short to be redacted is refused too: the redactor has no
	// length floor, so the loader holds the floor (service.MinSecretBytes).
	if strings.ContainsAny(key, "\r\n") || len(key) > 4096 || service.TooShortSecret(key) {
		return Settings{}, refuse(EnvAPIKey, "", "the live query adapter needs an API key", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	// ADR-0023, point 6: the organisation id is the numeric id Zetaris issues.
	org := strings.TrimSpace(env(EnvOrgID))
	if org == "" {
		if missing == "" {
			missing = EnvOrgID
		}
	} else if !orgID(org) {
		// ADR-0013 point 4: the id identifies Proven's account, so a refusal names the variable and
		// the rule and says the value is withheld, as it does for the key.
		return Settings{}, refuse(EnvOrgID, "", "not an organisation id; the value is not printed", "a non-empty string of digits, at most 18 characters")
	}
	// The row limit is bounded, so one query cannot pull a whole source into the
	// review.
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
	// ADR-0013 point 4a: an empty key means Zetaris is down. The review goes
	// on and names it. The reason names the variable and not the value.
	if missing == "" && strings.TrimSpace(key) == "" {
		missing = EnvAPIKey
	}
	if missing != "" {
		return Settings{}, &service.Down{Service: "zetaris", Reason: missing + " is empty"}
	}
	return Settings{BaseURL: base, APIKey: key, OrgID: org, RowLimit: limit, Timeout: timeout, Network: policy}, nil
}

// fromNetwork maps a refused network setting onto this package's ConfigError,
// so the shell reports it like any other refused setting. Any other error passes through.
func fromNetwork(err error) error {
	if ne, ok := network.As(err); ok {
		return &ConfigError{Field: ne.Field, Value: ne.Value, Reason: ne.Reason, Allowed: ne.Allowed}
	}
	return err
}

// refuse builds a ConfigError. The value of the key and of the organisation id
// is blanked here, in one place, so no caller can print it.
func refuse(field, value, reason, allowed string) *ConfigError {
	if field == EnvAPIKey || field == EnvOrgID {
		value = ""
	}
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

// checkBase accepts an absolute http or https URL. User info is refused, so a
// password cannot travel in the URL. A query, a fragment and a trailing slash
// are refused or trimmed, because the client appends the route to the path.
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

// orgID reports whether s is an organisation id: digits only, at most 18 of
// them, so it fits a 64-bit integer.
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

// clip shortens a refused value before it is printed, so a long paste does
// not flood the refusal.
func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
