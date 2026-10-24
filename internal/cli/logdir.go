package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// EnvLogDir is the trace directory. An empty value means the user cache.
// Sample mode reads it: the sample reviewer writes a trace too.
const EnvLogDir = "PROVEN_LOG_DIR"

// Environment is the trace directory row on the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{{
		Name:        EnvLogDir,
		Default:     envdoc.Tick("<user cache>/proven/logs"),
		Description: "where the trace is written. The directory is created with mode 0700. A path that resolves inside the review target, including through a symlink, is refused. Sample mode reads this",
	}}
}

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
	// The directory can sit outside the target while trace.jsonl is a symlink
	// into it. Opening the trace follows that final link, so the file is
	// checked before anything is written.
	fileInside, err := traceFileInsideTarget(resolved, root)
	if err != nil {
		return "", refuse(EnvLogDir, raw, "trace file cannot be resolved", logDirAllowed)
	}
	if fileInside {
		return "", refuse(EnvLogDir, raw, "trace file resolves inside the review target "+root, logDirAllowed)
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

// traceFileInsideTarget reports whether trace.jsonl is a symlink whose
// destination resolves inside the review target. A missing file is outside:
// the open would create it in logDir, which was already checked.
func traceFileInsideTarget(logDir, root string) (bool, error) {
	path := filepath.Join(logDir, traceFile)
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}
	landing, err := symlinkLanding(path)
	if err != nil {
		return false, err
	}
	return pathInside(landing, root)
}

// symlinkLanding returns the path a symlink chain names, including a final
// component that does not exist yet. The open would create that component.
func symlinkLanding(path string) (string, error) {
	const maxLinks = 64
	cur := path
	seen := make(map[string]struct{})
	for i := 0; i < maxLinks; i++ {
		abs, err := filepath.Abs(cur)
		if err != nil {
			return "", err
		}
		abs = filepath.Clean(abs)
		if _, ok := seen[abs]; ok {
			return "", fmt.Errorf("trace file symlink loop")
		}
		seen[abs] = struct{}{}
		info, err := os.Lstat(abs)
		if os.IsNotExist(err) {
			return resolvePath(abs)
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return resolvePath(abs)
		}
		dest, err := os.Readlink(abs)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(dest) {
			dest = filepath.Join(filepath.Dir(abs), dest)
		}
		cur = dest
	}
	return "", fmt.Errorf("trace file symlink loop")
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
