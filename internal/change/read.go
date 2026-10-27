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
	"bytes"
	"context"
	"fmt"
	"io"
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
func (p Policy) Read(root string) ([]File, error) {
	base := map[string][]byte{}
	chg := map[string][]byte{}
	if err := p.walk(filepath.Join(root, "base"), base); err != nil {
		return nil, err
	}
	if err := p.walk(filepath.Join(root, "change"), chg); err != nil {
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

// Changed keeps the files whose two sides differ (an edited file, an added one and a deleted one) and the files
// git names as changed (paths), which covers a change of mode only, whose two sides read the same bytes. Every
// other file is dropped, so a reader of a whole exported repository gets the change and never the repository,
// however few bytes differ. A side that is missing is nil and a side that is an empty file is not, so adding or
// deleting an empty file is kept. The result may be empty: a change no byte and no path of git names.
func Changed(files []File, paths []string) []File {
	named := make(map[string]bool, len(paths))
	for _, p := range paths {
		named[p] = true
	}
	out := make([]File, 0, len(files))
	for _, f := range files {
		if named[f.Path] || (f.Base == nil) != (f.Change == nil) || !bytes.Equal(f.Base, f.Change) {
			out = append(out, f)
		}
	}
	return out
}

type changedOnlyKey struct{}

// changedOnly is a live review's mark: the paths git names as changed, nil when the trees were given prepared.
type changedOnly struct{ paths []string }

// WithChangedOnly marks a live review: its readers hand the Planner only what changed (Changed), with paths the
// files git names between the baseline and the head. A sample review carries no mark, because its recorded
// planner answers are keyed by the prompt over every file of its fixture.
func WithChangedOnly(ctx context.Context, paths []string) context.Context {
	return context.WithValue(ctx, changedOnlyKey{}, changedOnly{paths: paths})
}

// ChangedFor is Changed(files, paths) in a context WithChangedOnly marked, and files as they are in any other.
func ChangedFor(ctx context.Context, files []File) []File {
	if m, ok := ctx.Value(changedOnlyKey{}).(changedOnly); ok {
		return Changed(files, m.paths)
	}
	return files
}

// walk reads every regular file under root into dst, keyed by its slash path
// relative to root.
func (p Policy) walk(root string, dst map[string][]byte) error {
	var total int64
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
		// The trees are repository content: a file larger than the bound is
		// refused by name before it is read, and so are trees that together
		// exceed the archive bound. The read itself stops at the bound, so a
		// file that grows after Stat is not read whole.
		b, err := ReadBounded(path, p.Bounds().File, "the file "+filepath.ToSlash(rel), EnvMaxFileBytes)
		if err != nil {
			return err
		}
		total += int64(len(b))
		if total > p.Bounds().Archive {
			return OverBound("the files under "+filepath.Base(root), p.Bounds().Archive, EnvMaxArchiveBytes)
		}
		dst[filepath.ToSlash(rel)] = b
		return nil
	})
}

// ReadBounded reads the file at path and refuses one larger than limit bytes.
// what names the file in the refusal and env names the setting that sets the
// bound; the refusal never carries the content. The size is checked before the
// read, and the read itself stops one byte past the bound, so a file that
// grows after the check is still never held whole. ReadBounded follows a
// symlink as os.Open does: a caller that must not follow one checks first.
func ReadBounded(path string, limit int64, what, env string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if info, err := f.Stat(); err == nil && info.Size() > limit {
		return nil, OverBound(what, limit, env)
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, OverBound(what, limit, env)
	}
	return b, nil
}
