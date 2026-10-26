package hmem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// maxLogLine bounds one line of the audit log or the dream log. A line longer
// than this is refused by name, so a damaged file cannot make a reader grow
// without limit.
const maxLogLine = 8 << 20

// historyPageMax is the most ledger entries one History call returns. A longer
// history is read page by page with HistoryPage, so one call never loads an
// unbounded trail.
const historyPageMax = 1000

// indexedActions are the ledger actions ledgerHas can ask about. Only these
// are kept in the in-memory set, so the index grows with the lines that
// recovery and approval read, not with every read audit line.
var indexedActions = map[string]bool{"create": true, "update": true, "dream_approved": true}

// logCursor reads an append-only JSONL file incrementally. It remembers how far
// it has read and each call reads only the complete lines added since, so a
// lookup costs the new bytes, not the whole file (E16: every read is bounded at
// the source). The owner keeps the index it builds from those lines; the file
// stays the source of truth, and a file that shrank is read again from the
// start.
type logCursor struct {
	path string
	// label names the log in an error ("ledger" or "dream log").
	label string
	// off is the byte offset just past the last complete line consumed.
	off int64
	// read counts every byte this cursor ever read from the file. The growth
	// tests assert it stays near the file size however many lookups run.
	read int64
}

// advance calls fn for each complete, non-empty line after the cursor, with the
// offset the line starts at. A trailing line with no newline is a write in
// progress or torn by a crash: it is not consumed. A missing file is an empty
// log. When the file is shorter than the cursor, reset runs first and the file
// is read from the start. If fn fails, the cursor stays before that line, so the
// same error is met again rather than the line being skipped.
func (c *logCursor) advance(reset func(), fn func(line []byte, start int64) error) error {
	// An error from fn already names itself; a file error is named here.
	wrap := func(err error) error { return fmt.Errorf("hmem: %s: %w", c.label, err) }
	f, err := os.Open(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			if c.off != 0 {
				reset()
				c.off = 0
			}
			return nil
		}
		return wrap(err)
	}
	// The file is only read here; a failed Close of a read-only file loses no data.
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return wrap(err)
	}
	if info.Size() < c.off {
		reset()
		c.off = 0
	}
	if info.Size() == c.off {
		return nil
	}
	if _, err := f.Seek(c.off, io.SeekStart); err != nil {
		return wrap(err)
	}
	r := bufio.NewReaderSize(f, 64*1024)
	for {
		line, err := readLine(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return wrap(err)
		}
		c.read += int64(len(line))
		if body := bytes.TrimSpace(line); len(body) > 0 {
			if err := fn(body, c.off); err != nil {
				return err
			}
		}
		c.off += int64(len(line))
	}
}

// readLine returns the next line including its newline. It returns io.EOF when
// what remains has no newline (nothing complete to consume) and an error when
// a line passes maxLogLine.
func readLine(r *bufio.Reader) ([]byte, error) {
	var line []byte
	for {
		part, err := r.ReadSlice('\n')
		line = append(line, part...)
		switch err {
		case nil:
			return line, nil
		case bufio.ErrBufferFull:
			if len(line) > maxLogLine {
				return nil, fmt.Errorf("a log line is longer than %d bytes", maxLogLine)
			}
		case io.EOF:
			return nil, io.EOF
		default:
			return nil, err
		}
	}
}

// ledgerIndex is what the audit log is read into: where each memory's lines
// start, and the set of lines recovery and approval ask about. It is kept up to
// date by the cursor, so it is built once and then extended by new lines only.
type ledgerIndex struct {
	cur logCursor
	// byID holds the start offset of each line, per memory id, oldest first.
	byID map[string][]int64
	// has holds action, id and proposal for the lines of indexedActions. The
	// proposal-less key is always there too.
	has map[string]bool
	// approved holds the ids of the proposals with a dream_approved line.
	approved map[string]bool
}

func newLedgerIndex(path string) *ledgerIndex {
	ix := &ledgerIndex{cur: logCursor{path: path, label: "ledger"}}
	ix.reset()
	return ix
}

func (ix *ledgerIndex) reset() {
	ix.byID = map[string][]int64{}
	ix.has = map[string]bool{}
	ix.approved = map[string]bool{}
}

func hasKey(action, id, proposal string) string { return action + "\x00" + id + "\x00" + proposal }

// syncLedger brings the index up to the end of the audit log. A line that is
// not valid JSON is refused, as it always was. The caller holds s.mu.
func (s *Store) syncLedger() error {
	ix := s.ledger
	return ix.cur.advance(ix.reset, func(line []byte, start int64) error {
		var e LedgerEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return fmt.Errorf("hmem: ledger line is not valid JSON: %w", err)
		}
		ix.byID[e.MemoryID] = append(ix.byID[e.MemoryID], start)
		if indexedActions[e.Action] {
			ix.has[hasKey(e.Action, e.MemoryID, "")] = true
			if p, ok := e.Details["proposal"].(string); ok {
				ix.has[hasKey(e.Action, e.MemoryID, p)] = true
			}
		}
		if e.Action == "dream_approved" {
			ix.approved[e.MemoryID] = true
		}
		return nil
	})
}

// readLedgerAt reads the one ledger line that starts at off.
func (s *Store) readLedgerAt(f *os.File, off int64) (LedgerEntry, error) {
	r := bufio.NewReaderSize(io.NewSectionReader(f, off, maxLogLine+1), 64*1024)
	line, err := readLine(r)
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("hmem: ledger: %w", err)
	}
	s.ledger.cur.read += int64(len(line))
	var e LedgerEntry
	if err := json.Unmarshal(bytes.TrimSpace(line), &e); err != nil {
		return LedgerEntry{}, fmt.Errorf("hmem: ledger line is not valid JSON: %w", err)
	}
	return e, nil
}

// History returns the ledger entries of one memory, oldest first, at most
// historyPageMax of them. It reads the audit file, not memory, so a history
// survives a reopen and a deleted record keeps its trail. A ledger not yet
// written is an empty history. A trail longer than one page is read with
// HistoryPage.
func (s *Store) History(id string) ([]LedgerEntry, error) {
	out, _, err := s.HistoryPage(id, 0, historyPageMax)
	return out, err
}

// HistoryPage returns up to limit ledger entries of one memory, oldest first,
// starting at the entry numbered cursor (zero is the first). The second result
// is the cursor of the next page, or -1 when this page reached the end. The
// ledger is append-only, so entry numbers never move and a cursor stays valid.
// A limit above historyPageMax is cut to it; a limit below one is refused.
func (s *Store) HistoryPage(id string, cursor, limit int) ([]LedgerEntry, int, error) {
	if limit < 1 || cursor < 0 {
		return nil, -1, fmt.Errorf("hmem: history page needs a cursor of at least 0 and a limit of at least 1")
	}
	limit = min(limit, historyPageMax)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.syncLedger(); err != nil {
		return nil, -1, err
	}
	offs := s.ledger.byID[id]
	out := []LedgerEntry{}
	if cursor >= len(offs) {
		return out, -1, nil
	}
	end := min(cursor+limit, len(offs))
	f, err := os.Open(s.auditPath)
	if err != nil {
		return nil, -1, fmt.Errorf("hmem: ledger: %w", err)
	}
	// The file is only read here; a failed Close of a read-only file loses no data.
	defer f.Close()
	for _, off := range offs[cursor:end] {
		e, err := s.readLedgerAt(f, off)
		if err != nil {
			return nil, -1, err
		}
		out = append(out, e)
	}
	if end >= len(offs) {
		return out, -1, nil
	}
	return out, end, nil
}
