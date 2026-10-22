// Package world is the minimum world model: a store, an ingest of the trace
// log, and a context builder. The trace log is internal/trace. Memory of a
// refuted finding is stored in H-MEM.
package world

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/trace"
)

const schemaVersion = 1

// Provenance is required on every entity write.
type Provenance struct {
	Kind string `json:"kind"`
	At   string `json:"at"`
	By   string `json:"by,omitempty"`
}

// Entity is one thing the world remembers. A finding is an entity.
type Entity struct {
	Type   string            `json:"type"`
	ID     string            `json:"id"`
	Name   string            `json:"name,omitempty"`
	Attrs  map[string]string `json:"attrs,omitempty"`
	Source Provenance        `json:"source"`
	State  string            `json:"state"`
}

// Finding is the folded view of one finding: claims, evidence, refutations, verdict.
type Finding struct {
	ID          string       `json:"id"`
	Claims      []string     `json:"claims,omitempty"`
	Evidence    []string     `json:"evidence,omitempty"`
	Refutations []Refutation `json:"refutations,omitempty"`
	Verdict     string       `json:"verdict,omitempty"`
	MemoryID    string       `json:"memory_id,omitempty"`
}

// Refutation is one skeptic attempt folded from the log.
type Refutation struct {
	Text    string `json:"text"`
	Refuted bool   `json:"refuted"`
	Reason  string `json:"reason"`
}

// Context is the structured payload a caller renders. It is not a prompt string.
type Context struct {
	FindingID   string        `json:"finding_id"`
	Claims      []string      `json:"claims"`
	Evidence    []string      `json:"evidence"`
	Refutations []Refutation  `json:"refutations"`
	Verdict     string        `json:"verdict"`
	MemoryID    string        `json:"memory_id,omitempty"`
	Memories    []hmem.Record `json:"memories"`
}

// Report is one ingest pass.
type Report struct {
	Applied int
	Skipped int
	Stopped bool
	Changed bool
}

// Options open a store. Dir is durable state. LogDir receives the provenance JSONL.
type Options struct {
	Dir      string
	LogDir   string
	RepoRoot string
	Memory   *hmem.Store
}

// Store folds the trace log. The generator is accepted and never called.
type Store struct {
	dir       string
	auditPath string
	mem       *hmem.Store
	gen       generator.Generator
	applied   map[string]struct{}
	findings  map[string]*Finding
	entities  map[string]Entity
}

type disk struct {
	Version  int       `json:"version"`
	Applied  []string  `json:"applied"`
	Findings []Finding `json:"findings"`
	Entities []Entity  `json:"entities"`
}

// Open loads state.json when it exists and refuses a bad file.
func Open(opts Options) (*Store, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("world: state directory is empty")
	}
	if opts.LogDir == "" {
		return nil, fmt.Errorf("world: logs directory is empty")
	}
	if opts.Memory == nil {
		return nil, fmt.Errorf("world: memory store is missing")
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("world: state directory: %w", err)
	}
	if err := refuseInside(dir, opts.RepoRoot); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("world: state directory: %w", err)
	}
	st, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("world: state path %s is not a directory", dir)
	}
	logDir, err := filepath.Abs(opts.LogDir)
	if err != nil {
		return nil, err
	}
	s := &Store{
		dir:       dir,
		auditPath: filepath.Join(logDir, "world.jsonl"),
		mem:       opts.Memory,
		applied:   map[string]struct{}{},
		findings:  map[string]*Finding{},
		entities:  map[string]Entity{},
	}
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("world: read state: %w", err)
	}
	var file disk
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, fmt.Errorf("world: state.json is not valid JSON: %w", err)
	}
	if file.Version != schemaVersion {
		return nil, fmt.Errorf("world: state version %d, want %d", file.Version, schemaVersion)
	}
	for _, id := range file.Applied {
		if id == "" {
			return nil, fmt.Errorf("world: applied list has an empty event id")
		}
		if _, ok := s.applied[id]; ok {
			return nil, fmt.Errorf("world: duplicate applied id %s", id)
		}
		s.applied[id] = struct{}{}
	}
	for _, f := range file.Findings {
		if f.ID == "" {
			return nil, fmt.Errorf("world: a finding has no id")
		}
		if _, ok := s.findings[f.ID]; ok {
			return nil, fmt.Errorf("world: duplicate finding %s", f.ID)
		}
		cp := f
		s.findings[f.ID] = &cp
	}
	for _, e := range file.Entities {
		if err := checkProvenance(e.Source); err != nil {
			return nil, fmt.Errorf("world: entity %s: %w", e.ID, err)
		}
		if _, ok := s.entities[e.ID]; ok {
			return nil, fmt.Errorf("world: duplicate entity %s", e.ID)
		}
		s.entities[e.ID] = e
	}
	return s, nil
}

func refuseInside(dir, root string) error {
	if root == "" {
		return nil
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return err
	}
	if rel == "." || !strings.HasPrefix(rel, "..") {
		return fmt.Errorf("world: state directory %s is inside the repository", dir)
	}
	return nil
}

// SetGenerator stores the model interface. Ingest does not call it.
func (s *Store) SetGenerator(g generator.Generator) { s.gen = g }

// Generator is the model interface this store was given.
func (s *Store) Generator() generator.Generator { return s.gen }

// UpsertEntity writes one entity. Missing provenance is refused.
func (s *Store) UpsertEntity(e Entity) error {
	if err := checkProvenance(e.Source); err != nil {
		return err
	}
	if e.ID == "" || e.Type == "" {
		return fmt.Errorf("world: an entity needs a type and an id")
	}
	if e.State == "" {
		e.State = "active"
	}
	s.entities[e.ID] = e
	return s.persist()
}

