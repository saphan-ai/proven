package hmem

import (
	"fmt"
	"io"
	"os"
)

// Append stages a fault hook is asked about, in order. The hook runs AFTER the
// step it names, so a refusal models a step that placed its bytes and then
// reported a failure, which is the case a hook before the write cannot cover.
const (
	// stagePartial follows a write of the first half of the bytes (a short write).
	stagePartial = "partial"
	// stageWritten follows the write of all the bytes.
	stageWritten = "written"
	// stageSynced follows the sync of the file.
	stageSynced = "synced"
	// stageClosed follows the close of the file.
	stageClosed = "closed"
	// stageRollback precedes the truncation that takes a failed append back out.
	stageRollback = "rollback"
)

// appendFault is the seam tests use to fail an append after it placed bytes.
// path names the log; production leaves the hook nil.
type appendFault func(path, stage string) error

// appendSynced appends b to the log at path, so that the file ends the call
// either with all of b and a synced tail, or exactly as it was.
//
// The rule: a failed append never leaves a line in the live log. A write can
// place bytes before it reports an error, and a sync or a close can fail after
// every byte is in place; the caller then undoes the change and the ledger must
// not keep a line for a change that did not happen, nor a partial line that
// would join the next append. So the size before the append is recorded, and any
// failure truncates the file back to it and syncs that. (The other way, writing
// a temp copy and renaming it, would rewrite the whole log on every append and
// so grow with the log; truncating costs the bytes of the one append.)
//
// A torn tail left by a crash (bytes after the last newline, never
// acknowledged) is cut off before the append, so a new line never joins it.
// Readers already ignore such a tail (see logCursor.advance).
//
// If the truncation itself fails, the error says so by name and gives the
// offset, because then the caller cannot trust the tail of that file.
func appendSynced(path string, b []byte, fault appendFault) (err error) {
	hook := func(stage string) error {
		if fault == nil {
			return nil
		}
		return fault(path, stage)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			f.Close()
		}
	}()
	// OpenFile sets the mode only when it creates; an older file is narrowed.
	if err := f.Chmod(0o600); err != nil {
		return err
	}
	size, err := cutTornTail(f)
	if err != nil {
		return err
	}
	fail := func(cause error) error {
		if !closed {
			closed = true
			f.Close()
		}
		if rerr := rollbackTo(path, size, hook); rerr != nil {
			return fmt.Errorf("%w; taking the failed append back out failed too (%v), so the tail of %s after byte %d is not trustworthy", cause, rerr, path, size)
		}
		return cause
	}
	half := len(b) / 2
	if _, err := f.Write(b[:half]); err != nil {
		return fail(err)
	}
	if err := hook(stagePartial); err != nil {
		return fail(err)
	}
	if _, err := f.Write(b[half:]); err != nil {
		return fail(err)
	}
	if err := hook(stageWritten); err != nil {
		return fail(err)
	}
	if err := f.Sync(); err != nil {
		return fail(err)
	}
	if err := hook(stageSynced); err != nil {
		return fail(err)
	}
	closed = true
	if err := f.Close(); err != nil {
		return fail(err)
	}
	if err := hook(stageClosed); err != nil {
		return fail(err)
	}
	return nil
}

// rollbackTo cuts the file at path back to size and syncs the cut.
func rollbackTo(path string, size int64, hook func(string) error) error {
	if err := hook(stageRollback); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := f.Truncate(size); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// cutTornTail truncates f after its last newline when the file does not end
// with one, and returns the size the file has afterwards. The scan goes back
// from the end in chunks and stops after maxLogLine bytes, so a damaged file
// cannot make an append read without limit; a tail that long is refused by name.
func cutTornTail(f *os.File) (int64, error) {
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	size := info.Size()
	if size == 0 {
		return 0, nil
	}
	last := make([]byte, 1)
	if _, err := f.ReadAt(last, size-1); err != nil && err != io.EOF {
		return 0, err
	}
	if last[0] == '\n' {
		return size, nil
	}
	const chunk = 64 * 1024
	buf := make([]byte, chunk)
	end := size
	for scanned := int64(0); end > 0; {
		start := max(0, end-chunk)
		n := int(end - start)
		if _, err := f.ReadAt(buf[:n], start); err != nil && err != io.EOF {
			return 0, err
		}
		for i := n - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				return truncateTo(f, start+int64(i)+1)
			}
		}
		scanned += int64(n)
		if scanned > maxLogLine {
			return 0, fmt.Errorf("the log ends with an unterminated line longer than %d bytes", maxLogLine)
		}
		end = start
	}
	return truncateTo(f, 0)
}

// truncateTo cuts f at size and returns size.
func truncateTo(f *os.File, size int64) (int64, error) {
	if err := f.Truncate(size); err != nil {
		return 0, fmt.Errorf("cutting an unterminated line: %w", err)
	}
	return size, nil
}
