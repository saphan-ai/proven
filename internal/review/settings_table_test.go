package review

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestEachSettingAtDefaultAndOneOtherValue is one subtest per proven.yaml key.
// Each subtest checks the default and one other value. The other value is one
// the review uses, or a refusal for a value the key does not allow.
// The property it proves: every key of ADR-0013's closed set has a working default and really
// changes the review when set.
func TestEachSettingAtDefaultAndOneOtherValue(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		if defaultSettings().Timeout != 0 {
			t.Fatalf("default %s", defaultSettings().Timeout)
		}
		if got := mustParse(t, "timeout: 30s\n").Timeout; got != 30*time.Second {
			t.Fatalf("other %s", got)
		}
		root := moduleRoot(t)
		open := reviewSample(t, "", root)
		if open.Partial {
			t.Fatal("default timeout bounded the review")
		}
		bounded := reviewSample(t, "timeout: 1ns\n", root)
		if !bounded.Partial {
			t.Fatal("1ns did not bound the review")
		}
	})

	t.Run("default_reviewer", func(t *testing.T) {
		dir := treeWith(t, "go.mod", "module example.com/x\n")
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if def.DefaultReviewer != "NVIDIA" || !strings.Contains(def.AuthorModelStatement, "default reviewer NVIDIA") {
			t.Fatalf("%s reviewer %s", def.AuthorModelStatement, def.DefaultReviewer)
		}
		if !strings.Contains(def.ReviewerStatements[0], "reviewer NVIDIA; role default") {
			t.Fatal(def.ReviewerStatements[0])
		}
		other := decideOn(t, mustParse(t, "default_reviewer: claude-opus\n"), Request{Target: dir})
		if other.DefaultReviewer != "claude-opus" || !strings.Contains(other.AuthorModelStatement, "default reviewer claude-opus") {
			t.Fatalf("%s reviewer %s", other.AuthorModelStatement, other.DefaultReviewer)
		}
		if !strings.Contains(other.ReviewerStatements[0], "reviewer claude-opus; role default; family Anthropic") {
			t.Fatal(other.ReviewerStatements[0])
		}
	})

	t.Run("special_reviewer", func(t *testing.T) {
		dir := treeWith(t, "go.mod", "module example.com/x\n")
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if def.SpecialReviewer != "" || len(def.ReviewerStatements) != 2 || def.ReviewerStatements[1] != "special-review model: not configured" {
			t.Fatalf("%#v", def.ReviewerStatements)
		}
		other := decideOn(t, mustParse(t, "special_reviewer: qwen-special\n"), Request{Target: dir})
		if other.SpecialReviewer != "qwen-special" || len(other.ReviewerModels) != 2 || other.ReviewerModels[1] != "qwen-special" {
			t.Fatalf("%#v", other.ReviewerModels)
		}
		if !strings.Contains(other.ReviewerStatements[1], "reviewer qwen-special; role special") {
			t.Fatal(other.ReviewerStatements[1])
		}
	})

	t.Run("families", func(t *testing.T) {
		def := defaultSettings()
		if familyOf("claude-3-opus", def.Families) != "Anthropic" || familyOf("gemini-2", def.Families) != "family unknown" {
			t.Fatal("default family table")
		}
		other := mustParse(t, "families: gemini-*=Google\n")
		if familyOf("gemini-2", other.Families) != "Google" {
			t.Fatal(familyOf("gemini-2", other.Families))
		}
		if familyOf("claude-3-opus", other.Families) != "family unknown" {
			t.Fatal("other table kept the default claude row")
		}
	})

	t.Run("path_rules", func(t *testing.T) {
		dir := treeWith(t, "go.mod", "module example.com/x\n")
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if def.Class != "security" || def.ClassRule != "go.mod" {
			t.Fatalf("%s by %s", def.Class, def.ClassRule)
		}
		other := decideOn(t, mustParse(t, "path_rules: **=docs\n"), Request{Target: dir})
		if other.Class != "docs" || other.ClassRule != "**" {
			t.Fatalf("%s by %s", other.Class, other.ClassRule)
		}
	})

	t.Run("class_order", func(t *testing.T) {
		dir := treeWith(t, "docs/api_secret.md", "# secret\n")
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if def.Class != "security" || def.ClassRule != "**/*secret*" {
			t.Fatalf("default %s by %s", def.Class, def.ClassRule)
		}
		body := "class_order: docs, user interface, mobile, backend, security\n"
		other := decideOn(t, mustParse(t, body), Request{Target: dir})
		if other.Class != "docs" || other.ClassRule != "**/*.md" {
			t.Fatalf("other %s by %s", other.Class, other.ClassRule)
		}
	})

	t.Run("fallback_class", func(t *testing.T) {
		dir := treeWith(t, "billing/discount.go", "package billing\n")
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if def.Class != "backend" || def.ClassRule != "fallback" {
			t.Fatalf("default %s by %s", def.Class, def.ClassRule)
		}
		other := decideOn(t, mustParse(t, "fallback_class: docs\n"), Request{Target: dir})
		if other.Class != "docs" || other.ClassRule != "fallback" {
			t.Fatalf("other %s by %s", other.Class, other.ClassRule)
		}
	})

	t.Run("doubled_classes", func(t *testing.T) {
		sec := treeWith(t, "go.mod", "module example.com/x\n")
		back := treeWith(t, "billing/discount.go", "package billing\n")
		if !decideOn(t, defaultSettings(), Request{Target: sec}).Doubled {
			t.Fatal("default security is not doubled")
		}
		if decideOn(t, defaultSettings(), Request{Target: back}).Doubled {
			t.Fatal("default backend is doubled")
		}
		cfg := mustParse(t, "doubled_classes: docs, backend\n")
		if decideOn(t, cfg, Request{Target: sec}).Doubled {
			t.Fatal("list form still doubles security")
		}
		got := decideOn(t, cfg, Request{Target: back})
		if !got.Doubled || len(got.ReviewerStatements) != 2 {
			t.Fatalf("backend doubled %t lines %#v", got.Doubled, got.ReviewerStatements)
		}
	})

	t.Run("author_commit_trailers", func(t *testing.T) {
		dir, base, head := commitWithTrailer(t)
		req := Request{Target: dir, BaselineCommit: base, HeadCommit: head}
		def := decideOn(t, defaultSettings(), req)
		if !strings.Contains(def.AuthorModelStatement, "author Ada:model-a family family unknown from commit") {
			t.Fatal(def.AuthorModelStatement)
		}
		off := decideOn(t, mustParse(t, "author_commit_trailers: off\n"), req)
		if !strings.Contains(off.AuthorModelStatement, "author model not stated") || strings.Contains(off.AuthorModelStatement, "model-a") {
			t.Fatal(off.AuthorModelStatement)
		}
	})

	t.Run("author_description", func(t *testing.T) {
		dir := treeWith(t, "go.mod", "module example.com/x\n")
		req := Request{Target: dir, Description: "Assisted-by: Claude:claude-3-opus\n"}
		def := decideOn(t, defaultSettings(), req)
		if !strings.Contains(def.AuthorModelStatement, "from description") {
			t.Fatal(def.AuthorModelStatement)
		}
		off := decideOn(t, mustParse(t, "author_description: off\n"), req)
		if !strings.Contains(off.AuthorModelStatement, "author model not stated") || strings.Contains(off.AuthorModelStatement, "claude-3-opus") {
			t.Fatal(off.AuthorModelStatement)
		}
	})

	t.Run("author_record", func(t *testing.T) {
		dir := treeWith(t, "go.mod", "module example.com/x\n")
		if err := os.WriteFile(filepath.Join(dir, "MODEL-a.txt"), []byte("Assisted-by: Ada:model-a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "MODEL-b.txt"), []byte("Assisted-by: Bea:model-b\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if strings.Contains(def.AuthorModelStatement, "from record") || strings.Contains(def.AuthorModelStatement, "model-a") {
			t.Fatal(def.AuthorModelStatement)
		}
		glob := decideOn(t, mustParse(t, "author_record: MODEL*.txt\n"), Request{Target: dir})
		if !strings.Contains(glob.AuthorModelStatement, "author Ada:model-a family family unknown from record") {
			t.Fatal(glob.AuthorModelStatement)
		}
		if !strings.Contains(glob.AuthorModelStatement, "author Bea:model-b family family unknown from record") {
			t.Fatal(glob.AuthorModelStatement)
		}
		mustRefuse(t, "author_record: /etc/passwd\n", "author_record", "absolute path")
		mustRefuse(t, "author_record: ../secret\n", "author_record", "path leaves the review target")
	})

	t.Run("author_trailer_keys", func(t *testing.T) {
		dir := treeWith(t, "go.mod", "module example.com/x\n")
		req := Request{Target: dir, Description: "Assisted-by: Claude:claude-3-opus\nModel-of-record: Qwen:qwen2.5\n"}
		def := decideOn(t, defaultSettings(), req)
		if !strings.Contains(def.AuthorModelStatement, "claude-3-opus") || strings.Contains(def.AuthorModelStatement, "qwen2.5") {
			t.Fatal(def.AuthorModelStatement)
		}
		other := decideOn(t, mustParse(t, "author_trailer_keys: Model-of-record\n"), req)
		if !strings.Contains(other.AuthorModelStatement, "qwen2.5") || strings.Contains(other.AuthorModelStatement, "claude-3-opus") {
			t.Fatal(other.AuthorModelStatement)
		}
	})

	t.Run("raise_class", func(t *testing.T) {
		if defaultSettings().RaiseClass {
			t.Fatal("raise_class defaults to on")
		}
		dir := treeWith(t, "billing/discount.go", "package billing\n")
		def := decideOn(t, defaultSettings(), Request{Target: dir})
		if def.Class != "backend" {
			t.Fatal(def.Class)
		}
		cfg := mustParse(t, "raise_class: on\n")
		if !cfg.RaiseClass {
			t.Fatal("raise_class on was not stored")
		}
		other := decideOn(t, cfg, Request{Target: dir})
		if other.Class != "backend" || other.ClassRule != "fallback" {
			t.Fatalf("raise_class changed the class to %s by %s", other.Class, other.ClassRule)
		}
	})
}

// mustParse parses body as proven.yaml and fails the test on a refusal.
func mustParse(t *testing.T, body string) fileSettings {
	t.Helper()
	cfg, err := parseConfig([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

// mustRefuse checks that body is refused on field, with reason and a non-empty allowed list.
func mustRefuse(t *testing.T, body, field, reason string) {
	t.Helper()
	_, err := parseConfig([]byte(body))
	cfg, ok := err.(*ConfigError)
	if !ok || cfg.Field != field || !strings.Contains(cfg.Reason, reason) || len(cfg.Allowed) == 0 {
		t.Fatalf("body %q err %v", body, err)
	}
}

// reviewSample runs one sample review of the discount fixture under root, with body as an explicit
// proven.yaml, or the defaults when body is empty.
func reviewSample(t *testing.T, body, root string) Result {
	t.Helper()
	opts := Options{Sample: true, LogDir: t.TempDir(), Env: os.Getenv}
	if body != "" {
		path := filepath.Join(t.TempDir(), "proven.yaml")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		opts.ConfigPath = path
	}
	rev, err := Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(t.Context(), Request{Target: filepath.Join(root, "testdata", "fixture", "discount")})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// moduleRoot walks up from the test's directory to the module root.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

// commitWithTrailer makes a two-commit repository whose head commit carries the trailer
// "Assisted-by: Ada:model-a", and returns the directory and both commits.
func commitWithTrailer(t *testing.T) (dir, base, head string) {
	t.Helper()
	dir = t.TempDir()
	gitAt(t, dir, "init")
	if err := os.MkdirAll(filepath.Join(dir, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "billing", "discount.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAt(t, dir, "add", "-A")
	gitAt(t, dir, "commit", "-m", "[fixture] change: base")
	base = strings.TrimSpace(gitAt(t, dir, "rev-parse", "HEAD"))
	if err := os.WriteFile(filepath.Join(dir, "billing", "discount.go"), []byte("package billing\n\nfunc Discount() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAt(t, dir, "add", "billing/discount.go")
	gitAt(t, dir, "commit", "-m", "[fixture] change: head\n\nAssisted-by: Ada:model-a")
	head = strings.TrimSpace(gitAt(t, dir, "rev-parse", "HEAD"))
	return dir, base, head
}

// gitAt runs git in dir with a fixed identity, no signing and a fixed default branch.
func gitAt(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{
		"-C", dir,
		"-c", "user.name=Proven Test",
		"-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
		"-c", "init.defaultBranch=main",
	}, args...)
	cmd := exec.Command("git", full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
