// Package hmem is the memory store and retrieval service, with sleep
// (sleep.go), conflict resolution (conflict.go), the trust ledger (ledger.go)
// and dreaming with approval (dream.go). None of them calls a model.
//
// It is the reviewer's memory across pull requests (ADR-0003 point 1), built
// to the Meterless H-MEM specification (ADR-0018 point 1). In a review it
// keeps the Skeptic's refutations: when the Skeptic refutes a finding that
// has evidence, the world package stores it here, and the next review of the
// same change matches it so the refuted finding does not come back
// (ADR-0006 point 4, ADR-0018 point 6). The store is a JSON file outside the
// repository, with an append-only audit log in the run's logs directory.
// A failed audit or persist leaves the in-memory records and the id sequence
// as they were, so memory and the state file agree.
package hmem

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/pathguard"
)

// schemaVersion is the state.json format. Open refuses any other version
// rather than guess at a file it does not know. Version 2 stores a refutation
// only with its evidence, so a version 1 file, whose records were not held to
// that rule, is refused rather than trusted.
const schemaVersion = 2

// Layers and types are the closed sets from the memory record.
const (
	LayerShort     = "short_term"
	LayerWork      = "working"
	LayerLong      = "long_term"
	TypeGeneral    = "general"
	TypeFactual    = "factual"
	TypePreference = "preference"
	TypePersonal   = "personal"
)

// Record is one memory. Source and provenance are mandatory. The JSON names
// follow the Meterless memory record, so the conformance suite reads them as is.
type Record struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	Type         string     `json:"type"`
	Layer        string     `json:"layer"`
	Timestamp    int64      `json:"timestamp"`
	LastAccessed int64      `json:"lastAccessed"`
	AccessCount  int        `json:"accessCount"`
	Tags         []string   `json:"tags"`
	Embedding    []float64  `json:"embedding"`
	Confidence   float64    `json:"confidence"`
	Source       string     `json:"source"`
	Provenance   Provenance `json:"provenance"`
	Domain       string     `json:"domain"`
	Namespace    string     `json:"namespace"`
	Entities     []string   `json:"entities"`
	RelatedTo    []string   `json:"relatedTo"`
	Supersedes   string     `json:"supersedes,omitempty"`
	SupersededBy string     `json:"supersededBy,omitempty"`
	DerivedFrom  []string   `json:"derivedFrom,omitempty"`
	ChatID       string     `json:"chatId,omitempty"`
	// ChangeKey and FindingKey are the exact index the next review matches.
	// Retrieval ranking does not read them.
	ChangeKey  string `json:"changeKey,omitempty"`
	FindingKey string `json:"findingKey,omitempty"`
}

// Provenance is where a record was learned.
type Provenance struct {
	Origin    string `json:"origin"`
	LearnedAt int64  `json:"learnedAt"`
	Label     string `json:"label"`
}

// AddInput is one new memory. Source is required. A nil Confidence means 0.7.
type AddInput struct {
	Content    string
	Type       string
	Layer      string
	Tags       []string
	Confidence *float64
	Source     string
	Origin     string
	Label      string
	ChatID     string
	// DerivedFrom names the records a synthesized or dreamed record came from.
	// Sleep never archives a record that carries lineage.
	DerivedFrom []string
	At          int64
	// fixedID, when set, is the id of the new record instead of the next
	// sequence id. An approved dream uses it so its record can be found again.
	fixedID string
	// deferWrites, when set, installs the record in memory only. The caller
	// writes the state and the ledger line together with its other changes, so
	// one failure undoes all of them (sleep synthesis uses it).
	deferWrites bool
}

// RefutationKey names a finding a skeptic refuted on one change. ChangeKey,
// RuleID, Path and Title identify it; Reason is the Skeptic's words and At
// the time to record, in Unix milliseconds (zero means now).
type RefutationKey struct {
	ChangeKey string
	RuleID    string
	Path      string
	Title     string
	Reason    string
	At        int64
}

// Options open a store. Dir is the durable directory. LogDir receives the audit JSONL.
// RepoRoot, when set, refuses a Dir inside that tree, so memory never lands
// in the repository under review. Clock, when set, replaces the wall clock.
type Options struct {
	Dir      string
	LogDir   string
	RepoRoot string
	Clock    func() time.Time
}

