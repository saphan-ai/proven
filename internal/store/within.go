package store

import (
	"os"
	"path/filepath"
	"strings"
)

// Within reports whether file resolves to a place inside dir, with links
// resolved on both sides. ADR-0047: the local file never lives in the tree
// under review, so a link from outside into the tree is caught as well as a
// plain path. The file need not exist yet: the deepest existing parent is
// resolved and the rest is appended.
func Within(file, dir string) (bool, error) {
	f, err := resolve(file)
	if err != nil {
		return false, err
	}
	d, err := resolve(dir)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(d, f)
	if err != nil {
		return false, nil
	}
	return rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}

// resolve makes path absolute and resolves every link on the part that
// exists.
func resolve(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rest := ""
	cur := abs
	for {
		real, err := filepath.EvalSymlinks(cur)
		if err == nil {
			return filepath.Join(real, rest), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs, nil
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}
