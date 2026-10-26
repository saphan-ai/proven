package change

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// exportRaw writes the tracked files of rev in target into dst, each as the blob
// git stores. It is ExportCommit for local mode (ADR-0034). git archive converts
// each file on the way out: it runs the filter driver that .gitattributes names
// and the user's git configuration defines, and a driver is a program Proven did
// not choose (git-lfs is one, and its smudge connects over HTTPS). Reading the
// tree with ls-tree and the blobs with cat-file, which convert nothing, starts
// no driver, whatever the repository or the configuration say. The same files
// are written as untar writes them: a link is left out and named, a submodule is
// an empty directory, and each file and the whole tree are bounded.
func exportRaw(target, rev, dst string) ([]string, error) {
	list := GitCommand("-C", target, "ls-tree", "-r", "-z", "--end-of-options", rev)
	listing, err := outputBounded(list, Bounds().Archive, "the list of the commit's files", EnvMaxArchiveBytes)
	if err != nil {
		return nil, err
	}
	var skipped []string
	var blobs []treeBlob
	for _, entry := range bytes.Split(listing, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		blob, kind, err := parseTreeEntry(string(entry))
		if err != nil {
			return nil, err
		}
		switch kind {
		case "symlink":
			skipped = append(skipped, blob.rel)
		case "submodule":
			if err := os.MkdirAll(filepath.Join(dst, filepath.FromSlash(blob.rel)), 0o755); err != nil {
				return nil, err
			}
		default:
			blobs = append(blobs, blob)
		}
	}
	if err := writeBlobs(target, dst, blobs); err != nil {
		return nil, err
	}
	return skipped, nil
}

// treeBlob is one file of a commit: its id, its path relative to the commit and its mode.
type treeBlob struct {
	oid  string
	rel  string
	mode os.FileMode
}

// parseTreeEntry reads one entry of ls-tree -z: "<mode> <type> <id>\t<path>". kind is
// "symlink", "submodule" or "file". A path that leaves the tree, or is a git
// directory, is refused as it is in an archive.
func parseTreeEntry(entry string) (treeBlob, string, error) {
	head, name, ok := strings.Cut(entry, "\t")
	fields := strings.Fields(head)
	if !ok || len(fields) != 3 {
		return treeBlob{}, "", fmt.Errorf("change: unreadable tree entry %q", clipText(entry))
	}
	rel, err := safeRel(name)
	if err != nil {
		return treeBlob{}, "", err
	}
	if rel == "" {
		return treeBlob{}, "", fmt.Errorf("change: tree entry %q names no file", clipText(entry))
	}
	blob := treeBlob{oid: fields[2], rel: rel, mode: 0o644}
	switch {
	case fields[1] == "commit":
		return blob, "submodule", nil
	case fields[0] == "120000":
		return blob, "symlink", nil
	case fields[0] == "100755":
		blob.mode = 0o755
	}
	return blob, "file", nil
}

// clipText keeps an error message short: a path in a repository can be as long as its author likes.
func clipText(s string) string {
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

// writeBlobs reads the blobs from one cat-file --batch process, in the order given, and writes each
// under dst. The size of a blob is read before its bytes, so a file over the bound is refused without
// being read, and the sizes together are held to the archive bound.
func writeBlobs(target, dst string, blobs []treeBlob) error {
	if len(blobs) == 0 {
		return nil
	}
	var ids strings.Builder
	for _, b := range blobs {
		ids.WriteString(b.oid + "\n")
	}
	cmd := GitCommand("-C", target, "cat-file", "--batch")
	cmd.Stdin = strings.NewReader(ids.String())
	stderr := clippedBuffer{max: maxFailureBytes}
	cmd.Stderr = &stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("git: %w", err)
	}
	fail := func(err error) error {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	out := bufio.NewReader(pipe)
	var total int64
	for _, b := range blobs {
		size, missing, err := readBatchHeader(out)
		if missing {
			_ = cmd.Wait()
			return &gitFailure{msg: offlineMessage(cmd, "object "+b.oid+" is missing")}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// git stopped before it answered: its own words say why (a refused fetch of a missing object).
			_ = cmd.Wait()
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = "cat-file ended early"
			}
			return &gitFailure{msg: offlineMessage(cmd, msg)}
		}
		if err != nil {
			return fail(err)
		}
		if size > Bounds().File {
			return fail(OverBound("the file "+b.rel, Bounds().File, EnvMaxFileBytes))
		}
		total += size
		if total > Bounds().Archive {
			return fail(OverBound("the archive of the commit", Bounds().Archive, EnvMaxArchiveBytes))
		}
		path := filepath.Join(dst, filepath.FromSlash(b.rel))
		if err := writeTarFile(io.LimitReader(out, size), path, b.mode); err != nil {
			return fail(err)
		}
		if _, err := out.Discard(1); err != nil { // the newline after the content
			return fail(err)
		}
	}
	if err := cmd.Wait(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return &gitFailure{msg: offlineMessage(cmd, msg)}
	}
	return nil
}

// readBatchHeader reads "<id> blob <size>" and returns the size. "<id> missing" is an object the
// repository does not hold, which in a partial clone is the one git would fetch.
func readBatchHeader(out *bufio.Reader) (size int64, missing bool, err error) {
	line, err := out.ReadString('\n')
	if err != nil {
		return 0, false, err
	}
	fields := strings.Fields(line)
	if len(fields) == 2 && fields[1] == "missing" {
		return 0, true, nil
	}
	if len(fields) != 3 || fields[1] != "blob" {
		return 0, false, fmt.Errorf("git: unreadable cat-file answer %q", clipText(line))
	}
	size, err = strconv.ParseInt(fields[2], 10, 64)
	if err != nil || size < 0 {
		return 0, false, fmt.Errorf("git: unreadable cat-file size %q", clipText(line))
	}
	return size, false, nil
}