// Store is the durable memory. The generator is accepted and never called:
// this package does not ask a model for a memory.
type Store struct {
	dir       string
	auditPath string
	clock     func() time.Time
	gen       generator.Generator
	mu        sync.Mutex
	seq       int
	records   map[string]Record
	// epoch counts the Clears this store has had. A dream proposal carries the
	// epoch it was made in and is bound only while that epoch is current, so
	// Clear voids every old proposal with the one state write that empties the
	// records (see Clear).
	epoch int
	// ledger and dreams are the incremental indexes of the two append-only
	// logs; see logindex.go.
	ledger *ledgerIndex
	dreams *dreamIndex
	// writeHook, when set, runs before each durable write ("dream", "state" or "audit:<action>") and may refuse it.
	// It is the seam tests use to inject a failed write; production leaves it nil.
	writeHook func(op string) error
	// appendFault, when set, runs after each step of a log append and may refuse
	// it, so a test can fail an append after its bytes were placed. Production
	// leaves it nil.
	appendFault appendFault
	// readDir lists a directory; backups use it only to migrate a store that has
	// no sequence file yet. Nil means os.ReadDir; tests count its calls.
	readDir func(dir string) ([]os.DirEntry, error)
}

// disk is the shape of state.json. Seq is kept so Add never reuses an id after
// a reopen. Clear sets it back to zero and writes that, so the next id starts again.
type disk struct {
	Version int `json:"version"`
	Seq     int `json:"seq"`
	// Epoch is the Clear count; a file written before it existed reads as 0.
	Epoch   int      `json:"epoch,omitempty"`
	Records []Record `json:"records"`
}

// Open creates the directory, loads state.json when it exists, and refuses a bad file.
func Open(opts Options) (*Store, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("hmem: state directory is empty")
	}
	if opts.LogDir == "" {
		return nil, fmt.Errorf("hmem: logs directory is empty")
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("hmem: state directory: %w", err)
	}
	if err := refuseInside(dir, opts.RepoRoot); err != nil {
		return nil, err
	}
	if err := privateDir(dir); err != nil {
		return nil, fmt.Errorf("hmem: state directory: %w", err)
	}
	st, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("hmem: state directory: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("hmem: state path %s is not a directory", dir)
	}
	logDir, err := filepath.Abs(opts.LogDir)
	if err != nil {
		return nil, fmt.Errorf("hmem: logs directory: %w", err)
	}
	if err := privateDir(logDir); err != nil {
		return nil, fmt.Errorf("hmem: logs directory: %w", err)
	}
	s := &Store{
		dir:       dir,
		auditPath: filepath.Join(logDir, "hmem.jsonl"),
		clock:     opts.Clock,
		records:   map[string]Record{},
	}
	s.ledger = newLedgerIndex(s.auditPath)
	s.dreams = newDreamIndex(filepath.Join(logDir, "dreams.jsonl"))
	path := filepath.Join(dir, "state.json")
	// A state file an earlier run left at a wider mode is narrowed before it is read.
	if err := os.Chmod(path, 0o600); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("hmem: state file: %w", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("hmem: read state: %w", err)
	}
	var file disk
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, fmt.Errorf("hmem: state.json is not valid JSON: %w", err)
	}
	if file.Version != schemaVersion {
		return nil, fmt.Errorf("hmem: state version %d, want %d; remove %s to start a new memory", file.Version, schemaVersion, dir)
	}
	for _, rec := range file.Records {
		if err := validate(rec); err != nil {
			return nil, err
		}
		if _, ok := s.records[rec.ID]; ok {
			return nil, fmt.Errorf("hmem: duplicate record id %s", rec.ID)
		}
		s.records[rec.ID] = rec
	}
	s.seq = file.Seq
	s.epoch = file.Epoch
	// An approval a crashed run left half done is finished now, so no person
	// has to notice it.
	if err := s.recoverApprovals(); err != nil {
		return nil, err
	}
	return s, nil
}

