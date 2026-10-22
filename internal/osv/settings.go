package osv

import "strings"

// EnvTable is the environment variable that names the OSV package table.
// Sample mode does not read it.
const EnvTable = "PROVEN_OSV_TABLE"

// DefaultTable is the table a review queries when EnvTable is empty.
const DefaultTable = "px_osvfile.px_osv_pkg"

// ConfigError names a refused OSV setting and the values that are allowed.
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
	msg := "configuration refused: " + e.Field + "=" + quote(e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

func quote(s string) string {
	return `"` + s + `"`
}

// LoadTable reads EnvTable. An empty value is DefaultTable.
func LoadTable(env func(string) string) (string, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvTable))
	if raw == "" {
		return DefaultTable, nil
	}
	if !ValidTable(raw) {
		return "", &ConfigError{
			Field: EnvTable, Value: raw, Reason: "not a table name",
			Allowed: []string{"a table or schema.table of letters, digits and underscores; empty means " + DefaultTable},
		}
	}
	return raw, nil
}

// ValidTable reports whether name can be placed in a FROM clause.
// One identifier, or schema.table. Nothing else.
func ValidTable(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) < 1 || len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if !ident(p) {
			return false
		}
	}
	return true
}

func ident(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c == '_':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}
