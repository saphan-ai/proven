package hmem_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
)

// approveEnv is a store with two related preferences and a pending proposal,
// with the directories kept so a test can reopen the same store.
type approveEnv struct {
	dir, logs string
	s         *hmem.Store
	proposal  string
	base      int // records before any approval
}

func openApprove(t *testing.T, dir, logs string) *hmem.Store {
	t.Helper()
	s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: func() time.Time { return time.UnixMilli(fixed) }})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newApproveEnv(t *testing.T) *approveEnv {
	t.Helper()
	e := &approveEnv{dir: t.TempDir(), logs: t.TempDir()}
	e.s = openApprove(t, e.dir, e.logs)
	e.s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
	e.s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
	ps, err := e.s.Dream()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		if p.Type != "domain" {
			e.proposal = p.ID
		}
	}
	if e.proposal == "" {
		t.Fatal("no non-domain proposal")
	}
	e.base = len(e.s.All())
	return e
}

// failing refuses the writes for which refuse returns true, with one error.
func failing(refuse func(op string) bool) (func(string) error, error) {
	boom := errors.New("injected write failure")
	return func(op string) error {
		if refuse(op) {
			return boom
		}
		return nil
	}, boom
}

// assertApprovedOnce proves the end state every failure point must reach: one
// long-term record with lineage, the proposal approved, none pending, and a
// history that names only records that exist. It reads from a store reopened
// on the same files, so what it checks is what is durable.
func assertApprovedOnce(t *testing.T, e *approveEnv) {
	t.Helper()
	s := openApprove(t, e.dir, e.logs)
	all := s.All()
	if len(all) != e.base+1 {
		t.Fatalf("%d records, want exactly %d", len(all), e.base+1)
	}
	long := 0
	var rec hmem.Record
	for _, r := range all {
		if r.Layer == hmem.LayerLong && r.Source == "dreaming" {
			long++
			rec = r
		}
	}
	if long != 1 {
		t.Fatalf("%d approved long-term records, want 1", long)
	}
	if len(rec.DerivedFrom) != 2 {
		t.Fatalf("approved record lineage %v, want the two source records", rec.DerivedFrom)
	}
	if pend, _ := s.Pending(); len(pend) != 0 {
		t.Fatalf("%d proposals still pending, want 0", len(pend))
	}
	if st := hmem.ProposalStatus(s, e.proposal); st != hmem.ProposalApproved {
		t.Fatalf("proposal status %q, want approved", st)
	}
	creates := 0
	approvals := 0
	for _, r := range all {
		h, err := s.History(r.ID)
		if err != nil {
			t.Fatal(err)
		}
		if r.ID == rec.ID {
			for _, en := range h {
				if en.Action == "create" {
					creates++
				}
			}
		}
	}
	if creates != 1 {
		t.Fatalf("history of %s has %d create lines, want 1", rec.ID, creates)
	}
	hp, _ := s.History(e.proposal)
	for _, en := range hp {
		if en.Action == "dream_approved" {
			approvals++
		}
	}
	if approvals != 1 {
		t.Fatalf("%d dream_approved lines, want 1", approvals)
	}
	assertNoPhantomCreate(t, e.logs, all)
}

// assertNoPhantomCreate proves the ledger names no created record that does not
// exist: every "create" line in the audit log belongs to a record in all.
func assertNoPhantomCreate(t *testing.T, logs string, all []hmem.Record) {
	t.Helper()
	exists := map[string]bool{}
	for _, r := range all {
		exists[r.ID] = true
	}
	b, err := os.ReadFile(filepath.Join(logs, "hmem.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var e hmem.LedgerEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		if e.Action == "create" && !exists[e.MemoryID] {
			t.Fatalf("the ledger says %s was created, but no such record exists", e.MemoryID)
		}
	}
}

// TestApproveRetriesAfterEachFailedWrite proves approval is idempotent: when
// any one durable write fails, a retry in the same process ends with exactly
// one long-term record with lineage and an approved proposal. The cases are the
// state write (nothing may be left behind), the create line, the proposal-state
// write and the approval line.
func TestApproveRetriesAfterEachFailedWrite(t *testing.T) {
	for _, op := range []string{"state", "audit:create", "dream", "audit:dream_approved"} {
		t.Run(op, func(t *testing.T) {
			e := newApproveEnv(t)
			hook, boom := failing(func(o string) bool { return o == op })
			e.s.FailWrites(hook)
			if _, err := e.s.Approve(e.proposal); !errors.Is(err, boom) {
				t.Fatalf("approve error = %v, want the injected failure", err)
			}
			if op == "state" {
				if got := len(e.s.All()); got != e.base {
					t.Fatalf("a failed state write left %d records, want %d", got, e.base)
				}
				assertNoPhantomCreate(t, e.logs, e.s.All())
			}
			e.s.FailWrites(nil)
			if rec, err := e.s.Approve(e.proposal); err != nil || rec == nil {
				t.Fatalf("retry: %v %v", rec, err)
			}
			assertApprovedOnce(t, e)
		})
	}
}