// refuseInside refuses a state directory at or below root.
func refuseInside(dir, root string) error {
	if root == "" {
		return nil
	}
	// Both paths are resolved through their links first, so a state path that
	// reaches the repository through a link is refused; a link that does not
	// resolve is refused too.
	inside, err := pathguard.Inside(dir, root)
	if err != nil {
		return fmt.Errorf("hmem: state directory %s cannot be placed against the repository: %w", dir, err)
	}
	if inside {
		return fmt.Errorf("hmem: state directory %s is inside the repository", dir)
	}
	return nil
}

// validate refuses a record outside the closed sets or without its source
// and provenance. Open runs it on every loaded record, so a hand-edited file
// cannot bring in a record the store would never write.
func validate(rec Record) error {
	if rec.ID == "" || strings.TrimSpace(rec.Content) == "" {
		return fmt.Errorf("hmem: record %q is missing id or content", rec.ID)
	}
	if strings.TrimSpace(rec.Source) == "" {
		return fmt.Errorf("hmem: record %s has no source", rec.ID)
	}
	switch rec.Type {
	case TypeGeneral, TypeFactual, TypePersonal, TypePreference:
	default:
		return fmt.Errorf("hmem: record %s has type %q", rec.ID, rec.Type)
	}
	switch rec.Layer {
	case LayerShort, LayerWork, LayerLong:
	default:
		return fmt.Errorf("hmem: record %s has layer %q", rec.ID, rec.Layer)
	}
	if rec.Confidence < 0 || rec.Confidence > 1 {
		return fmt.Errorf("hmem: record %s has confidence %v", rec.ID, rec.Confidence)
	}
	if rec.Provenance.Origin == "" || rec.Provenance.Label == "" {
		return fmt.Errorf("hmem: record %s has no provenance", rec.ID)
	}
	return nil
}

// SetGenerator stores the model interface. RememberRefutation, Add and Query do not call it.
// ADR-0018: every model call goes through an injected generator, never inside the engine.
func (s *Store) SetGenerator(g generator.Generator) { s.gen = g }

// Generator is the model interface this store was given.
func (s *Store) Generator() generator.Generator { return s.gen }

// Add inserts one record. A blank source is refused.
func (s *Store) Add(in AddInput) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addLocked(in)
}

// addLocked is Add for a caller that already holds s.mu: sleep synthesis and
// an approved dream both create a record inside their own locked step.
func (s *Store) addLocked(in AddInput) (Record, error) {
	if strings.TrimSpace(in.Source) == "" {
		return Record{}, fmt.Errorf("hmem: a record with no source is refused")
	}
	now := in.At
	if now == 0 {
		now = s.now()
	}
	typ := in.Type
	if typ == "" {
		typ = TypeGeneral
	}
	layer := in.Layer
	if layer == "" {
		layer = LayerWork
	}
	conf := 0.7
	if in.Confidence != nil {
		conf = *in.Confidence
	}
	origin := in.Origin
	if origin == "" {
		origin = "manual"
	}
	label := in.Label
	if label == "" {
		label = in.Source
	}
	prevSeq := s.seq
	id := in.fixedID
	if id == "" {
		s.seq++
		id = fmt.Sprintf("mem-%d", s.seq)
	}
	rec := Record{
		ID:           id,
		Content:      in.Content,
		Type:         typ,
		Layer:        layer,
		Timestamp:    now,
		LastAccessed: now,
		Tags:         append([]string(nil), in.Tags...),
		Embedding:    Embed(in.Content),
		Confidence:   clamp01(conf),
		Source:       in.Source,
		Provenance:   Provenance{Origin: origin, LearnedAt: now, Label: label},
		Domain:       classifyDomain(in.Content, in.Tags, typ),
		Entities:     extractEntities(in.Content),
		RelatedTo:    []string{},
		ChatID:       in.ChatID,
		DerivedFrom:  append([]string(nil), in.DerivedFrom...),
	}
	if len(rec.DerivedFrom) == 0 {
		rec.DerivedFrom = nil
	}
	rec.Namespace = namespace(rec.Domain, rec.Type)
	// A refused record gives its sequence number back, so ids stay dense.
	if err := validate(rec); err != nil {
		s.seq = prevSeq
		return Record{}, err
	}
	s.records[rec.ID] = rec
	if in.deferWrites {
		return rec, nil
	}
	if in.fixedID != "" {
		// The caller writes the ledger line after this, so a failed state
		// write leaves neither a record nor a line that names one.
		if err := s.persist(); err != nil {
			delete(s.records, rec.ID)
			return Record{}, err
		}
		return rec, nil
	}
	if err := s.publish("create", rec.ID, in.Source, prevSeq, nil, []string{rec.ID}); err != nil {
		return Record{}, err
	}
	return rec, nil
}

