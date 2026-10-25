package review

// This file reads proven.yaml, the reviewer's own settings (ADR-0013): the timeout, the two reviewer
// models (ADR-0009, ADR-0010), the family table and author sources (ADR-0011), and the class rules
// (ADR-0014), the two settings of proposed fixes: whether they are made and how many attempts one
// gets (ADR-0042), and Scout Intent's switch and its plan for each kind of change (ADR-0049). Every key
// is closed: an unknown key or value is refused with the values allowed.

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/change"
)

// configKeyTimeout is the timeout key, named once because the refusal text is keyed on it too.
const configKeyTimeout = "timeout"

// allowedTimeout is the value list a refused timeout names.
const allowedTimeout = "a Go duration of zero or greater, for example 30s, 5m or 0s"

// The bounds of fix_attempts (ADR-0042 point 6). The default is two: one try and one retry with the
// failure in hand. The upper bound keeps a setting from turning a retry into an open loop of model calls.
const (
	defaultFixAttempts = 2
	maxFixAttempts     = 5
)

// allowedFixAttempts is the value list a refused fix_attempts names.
const allowedFixAttempts = "a whole number from 1 to 5"

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
	"fixes",
	"fix_attempts",
	"scout_intent",
	"plan_dependency",
	"plan_refactor",
	"plan_feature",
	"plan_security",
	"plan_docs",
	"plan_config",
	"prose_suffixes",
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
	// DoubledSet is true when proven.yaml named doubled_classes. With a second reviewer connected,
	// an unset list means every class is doubled, and a set one narrows it (ADR-0037 point 3).
	DoubledSet      bool
	ReadCommits     bool
	ReadDescription bool
	Record          string
	TrailerKeys     []string
	RaiseClass      bool
	// Fixes turns proposed fixes on or off (ADR-0042). FixAttempts is the most attempts at one fix.
	Fixes       bool
	FixAttempts int
	// ScoutIntent switches the review plan on (ADR-0049). Off, every review runs the full plan.
	ScoutIntent bool
	// Plans is the plan for each kind of change, keyed by the kind's registry id (ADR-0049 point 2).
	Plans map[string]planSetting
	// ProseSuffixes is the closed list of file suffixes that count as prose never built or executed
	// (ADR-0049). A review plans as documentation-only, and runs no model, only when every changed file
	// has one of these suffixes and is not executable.
	ProseSuffixes []string
}

// familyRule maps a model-name pattern to a family, one row of the families table (ADR-0011 point 4).
type familyRule struct {
	Pattern string
	Family  string
}

// pathRule maps a changed-path pattern to a review class (ADR-0014 point 2).
type pathRule struct {
	Pattern string
	Class   string
}

// classNames is the closed list of review classes (ADR-0014 point 7), in the default order of risk,
// highest first. It is a function, not a variable, so no caller can change the shared list.
func classNames() []string {
	return []string{"security", "backend", "mobile", "user interface", "docs"}
}

// defaultSettings is what Proven uses with no proven.yaml. Each value can be replaced by a named
// key (ADR-0013): NVIDIA reviews by default (ADR-0008 point 3), security is doubled
// (ADR-0009), backend is the fallback class (ADR-0014 point 4), and Assisted-by is the trailer key
// (ADR-0011 point 1). The timeout is zero: no deadline unless the shell or the file sets one.
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
		// ADR-0042: proposed fixes are on by default, with one try and one retry.
		Fixes:       true,
		FixAttempts: defaultFixAttempts,
		// ADR-0049: the engine lands switched off; it is switched on only once its measurement holds.
		ScoutIntent: false,
		Plans:       defaultPlans(),
		// ADR-0049: only plain prose may skip the model. Anything else, an unknown suffix included, is
		// not known to be inert, and a file that is built or executed must be reviewed.
		ProseSuffixes: []string{".md", ".markdown", ".rst", ".txt", ".adoc"},
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
		"composer.json",
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
	// ADR-0013: the file is read no further than the configuration bound.
	b, err := change.ReadBounded(path, change.Bounds().Config, "the configuration file "+filepath.Base(path), change.EnvMaxConfigBytes)
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
	// Blob is bounded by the file limit; proven.yaml has a smaller bound of its own.
	if bound := change.Bounds().Config; int64(len(b)) > bound {
		return fileSettings{}, change.OverBound("proven.yaml", bound, change.EnvMaxConfigBytes)
	}
	return parseConfig(b)
}

