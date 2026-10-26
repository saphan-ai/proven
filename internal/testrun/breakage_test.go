package testrun

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBreakageArgvsNameEveryTouchedPackageOnce proves the commands that show a patch broke nothing
// it touched are built from the edit paths alone: one go test for the package of each edited file,
// once each, in name order; a file under testdata belongs to the package above it; and a module
// manifest runs every package below it. No word of an edit's text reaches a command.
func TestBreakageArgvsNameEveryTouchedPackageOnce(t *testing.T) {
	cases := map[string]struct {
		edits []Edit
		want  []string
	}{
		"one package":            {[]Edit{{Path: "billing/price.go", Old: "SECRET OLD", New: "$(rm -rf .)"}}, []string{"./billing"}},
		"two files, one package": {[]Edit{{Path: "billing/a.go"}, {Path: "billing/b.go"}}, []string{"./billing"}},
		"two packages in order":  {[]Edit{{Path: "shipping/s.go"}, {Path: "billing/a.go"}}, []string{"./billing", "./shipping"}},
		"the root package":       {[]Edit{{Path: "main.go"}}, []string{"."}},
		"a testdata file":        {[]Edit{{Path: "billing/testdata/in.txt"}}, []string{"./billing"}},
		"a root testdata file":   {[]Edit{{Path: "testdata/in.txt"}}, []string{"."}},
		"the root manifest":      {[]Edit{{Path: "go.mod"}, {Path: "billing/a.go"}}, []string{"./..."}},
		"the checksum file":      {[]Edit{{Path: "go.sum"}}, []string{"./..."}},
		"a nested manifest":      {[]Edit{{Path: "tools/go.mod"}}, []string{"./tools/..."}},
	}
	for name, tc := range cases {
		argvs := BreakageArgvs(tc.edits)
		var got []string
		for _, a := range argvs {
			if len(a) != 4 || a[0] != "go" || a[1] != "test" || a[2] != jsonFlag {
				t.Fatalf("%s: not a Proven-built go test -json argv: %v", name, a)
			}
			got = append(got, a[3])
		}
		if strings.Join(got, " ") != strings.Join(tc.want, " ") {
			t.Fatalf("%s: %v, want %v", name, got, tc.want)
		}
		for _, a := range argvs {
			if strings.ContainsAny(strings.Join(a, " "), "$()") || strings.Contains(strings.Join(a, " "), "SECRET") {
				t.Fatalf("%s: edit text reached a command: %v", name, a)
			}
		}
	}
}

// TestTheSampleReplaysARecordedPackageRunAndNamesOneThatWasNeverRun proves the replay of a breakage
// run: a package a recorded fix ran on the patched change replays as passed, a package no recording
// ran is reported as not run (never as passed), and a patch that does not apply is refused.
func TestTheSampleReplaysARecordedPackageRunAndNamesOneThatWasNeverRun(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"billing/a.go", "other/b.go"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, f)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, f), []byte("package x\n\nvar v = 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s := NewSample(nil, RecordedFix{FindingID: "F1", Attempt: 1, Runs: []FixRun{{
		Label: RunPackage, Command: "go test -json ./billing", Result: ResultPassed, ExitCode: 0,
		Output: "--- PASS: TestA (0.00s)\nok  \texample.com/billing\t0.001s",
	}}})
	out, err := s.RunBreakage(context.Background(), BreakageRequest{ChangeDir: dir, Edits: []Edit{
		{Path: "billing/a.go", Old: "v = 1", New: "v = 2"}, {Path: "other/b.go", Old: "v = 1", New: "v = 2"},
	}})
	if err != nil || out.Refused != "" || len(out.Runs) != 2 {
		t.Fatalf("%v %#v", err, out)
	}
	if r := out.Runs[0]; r.Label != "breakage ./billing" || r.Result != ResultPassed {
		t.Fatalf("the recorded package did not replay: %#v", r)
	}
	if r := out.Runs[1]; r.Label != "breakage ./other" || r.Result != ResultNotRun {
		t.Fatalf("a package no recording ran was not reported as not run: %#v", r)
	}
	out, err = s.RunBreakage(context.Background(), BreakageRequest{ChangeDir: dir, Edits: []Edit{{Path: "billing/a.go", Old: "absent", New: "x"}}})
	if err != nil || out.Refused != ReasonEditNotFound || len(out.Runs) != 0 {
		t.Fatalf("%v %#v", err, out)
	}
}

// TestAContainerWithNoRuntimeIsTheProverOffForABreakageRun proves a breakage run that cannot start
// is an error the caller leaves open, never an empty outcome that could be read as a pass.
func TestAContainerWithNoRuntimeIsTheProverOffForABreakageRun(t *testing.T) {
	var c *Container
	out, err := c.RunBreakage(context.Background(), BreakageRequest{ChangeDir: t.TempDir(), Edits: []Edit{{Path: "a.go", Old: "a", New: "b"}}})
	if err == nil || len(out.Runs) != 0 {
		t.Fatalf("%v %#v", err, out)
	}
}

// TestABreakageTargetOutsideEveryPackageIsNotRun proves a target go test cannot
// resolve at the module root is named as a run that could not happen, not a
// package that failed: a directory holding no Go file, and a directory inside a
// nested module. A package directory and the whole module resolve. The mutated
// line is unresolvable's verdict: made to return "" always, the first two cases
// fail.
func TestABreakageTargetOutsideEveryPackageIsNotRun(t *testing.T) {
	root := t.TempDir()
	write := func(rel string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("scripts/build.sh")
	write("pkg/a.go")
	write("tools/go.mod")
	write("tools/gen/main.go")
	for _, c := range []struct {
		target string
		notRun bool
	}{
		{"./scripts", true},
		{"./tools/gen", true},
		{"./tools/...", true},
		{"./pkg", false},
		{"./...", false},
	} {
		if got := unresolvable(root, c.target); (got != "") != c.notRun {
			t.Errorf("unresolvable(%s) = %q, want not run %v", c.target, got, c.notRun)
		}
	}
}