// RememberRefutation stores a refuted finding. The same key returns the same record
// and does not append a second one. The world package calls it when it
// ingests a refutation that had evidence, so a replayed trace stores nothing twice.
func (s *Store) RememberRefutation(k RefutationKey) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if k.ChangeKey == "" || k.RuleID == "" || k.Path == "" {
		return Record{}, fmt.Errorf("hmem: a refutation needs a change, a rule and a path")
	}
	// The id is derived from the key, not from the sequence: that is what
	// makes a second call with the same key find the first record.
	fk := findingKey(k.RuleID, k.Path, k.Title)
	sum := sha256.Sum256([]byte(k.ChangeKey + "\n" + fk))
	id := "mem_" + hex.EncodeToString(sum[:])[:16]
	if rec, ok := s.records[id]; ok {
		return rec, nil
	}
	now := k.At
	if now == 0 {
		now = s.now()
	}
	// A refutation is a long-term fact with high confidence: the Skeptic
	// refuted it with evidence, not by opinion.
	content := fmt.Sprintf("refuted %s %s: %s", k.RuleID, k.Path, k.Reason)
	conf := 0.9
	rec := Record{
		ID:           id,
		Content:      content,
		Type:         TypeFactual,
		Layer:        LayerLong,
		Timestamp:    now,
		LastAccessed: now,
		Tags: []string{
			"change:" + k.ChangeKey,
			"rule:" + k.RuleID,
			"path:" + k.Path,
		},
		Embedding:  Embed(content),
		Confidence: conf,
		Source:     "skeptic",
		Provenance: Provenance{Origin: "model_response", LearnedAt: now, Label: "skeptic refutation"},
		Domain:     classifyDomain(content, nil, TypeFactual),
		Entities:   extractEntities(content),
		RelatedTo:  []string{},
		ChangeKey:  k.ChangeKey,
		FindingKey: fk,
	}
	rec.Namespace = namespace(rec.Domain, rec.Type)
	if err := validate(rec); err != nil {
		return Record{}, err
	}
	s.records[id] = rec
	if err := s.publish("create", id, "skeptic", s.seq, nil, []string{id}); err != nil {
		return Record{}, err
	}
	return rec, nil
}

// MatchRefutation returns the stored refutation of this finding on this change.
// It matches the exact keys, not the ranked retrieval: a refutation must come
// back for the same finding only, never for a finding that merely reads alike.
func (s *Store) MatchRefutation(changeKey, ruleID, path, title string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	want := findingKey(ruleID, path, title)
	var found Record
	ok := false
	for _, rec := range s.records {
		if rec.ChangeKey != changeKey || rec.FindingKey != want {
			continue
		}
		// The lowest id wins, so the answer does not depend on map order.
		if !ok || rec.ID < found.ID {
			found = rec
			ok = true
		}
	}
	return found, ok
}

