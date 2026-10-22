package review

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
func applyDecision(out *Result, cfg fileSettings, req Request) error {
	files, err := changedPaths(req)
	if err != nil {
		return err
	}
	class, rule := chooseClass(files, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
	authors, err := readAuthors(cfg, req)
	if err != nil {
		return err
	}
	for i := range authors {
		authors[i].Family = familyOf(authors[i].Model, cfg.Families)
	}
	doubled := cfg.Doubled[class]
	models := []string{cfg.DefaultReviewer}
	lines := []string{reviewerLine("default", cfg.DefaultReviewer, familyOf(cfg.DefaultReviewer, cfg.Families), authors)}
	if doubled {
		if cfg.SpecialReviewer != "" {
			models = append(models, cfg.SpecialReviewer)
			lines = append(lines, reviewerLine("special", cfg.SpecialReviewer, familyOf(cfg.SpecialReviewer, cfg.Families), authors))
		} else {
			lines = append(lines, "special-review model: not configured")
		}
	}
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
	out.AuthorModelStatement = authorStatement(authors, cfg.DefaultReviewer)
	return nil
}

func changedPaths(req Request) ([]string, error) {
	if req.BaselineCommit != "" {
		return change.Between(req.Target, req.BaselineCommit, req.HeadCommit)
	}
	files, err := change.Read(req.Target)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out, nil
}

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
		if r < bestRank {
			best = m
			bestRank = r
		}
	}
	return best, true
}

func classRank(order []string, class string) int {
	for i, c := range order {
		if c == class {
			return i
		}
	}
	return len(order)
}

func familyOf(model string, table []familyRule) string {
	for _, r := range table {
		if matchModel(r.Pattern, model) {
			return r.Family
		}
	}
	return "family unknown"
}

func matchModel(pattern, model string) bool {
	if !strings.ContainsAny(pattern, "*?[") {
		return pattern == model
	}
	ok, err := path.Match(pattern, model)
	return err == nil && ok
}

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

func matchSegs(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
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

func readAuthors(cfg fileSettings, req Request) ([]AuthorModel, error) {
	var found []AuthorModel
	if cfg.ReadCommits && req.BaselineCommit != "" {
		msgs, err := change.Messages(req.Target, req.BaselineCommit, req.HeadCommit)
		if err != nil {
			return nil, err
		}
		for _, m := range msgs {
			found = append(found, scanTrailers(m, cfg.TrailerKeys, "commit")...)
		}
	}
	if cfg.ReadDescription && req.Description != "" {
		found = append(found, scanTrailers(req.Description, cfg.TrailerKeys, "description")...)
	}
	if cfg.Record != "" {
		text, err := readRecord(req.Target, cfg.Record)
		if err != nil {
			return nil, err
		}
		found = append(found, scanTrailers(text, cfg.TrailerKeys, "record")...)
	}
	return dedupeAuthors(found), nil
}

func scanTrailers(text string, keys []string, source string) []AuthorModel {
	want := map[string]struct{}{}
	for _, k := range keys {
		k = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(k), ":"))
		if k != "" {
			want[k] = struct{}{}
		}
	}
	var out []AuthorModel
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
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
			continue
		}
		out = append(out, AuthorModel{Agent: agent, Model: model, Source: source})
	}
	return out
}

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

func readRecord(target, pattern string) (string, error) {
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
	for _, p := range matches {
		body, err := os.ReadFile(p)
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
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	if !pathInside(root, resolved) {
		return "", recordRefused(pattern, "path resolves outside the review target")
	}
	return resolved, nil
}

func recordRefused(value, reason string) *ConfigError {
	return &ConfigError{
		Field: "author_record", Value: value, Reason: reason,
		Allowed: []string{"off, or a relative path or pattern inside the review target"},
	}
}

func pathInside(root, resolved string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(resolved))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func authorStatement(models []AuthorModel, defaultReviewer string) string {
	if len(models) == 0 {
		return "author model not stated; default reviewer " + defaultReviewer
	}
	parts := make([]string, len(models))
	for i, m := range models {
		parts[i] = fmt.Sprintf("author %s:%s family %s from %s", m.Agent, m.Model, m.Family, m.Source)
	}
	return strings.Join(parts, "; ")
}

func reviewerLine(role, model, family string, authors []AuthorModel) string {
	head := fmt.Sprintf("reviewer %s; role %s; family %s", model, role, family)
	if len(authors) == 0 {
		return head + "; author model not stated"
	}
	parts := []string{head}
	for _, a := range authors {
		parts = append(parts, fmt.Sprintf("%s (%s): %s", a.Model, a.Family, familyRelation(family, a.Family)))
	}
	return strings.Join(parts, "; ")
}

func familyRelation(reviewerFamily, authorFamily string) string {
	if reviewerFamily == "family unknown" || authorFamily == "family unknown" {
		return "family unknown"
	}
	if reviewerFamily == authorFamily {
		return "yes"
	}
	return "no"
}
