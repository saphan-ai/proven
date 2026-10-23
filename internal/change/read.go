// Package change reads the change under review, before any agent sees it.
//
// It has two sources. A sample fixture is two plain trees, base/ and change/,
// read by Read. A real repository is two commits, read through git by Between,
// Blob and Messages, or written out as those same two trees by ExportCommit
// and ExportWorktree so the Planner and the Prover can read them. A tracked
// symlink is not written into those trees and is not followed. Read still
// refuses a symlink in a tree it was given. Select then narrows the changed
// files with the include and exclude patterns of the review request.
//
// The Planner reads the files this package returns and names the risks; the
// pipeline uses the file list to decide what the review covers; the review
// shell reads proven.yaml at the baseline with Blob and the commit messages
// with Messages, where a change names its model (ADR-0011).
package change

import (
	"fmt"
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
// A symlink in either tree is refused, and the error names it.
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
	// Sorted, so the planner prompt built from these files is the same on every
	// run; a fixture pins that prompt by its hash.
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

// walk reads every regular file under root into dst, keyed by its slash path
// relative to root.
func walk(root string, dst map[string][]byte) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Walk uses Lstat, so a link is visible here. ReadFile would follow it
		// and a fixture could name a file outside this tree.
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("change: symlink %s", path)
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
