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

// detached returns the record with its own copy of every slice field. Every
// public call that hands out a Record returns it through detached, so editing
// what a caller was given never changes a stored record: the stored tags would
// otherwise change without the domain being classified again (see Dream).
func (r Record) detached() Record {
	r.Tags = cloneStrings(r.Tags)
	if r.Embedding != nil {
		r.Embedding = append(make([]float64, 0, len(r.Embedding)), r.Embedding...)
	}
	r.Entities = cloneStrings(r.Entities)
	r.RelatedTo = cloneStrings(r.RelatedTo)
	r.DerivedFrom = cloneStrings(r.DerivedFrom)
	return r
}

// cloneStrings copies a slice and keeps nil as nil and empty as empty, so a
// detached record encodes to the same JSON as the stored one.
func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	return append(make([]string, 0, len(in)), in...)
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
	// writeHook, when set, runs before each durable write ("dream", "state" or "audit:<action>") and may refuse it.
	// It is the seam tests use to inject a failed write; production leaves it nil.
	writeHook func(op string) error
}

// disk is the shape of state.json. Seq is kept so Add never reuses an id after
// a reopen. Clear sets it back to zero and writes that, so the next id starts again.
type disk struct {
	Version int      `json:"version"`
	Seq     int      `json:"seq"`
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
	if in.fixedID != "" {
		// The caller writes the ledger line after this, so a failed state
		// write leaves neither a record nor a line that names one.
		if err := s.persist(); err != nil {
			delete(s.records, rec.ID)
			return Record{}, err
		}
		return rec.detached(), nil
	}
	if err := s.publish("create", rec.ID, in.Source, prevSeq, nil, []string{rec.ID}); err != nil {
		return Record{}, err
	}
	return rec.detached(), nil
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
		return rec.detached(), nil
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
	return rec.detached(), nil
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
	return found.detached(), ok
}

// Clear drops every record, writes one clear audit line, resets the id
// sequence, and writes the empty store. Earlier audit lines stay.
// A failed audit or persist leaves the records and the sequence as they were.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev := s.records
	prevSeq := s.seq
	s.records = map[string]Record{}
	s.seq = 0
	// Clear has no caller name, so the actor is the store itself.
	if err := s.publish("clear", "all", "store", prevSeq, prev, nil); err != nil {
		return err
	}
	return nil
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

// Get returns one record. The caller owns the copy: editing its slices does
// not change the store.
func (s *Store) Get(id string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	return rec.detached(), ok
}

// All returns every record, sorted by id. The caller owns the copies: editing
// their slices does not change the store.
func (s *Store) All() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.snapshot()
	for i := range out {
		out[i] = out[i].detached()
	}
	return out
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

// publish writes the audit line and then the state file. On either failure it
// puts seq and saved records back and drops the new ids, so a refused write
// leaves memory equal to the state file.
func (s *Store) publish(action, id, actor string, seq int, saved map[string]Record, dropped []string) error {
	return s.publishWith(action, id, actor, nil, seq, saved, dropped)
}

// publishWith is publish with extra fields on the audit line: the ledger
// states (previousState, newState) and details a history is rebuilt from.
func (s *Store) publishWith(action, id, actor string, extra map[string]any, seq int, saved map[string]Record, dropped []string) error {
	err := s.auditWith(action, id, actor, extra)
	if err == nil {
		err = s.persist()
	}
	if err == nil {
		return nil
	}
	// Clear installed a fresh map; copying the saved records back into it
	// restores the old ones. Other callers pass the map they mutated, and
	// the saved values overwrite those entries.
	s.seq = seq
	for rid, rec := range saved {
		s.records[rid] = rec
	}
	for _, rid := range dropped {
		delete(s.records, rid)
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
	file := disk{Version: schemaVersion, Seq: s.seq, Records: s.snapshot()}
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

// auditWith is audit with extra fields. The four base fields cannot be
// overwritten by extra, so every line stays readable as a ledger entry.
func (s *Store) auditWith(action, id, actor string, extra map[string]any) error {
	if s.writeHook != nil {
		if err := s.writeHook("audit:" + action); err != nil {
			return err
		}
	}
	line := map[string]any{}
	for k, v := range extra {
		line[k] = v
	}
	line["memoryId"] = id
	line["action"] = action
	line["actor"] = actor
	line["timestamp"] = s.now()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(line); err != nil {
		return err
	}
	f, err := os.OpenFile(s.auditPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("hmem: audit: %w", err)
	}
	// OpenFile sets the mode only when it creates; an existing file is narrowed too.
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		return fmt.Errorf("hmem: audit: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("hmem: audit: %w", err)
	}
	return f.Sync()
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
	if _, err := f.Write(b); err != nil {
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
