// Package snapshot copies a directory tree into a read-only snapshot.
package snapshot

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Copy copies the tree at src to dst. Every file in the copy has mode 0644
// and every directory has mode 0755.
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
			return os.Chmod(target, 0o755)
		}
		if err := copyFile(path, target); err != nil {
			return err
		}
		return os.Chmod(target, 0o644)
	})
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
