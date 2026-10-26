package hmem_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// domainEnv is a store holding one pending domain proposal for one record.
type domainEnv struct {
	dir, logs, proposal, record string
}

// newDomainEnv makes the stored domain stale by editing state.json, because
// the ordinary Add path assigns the domain with the same classifier Dream
// uses, so a stale domain cannot be made through Add alone.
func newDomainEnv(t *testing.T) *domainEnv {
	t.Helper()
	e := &domainEnv{dir: t.TempDir(), logs: t.TempDir()}
	s := openApprove(t, e.dir, e.logs)
	if _, err := s.Add(hmem.AddInput{Content: "deploy the service", Tags: []string{"tech:go"}, Source: "t"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(e.dir, "state.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte(`"domain": "tech"`), []byte(`"domain": "general"`), 1)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	s = openApprove(t, e.dir, e.logs)
	ps, err := s.Dream()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		if p.Type == "domain" {
			e.proposal = p.ID
		}
	}
	if e.proposal == "" {
		t.Fatalf("the fixture produced no domain proposal: %v", ps)
	}
	e.record = s.All()[0].ID
	return e
}

// TestReopenFinishesAnInterruptedDomainApproval proves, for each durable write
// of a domain approval, that reopening the store alone (no second Approve)
// ends with one approved proposal state, the record in the suggested domain
// once, and one proposal-linked update line. Recovery removed, the cases that
// stopped after the first write stay unfinished and this test fails.
func TestReopenFinishesAnInterruptedDomainApproval(t *testing.T) {
	cases := []struct {
		name string
		stop func(op string, dreams int) bool
	}{
		{"before the move (state write)", func(op string, _ int) bool { return op == "state" }},
		{"after the move, before its update line", func(op string, _ int) bool { return op == "audit:update" }},
		{"after the update line, before the approved state", func(op string, d int) bool { return op == "dream" && d == 2 }},
		{"before the dream_approved line", func(op string, _ int) bool { return op == "audit:dream_approved" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newDomainEnv(t)
			s := openApprove(t, e.dir, e.logs)
			dreams := 0
			hook, boom := failing(func(op string) bool {
				if op == "dream" {
					dreams++
				}
				return c.stop(op, dreams)
			})
			s.FailWrites(hook)
			if _, err := s.Approve(e.proposal); !errors.Is(err, boom) {
				t.Fatalf("approve error = %v, want the injected failure", err)
			}
			// A fresh process: no Approve is called before the checks.
			s = openApprove(t, e.dir, e.logs)
			if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalApproved {
				t.Fatalf("proposal state %q after reopen, want approved", got)
			}
			rec, ok := s.Get(e.record)
			if !ok || rec.Domain != "tech" {
				t.Fatalf("record domain %q (found %v), want tech", rec.Domain, ok)
			}
			h, _ := s.History(e.record)
			updates := 0
			for _, en := range h {
				if en.Action == "update" && en.Details["proposal"] == e.proposal {
					updates++
				}
			}
			if updates != 1 {
				t.Fatalf("%d proposal-linked update lines, want 1", updates)
			}
			if pend, _ := s.Pending(); len(pend) != 0 {
				t.Fatalf("%d proposals pending after reopen", len(pend))
			}
			if len(s.All()) != 1 {
				t.Fatalf("%d records, want 1", len(s.All()))
			}
			// A second reopen changes nothing: recovery is idempotent.
			s = openApprove(t, e.dir, e.logs)
			h2, _ := s.History(e.record)
			if len(h2) != len(h) {
				t.Fatalf("history grew from %d to %d lines on a second reopen", len(h), len(h2))
			}
		})
	}
}

// TestReopenDoesNotApproveAPendingDomainProposal proves recovery finishes only
// what a person decided: a pending domain proposal stays pending and its
// record stays where it was.
func TestReopenDoesNotApproveAPendingDomainProposal(t *testing.T) {
	e := newDomainEnv(t)
	s := openApprove(t, e.dir, e.logs)
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalProposed {
		t.Fatalf("proposal state %q, want proposed", got)
	}
	if rec, _ := s.Get(e.record); rec.Domain != "general" {
		t.Fatalf("record domain %q, want general", rec.Domain)
	}
}