func checkProvenance(p Provenance) error {
	if p.Kind == "" || p.At == "" {
		return fmt.Errorf("world: a write with no provenance is refused")
	}
	return nil
}

// Ingest reads a JSONL trace. stopAfter limits how many new events this pass
// applies; zero means the rest of the log. An event id already applied is skipped.
// A line with no event_id takes its id from the sha256 of the line.
func (s *Store) Ingest(r io.Reader, stopAfter int) (Report, error) {
	before := s.Canonical()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	var rep Report
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev trace.Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return rep, fmt.Errorf("world: log line: %w", err)
		}
		if ev.EventID == "" {
			sum := sha256.Sum256(line)
			ev.EventID = hex.EncodeToString(sum[:])[:16]
		}
		if _, ok := s.applied[ev.EventID]; ok {
			rep.Skipped++
			continue
		}
		if err := s.apply(ev); err != nil {
			return rep, err
		}
		s.applied[ev.EventID] = struct{}{}
		if err := s.audit(ev); err != nil {
			return rep, err
		}
		rep.Applied++
		if stopAfter > 0 && rep.Applied >= stopAfter {
			rep.Stopped = true
			break
		}
	}
	if err := sc.Err(); err != nil {
		return rep, err
	}
	if err := s.persist(); err != nil {
		return rep, err
	}
	rep.Changed = s.Canonical() != before
	return rep, nil
}

func (s *Store) apply(ev trace.Event) error {
	if ev.Kind == "selection" || ev.Kind == "memory" || ev.FindingID == "" {
		return nil
	}
	f := s.findings[ev.FindingID]
	if f == nil {
		f = &Finding{ID: ev.FindingID}
		s.findings[ev.FindingID] = f
	}
	switch ev.Kind {
	case "claim":
		f.Claims = append(f.Claims, ev.Text)
	case "evidence":
		f.Evidence = append(f.Evidence, ev.Text)
	case "refutation":
		ref := Refutation{Text: ev.Text}
		if ev.Refutation != nil {
			ref.Refuted = ev.Refutation.Refuted
			ref.Reason = ev.Refutation.Reason
		}
		f.Refutations = append(f.Refutations, ref)
		if ref.Refuted && ev.HasEvidence {
			at := int64(len(s.applied) + 1)
			rec, err := s.mem.RememberRefutation(hmem.RefutationKey{
				ChangeKey: ev.ChangeKey,
				RuleID:    ev.RuleID,
				Path:      ev.Path,
				Title:     ev.Title,
				Reason:    ref.Reason,
				At:        at,
			})
			if err != nil {
				return err
			}
			f.MemoryID = rec.ID
		}
	case "verdict":
		f.Verdict = ev.Text
	}
	if ev.RuleID == "" && ev.Title == "" {
		return nil
	}
	id := ExternalID("finding", "review", ev.FindingID)
	return s.putEntity(Entity{
		Type:  "finding",
		ID:    id,
		Name:  ev.Title,
		Attrs: map[string]string{"rule": ev.RuleID, "path": ev.Path},
		Source: Provenance{
			Kind: "agent",
			At:   "event:" + ev.EventID,
			By:   ev.Agent,
		},
		State: "active",
	})
}

func (s *Store) putEntity(e Entity) error {
	if err := checkProvenance(e.Source); err != nil {
		return err
	}
	s.entities[e.ID] = e
	return nil
}

// Context builds the read model for one finding, including a matched memory.
func (s *Store) Context(findingID, changeKey, ruleID, path, title string) Context {
	out := Context{FindingID: findingID, Claims: []string{}, Evidence: []string{}, Refutations: []Refutation{}, Memories: []hmem.Record{}}
	if f := s.findings[findingID]; f != nil {
		out.Claims = append([]string(nil), f.Claims...)
		out.Evidence = append([]string(nil), f.Evidence...)
		out.Refutations = append([]Refutation(nil), f.Refutations...)
		out.Verdict = f.Verdict
		out.MemoryID = f.MemoryID
	}
	if rec, ok := s.mem.MatchRefutation(changeKey, ruleID, path, title); ok {
		out.Memories = []hmem.Record{rec}
		if out.MemoryID == "" {
			out.MemoryID = rec.ID
		}
	}
	return out
}

// Canonical is the folded state, with ids sorted, so two runs can be compared.
func (s *Store) Canonical() string {
	file := disk{
		Version:  schemaVersion,
		Applied:  make([]string, 0, len(s.applied)),
		Findings: make([]Finding, 0, len(s.findings)),
		Entities: make([]Entity, 0, len(s.entities)),
	}
	for id := range s.applied {
		file.Applied = append(file.Applied, id)
	}
	sort.Strings(file.Applied)
	ids := make([]string, 0, len(s.findings))
	for id := range s.findings {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		file.Findings = append(file.Findings, *s.findings[id])
	}
	eids := make([]string, 0, len(s.entities))
	for id := range s.entities {
		eids = append(eids, id)
	}
	sort.Strings(eids)
	for _, id := range eids {
		file.Entities = append(file.Entities, s.entities[id])
	}
	b, err := json.Marshal(file)
	if err != nil {
		return ""
	}
	return string(b)
}

func (s *Store) persist() error {
	var file disk
	if err := json.Unmarshal([]byte(s.Canonical()), &file); err != nil {
		return err
	}
	b, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := filepath.Join(s.dir, "state.json.tmp")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.dir, "state.json"))
}

func (s *Store) audit(ev trace.Event) error {
	line := map[string]string{
		"event_id":   ev.EventID,
		"kind":       ev.Kind,
		"finding_id": ev.FindingID,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(line); err != nil {
		return err
	}
	f, err := os.OpenFile(s.auditPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("world: audit: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		return err
	}
	return f.Sync()
}
