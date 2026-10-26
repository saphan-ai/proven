package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
	// The real reviewer follows the directory the command chose. A sample
	// review of the fixture must exit 0, write trace.jsonl in that directory,
	// and must not create logs inside the fixture.
	fixture := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"scan", "--sample", fixture}, emptyEnv, &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("sample review exit %d\n%s", code, stderr.String())
	}
	tracePath := filepath.Join(want, "trace.jsonl")
	if _, err := os.Stat(tracePath); err != nil {
		t.Fatalf("trace file %s: %v", tracePath, err)
	}
	if info, err := os.Stat(filepath.Join(fixture, "logs")); err == nil {
		t.Fatalf("sample scan created logs inside %s mode %o\nstderr: %s", fixture, info.Mode().Perm(), stderr.String())
	}
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
	t.Run("..", func(t *testing.T) {
		// Join would clean "..". The value the command sees still contains it.
		base := filepath.Base(target)
		dir := target + string(filepath.Separator) + ".." + string(filepath.Separator) + base + string(filepath.Separator) + "logs"
		refuseLogDir(t, target, dir)
	})
}

// TestTraceFileSymlinkIntoTheTargetIsRefused proves a trace.jsonl symlink
// in a directory outside the target is refused when the link lands inside
// the target. The directory check alone follows that link on open.
// Mutated field: trace.jsonl in PROVEN_LOG_DIR.
func TestTraceFileSymlinkIntoTheTargetIsRefused(t *testing.T) {
	target := t.TempDir()
	copyTree(t, filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"), target)
	outside := t.TempDir()
	victim := filepath.Join(target, "trace.jsonl")
	if err := os.Symlink(victim, filepath.Join(outside, "trace.jsonl")); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{EnvLogDir: outside}), &stdout, &stderr, nil)
	if code != exitConfig {
		_, statErr := os.Lstat(victim)
		t.Fatalf("exit %d target trace exists=%v stderr %s", code, statErr == nil, stderr.String())
	}
	if _, err := os.Lstat(victim); !os.IsNotExist(err) {
		body, _ := os.ReadFile(victim)
		t.Fatalf("trace file inside the target: err %v body %q", err, body)
	}
	root, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	msg := stderr.String()
	if !strings.Contains(msg, EnvLogDir) || !strings.Contains(msg, root) {
		t.Fatalf("stderr %q", msg)
	}
}

// TestTraceFileHardLinkIntoTheTargetIsRefused proves a regular trace.jsonl
// outside the target is refused when that file is a hard link to a file
// inside the target. The two names are one file, so an append through the
// outside name writes the target. The open does not follow a path it can
// be compared with; the link count is what shows the second name.
// Unplug: the link-count check in traceFileInsideTarget. The load exits 0
// and the target file's bytes change.
// Mutated field: the link count of trace.jsonl in PROVEN_LOG_DIR.
func TestTraceFileHardLinkIntoTheTargetIsRefused(t *testing.T) {
	// One parent, then two directories. A hard link has to share a filesystem.
	parent := t.TempDir()
	target := filepath.Join(parent, "target")
	copyTree(t, filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"), target)
	victim := filepath.Join(target, "kept.txt")
	// Empty, so an append is the whole change. A first line that is not
	// JSON makes the run fail for that line instead of for the link.
	before := []byte{}
	if err := os.WriteFile(victim, before, 0o644); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	tracePath := filepath.Join(outside, "trace.jsonl")
	linked := victim
	if cross := os.Link(victim, tracePath); cross != nil {
		if !errors.Is(cross, syscall.EXDEV) {
			t.Skipf("hard link not available here: %v", cross)
		}
		// The OS refused a link whose names are in different directories.
		// The second name is created in the log directory instead.
		// The load reads the link count and not the other path.
		linked = filepath.Join(outside, "kept.txt")
		if err := os.WriteFile(linked, before, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(linked, tracePath); err != nil {
			t.Skipf("hard link not available here: %v", err)
		}
		t.Logf("measured cross-directory link: %v", cross)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{EnvLogDir: outside}), &stdout, &stderr, nil)
	if code != exitConfig {
		after, readErr := os.ReadFile(linked)
		t.Fatalf("exit %d linked bytes changed=%v read %v stderr %s", code, readErr == nil && !bytes.Equal(before, after), readErr, stderr.String())
	}
	after, err := os.ReadFile(linked)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("linked file was written: before %q after %q", before, after)
	}
	msg := stderr.String()
	if !strings.Contains(msg, EnvLogDir) || !strings.Contains(msg, "trace file has 2 names") || !strings.Contains(msg, "remove or rename "+traceFile) {
		t.Fatalf("stderr %q", msg)
	}
}

