package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saphan-ai/proven/internal/trace"
)

// EnvLogDir is the trace directory setting, declared beside the trace
// (trace.EnvLogDir) so the configuration page can list it without importing
// this package.
const EnvLogDir = trace.EnvLogDir

// logDirAllowed is the form a refused log directory names (ADR-0013).
var logDirAllowed = []string{"an absolute path outside the review target"}

// resolveLogDir chooses the trace directory and creates it, mode 0700.
// A directory that resolves inside the review target is refused before it
// is created. The trace is the record of the review; it must not land in
// the repository under review, where it would be readable with that tree.
func resolveLogDir(getenv func(string) string, target string) (string, error) {
	raw := getenv(EnvLogDir)
	if strings.TrimSpace(raw) == "" {
		if raw != "" {
			return "", refuse(EnvLogDir, raw, "log directory is empty", logDirAllowed)
		}
		dir, err := defaultLogDir()
		if err != nil {
			return "", err
		}
		raw = dir
	} else if strings.ContainsAny(raw, "\x00\n\r") {
		return "", refuse(EnvLogDir, raw, "log directory is not a path", logDirAllowed)
	}
	resolved, err := resolvePath(raw)
	if err != nil {
		return "", refuse(EnvLogDir, raw, "log directory cannot be resolved", logDirAllowed)
	}
	root, err := resolvePath(target)
	if err != nil {
		return "", refuse(EnvLogDir, raw, "review target cannot be resolved", logDirAllowed)
	}
	inside, err := pathInside(resolved, root)
	if err != nil {
		return "", refuse(EnvLogDir, raw, "log directory cannot be compared with the review target", logDirAllowed)
	}
	if inside {
		// Refuse before creating the directory. A trace inside the review
		// target would put the review's record in the repository it reviews.
		return "", refuse(EnvLogDir, raw, "log directory resolves inside the review target "+root, logDirAllowed)
	}
	// The directory can sit outside the target while a file Proven writes in
	// it is a symlink into the target, or one name of a file that has another
	// name. Each open follows a symlink and writes every name of that file,
	// so every such file is checked before anything is written.
	reason, err := logDirFilesRefusal(resolved, root)
	if err != nil {
		return "", refuse(EnvLogDir, raw, err.Error(), logDirAllowed)
	}
	if reason != "" {
		return "", refuse(EnvLogDir, raw, reason, logDirAllowed)
	}
	if err := privateLogDir(resolved); err != nil {
		return "", refuse(EnvLogDir, raw, err.Error(), logDirAllowed)
	}
	return resolved, nil
}

// userCacheDir is the cache root for the default trace directory.
// Production leaves it as os.UserCacheDir. A test replaces it with a
// temporary directory. os.UserCacheDir reads $XDG_CACHE_HOME on Unix and
// $HOME/Library/Caches on Darwin, so a test that sets only the Unix
// variable does not name the directory Darwin uses, and it must not
// create a directory in the real user cache.
var userCacheDir = os.UserCacheDir

// defaultLogDir is the trace directory when PROVEN_LOG_DIR is unset.
// It sits in the user cache, outside any repository, so a scan started in
// the checkout does not write its trace into that checkout.
func defaultLogDir() (string, error) {
	cache, err := userCacheDir()
	if err != nil {
		return "", refuse(EnvLogDir, "", "the user cache directory is not available: "+err.Error(), logDirAllowed)
	}
	return filepath.Join(cache, "proven", "logs"), nil
}

// resolvePath returns path with symlinks resolved. A missing final component
// is kept: the directory may be created after the check. An ancestor that
// is a symlink is resolved, so a link into the review target is visible.
func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	var rest []string
	cur := abs
	for {
		resolved, err := filepath.EvalSymlinks(cur)
		if err == nil {
			for i := len(rest) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, rest[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", err
		}
		rest = append(rest, filepath.Base(cur))
		cur = parent
	}
}

// traceFile is the name trace.Open appends to. A symlink at this name is
// followed by that open.
const traceFile = "trace.jsonl"

