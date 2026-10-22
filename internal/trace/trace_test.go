package trace_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/trace"
)

func TestAppendDoesNotRewrite(t *testing.T) {
	dir := t.TempDir()
	w, err := trace.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(trace.Event{FindingID: "F3", Kind: "claim", Agent: "planner", Text: "one"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Append(trace.Event{FindingID: "F3", Kind: "verdict", Agent: "judge", Text: "dropped: no evidence"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(w.Path())
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" || !bytesContains(b, "dropped: no evidence") || !bytesContains(b, "\"finding_id\":\"F3\"") {
		t.Fatalf("%s", b)
	}
}

func TestEveryEventHasRunIDAndTimeAndClose(t *testing.T) {
	dir := t.TempDir()
	w, err := trace.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(trace.Event{FindingID: "F1", Kind: "claim", Agent: "planner", Text: "one"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Append(trace.Event{FindingID: "F1", Kind: "verdict", Agent: "judge", Text: "kept"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := w.Append(trace.Event{Text: "after"}); err == nil {
		t.Fatal("append after close")
	}
	b, err := os.ReadFile(w.Path())
	if err != nil {
		t.Fatal(err)
	}
	var runID string
	n := 0
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e trace.Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		n++
		if e.RunID == "" {
			t.Fatalf("empty run id %s", line)
		}
		if runID == "" {
			runID = e.RunID
		}
		if e.RunID != runID {
			t.Fatalf("run id %s then %s", runID, e.RunID)
		}
		tm, err := time.Parse(time.RFC3339, e.Time)
		if err != nil {
			t.Fatalf("time %q: %v", e.Time, err)
		}
		if tm.Location() != time.UTC {
			t.Fatalf("time %s is not UTC", e.Time)
		}
		if !strings.HasSuffix(e.Time, "Z") {
			t.Fatalf("time %s is not UTC", e.Time)
		}
	}
	if n != 2 {
		t.Fatalf("events %d", n)
	}
}

func bytesContains(b []byte, s string) bool {
	return len(b) >= len(s) && (string(b) == s || len(s) == 0 || (len(b) > 0 && contains(string(b), s)))
}

func contains(h, n string) bool {
	return len(n) == 0 || (len(h) >= len(n) && (h == n || len(h) > 0 && search(h, n)))
}

func search(h, n string) bool {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return true
		}
	}
	return false
}
