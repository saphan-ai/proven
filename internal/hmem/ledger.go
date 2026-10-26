package hmem

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
