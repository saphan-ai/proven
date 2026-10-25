package store

import (
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// TestLoadDefaultsToZetaris proves an empty PROVEN_EVIDENCE_STORE is
// zetaris and does not require a file, couchdb is a store that needs no file,
// and a word that is not a store is refused naming all three. The mutated
// field is the variable: empty, then couchdb, then a word that is not a store.
func TestLoadDefaultsToZetaris(t *testing.T) {
	got, err := Load(func(string) string { return "" })
	if err != nil || got.Name != NameZetaris || got.File != "" {
		t.Fatalf("%#v %v", got, err)
	}
	got, err = Load(func(k string) string {
		if k == EnvStore {
			return NameCouchDB
		}
		return ""
	})
	if err != nil || got.Name != NameCouchDB || got.File != "" {
		t.Fatalf("couchdb: %#v %v", got, err)
	}
	var cfg *ConfigError
	_, err = Load(func(k string) string {
		if k == EnvStore {
			return "postgres"
		}
		return ""
	})
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvStore || cfg.Value != "postgres" {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "couchdb") {
		t.Fatalf("the refusal does not name couchdb as allowed: %v", err)
	}
}

// couchEnv builds an environment function over a map.
func couchEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// TestLoadCouchReadsTheSettings proves a full set loads with its defaults for
// the optional ones, and a trailing slash on the address is dropped
// (ADR-0013). The mutated fields are the optional ones, left empty.
func TestLoadCouchReadsTheSettings(t *testing.T) {
	got, err := LoadCouch(couchEnv(map[string]string{
		EnvCouchURL: "http://couch.example:5984/", EnvCouchUser: "reader", EnvCouchPassword: "a-long-password",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := CouchSettings{BaseURL: "http://couch.example:5984", Prefix: "proven_", User: "reader", Password: "a-long-password", RowLimit: 100, Timeout: 60 * time.Second}
	if got != want {
		t.Fatalf("%#v, want %#v", got, want)
	}
}

// TestLoadCouchEmptyIsADownStoreAndWrongIsRefused proves an empty address,
// user or password is a down store named couchdb with the variable's name,
// and a value that is set wrongly is refused by name without printing a
// secret (ADR-0013 points 4 and 4a). The mutated field is each variable in
// turn; a bad address is checked before an empty password, so it is never
// hidden behind one.
func TestLoadCouchEmptyIsADownStoreAndWrongIsRefused(t *testing.T) {
	full := map[string]string{EnvCouchURL: "http://c:5984", EnvCouchUser: "reader", EnvCouchPassword: "a-long-password"}
	with := func(k, v string) map[string]string {
		m := map[string]string{}
		for a, b := range full {
			m[a] = b
		}
		m[k] = v
		return m
	}
	for _, name := range []string{EnvCouchURL, EnvCouchUser, EnvCouchPassword} {
		_, err := LoadCouch(couchEnv(with(name, "")))
		d, ok := service.As(err)
		if !ok || d.Service != "couchdb" || !strings.Contains(d.Reason, name) {
			t.Fatalf("%s empty: %v", name, err)
		}
	}
	for name, bad := range map[string]string{
		EnvCouchURL:      "ftp://c",
		EnvCouchPrefix:   "Bad-Prefix",
		EnvCouchUser:     "a:b",
		EnvCouchPassword: "short",
		EnvCouchRowLimit: "0",
		EnvCouchTimeout:  "-1s",
	} {
		_, err := LoadCouch(couchEnv(with(name, bad)))
		var cfg *ConfigError
		if err == nil || !asConfig(err, &cfg) || cfg.Field != name {
			t.Fatalf("%s=%q: %v", name, bad, err)
		}
		if (name == EnvCouchPassword || name == EnvCouchUser) && (cfg.Value != "" || strings.Contains(err.Error(), bad)) {
			t.Fatalf("%s: the refusal printed the value: %v", name, err)
		}
	}
	// A bad address with an empty password is refused, not called down.
	_, err := LoadCouch(couchEnv(map[string]string{EnvCouchURL: "http://u:p@c", EnvCouchUser: "r"}))
	var cfg *ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvCouchURL {
		t.Fatalf("user info in the address: %v", err)
	}
	// Property: a refused address never carries the user info it was given
	// (ADR-0013). The mutated field is PROVEN_COUCHDB_URL.
	if strings.Contains(err.Error(), "u:p") || strings.Contains(err.Error(), "p@") || strings.Contains(cfg.Value, "p@") {
		t.Fatalf("the refusal printed the user info: %v", err)
	}
}

// TestEnvironmentListsEveryCouchSetting proves the settings page rows hold
// each CouchDB variable and mark only the password secret (ADR-0013). The
// mutated field is the Secret flag.
func TestEnvironmentListsEveryCouchSetting(t *testing.T) {
	secret := map[string]bool{}
	for _, row := range Environment() {
		secret[row.Name] = row.Secret
	}
	for _, name := range []string{EnvCouchURL, EnvCouchPrefix, EnvCouchUser, EnvCouchPassword, EnvCouchRowLimit, EnvCouchTimeout} {
		s, ok := secret[name]
		if !ok || s != (name == EnvCouchPassword) {
			t.Fatalf("%s: listed %v, secret %v", name, ok, s)
		}
	}
}

// TestLoadLocalReadsTheFileAndRefusesABrokenPath proves the file applies
// only to the local store, an empty path is refused by name (there is no
// default), and a path with a newline is refused by name. The mutated field
// is PROVEN_EVIDENCE_FILE; negative control: restoring a default turns the
// empty case red.
func TestLoadLocalReadsTheFileAndRefusesABrokenPath(t *testing.T) {
	// ADR-0047: there is no default file; empty and blank are refused by name.
	for _, blank := range []string{"", "  "} {
		_, err := Load(func(k string) string {
			switch k {
			case EnvStore:
				return NameLocal
			case EnvFile:
				return blank
			}
			return ""
		})
		var refused *ConfigError
		if err == nil || !asConfig(err, &refused) || refused.Field != EnvFile || !strings.Contains(err.Error(), "required") {
			t.Fatalf("%q: %v", blank, err)
		}
	}
	got, err := Load(func(k string) string {
		switch k {
		case EnvStore:
			return NameLocal
		case EnvFile:
			return "room/evidence.sql"
		default:
			return ""
		}
	})
	if err != nil || got.File != "room/evidence.sql" {
		t.Fatalf("%#v %v", got, err)
	}
	_, err = Load(func(k string) string {
		switch k {
		case EnvStore:
			return NameLocal
		case EnvFile:
			return "evidence.sql\n"
		default:
			return ""
		}
	})
	var cfg *ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvFile {
		t.Fatal(err)
	}
	// A zetaris review does not read a broken file path. Mutated field: the store name.
	got, err = Load(func(k string) string {
		if k == EnvFile {
			return "evidence.sql\n"
		}
		return ""
	})
	if err != nil || got.Name != NameZetaris {
		t.Fatalf("%#v %v", got, err)
	}
}

func asConfig(err error, dst **ConfigError) bool {
	cfg, ok := err.(*ConfigError)
	if ok {
		*dst = cfg
	}
	return ok
}
