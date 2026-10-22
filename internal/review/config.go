package review

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/change"
)

const configKeyTimeout = "timeout"

// allowedTimeout is the value list a refused timeout names.
const allowedTimeout = "a Go duration of zero or greater, for example 30s, 5m or 0s"

// configKeys is the closed key set, in the order a refusal lists them.
var configKeys = []string{
	"timeout",
	"default_reviewer",
	"special_reviewer",
	"families",
	"path_rules",
	"class_order",
	"fallback_class",
	"doubled_classes",
	"author_commit_trailers",
	"author_description",
	"author_record",
	"author_trailer_keys",
	"raise_class",
}

// fileSettings is one loaded proven.yaml. A missing file is defaultSettings.
// A key that is set replaces that setting; it does not merge with the default list.
type fileSettings struct {
	Timeout         time.Duration
	DefaultReviewer string
	SpecialReviewer string
	Families        []familyRule
	PathRules       []pathRule
	ClassOrder      []string
	Fallback        string
	Doubled         map[string]bool
	ReadCommits     bool
	ReadDescription bool
	Record          string
	TrailerKeys     []string
	RaiseClass      bool
}

type familyRule struct {
	Pattern string
	Family  string
}

type pathRule struct {
	Pattern string
	Class   string
}

func classNames() []string {
	return []string{"security", "backend", "mobile", "user interface", "docs"}
}

func familyNames() []string {
	return []string{"Anthropic", "OpenAI", "NVIDIA", "Alibaba"}
}

func defaultSettings() fileSettings {
	return fileSettings{
		DefaultReviewer: "NVIDIA",
		// NVIDIA is an exact name because the default reviewer is the word NVIDIA,
		// and the ADR-0011 globs (nemotron-*) do not match that word.
		Families: []familyRule{
			{"claude-*", "Anthropic"},
			{"gpt-*", "OpenAI"},
			{"codex-*", "OpenAI"},
			{"nemotron-*", "NVIDIA"},
			{"qwen*", "Alibaba"},
			{"NVIDIA", "NVIDIA"},
		},
		PathRules:       defaultPathRules(),
		ClassOrder:      classNames(),
		Fallback:        "backend",
		Doubled:         map[string]bool{"security": true},
		ReadCommits:     true,
		ReadDescription: true,
		TrailerKeys:     []string{"Assisted-by"},
	}
}

// defaultPathRules ships ADR-0014's examples. Dependency manifests are the
// lockfiles and module files a change commonly touches. A pattern with no
// slash also matches the base name, so go.mod matches a nested module.
func defaultPathRules() []pathRule {
	sec := []string{
		"**/auth/**",
		"**/payment*/**",
		"**/*secret*",
		"**/migrations/**",
		"go.mod",
		"go.sum",
		"package.json",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"Cargo.toml",
		"Cargo.lock",
		"requirements.txt",
		"Pipfile.lock",
		"poetry.lock",
		"Gemfile.lock",
		"composer.lock",
		"pom.xml",
		"build.gradle",
		"build.gradle.kts",
	}
	out := make([]pathRule, 0, len(sec)+2)
	for _, p := range sec {
		out = append(out, pathRule{Pattern: p, Class: "security"})
	}
	out = append(out, pathRule{Pattern: "**/*.md", Class: "docs"})
	out = append(out, pathRule{Pattern: "docs/**", Class: "docs"})
	return out
}

// loadConfigFile reads an explicit configuration path. A missing path is refused.
func loadConfigFile(path string) (fileSettings, error) {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileSettings{}, &ConfigError{
				Field: "ConfigPath", Value: path, Reason: "file does not exist",
				Allowed: []string{"a path to an existing file"},
			}
		}
		return fileSettings{}, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: err.Error(),
			Allowed: []string{"a path to an existing file"},
		}
	}
	if st.IsDir() {
		return fileSettings{}, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: "path is a directory",
			Allowed: []string{"a path to an existing file"},
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return fileSettings{}, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: err.Error(),
			Allowed: []string{"a path to an existing file"},
		}
	}
	return parseConfig(b)
}

