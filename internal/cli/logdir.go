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
	if err := privateLogDir(resolved); err != nil {
		return "", refuse(EnvLogDir, raw, err.Error(), logDirAllowed)
	}
	return resolved, nil
}

// defaultLogDir is the trace directory when PROVEN_LOG_DIR is unset.
// It sits in the user cache, outside any repository, so a scan started in
// the checkout does not write its trace into that checkout.
func defaultLogDir() (string, error) {
	cache, err := os.UserCacheDir()
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

// pathInside reports whether dir is root or a path below root.
// Both arguments are already absolute and symlink-resolved.
func pathInside(dir, root string) (bool, error) {
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
