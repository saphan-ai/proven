package pipeline

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/saphan-ai/proven/internal/trace"
)

func TestDeadlineOmitsUnjudgedFindings(t *testing.T) {
	stages := []string{"planned", "investigated", "proved", "refuted"}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := buildSample(t, t.TempDir(), nil)
			p.atBoundary = func(name string) {
				if name == stage {
					cancel()
				}
			}
			d, err := p.Review(ctx, Request{Target: fixturePath(t)})
			if err != nil {
				t.Fatal(err)
			}
			if !d.Partial {
				t.Fatal("want partial")
			}
			if len(d.Findings) != 0 {
				ids := make([]string, len(d.Findings))
				for i, f := range d.Findings {
					ids[i] = f.ID
				}
				t.Fatalf("unjudged findings %v", ids)
			}
			got := map[string]string{}
			for _, e := range readTrace(t, d.TracePath) {
				if e.Kind == "verdict" {
					got[e.FindingID] = e.Text
				}
			}
			for _, id := range []string{"F1", "F2", "F3", "F4"} {
				if got[id] != "not judged: deadline" {
					t.Fatalf("%s verdict %q", id, got[id])
				}
			}
		})
	}
}

func readTrace(t *testing.T, path string) []trace.Event {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []trace.Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e trace.Event
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
		out = append(out, e)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
