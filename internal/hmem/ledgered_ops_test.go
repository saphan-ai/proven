package hmem_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// ledgerOp is one operation that changes durable state and writes ledger lines,
// with the ledger action it writes and the memory id the line is about.
type ledgerOp struct {
	name   string
	action string
	id     string
	run    func(s *hmem.Store) error
}

// ledgerOps lists the operations a failed write must leave truthful. Each runs
// on a store holding mem-1 and mem-2, two related preferences.
var ledgerOps = []ledgerOp{
	{"add", "create", "mem-3", func(s *hmem.Store) error {
		_, err := s.Add(hmem.AddInput{Content: "a third fact about rivers", Source: "t"})
		return err
	}},
	{"supersede", "update", "mem-1", func(s *hmem.Store) error { return s.MarkSuperseded("mem-1", "mem-2") }},
	{"feedback", "feedback", "mem-1", func(s *hmem.Store) error { return s.Feedback("mem-1", "helpful") }},
	{"query", "read", "mem-1", func(s *hmem.Store) error {
		_, err := s.Query("prefer short bullet answers", 5, 0)
		return err
	}},
	{"dream", "dream_proposed", "dream-1", func(s *hmem.Store) error {
		_, err := s.Dream()
		return err
	}},
	{"reject", "dream_rejected", "dream-1", func(s *hmem.Store) error { return s.Reject("dream-1") }},
}

// TestALedgerLineExistsOnlyForAChangeThatHappened proves, for every operation
// that writes both state and a ledger line, that a failure of any one or two
// writes leaves the ledger and the durable state in agreement (ADR-0018 point
// 3). After each, a new process is opened: the ledger holds a line for the
// operation only when the operation's effect is durable, and the returned error
// is nil only then.
func TestALedgerLineExistsOnlyForAChangeThatHappened(t *testing.T) {
	for _, op := range ledgerOps {
		for _, fail := range failureSets(4) {
			t.Run(fmt.Sprintf("%s/%v", op.name, fail), func(t *testing.T) {
				dir, logs := t.TempDir(), t.TempDir()
				s := openApprove(t, dir, logs)
				s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
				s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
				if op.name == "reject" {
					if _, err := s.Dream(); err != nil {
						t.Fatal(err)
					}
				}
				before := recordIDs(s)
				s.FailWrites(failOnCalls(fail...))
				err := op.run(s)
				s.FailWrites(nil)

				re := openApprove(t, dir, logs)
				if got := recordIDs(s); got != recordIDs(re) {
					t.Fatalf("memory holds %q but a new process loads %q (err %v)", got, recordIDs(re), err)
				}
				for _, r := range re.All() {
					if m, ok := s.Get(r.ID); !ok || m.Confidence != r.Confidence || m.AccessCount != r.AccessCount || m.SupersededBy != r.SupersededBy {
						t.Fatalf("record %s differs between memory and a new process (err %v)", r.ID, err)
					}
				}
				lines := ledgerCount(t, re, op.id, op.action)
				effect := durableEffect(t, op, re, before)
				switch {
				case err == nil && (!effect || lines != 1):
					t.Fatalf("the call succeeded but effect=%v and the ledger has %d %s lines", effect, lines, op.action)
				case err != nil && lines != 0 && !effect:
					t.Fatalf("the ledger holds a %s line for a change that did not happen (err %v)", op.action, err)
				case err != nil && effect && !strings.Contains(err.Error(), "state file") && !strings.Contains(err.Error(), "without a ledger line"):
					// A durable change that was reported failed is allowed
					// only when the error names it.
					t.Fatalf("the change is durable but the error does not say so: %v", err)
				}
			})
		}
	}
}

// durableEffect reports whether the operation's effect is in the state a new
// process loaded.
func durableEffect(t *testing.T, op ledgerOp, re *hmem.Store, before string) bool {
	t.Helper()
	switch op.name {
	case "add":
		return recordIDs(re) != before
	case "supersede":
		r, _ := re.Get("mem-1")
		return r.SupersededBy == "mem-2"
	case "feedback":
		r, _ := re.Get("mem-1")
		return r.AccessCount == 1
	case "query":
		r, _ := re.Get("mem-1")
		return r.AccessCount == 1
	case "dream":
		pend, err := re.Pending()
		if err != nil {
			t.Fatal(err)
		}
		return len(pend) > 0
	case "reject":
		return hmem.ProposalStatus(re, "dream-1") == hmem.ProposalRejected
	}
	t.Fatalf("no durable effect defined for %s", op.name)
	return false
}
