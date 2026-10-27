package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFamilyTableDefaults proves the shipped family table of ADR-0011 point 4, and that a model it
// does not know, even a near miss such as gpt4, is "family unknown", never guessed.
func TestFamilyTableDefaults(t *testing.T) {
	cfg := defaultSettings()
	cases := []struct {
		model  string
		family string
	}{
		{"claude-3-opus", "Anthropic"},
		{"claude-opus", "Anthropic"},
		{"gpt-4o", "OpenAI"},
		{"codex-mini", "OpenAI"},
		{"nemotron-super", "NVIDIA"},
		{"qwen2.5", "Alibaba"},
		{"qwen", "Alibaba"},
		{"NVIDIA", "NVIDIA"},
		{"gemini-2", "family unknown"},
		{"gpt4", "family unknown"},
	}
	for _, c := range cases {
		if got := familyOf(c.model, cfg.Families); got != c.family {
			t.Errorf("%s: got %s want %s", c.model, got, c.family)
		}
	}
}

// TestOpenFamilyNameIsAccepted proves that a family name outside the shipped table is accepted: the
// family names are open, only the classes are a closed list.
func TestOpenFamilyNameIsAccepted(t *testing.T) {
	cfg, err := parseConfig([]byte("families: gemini-*=Google\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := familyOf("gemini-2", cfg.Families); got != "Google" {
		t.Fatalf("gemini-2: %s", got)
	}
	if got := familyOf("claude-3-opus", cfg.Families); got != "family unknown" {
		t.Fatalf("replaced table still maps claude-3-opus to %s", got)
	}
}

// TestConfiguredFamilyTableReplacesTheDefault proves that a configured families table replaces the
// shipped one instead of merging with it.
func TestConfiguredFamilyTableReplacesTheDefault(t *testing.T) {
	cfg, err := parseConfig([]byte("families: gemini-*=OpenAI\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := familyOf("gemini-2", cfg.Families); got != "OpenAI" {
		t.Fatalf("gemini-2: %s", got)
	}
	if got := familyOf("claude-3-opus", cfg.Families); got != "family unknown" {
		t.Fatalf("replaced table still maps claude-3-opus to %s", got)
	}
}

// TestClassRules proves the shipped path rules of ADR-0014 point 2: each sample path gets its class and
// the rule that chose it, a near miss (author/) falls to the fallback, and a base-name rule matches
// in a subdirectory.
func TestClassRules(t *testing.T) {
	cfg := defaultSettings()
	cases := []struct {
		file  string
		class string
		rule  string
	}{
		{"pkg/auth/login.go", "security", "**/auth/**"},
		{"auth/login.go", "security", "**/auth/**"},
		{"pkg/author/login.go", "backend", "fallback"},
		{"billing/payments/charge.go", "security", "**/payment*/**"},
		{"config/api_secret.go", "security", "**/*secret*"},
		{"db/migrations/001.sql", "security", "**/migrations/**"},
		{"go.mod", "security", "go.mod"},
		{"service/go.mod", "security", "go.mod"},
		{"composer.json", "security", "composer.json"},
		{"app/composer.lock", "security", "composer.lock"},
		{"build.gradle.kts", "security", "build.gradle.kts"},
		{"gradle/libs.versions.toml", "security", "libs.versions.toml"},
		{"gradle.lockfile", "security", "gradle.lockfile"},
		{"README.md", "docs", "**/*.md"},
		{"docs/guide.md", "docs", "**/*.md"},
		{"billing/discount.go", "backend", "fallback"},
	}
	for _, c := range cases {
		class, rule := chooseClass([]string{c.file}, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
		if class != c.class || rule != c.rule {
			t.Errorf("%s: got %s by %s, want %s by %s", c.file, class, rule, c.class, c.rule)
		}
	}
}

// TestHighestRiskWinsOverEarlierLowerRule proves ADR-0014 point 3: the highest-risk class wins, not
// the rule listed first.
func TestHighestRiskWinsOverEarlierLowerRule(t *testing.T) {
	rules := []pathRule{
		{Pattern: "**/*.md", Class: "docs"},
		{Pattern: "**/*secret*", Class: "security"},
	}
	class, rule := chooseClass([]string{"docs/api_secret.md"}, rules, classNames(), "backend")
	if class != "security" || rule != "**/*secret*" {
		t.Fatalf("got %s by %s, want security by **/*secret*", class, rule)
	}
}

// TestConfiguredOrderCanPutDocsAboveBackend proves classes besides security may be reordered.
func TestConfiguredOrderCanPutDocsAboveBackend(t *testing.T) {
	cfg, err := parseConfig([]byte("class_order: security, docs, user interface, mobile, backend\n"))
	if err != nil {
		t.Fatal(err)
	}
	rules := []pathRule{
		{Pattern: "**/*.md", Class: "docs"},
		{Pattern: "**/*secret*", Class: "backend"},
	}
	class, rule := chooseClass([]string{"docs/api_secret.md"}, rules, cfg.ClassOrder, cfg.Fallback)
	if class != "docs" || rule != "**/*.md" {
		t.Fatalf("got %s by %s, want docs by **/*.md", class, rule)
	}
}

// TestMobileOutranksUserInterface proves that with several changed files, the class is decided across
// all of them by the default order.
func TestMobileOutranksUserInterface(t *testing.T) {
	rules := []pathRule{
		{Pattern: "**/*.tsx", Class: "user interface"},
		{Pattern: "**/*.swift", Class: "mobile"},
	}
	class, rule := chooseClass([]string{"ios/app.swift", "web/app.tsx"}, rules, classNames(), "backend")
	if class != "mobile" || rule != "**/*.swift" {
		t.Fatalf("got %s by %s", class, rule)
	}
}

// TestSecurityIsDoubledAndBackendIsNot proves the default doubling (ADR-0009): a security change is
// doubled and says when no special model is configured; a backend change runs one reviewer.
func TestSecurityIsDoubledAndBackendIsNot(t *testing.T) {
	cfg := defaultSettings()
	secDir := treeWith(t, "go.mod", "module example.com/sec\n")
	sec := decideOn(t, cfg, Request{Target: secDir})
	if sec.Class != "security" || !sec.Doubled {
		t.Fatalf("class %s doubled %t", sec.Class, sec.Doubled)
	}
	if len(sec.ReviewerStatements) != 2 || sec.ReviewerStatements[1] != "special-review model: not configured" {
		t.Fatalf("%#v", sec.ReviewerStatements)
	}

	backDir := treeWith(t, "billing/discount.go", "package billing\n")
	back := decideOn(t, cfg, Request{Target: backDir})
	if back.Class != "backend" || back.ClassRule != "fallback" || back.Doubled {
		t.Fatalf("class %s rule %s doubled %t", back.Class, back.ClassRule, back.Doubled)
	}
	if len(back.ReviewerStatements) != 1 {
		t.Fatalf("%#v", back.ReviewerStatements)
	}
}

// treeWith builds a base/change pair of directories whose only difference is one new file at rel.
func treeWith(t *testing.T, rel, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "change", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// decideOn runs applyDecision alone, without the agents, and returns what it filled in.
func decideOn(t *testing.T, cfg fileSettings, req Request) Result {
	t.Helper()
	var out Result
	if err := applyDecision(&out, cfg, req, nil); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestDoubledClassAddsTheSpecialReviewer proves ADR-0009 point 5: a reviewer line names its model and
// role and says, per author model, whether it shares the family: no, yes, or family unknown.
func TestDoubledClassAddsTheSpecialReviewer(t *testing.T) {
	cfg, err := parseConfig([]byte("special_reviewer: qwen-special\n"))
	if err != nil {
		t.Fatal(err)
	}
	authors := []AuthorModel{{Agent: "Claude", Model: "claude-3-opus", Family: "Anthropic", Source: "commit"}}
	line := reviewerLine("special", cfg.SpecialReviewer, familyOf(cfg.SpecialReviewer, cfg.Families), authors, nil)
	if !strings.Contains(line, "reviewer qwen-special; role special") {
		t.Fatal(line)
	}
	if !strings.Contains(line, "claude-3-opus (Anthropic): no") {
		t.Fatal(line)
	}
	same := reviewerLine("default", "NVIDIA", "NVIDIA", []AuthorModel{{Model: "nemotron-super", Family: "NVIDIA"}}, nil)
	if !strings.Contains(same, "nemotron-super (NVIDIA): yes") {
		t.Fatal(same)
	}
	unknown := reviewerLine("default", "NVIDIA", "NVIDIA", []AuthorModel{{Model: "gemini-2", Family: "family unknown"}}, nil)
	if !strings.Contains(unknown, "gemini-2 (family unknown): family unknown") {
		t.Fatal(unknown)
	}
}

// TestAuthorStatementWhenNothingIsNamed proves ADR-0011 point 6: with no author model named, the review
// says "author model not stated" and names the default reviewer.
func TestAuthorStatementWhenNothingIsNamed(t *testing.T) {
	got := authorStatement(nil, nil, "NVIDIA")
	if got != "author model not stated; default reviewer NVIDIA" {
		t.Fatal(got)
	}
	line := reviewerLine("default", "NVIDIA", "NVIDIA", nil, nil)
	if !strings.Contains(line, "author model not stated") {
		t.Fatal(line)
	}
}

// TestScanTrailersReadsTheAgentAndModel proves that a trailer is read as agent:model, that a bad value
// is kept as unreadable, that an indented line is not a trailer, and that a repeated model counts once.
func TestScanTrailersReadsTheAgentAndModel(t *testing.T) {
	text := "[fixture] change: second\n\nAssisted-by: Claude:claude-3-opus\nAssisted-by: not-a-pair\n  Assisted-by: Indented:nope\n"
	got, bad := scanTrailers(text, []string{"Assisted-by"}, "commit")
	if len(got) != 1 || got[0].Agent != "Claude" || got[0].Model != "claude-3-opus" || got[0].Source != "commit" {
		t.Fatalf("%#v", got)
	}
	if len(bad) != 1 || bad[0] != "Assisted-by: not-a-pair" {
		t.Fatalf("unreadable %#v", bad)
	}
	two := "Assisted-by: Claude:claude-3-opus\nAssisted-by: Nemotron:nemotron-super\n"
	parsed, _ := scanTrailers(two, []string{"Assisted-by"}, "description")
	got = dedupeAuthors(parsed)
	if len(got) != 2 || got[1].Model != "nemotron-super" {
		t.Fatalf("%#v", got)
	}
	again := two + "Assisted-by: Claude:claude-3-opus\n"
	parsed, _ = scanTrailers(again, []string{"Assisted-by"}, "description")
	got = dedupeAuthors(parsed)
	if len(got) != 2 {
		t.Fatalf("duplicate model kept: %#v", got)
	}
}

// TestUnreadableTrailerIsNamed proves that a trailer the reader cannot split is quoted in the review
// as unreadable, and is never reported as "author model not stated".
func TestUnreadableTrailerIsNamed(t *testing.T) {
	text := "Assisted-by: claude-opus-5-5\nCo-authored-by: Claude <noreply@example.com>\n"
	got, bad := scanTrailers(text, []string{"Assisted-by", "Co-authored-by"}, "description")
	if len(got) != 0 {
		t.Fatalf("parsed %#v", got)
	}
	if len(bad) != 2 || bad[0] != "Assisted-by: claude-opus-5-5" || bad[1] != "Co-authored-by: Claude <noreply@example.com>" {
		t.Fatalf("%#v", bad)
	}
	stmt := authorStatement(nil, bad, "NVIDIA")
	if strings.Contains(stmt, "not stated") {
		t.Fatal(stmt)
	}
	if !strings.Contains(stmt, "author statement unreadable: Assisted-by: claude-opus-5-5") {
		t.Fatal(stmt)
	}
	if !strings.Contains(stmt, "author statement unreadable: Co-authored-by: Claude <noreply@example.com>") {
		t.Fatal(stmt)
	}
	line := reviewerLine("default", "NVIDIA", "NVIDIA", nil, bad)
	if strings.Contains(line, "not stated") || !strings.Contains(line, "author statement unreadable: Assisted-by: claude-opus-5-5") {
		t.Fatal(line)
	}
}

// TestConfigRefusals proves that each bad value is refused with its field, its value, its reason and the
// allowed values, all in the printed message (ADR-0013, ADR-0014 point 7).
func TestConfigRefusals(t *testing.T) {
	cases := []struct {
		body   string
		field  string
		value  string
		reason string
	}{
		{"families: not-a-pair\n", "families", "not-a-pair", "want pattern=value"},
		{"families: gemini-*=\n", "families", "gemini-*=", "want pattern=value"},
		{"path_rules: **/*.go=frontend\n", "path_rules", "frontend", "unknown class"},
		{"class_order: security, backend\n", "class_order", "security, backend", "leaves out"},
		{"colour: red\n", "colour", "red", "unknown key"},
		{"fallback_class: frontend\n", "fallback_class", "frontend", "unknown class"},
		{"doubled_classes: frontend\n", "doubled_classes", "frontend", "unknown class"},
		{"raise_class: maybe\n", "raise_class", "maybe", "not on or off"},
		{"author_commit_trailers: yes\n", "author_commit_trailers", "yes", "not on or off"},
	}
	for _, c := range cases {
		_, err := parseConfig([]byte(c.body))
		cfg, ok := err.(*ConfigError)
		if !ok {
			t.Fatalf("%s: %v", c.body, err)
		}
		if cfg.Field != c.field || cfg.Value != c.value || !strings.Contains(cfg.Reason, c.reason) {
			t.Fatalf("%s: field %q value %q reason %q", strings.TrimSpace(c.body), cfg.Field, cfg.Value, cfg.Reason)
		}
		if len(cfg.Allowed) == 0 {
			t.Fatalf("%s: no allowed values", c.body)
		}
		if !strings.Contains(cfg.Error(), c.field) || !strings.Contains(cfg.Error(), c.value) {
			t.Fatalf("message %q", cfg.Error())
		}
	}
}

// TestClassOrderThatListsEveryClassIsAccepted proves that a reordered list naming every class is accepted.
func TestClassOrderThatListsEveryClassIsAccepted(t *testing.T) {
	_, err := parseConfig([]byte("class_order: security, docs, user interface, mobile, backend\n"))
	if err != nil {
		t.Fatal(err)
	}
}

// TestRaiseClassOnDoesNotChangeTheClass proves that raise_class is read and kept but changes no class:
// no model raises a class in this build (ADR-0014 point 9).
func TestRaiseClassOnDoesNotChangeTheClass(t *testing.T) {
	cfg, err := parseConfig([]byte("raise_class: on\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RaiseClass {
		t.Fatal("raise_class was not stored")
	}
	class, rule := chooseClass([]string{"billing/discount.go"}, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
	if class != "backend" || rule != "fallback" {
		t.Fatalf("raise_class changed the class to %s by %s", class, rule)
	}
}

// TestDoubledOffIsAllowed proves that doubled_classes: off doubles no class.
func TestDoubledOffIsAllowed(t *testing.T) {
	cfg, err := parseConfig([]byte("doubled_classes: off\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Doubled) != 0 {
		t.Fatalf("%v", cfg.Doubled)
	}
}