// loadSettingsAtBaseline reads proven.yaml as it is in the baseline commit.
// The change's copy is not consulted. A change must not choose how strictly
// it is reviewed. A baseline with no proven.yaml is the defaults.
func loadSettingsAtBaseline(target, baseline string) (fileSettings, error) {
	b, ok, err := change.Blob(target, baseline, "proven.yaml")
	if err != nil {
		return fileSettings{}, err
	}
	if !ok {
		return defaultSettings(), nil
	}
	return parseConfig(b)
}

// loadSettingsInTarget reads proven.yaml in the review target when Open was
// given no path and the request has no baseline commit. A missing file is the defaults.
func loadSettingsInTarget(target string) (fileSettings, error) {
	path := filepath.Join(target, "proven.yaml")
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSettings(), nil
		}
		return fileSettings{}, err
	}
	if st.IsDir() {
		return fileSettings{}, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: "path is a directory",
			Allowed: []string{"a path to an existing file"},
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return fileSettings{}, err
	}
	return parseConfig(b)
}

// parseConfig reads one flat YAML mapping. Lists are comma-separated scalars.
// An empty file is the defaults. Comments, quotes and one document are read.
// A nested value is refused. An unknown key is refused.
func parseConfig(b []byte) (fileSettings, error) {
	cfg := defaultSettings()
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(b)) == 0 {
		return cfg, nil
	}
	seen := map[string]bool{}
	seenDoc := false
	for _, raw := range bytes.Split(b, []byte("\n")) {
		line := strings.TrimRight(string(raw), "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "---" {
			if seenDoc {
				return fileSettings{}, configRefused("proven.yaml", trimmed, "more than one document", "one flat mapping")
			}
			seenDoc = true
			continue
		}
		if trimmed == "..." {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			return fileSettings{}, configRefused("proven.yaml", trimmed, "nested value", "a flat mapping of scalar values")
		}
		if strings.HasPrefix(trimmed, "- ") || trimmed == "-" {
			return fileSettings{}, configRefused("proven.yaml", trimmed, "the file must be a mapping", "a flat mapping of scalar values")
		}
		key, val, err := splitField(trimmed)
		if err != nil {
			return fileSettings{}, err
		}
		if !knownKey(key) {
			return fileSettings{}, &ConfigError{
				Field: key, Value: val, Reason: "unknown key",
				Allowed: append([]string(nil), configKeys...),
			}
		}
		if seen[key] {
			return fileSettings{}, configRefused(key, val, "duplicate key", allowedText(key))
		}
		seen[key] = true
		if err := applyKey(&cfg, key, val); err != nil {
			return fileSettings{}, err
		}
	}
	return cfg, nil
}

func knownKey(key string) bool {
	for _, k := range configKeys {
		if k == key {
			return true
		}
	}
	return false
}

func allowedText(key string) string {
	switch key {
	case configKeyTimeout:
		return allowedTimeout
	case "author_commit_trailers", "author_description", "raise_class":
		return "on, off"
	case "fallback_class", "doubled_classes", "class_order", "path_rules":
		return strings.Join(classNames(), ", ")
	case "families":
		return strings.Join(familyNames(), ", ")
	case "author_record":
		return "off, or a relative path or pattern inside the review target"
	case "default_reviewer", "special_reviewer":
		return "a model name"
	case "author_trailer_keys":
		return "trailer keys, for example Assisted-by"
	default:
		return strings.Join(configKeys, ", ")
	}
}

