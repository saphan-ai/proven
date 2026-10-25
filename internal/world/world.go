// Package world is the minimum world model: a store, an ingest of the trace
// log, and a context builder. The trace log is internal/trace. Memory of a
// refuted finding is stored in H-MEM.
//
// It holds the state of one review (ADR-0003 point 2), built to the Meterless
// World Model specification (ADR-0018 point 1). After the six agents have
// written their claims, evidence, refutations and verdicts to the trace, the
// pipeline folds that trace into this store. Folding is idempotent and
// resumable (ADR-0018 point 5): an event already applied is skipped, so a run
// killed halfway and started again reaches the same state. A refutation the
// Skeptic made with evidence is handed to H-MEM, which backs this store's
// memory (ADR-0018 point 4). A failed audit or persist puts the in-memory
// fold back, so memory and the state file agree.
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

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/pathguard"
	"github.com/saphan-ai/proven/internal/trace"
)

// schemaVersion is the state.json format. Open refuses any other version.
const schemaVersion = 1

// Provenance is required on every entity write. Kind says how the entity was
// learned, At when (for an ingested event, the event id), and By which agent.
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

// Report is one ingest pass. Applied and Skipped count new and already
// applied events, Stopped says the pass hit its stopAfter limit, and Changed
// says the folded state differs from before the pass.
type Report struct {
	Applied int
	Skipped int
	Stopped bool
	Changed bool
}

// Options open a store. Dir is durable state. LogDir receives the provenance JSONL.
// RepoRoot, when set, refuses a Dir inside that tree. Memory is the H-MEM
// store that keeps refutations; it is required.
type Options struct {
	Dir      string
	LogDir   string
	RepoRoot string
	Memory   *hmem.Store
}

// Store folds the trace log. The generator is accepted and never called
// (ADR-0018: model calls go through an injected generator, never inside the engine).
type Store struct {
	dir       string
	auditPath string
	mem       *hmem.Store
	gen       generator.Generator
	applied   map[string]struct{}
	findings  map[string]*Finding
	entities  map[string]Entity
}

// disk is the shape of state.json. Applied lists every event id already
// folded, which is what lets an ingest resume.
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
	if err := privateDir(dir); err != nil {
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
	// A state file an earlier run left at a wider mode is narrowed before it is read.
	if err := os.Chmod(filepath.Join(dir, "state.json"), 0o600); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("world: state file: %w", err)
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
	// Every list is checked for empty and duplicate ids: a damaged file is
	// refused rather than folded into a state that could skip or repeat events.
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

// refuseInside refuses a state directory at or below root, so review state
// never lands in the repository under review.
func refuseInside(dir, root string) error {
	if root == "" {
		return nil
	}
	// Both paths are resolved through their links first, so a state path that
	// reaches the repository through a link is refused; a link that does not
	// resolve is refused too.
	inside, err := pathguard.Inside(dir, root)
	if err != nil {
		return fmt.Errorf("world: state directory %s cannot be placed against the repository: %w", dir, err)
	}
	if inside {
		return fmt.Errorf("world: state directory %s is inside the repository", dir)
	}
	return nil
}

// SetGenerator stores the model interface. Ingest does not call it.
func (s *Store) SetGenerator(g generator.Generator) { s.gen = g }

// Generator is the model interface this store was given.
func (s *Store) Generator() generator.Generator { return s.gen }

// UpsertEntity writes one entity. Missing provenance is refused.
// A failed audit or persist leaves the entity map as it was.
func (s *Store) UpsertEntity(e Entity) error {
	if err := checkProvenance(e.Source); err != nil {
		return err
	}
	if e.ID == "" || e.Type == "" {
		return fmt.Errorf("world: an entity needs a type and an id")
	}
	// An entity written with no state is a live one.
	if e.State == "" {
		e.State = "active"
	}
	prev, existed := s.entities[e.ID]
	s.entities[e.ID] = e
	actor := e.Source.By
	if actor == "" {
		actor = "system"
	}
	// The audit id is built from the entity and its provenance time, so the
	// same write names the same line.
	err := s.audit(trace.Event{
		EventID:   "upsert:" + e.ID + ":" + e.Source.At,
		Kind:      "upsert",
		Agent:     actor,
		FindingID: e.ID,
		Text:      e.Name,
		Title:     e.Name,
	})
	if err == nil {
		err = s.persist()
	}
	if err != nil {
		if existed {
			s.entities[e.ID] = prev
		} else {
			delete(s.entities, e.ID)
		}
		return err
	}
	return nil
}

// checkProvenance refuses a write that does not say how and when it was learned.
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
	// The whole pass is one write. Any failure puts this snapshot back, so a
	// retry on the same store applies the events again instead of skipping
	// ones the state file does not hold (ADR-0018 point 5). H-MEM is not in
	// the snapshot: a refutation it already stored stays, and the retry
	// finds that same record.
	snap := s.cloneState()
	sc := bufio.NewScanner(r)
	// A trace line can carry a whole query result, so lines up to 4 MiB are read.
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	var rep Report
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev trace.Event
		if err := json.Unmarshal(line, &ev); err != nil {
			s.restoreState(snap)
			return Report{}, fmt.Errorf("world: log line: %w", err)
		}
		if ev.EventID == "" {
			sum := sha256.Sum256(line)
			ev.EventID = hex.EncodeToString(sum[:])[:16]
		}
		// ADR-0018 point 5: an event already applied is skipped, never folded twice.
		if _, ok := s.applied[ev.EventID]; ok {
			rep.Skipped++
			continue
		}
		if err := s.apply(ev); err != nil {
			s.restoreState(snap)
			return Report{}, err
		}
		s.applied[ev.EventID] = struct{}{}
		if err := s.audit(ev); err != nil {
			s.restoreState(snap)
			return Report{}, err
		}
		rep.Applied++
		if stopAfter > 0 && rep.Applied >= stopAfter {
			rep.Stopped = true
			break
		}
	}
	if err := sc.Err(); err != nil {
		s.restoreState(snap)
		return Report{}, err
	}
	// State is written once per pass. A pass killed before this line leaves
	// the old state.json, and the next pass applies the same events again.
	// A persist error is the same fact for the live store: roll the fold back.
	if err := s.persist(); err != nil {
		s.restoreState(snap)
		return Report{}, err
	}
	rep.Changed = s.Canonical() != before
	return rep, nil
}

