package review

// This file decides what a review states about itself: its class and the rule that chose it
// (ADR-0012, ADR-0014), the author models the change names (ADR-0011), and each reviewer model with
// its family and whether it shares the author's (ADR-0008, ADR-0009). None of it is a model's guess.

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/change"
)

// applyDecision fills the class, the reviewer models and the author-model
// statement. raise_class is validated on load and is not applied: this wagon
// does not ask a model to raise a class (ADR-0014 point 9, off for the hackathon).
// planned, when set, is the plan Scout Intent chose (ADR-0049): its class and doubling are the review's,
// so the statements describe the review that ran. Nil states the class the path rules chose.
func applyDecision(out *Result, cfg fileSettings, req Request, planned *reviewPlan) error {
	files, err := changedPaths(req)
	if err != nil {
		return err
	}
	class, rule := chooseClass(files, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
	doubled := cfg.Doubled[class]
	if planned != nil {
		class, rule, doubled = planned.Class, planned.Rule, planned.Doubled
	}
	authors, unreadable, err := readAuthors(cfg, req)
	if err != nil {
		return err
	}
	for i := range authors {
		authors[i].Family = familyOf(authors[i].Model, cfg.Families)
	}
	// ADR-0009: every review uses the default reviewer; a doubled class adds the special model,
	// even when it is the author's own model (a fresh agent is still a second look).
	models := []string{cfg.DefaultReviewer}
	lines := []string{reviewerLine("default", cfg.DefaultReviewer, familyOf(cfg.DefaultReviewer, cfg.Families), authors, unreadable)}
	if doubled {
		if cfg.SpecialReviewer != "" {
			models = append(models, cfg.SpecialReviewer)
			lines = append(lines, reviewerLine("special", cfg.SpecialReviewer, familyOf(cfg.SpecialReviewer, cfg.Families), authors, unreadable))
		} else {
			// A doubled class with no special model is stated, not hidden: the reader sees one review ran.
			lines = append(lines, "special-review model: not configured")
		}
	}
	// ADR-0037 point 5: each reviewer is named, with whether it is the author's model.
	reviewers := []ReviewerInfo{reviewerOf(RoleDefault, cfg.DefaultReviewer, cfg.Families, authors, unreadable)}
	if doubled && cfg.SpecialReviewer != "" {
		reviewers = append(reviewers, reviewerOf(RoleSpecial, cfg.SpecialReviewer, cfg.Families, authors, unreadable))
	}
	out.Reviewers = reviewers
	// Never nil, so "no author models" prints as an empty list.
	if authors == nil {
		authors = []AuthorModel{}
	}
	out.Class = class
	out.ClassRule = rule
	out.Doubled = doubled
	out.DefaultReviewer = cfg.DefaultReviewer
	out.SpecialReviewer = cfg.SpecialReviewer
	out.ReviewerModels = models
	out.ReviewerStatements = lines
	out.AuthorModels = authors
	out.AuthorModelStatement = authorStatement(authors, unreadable, cfg.DefaultReviewer)
	// A change to proven.yaml is named, because the change's own copy of the settings does not
	// govern this review (see loadSettingsAtBaseline).
	for _, f := range files {
		if f == "proven.yaml" {
			out.SettingsStatement = "the change edits proven.yaml"
			break
		}
	}
	return nil
}

// changedPaths lists the paths the change touches: the diff from the baseline to the head when
// there is a baseline, else the working tree's changes.
func changedPaths(req Request) ([]string, error) {
	if req.BaselineCommit != "" {
		return req.read.Between(req.Target, req.BaselineCommit, req.HeadCommit)
	}
	files, err := req.read.Read(req.Target)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out, nil
}

// chooseClass returns the review class and the pattern that chose it (ADR-0014). Every rule that
// matches any changed path is a candidate; the highest-risk class wins (point 3). With no match the
// class is the configured fallback and the rule is "fallback" (point 4). No heuristic is added (point 8).
func chooseClass(files []string, rules []pathRule, order []string, fallback string) (string, string) {
	var matches []pathRule
	for _, r := range rules {
		for _, f := range files {
			if pathMatches(r.Pattern, f) {
				matches = append(matches, r)
				break
			}
		}
	}
	best, ok := highest(matches, order)
	if !ok {
		return fallback, "fallback"
	}
	return best.Class, best.Pattern
}

// highest returns the match whose class is earliest in order.
// order[0] is the highest risk. Among matches of that class, the earliest rule wins.
func highest(matches []pathRule, order []string) (pathRule, bool) {
	if len(matches) == 0 {
		return pathRule{}, false
	}
	best := matches[0]
	bestRank := classRank(order, best.Class)
	for _, m := range matches[1:] {
		r := classRank(order, m.Class)
		// Strictly less, so a later rule of the same class never replaces an earlier one.
		if r < bestRank {
			best = m
			bestRank = r
		}
	}
	return best, true
}

// classRank is the class's place in order. A class not in order ranks after every listed one;
// parseClassOrder makes a configured order list every class, so this is only a guard.
func classRank(order []string, class string) int {
	for i, c := range order {
		if c == class {
			return i
		}
	}
	return len(order)
}

// familyOf looks a model up in the configured family table; the first matching row wins.
// ADR-0011 point 4: a model the table does not know is "family unknown", never guessed.
func familyOf(model string, table []familyRule) string {
	for _, r := range table {
		if matchModel(r.Pattern, model) {
			return r.Family
		}
	}
	return "family unknown"
}

// matchModel matches one family pattern. A pattern with no wildcard is compared exactly, so
// a model name with a bracket in it is not read as a character class by mistake.
func matchModel(pattern, model string) bool {
	if !strings.ContainsAny(pattern, "*?[") {
		return pattern == model
	}
	ok, err := path.Match(pattern, model)
	return err == nil && ok
}

// pathMatches matches a path rule against one changed path, segment by segment, with ** for
// any number of directories. A pattern with no slash also matches the base name, so go.mod
// matches a go.mod in any directory.
func pathMatches(pattern, file string) bool {
	file = filepath.ToSlash(file)
	pattern = filepath.ToSlash(pattern)
	if matchSegs(strings.Split(pattern, "/"), strings.Split(file, "/")) {
		return true
	}
	if !strings.Contains(pattern, "/") {
		return matchSegs(strings.Split(pattern, "/"), []string{path.Base(file)})
	}
	return false
}

// matchSegs matches pattern segments against path segments. A ** segment matches zero or more
// path segments; any other segment matches exactly one, with path.Match. A bad segment fails closed.
func matchSegs(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
		// Collapse a run of ** into one, then try "matches nothing here" before "eats one segment".
		rest := pat
		for len(rest) > 0 && rest[0] == "**" {
			rest = rest[1:]
		}
		if matchSegs(rest, name) {
			return true
		}
		if len(name) > 0 && matchSegs(pat, name[1:]) {
			return true
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], name[0])
	if err != nil || !ok {
		return false
	}
	return matchSegs(pat[1:], name[1:])
}

