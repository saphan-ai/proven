package hmem_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

var appendStages = []string{hmem.StagePartial, hmem.StageWritten, hmem.StageSynced, hmem.StageClosed}

// failAppendAt fails every append to a log whose name ends in logName at stage,
// after that step ran.
func failAppendAt(logName, stage string) (func(path, stage string) error, error) {
	boom := errors.New("injected append failure")
	return func(path, st string) error {
		if st == stage && filepath.Base(path) == logName {
			return boom
		}
		return nil
	}, boom
}

// assertCleanLog proves a log file holds only whole, valid JSON lines: it is
// empty or ends with a newline and every line parses.
func assertCleanLog(t *testing.T, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatal(err)
	}
	if len(b) == 0 {
		return
	}
	if b[len(b)-1] != '\n' {
		t.Fatalf("%s ends inside a line: %q", path, b[max(0, len(b)-40):])
	}
	for _, line := range bytes.Split(bytes.TrimSuffix(b, []byte("\n")), []byte("\n")) {
		var v map[string]any
		if err := json.Unmarshal(line, &v); err != nil {
			t.Fatalf("%s holds a line that is not JSON: %q", path, line)
		}
	}
}

// TestFailedLedgerAppendLeavesNoLineAfterBytesLanded proves property 1 for the
// ledger: when a write fails after its bytes were placed (a short write, a
// write that reports an error, a failed sync, a failed close), the ledger file
// is byte-for-byte what it was, no change is visible after a reopen, and the
// next change is written and read without trouble. A store that only restores
// its state leaves a full line for a change that was rolled back, or a partial
// line that the next append joins.
func TestFailedLedgerAppendLeavesNoLineAfterBytesLanded(t *testing.T) {
	for _, stage := range appendStages {
		t.Run(stage, func(t *testing.T) {
			dir, logs := t.TempDir(), t.TempDir()
			ledger := filepath.Join(logs, "hmem.jsonl")
			s := openApprove(t, dir, logs)
			if _, err := s.Add(hmem.AddInput{Content: "first memory", Source: "t"}); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			hook, boom := failAppendAt("hmem.jsonl", stage)
			s.FailAppends(hook)
			if _, err := s.Add(hmem.AddInput{Content: "second memory", Source: "t"}); !errors.Is(err, boom) {
				t.Fatalf("Add error = %v, want the injected failure", err)
			}
			after, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatalf("the failed append left %d bytes in the ledger: %q", len(after)-len(before), after[len(before):])
			}
			s = openApprove(t, dir, logs)
			if len(s.All()) != 1 {
				t.Fatalf("%d records after reopen, want 1", len(s.All()))
			}
			// The next change lands whole beside the first.
			if _, err := s.Add(hmem.AddInput{Content: "third memory", Source: "t"}); err != nil {
				t.Fatal(err)
			}
			assertCleanLog(t, ledger)
			if h, err := s.History("mem-2"); err != nil || len(h) != 1 {
				t.Fatalf("History(mem-2) = %d entries, err %v, want 1", len(h), err)
			}
		})
	}
}

// TestFailedDreamAppendLeavesNoProposalAfterBytesLanded is property 1 for the
// dream log: a proposal-state append that fails after its bytes landed leaves
// the dream log as it was, so a returned error never hides a visible proposal.
func TestFailedDreamAppendLeavesNoProposalAfterBytesLanded(t *testing.T) {
	for _, stage := range appendStages {
		t.Run(stage, func(t *testing.T) {
			dir, logs := t.TempDir(), t.TempDir()
			s := openApprove(t, dir, logs)
			s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
			s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
			hook, boom := failAppendAt("dreams.jsonl", stage)
			s.FailAppends(hook)
			if _, err := s.Dream(); !errors.Is(err, boom) {
				t.Fatalf("Dream error = %v, want the injected failure", err)
			}
			if b, err := os.ReadFile(filepath.Join(logs, "dreams.jsonl")); err == nil && len(b) != 0 {
				t.Fatalf("the failed append left bytes in the dream log: %q", b)
			}
			s = openApprove(t, dir, logs)
			if pend, err := s.Pending(); err != nil || len(pend) != 0 {
				t.Fatalf("Pending = %d, err %v, want none", len(pend), err)
			}
			if ps, err := s.Dream(); err != nil || len(ps) == 0 {
				t.Fatalf("the next Dream = %d proposals, err %v, want some", len(ps), err)
			}
			assertCleanLog(t, filepath.Join(logs, "dreams.jsonl"))
		})
	}
}

// TestFailedRollbackOfAnAppendIsNamed proves an append whose bytes cannot be
// taken back says so: the error names the log and the offset after which its
// tail is not trustworthy, so a person is never told the ledger is clean when it
// is not.
func TestFailedRollbackOfAnAppendIsNamed(t *testing.T) {
	logs := t.TempDir()
	s := openApprove(t, t.TempDir(), logs)
	boom := errors.New("injected failure")
	s.FailAppends(func(path, stage string) error {
		if stage == hmem.StageWritten || stage == hmem.StageRollback {
			return boom
		}
		return nil
	})
	_, err := s.Add(hmem.AddInput{Content: "a memory", Source: "t"})
	if err == nil || !strings.Contains(err.Error(), "not trustworthy") || !strings.Contains(err.Error(), "hmem.jsonl") {
		t.Fatalf("error = %v, want it to name the untrusted ledger tail", err)
	}
}

// TestATornTailFromACrashDoesNotJoinTheNextLine proves a log that ends in half a
// line (a crash between two writes) is read around and then repaired: readers
// skip the unfinished line, and the next append cuts it off so it does not join
// the new line and make later reads fail.
func TestATornTailFromACrashDoesNotJoinTheNextLine(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	ledger := filepath.Join(logs, "hmem.jsonl")
	s := openApprove(t, dir, logs)
	if _, err := s.Add(hmem.AddInput{Content: "first memory", Source: "t"}); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(ledger, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"memoryId":"mem-9","action":"cre`); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if h, err := s.History("mem-1"); err != nil || len(h) != 1 {
		t.Fatalf("with a torn tail History = %d entries, err %v, want 1", len(h), err)
	}
	if _, err := s.Add(hmem.AddInput{Content: "second memory", Source: "t"}); err != nil {
		t.Fatal(err)
	}
	assertCleanLog(t, ledger)
	s = openApprove(t, dir, logs)
	if h, err := s.History("mem-2"); err != nil || len(h) != 1 {
		t.Fatalf("History(mem-2) = %d entries, err %v, want 1", len(h), err)
	}
}
