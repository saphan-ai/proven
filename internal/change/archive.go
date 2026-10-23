package change

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ExportCommit writes the tracked files of rev in target into dst.
// The git directory is not copied. dst is created if needed.
func ExportCommit(target, rev, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	// gitArchive lists the commit's tracked files and does not emit .git.
	out, err := gitArchive(target, rev)
	if err != nil {
		return err
	}
	return untar(out, dst)
}

// ExportWorktree writes the tracked files of target's working tree into dst.
// An uncommitted edit of a tracked file is included. An untracked file is not.
// A file deleted in the working tree is omitted. The git directory is not copied.
func ExportWorktree(target, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	// ls-files without --others is the index: tracked paths only. -z keeps
	// unusual names intact. An empty head means this tree, not HEAD, so the
	// bytes are read from the working tree rather than from a commit.
	names, err := gitZ([]string{"-C", target, "ls-files", "-z", "--end-of-options"})
	if err != nil {
		return err
	}
	for _, rel := range names {
		clean, err := safeRel(rel)
		if err != nil {
			return err
		}
		if clean == "" {
			continue
		}
		src := filepath.Join(target, filepath.FromSlash(clean))
		out := filepath.Join(dst, filepath.FromSlash(clean))
		if err := copyTracked(src, out); err != nil {
			return err
		}
	}
	return nil
}

// untar extracts a git archive into dst. A path that leaves dst, or a git
// directory, is refused: the Prover mounts this tree, and the archive must
// not name anything outside it.
func untar(data []byte, dst string) error {
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		clean, err := safeRel(hdr.Name)
		if err != nil {
			return err
		}
		if clean == "" {
			continue
		}
		target := filepath.Join(dst, filepath.FromSlash(clean))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			// git archive writes a pax global header under a regular-file
			// type. It is not a file of the commit, so it is not copied.
			if isPaxHeader(hdr.Name) {
				if err := discard(tr); err != nil {
					return err
				}
				continue
			}
			if err := writeTarFile(tr, target, os.FileMode(hdr.Mode).Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			// change.Read refuses a symlink. Refuse it here too, so the tree
			// the Prover mounts never contains a link.
			return fmt.Errorf("change: symlink %s", hdr.Name)
		case tar.TypeXHeader, tar.TypeXGlobalHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
			if err := discard(tr); err != nil {
				return err
			}
		default:
			if isPaxHeader(hdr.Name) {
				if err := discard(tr); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("change: unsupported file %s", hdr.Name)
		}
	}
}

// isPaxHeader reports git archive's pax bookkeeping entry. It is not a tracked file.
func isPaxHeader(name string) bool {
	base := path.Base(path.Clean(name))
	return base == "pax_global_header" || base == "PaxHeaders" || strings.Contains(name, "PaxHeaders")
}

// discard reads a tar entry the tree does not keep, so the next header stays aligned.
func discard(tr *tar.Reader) error {
	_, err := io.Copy(io.Discard, tr)
	return err
}

// safeRel cleans a slash path from git. An empty result means the root entry,
// which the caller skips. A path that climbs out, or a git directory, is refused.
func safeRel(name string) (string, error) {
	clean := path.Clean(strings.TrimSuffix(name, "/"))
	if clean == "." || clean == "" {
		return "", nil
	}
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", fmt.Errorf("change: path %q leaves the tree", name)
	}
	for _, part := range strings.Split(clean, "/") {
		if part == ".git" {
			return "", fmt.Errorf("change: path %q is a git directory", name)
		}
	}
	return clean, nil
}

// writeTarFile writes one regular file from the archive. A zero mode becomes
// the usual file mode, because a header can omit it.
func writeTarFile(tr *tar.Reader, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if mode == 0 {
		mode = 0o644
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, tr)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// copyTracked copies one tracked working-tree file. A path that is gone was
// deleted in the working tree, so the change side has no file. A directory is
// a submodule, which is another repository, and is not copied. A symlink is
// refused, matching Read.
func copyTracked(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("change: symlink %s", src)
	}
	if info.IsDir() {
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("change: unsupported file %s", src)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
