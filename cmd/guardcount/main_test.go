package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/trace"
)

// update rewrites the golden file from the current output. Run it only after
// reading the diff: the golden file is what the test holds the command to.
var update = flag.Bool("update", false, "rewrite testdata/counts.golden")

// Property: on the committed sample traces the command prints exactly the
// golden table: rows per UTC day, side, verdict and backend in a stable
// order, the median latency, an offset time counted on its UTC day, an older
// down line counted as down:unstructured, and no row for "no guardrail" or
// for a line that is not a guardrail check.
func TestCountsMatchTheGoldenFile(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"testdata/live.trace.jsonl", "testdata/sample.trace.jsonl"}, &out, &errOut)
	if code != exitOK || errOut.Len() != 0 {
		t.Fatalf("exit %d stderr %q", code, errOut.String())
	}
	golden := filepath.Join("testdata", "counts.golden")
	if *update {
		if err := os.WriteFile(golden, out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("output differs from %s\ngot:\n%s\nwant:\n%s", golden, out.Bytes(), want)
	}
}

// Property: the output shown in docs/CONFIGURATION.md is the golden output,
// so the documented table cannot drift from what the command prints.
func TestDocsShowTheGoldenOutput(t *testing.T) {
	doc := readFile(t, filepath.Join("..", "..", "docs", "CONFIGURATION.md"))
	golden := readFile(t, filepath.Join("testdata", "counts.golden"))
	if !bytes.Contains(doc, append([]byte("```text\n"), golden...)) {
		t.Fatal("docs/CONFIGURATION.md does not show testdata/counts.golden")
	}
}

// Property: the order of the files does not change the table.
func TestFileOrderDoesNotChangeTheTable(t *testing.T) {
	var a, b bytes.Buffer
	run([]string{"testdata/live.trace.jsonl", "testdata/sample.trace.jsonl"}, &a, &bytes.Buffer{})
	run([]string{"testdata/sample.trace.jsonl", "testdata/live.trace.jsonl"}, &b, &bytes.Buffer{})
	if a.String() != b.String() {
		t.Fatalf("order changed the table:\n%s\n%s", a.String(), b.String())
	}
}

// Property (unplug, law 41): a review run with the guardrail switched off
// writes its one "no guardrail" line exactly as before, with no structured
// check, and the command prints the header and no row for it, with no error.
func TestGuardrailOffTraceHasNoRows(t *testing.T) {
	path := reviewWithGuardOff(t)
	var noGuard []trace.Event
	for _, line := range bytes.Split(readFile(t, path), []byte("\n")) {
		if !bytes.Contains(line, []byte(`"kind":"guardrail"`)) {
			continue
		}
		var e trace.Event
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		noGuard = append(noGuard, e)
	}
	if len(noGuard) != 1 {
		t.Fatalf("guardrail lines %d, want 1", len(noGuard))
	}
	if e := noGuard[0]; e.Text != "no guardrail" || e.Agent != "pipeline" || e.Guardrail != nil {
		t.Fatalf("the no-guardrail line changed: %#v", e)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{path}, &out, &errOut); code != exitOK || errOut.Len() != 0 {
		t.Fatalf("exit %d stderr %q", code, errOut.String())
	}
	if out.String() != "day  side  verdict  backend  count  median_ms\n" {
		t.Fatalf("output %q", out.String())
	}
}

// Property: a line that is not JSON is refused with the file and line number,
// not skipped, so a count is never made from half a file.
func TestBrokenLineIsRefusedWithItsPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, []byte("{\"kind\":\"claim\"}\nnot json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{path}, &out, &errOut); code != exitError {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(errOut.String(), path+":2:") {
		t.Fatalf("stderr %q", errOut.String())
	}
}

// Property: a guardrail line whose time is not RFC 3339 is refused, because
// its day cannot be known.
func TestGuardrailLineWithoutTimeIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	line := `{"kind":"guardrail","text":"prompt safe","guardrail":{"side":"prompt","verdict":"safe","latency_ms":1}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	var errOut bytes.Buffer
	if code := run([]string{path}, &bytes.Buffer{}, &errOut); code != exitError || !strings.Contains(errOut.String(), ":1:") {
		t.Fatalf("exit %d stderr %q", code, errOut.String())
	}
}

// Property: no argument, or an unknown flag, is a usage error; --help is not.
func TestUsage(t *testing.T) {
	cases := []struct {
		args []string
		code int
	}{
		{nil, exitUsage},
		{[]string{"-x"}, exitUsage},
		{[]string{"--help"}, exitOK},
		{[]string{filepath.Join(t.TempDir(), "missing.jsonl")}, exitError},
	}
	for _, tc := range cases {
		if code := run(tc.args, &bytes.Buffer{}, &bytes.Buffer{}); code != tc.code {
			t.Errorf("%v: exit %d, want %d", tc.args, code, tc.code)
		}
	}
}

// Property: the median of an even count is the mean of the two middle values,
// rounded down; of an odd count, the middle value.
func TestMedian(t *testing.T) {
	if m := median([]int64{120, 90, 100}); m != 100 {
		t.Fatalf("odd median %d", m)
	}
	if m := median([]int64{410, 300}); m != 355 {
		t.Fatalf("even median %d", m)
	}
	if m := median([]int64{1, 2}); m != 1 {
		t.Fatalf("rounded median %d", m)
	}
}

// reviewWithGuardOff runs the discount sample review with the guardrail
// switched off and returns its trace path.
func reviewWithGuardOff(t *testing.T) string {
	t.Helper()
	fx, err := fixture.Load(filepath.Join("..", "..", "testdata", "fixture", "discount"))
	if err != nil {
		t.Fatal(err)
	}
	gen, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := pipeline.Build(pipeline.Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: gen,
		GuardOff:  true,
		Env:       func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), pipeline.Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	return d.TracePath
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
