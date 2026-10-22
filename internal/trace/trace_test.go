package trace_test

import (
	"os"
	"testing"

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
