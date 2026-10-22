package review

import (
	"strings"
	"testing"
)

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

func TestConfiguredOrderCanPutDocsAboveSecurity(t *testing.T) {
	cfg, err := parseConfig([]byte("class_order: docs, user interface, mobile, backend, security\n"))
	if err != nil {
		t.Fatal(err)
	}
	rules := []pathRule{
		{Pattern: "**/*.md", Class: "docs"},
		{Pattern: "**/*secret*", Class: "security"},
	}
	class, rule := chooseClass([]string{"docs/api_secret.md"}, rules, cfg.ClassOrder, cfg.Fallback)
	if class != "docs" || rule != "**/*.md" {
		t.Fatalf("got %s by %s, want docs by **/*.md", class, rule)
	}
}

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

func TestSecurityIsDoubledAndBackendIsNot(t *testing.T) {
	cfg := defaultSettings()
	sec := reviewerLine("default", cfg.DefaultReviewer, "NVIDIA", nil)
	lines := []string{sec, "special-review model: not configured"}
	if !cfg.Doubled["security"] {
		t.Fatal("security is not doubled")
	}
	if cfg.Doubled["backend"] || cfg.Doubled["docs"] {
		t.Fatal("a class other than security is doubled")
	}
	if lines[1] != "special-review model: not configured" {
		t.Fatal(lines[1])
	}
}

func TestDoubledClassAddsTheSpecialReviewer(t *testing.T) {
	cfg, err := parseConfig([]byte("special_reviewer: qwen-special\n"))
	if err != nil {
		t.Fatal(err)
	}
	authors := []AuthorModel{{Agent: "Claude", Model: "claude-3-opus", Family: "Anthropic", Source: "commit"}}
	line := reviewerLine("special", cfg.SpecialReviewer, familyOf(cfg.SpecialReviewer, cfg.Families), authors)
	if !strings.Contains(line, "reviewer qwen-special; role special") {
		t.Fatal(line)
	}
	if !strings.Contains(line, "claude-3-opus (Anthropic): no") {
		t.Fatal(line)
	}
	same := reviewerLine("default", "NVIDIA", "NVIDIA", []AuthorModel{{Model: "nemotron-super", Family: "NVIDIA"}})
	if !strings.Contains(same, "nemotron-super (NVIDIA): yes") {
		t.Fatal(same)
	}
	unknown := reviewerLine("default", "NVIDIA", "NVIDIA", []AuthorModel{{Model: "gemini-2", Family: "family unknown"}})
	if !strings.Contains(unknown, "gemini-2 (family unknown): family unknown") {
		t.Fatal(unknown)
	}
}

func TestAuthorStatementWhenNothingIsNamed(t *testing.T) {
	got := authorStatement(nil, "NVIDIA")
	if got != "author model not stated; default reviewer NVIDIA" {
		t.Fatal(got)
	}
	line := reviewerLine("default", "NVIDIA", "NVIDIA", nil)
	if !strings.Contains(line, "author model not stated") {
		t.Fatal(line)
	}
}

func TestScanTrailersReadsTheAgentAndModel(t *testing.T) {
	text := "[fixture] change: second\n\nAssisted-by: Claude:claude-3-opus\nAssisted-by: not-a-pair\n  Assisted-by: Indented:nope\n"
	got := scanTrailers(text, []string{"Assisted-by"}, "commit")
	if len(got) != 1 || got[0].Agent != "Claude" || got[0].Model != "claude-3-opus" || got[0].Source != "commit" {
		t.Fatalf("%#v", got)
	}
	two := "Assisted-by: Claude:claude-3-opus\nAssisted-by: Nemotron:nemotron-super\n"
	got = dedupeAuthors(scanTrailers(two, []string{"Assisted-by"}, "description"))
	if len(got) != 2 || got[1].Model != "nemotron-super" {
		t.Fatalf("%#v", got)
	}
	again := two + "Assisted-by: Claude:claude-3-opus\n"
	got = dedupeAuthors(scanTrailers(again, []string{"Assisted-by"}, "description"))
	if len(got) != 2 {
		t.Fatalf("duplicate model kept: %#v", got)
	}
}

func TestConfigRefusals(t *testing.T) {
	cases := []struct {
		body   string
		field  string
		value  string
		reason string
	}{
		{"families: gemini-*=Google\n", "families", "Google", "unknown family"},
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

func TestClassOrderThatListsEveryClassIsAccepted(t *testing.T) {
	_, err := parseConfig([]byte("class_order: docs, user interface, mobile, backend, security\n"))
	if err != nil {
		t.Fatal(err)
	}
}

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

func TestDoubledOffIsAllowed(t *testing.T) {
	cfg, err := parseConfig([]byte("doubled_classes: off\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Doubled) != 0 {
		t.Fatalf("%v", cfg.Doubled)
	}
}