// logDirFileNames are the files Proven opens directly in the log directory,
// in the order the pipeline opens them. trace.Open writes trace.jsonl.
// hmem.Store.audit writes hmem.jsonl. world.Store.audit writes world.jsonl.
// Each open follows a final symlink. Measured 2026-10-25 by reading every
// production OpenFile and WriteFile that joins a log directory: these three
// names, and no others. State files are written under the state directory.
var logDirFileNames = []string{traceFile, "hmem.jsonl", "world.jsonl"}

// logDirFilesRefusal is the reason opening a file in the log directory must
// be refused, or empty when every open may proceed. The first file that
// must be refused ends the check.
func logDirFilesRefusal(logDir, root string) (string, error) {
	for _, name := range logDirFileNames {
		reason, err := logFileRefusal(logDir, root, name)
		if err != nil {
			return "", fmt.Errorf("%s cannot be resolved", name)
		}
		if reason != "" {
			return reason, nil
		}
	}
	return "", nil
}

// logFileRefusal is the reason opening name in logDir must be refused, or
// empty when the open may proceed. A missing file is allowed: the open
// would create it in logDir, which was already checked. A symlink is
// followed to the file the open would write. That landing is refused when
// it is inside the target. A landing, or a regular file, with more than one
// name is the same file as that other name, and the other name can sit
// inside the target. The kernel reports the count and not the other path,
// so any count above one is refused.
func logFileRefusal(logDir, root, name string) (string, error) {
	path := filepath.Join(logDir, name)
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	// Any symbolic link is refused, wherever it points. Following one to its
	// landing and judging that landing is how a link written as sub/../x, with
	// sub itself a link into the target, passed: the text was cleaned before
	// the kernel would resolve it. Proven writes these files only as files.
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Sprintf("%s is a symbolic link; Proven writes %s only as a regular file, so remove the link (review target %s)", logFileInside(name), name, root), nil
	}
	if !info.Mode().IsRegular() {
		return "", nil
	}
	return manyNames(path, name, root)
}

// logFileInside is the subject of a refusal sentence. The trace file keeps
// the wording the trace checks already match.
func logFileInside(name string) string {
	if name == traceFile {
		return "trace file"
	}
	return name
}

// manyNames refuses path when it has more than one name. Its caller has
// already read path with Lstat and passes only an existing regular file. The sentence
// names the count that was read, the file to remove or rename, and the
// review target the scan was about to write beside. It does not claim that
// the other name is inside that target, because the count does not say.
func manyNames(path, name, root string) (string, error) {
	n, err := linkCount(path)
	if err != nil {
		return "", err
	}
	if n > 1 {
		return fmt.Sprintf("%s has %d names; remove or rename %s (review target %s)", logFileInside(name), n, name, root), nil
	}
	return "", nil
}

// pathInside reports whether dir is root or a path below root.
// When root exists, it walks dir and compares each existing ancestor with
// root by os.SameFile. A string comparison treats a letter-case alias, or
// another name for the same directory, as a different path. A root that
// does not exist has no inode, so the comparison is lexical.
func pathInside(dir, root string) (bool, error) {
	rootInfo, err := os.Stat(root)
	if os.IsNotExist(err) {
		return lexicalInside(dir, root)
	}
	if err != nil {
		return false, err
	}
	cur, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	cur = filepath.Clean(cur)
	for {
		info, err := os.Stat(cur)
		if err == nil {
			if os.SameFile(info, rootInfo) {
				return true, nil
			}
		} else if !os.IsNotExist(err) {
			return false, err
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return false, nil
		}
		cur = parent
	}
}

func lexicalInside(dir, root string) (bool, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return true, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}

// privateLogDir creates dir, or narrows an existing directory, to 0700.
// The trace is readable only by its user. MkdirAll applies a mode only when
// it creates, and the process umask can widen that mode.
func privateLogDir(dir string) error {
	info, err := os.Lstat(dir)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("log directory is not a directory")
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.Chmod(dir, 0o700)
}
