package review

// This file reads the organisation policy and its exceptions (ADR-0036). The policy is one file in the
// flat format of proven.yaml, named by the setting PROVEN_POLICY. It is read before a repository's
// proven.yaml: a repository may make a policy setting stricter, never weaker, and a weaker local value
// is refused on load. An exception accepts one known finding until an expiry date, with an owner and a
// reason, and lives in the policy file only.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/envdoc"
)

// EnvPolicy is the setting that names the organisation policy file: a local path. Empty means no policy.
const EnvPolicy = "PROVEN_POLICY"

// exceptionPrefix starts every exception key: exception.<id>.<field>.
const exceptionPrefix = "exception."

// dateLayout is the one date format an expiry is written in.
const dateLayout = "2006-01-02"

// policyKeys are the proven.yaml keys a policy may set, in the order a refusal lists them. A key is here
// only when two of its values can be put in order, stricter first (docs/POLICY.md, the stricter-than
// table). The rest have no such order and a policy that sets one is refused.
var policyKeys = []string{
	"path_rules",
	"class_order",
	"fallback_class",
	"doubled_classes",
	"author_commit_trailers",
	"author_description",
	"author_trailer_keys",
	"raise_class",
}

// exceptionFields are the five fields every exception names, in the order a refusal lists them.
var exceptionFields = []string{"rule", "location", "reason", "owner", "expires"}

// exception accepts one known finding until Expires (ADR-0036 point 2).
type exception struct {
	ID      string
	Rule    string // the finding's rule id
	Path    string // slash-separated
	Line    int    // 0 means any line of the path
	Where   string // the location as written, for the review's own sentences
	Reason  string
	Owner   string
	Expires string // YYYY-MM-DD, the last day the exception holds
}

// policy is one loaded organisation policy. A nil *policy is no policy, and its methods say so.
type policy struct {
	path       string
	settings   fileSettings      // the defaults, then the policy's own keys
	given      map[string]string // the raw value of each key the policy set
	exceptions []exception
}

// EnvironmentPolicy lists the policy setting for the configuration page.
func EnvironmentPolicy() []envdoc.Setting {
	return []envdoc.Setting{{
		Name:        EnvPolicy,
		Default:     "none",
		Description: "a path to the organisation policy file, read before each repository's `proven.yaml`; see [POLICY.md](POLICY.md). A path inside the repository under review is refused, and so is a missing or unreadable file. Sample mode reads it too",
	}}
}

// loadPolicy reads the policy file at path. A missing, unreadable, over-bound or malformed file is a
// *ConfigError naming the setting and the file.
func loadPolicy(path string) (*policy, error) {
	allowed := []string{"a path to an existing file outside the repository under review"}
	st, err := os.Stat(path)
	if err != nil {
		reason := err.Error()
		if os.IsNotExist(err) {
			reason = "file does not exist"
		}
		return nil, &ConfigError{Field: EnvPolicy, Value: path, Reason: reason, Allowed: allowed}
	}
	if st.IsDir() {
		return nil, &ConfigError{Field: EnvPolicy, Value: path, Reason: "path is a directory", Allowed: allowed}
	}
	// ADR-0013: the policy is read no further than the configuration bound, as proven.yaml is.
	b, err := change.ReadBounded(path, change.Bounds().Config, "the organisation policy "+filepath.Base(path), change.EnvMaxConfigBytes)
	if err != nil {
		return nil, &ConfigError{Field: EnvPolicy, Value: path, Reason: err.Error(), Allowed: allowed}
	}
	p := &policy{path: path}
	col := &exceptionReader{byID: map[string]map[string]string{}}
	cfg, given, err := parseConfigOver(b, defaultSettings(), col.take)
	if err != nil {
		return nil, p.named(err)
	}
	p.settings, p.given = cfg, given
	p.exceptions, err = col.finish()
	if err != nil {
		return nil, p.named(err)
	}
	return p, nil
}

// named says in a refusal's reason that it came from the policy file, so a reader does not look for
// the fault in proven.yaml.
func (p *policy) named(err error) error {
	if ce, ok := err.(*ConfigError); ok {
		out := *ce
		out.Reason = "in the organisation policy " + p.path + ": " + ce.Reason
		return &out
	}
	return err
}

