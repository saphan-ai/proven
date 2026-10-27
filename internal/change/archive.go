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
// A tracked symlink is not written into dst and is not followed.
// skipped names each such path, slash-separated, relative to the commit.
//
// rev is checked as a commit before git sees it, the way Between checks its
// revisions: a value that is not a commit, including one that begins with a
// dash, is refused by name and never reaches the archive command as an
// argument. Each file and the whole archive are bounded (ADR-0013).
func (p Policy) ExportCommit(target, rev, dst string) ([]string, error) {
	if err := p.ensureCommit(target, rev, "exported"); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return nil, err
	}
	// In local mode no filter driver may run, and git archive runs the one the
	// attributes name; the blobs are read as stored instead (ADR-0034).
	if !p.public {
		return p.exportRaw(target, rev, dst)
	}
	// gitArchive lists the commit's tracked files and does not emit .git.
	out, err := p.gitArchive(target, rev)
	if err != nil {
		return nil, err
	}
	return p.untar(out, dst)
}

// ExportWorktree writes the tracked files of target's working tree into dst.
// An uncommitted edit of a tracked file is included. An untracked file is not.
// A file deleted in the working tree is omitted. The git directory is not copied.
// A tracked symlink is not written and is not followed. skipped names each
// such path, slash-separated, relative to the repository. Each file is bounded
// by the file limit and the files together by the archive limit, as in
// ExportCommit, because the working tree is as untrusted as a commit.
func (p Policy) ExportWorktree(target, dst string) ([]string, error) {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return nil, err
	}
	// ls-files without --others is the index: tracked paths only. -z keeps
	// unusual names intact. An empty head means this tree, not HEAD, so the
	// bytes are read from the working tree rather than from a commit.
	names, err := p.gitZ([]string{"-C", target, "ls-files", "-z", "--end-of-options"})
	if err != nil {
		return nil, err
	}
	var skipped []string
	var total int64
	for _, rel := range names {
		clean, err := safeRel(rel)
		if err != nil {
			return nil, err
		}
		if clean == "" {
			continue
		}
		src := filepath.Join(target, filepath.FromSlash(clean))
		out := filepath.Join(dst, filepath.FromSlash(clean))
		// ADR-0025: a tracked path whose parent directory is a link in the working tree
		// would be read from wherever the link points, so it is left out like a tracked link.
		linked, err := behindALink(target, clean)
		if err != nil {
			return nil, err
		}
		if linked {
			skipped = append(skipped, clean)
			continue
		}
		skip, n, err := p.copyTracked(src, out, clean)
		if err != nil {
			return nil, err
		}
		total += n
		if total > p.Bounds().Archive {
			return nil, OverBound("the tracked files of the working tree", p.Bounds().Archive, EnvMaxArchiveBytes)
		}
		if skip {
			skipped = append(skipped, clean)
		}
	}
	return skipped, nil
}

// untar extracts a git archive into dst. A path that leaves dst, or a git
// directory, is refused: the Prover mounts this tree, and the archive must
// not name anything outside it. skipped names each symlink that was left out.
func (p Policy) untar(data []byte, dst string) ([]string, error) {
	tr := tar.NewReader(bytes.NewReader(data))
	var skipped []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return skipped, nil
		}
		if err != nil {
			return nil, err
		}
		clean, err := safeRel(hdr.Name)
		if err != nil {
			return nil, err
		}
		if clean == "" {
			continue
		}
		target := filepath.Join(dst, filepath.FromSlash(clean))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, err
			}
		case tar.TypeReg, tar.TypeRegA:
			// git archive writes a pax global header under a regular-file
			// type. It is not a file of the commit, so it is not copied.
			if isPaxHeader(hdr.Name) {
				if err := discard(tr); err != nil {
					return nil, err
				}
				continue
			}
			// The header says how large the file is before any byte is written.
			if hdr.Size > p.Bounds().File {
				return nil, OverBound("the file "+clean, p.Bounds().File, EnvMaxFileBytes)
			}
			if err := writeTarFile(tr, target, os.FileMode(hdr.Mode).Perm()); err != nil {
				return nil, err
			}
		case tar.TypeSymlink:
			// F31: the Prover mounts this tree. A symlink can point outside it,
			// so the link is not written and not followed. The body is discarded
			// so the next header stays aligned. The path is returned for the trace.
			if err := discard(tr); err != nil {
				return nil, err
			}
			skipped = append(skipped, clean)
		case tar.TypeXHeader, tar.TypeXGlobalHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
			if err := discard(tr); err != nil {
				return nil, err
			}
		default:
			if isPaxHeader(hdr.Name) {
				if err := discard(tr); err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("change: unsupported file %s", hdr.Name)
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

// treeFileMode is the mode a materialized file gets: 0644, or 0755 when the
// source is executable by its owner. The group and other write bits are never
// carried over from the source or from the umask, so a tree written under a
// permissive umask is not writable by group or others. A header with no mode
// gets the usual file mode.
func treeFileMode(src os.FileMode) os.FileMode {
	if src&0o100 != 0 {
		return 0o755
	}
	return 0o644
}

// writeTarFile writes one regular file from the archive, at the mode treeFileMode gives it.
func writeTarFile(tr io.Reader, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, treeFileMode(mode))
	if err != nil {
		return err
	}
	// OpenFile applies the umask and keeps the mode of a file that exists; the
	// mode is set outright so it never follows either.
	if err := f.Chmod(treeFileMode(mode)); err != nil {
		f.Close()
		return err
	}
	_, copyErr := io.Copy(f, tr)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// behindALink reports whether a directory on the way from root to the slash path rel is a symbolic
// link. Each directory is checked with Lstat, which does not follow it. A directory that is gone is
// not a link: copyTracked treats the missing file as a deletion.
func behindALink(root, rel string) (bool, error) {
	parts := strings.Split(rel, "/")
	cur := root
	for _, p := range parts[:len(parts)-1] {
		cur = filepath.Join(cur, p)
		info, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}
	return false, nil
}

// copyTracked copies one tracked working-tree file and returns the bytes it
// copied. A path that is gone was deleted in the working tree, so the change
// side has no file. A directory is a submodule, which is another repository,
// and is not copied. skip is true when src is a symlink: F31 leaves it out, and
// the caller names the path. name is the path the refusal of an oversized file
// names; the file is read no further than the file limit, even if it grows
// while it is copied.
func (p Policy) copyTracked(src, dst, name string) (skip bool, copied int64, err error) {
	info, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		// Lstat has not followed the link. Return before Open, which would.
		return true, 0, nil
	}
	if info.IsDir() {
		return false, 0, nil
	}
	if !info.Mode().IsRegular() {
		return false, 0, fmt.Errorf("change: unsupported file %s", src)
	}
	limit := p.Bounds().File
	if info.Size() > limit {
		return false, 0, OverBound("the file "+name, limit, EnvMaxFileBytes)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, 0, err
	}
	in, err := os.Open(src)
	if err != nil {
		return false, 0, err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, treeFileMode(info.Mode().Perm()))
	if err != nil {
		return false, 0, err
	}
	if err := out.Chmod(treeFileMode(info.Mode().Perm())); err != nil {
		out.Close()
		return false, 0, err
	}
	// One byte past the limit shows a file that grew after Lstat.
	n, copyErr := io.Copy(out, io.LimitReader(in, limit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return false, 0, copyErr
	}
	if n > limit {
		return false, 0, OverBound("the file "+name, limit, EnvMaxFileBytes)
	}
	return false, n, closeErr
}
