package hmem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Sleep thresholds, from the Meterless H-MEM specification (sleep cycle, section 8).
const (
	consolidationDays = 7
	archiveMinDays    = 30
	archiveMaxAccess  = 2
	synthesisCosine   = 0.82
	// hubThreshold is the number of relations that makes a record a hub.
	hubThreshold = 3
	dayMillis    = int64(24 * 60 * 60 * 1000)
)

// SleepPreview is the full plan of one sleep cycle. Preview computes it and
// changes nothing; Execute carries it out.
type SleepPreview struct {
	ToConsolidate []string   `json:"toConsolidate"`
	ToArchive     []string   `json:"toArchive"`
	ToSynthesize  [][]string `json:"toSynthesize"`
}

// SleepReport is what Execute did. BackupID names the snapshot taken first.
type SleepReport struct {
	BackupID     string   `json:"backupId"`
	Consolidated int      `json:"consolidated"`
	Archived     int      `json:"archived"`
	Synthesized  int      `json:"synthesized"`
	ActionLog    []string `json:"actionLog"`
}

// backup is a snapshot of the whole store, written before sleep changes it.
type backup struct {
	Seq     int      `json:"seq"`
	Records []Record `json:"records"`
}

// backupName matches the ids Execute makes, so Restore never builds a path
// from caller text that could leave the backups directory.
var backupName = regexp.MustCompile(`^backup-[0-9]+$`)

// SleepPreview returns the plan for a sleep cycle at the store's clock. It
// changes nothing: preview first, then execute (H-MEM section 8). If the dream
// log cannot be read, the sources of pending proposals are unknown, so the
// preview plans no archive at all: it fails closed, and SleepExecute returns
// the read error.
func (s *Store) SleepPreview() SleepPreview {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, err := s.sleepPlan()
	if err != nil {
		plan.ToArchive = []string{}
	}
	return plan
}

// sleepPlan computes the plan. The caller holds s.mu.
func (s *Store) sleepPlan() (SleepPreview, error) {
	now := s.now()
	all := s.snapshot()
	guard, err := s.guardsNow(all)
	if err != nil {
		return SleepPreview{ToConsolidate: []string{}, ToArchive: []string{}, ToSynthesize: [][]string{}}, err
	}
	plan := SleepPreview{ToConsolidate: []string{}, ToArchive: []string{}, ToSynthesize: [][]string{}}
	for _, m := range all {
		if m.Layer == LayerShort && now-m.LastAccessed >= consolidationDays*dayMillis {
			plan.ToConsolidate = append(plan.ToConsolidate, m.ID)
		}
	}
	for _, m := range all {
		if archiveEligible(m, now, guard) {
			plan.ToArchive = append(plan.ToArchive, m.ID)
		}
	}
	// A record that is itself derived, or already a source of a derived
	// record, is not synthesized again: sources stay in the store, so without
	// this every cycle would add another synthesis of the same memories.
	used := map[string]bool{}
	for _, m := range all {
		if len(m.DerivedFrom) > 0 || guard.sources[m.ID] {
			used[m.ID] = true
		}
	}
	for i := range all {
		if used[all[i].ID] {
			continue
		}
		group := []string{all[i].ID}
		for j := i + 1; j < len(all); j++ {
			if used[all[j].ID] {
				continue
			}
			if Cosine(all[i].Embedding, all[j].Embedding) >= synthesisCosine {
				group = append(group, all[j].ID)
			}
		}
		if len(group) >= 2 {
			for _, id := range group {
				used[id] = true
			}
			plan.ToSynthesize = append(plan.ToSynthesize, group)
		}
	}
	return plan, nil
}

// guards is what keeps a record out of the archive, computed from the whole
// store at one moment.
type guards struct {
	// referenced holds every id another record names: as a relation, as a
	// supersession partner or as a derivation source.
	referenced map[string]bool
	// sources holds the ids a derived record was made from.
	sources map[string]bool
	// conflicted holds the ids of both sides of an unresolved conflict.
	conflicted map[string]bool
	// pending holds the ids a pending dream proposal names as sources.
	pending map[string]bool
}

// guardsNow computes the guards over all records and over the dream log. The
// caller holds s.mu.
//
// A pending proposal (proposed, or approving and not yet finished) names its
// sources as the lineage of the record that approval will write. If sleep
// archived a source meanwhile, approval would write a lineage that points at
// nothing, and the archive backup is not a live record. So every source of a
// pending proposal is protected until the proposal is approved or rejected;
// after that the approved record's own derivedFrom protects the sources, and a
// rejection frees them.
func (s *Store) guardsNow(all []Record) (guards, error) {
	g := protections(all)
	ps, err := s.proposals()
	if err != nil {
		return guards{}, err
	}
	for _, p := range ps {
		if p.Status != ProposalProposed && p.Status != ProposalApproving {
			continue
		}
		for _, id := range p.DerivedFrom {
			g.pending[id] = true
		}
	}
	return g, nil
}

