package testrun

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
)

// fixTrees writes a base tree and a change tree that differ in one line of
// billing/price.go, and returns them. The base is right; the change is wrong.
func fixTrees(t *testing.T) (base, change string) {
	t.Helper()
	base, change = t.TempDir(), t.TempDir()
	write := func(dir, rel, text string) {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(base, "billing/price.go", "package billing\n\nfunc Price() int { return 80 }\n")
	write(change, "billing/price.go", "package billing\n\nfunc Price() int { return 60 }\n")
	for _, dir := range []string{base, change} {
		write(dir, "billing/price_test.go", "package billing\n")
	}
	return base, change
}

// fixRequest is the request every test below starts from. The edit is the right fix.
func fixRequest(base, change string) FixRequest {
	return FixRequest{
		FindingID: "F1", Path: "billing/price_test.go", Name: "TestPrice",
		BaseDir: base, ChangeDir: change, Attempt: 1,
		Edits: []Edit{{Path: "billing/price.go", Old: "return 60", New: "return 80"}},
	}
}

// pricedRuntime stands in for the container runtime. A mounted tree passes the
// proving test when its price.go returns 80; it passes the package run when it
// also has no file named broken.go. It records the host directory of every run.
func pricedRuntime(t *testing.T, mounts *[]string) func(context.Context, []string) (string, error) {
	t.Helper()
	fail := exitCode(t, 1)
	return func(_ context.Context, argv []string) (string, error) {
		if argv[1] == "rm" {
			return "", nil
		}
		host := mountHost(argv)
		*mounts = append(*mounts, host)
		src, err := os.ReadFile(filepath.Join(host, "billing", "price.go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), "return 80") {
			return "FAIL got 60 want 80", fail
		}
		// Only a run of the whole package meets the neighbour's test.
		packageRun := !strings.Contains(strings.Join(argv, " "), " -run ")
		if _, err := os.Stat(filepath.Join(host, "billing", "broken.go")); err == nil && packageRun {
			return "FAIL TestNeighbour", fail
		}
		return "ok", nil
	}
}

// TestRunFixRunsTheProofThreeTimesAndThePackage proves ADR-0042 point 2: the
// proving test runs on the base, on the change and on the change with the patch,
// each with its command and exit code, and the package runs on the patched change.
// It also proves the patch lands in a scratch copy and never in the change tree.
func TestRunFixRunsTheProofThreeTimesAndThePackage(t *testing.T) {
	base, change := fixTrees(t)
	c := mustContainer(t, DefaultContainerConfig())
	var mounts []string
	c.run = pricedRuntime(t, &mounts)

	out, err := c.RunFix(context.Background(), fixRequest(base, change))
	if err != nil {
		t.Fatal(err)
	}
	if out.Refused != "" || len(out.Runs) != 4 {
		t.Fatalf("%#v", out)
	}
	want := []struct {
		label, result, command string
		code                   int
	}{
		{RunBase, ResultPassed, "go test -run '^TestPrice$' ./billing", 0},
		{RunChange, ResultFailed, "go test -run '^TestPrice$' ./billing", 1},
		{RunPatched, ResultPassed, "go test -run '^TestPrice$' ./billing", 0},
		{RunPackage, ResultPassed, "go test ./billing", 0},
	}
	for i, w := range want {
		got := out.Runs[i]
		if got.Label != w.label || got.Result != w.result || got.Command != w.command || got.ExitCode != w.code {
			t.Fatalf("run %d: %#v, want %#v", i, got, w)
		}
	}
	if mounts[0] != base || mounts[1] != change {
		t.Fatalf("the base and the change run on their own trees: %v", mounts)
	}
	if mounts[2] == change || mounts[2] == base || mounts[3] != mounts[2] {
		t.Fatalf("the patched runs share one scratch copy: %v", mounts)
	}
	// The change tree is the review's own; the patch must never reach it.
	got, err := os.ReadFile(filepath.Join(change, "billing", "price.go"))
	if err != nil || !strings.Contains(string(got), "return 60") {
		t.Fatalf("the change tree was written: %q %v", got, err)
	}
	if _, err := os.Stat(mounts[2]); err == nil {
		t.Fatalf("the scratch copy %s was not removed", mounts[2])
	}
}

// TestRunFixSkipsThePackageWhenThePatchedTestFails proves a patch that does not make
// the proving test pass ends the attempt there (ADR-0042 point 3): three runs, no package run.
func TestRunFixSkipsThePackageWhenThePatchedTestFails(t *testing.T) {
	base, change := fixTrees(t)
	c := mustContainer(t, DefaultContainerConfig())
	var mounts []string
	c.run = pricedRuntime(t, &mounts)
	req := fixRequest(base, change)
	// The mutated field: the new text no longer returns 80, so the patched test fails.
	req.Edits[0].New = "return 70"

	out, err := c.RunFix(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Runs) != 3 || out.Runs[2].Label != RunPatched || out.Runs[2].Result != ResultFailed || out.Runs[2].ExitCode != 1 {
		t.Fatalf("%#v", out.Runs)
	}
}

// TestRunFixReportsABrokenNeighbourTest proves a patch that fixes the proving test
// but breaks another test of the package is recorded as a failed package run.
func TestRunFixReportsABrokenNeighbourTest(t *testing.T) {
	base, change := fixTrees(t)
	// The change carries a neighbour file the patch will not remove: any patched tree has it.
	if err := os.WriteFile(filepath.Join(change, "billing", "broken.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := mustContainer(t, DefaultContainerConfig())
	var mounts []string
	c.run = pricedRuntime(t, &mounts)

	out, err := c.RunFix(context.Background(), fixRequest(base, change))
	if err != nil {
		t.Fatal(err)
	}
	last := out.Runs[len(out.Runs)-1]
	if len(out.Runs) != 4 || out.Runs[2].Result != ResultPassed || last.Label != RunPackage || last.Result != ResultFailed {
		t.Fatalf("%#v", out.Runs)
	}
}

// TestRunFixStartsNothingForARefusedPatch proves each refusal reason names the one
// field it guards and starts no container: a model's patch is checked before it runs.
func TestRunFixStartsNothingForARefusedPatch(t *testing.T) {
	base, change := fixTrees(t)
	c := mustContainer(t, DefaultContainerConfig())
	c.run = func(context.Context, []string) (string, error) {
		t.Fatal("a container was started for a refused patch")
		return "", nil
	}
	cases := []struct {
		name   string
		mutate func(*FixRequest)
		reason string
	}{
		{"no edit", func(r *FixRequest) { r.Edits = nil }, ReasonNoEdits},
		{"edit of the proving test", func(r *FixRequest) { r.Edits[0].Path = "billing/price_test.go" }, ReasonEditsATest},
		{"absolute path", func(r *FixRequest) { r.Edits[0].Path = "/etc/passwd" }, ReasonEditPath},
		{"path leaves the checkout", func(r *FixRequest) { r.Edits[0].Path = "../price.go" }, ReasonEditPath},
		{"hidden directory", func(r *FixRequest) { r.Edits[0].Path = ".git/config" }, ReasonEditPath},
		{"missing file", func(r *FixRequest) { r.Edits[0].Path = "billing/gone.go" }, ReasonEditMissing},
		{"empty old text", func(r *FixRequest) { r.Edits[0].Old = "" }, ReasonEditNoOld},
		{"old text not in the file", func(r *FixRequest) { r.Edits[0].Old = "return 61" }, ReasonEditNotFound},
		{"old text twice in the file", func(r *FixRequest) { r.Edits[0].Old = "n" }, ReasonEditAmbiguous},
		{"model-written test name", func(r *FixRequest) { r.Name = "TestX; rm -rf /" }, ReasonBadName},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := fixRequest(base, change)
			tc.mutate(&req)
			out, err := c.RunFix(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if out.Refused != tc.reason || len(out.Runs) != 0 {
				t.Fatalf("refused %q, want %q (%#v)", out.Refused, tc.reason, out)
			}
		})
	}
}

// TestRefuseEditsSeesEarlierEditsOfTheSameFile proves edits are planned in order:
// the second edit may name text the first one wrote, and a patch is never half applied.
func TestRefuseEditsSeesEarlierEditsOfTheSameFile(t *testing.T) {
	_, change := fixTrees(t)
	ok := []Edit{
		{Path: "billing/price.go", Old: "return 60", New: "return 61"},
		{Path: "billing/price.go", Old: "return 61", New: "return 80"},
	}
	if reason := RefuseEdits(change, ok, 0); reason != "" {
		t.Fatal(reason)
	}
	// The second edit names text that no earlier edit wrote.
	bad := []Edit{ok[0], {Path: "billing/price.go", Old: "return 99", New: "return 80"}}
	if reason := RefuseEdits(change, bad, 0); reason != ReasonEditNotFound {
		t.Fatal(reason)
	}
	got, _ := os.ReadFile(filepath.Join(change, "billing", "price.go"))
	if !strings.Contains(string(got), "return 60") {
		t.Fatalf("checking the patch wrote to the tree: %q", got)
	}
}

// TestRefuseEditsBoundsThePatch proves the edit count and the size are bounded.
func TestRefuseEditsBoundsThePatch(t *testing.T) {
	_, change := fixTrees(t)
	many := make([]Edit, maxEdits+1)
	for i := range many {
		many[i] = Edit{Path: "billing/price.go", Old: "return 60", New: "return 60"}
	}
	if reason := RefuseEdits(change, many, 0); reason != ReasonTooManyEdits {
		t.Fatal(reason)
	}
	big := []Edit{{Path: "billing/price.go", Old: "return 60", New: strings.Repeat("x", 100)}}
	if reason := RefuseEdits(change, big, 50); reason != ReasonPatchTooLarge {
		t.Fatal(reason)
	}
}

// TestRefuseEditsDoesNotFollowALink proves an edit cannot reach a file outside the
// checkout through a symbolic link inside it.
func TestRefuseEditsDoesNotFollowALink(t *testing.T) {
	_, change := fixTrees(t)
	outside := filepath.Join(t.TempDir(), "secret.go")
	if err := os.WriteFile(outside, []byte("return 60"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(change, "billing", "link.go")); err != nil {
		t.Skip("symbolic links are not available here")
	}
	edits := []Edit{{Path: "billing/link.go", Old: "return 60", New: "return 80"}}
	if reason := RefuseEdits(change, edits, 0); reason != ReasonEditPath {
		t.Fatal(reason)
	}
}

// TestSampleReplaysARecordedAttempt proves the sample returns the recorded runs of
// the attempt asked for, runs no code, and says so when an attempt has no recording.
func TestSampleReplaysARecordedAttempt(t *testing.T) {
	_, change := fixTrees(t)
	runs := []FixRun{{Label: RunBase, Command: "go test", Result: ResultPassed}}
	s := NewSample(nil, RecordedFix{FindingID: "F1", Attempt: 2, Runs: runs})
	req := FixRequest{FindingID: "F1", ChangeDir: change, Attempt: 2,
		Edits: []Edit{{Path: "billing/price.go", Old: "return 60", New: "return 80"}}}

	out, err := s.RunFix(context.Background(), req)
	if err != nil || len(out.Runs) != 1 || out.Runs[0].Label != RunBase {
		t.Fatalf("%#v %v", out, err)
	}
	req.Attempt = 1
	out, err = s.RunFix(context.Background(), req)
	if err != nil || len(out.Runs) != 0 || !strings.Contains(out.Refused, "no run") {
		t.Fatalf("%#v %v", out, err)
	}
}

// TestSampleChecksThePatchAgainstTheRealTree proves a recorded attempt is still
// refused when its edit does not apply to the tree under review.
func TestSampleChecksThePatchAgainstTheRealTree(t *testing.T) {
	_, change := fixTrees(t)
	s := NewSample(nil, RecordedFix{FindingID: "F1", Attempt: 1, Runs: []FixRun{{Label: RunBase}}})
	req := FixRequest{FindingID: "F1", ChangeDir: change, Attempt: 1,
		Edits: []Edit{{Path: "billing/price.go", Old: "return 99", New: "return 80"}}}
	out, err := s.RunFix(context.Background(), req)
	if err != nil || out.Refused != ReasonEditNotFound {
		t.Fatalf("%#v %v", out, err)
	}
}

// TestOffAndLiveProveNoFix proves an adapter that runs no container names the Prover
// off for a fix, as it does for a test.
func TestOffAndLiveProveNoFix(t *testing.T) {
	for name, r := range map[string]FixRunner{"off": Off{Reason: "Prover off: no runtime"}, "live": Live{}} {
		_, err := r.RunFix(context.Background(), FixRequest{})
		if d, ok := service.As(err); !ok || d.Service != "prover" {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
