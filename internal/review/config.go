package review

import (
	"bytes"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const configKeyTimeout = "timeout"

// allowedTimeout is the value list a refused timeout names.
const allowedTimeout = "a Go duration of zero or greater, for example 30s, 5m or 0s"

// loadConfigFile reads an explicit configuration path. A missing path is refused.
func loadConfigFile(path string) (time.Duration, error) {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, &ConfigError{
				Field: "ConfigPath", Value: path, Reason: "file does not exist",
				Allowed: []string{"a path to an existing file"},
			}
		}
		return 0, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: err.Error(),
			Allowed: []string{"a path to an existing file"},
		}
	}
	if st.IsDir() {
		return 0, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: "path is a directory",
			Allowed: []string{"a path to an existing file"},
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: err.Error(),
			Allowed: []string{"a path to an existing file"},
		}
	}
	return parseConfig(b)
}

// parseConfig reads one flat YAML mapping. The only key is timeout.
// An empty file is the defaults. Comments, quotes and one document are read.
// A nested value is refused.
func parseConfig(b []byte) (time.Duration, error) {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(b)) == 0 {
		return 0, nil
	}
	var timeout string
	seenTimeout := false
	seenDoc := false
	for _, raw := range bytes.Split(b, []byte("\n")) {
		line := strings.TrimRight(string(raw), "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "---" {
			if seenDoc {
				return 0, configRefused("proven.yaml", trimmed, "more than one document", "one flat mapping")
			}
			seenDoc = true
			continue
		}
		if trimmed == "..." {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			return 0, configRefused("proven.yaml", trimmed, "nested value", "a flat mapping with key timeout")
		}
		if strings.HasPrefix(trimmed, "- ") || trimmed == "-" {
			return 0, configRefused("proven.yaml", trimmed, "the file must be a mapping", "a flat mapping with key timeout")
		}
		key, val, err := splitField(trimmed)
		if err != nil {
			return 0, err
		}
		if key != configKeyTimeout {
			return 0, &ConfigError{
				Field: key, Value: val, Reason: "unknown key",
				Allowed: []string{configKeyTimeout},
			}
		}
		if seenTimeout {
			return 0, configRefused(configKeyTimeout, val, "duplicate key", allowedTimeout)
		}
		seenTimeout = true
		timeout = val
	}
	if !seenTimeout || timeout == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(timeout)
	if err != nil || d < 0 {
		return 0, configRefused(configKeyTimeout, timeout, "not a Go duration", allowedTimeout)
	}
	return d, nil
}

func configRefused(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

func splitField(line string) (string, string, error) {
	key, rest, err := takeKey(line)
	if err != nil {
		return "", "", err
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", "", configRefused("proven.yaml", line, "missing value", "key: value")
	}
	if !strings.HasPrefix(rest, ":") {
		return "", "", configRefused("proven.yaml", line, "missing colon", "key: value")
	}
	val, err := takeValue(strings.TrimSpace(rest[1:]))
	if err != nil {
		return "", "", err
	}
	return key, val, nil
}

func takeKey(line string) (string, string, error) {
	if line == "" {
		return "", "", configRefused("proven.yaml", line, "empty key", configKeyTimeout)
	}
	if line[0] == '"' || line[0] == '\'' {
		key, rest, err := unquote(line)
		if err != nil {
			return "", "", err
		}
		return key, rest, nil
	}
	i := strings.IndexByte(line, ':')
	if i <= 0 {
		return "", "", configRefused("proven.yaml", line, "missing colon", "key: value")
	}
	return strings.TrimSpace(line[:i]), line[i:], nil
}

func takeValue(s string) (string, error) {
	if s == "" || s == "~" || s == "null" || s == "Null" || s == "NULL" {
		return "", nil
	}
	switch s[0] {
	case '"', '\'':
		v, rest, err := unquote(s)
		if err != nil {
			return "", err
		}
		rest = strings.TrimSpace(rest)
		if rest != "" && !strings.HasPrefix(rest, "#") {
			return "", configRefused("proven.yaml", s, "trailing junk", "a scalar value")
		}
		return v, nil
	case '[', '{', '|', '>':
		return "", configRefused("proven.yaml", s, "nested value", "a scalar value")
	}
	if i := strings.Index(s, " #"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if s == "" {
		return "", nil
	}
	return s, nil
}

func unquote(s string) (string, string, error) {
	if s == "" {
		return "", "", configRefused("proven.yaml", s, "empty quote", "a quoted scalar")
	}
	q := s[0]
	if q != '"' && q != '\'' {
		return "", "", configRefused("proven.yaml", s, "not a quoted scalar", "a quoted scalar")
	}
	var b strings.Builder
	for i := 1; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return "", "", configRefused("proven.yaml", s, "invalid text", "utf-8")
		}
		if r == rune(q) {
			return b.String(), s[i+size:], nil
		}
		if q == '"' && r == '\\' {
			if i+size >= len(s) {
				return "", "", configRefused("proven.yaml", s, "broken escape", "a quoted scalar")
			}
			esc, esz := utf8.DecodeRuneInString(s[i+size:])
			switch esc {
			case '\\', '"':
				b.WriteRune(esc)
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				return "", "", configRefused("proven.yaml", s, "broken escape", "a quoted scalar")
			}
			i += size + esz
			continue
		}
		b.WriteRune(r)
		i += size
	}
	return "", "", configRefused("proven.yaml", s, "unclosed quote", "a quoted scalar")
}
