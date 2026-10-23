package review

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestLinkedConfigIsRefusedByName proves a proven.yaml that is a symbolic link is refused (ADR-0013,
// ADR-0025): the refusal names the file and never carries what the link points at. Mutated field:
// only the file type; the linked file is a valid configuration.
func TestLinkedConfigIsRefusedByName(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "elsewhere.yaml")
	if err := os.WriteFile(outside, []byte("timeout: 5m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(target, "proven.yaml")); err != nil {
		t.Fatal(err)
	}
	_, err := loadSettingsInTarget(target)
	if err == nil {
		t.Fatal("a symbolic link named proven.yaml was read")
	}
	if !strings.Contains(err.Error(), "proven.yaml") || strings.Contains(err.Error(), "5m") {
		t.Fatalf("the refusal must name the file and not quote the linked content: %v", err)
	}
}

// TestRegularConfigInTheTargetIsStillRead proves the link refusal does not refuse a regular file.
func TestRegularConfigInTheTargetIsStillRead(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "proven.yaml"), []byte("timeout: 5m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadSettingsInTarget(target)
	if err != nil || cfg.Timeout.String() != "5m0s" {
		t.Fatalf("a regular proven.yaml was not read: %v %v", cfg.Timeout, err)
	}
}

// TestBaselineConfigOverTheConfigBoundIsRefused proves the proven.yaml read from the baseline commit
// has the configuration bound, which is smaller than the file bound a blob has (ADR-0013). Mutated
// field: only the size; the file is under the file bound.
func TestBaselineConfigOverTheConfigBoundIsRefused(t *testing.T) {
	dir := t.TempDir()
	body := "timeout: 5m\n" + strings.Repeat("# padding-line\n", (2<<20)/15)
	if err := os.WriteFile(filepath.Join(dir, "proven.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	baseline := commitAll(t, dir)
	_, err := loadSettingsAtBaseline(dir, baseline)
	if err == nil {
		t.Fatalf("a %d-byte proven.yaml was read from the baseline", len(body))
	}
	if !strings.Contains(err.Error(), "proven.yaml") || strings.Contains(err.Error(), "padding-line") {
		t.Fatalf("the refusal must name the file and not quote it: %v", err)
	}
}

// TestAuthorRecordIsBoundedAsASum proves the record bound covers every file the setting names
// together: each file is under the bound and their sum is over it (ADR-0013).
func TestAuthorRecordIsBoundedAsASum(t *testing.T) {
	target := t.TempDir()
	half := strings.Repeat("x", 600<<10)
	for _, name := range []string{"author-a.txt", "author-b.txt"} {
		if err := os.WriteFile(filepath.Join(target, name), []byte(half), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := readRecord(target, "author-a.txt"); err != nil {
		t.Fatalf("one file under the bound was refused: %v", err)
	}
	_, err := readRecord(target, "author-*.txt")
	if err == nil {
		t.Fatal("two files whose sum is over the record bound were read")
	}
	if !strings.Contains(err.Error(), "author-*.txt") || strings.Contains(err.Error(), "xxxx") {
		t.Fatalf("the refusal must name the record and not quote it: %v", err)
	}
}

// TestOpenAppliesTheSizeSettings proves Open reads the size bounds from its environment, refuses a
// wrong value without replacing it by the default, and that a sample-mode Open leaves the defaults in
// force instead of a previous review's bounds (ADR-0013).
func TestOpenAppliesTheSizeSettings(t *testing.T) {
	defer change.UseLimits(change.UseLimits(change.DefaultLimits()))
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}

	if _, err := Open(Options{LogDir: t.TempDir(), Env: env(map[string]string{change.EnvMaxFileBytes: "1234"})}); err != nil {
		t.Fatal(err)
	}
	if got := change.Bounds().File; got != 1234 {
		t.Fatalf("Open did not apply %s: file bound %d", change.EnvMaxFileBytes, got)
	}

	// A wrong value is refused, and the bound in force stays the last good one.
	_, err := Open(Options{LogDir: t.TempDir(), Env: env(map[string]string{change.EnvMaxFileBytes: "lots"})})
	var cfg *ConfigError
	if !errors.As(err, &cfg) || !strings.Contains(cfg.Field, change.EnvMaxFileBytes) {
		t.Fatalf("a wrong size setting was not refused by name: %v", err)
	}

	if _, err := Open(Options{Sample: true, LogDir: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if got := change.Bounds(); got != change.DefaultLimits() {
		t.Fatalf("a sample-mode Open kept the bounds of an earlier review: %+v", got)
	}
}

// commitAll makes dir a repository with everything in it committed and returns the commit.
func commitAll(t *testing.T, dir string) string {
	t.Helper()
	run := func(args ...string) string {
		t.Helper()
		full := append([]string{
			"-c", "user.name=Proven Test", "-c", "user.email=test@example.com",
			"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init")
	run("add", ".")
	run("commit", "-m", "[review] test: first")
	return run("rev-parse", "HEAD")
}
