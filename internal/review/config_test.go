package review_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/review"
)

func TestUnknownConfigKeyIsRefused(t *testing.T) {
	path := writeConfig(t, "colour: red\n")
	_, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) {
		t.Fatal(err)
	}
	if cfg.Field != "colour" {
		t.Fatalf("field %q", cfg.Field)
	}
	if cfg.Error() != `configuration refused: colour="red": unknown key; allowed: timeout, default_reviewer, special_reviewer, families, path_rules, class_order, fallback_class, doubled_classes, author_commit_trailers, author_description, author_record, author_trailer_keys, raise_class` {
		t.Fatalf("error %q", cfg.Error())
	}
}

func TestBadTimeoutIsRefused(t *testing.T) {
	path := writeConfig(t, "timeout: tomorrow\n")
	_, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "timeout" || cfg.Value != "tomorrow" {
		t.Fatal(err)
	}
}

func TestNegativeTimeoutIsRefused(t *testing.T) {
	path := writeConfig(t, "timeout: -1s\n")
	_, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "timeout" {
		t.Fatal(err)
	}
}

func TestMissingConfigFileIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such.yaml")
	_, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "ConfigPath" || cfg.Value != path {
		t.Fatal(err)
	}
}

func TestTimeoutBoundsAReviewWithNoDeadline(t *testing.T) {
	path := writeConfig(t, "timeout: 1ns\n")
	rev, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Partial {
		t.Fatal("a 1ns timeout must bound a review whose context has no deadline")
	}
}

func TestParentDeadlineWinsOverConfigTimeout(t *testing.T) {
	path := writeConfig(t, "timeout: 1ns\n")
	rev, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	res, err := rev.Review(ctx, review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("the context deadline wins; the config timeout must not replace it")
	}
	if _, ok := find(res, "F1"); !ok {
		t.Fatal("F1 missing")
	}
}

func TestProvenYAMLInTargetIsRefusedOnReview(t *testing.T) {
	dir := copyFixtureTree(t)
	if err := os.WriteFile(filepath.Join(dir, "proven.yaml"), []byte("colour: red\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rev, err := review.Open(review.Options{Sample: true, LogDir: t.TempDir(), Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	_, err = rev.Review(context.Background(), review.Request{Target: dir})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "colour" {
		t.Fatal(err)
	}
}

func TestExplicitConfigIgnoresTargetFile(t *testing.T) {
	dir := copyFixtureTree(t)
	if err := os.WriteFile(filepath.Join(dir, "proven.yaml"), []byte("colour: red\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := writeConfig(t, "timeout: 1h\n")
	rev, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := find(res, "F1"); !ok {
		t.Fatal("F1 missing")
	}
}

func TestZeroTimeoutLeavesTheReviewUnbounded(t *testing.T) {
	path := writeConfig(t, "timeout: 0s\n")
	rev, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("0s must not add a deadline")
	}
	if _, ok := find(res, "F1"); !ok {
		t.Fatal("F1 missing")
	}
}

func TestQuotedTimeoutWithCommentBoundsTheReview(t *testing.T) {
	path := writeConfig(t, "# review bound\ntimeout: \"1ns\" # soon\n")
	rev, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Partial {
		t.Fatal("quoted timeout with a comment was not applied")
	}
}

func TestDuplicateTimeoutIsRefused(t *testing.T) {
	path := writeConfig(t, "timeout: 1h\ntimeout: 1ns\n")
	_, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "timeout" {
		t.Fatal(err)
	}
}

func TestNestedConfigIsRefused(t *testing.T) {
	path := writeConfig(t, "timeout:\n  seconds: 1\n")
	_, err := review.Open(review.Options{
		ConfigPath: path,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	})
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) {
		t.Fatal(err)
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