// protections computes the guards over all records. A resolved conflict is a
// supersession link, so checkConflict no longer reports it and the link guards
// both sides instead.
func protections(all []Record) guards {
	g := guards{referenced: map[string]bool{}, sources: map[string]bool{}, conflicted: map[string]bool{}, pending: map[string]bool{}}
	for _, m := range all {
		for _, r := range m.RelatedTo {
			g.referenced[r] = true
		}
		if m.Supersedes != "" {
			g.referenced[m.Supersedes] = true
		}
		if m.SupersededBy != "" {
			g.referenced[m.SupersededBy] = true
		}
		for _, d := range m.DerivedFrom {
			g.referenced[d] = true
			g.sources[d] = true
		}
	}
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			if _, ok := checkConflict(all[i], all[j]); ok {
				g.conflicted[all[i].ID] = true
				g.conflicted[all[j].ID] = true
			}
		}
	}
	return g
}

// archiveEligible reports whether sleep may archive m now. Guardrails (H-MEM
// section 8.2): a record that supersedes or is superseded, a record derived
// from others, a derivation source, a source of a pending dream proposal, a related record, a conflict member and
// a relationship hub are never archived, because losing one loses the reason
// for something else. Preview and Execute both ask it, so a supplied plan
// gets no way around the guard.
func archiveEligible(m Record, now int64, g guards) bool {
	if now-m.Timestamp < archiveMinDays*dayMillis || m.AccessCount > archiveMaxAccess {
		return false
	}
	if g.referenced[m.ID] || g.conflicted[m.ID] || g.pending[m.ID] {
		return false
	}
	return m.Supersedes == "" && m.SupersededBy == "" && len(m.DerivedFrom) == 0 && len(m.RelatedTo) < hubThreshold
}

// SleepExecute carries out a plan, after writing a backup of the whole store.
// A nil plan means the plan computed now. An id in the plan that is no longer
// in the store is skipped. An id that is no longer eligible for archive (the
// record changed since the preview, or the caller supplied it) is kept and
// logged: the guard is checked here, under the lock, for every id. Every step writes a ledger line; a failed write
// puts the records and the sequence back as they were.
func (s *Store) SleepExecute(plan *SleepPreview) (SleepReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if plan == nil {
		p, err := s.sleepPlan()
		if err != nil {
			return SleepReport{}, err
		}
		plan = &p
	}
	backupID, err := s.writeBackup()
	if err != nil {
		return SleepReport{}, err
	}
	prev := make(map[string]Record, len(s.records))
	for id, rec := range s.records {
		prev[id] = rec
	}
	prevSeq := s.seq
	rollback := func(err error) (SleepReport, error) {
		s.records = prev
		s.seq = prevSeq
		// A synthesis writes the state file as it adds; writing the restored
		// records keeps the file equal to memory. Its own failure adds nothing
		// to the error already being returned.
		_ = s.persist()
		return SleepReport{}, err
	}
	rep := SleepReport{BackupID: backupID, ActionLog: []string{"backup " + backupID + " created"}}
	for _, id := range plan.ToConsolidate {
		rec, ok := s.records[id]
		if !ok {
			continue
		}
		before := ledgerState(rec)
		rec.Layer = LayerLong
		s.records[id] = rec
		after := ledgerState(rec)
		if err := s.auditWith("sleep_consolidate", id, "sleep", map[string]any{
			"previousState": before, "newState": after,
			"details": map[string]any{"from": LayerShort, "to": LayerLong},
		}); err != nil {
			return rollback(err)
		}
		rep.Consolidated++
		rep.ActionLog = append(rep.ActionLog, "consolidated "+id)
	}
	// Guards are computed once, before any archive, from the store as it is
	// now: archiving one record must not unprotect another in the same cycle.
	guard, err := s.guardsNow(s.snapshot())
	if err != nil {
		return rollback(err)
	}
	for _, id := range plan.ToArchive {
		rec, ok := s.records[id]
		if !ok {
			continue
		}
		if !archiveEligible(rec, s.now(), guard) {
			rep.ActionLog = append(rep.ActionLog, "kept "+id+": protected or not old enough")
			continue
		}
		// The backup above is the way back: archiving removes the record from
		// the live store and the ledger keeps the line that says why.
		if err := s.auditWith("sleep_archive", id, "sleep", map[string]any{
			"details": map[string]any{"age": s.now() - rec.Timestamp, "accessCount": rec.AccessCount},
		}); err != nil {
			return rollback(err)
		}
		delete(s.records, id)
		rep.Archived++
		rep.ActionLog = append(rep.ActionLog, "archived "+id)
	}
	for _, group := range plan.ToSynthesize {
		var members []Record
		for _, id := range group {
			// A record already derived or already a source stays out, as in the plan.
			if rec, ok := s.records[id]; ok && len(rec.DerivedFrom) == 0 && !guard.sources[id] {
				members = append(members, rec)
			}
		}
		if len(members) < 2 {
			continue
		}
		merged, err := s.synthesize(members)
		if err != nil {
			return rollback(err)
		}
		if err := s.auditWith("sleep_synthesize", merged.ID, "sleep", map[string]any{
			"details": map[string]any{"sources": group},
		}); err != nil {
			return rollback(err)
		}
		// The sources stay: a synthesis is lineage, not replacement, and the
		// derived record's DerivedFrom must name live records.
		rep.Synthesized++
		rep.ActionLog = append(rep.ActionLog, fmt.Sprintf("synthesized %s from %s", merged.ID, strings.Join(group, ", ")))
	}
	if err := s.persist(); err != nil {
		return rollback(err)
	}
	return rep, nil
}