// exceptionReader collects the exception.<id>.<field> keys of a policy file as the reader meets them.
type exceptionReader struct {
	order []string
	byID  map[string]map[string]string
}

// take is the key hook of the policy file. It takes every exception key, refuses a proven.yaml key
// that has no stricter-than order, and leaves the policy keys to the ordinary reader.
func (c *exceptionReader) take(key, val string) (bool, error) {
	if !strings.HasPrefix(key, exceptionPrefix) {
		if knownKey(key) && !policyKey(key) {
			return false, configRefused(key, val, "not a policy setting: no order says which of its values is stricter", strings.Join(policyKeys, ", "))
		}
		return false, nil
	}
	const shape = "exception.<id>.<field>, the id from a-z, 0-9 and -, the field one of " + "rule, location, reason, owner, expires"
	rest := strings.TrimPrefix(key, exceptionPrefix)
	i := strings.IndexByte(rest, '.')
	if i <= 0 || !validExceptionID(rest[:i]) {
		return false, configRefused(key, val, "bad exception key", shape)
	}
	id, field := rest[:i], rest[i+1:]
	if !contains(exceptionFields, field) {
		return false, configRefused(key, val, "unknown exception field", shape)
	}
	if c.byID[id] == nil {
		c.byID[id] = map[string]string{}
		c.order = append(c.order, id)
	}
	if _, dup := c.byID[id][field]; dup {
		return false, configRefused(key, val, "duplicate key", shape)
	}
	c.byID[id][field] = val
	return true, nil
}

// finish checks that every exception names all five fields and reads them. An exception without an
// owner or an expiry is refused here, on load (ADR-0036 point 3), so it is never half applied.
func (c *exceptionReader) finish() ([]exception, error) {
	ids := append([]string(nil), c.order...)
	sort.Strings(ids)
	out := make([]exception, 0, len(ids))
	for _, id := range ids {
		f := c.byID[id]
		for _, field := range exceptionFields {
			if strings.TrimSpace(f[field]) == "" {
				key := exceptionPrefix + id + "." + field
				return nil, configRefused(key, f[field], "missing: an exception names its rule, location, reason, owner and expiry", "a non-empty value")
			}
		}
		e := exception{ID: id, Rule: f["rule"], Where: f["location"], Reason: f["reason"], Owner: f["owner"], Expires: f["expires"]}
		if _, err := time.Parse(dateLayout, e.Expires); err != nil {
			return nil, configRefused(exceptionPrefix+id+".expires", e.Expires, "not a date", "a date written YYYY-MM-DD, for example 2027-03-31")
		}
		e.Path, e.Line = splitLocation(e.Where)
		if e.Path == "" {
			return nil, configRefused(exceptionPrefix+id+".location", e.Where, "no path", "a path, or a path and a line: src/db.go or src/db.go:42")
		}
		out = append(out, e)
	}
	return out, nil
}

// splitLocation reads "path" or "path:line". A suffix after the last colon that is not a positive whole
// number is part of the path.
func splitLocation(where string) (string, int) {
	where = filepath.ToSlash(strings.TrimSpace(where))
	if i := strings.LastIndexByte(where, ':'); i >= 0 {
		if n, err := strconv.Atoi(where[i+1:]); err == nil && n > 0 {
			return strings.TrimPrefix(where[:i], "./"), n
		}
	}
	return strings.TrimPrefix(where, "./"), 0
}

// validExceptionID accepts a-z, 0-9 and -, not starting with a dash, so the id reads the same in every file.
func validExceptionID(id string) bool {
	if id == "" || id[0] == '-' {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' {
			return false
		}
	}
	return true
}

// policyKey reports whether a policy may set key.
func policyKey(key string) bool { return contains(policyKeys, key) }

// contains reports whether list holds s.
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// base is what a repository starts from: the policy's settings, or the defaults with no policy.
func (p *policy) base() fileSettings {
	if p == nil {
		return defaultSettings()
	}
	return p.settings
}

// parseLocal reads a repository's proven.yaml over the policy and refuses a value weaker than the
// policy's. A key the repository does not set keeps the policy's value.
func (p *policy) parseLocal(b []byte) (fileSettings, error) {
	if p == nil {
		return parseConfig(b)
	}
	cfg, given, err := parseConfigOver(b, p.settings, nil)
	if err != nil {
		return fileSettings{}, err
	}
	if err := p.refuseWeaker(cfg, given); err != nil {
		return fileSettings{}, err
	}
	return cfg, nil
}

