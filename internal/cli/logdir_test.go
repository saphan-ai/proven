package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestDefaultLogDirIsOutsideTheTargetAndPrivate proves the default trace
// directory is the user cache's proven/logs, outside the review target, and
// that the directory is created with mode 0700.
// Mutated field: the cache root the default log directory is built from,
// then the directory mode.
func TestDefaultLogDirIsOutsideTheTargetAndPrivate(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)
	cache := t.TempDir()
	xdg, home := setCacheRoot(t, cache)

	var got string
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "."}, emptyEnv, &stdout, &stderr, func(opts review.Options) (review.Reviewer, error) {
		got = opts.LogDir
		return &captureReviewer{}, nil
	})
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	inside := filepath.Join(target, "logs")
	if info, err := os.Stat(inside); err == nil {
		t.Fatalf("property: the default trace lives outside the review target; logs was created inside %s mode %o", inside, info.Mode().Perm())
	}
	resolved, err := filepath.EvalSymlinks(cache)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(resolved, "proven", "logs")
	if got != want {
		t.Fatalf("LogDir %q, want %q", got, want)
	}
	info, err := os.Stat(want)
	if err != nil {
		t.Fatalf("default log dir %s: %v", want, err)
	}
	if !info.IsDir() {
		t.Fatalf("default log dir %s is not a directory", want)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("default log dir mode %o, want 0700", info.Mode().Perm())
	}
	if insideTarget(t, want, target) {
		t.Fatalf("default log dir %s is inside the target %s", want, target)
	}
	// The real reviewer follows the directory the command chose. A sample scan
	// of this target must not create logs inside it.
	stdout.Reset()
	stderr.Reset()
	Run([]string{"scan", "--sample", "."}, emptyEnv, &stdout, &stderr, nil)
	if info, err := os.Stat(inside); err == nil {
		t.Fatalf("sample scan created logs inside %s mode %o\nstderr: %s", inside, info.Mode().Perm(), stderr.String())
	}
	assertOSCacheUntouched(t, xdg, home)
}

// TestConfiguredLogDirOutsideTheTargetIsUsed proves PROVEN_LOG_DIR replaces
// the default when the directory resolves outside the review target.
// Mutated field: PROVEN_LOG_DIR.
func TestConfiguredLogDirOutsideTheTargetIsUsed(t *testing.T) {
	target := t.TempDir()
	outside := t.TempDir()
	configured := filepath.Join(outside, "custom-logs")
	if err := os.MkdirAll(configured, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(configured, 0o755); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(outside)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(resolved, "custom-logs")
	var got string
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", target}, envOf(map[string]string{EnvLogDir: configured}), &stdout, &stderr, func(opts review.Options) (review.Reviewer, error) {
		got = opts.LogDir
		return &captureReviewer{}, nil
	})
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if got != want {
		t.Fatalf("LogDir %q, want %q", got, want)
	}
	info, err := os.Stat(want)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("configured log dir mode %o, want 0700", info.Mode().Perm())
	}
}