// synthesize merges members into one long-term record with lineage. The caller
// holds s.mu. The members stay in the store.
func (s *Store) synthesize(members []Record) (Record, error) {
	contents := make([]string, 0, len(members))
	ids := make([]string, 0, len(members))
	seen := map[string]bool{}
	var tags []string
	conf := 0.0
	for _, m := range members {
		contents = append(contents, m.Content)
		ids = append(ids, m.ID)
		for _, t := range m.Tags {
			if !seen[t] {
				seen[t] = true
				tags = append(tags, t)
			}
		}
		if m.Confidence > conf {
			conf = m.Confidence
		}
	}
	return s.addLocked(AddInput{
		Content:     "Synthesis: " + strings.Join(contents, " | "),
		Type:        TypeFactual,
		Layer:       LayerLong,
		Tags:        tags,
		Confidence:  &conf,
		Source:      "sleep:synthesis",
		Origin:      "sleep_synthesis",
		Label:       fmt.Sprintf("synthesized from %d memories", len(members)),
		DerivedFrom: ids,
	})
}

// writeBackup snapshots the whole store to backups/backup-N.json in the state
// directory, with the next free N, and returns the id. The caller holds s.mu.
func (s *Store) writeBackup() (string, error) {
	dir := filepath.Join(s.dir, "backups")
	if err := privateDir(dir); err != nil {
		return "", fmt.Errorf("hmem: backup directory: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("hmem: backup directory: %w", err)
	}
	id := fmt.Sprintf("backup-%d", len(entries)+1)
	b, err := json.MarshalIndent(backup{Seq: s.seq, Records: s.snapshot()}, "", "  ")
	if err != nil {
		return "", err
	}
	if err := writeAtomic(filepath.Join(dir, id+".json"), append(b, '\n')); err != nil {
		return "", fmt.Errorf("hmem: backup: %w", err)
	}
	return id, nil
}

// Restore puts the store back to a backup Execute made. An id that is not one
// of those is refused. A failed audit or write leaves the store as it was.
func (s *Store) Restore(backupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !backupName.MatchString(backupID) {
		return fmt.Errorf("hmem: unknown backup %q", backupID)
	}
	b, err := os.ReadFile(filepath.Join(s.dir, "backups", backupID+".json"))
	if err != nil {
		return fmt.Errorf("hmem: unknown backup %q", backupID)
	}
	var bk backup
	if err := json.Unmarshal(b, &bk); err != nil {
		return fmt.Errorf("hmem: backup %s is not valid JSON: %w", backupID, err)
	}
	prev := s.records
	prevSeq := s.seq
	s.records = map[string]Record{}
	for _, rec := range bk.Records {
		if err := validate(rec); err != nil {
			s.records, s.seq = prev, prevSeq
			return err
		}
		s.records[rec.ID] = rec
	}
	s.seq = bk.Seq
	if err := s.auditWith("restore", "*", "sleep", map[string]any{"details": map[string]any{"backupId": backupID}}); err != nil {
		s.records, s.seq = prev, prevSeq
		return err
	}
	if err := s.persist(); err != nil {
		s.records, s.seq = prev, prevSeq
		return err
	}
	return nil
}
