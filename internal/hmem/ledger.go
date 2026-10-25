package hmem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// LedgerState is the part of a record a ledger entry keeps before and after a
// change. It is a snapshot of the fields a change can move, not the whole record.
type LedgerState struct {
	Confidence   float64  `json:"confidence"`
	AccessCount  int      `json:"accessCount"`
	Layer        string   `json:"layer"`
	Domain       string   `json:"domain"`
	Tags         []string `json:"tags"`
	SupersededBy string   `json:"supersededBy,omitempty"`
}

// LedgerEntry is one line of the trust ledger. The ledger is the store's audit
// JSONL: append-only, one line per change (ADR-0018 point 3).
type LedgerEntry struct {
	MemoryID      string         `json:"memoryId"`
	Action        string         `json:"action"`
	Actor         string         `json:"actor"`
	Timestamp     int64          `json:"timestamp"`
	PreviousState *LedgerState   `json:"previousState,omitempty"`
	NewState      *LedgerState   `json:"newState,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
}

// ledgerState snapshots the movable fields of rec for a ledger line.
func ledgerState(rec Record) LedgerState {
	return LedgerState{
		Confidence:   rec.Confidence,
		AccessCount:  rec.AccessCount,
		Layer:        rec.Layer,
		Domain:       rec.Domain,
		Tags:         append([]string{}, rec.Tags...),
		SupersededBy: rec.SupersededBy,
	}
}

// History returns every ledger entry for one memory, oldest first. It reads the
// audit file, not memory, so a history survives a reopen and a deleted record
// keeps its trail. A ledger not yet written is an empty history.
func (s *Store) History(id string) ([]LedgerEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.auditPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []LedgerEntry{}, nil
		}
		return nil, fmt.Errorf("hmem: ledger: %w", err)
	}
	defer f.Close()
	out := []LedgerEntry{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var e LedgerEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, fmt.Errorf("hmem: ledger line is not valid JSON: %w", err)
		}
		if e.MemoryID == id {
			out = append(out, e)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("hmem: ledger: %w", err)
	}
	return out, nil
}
