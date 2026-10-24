package store

import (
	"strings"
	"testing"
)

// TestLoadDefaultsToZetaris proves an empty PROVEN_EVIDENCE_STORE is
// zetaris and does not require a file. The mutated field is the variable:
// empty, then couchdb, then a word that is not a store.
func TestLoadDefaultsToZetaris(t *testing.T) {
	got, err := Load(func(string) string { return "" })
	if err != nil || got.Name != NameZetaris || got.File != "" {
		t.Fatalf("%#v %v", got, err)
	}
	_, err = Load(func(k string) string {
		if k == EnvStore {
			return NameCouchDB
		}
		return ""
	})
	var cfg *ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvStore || cfg.Value != NameCouchDB {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "not built") || !strings.Contains(err.Error(), "zetaris") {
		t.Fatal(err)
	}
	_, err = Load(func(k string) string {
		if k == EnvStore {
			return "postgres"
		}
		return ""
	})
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvStore || cfg.Value != "postgres" {
		t.Fatal(err)
	}
}

// TestLoadLocalReadsTheFileAndRefusesABrokenPath proves the file applies
// only to the local store, an empty path takes the default, and a path
// with a newline is refused by name. The mutated field is PROVEN_EVIDENCE_FILE.
func TestLoadLocalReadsTheFileAndRefusesABrokenPath(t *testing.T) {
	got, err := Load(func(k string) string {
		if k == EnvStore {
			return NameLocal
		}
		return ""
	})
	if err != nil || got.Name != NameLocal || got.File != defaultFile {
		t.Fatalf("%#v %v", got, err)
	}
	got, err = Load(func(k string) string {
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