// Clear drops every record, resets the id sequence, writes the empty store and
// then one clear audit line. Earlier audit lines stay.
// A failed write leaves the records and the sequence as they were.
//
// The rule: a proposal never outlives the records it names. The reset sequence
// hands out the same ids again, and a proposal that survived would bind its old
// content to unrelated new records (a false lineage). Clear therefore raises
// the store's epoch in the same state.json write that empties the records, and
// a proposal made in an older epoch reads as void with the reason "cleared"
// unless it finished (approved with its dream_approved line, or rejected). A
// void proposal is refused by Approve by name.
//
// Clear is one transition (ADR-0018 point 3: no half-applied change), and its
// only commit point is the state.json write: before it nothing durable has
// changed, after it records and proposals are both gone, so no crash can leave
// void proposals beside live records. The clear audit line comes after the
// commit, so the ledger never reports a clear that did not happen. If the audit
// write fails, Clear writes the old state back; if that fails too, memory is
// left equal to the state file (cleared) and the error says the audit line is
// missing (see commit).
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev, prevSeq, prevEpoch := s.records, s.seq, s.epoch
	s.records = map[string]Record{}
	s.seq = 0
	s.epoch++
	undo := func() { s.records, s.seq, s.epoch = prev, prevSeq, prevEpoch }
	redo := func() { s.records, s.seq, s.epoch = map[string]Record{}, 0, prevEpoch+1 }
	// Clear has no caller name, so the actor is the store itself.
	return s.commit([]auditEntry{{action: "clear", id: "all", actor: "store"}}, undo, redo)
}

// MarkSuperseded records that id was replaced by by.
// Both records stay. id's SupersededBy and by's Supersedes are both set,
// so either side names the other. Query ranks the superseded one lower.
// The audit actor is by: the record that superseded, not the retrieval role.
func (s *Store) MarkSuperseded(id, by string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	if !ok {
		return fmt.Errorf("hmem: unknown memory %s", id)
	}
	repl, ok := s.records[by]
	if !ok {
		return fmt.Errorf("hmem: unknown memory %s", by)
	}
	prev := map[string]Record{id: rec, by: repl}
	rec.SupersededBy = by
	repl.Supersedes = id
	// One record replacing itself keeps both links on that one value.
	if id == by {
		rec.Supersedes = id
		s.records[id] = rec
	} else {
		s.records[id] = rec
		s.records[by] = repl
	}
	if err := s.publish("update", id, by, s.seq, prev, nil); err != nil {
		return err
	}
	return nil
}

// Get returns one record.
func (s *Store) Get(id string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	return rec, ok
}

// All returns every record, sorted by id.
func (s *Store) All() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshot()
}