// TestConfiguredLogDirInsideTheTargetIsRefused proves a log directory that
// resolves inside the review target is refused at load with exit 7.
// The direct case names the path itself. The symlink cases name a path that
// looks outside until the links are resolved.
// Mutated field: PROVEN_LOG_DIR, direct and through a symlink.
func TestConfiguredLogDirInsideTheTargetIsRefused(t *testing.T) {
	target := t.TempDir()
	t.Chdir(t.TempDir())

	t.Run("direct", func(t *testing.T) {
		dir := filepath.Join(target, "logs")
		refuseLogDir(t, target, dir)
	})
	t.Run("symlink", func(t *testing.T) {
		real := filepath.Join(target, "hidden")
		if err := os.Mkdir(real, 0o755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(t.TempDir(), "logs-link")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		refuseLogDir(t, target, link)
	})
	t.Run("target symlink", func(t *testing.T) {
		alias := filepath.Join(t.TempDir(), "target-link")
		if err := os.Symlink(target, alias); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(target, "logs")
		refuseLogDir(t, alias, dir)
	})
}

// TestLogDirThatIsNotADirectoryIsRefused proves a PROVEN_LOG_DIR that is a
// file is refused at load, and the refusal names the allowed form.
// Mutated field: PROVEN_LOG_DIR.
func TestLogDirThatIsNotADirectoryIsRefused(t *testing.T) {
	target := t.TempDir()
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	opened := false
	code := Run([]string{"scan", target}, envOf(map[string]string{EnvLogDir: file}), &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		opened = true
		return &captureReviewer{}, nil
	})
	if code != exitConfig {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if opened {
		t.Fatal("reviewer opened after a refused log directory")
	}
	if !strings.Contains(stderr.String(), EnvLogDir) || !strings.Contains(stderr.String(), "allowed:") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

// TestDefaultLogDirInsideTheTargetIsRefused proves the default cache path is
// held to the same rule as a configured directory.
// Mutated field: the cache root, placed inside the review target.
func TestDefaultLogDirInsideTheTargetIsRefused(t *testing.T) {
	target := t.TempDir()
	cache := filepath.Join(target, "cache")
	if err := os.Mkdir(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	xdg, home := setCacheRoot(t, cache)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", target}, emptyEnv, &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		t.Fatal("reviewer opened")
		return nil, nil
	})
	if code != exitConfig {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	root, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), EnvLogDir) || !strings.Contains(stderr.String(), root) {
		t.Fatalf("stderr %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(cache, "proven", "logs")); err == nil {
		t.Fatal("default log dir was created inside the target")
	}
	assertOSCacheUntouched(t, xdg, home)
}

func refuseLogDir(t *testing.T, target, logDir string) {
	t.Helper()
	_, existed := os.Lstat(logDir)
	var stdout, stderr bytes.Buffer
	opened := false
	code := Run([]string{"scan", target}, envOf(map[string]string{EnvLogDir: logDir}), &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		opened = true
		return &captureReviewer{}, nil
	})
	if code != exitConfig {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if opened {
		t.Fatal("reviewer opened after a refused log directory")
	}
	msg := stderr.String()
	if !strings.Contains(msg, EnvLogDir) || !strings.Contains(msg, "allowed:") {
		t.Fatalf("stderr %q", msg)
	}
	root, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, root) {
		t.Fatalf("stderr does not name the target %s\n%s", root, msg)
	}
	if existed != nil {
		if _, err := os.Lstat(logDir); err == nil {
			t.Fatalf("refused log dir was created: %s", logDir)
		}
	}
}

// setCacheRoot points the default trace directory at dir.
// HOME and XDG_CACHE_HOME point elsewhere. Darwin's os.UserCacheDir
// reads the first and Unix reads the second. Neither is dir, so a
// default that calls os.UserCacheDir fails on every OS, and the write
// stays in a temporary directory.
func setCacheRoot(t *testing.T, dir string) (xdg, home string) {
	t.Helper()
	prev := userCacheDir
	userCacheDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userCacheDir = prev })
	home = t.TempDir()
	xdg = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", xdg)
	return xdg, home
}

// assertOSCacheUntouched proves the run did not follow either host rule
// for the user cache. XDG_CACHE_HOME is the Unix rule. HOME/Library/Caches
// is the Darwin rule.
func assertOSCacheUntouched(t *testing.T, xdg, home string) {
	t.Helper()
	if err := absentDir(filepath.Join(xdg, "proven", "logs")); err != nil {
		t.Fatalf("default log dir followed XDG_CACHE_HOME: %v", err)
	}
	darwin := filepath.Join(home, "Library", "Caches", "proven", "logs")
	if err := absentDir(darwin); err != nil {
		t.Fatalf("default log dir followed HOME/Library/Caches: %v", err)
	}
}

func absentDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("directory exists: %s", path)
}

func insideTarget(t *testing.T, dir, root string) bool {
	t.Helper()
	dir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		t.Fatal(err)
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