// TestTraceFileSymlinkToALinkedFileIsRefused proves a trace.jsonl that is a
// symlink is refused when the file the open would write has another name.
// The directory check follows the link only to see where it lands. A landing
// outside the target used to pass, and the append then wrote every name of
// that file, including a name inside the target.
// Unplug: the link-count check of the symlink's landing. The load exits 0
// and a name of that file changes.
// Mutated field: the link count of the file trace.jsonl names.
func TestTraceFileSymlinkToALinkedFileIsRefused(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "target")
	copyTree(t, filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"), target)
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	outsideFile := filepath.Join(outside, "kept.txt")
	before := []byte{}
	if err := os.WriteFile(outsideFile, before, 0o644); err != nil {
		t.Fatal(err)
	}
	insideName := filepath.Join(target, "kept.txt")
	second := insideName
	if cross := os.Link(outsideFile, insideName); cross != nil {
		if !errors.Is(cross, syscall.EXDEV) {
			t.Skipf("hard link not available here: %v", cross)
		}
		// The OS refused a link whose names are in different directories.
		// The second name is created beside the outside file instead.
		// The load reads the link count of the landing and not the other path.
		second = filepath.Join(outside, "kept-link.txt")
		if err := os.Link(outsideFile, second); err != nil {
			t.Skipf("hard link not available here: %v", err)
		}
		t.Logf("measured cross-directory link: %v", cross)
	}
	if err := os.Symlink(outsideFile, filepath.Join(outside, traceFile)); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{EnvLogDir: outside}), &stdout, &stderr, nil)
	outsideAfter, outsideErr := os.ReadFile(outsideFile)
	secondAfter, secondErr := os.ReadFile(second)
	outsideChanged := outsideErr == nil && !bytes.Equal(before, outsideAfter)
	secondChanged := secondErr == nil && !bytes.Equal(before, secondAfter)
	if code != exitConfig || outsideChanged || secondChanged {
		t.Fatalf("exit %d outside changed=%v second changed=%v read %v %v stderr %s", code, outsideChanged, secondChanged, outsideErr, secondErr, stderr.String())
	}
	msg := stderr.String()
	if !strings.Contains(msg, EnvLogDir) || !strings.Contains(msg, "trace file is a symbolic link") || !strings.Contains(msg, "remove the link") {
		t.Fatalf("stderr %q", msg)
	}
}

// siblingLogFiles are the names, other than trace.jsonl, that Proven opens
// directly in the log directory. world.Store.audit writes world.jsonl.
// hmem.Store.audit writes hmem.jsonl. Both opens follow a final symlink.
var siblingLogFiles = []string{"world.jsonl", "hmem.jsonl"}

// TestSiblingLogFileSymlinkIntoTheTargetIsRefused proves a world.jsonl or
// hmem.jsonl symlink in a directory outside the target is refused when the
// link lands inside the target. The directory check alone follows that link
// on open.
// Unplug: the symlink-landing check for that file name. The load exits 0.
// For world.jsonl the target file is created. For hmem.jsonl the sample
// fixture writes no audit line, so the target file stays absent and the
// exit code is the difference.
// Mutated field: that file name in PROVEN_LOG_DIR.
func TestSiblingLogFileSymlinkIntoTheTargetIsRefused(t *testing.T) {
	for _, name := range siblingLogFiles {
		t.Run(name, func(t *testing.T) {
			target := t.TempDir()
			copyTree(t, filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"), target)
			outside := t.TempDir()
			victim := filepath.Join(target, name)
			if err := os.Symlink(victim, filepath.Join(outside, name)); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{EnvLogDir: outside}), &stdout, &stderr, nil)
			_, statErr := os.Lstat(victim)
			nbytes := 0
			if statErr == nil {
				body, readErr := os.ReadFile(victim)
				if readErr == nil {
					nbytes = len(body)
				}
			}
			if code != exitConfig {
				t.Fatalf("exit %d target file exists=%v bytes=%d stat %v stderr %s", code, statErr == nil, nbytes, statErr, stderr.String())
			}
			if _, err := os.Lstat(victim); !os.IsNotExist(err) {
				body, _ := os.ReadFile(victim)
				t.Fatalf("log file inside the target: err %v bytes %d body %q", err, len(body), body)
			}
			root, err := filepath.EvalSymlinks(target)
			if err != nil {
				t.Fatal(err)
			}
			msg := stderr.String()
			if !strings.Contains(msg, EnvLogDir) || !strings.Contains(msg, name) || !strings.Contains(msg, root) {
				t.Fatalf("stderr %q", msg)
			}
		})
	}
}

