package store

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// Environment variables of the evidence store. The set is closed.
// ADR-0013: every value is a setting, validated when it loads.
const (
	EnvStore = "PROVEN_EVIDENCE_STORE"
	EnvFile  = "PROVEN_EVIDENCE_FILE"
)

// The store names. CouchDB is refused by name until it is built (W49).
const (
	NameZetaris = "zetaris"
	NameLocal   = "local"
	NameCouchDB = "couchdb"
)

// defaultFile is the one file used when the store is local and
// PROVEN_EVIDENCE_FILE is empty. It is a relative path, so it sits in the
// process's working directory.
const defaultFile = "proven-evidence.sql"

// Environment lists the store settings in the order the page shows.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvStore,
			Default:     envdoc.Tick(NameZetaris),
			Description: "which evidence store a live review uses: `zetaris` or `local`. `couchdb` is refused by name because it is not built yet. Empty means `zetaris`. Sample mode uses the local store and does not read this variable",
		},
		{
			Name:        EnvFile,
			Default:     envdoc.Tick(defaultFile),
			Description: "the one file of the local store, read only when the store is `local`. Empty means `proven-evidence.sql`. The file holds tables and no server is started",
		},
	}
}

// ConfigError names a refused store setting and the values that are allowed.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the setting, the refused value, why, and what is allowed, as
// ADR-0013 asks of every refusal.
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

// Choice is the store a live review will open.
// File is set only when Name is local. It does not apply to zetaris.
type Choice struct {
	Name string
	File string
}

// Load reads PROVEN_EVIDENCE_STORE. An empty value is zetaris.
// couchdb is refused by name. Any other word is refused. The file path is
// checked only when the store is local: it does not apply to zetaris.
func Load(env func(string) string) (Choice, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	name := strings.TrimSpace(env(EnvStore))
	if name == "" {
		name = NameZetaris
	}
	switch name {
	case NameCouchDB:
		// ADR-0047: CouchDB is a later wagon. Naming it must not fall through
		// to zetaris.
		return Choice{}, refuse(EnvStore, name, "couchdb is not built yet", "zetaris, local")
	case NameZetaris:
		// The file path applies only when the store is local. A zetaris
		// review does not read it, so a stray value there is not a refusal.
		return Choice{Name: name}, nil
	case NameLocal:
		// A line break is a wrong value, not whitespace to trim away. Trim
		// happens only after that check, so "evidence.sql\n" is refused
		// rather than read as evidence.sql.
		raw := env(EnvFile)
		if raw != "" {
			if err := checkFile(raw); err != nil {
				return Choice{}, err
			}
		}
		file := strings.TrimSpace(raw)
		if file == "" {
			file = defaultFile
		}
		return Choice{Name: name, File: file}, nil
	default:
		return Choice{}, refuse(EnvStore, clip(name), "not a store", "zetaris, local; couchdb is not built yet")
	}
}

func refuse(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

// checkFile refuses a path that is not a single line of text.
// An absolute path and a relative path are both accepted: the operator
// names the file.
func checkFile(path string) error {
	allowed := "a file path of at most 4096 bytes, with no control character"
	if path == "" || len(path) > 4096 || !utf8.ValidString(path) {
		return refuse(EnvFile, clip(path), "not a file path", allowed)
	}
	for i := 0; i < len(path); i++ {
		if path[i] < 0x20 || path[i] == 0x7f {
			return refuse(EnvFile, clip(path), "not a file path", allowed)
		}
	}
	return nil
}

func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