func applyKey(cfg *fileSettings, key, val string) error {
	switch key {
	case configKeyTimeout:
		if val == "" {
			cfg.Timeout = 0
			return nil
		}
		d, err := time.ParseDuration(val)
		if err != nil || d < 0 {
			return configRefused(configKeyTimeout, val, "not a Go duration", allowedTimeout)
		}
		cfg.Timeout = d
		return nil
	case "default_reviewer":
		if val == "" {
			return configRefused(key, val, "missing value", "a model name, for example NVIDIA")
		}
		cfg.DefaultReviewer = val
		return nil
	case "special_reviewer":
		cfg.SpecialReviewer = val
		return nil
	case "families":
		rules, err := parseFamilies(val)
		if err != nil {
			return err
		}
		cfg.Families = rules
		return nil
	case "path_rules":
		rules, err := parsePathRules(val)
		if err != nil {
			return err
		}
		cfg.PathRules = rules
		return nil
	case "class_order":
		order, err := parseClassOrder(val)
		if err != nil {
			return err
		}
		cfg.ClassOrder = order
		return nil
	case "fallback_class":
		if !knownClass(val) {
			return &ConfigError{Field: key, Value: val, Reason: "unknown class", Allowed: classNames()}
		}
		cfg.Fallback = val
		return nil
	case "doubled_classes":
		if val == "off" {
			cfg.Doubled = map[string]bool{}
			return nil
		}
		parts, err := parseClassTokens(key, val)
		if err != nil {
			return err
		}
		m := map[string]bool{}
		for _, p := range parts {
			m[p] = true
		}
		cfg.Doubled = m
		return nil
	case "author_commit_trailers":
		on, err := parseOnOff(key, val)
		if err != nil {
			return err
		}
		cfg.ReadCommits = on
		return nil
	case "author_description":
		on, err := parseOnOff(key, val)
		if err != nil {
			return err
		}
		cfg.ReadDescription = on
		return nil
	case "author_record":
		rec, err := parseRecord(val)
		if err != nil {
			return err
		}
		cfg.Record = rec
		return nil
	case "author_trailer_keys":
		keys, err := parseTrailerKeys(val)
		if err != nil {
			return err
		}
		cfg.TrailerKeys = keys
		return nil
	case "raise_class":
		on, err := parseOnOff(key, val)
		if err != nil {
			return err
		}
		cfg.RaiseClass = on
		return nil
	default:
		return &ConfigError{Field: key, Value: val, Reason: "unknown key", Allowed: append([]string(nil), configKeys...)}
	}
}

func parseOnOff(field, value string) (bool, error) {
	switch value {
	case "on":
		return true, nil
	case "off":
		return false, nil
	default:
		return false, &ConfigError{Field: field, Value: value, Reason: "not on or off", Allowed: []string{"on", "off"}}
	}
}

func parseFamilies(value string) ([]familyRule, error) {
	pairs, err := parsePairs("families", value)
	if err != nil {
		return nil, err
	}
	out := make([]familyRule, 0, len(pairs))
	for _, p := range pairs {
		if err := validPattern(p[0]); err != nil {
			return nil, &ConfigError{Field: "families", Value: p[0], Reason: "bad pattern", Allowed: []string{"a model-name pattern, for example claude-*"}}
		}
		if !knownFamily(p[1]) {
			return nil, &ConfigError{Field: "families", Value: p[1], Reason: "unknown family", Allowed: familyNames()}
		}
		out = append(out, familyRule{Pattern: p[0], Family: p[1]})
	}
	return out, nil
}

func parsePathRules(value string) ([]pathRule, error) {
	pairs, err := parsePairs("path_rules", value)
	if err != nil {
		return nil, err
	}
	out := make([]pathRule, 0, len(pairs))
	for _, p := range pairs {
		if err := validPattern(p[0]); err != nil {
			return nil, &ConfigError{Field: "path_rules", Value: p[0], Reason: "bad pattern", Allowed: []string{"a path pattern using * and **"}}
		}
		if !knownClass(p[1]) {
			return nil, &ConfigError{Field: "path_rules", Value: p[1], Reason: "unknown class", Allowed: classNames()}
		}
		out = append(out, pathRule{Pattern: p[0], Class: p[1]})
	}
	return out, nil
}

