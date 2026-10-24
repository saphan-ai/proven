// Package pathguard compares directories by what they resolve to, so a state,
// trace or work directory is never taken for outside a tree it reaches through
// a symbolic link.
package pathguard

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Resolve returns the absolute form of p with every symbolic link on its way
// resolved. A path that does not exist yet is resolved through its nearest
// existing parent and the missing names are appended, so a directory about to be
// created is placed where it would really land. A link that cannot be resolved
// (one that points nowhere, or loops) is an error, never a guess.
func Resolve(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	var rest []string
	cur := abs
	for {
		resolved, err := filepath.EvalSymlinks(cur)
		if err == nil {
			for i := len(rest) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, rest[i])
			}
			return resolved, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		// A name that exists but does not resolve is a link to nowhere.
		if _, lerr := os.Lstat(cur); lerr == nil {
			return "", fmt.Errorf("%s is a link that does not resolve", cur)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", err
		}
		rest = append(rest, filepath.Base(cur))
		cur = parent
	}
}

// Inside reports whether dir is root or lies below it, comparing the resolved
// paths of both. A name such as "..state" starts with ".." and is still inside;
// only ".." itself or a "../" prefix leaves root.
func Inside(dir, root string) (bool, error) {
	rd, err := Resolve(dir)
	if err != nil {
		return false, err
	}
	rr, err := Resolve(root)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(rr, rd)
	if err != nil {
		return false, err
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))), nil
}
