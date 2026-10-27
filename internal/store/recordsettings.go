package store

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/service"
)

// Environment variables of the evidence records and their viewer. The set is
// closed. ADR-0013: every value is a setting, validated when it loads. The
// writing password comes from the environment only and is redacted.
const (
	EnvRecordsDB       = "PROVEN_COUCHDB_RECORDS"
	EnvRecordsUser     = "PROVEN_COUCHDB_RECORDS_USER"
	EnvRecordsPassword = "PROVEN_COUCHDB_RECORDS_PASSWORD"
	EnvEvidenceLink    = "PROVEN_EVIDENCE_LINK_BASE_URL"
)

// recordsDBFormat is what the records database may be called. CouchDB database
// names start with a lower-case letter; the name keeps to the characters that
// need no escaping in a URL path.
var recordsDBFormat = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

// RecordsEnvironment lists the records and viewer settings in the order the
// page shows. Each Name is the Env constant beside this function.
func RecordsEnvironment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvRecordsDB,
			Default:     "none (records are not written)",
			Description: "the CouchDB database that holds a review's evidence records, on the server `PROVEN_COUCHDB_URL` names. Set: each finding of a review is written to it as one document, with the design document and the viewer page. Empty: nothing is written. A lower-case letter, then at most 63 lower-case letters, digits, `_` or `-`",
		},
		{
			Name:        EnvRecordsUser,
			Default:     "none (required when `PROVEN_COUCHDB_RECORDS` is set)",
			Description: "the CouchDB user Proven writes the records as. It must be a different user from `PROVEN_COUCHDB_USER`, the read-only account the viewer is opened with. One line of printable ASCII with no `:`",
		},
		{
			Name:        EnvRecordsPassword,
			Default:     "none (required when `PROVEN_COUCHDB_RECORDS` is set)",
			Secret:      true,
			Description: "that writing user's password. Secret: never printed",
		},
		{
			Name:        EnvEvidenceLink,
			Default:     "none (no evidence-record link)",
			Description: "the address a reader reaches CouchDB at, as an absolute `http` or `https` URL with no user info, query or fragment. Set, with `PROVEN_COUCHDB_RECORDS`: each finding carries a link to its record in the viewer page. Empty: no link. It is never `/_utils`, the admin console",
		},
	}
}

// RecordsSettings is one validated records configuration. A zero value is
// off: nothing is written.
// Password is held so the call can send it. It is not a value to print.
type RecordsSettings struct {
	BaseURL  string
	Database string
	User     string
	Password string
	Timeout  time.Duration
	// Network is the air-gapped policy of ADR-0034. Nil means public.
	Network *network.Policy
}

// Enabled reports whether records are written.
func (s RecordsSettings) Enabled() bool { return s.Database != "" }