// refuseWeaker compares each key the repository set with the same key of the policy, when the policy
// set it, and refuses the first value that is weaker (ADR-0036 point 1). cfg holds the repository's
// values; given holds the raw text of the keys it set.
func (p *policy) refuseWeaker(cfg fileSettings, given map[string]string) error {
	for _, key := range policyKeys {
		local, set := given[key]
		if !set {
			continue
		}
		if why := p.weaker(key, cfg); why != "" {
			return p.refuse(key, local, why)
		}
	}
	return nil
}

// weaker returns why the repository's value of key is weaker than the policy's, or "" when it is as
// strict or stricter, or when the policy did not set the key.
func (p *policy) weaker(key string, cfg fileSettings) string {
	pol := p.settings
	order := pol.ClassOrder
	_, polSet := p.given[key]
	switch key {
	case "class_order":
		// The order ranks every class a path rule or the fallback names, so a policy that sets either also
		// fixes the order; a repository that reorders the classes would move the policy's own paths.
		if !p.fixesClasses() {
			return ""
		}
		if strings.Join(cfg.ClassOrder, ", ") != strings.Join(order, ", ") {
			return "the policy fixes the order of the classes as " + strings.Join(order, ", ")
		}
	case "path_rules":
		if !p.fixesClasses() {
			return ""
		}
		return p.weakerRules(cfg)
	case "fallback_class":
		if !p.fixesClasses() {
			return ""
		}
		if classRank(order, cfg.Fallback) > classRank(order, pol.Fallback) {
			return "it falls back to " + cfg.Fallback + ", below " + pol.Fallback
		}
		if cfg.Fallback != pol.Fallback && p.leavesDoubled(cfg, cfg.Fallback) {
			return "it falls back to " + cfg.Fallback + ", above " + pol.Fallback + ", and does not double " + cfg.Fallback + " where the policy doubles " + p.doubledList()
		}
	case "doubled_classes":
		if polSet {
			for _, c := range classNames() {
				if pol.Doubled[c] && !cfg.Doubled[c] {
					return "it stops doubling " + c
				}
			}
			return ""
		}
		if !p.fixesClasses() {
			return ""
		}
		// The policy did not name the list, but it fixes the classes, and with a second reviewer
		// connected an unset list doubles every class: compare what each side then doubles.
		polDoubled := withSecond(pol, "second").Doubled
		cfgDoubled := withSecond(cfg, "second").Doubled
		for _, c := range classNames() {
			if polDoubled[c] && !cfgDoubled[c] {
				return "it stops doubling " + c + " when a second reviewer is connected"
			}
		}
	case "author_commit_trailers":
		if polSet && pol.ReadCommits && !cfg.ReadCommits {
			return "it turns the reading of commit trailers off"
		}
	case "author_description":
		if polSet && pol.ReadDescription && !cfg.ReadDescription {
			return "it turns the reading of the description off"
		}
	case "raise_class":
		if polSet && pol.RaiseClass && !cfg.RaiseClass {
			return "it turns raise_class off"
		}
	case "author_trailer_keys":
		if !polSet {
			return ""
		}
		have := map[string]bool{}
		for _, k := range cfg.TrailerKeys {
			have[strings.ToLower(k)] = true
		}
		for _, k := range pol.TrailerKeys {
			if !have[strings.ToLower(k)] {
				return "it drops the key " + k
			}
		}
	}
	return ""
}

// fixesClasses reports whether the policy sets any key that decides the class of a path: path_rules,
// class_order or fallback_class, or doubled_classes, which names the classes whose paths are reviewed
// twice and so says which paths those are (with a second reviewer connected, security is doubled even
// under "doubled_classes: off"). Then the policy's effective classification (its rules, or the defaults,
// and its fallback) is a floor for every path, and a repository's own is compared with it.
func (p *policy) fixesClasses() bool {
	for _, key := range []string{"path_rules", "class_order", "fallback_class", "doubled_classes"} {
		if _, ok := p.given[key]; ok {
			return true
		}
	}
	return false
}