// apply folds one event into its finding. Selection and memory events, and
// events with no finding, are marked applied but change nothing.
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
		evidenced := false
		if ev.Refutation != nil {
			ref.Refuted = ev.Refutation.Refuted
			ref.Reason = ev.Refutation.Reason
			evidenced = ev.Refutation.Evidenced
		}
		f.Refutations = append(f.Refutations, ref)
		// Only a refutation of a finding that had evidence, and that carries
		// evidence of its own, is remembered (ADR-0006 point 4): refuting a guess
		// teaches the next review nothing, and the Skeptic's word alone is not a
		// refutation. The time is the event's place in the fold, not the clock,
		// so a replay stores the same record.
		if ref.Refuted && ev.HasEvidence && evidenced {
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
	// An event that names a rule or a title also records the finding as an entity.
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

// putEntity stores an entity from an ingested event. Unlike UpsertEntity it
// writes no audit line or state, because Ingest does both once per event and pass.
func (s *Store) putEntity(e Entity) error {
	if err := checkProvenance(e.Source); err != nil {
		return err
	}
	s.entities[e.ID] = e
	return nil
}

// Context builds the read model for one finding, including a matched memory.
// The memory is matched by the finding's keys, so a finding refuted in an
// earlier review shows that refutation here.
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
		// The finding's own memory id, set during the fold, wins over the match.
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

// foldSnap is the in-memory fold before a pass, deep enough that later
// appends do not change it.
type foldSnap struct {
	applied  map[string]struct{}
	findings map[string]*Finding
	entities map[string]Entity
}

func (s *Store) cloneState() foldSnap {
	applied := make(map[string]struct{}, len(s.applied))
	for id := range s.applied {
		applied[id] = struct{}{}
	}
	findings := make(map[string]*Finding, len(s.findings))
	for id, f := range s.findings {
		findings[id] = cloneFinding(f)
	}
	entities := make(map[string]Entity, len(s.entities))
	for id, e := range s.entities {
		e.Attrs = cloneAttrs(e.Attrs)
		entities[id] = e
	}
	return foldSnap{applied: applied, findings: findings, entities: entities}
}

func (s *Store) restoreState(snap foldSnap) {
	s.applied = snap.applied
	s.findings = snap.findings
	s.entities = snap.entities
}

func cloneFinding(f *Finding) *Finding {
	cp := *f
	cp.Claims = append([]string(nil), f.Claims...)
	cp.Evidence = append([]string(nil), f.Evidence...)
	cp.Refutations = append([]Refutation(nil), f.Refutations...)
	return &cp
}

func cloneAttrs(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// persist writes the canonical state indented, through a temporary file and
// a rename, so a crash never leaves a half-written state.json.
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
	// A temporary file left by an earlier run is removed first and the new one
	// is created exclusively at 0600, so the mode of an old file never passes
	// through the rename onto state.json.
	tmp := filepath.Join(s.dir, "state.json.tmp")
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
	return os.Rename(tmp, filepath.Join(s.dir, "state.json"))
}

// audit appends one line to world.jsonl and syncs it (ADR-0018 point 3). The
// line names the event, its kind and its finding; the full event stays in the trace.
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
	f, err := os.OpenFile(s.auditPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("world: audit: %w", err)
	}
	// OpenFile sets the mode only when it creates; an existing file is narrowed too.
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		return fmt.Errorf("world: audit: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		return err
	}
	return f.Sync()
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