// TestApproveSurvivesAFailedSecondDreamWrite proves there is no second write
// that can strand a proposal: the proposal-state write fails, and so does every
// dream-log write after it (the write the earlier design used to compensate).
// Nothing is rolled back; a retry after a reopen finishes the one approval.
func TestApproveSurvivesAFailedSecondDreamWrite(t *testing.T) {
	e := newApproveEnv(t)
	hook, boom := failing(func(op string) bool { return op == "audit:dream_approved" || op == "dream" })
	e.s.FailWrites(hook)
	if _, err := e.s.Approve(e.proposal); !errors.Is(err, boom) {
		t.Fatalf("approve error = %v, want the injected failure", err)
	}
	s := openApprove(t, e.dir, e.logs)
	if _, err := s.Approve(e.proposal); err == nil || !strings.Contains(err.Error(), "already approved") {
		// Open already finished the approval, so a second call is refused.
		t.Fatalf("approve after reopen = %v, want the approval already finished", err)
	}
	assertApprovedOnce(t, e)
}

// TestReopenFinishesAnInterruptedApproval proves recovery needs no person: the
// process stops between the record write and the proposal write, and the next
// Open creates nothing new but completes the proposal and the ledger.
func TestReopenFinishesAnInterruptedApproval(t *testing.T) {
	for _, op := range []string{"audit:create", "dream", "audit:dream_approved"} {
		t.Run(op, func(t *testing.T) {
			e := newApproveEnv(t)
			hook, _ := failing(func(o string) bool { return o == op })
			e.s.FailWrites(hook)
			if _, err := e.s.Approve(e.proposal); err == nil {
				t.Fatal("approve succeeded despite the injected failure")
			}
			assertApprovedOnce(t, e) // this reopens the store, which recovers
		})
	}
}

// TestApproveCompletesAnApprovedProposalWhoseRecordIsMissing proves an older
// log is not left terminal: a proposal marked approved with no approval line
// and no record is completed by an explicit Approve, with exactly one record.
func TestApproveCompletesAnApprovedProposalWhoseRecordIsMissing(t *testing.T) {
	e := newApproveEnv(t)
	// Write the approved state by hand, as an older round left it.
	b, err := os.ReadFile(filepath.Join(e.logs, "dreams.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var p map[string]any
	for _, l := range bytes.Split(bytes.TrimSpace(b), []byte("\n")) {
		var q map[string]any
		if err := json.Unmarshal(l, &q); err != nil {
			t.Fatal(err)
		}
		if q["id"] == e.proposal {
			p = q
		}
	}
	p["status"] = "approved"
	line, _ := json.Marshal(p)
	f, err := os.OpenFile(filepath.Join(e.logs, "dreams.jsonl"), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(append(line, '\n'))
	f.Close()
	s := openApprove(t, e.dir, e.logs)
	if got := len(s.All()); got != e.base {
		t.Fatalf("reopen created %d records, want none", got-e.base)
	}
	if rec, err := s.Approve(e.proposal); err != nil || rec == nil {
		t.Fatalf("approve of the unfinished proposal: %v %v", rec, err)
	}
	assertApprovedOnce(t, e)
}

// TestApproveRefusesAFinishedApproval proves idempotence has a limit: once the
// approval is complete, a second Approve is refused and creates nothing.
func TestApproveRefusesAFinishedApproval(t *testing.T) {
	e := newApproveEnv(t)
	if _, err := e.s.Approve(e.proposal); err != nil {
		t.Fatal(err)
	}
	if _, err := e.s.Approve(e.proposal); err == nil {
		t.Fatal("a finished approval was approved again")
	}
	assertApprovedOnce(t, e)
}

// TestApproveDomainProposalRecoversForward proves a domain approval is one
// recoverable transition: a failed proposal-state write leaves the record in
// its new domain, the proposal pending, one update line, and a retry moves
// nothing twice.
func TestApproveDomainProposalRecoversForward(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	if _, err := s.Add(hmem.AddInput{Content: "deploy the service", Tags: []string{"tech:go"}, Source: "t"}); err != nil {
		t.Fatal(err)
	}
	// Make the stored domain stale ("general") so Dream suggests a move back.
	path := filepath.Join(dir, "state.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte(`"domain": "tech"`), []byte(`"domain": "general"`), 1)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	s = openApprove(t, dir, logs)
	ps, err := s.Dream()
	if err != nil {
		t.Fatal(err)
	}
	id := ""
	for _, p := range ps {
		if p.Type == "domain" {
			id = p.ID
		}
	}
	if id == "" {
		t.Fatalf("the fixture produced no domain proposal: %v", ps)
	}
	hook, boom := failing(func(op string) bool { return op == "dream" })
	s.FailWrites(hook)
	if _, err := s.Approve(id); !errors.Is(err, boom) {
		t.Fatalf("approve error = %v", err)
	}
	if pend, _ := s.Pending(); len(pend) != 1 {
		t.Fatalf("%d pending proposals, want 1", len(pend))
	}
	s.FailWrites(nil)
	if _, err := s.Approve(id); err != nil {
		t.Fatal(err)
	}
	s = openApprove(t, dir, logs)
	rec := s.All()[0]
	if rec.Domain != "tech" {
		t.Fatalf("domain %q after the retry, want tech", rec.Domain)
	}
	h, _ := s.History(rec.ID)
	updates := 0
	for _, en := range h {
		if en.Action == "update" {
			updates++
		}
	}
	if updates != 1 {
		t.Fatalf("%d update lines, want 1", updates)
	}
	if pend, _ := s.Pending(); len(pend) != 0 {
		t.Fatalf("%d proposals pending after the approval", len(pend))
	}
}
