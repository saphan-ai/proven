// Package snapshot copies a directory tree into a read-only snapshot.
package snapshot

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// readBits is the read permission for owner, group and others.
const readBits fs.FileMode = 0o444

// Copy copies the tree at src to dst. The copy adds the read bits for
// everyone and leaves the rest of each mode as it was.
func Copy(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			if err := os.MkdirAll(target, 0o777); err != nil {
				return err
			}
		} else if err := copyFile(path, target); err != nil {
			return err
		}
		return addReadBits(target)
	})
}

func addReadBits(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.Chmod(path, info.Mode().Perm()|readBits)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
