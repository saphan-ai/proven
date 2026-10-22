// Package change reads a fixture's base and change trees.
package change

import (
	"os"
	"path/filepath"
	"sort"
)

// File is one path in the change, with both sides. A missing side is empty.
type File struct {
	Path   string
	Base   []byte
	Change []byte
}

// Read walks base/ and change/ under root. Paths are slash-separated and sorted.
func Read(root string) ([]File, error) {
	base := map[string][]byte{}
	chg := map[string][]byte{}
	if err := walk(filepath.Join(root, "base"), base); err != nil {
		return nil, err
	}
	if err := walk(filepath.Join(root, "change"), chg); err != nil {
		return nil, err
	}
	names := map[string]struct{}{}
	for n := range base {
		names[n] = struct{}{}
	}
	for n := range chg {
		names[n] = struct{}{}
	}
	order := make([]string, 0, len(names))
	for n := range names {
		order = append(order, n)
	}
	sort.Strings(order)
	out := make([]File, 0, len(order))
	for _, n := range order {
		out = append(out, File{Path: n, Base: base[n], Change: chg[n]})
	}
	return out, nil
}

func walk(root string, dst map[string][]byte) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dst[filepath.ToSlash(rel)] = b
		return nil
	})
}