// validPattern checks a pattern on load, so a bad pattern is refused with the setting's name
// (ADR-0013) rather than silently matching nothing during a review.
func validPattern(pattern string) error {
	if strings.TrimSpace(pattern) == "" {
		return fmt.Errorf("empty pattern")
	}
	for _, seg := range strings.Split(strings.ReplaceAll(pattern, "\\", "/"), "/") {
		if seg == "" || seg == "**" {
			continue
		}
		if _, err := path.Match(seg, "x"); err != nil {
			return err
		}
	}
	return nil
}

// readAuthors collects the author models: those declared in settings (author_models), then, from the
// sources the settings switch on, in ADR-0011's order: commit trailers, the pull-request description,
// the record file. It also
// returns the trailer lines it could not read, so the review can quote them instead of dropping them.
// ADR-0011 point 7: it only reads; it never writes a trailer.
func readAuthors(cfg fileSettings, req Request) ([]AuthorModel, []string, error) {
	found := append([]AuthorModel{}, cfg.Declared...)
	var unreadable []string
	take := func(text, source string) {
		authors, bad := scanTrailers(text, cfg.TrailerKeys, source)
		found = append(found, authors...)
		unreadable = append(unreadable, bad...)
	}
	// Commits are read only between a baseline and the head; without a baseline there are no
	// commits of the change to read.
	if cfg.ReadCommits && req.BaselineCommit != "" {
		msgs, err := req.read.Messages(req.Target, req.BaselineCommit, req.HeadCommit)
		if err != nil {
			return nil, nil, err
		}
		for _, m := range msgs {
			take(m, "commit")
		}
	}
	if cfg.ReadDescription && req.Description != "" {
		take(req.Description, "description")
	}
	if cfg.Record != "" {
		text, err := readRecord(req.read, req.Target, cfg.Record)
		if err != nil {
			return nil, nil, err
		}
		take(text, "record")
	}
	return dedupeAuthors(found), dedupeLines(unreadable), nil
}

