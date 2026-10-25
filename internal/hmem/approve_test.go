package hmem_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
)

// approveFixture opens a store with two related preferences and returns it with
// the id of one non-domain proposal.
func approveFixture(t *testing.T) (*hmem.Store, string) {
	t.Helper()
	s, err := hmem.Open(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Clock: func() time.Time { return time.UnixMilli(fixed) }})
	if err != nil {
		t.Fatal(err)
	}
	s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
	s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
	ps, err := s.Dream()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		if p.Type != "domain" {
			return s, p.ID
		}
	}
	t.Fatal("no non-domain proposal")
	return nil, ""
}

// assertApproveIsAtomic proves the property: when the named write fails, no
// long-term record remains, the proposal is still pending, the error is
// returned, and a retry creates exactly one record.
func assertApproveIsAtomic(t *testing.T, failOp string) {
	s, id := approveFixture(t)
	count := len(s.All())
	boom := errors.New("injected write failure")
	s.FailWrites(func(op string) error {
		if op == failOp {
			return boom
		}
		return nil
	})
	if _, err := s.Approve(id); !errors.Is(err, boom) {
		t.Fatalf("approve error = %v, want the injected failure", err)
	}
	if got := len(s.All()); got != count {
		t.Fatalf("a failed approval left %d records, want %d", got, count)
	}
	pend, _ := s.Pending()
	found := false
	for _, p := range pend {
		found = found || p.ID == id
	}
	if !found {
		t.Fatal("the proposal is no longer pending after a failed approval")
	}
	s.FailWrites(nil)
	rec, err := s.Approve(id)
	if err != nil || rec == nil {
		t.Fatalf("retry: %v %v", rec, err)
	}
	if got := len(s.All()); got != count+1 {
		t.Fatalf("after a retry there are %d records, want exactly %d", got, count+1)
	}
}

// TestApproveRollsBackWhenTheProposalStateWriteFails proves approval is one
// transition when the dream-log write fails after the record was created.
func TestApproveRollsBackWhenTheProposalStateWriteFails(t *testing.T) {
	assertApproveIsAtomic(t, "dream")
}

// TestApproveRollsBackWhenTheAuditWriteFails proves approval is one transition
// when the dream_approved audit write fails after the record and proposal state.
func TestApproveRollsBackWhenTheAuditWriteFails(t *testing.T) {
	assertApproveIsAtomic(t, "audit:dream_approved")
}

// TestApproveDomainProposalRollsBackReclassification proves a failed write
// after a domain proposal moved a record puts its domain back and keeps the
// proposal pending, and that a retry moves it once.
func TestApproveDomainProposalRollsBackReclassification(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	clock := func() time.Time { return time.UnixMilli(fixed) }
	s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: clock})
	if err != nil {
		t.Fatal(err)
	}
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
	if s, err = hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: clock}); err != nil {
		t.Fatal(err)
	}
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
	boom := errors.New("injected write failure")
	s.FailWrites(func(op string) error {
		if op == "dream" {
			return boom
		}
		return nil
	})
	if _, err := s.Approve(id); !errors.Is(err, boom) {
		t.Fatalf("approve error = %v", err)
	}
	if d := s.All()[0].Domain; d != "general" {
		t.Fatalf("domain %q after a failed approval, want general restored", d)
	}
	if pend, _ := s.Pending(); len(pend) != 1 {
		t.Fatalf("%d pending proposals, want 1", len(pend))
	}
	s.FailWrites(nil)
	if _, err := s.Approve(id); err != nil {
		t.Fatal(err)
	}
	if d := s.All()[0].Domain; d != "tech" {
		t.Fatalf("domain %q after the retry, want tech", d)
	}
}
