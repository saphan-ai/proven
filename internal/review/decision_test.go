package review_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

func TestNoTrailerSaysAuthorModelNotStated(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Class != "security" || res.ClassRule != "go.mod" {
		t.Fatalf("class %s rule %s", res.Class, res.ClassRule)
	}
	if !res.Doubled {
		t.Fatal("security is not doubled")
	}
	if res.AuthorModelStatement != "author model not stated; default reviewer NVIDIA" {
		t.Fatal(res.AuthorModelStatement)
	}
	if len(res.ReviewerModels) != 1 || res.ReviewerModels[0] != "NVIDIA" {
		t.Fatalf("models %#v", res.ReviewerModels)
	}
	joined := strings.Join(res.ReviewerStatements, "\n")
	if !strings.Contains(joined, "reviewer NVIDIA; role default; family NVIDIA; author model not stated") {
		t.Fatal(joined)
	}
	if !strings.Contains(joined, "special-review model: not configured") {
		t.Fatal(joined)
	}
	if _, ok := find(res, "F1"); !ok {
		t.Fatal("F1 missing")
	}
}

func TestDescriptionNamesTheAuthorModel(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target:      fixturePath(t),
		Description: "Notes.\n\nAssisted-by: Claude:claude-3-opus\nAssisted-by: Nemotron:nemotron-super\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.AuthorModelStatement, "author Claude:claude-3-opus family Anthropic from description") {
		t.Fatal(res.AuthorModelStatement)
	}
	if !strings.Contains(res.AuthorModelStatement, "author Nemotron:nemotron-super family NVIDIA from description") {
		t.Fatal(res.AuthorModelStatement)
	}
	joined := strings.Join(res.ReviewerStatements, "\n")
	if !strings.Contains(joined, "claude-3-opus (Anthropic): no") || !strings.Contains(joined, "nemotron-super (NVIDIA): yes") {
		t.Fatal(joined)
	}
}

func TestDescriptionOffIgnoresThePullRequest(t *testing.T) {
	path := writeConfig(t, "author_description: off\n")
	rev, err := review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{
		Target:      fixturePath(t),
		Description: "Assisted-by: Claude:claude-3-opus\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.AuthorModelStatement, "author model not stated") {
		t.Fatal(res.AuthorModelStatement)
	}
}