// loadSettingsInTarget reads proven.yaml in the review target when Open was
// given no path and the request has no baseline commit. A missing file is the defaults.
func loadSettingsInTarget(target string) (fileSettings, error) {
	path := filepath.Join(target, "proven.yaml")
	// ADR-0013, ADR-0025: the target's copy is repository content, so a link is refused by name
	// before it is opened. Stat and ReadFile would follow it to a file outside the checkout, and
	// the refusal never quotes what the link points at.
	st, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSettings(), nil
		}
		return fileSettings{}, err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		return fileSettings{}, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: "proven.yaml is a symbolic link; Proven reads only a regular file inside the checkout",
			Allowed: []string{"a regular file named proven.yaml"},
		}
	}
	if st.IsDir() {
		return fileSettings{}, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: "path is a directory",
			Allowed: []string{"a path to an existing file"},
		}
	}
	// ADR-0013: the target's copy is repository content, so it is bounded.
	b, err := change.ReadBounded(path, change.Bounds().Config, "proven.yaml", change.EnvMaxConfigBytes)
	if err != nil {
		return fileSettings{}, err
	}
	return parseConfig(b)
}

// parseConfig reads one flat YAML mapping. It is a small reader of its own, not a YAML library,
// because the file is one flat mapping and anything richer must be refused, not interpreted. Lists are comma-separated scalars.
// An empty file is the defaults. Comments, quotes and one document are read.
// A nested value is refused. An unknown key is refused.
func parseConfig(b []byte) (fileSettings, error) {
	cfg := defaultSettings()
	// An editor may save a UTF-8 byte order mark; it is not part of the first key.
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
		// "..." ends a YAML document; with one document allowed, nothing follows that matters.
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
		// A key set twice is refused, so a reader of the file never has to guess which line wins.
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

// knownKey reports whether key is in the closed key set.
func knownKey(key string) bool {
	for _, k := range configKeys {
		if k == key {
			return true
		}
	}
	return false
}

// allowedText is the allowed-values text a refusal of key names (ADR-0013).
func allowedText(key string) string {
	switch key {
	case configKeyTimeout:
		return allowedTimeout
	case "author_commit_trailers", "author_description", "raise_class", "fixes":
		return "on, off"
	case "fix_attempts":
		return allowedFixAttempts
	case "scout_intent":
		return "on, off"
	case "plan_dependency", "plan_refactor", "plan_feature", "plan_security", "plan_docs", "plan_config":
		return allowedPlan()
	case "fallback_class", "doubled_classes", "class_order", "path_rules":
		return strings.Join(classNames(), ", ")
	case "families":
		return "pattern=name, for example claude-*=Anthropic"
	case "author_record":
		return "off, or a relative path or pattern inside the review target"
	case "default_reviewer", "special_reviewer":
		return "a model name"
	case "author_trailer_keys":
		return "trailer keys, for example Assisted-by"
	case "prose_suffixes":
		return allowedProseSuffixes
	default:
		return strings.Join(configKeys, ", ")
	}
}

// applyKey checks one value and sets it on cfg. A key that is set replaces the default; lists are
// never merged with the default list, so the file says the whole setting.
func applyKey(cfg *fileSettings, key, val string) error {
	switch key {
	case configKeyTimeout:
		// An empty timeout means no deadline from the file, the same as 0s.
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
		// Every review uses the default reviewer (ADR-0009 point 2), so it may not be empty.
		if val == "" {
			return configRefused(key, val, "missing value", "a model name, for example NVIDIA")
		}
		cfg.DefaultReviewer = val
		return nil
	case "special_reviewer":
		// Empty is allowed: no special model, and a doubled class says so in the review.
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
		// "off" doubles no class; an empty list would be refused, so off is the way to say none.
		cfg.DoubledSet = true
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
	case "prose_suffixes":
		sfx, err := parseProseSuffixes(val)
		if err != nil {
			return err
		}
		cfg.ProseSuffixes = sfx
		return nil
	case "raise_class":
		on, err := parseOnOff(key, val)
		if err != nil {
			return err
		}
		// ADR-0014 point 9: the setting is checked and kept, but no model raises a class yet.
		cfg.RaiseClass = on
		return nil
	case "fixes":
		on, err := parseOnOff(key, val)
		if err != nil {
			return err
		}
		cfg.Fixes = on
		return nil
	case "fix_attempts":
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 || n > maxFixAttempts {
			return configRefused(key, val, "not a number of attempts in range", allowedFixAttempts)
		}
		cfg.FixAttempts = n
		return nil
	case "scout_intent":
		on, err := parseOnOff(key, val)
		if err != nil {
			return err
		}
		cfg.ScoutIntent = on
		return nil
	case "plan_dependency", "plan_refactor", "plan_feature", "plan_security", "plan_docs", "plan_config":
		ps, err := parsePlanSetting(key, val)
		if err != nil {
			return err
		}
		// The map is copied before the change, so the shared defaults are never written to.
		plans := make(map[string]planSetting, len(cfg.Plans))
		for k, v := range cfg.Plans {
			plans[k] = v
		}
		plans[planKind(key)] = ps
		cfg.Plans = plans
		return nil
	default:
		return &ConfigError{Field: key, Value: val, Reason: "unknown key", Allowed: append([]string(nil), configKeys...)}
	}
}

// parseOnOff reads a switch. Only "on" and "off" are accepted, not YAML's yes, true or 1, so a
// switch reads one way in every file.
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

// parseFamilies reads "pattern=family, ..." into the family table, in the file's order: the first
// matching row wins in familyOf.
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
		out = append(out, familyRule{Pattern: p[0], Family: p[1]})
	}
	return out, nil
}

