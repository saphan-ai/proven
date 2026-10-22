package review_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// Request-contract tests. At this base every case fails on behaviour except
// TestContractW1cT8Defaults, which guards an empty configuration path.

func TestContractW1cT1Exclude(t *testing.T) {
	rev := contractW1cOpen(t, "")
	res, err := rev.Review(context.Background(), review.Request{
		Target:  contractW1cFixture(t),
		Exclude: []string{"*"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("Exclude ignored: %d findings remain (%s)", len(res.Findings), contractW1cLocations(res))
	}
}

func TestContractW1cT2Include(t *testing.T) {
	rev := contractW1cOpen(t, "")
	res, err := rev.Review(context.Background(), review.Request{
		Target:  contractW1cFixture(t),
		Include: []string{"go.mod"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if contractW1cHasPath(res, "billing/discount.go") {
		t.Fatalf("Include ignored: finding remains at billing/discount.go (%s)", contractW1cLocations(res))
	}
	if !contractW1cHasIDAt(res, "F2", "go.mod") {
		t.Fatalf("Include ignored: want F2 at go.mod, got %s", contractW1cLocations(res))
	}
	for _, f := range res.Findings {
		if f.Location.Path != "go.mod" {
			t.Fatalf("Include ignored: finding %s at %s", f.ID, f.Location.Path)
		}
	}
}

func TestContractW1cT3ExcludeWins(t *testing.T) {
	rev := contractW1cOpen(t, "")
	res, err := rev.Review(context.Background(), review.Request{
		Target:  contractW1cFixture(t),
		Include: []string{"*"},
		Exclude: []string{"discount.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if contractW1cHasPath(res, "billing/discount.go") {
		t.Fatalf("Include checked after Exclude: finding remains at billing/discount.go (%s)", contractW1cLocations(res))
	}
	if !contractW1cHasIDAt(res, "F2", "go.mod") {
		t.Fatalf("want only go.mod findings, got %s", contractW1cLocations(res))
	}
	for _, f := range res.Findings {
		if f.Location.Path != "go.mod" {
			t.Fatalf("want only go.mod findings, got %s at %s", f.ID, f.Location.Path)
		}
	}
}

func TestContractW1cT4MissingBaseline(t *testing.T) {
	rev := contractW1cOpen(t, "")
	res, err := rev.Review(context.Background(), review.Request{
		Target:         contractW1cFixture(t),
		BaselineCommit: "does-not-exist",
	})
	if err == nil {
		t.Fatalf("BaselineCommit ignored: review returned %d findings and a nil error (%s)", len(res.Findings), contractW1cLocations(res))
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Fatalf("error %q does not name does-not-exist", err.Error())
	}
}

func TestContractW1cT5UnknownKey(t *testing.T) {
	path := contractW1cWriteConfig(t, "colour: red\n")
	_, err := review.Open(contractW1cOptions(t, path))
	cfg := contractW1cAsConfig(err)
	if cfg == nil {
		t.Fatalf("ConfigPath ignored: Open error = %v, want *ConfigError for colour", err)
	}
	if cfg.Field != "colour" {
		t.Fatalf("Field = %q, want colour", cfg.Field)
	}
	if len(cfg.Allowed) == 0 || !contractW1cContains(cfg.Allowed, "timeout") {
		t.Fatalf("Allowed = %q, want a non-empty set containing timeout", cfg.Allowed)
	}
}

func TestContractW1cT6BadTimeout(t *testing.T) {
	path := contractW1cWriteConfig(t, "timeout: banana\n")
	_, err := review.Open(contractW1cOptions(t, path))
	cfg := contractW1cAsConfig(err)
	if cfg == nil || cfg.Field != "timeout" {
		t.Fatalf("timeout value not parsed: err=%v field=%s", err, contractW1cField(cfg))
	}
}

func TestContractW1cT7MissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-config.yaml")
	_, err := review.Open(contractW1cOptions(t, missing))
	cfg := contractW1cAsConfig(err)
	if cfg == nil {
		t.Fatalf("missing file read as defaults: err=%v", err)
	}
	if !strings.Contains(cfg.Value, missing) && !strings.Contains(cfg.Reason, missing) {
		t.Fatalf("Value %q Reason %q do not name %s", cfg.Value, cfg.Reason, missing)
	}
}

func TestContractW1cT8Defaults(t *testing.T) {
	rev := contractW1cOpen(t, "")
	res, err := rev.Review(context.Background(), review.Request{Target: contractW1cFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !contractW1cHasIDAt(res, "F1", "billing/discount.go") {
		t.Fatalf("F1 missing at billing/discount.go (%s)", contractW1cLocations(res))
	}
	if !contractW1cHasIDAt(res, "F2", "go.mod") {
		t.Fatalf("F2 missing at go.mod (%s)", contractW1cLocations(res))
	}
}

func TestContractW1cT9TimeoutBounds(t *testing.T) {
	path := contractW1cWriteConfig(t, "timeout: 1ns\n")
	rev := contractW1cOpen(t, path)
	res, err := rev.Review(context.Background(), review.Request{Target: contractW1cFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial || contractW1cProvedCount(res) == 0 {
		return
	}
	t.Fatalf("config timeout never applied: Partial=false and %d findings proved (%s)", contractW1cProvedCount(res), contractW1cLocations(res))
}

func contractW1cOpen(t *testing.T, configPath string) review.Reviewer {
	t.Helper()
	rev, err := review.Open(contractW1cOptions(t, configPath))
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

func contractW1cOptions(t *testing.T, configPath string) review.Options {
	t.Helper()
	return review.Options{
		ConfigPath: configPath,
		Sample:     true,
		LogDir:     t.TempDir(),
		Env:        os.Getenv,
	}
}

func contractW1cWriteConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func contractW1cFixture(t *testing.T) string {
	t.Helper()
	return filepath.Join(contractW1cRepoRoot(t), "testdata", "fixture", "discount")
}

func contractW1cRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, readErr := os.ReadFile(filepath.Join(dir, "go.mod"))
		if readErr == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

func contractW1cAsConfig(err error) *review.ConfigError {
	var cfg *review.ConfigError
	if !errors.As(err, &cfg) {
		return nil
	}
	return cfg
}

func contractW1cField(cfg *review.ConfigError) string {
	if cfg == nil {
		return ""
	}
	return cfg.Field
}

func contractW1cContains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func contractW1cHasPath(res review.Result, path string) bool {
	for _, f := range res.Findings {
		if f.Location.Path == path {
			return true
		}
	}
	return false
}

func contractW1cHasIDAt(res review.Result, id, path string) bool {
	for _, f := range res.Findings {
		if f.ID == id && f.Location.Path == path {
			return true
		}
	}
	return false
}

func contractW1cLocations(res review.Result) string {
	parts := make([]string, 0, len(res.Findings))
	for _, f := range res.Findings {
		parts = append(parts, f.ID+"@"+f.Location.Path)
	}
	return strings.Join(parts, ", ")
}

func contractW1cProvedCount(res review.Result) int {
	n := 0
	for _, f := range res.Findings {
		if f.Unresolved || len(f.Evidence) == 0 {
			continue
		}
		n++
	}
	return n
}