// TestSiblingLogFileHardLinkIntoTheTargetIsRefused proves a regular
// world.jsonl or hmem.jsonl outside the target is refused when that file
// has a second name. An append through the outside name writes every name.
// Unplug: the link-count check for that file name. The load exits 0.
// Mutated field: the link count of that file in PROVEN_LOG_DIR.
func TestSiblingLogFileHardLinkIntoTheTargetIsRefused(t *testing.T) {
	for _, name := range siblingLogFiles {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			target := filepath.Join(parent, "target")
			copyTree(t, filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"), target)
			before := []byte{}
			victim := filepath.Join(target, "kept-"+name)
			if err := os.WriteFile(victim, before, 0o644); err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(parent, "outside")
			if err := os.MkdirAll(outside, 0o755); err != nil {
				t.Fatal(err)
			}
			logPath := filepath.Join(outside, name)
			linked := victim
			if cross := os.Link(victim, logPath); cross != nil {
				if !errors.Is(cross, syscall.EXDEV) {
					t.Skipf("hard link not available here: %v", cross)
				}
				linked = filepath.Join(outside, "kept-"+name)
				if err := os.WriteFile(linked, before, 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(linked, logPath); err != nil {
					t.Skipf("hard link not available here: %v", err)
				}
				t.Logf("measured cross-directory link: %v", cross)
			}
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{EnvLogDir: outside}), &stdout, &stderr, nil)
			after, readErr := os.ReadFile(linked)
			changed := readErr == nil && !bytes.Equal(before, after)
			if code != exitConfig || changed {
				t.Fatalf("exit %d linked bytes changed=%v read %v stderr %s", code, changed, readErr, stderr.String())
			}
			root, err := filepath.EvalSymlinks(target)
			if err != nil {
				t.Fatal(err)
			}
			msg := stderr.String()
			if !strings.Contains(msg, EnvLogDir) || !strings.Contains(msg, name) || !strings.Contains(msg, root) || !strings.Contains(msg, "remove or rename "+name) {
				t.Fatalf("stderr %q", msg)
			}
		})
	}
}

// TestLogDirCaseAliasInsideTheTargetIsRefused proves a log directory that
// differs from the target only by letter case is refused. That refusal is
// exercised only on volumes that ignore case (darwin, Windows). On any
// other volume this test skips, and the skip is a measurement of this
// temp filesystem, not an assumption.
// Mutated field: letter case of PROVEN_LOG_DIR.
func TestLogDirCaseAliasInsideTheTargetIsRefused(t *testing.T) {
	folds, detail := tempFilesystemFoldsCase(t)
	if !folds {
		t.Skip("measured case-sensitive temp filesystem: " + detail)
	}
	parent := t.TempDir()
	target := filepath.Join(parent, "Repo")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "REPO", "logs")
	refuseLogDir(t, target, alias)
	if _, err := os.Stat(filepath.Join(target, "logs")); err == nil {
		t.Fatal("letter-case alias created logs inside the target")
	}
}

// tempFilesystemFoldsCase probes the temp filesystem. It creates a directory
// named Repo and stats REPO. The same file means letter case is not distinct.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, in)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func tempFilesystemFoldsCase(t *testing.T) (bool, string) {
	t.Helper()
	dir := t.TempDir()
	upper := filepath.Join(dir, "Repo")
	if err := os.Mkdir(upper, 0o755); err != nil {
		t.Fatal(err)
	}
	lower := filepath.Join(dir, "REPO")
	lowerInfo, err := os.Stat(lower)
	if os.IsNotExist(err) {
		return false, "Stat(REPO) after Mkdir(Repo): not exist"
	}
	if err != nil {
		t.Fatal(err)
	}
	upperInfo, err := os.Stat(upper)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(upperInfo, lowerInfo) {
		return true, "Stat(REPO) and Stat(Repo) are the same file"
	}
	return false, "Stat(REPO) exists and is a different file"
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

// TestLogFileLinkThroughDotDotIsRefused proves a log-directory file that is a
// symbolic link is refused even when its text, cleaned as text, points
// outside the target: sub/../x with sub a link into the target lands inside
// once the kernel resolves sub first. Mutated field: the symlink branch of
// logFileRefusal (judge the cleaned landing instead of refusing).
func TestLogFileLinkThroughDotDotIsRefused(t *testing.T) {
	// A real fixture, so a run that is not refused does write its logs and
	// the "not written inside the target" check can fail.
	target := t.TempDir()
	copyTree(t, filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"), target)
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "deep"), filepath.Join(outside, "sub")); err != nil {
		t.Skipf("symlinks not available here: %v", err)
	}
	for _, name := range logDirFileNames {
		if err := os.Symlink(filepath.Join("sub", "..", "x-"+name), filepath.Join(outside, name)); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{EnvLogDir: outside}), &stdout, &stderr, nil)
	if code != exitConfig {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	for _, name := range logDirFileNames {
		if _, err := os.Stat(filepath.Join(target, "x-"+name)); err == nil {
			t.Fatalf("%s was written inside the target", name)
		}
	}
	if !strings.Contains(stderr.String(), "is a symbolic link") {
		t.Fatalf("stderr %q", stderr.String())
	}
}