func TestRecordFileOffByDefault(t *testing.T) {
	dir := copyFixtureTree(t)
	if err := os.WriteFile(filepath.Join(dir, "MODEL.txt"), []byte("Assisted-by: Qwen:qwen2.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.AuthorModelStatement, "author model not stated") {
		t.Fatal(res.AuthorModelStatement)
	}
	path := writeConfig(t, "author_record: MODEL.txt\n")
	rev, err = review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err = rev.Review(context.Background(), review.Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.AuthorModelStatement, "author Qwen:qwen2.5 family Alibaba from record") {
		t.Fatal(res.AuthorModelStatement)
	}
}

func TestCommitTrailerUsesTheConfiguredKey(t *testing.T) {
	dir := copyFixtureTree(t)
	gitRun(t, dir, "init")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-m", "[fixture] change: base")
	base := strings.TrimSpace(string(gitRun(t, dir, "rev-parse", "HEAD")))
	if err := os.MkdirAll(filepath.Join(dir, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "billing", "discount.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "billing/discount.go")
	gitRun(t, dir, "commit", "-m", "[fixture] change: head\n\nModel-of-record: Claude:claude-3-opus")
	head := strings.TrimSpace(string(gitRun(t, dir, "rev-parse", "HEAD")))

	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target: dir, BaselineCommit: base, HeadCommit: head,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.AuthorModelStatement, "author model not stated") {
		t.Fatalf("default key read a custom trailer: %s", res.AuthorModelStatement)
	}

	path := writeConfig(t, "author_trailer_keys: Model-of-record\n")
	rev, err = review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err = rev.Review(context.Background(), review.Request{
		Target: dir, BaselineCommit: base, HeadCommit: head,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.AuthorModelStatement, "author Claude:claude-3-opus family Anthropic from commit") {
		t.Fatal(res.AuthorModelStatement)
	}
	if res.Class != "backend" || res.ClassRule != "fallback" {
		t.Fatalf("class %s rule %s; the commit only touched discount.go", res.Class, res.ClassRule)
	}
	if res.Doubled {
		t.Fatal("backend was doubled")
	}
}

func TestSpecialReviewerIsAddedForSecurityOnly(t *testing.T) {
	path := writeConfig(t, "special_reviewer: qwen-special\n")
	rev, err := review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.ReviewerModels) != 2 || res.ReviewerModels[1] != "qwen-special" {
		t.Fatalf("%#v", res.ReviewerModels)
	}
	joined := strings.Join(res.ReviewerStatements, "\n")
	if strings.Contains(joined, "not configured") || !strings.Contains(joined, "role special") {
		t.Fatal(joined)
	}

	path = writeConfig(t, "special_reviewer: qwen-special\npath_rules: **/*.md=docs\n")
	rev, err = review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err = rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Class != "backend" || res.Doubled {
		t.Fatalf("class %s doubled %t", res.Class, res.Doubled)
	}
	if len(res.ReviewerModels) != 1 {
		t.Fatalf("%#v", res.ReviewerModels)
	}
}

func TestChangeCannotChooseItsOwnStrictness(t *testing.T) {
	dir := copyFixtureTree(t)
	gitRun(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "proven.yaml"), []byte("default_reviewer: from-base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-m", "[fixture] change: base")
	base := strings.TrimSpace(string(gitRun(t, dir, "rev-parse", "HEAD")))

	// A root go.mod is outside base/ and change/, so the sample fixture hash stays
	// put. Under the base rules it is security; the change's proven.yaml would make it docs.
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lower := "default_reviewer: from-change\npath_rules: **=docs\nclass_order: docs, user interface, mobile, backend, security\ndoubled_classes: off\n"
	if err := os.WriteFile(filepath.Join(dir, "proven.yaml"), []byte(lower), 0o644); err != nil {
		t.Fatal(err)
	}

	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target: dir, BaselineCommit: base,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Class != "security" || res.ClassRule != "go.mod" {
		t.Fatalf("class %s rule %s; the change's proven.yaml was applied", res.Class, res.ClassRule)
	}
	if !res.Doubled {
		t.Fatal("security was not doubled; the change turned doubling off")
	}
	if res.SettingsStatement != "the change edits proven.yaml" {
		t.Fatalf("settings statement %q", res.SettingsStatement)
	}
	if res.DefaultReviewer != "from-base" {
		t.Fatalf("reviewer %q; settings were not read from the baseline", res.DefaultReviewer)
	}
}

func TestSymlinkedRecordOutsideTheTargetIsRefused(t *testing.T) {
	dir := copyFixtureTree(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("Assisted-by: Outside:outside-model\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "MODEL.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	path := writeConfig(t, "author_record: MODEL.txt\n")
	rev, err := review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: dir})
	if err == nil {
		if strings.Contains(res.AuthorModelStatement, "outside-model") || strings.Contains(res.AuthorModelStatement, "from record") {
			t.Fatalf("symlink was read: %s", res.AuthorModelStatement)
		}
		return
	}
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "author_record" || !strings.Contains(cfg.Reason, "symlink") {
		t.Fatal(err)
	}
	if !strings.Contains(cfg.Error(), "MODEL.txt") {
		t.Fatalf("refusal does not name the link: %s", cfg.Error())
	}
}

func TestSymlinkedRecordGlobIsRefused(t *testing.T) {
	dir := copyFixtureTree(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("Assisted-by: Outside:outside-model\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "MODEL.txt")); err != nil {
		t.Fatal(err)
	}
	path := writeConfig(t, "author_record: MODEL*.txt\n")
	rev, err := review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: dir})
	if err == nil {
		if strings.Contains(res.AuthorModelStatement, "outside-model") || strings.Contains(res.AuthorModelStatement, "from record") {
			t.Fatalf("symlink was read: %s", res.AuthorModelStatement)
		}
		return
	}
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "author_record" || !strings.Contains(cfg.Reason, "symlink") {
		t.Fatal(err)
	}
}

func TestUnknownFamilyAndUnknownClassAreRefusedOnLoad(t *testing.T) {
	path := writeConfig(t, "families: gemini-*=Google\n")
	_, err := review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "families" || cfg.Value != "Google" || !strings.Contains(cfg.Reason, "unknown family") {
		t.Fatal(err)
	}
	path = writeConfig(t, "path_rules: **/*.go=frontend\n")
	_, err = review.Open(review.Options{ConfigPath: path, Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	cfg = nil
	if !errors.As(err, &cfg) || cfg.Field != "path_rules" || cfg.Value != "frontend" || !strings.Contains(cfg.Reason, "unknown class") {
		t.Fatal(err)
	}
}