// doubledRank is the place, in the policy's order, of the least strict class the policy doubles. It is
// false when the policy does not name doubled_classes or names none: then no path is doubled by the
// policy's word and a class a repository moves a path to cannot undouble it.
func (p *policy) doubledRank() (int, bool) {
	if _, set := p.given["doubled_classes"]; !set {
		return 0, false
	}
	least, any := 0, false
	for _, c := range classNames() {
		if p.settings.Doubled[c] {
			if r := classRank(p.settings.ClassOrder, c); !any || r > least {
				least, any = r, true
			}
		}
	}
	return least, any
}

// leavesDoubled reports whether a path the repository puts in class could have been a path the policy
// doubles, and the repository does not double class. The repository's classification is never lower than
// the policy's (weakerRules), so a path it moves to class comes from a class below it; when the policy
// doubles a class below class, the move may take the path out of a doubled class into one that is not.
// The check does not list paths: a pattern cannot be shown disjoint from every other.
func (p *policy) leavesDoubled(cfg fileSettings, class string) bool {
	least, ok := p.doubledRank()
	return ok && classRank(p.settings.ClassOrder, class) < least && !cfg.Doubled[class]
}

// doubledList names the classes the policy doubles, in the order of the classes.
func (p *policy) doubledList() string {
	var out []string
	for _, c := range p.settings.ClassOrder {
		if p.settings.Doubled[c] {
			out = append(out, c)
		}
	}
	return strings.Join(out, ", ")
}

// weakerRules compares the class every path gets under the repository's rules and fallback with the
// class it gets under the policy's, and returns why some path would get a lower one, or "". A path takes
// the highest class of the rules that match it, else the fallback (ADR-0014). Two facts make the
// comparison sound without listing paths:
//
//  1. A path a policy rule covers keeps its class when the repository holds that rule at the same class
//     or a higher one, because extra rules only add candidates and the highest class wins.
//  2. A path no policy rule covers gets the policy's fallback, so a repository rule that puts any path
//     below that fallback may lower it.
//
// A rule of the repository that is not one of the policy's and sits below the policy's fallback is refused
// even if it could only match paths a policy rule covers: a pattern cannot be shown disjoint from every
// other, so the refusal takes the stricter reading.
func (p *policy) weakerRules(cfg fileSettings) string {
	pol, order := p.settings, p.settings.ClassOrder
	// The rule that lowers a path is named first: it is the one a reader must change.
	floor := classRank(order, pol.Fallback)
	for _, r := range cfg.PathRules {
		if classRank(order, r.Class) <= floor {
			continue
		}
		if _, ours := ruleFor(pol.PathRules, r.Pattern); ours {
			continue
		}
		return "its rule " + r.Pattern + "=" + r.Class + " gives paths " + r.Class + " where the policy's fallback gives " + pol.Fallback + ", below it"
	}
	// A rule the policy does not hold raises the paths it matches to its class. Raised past a class the
	// policy doubles into one the repository does not double, a doubled path would be reviewed once.
	for _, r := range cfg.PathRules {
		if !hasRule(pol.PathRules, r) && p.leavesDoubled(cfg, r.Class) {
			return "its rule " + r.Pattern + "=" + r.Class + " may move a path out of a class the policy doubles (" + p.doubledList() + ") into " + r.Class + ", which it does not double"
		}
	}
	for _, want := range pol.PathRules {
		got, ok := ruleFor(cfg.PathRules, want.Pattern)
		if !ok {
			return "it leaves out the rule " + want.Pattern + "=" + want.Class
		}
		if classRank(order, got.Class) > classRank(order, want.Class) {
			return "it puts " + want.Pattern + " in " + got.Class + ", below " + want.Class
		}
	}
	return ""
}

// hasRule reports whether rules holds r: the same pattern at the same class.
func hasRule(rules []pathRule, r pathRule) bool {
	for _, have := range rules {
		if have == r {
			return true
		}
	}
	return false
}

// ruleFor finds the rule for one pattern.
func ruleFor(rules []pathRule, pattern string) (pathRule, bool) {
	for _, r := range rules {
		if r.Pattern == pattern {
			return r, true
		}
	}
	return pathRule{}, false
}

