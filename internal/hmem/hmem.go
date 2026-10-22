// Package hmem is the minimum memory store and retrieval service.
// Dreaming, sleep and conflict resolution are not in this package.
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
)

const schemaVersion = 1

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

// Record is one memory. Source and provenance are mandatory.
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
	At         int64
}

// RefutationKey names a finding a skeptic refuted on one change.
type RefutationKey struct {
	ChangeKey string
	RuleID    string
	Path      string
	Title     string
	Reason    string
	At        int64
}

// Options open a store. Dir is the durable directory. LogDir receives the audit JSONL.
// RepoRoot, when set, refuses a Dir inside that tree.
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
}

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
	if err := os.MkdirAll(dir, 0o755); err != nil {
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
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("hmem: logs directory: %w", err)
	}
	s := &Store{
		dir:       dir,
		auditPath: filepath.Join(logDir, "hmem.jsonl"),
		clock:     opts.Clock,
		records:   map[string]Record{},
	}
	path := filepath.Join(dir, "state.json")
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
		return nil, fmt.Errorf("hmem: state version %d, want %d", file.Version, schemaVersion)
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
	return s, nil
}

func refuseInside(dir, root string) error {
	if root == "" {
		return nil
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("hmem: repository root: %w", err)
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return fmt.Errorf("hmem: state directory: %w", err)
	}
	if rel == "." || !strings.HasPrefix(rel, "..") {
		return fmt.Errorf("hmem: state directory %s is inside the repository", dir)
	}
	return nil
}

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

// SetGenerator stores the model interface. Remember, Add and Query do not call it.
func (s *Store) SetGenerator(g generator.Generator) { s.gen = g }

// Generator is the model interface this store was given.
func (s *Store) Generator() generator.Generator { return s.gen }

// Add inserts one record. A blank source is refused.
func (s *Store) Add(in AddInput) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.seq++
	rec := Record{
		ID:           fmt.Sprintf("mem-%d", s.seq),
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
	}
	rec.Namespace = namespace(rec.Domain, rec.Type)
	if err := validate(rec); err != nil {
		s.seq--
		return Record{}, err
	}
	s.records[rec.ID] = rec
	if err := s.audit("create", rec.ID, in.Source); err != nil {
		return Record{}, err
	}
	if err := s.persist(); err != nil {
		return Record{}, err
	}
	return rec, nil
}

// RememberRefutation stores a refuted finding. The same key returns the same record
// and does not append a second one.
func (s *Store) RememberRefutation(k RefutationKey) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if k.ChangeKey == "" || k.RuleID == "" || k.Path == "" {
		return Record{}, fmt.Errorf("hmem: a refutation needs a change, a rule and a path")
	}
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
	if err := s.audit("create", id, "skeptic"); err != nil {
		return Record{}, err
	}
	if err := s.persist(); err != nil {
		return Record{}, err
	}
	return rec, nil
}

// MatchRefutation returns the stored refutation of this finding on this change.
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
		if !ok || rec.ID < found.ID {
			found = rec
			ok = true
		}
	}
	return found, ok
}

// Clear drops every record and writes the empty store. The audit log is left in place.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = map[string]Record{}
	return s.persist()
}

// MarkSuperseded records that id was replaced by by.
// Both records stay. Query ranks the superseded one lower.
func (s *Store) MarkSuperseded(id, by string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	if !ok {
		return fmt.Errorf("hmem: unknown memory %s", id)
	}
	if _, ok := s.records[by]; !ok {
		return fmt.Errorf("hmem: unknown memory %s", by)
	}
	rec.SupersededBy = by
	s.records[id] = rec
	if err := s.audit("update", id, "retrieval"); err != nil {
		return err
	}
	return s.persist()
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

func (s *Store) snapshot() []Record {
	out := make([]Record, 0, len(s.records))
	for _, rec := range s.records {
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) now() int64 {
	if s.clock != nil {
		return s.clock().UnixMilli()
	}
	return time.Now().UnixMilli()
}

func (s *Store) persist() error {
	file := disk{Version: schemaVersion, Seq: s.seq, Records: s.snapshot()}
	b, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return writeAtomic(filepath.Join(s.dir, "state.json"), b)
}

func (s *Store) audit(action, id, actor string) error {
	line := map[string]any{
		"memoryId":  id,
		"action":    action,
		"actor":     actor,
		"timestamp": s.now(),
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(line); err != nil {
		return err
	}
	f, err := os.OpenFile(s.auditPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("hmem: audit: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("hmem: audit: %w", err)
	}
	return f.Sync()
}

func findingKey(rule, path, title string) string {
	return rule + "\n" + path + "\n" + title
}

func writeAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
