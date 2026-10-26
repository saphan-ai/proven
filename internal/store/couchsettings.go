package store

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/service"
)

// Environment variables of the CouchDB store. The set is closed.
// ADR-0013: every value is a setting, validated when it loads. The password
// comes from the environment only and is redacted.
const (
	EnvCouchURL      = "PROVEN_COUCHDB_URL"
	EnvCouchPrefix   = "PROVEN_COUCHDB_PREFIX"
	EnvCouchUser     = "PROVEN_COUCHDB_USER"
	EnvCouchPassword = "PROVEN_COUCHDB_PASSWORD"
	EnvCouchRowLimit = "PROVEN_COUCHDB_ROW_LIMIT"
	EnvCouchTimeout  = "PROVEN_COUCHDB_TIMEOUT"
)

// The defaults of the optional CouchDB settings. CouchEnvironment prints them
// from these constants, so the page and the code cannot disagree.
const (
	defaultCouchPrefix   = "proven_"
	defaultCouchRowLimit = 100
	defaultCouchTimeout  = 60 * time.Second
)

// couchService is the name a review prints when CouchDB fails.
const couchService = "couchdb"

// prefixFormat is what a database prefix may be. CouchDB database names start
// with a lower-case letter and then hold lower-case letters, digits and a few
// marks; the prefix keeps to the letters, digits and underscore so the name
// built from it needs no escaping in a URL path.
var prefixFormat = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

// userFormat is what a CouchDB user name may be here: one line of printable
// ASCII with no colon, because the colon separates user and password in HTTP
// Basic authentication.
var userFormat = regexp.MustCompile(`^[\x21-\x39\x3b-\x7e]{1,128}$`)

// CouchEnvironment lists the CouchDB settings in the order the page shows.
// Each Name is the Env constant beside this function.
func CouchEnvironment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvCouchURL,
			Default:     "none (required when the store is `couchdb`)",
			Description: "CouchDB's address; an absolute `http` or `https` URL with no user info, query or fragment. Empty: CouchDB is down and the review names it. In `PROVEN_NETWORK=local` an address outside the organisation's network is refused at load, and the store's client refuses to dial one",
		},
		{
			Name:        EnvCouchPrefix,
			Default:     envdoc.Tick(defaultCouchPrefix),
			Description: "the start of every database name. One database holds one registered table, named the prefix, then the table name in lower case with `.` written as `-`. A lower-case letter, then at most 31 lower-case letters, digits or `_`",
		},
		{
			Name:        EnvCouchUser,
			Default:     "none (required when the store is `couchdb`)",
			Description: "the CouchDB user Proven reads as. One line of printable ASCII with no `:`. Empty: CouchDB is down and the review names it",
		},
		{
			Name:        EnvCouchPassword,
			Default:     "none (required when the store is `couchdb`)",
			Secret:      true,
			Description: "that user's password. Secret: never printed. Empty: CouchDB is down and the review names it",
		},
		{
			Name:        EnvCouchRowLimit,
			Default:     envdoc.Tick(strconv.Itoa(defaultCouchRowLimit)),
			Description: "the most rows one query returns",
		},
		{
			Name:        EnvCouchTimeout,
			Default:     envdoc.Tick(envdoc.SecondSpec(defaultCouchTimeout)),
			Description: "the limit for one request to CouchDB",
		},
	}
}

// CouchSettings is one validated CouchDB configuration.
// Password is held so the call can send it. It is not a value to print.
type CouchSettings struct {
	BaseURL  string
	Prefix   string
	User     string
	Password string
	RowLimit int
	Timeout  time.Duration
	// Network is the air-gapped policy of ADR-0034. Nil means public.
	Network *network.Policy
}

// LoadCouch reads the CouchDB settings. A value that is set wrongly is
// refused by name before a review runs (ADR-0013). An empty URL, user or
// password is a down store named couchdb, as it is for Zetaris: the review
// goes on without CouchDB evidence and says so (ADR-0013 point 4a). The first
// empty one is named only after every set value has been checked, so a bad
// value is never hidden behind an empty one.
func LoadCouch(env func(string) string) (CouchSettings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	policy, perr := network.Load(env)
	if perr != nil {
		return CouchSettings{}, fromNetwork(perr)
	}
	var missing string
	base := strings.TrimSpace(env(EnvCouchURL))
	if base == "" {
		missing = EnvCouchURL
	} else {
		var err error
		if base, err = checkCouchBase(base); err != nil {
			return CouchSettings{}, err
		}
		// ADR-0034: in the air-gapped mode an address outside the network is
		// refused here, by name, before the credential is ever sent to it.
		if err := policy.CheckURL(EnvCouchURL, base); err != nil {
			return CouchSettings{}, fromNetwork(err)
		}
	}
	prefix := env(EnvCouchPrefix)
	if prefix == "" {
		prefix = defaultCouchPrefix
	} else if !prefixFormat.MatchString(prefix) {
		return CouchSettings{}, refuse(EnvCouchPrefix, clip(prefix), "not a database prefix", "a lower-case letter, then at most 31 lower-case letters, digits or `_`; empty means "+defaultCouchPrefix)
	}
	user := env(EnvCouchUser)
	if user != "" && !userFormat.MatchString(user) {
		// The user name is not echoed: it is half of a credential.
		return CouchSettings{}, refuse(EnvCouchUser, "", "not a CouchDB user name; the value is not printed", "one line of printable ASCII with no `:`, at most 128 characters")
	}
	pass := env(EnvCouchPassword)
	// A password too short to be redacted is refused too: the redactor has no
	// length floor, so the loader holds it (service.MinSecretBytes).
	if strings.ContainsAny(pass, "\r\n") || len(pass) > 4096 || service.TooShortSecret(pass) {
		return CouchSettings{}, refuse(EnvCouchPassword, "", "the CouchDB store needs a password; the value is not printed", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	limit := defaultCouchRowLimit
	if raw := strings.TrimSpace(env(EnvCouchRowLimit)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 10000 {
			return CouchSettings{}, refuse(EnvCouchRowLimit, clip(raw), "not a row limit", "an integer from 1 to 10000; empty means 100")
		}
		limit = n
	}
	timeout := defaultCouchTimeout
	if raw := strings.TrimSpace(env(EnvCouchTimeout)); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return CouchSettings{}, refuse(EnvCouchTimeout, clip(raw), "not a Go duration greater than zero", "a Go duration greater than zero, for example 60s; empty means 60s")
		}
		timeout = d
	}
	switch {
	case missing != "":
	case user == "":
		missing = EnvCouchUser
	case strings.TrimSpace(pass) == "":
		missing = EnvCouchPassword
	}
	if missing != "" {
		// The reason names the variable and never a value.
		return CouchSettings{}, &service.Down{Service: couchService, Reason: missing + " is empty"}
	}
	return CouchSettings{BaseURL: base, Prefix: prefix, User: user, Password: pass, RowLimit: limit, Timeout: timeout, Network: policy}, nil
}

// fromNetwork maps a refused network setting onto this package's ConfigError.
func fromNetwork(err error) error {
	if ne, ok := network.As(err); ok {
		return &ConfigError{Field: ne.Field, Value: ne.Value, Reason: ne.Reason, Allowed: ne.Allowed}
	}
	return err
}

// checkCouchBase accepts the CouchDB address. The rules are checkBaseURL's,
// which the records settings share.
func checkCouchBase(raw string) (string, error) {
	return checkBaseURL(EnvCouchURL, raw)
}