// scanTrailers finds the lines "Key: agent:model" whose key is one of keys, matched without case.
// A known key with a value that is not agent:model is returned as unreadable.
func scanTrailers(text string, keys []string, source string) ([]AuthorModel, []string) {
	want := map[string]struct{}{}
	for _, k := range keys {
		k = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(k), ":"))
		if k != "" {
			want[k] = struct{}{}
		}
	}
	var out []AuthorModel
	var unreadable []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		// An indented line is a continuation or quoted text, not a trailer.
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}
		i := strings.IndexByte(line, ':')
		if i <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:i]))
		if _, ok := want[key]; !ok {
			continue
		}
		agent, model, ok := splitAgentModel(strings.TrimSpace(line[i+1:]))
		if !ok {
			unreadable = append(unreadable, line)
			continue
		}
		out = append(out, AuthorModel{Agent: agent, Model: model, Source: source})
	}
	return out, unreadable
}

// splitAgentModel splits "agent:model" at the first colon, so a model name may itself hold a colon.
// Both parts must be non-empty.
func splitAgentModel(val string) (string, string, bool) {
	i := strings.IndexByte(val, ':')
	if i <= 0 || i >= len(val)-1 {
		return "", "", false
	}
	agent := strings.TrimSpace(val[:i])
	model := strings.TrimSpace(val[i+1:])
	if agent == "" || model == "" {
		return "", "", false
	}
	return agent, model, true
}

// dedupeLines drops repeated lines and keeps the first order.
func dedupeLines(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// dedupeAuthors keeps the first entry for each model. The same model named in a commit and in
// the description is one author; the earlier source in ADR-0011's order is the one kept.
func dedupeAuthors(in []AuthorModel) []AuthorModel {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]AuthorModel, 0, len(in))
	for _, a := range in {
		if _, ok := seen[a.Model]; ok {
			continue
		}
		seen[a.Model] = struct{}{}
		out = append(out, a)
	}
	return out
}