// snapshot returns the records sorted by id. The caller holds s.mu.
func (s *Store) snapshot() []Record {
	out := make([]Record, 0, len(s.records))
	for _, rec := range s.records {
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// now is the store's time in Unix milliseconds, the unit of the memory record.
func (s *Store) now() int64 {
	if s.clock != nil {
		return s.clock().UnixMilli()
	}
	return time.Now().UnixMilli()
}

// publish writes the state file and then the audit line. On a failure it puts
// seq and saved records back and drops the new ids, so a refused write leaves
// memory equal to the state file and the ledger without a line for a change
// that did not happen.
func (s *Store) publish(action, id, actor string, seq int, saved map[string]Record, dropped []string) error {
	return s.publishWith(action, id, actor, nil, seq, saved, dropped)
}

// publishWith is publish with extra fields on the audit line: the ledger
// states (previousState, newState) and details a history is rebuilt from.
func (s *Store) publishWith(action, id, actor string, extra map[string]any, seq int, saved map[string]Record, dropped []string) error {
	return s.commitRecords([]auditEntry{{action: action, id: id, actor: actor, extra: extra}}, seq, saved, dropped)
}

// commitRecords is commit for a change to named records: saved holds the
// records as they were before the change, dropped the ids the change added,
// and seq the sequence before it. The change is already in memory.
func (s *Store) commitRecords(entries []auditEntry, seq int, saved map[string]Record, dropped []string) error {
	// What memory holds after the change, kept so a failed write-back can put
	// memory equal to the state file again.
	post := make(map[string]Record, len(saved)+len(dropped))
	for rid := range saved {
		post[rid] = s.records[rid]
	}
	for _, rid := range dropped {
		post[rid] = s.records[rid]
	}
	postSeq := s.seq
	undo := func() {
		s.seq = seq
		for rid, rec := range saved {
			s.records[rid] = rec
		}
		for _, rid := range dropped {
			delete(s.records, rid)
		}
	}
	redo := func() {
		s.seq = postSeq
		for rid, rec := range post {
			s.records[rid] = rec
		}
	}
	return s.commit(entries, undo, redo)
}

// commit makes a change that is already in memory durable: the state file
// first, then the audit lines in one write. The state file is the commit
// point, so the ledger never holds a line for a change that is not in it.
// undo puts memory back as it was and redo puts the change back.
//
// If the state write fails, undo runs and nothing durable changed. If the audit
// write fails after the state is written, undo runs and the old state is
// written back; the error is the audit failure. If that write-back fails as
// well, the state file holds the change, so redo runs: memory equals the state
// file, and the error says both that the state file holds the change and that
// the ledger has no line for it. It never claims a restore the file contradicts.
func (s *Store) commit(entries []auditEntry, undo, redo func()) error {
	if err := s.persist(); err != nil {
		undo()
		return err
	}
	err := s.auditBatch(entries)
	if err == nil {
		return nil
	}
	undo()
	if perr := s.persist(); perr != nil {
		redo()
		return fmt.Errorf("%w; the state file already holds the change and writing the old state back failed (%v), so memory now matches the state file and the audit ledger has no line for it", err, perr)
	}
	return err
}

// persist writes the whole store. Sorted records make the file the same for the same state.
func (s *Store) persist() error {
	if s.writeHook != nil {
		if err := s.writeHook("state"); err != nil {
			return err
		}
	}
	file := disk{Version: schemaVersion, Seq: s.seq, Epoch: s.epoch, Records: s.snapshot()}
	b, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return writeAtomic(filepath.Join(s.dir, "state.json"), b)
}

// audit appends one line to hmem.jsonl and syncs it (ADR-0018 point 3:
// audit is append-only JSONL in the run's logs).
func (s *Store) audit(action, id, actor string) error {
	return s.auditWith(action, id, actor, nil)
}

// auditEntry is one ledger line to write: the action, the memory id, the actor
// and the extra fields (ledger states and details).
type auditEntry struct {
	action, id, actor string
	extra             map[string]any
}

// auditWith is audit with extra fields. The four base fields cannot be
// overwritten by extra, so every line stays readable as a ledger entry.
func (s *Store) auditWith(action, id, actor string, extra map[string]any) error {
	return s.auditBatch([]auditEntry{{action: action, id: id, actor: actor, extra: extra}})
}

// auditBatch appends the lines of entries in one write, so a failure writes
// none of them: a change that needs several ledger lines never leaves some
// of them behind. The write hook is asked once per line before anything is
// written.
func (s *Store) auditBatch(entries []auditEntry) error {
	if len(entries) == 0 {
		return nil
	}
	if s.writeHook != nil {
		for _, e := range entries {
			if err := s.writeHook("audit:" + e.action); err != nil {
				return err
			}
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	now := s.now()
	for _, e := range entries {
		line := map[string]any{}
		for k, v := range e.extra {
			line[k] = v
		}
		line["memoryId"] = e.id
		line["action"] = e.action
		line["actor"] = e.actor
		line["timestamp"] = now
		if err := enc.Encode(line); err != nil {
			return err
		}
	}
	if err := appendSynced(s.auditPath, buf.Bytes(), s.appendFault); err != nil {
		return fmt.Errorf("hmem: audit: %w", err)
	}
	return nil
}

// findingKey joins the parts that name a finding across reviews.
func findingKey(rule, path, title string) string {
	return rule + "\n" + path + "\n" + title
}

// writeAtomic writes beside path and renames, so a crash never leaves a half-written state.json.
func writeAtomic(path string, b []byte) error {
	// A temporary file left by an earlier run is removed first and the new one
	// is created exclusively at 0600, so the mode of an old file never passes
	// through the rename onto state.json.
	tmp := path + ".tmp"
	// A directory at that name is not removed: the exclusive create below then fails.
	if info, err := os.Lstat(tmp); err == nil && !info.IsDir() {
		if err := os.Remove(tmp); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	// The data is synced before the rename, so a crash after the rename cannot
	// leave an empty state.json. The Close calls on the failure paths drop
	// their error on purpose: the write error is the one being returned.
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// clamp01 bounds x to [0, 1].
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// privateDir creates dir, or narrows an existing one, to 0700: MkdirAll sets a
// mode only when it creates, and a directory an earlier run left at 0755 must
// not stay readable by others (Proven's own state is readable only by its user).
func privateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.Chmod(dir, 0o700)
}