// LoadRecords reads the records settings. An empty database name is off and is
// not an error (ADR-0013). A database name that is set needs the server's
// address and a writing account, and a missing one is refused by name: an
// operator who asked for records and got none would not find out until a
// reader looked for them. The writing account must differ from the read-only
// account, so the account the viewer is opened with can never write.
func LoadRecords(env func(string) string) (RecordsSettings, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	db := strings.TrimSpace(env(EnvRecordsDB))
	if db == "" {
		return RecordsSettings{}, nil
	}
	if !recordsDBFormat.MatchString(db) {
		return RecordsSettings{}, refuse(EnvRecordsDB, clip(db), "not a database name", "a lower-case letter, then at most 63 lower-case letters, digits, `_` or `-`; empty means records are not written")
	}
	rawBase := strings.TrimSpace(env(EnvCouchURL))
	if rawBase == "" {
		return RecordsSettings{}, refuse(EnvCouchURL, "", "records need the CouchDB address", "an absolute http or https URL with no user info, query or fragment")
	}
	base, err := checkBaseURL(EnvCouchURL, rawBase)
	if err != nil {
		return RecordsSettings{}, err
	}
	// ADR-0034: in the air-gapped mode an address outside the network is
	// refused here, by name, before the writing credential is ever sent to it.
	policy, perr := network.Load(env)
	if perr != nil {
		return RecordsSettings{}, fromNetwork(perr)
	}
	if err := policy.CheckURL(EnvCouchURL, base); err != nil {
		return RecordsSettings{}, fromNetwork(err)
	}
	user := env(EnvRecordsUser)
	// The user name is not echoed: it is half of a credential.
	if !userFormat.MatchString(user) {
		return RecordsSettings{}, refuse(EnvRecordsUser, "", "records need a writing CouchDB user; the value is not printed", "one line of printable ASCII with no `:`, at most 128 characters")
	}
	if user == env(EnvCouchUser) {
		// ADR-0013: the read-only account opens the viewer; if it could write,
		// a page served to a reader would hold a write right.
		return RecordsSettings{}, refuse(EnvRecordsUser, "", "the writing user is the read-only user; the value is not printed", "a user different from "+EnvCouchUser)
	}
	pass := env(EnvRecordsPassword)
	if strings.ContainsAny(pass, "\r\n") || len(pass) > 4096 || strings.TrimSpace(pass) == "" || service.TooShortSecret(pass) {
		return RecordsSettings{}, refuse(EnvRecordsPassword, "", "records need the writing user's password; the value is not printed", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	timeout := defaultCouchTimeout
	if raw := strings.TrimSpace(env(EnvCouchTimeout)); raw != "" {
		d, derr := time.ParseDuration(raw)
		if derr != nil || d <= 0 {
			return RecordsSettings{}, refuse(EnvCouchTimeout, clip(raw), "not a Go duration greater than zero", "a Go duration greater than zero, for example 60s; empty means 60s")
		}
		timeout = d
	}
	return RecordsSettings{BaseURL: base, Database: db, User: user, Password: pass, Timeout: timeout, Network: policy}, nil
}

// LoadLinkBase reads the address evidence-record links are built from. Empty
// is "no link" and is not an error (ADR-0013).
func LoadLinkBase(env func(string) string) (string, error) {
	if env == nil {
		return "", nil
	}
	raw := strings.TrimSpace(env(EnvEvidenceLink))
	if raw == "" {
		return "", nil
	}
	return checkBaseURL(EnvEvidenceLink, raw)
}

// EvidenceURL is the address of one evidence record in the viewer page: the
// page is an attachment of the design document, so CouchDB serves it, and the
// document id rides in the query so the page reads that one record. It is ""
// when the base address, the database or the id is empty, so a review with the
// setting unset carries no link. The path never names /_utils: that is the
// admin console, which can edit and delete.
func EvidenceURL(base, database, id string) string {
	if base == "" || database == "" || id == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + "/" + database + "/_design/" + designName + "/" + viewerName + "?doc=" + url.QueryEscape(id)
}

// checkBaseURL accepts an absolute http or https URL for the named setting.
// User info is refused, so a password cannot travel in the URL. A query, a
// fragment and a trailing slash are refused or removed, so the paths built
// from it are exact.
func checkBaseURL(field, raw string) (string, error) {
	allowed := "an absolute http or https URL with no user info, query or fragment"
	// ADR-0013: an address that may carry user info is never echoed.
	shown := clip(raw)
	if strings.Contains(raw, "@") {
		shown = ""
	}
	if len(raw) > 2048 || strings.ContainsAny(raw, "\r\n\t ") {
		return "", refuse(field, shown, "not a base URL", allowed)
	}
	u, err := url.Parse(raw)
	// Defensive only: the @ guard above already blanks an address that carries user info.
	if err != nil {
		shown = ""
	}
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.Contains(raw, "?") || strings.Contains(raw, "#") {
		return "", refuse(field, shown, "not a base URL", allowed)
	}
	return strings.TrimRight(raw, "/"), nil
}