// refuse is the *ConfigError for a weaker local value. It names the key, the local value, the policy's
// value and the policy file, so the reader fixes the right file.
func (p *policy) refuse(key, local, why string) *ConfigError {
	polVal, ok := p.given[key]
	if !ok {
		polVal = "(not set; the policy's rules still apply)"
	}
	return &ConfigError{
		Field: key, Value: local,
		Reason:  fmt.Sprintf("weaker than the organisation policy %s, which has %s=%q: %s", p.path, key, polVal, why),
		Allowed: []string{"a value as strict as the policy's, or stricter"},
	}
}

// refuseInside refuses a policy file that lies inside the repository under review: the change under
// review could edit it, so a change would carry its own policy (ADR-0036 point 1). The repository is the
// target and, when the target lies in a git working tree, that whole tree from its top level: a target may
// be one directory of a larger repository. Links are resolved on both sides before they are compared.
func (p *policy) refuseInside(target string) error {
	if p == nil {
		return nil
	}
	roots := make([]string, 0, 2)
	root, err := resolvedAbs(target)
	if err != nil {
		return err
	}
	roots = append(roots, root)
	top, in, err := change.TopLevel(root)
	if err != nil {
		return err
	}
	if in {
		if top, err = resolvedAbs(top); err != nil {
			return err
		}
		roots = append(roots, top)
	}
	file, err := resolvedAbs(p.path)
	if err != nil {
		return err
	}
	if repositoryHolds(roots, file) {
		return &ConfigError{
			Field: EnvPolicy, Value: p.path,
			Reason:  "the policy file lies inside the repository under review, so a change could carry its own policy",
			Allowed: []string{"a path to a file outside the repository under review"},
		}
	}
	return nil
}

// repositoryHolds reports whether file lies inside any of roots.
func repositoryHolds(roots []string, file string) bool {
	for _, root := range roots {
		if pathInside(root, file) {
			return true
		}
	}
	return false
}

// resolvedAbs is the absolute path with every link resolved.
func resolvedAbs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// matches reports whether the exception covers the finding: the same rule, the same path, and the
// line inside the finding's lines when the exception names one.
func (e exception) matches(f Finding) bool {
	if e.Rule != f.RuleID || e.Path != strings.TrimPrefix(filepath.ToSlash(f.Location.Path), "./") {
		return false
	}
	if e.Line == 0 {
		return true
	}
	end := f.Location.EndLine
	if end < f.Location.StartLine {
		end = f.Location.StartLine
	}
	return e.Line >= f.Location.StartLine && e.Line <= end
}

// applyExceptions moves each finding an unexpired exception covers out of Findings into Accepted, with
// the exception's reason, owner and expiry, so it is shown and is not counted (ADR-0036 point 2). A
// finding covered only by expired exceptions stays in Findings and each such exception is named in
// Expired. The exception holds through the whole of its expiry date, UTC; now is the review's clock.
func (p *policy) applyExceptions(out *Result, now time.Time) {
	if p == nil || len(p.exceptions) == 0 {
		return
	}
	today := now.UTC().Format(dateLayout)
	kept := make([]Finding, 0, len(out.Findings))
	for _, f := range out.Findings {
		var live *exception
		var expired []exception
		for i := range p.exceptions {
			e := p.exceptions[i]
			if !e.matches(f) {
				continue
			}
			// Dates in this layout order the same as text and as time.
			if e.Expires < today {
				expired = append(expired, e)
			} else if live == nil {
				live = &p.exceptions[i]
			}
		}
		if live != nil {
			out.Accepted = append(out.Accepted, AcceptedFinding{
				Finding: f, ExceptionID: live.ID, Reason: live.Reason, Owner: live.Owner, Expires: live.Expires,
			})
			continue
		}
		for _, e := range expired {
			out.Expired = append(out.Expired, ExpiredException{
				ExceptionID: e.ID, Finding: findingName(f), Rule: e.Rule, Location: e.Where,
				Reason: e.Reason, Owner: e.Owner, Expires: e.Expires,
			})
		}
		kept = append(kept, f)
	}
	out.Findings = kept
	// An accepted finding is not active: the proof totals count active findings only.
	for level := range out.ProofCounts {
		out.ProofCounts[level] = 0
	}
	for _, f := range kept {
		if f.ProofLevel != "" && out.ProofCounts != nil {
			out.ProofCounts[f.ProofLevel]++
		}
	}
}

// findingName is how a review names a finding: its id, else its rule.
func findingName(f Finding) string {
	if f.ID != "" {
		return f.ID
	}
	return f.RuleID
}