// readRecord reads the author record file or files the setting names, in sorted order so the
// statement is the same on every run. A record that is configured but missing is refused: the team
// said where the record is, so its absence is a setting to fix, not "author model not stated".
func readRecord(rd change.Policy, target, pattern string) (string, error) {
	const allowed = "off, or a relative path or pattern inside the review target"
	matches, err := recordPaths(target, pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", &ConfigError{
			Field: "author_record", Value: pattern, Reason: "file does not exist",
			Allowed: []string{allowed},
		}
	}
	sort.Strings(matches)
	var b strings.Builder
	// ADR-0013: the record is bounded as a whole. Each file may use only what
	// the files before it left, so a pattern that matches many files cannot add up
	// to a read without bound.
	remaining := rd.Bounds().Record
	for _, p := range matches {
		body, err := change.ReadBounded(p, remaining, "the author record "+pattern, change.EnvMaxRecordBytes)
		if err == nil {
			remaining -= int64(len(body))
		}
		if err != nil {
			if os.IsNotExist(err) {
				return "", &ConfigError{
					Field: "author_record", Value: pattern, Reason: "file does not exist",
					Allowed: []string{allowed},
				}
			}
			return "", err
		}
		b.Write(body)
		// End each file with a newline, so the last line of one file never joins the first of the next.
		if !bytes.HasSuffix(body, []byte("\n")) {
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

// recordPaths returns regular files inside the review target. A symlink is
// refused by name before it is read, and so is any path that resolves outside
// the target. os.Stat and os.ReadFile follow links, so the check uses Lstat.
func recordPaths(target, pattern string) ([]string, error) {
	root, err := filepath.EvalSymlinks(target)
	if err != nil {
		return nil, err
	}
	if !strings.ContainsAny(pattern, "*?[") {
		p := filepath.Join(target, filepath.FromSlash(pattern))
		resolved, err := oneRecordPath(root, p, pattern)
		if err != nil || resolved == "" {
			return nil, err
		}
		return []string{resolved}, nil
	}
	var matches []string
	err = filepath.Walk(target, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Git's own files are never a record.
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(target, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !pathMatches(pattern, rel) {
			return nil
		}
		// filepath.Walk reports a link with Lstat, so a symlink is seen here before anything follows it.
		if info.Mode()&os.ModeSymlink != 0 {
			return recordRefused(rel, "symlink")
		}
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return err
		}
		if !pathInside(root, resolved) {
			return recordRefused(rel, "path resolves outside the review target")
		}
		matches = append(matches, resolved)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// oneRecordPath checks one exact record path. It returns "" with no error when the file does not
// exist, so readRecord can refuse it with the setting's name.
func oneRecordPath(root, p, pattern string) (string, error) {
	st, err := os.Lstat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		return "", recordRefused(pattern, "symlink")
	}
	if st.IsDir() {
		return "", recordRefused(pattern, "path is a directory")
	}
	// The file itself is not a link, but a directory on its path may be; resolve and check again.
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	if !pathInside(root, resolved) {
		return "", recordRefused(pattern, "path resolves outside the review target")
	}
	return resolved, nil
}

// recordRefused is the *ConfigError for a record path that is refused.
func recordRefused(value, reason string) *ConfigError {
	return &ConfigError{
		Field: "author_record", Value: value, Reason: reason,
		Allowed: []string{"off, or a relative path or pattern inside the review target"},
	}
}

// pathInside reports whether resolved is root or below it. Both are already resolved, so a
// ".." here can only mean the path left the root.
func pathInside(root, resolved string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(resolved))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// authorStatement is the review's one line about who wrote the change. ADR-0011 point 6: when no
// source names a model, it says so and names the default reviewer that was used.
func authorStatement(models []AuthorModel, unreadable []string, defaultReviewer string) string {
	parts := make([]string, 0, len(models)+len(unreadable))
	for _, m := range models {
		parts = append(parts, fmt.Sprintf("author %s:%s %s from %s", m.Agent, m.Model, familyPhrase(m.Family), m.Source))
	}
	for _, line := range unreadable {
		parts = append(parts, "author statement unreadable: "+line)
	}
	if len(parts) == 0 {
		return "author model not stated; default reviewer " + defaultReviewer
	}
	return strings.Join(parts, "; ")
}

// familyPhrase writes "family <name>", except for "family unknown", which
// already carries the word and would otherwise print it twice.
func familyPhrase(family string) string {
	if strings.HasPrefix(family, "family ") {
		return family
	}
	return "family " + family
}

// reviewerLine is one reviewer's statement: its model, role and family, and for each author model
// whether it shares the author's family. ADR-0009 point 5: a reader sees which independence the review had.
func reviewerLine(role, model, family string, authors []AuthorModel, unreadable []string) string {
	head := fmt.Sprintf("reviewer %s; role %s; %s", model, role, familyPhrase(family))
	if len(authors) == 0 && len(unreadable) == 0 {
		return head + "; author model not stated"
	}
	parts := []string{head}
	for _, a := range authors {
		parts = append(parts, fmt.Sprintf("%s (%s): %s", a.Model, a.Family, familyRelation(family, a.Family)))
	}
	for _, line := range unreadable {
		parts = append(parts, "author statement unreadable: "+line)
	}
	return strings.Join(parts, "; ")
}

// reviewerOf is the structured form of reviewerLine: the reviewer, its family, and whether it is the
// author's model. A reviewer is the author's model when the change names that exact model, or a
// model of the same family (ADR-0008); with no author named, or an unknown family, the answer is
// "unknown", never a guess.
func reviewerOf(role, model string, families []familyRule, authors []AuthorModel, unreadable []string) ReviewerInfo {
	family := familyOf(model, families)
	r := ReviewerInfo{Role: role, Model: model, Family: family, IsAuthorModel: "unknown"}
	if len(authors) == 0 {
		return r
	}
	if len(unreadable) > 0 {
		// An author statement the review could not read may name this very model.
		return r
	}
	answer := "no"
	for _, a := range authors {
		if a.Model == model {
			answer = "yes"
			break
		}
		switch familyRelation(family, a.Family) {
		case "yes":
			answer = "yes"
		case "family unknown":
			if answer == "no" {
				answer = "unknown"
			}
		}
		if answer == "yes" {
			break
		}
	}
	r.IsAuthorModel = answer
	return r
}

// familyRelation answers "is the reviewer in the author's family?". When either family is unknown
// the answer is "family unknown", never a yes or no.
func familyRelation(reviewerFamily, authorFamily string) string {
	if reviewerFamily == "family unknown" || authorFamily == "family unknown" {
		return "family unknown"
	}
	if reviewerFamily == authorFamily {
		return "yes"
	}
	return "no"
}