// parsePathRules reads "pattern=class, ...". ADR-0014 point 7: a rule that names a class not on
// the closed list is refused when the file is loaded.
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

// parsePairs splits a comma list of pattern=value items, at the first "=". A pattern given twice
// is refused, because only one of its values could ever apply.
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

// parseClassOrder reads the order of risk, highest first. ADR-0014 point 7: an order that leaves a
// class out is refused, and the refusal names the classes it left out.
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

// parseClassTokens reads a comma list of class names; each must be a known class, named once.
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

// parseTrailerKeys reads the trailer keys that name an author model (ADR-0011 point 3). A trailing
// colon is dropped, so "Assisted-by:" and "Assisted-by" are the same key; keys compare without case.
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

// parseRecord checks the author record setting on load. The record must lie inside the review
// target: an absolute path or a ".." step is refused here, and recordPaths checks links when it reads.
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

// splitCSV splits a comma list and trims each item. An empty list or an empty item is refused,
// so a stray comma never becomes a rule that matches nothing.
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

// knownClass reports whether name is on the closed class list.
func knownClass(name string) bool {
	for _, c := range classNames() {
		if c == name {
			return true
		}
	}
	return false
}

// configRefused is the *ConfigError with one allowed-values text.
func configRefused(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

// splitField splits one "key: value" line into the key and the unquoted value.
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

// takeKey reads the key, plain or quoted, and returns the rest of the line from the colon on.
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

// takeValue reads one scalar value. YAML's null spellings read as empty. A flow list, a flow map
// or a block scalar is refused as nested. An unquoted value ends at " #", where a comment starts.
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

// unquote reads one quoted scalar and returns it with the text after the closing quote. Inside
// double quotes it reads only the escapes \\, \", \n and \t; any other escape is refused rather than
// guessed. Single quotes take every character as it is.
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
