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

// TestUnknownConfigKeyIsRefused proves that a key outside the closed set is refused on Open,
// and that the refusal lists every allowed key in order (ADR-0013).
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

// TestBadTimeoutIsRefused proves that a timeout that is not a Go duration is refused, naming the field and the value.
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

// TestNegativeTimeoutIsRefused proves that a negative duration is refused rather than read as "no deadline".
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

// TestMissingConfigFileIsRefused proves that an explicit --config path that does not exist is refused,
// not replaced by the defaults.
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

// TestTimeoutBoundsAReviewWithNoDeadline proves that the timeout setting bounds a review when the shell set
// no deadline: the review comes back Partial, not as an error.
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

// TestParentDeadlineWinsOverConfigTimeout proves that the shell's deadline wins over the timeout setting:
// a 1ns setting does not cut short a review the shell gave an hour.
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

// TestProvenYAMLInTargetIsRefusedOnReview proves that, with no --config, a bad proven.yaml in the review
// target is refused when Review reads it.
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

// TestExplicitConfigIgnoresTargetFile proves that an explicit --config replaces the target's proven.yaml:
// the target's bad file is never read.
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

// TestZeroTimeoutLeavesTheReviewUnbounded proves that a 0s timeout adds no deadline.
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

// TestQuotedTimeoutWithCommentBoundsTheReview proves that a quoted value with a trailing comment, after a
// comment line, is read as its value and applied.
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

// TestDuplicateTimeoutIsRefused proves that a key set twice is refused rather than the last line winning.
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

// TestNestedConfigIsRefused proves that a nested mapping is refused: proven.yaml is one flat mapping.
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

// writeConfig writes body to a proven.yaml in a fresh temporary directory and returns its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