func parsePairs(field, value string) ([][2]string, error) {
	items, err := splitCSV(field, value)
	if err != nil {
		return nil, err
	}
	out := make([][2]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		i := strings.IndexByte(item, '=')
		if i <= 0 || i == len(item)-1 {
			return nil, configRefused(field, item, "want pattern=value", "pattern=value")
		}
		left := strings.TrimSpace(item[:i])
		right := strings.TrimSpace(item[i+1:])
		if left == "" || right == "" {
			return nil, configRefused(field, item, "want pattern=value", "pattern=value")
		}
		if _, ok := seen[left]; ok {
			return nil, configRefused(field, left, "duplicate pattern", "each pattern once")
		}
		seen[left] = struct{}{}
		out = append(out, [2]string{left, right})
	}
	return out, nil
}

func parseClassOrder(value string) ([]string, error) {
	parts, err := parseClassTokens("class_order", value)
	if err != nil {
		return nil, err
	}
	have := map[string]struct{}{}
	for _, p := range parts {
		have[p] = struct{}{}
	}
	var miss []string
	for _, c := range classNames() {
		if _, ok := have[c]; !ok {
			miss = append(miss, c)
		}
	}
	if len(miss) > 0 {
		return nil, &ConfigError{
			Field: "class_order", Value: value,
			Reason:  "leaves out " + strings.Join(miss, ", "),
			Allowed: classNames(),
		}
	}
	return parts, nil
}

func parseClassTokens(field, value string) ([]string, error) {
	parts, err := splitCSV(field, value)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	for _, p := range parts {
		if !knownClass(p) {
			return nil, &ConfigError{Field: field, Value: p, Reason: "unknown class", Allowed: classNames()}
		}
		if _, ok := seen[p]; ok {
			return nil, &ConfigError{Field: field, Value: p, Reason: "duplicate class", Allowed: classNames()}
		}
		seen[p] = struct{}{}
	}
	return parts, nil
}

func parseTrailerKeys(value string) ([]string, error) {
	parts, err := splitCSV("author_trailer_keys", value)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(strings.TrimSuffix(p, ":"))
		if p == "" || strings.Contains(p, ":") {
			return nil, configRefused("author_trailer_keys", p, "bad trailer key", "a key such as Assisted-by")
		}
		low := strings.ToLower(p)
		if _, ok := seen[low]; ok {
			return nil, configRefused("author_trailer_keys", p, "duplicate key", "each trailer key once")
		}
		seen[low] = struct{}{}
		out = append(out, p)
	}
	return out, nil
}

func parseRecord(value string) (string, error) {
	const allowed = "off, or a relative path or pattern inside the review target"
	if value == "" || value == "off" {
		return "", nil
	}
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "\\") {
		return "", &ConfigError{Field: "author_record", Value: value, Reason: "absolute path", Allowed: []string{allowed}}
	}
	for _, seg := range strings.Split(strings.ReplaceAll(value, "\\", "/"), "/") {
		if seg == ".." {
			return "", &ConfigError{Field: "author_record", Value: value, Reason: "path leaves the review target", Allowed: []string{allowed}}
		}
	}
	if err := validPattern(value); err != nil {
		return "", &ConfigError{Field: "author_record", Value: value, Reason: "bad pattern", Allowed: []string{allowed}}
	}
	return value, nil
}

func splitCSV(field, value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, configRefused(field, value, "missing value", "a comma-separated list")
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, configRefused(field, value, "empty list item", "a comma-separated list")
		}
		out = append(out, p)
	}
	return out, nil
}

func knownClass(name string) bool {
	for _, c := range classNames() {
		if c == name {
			return true
		}
	}
	return false
}

func knownFamily(name string) bool {
	for _, f := range familyNames() {
		if f == name {
			return true
		}
	}
	return false
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
		return "", "", configRefused("proven.yaml", line, "empty key", strings.Join(configKeys, ", "))
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
