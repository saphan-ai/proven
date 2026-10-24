package testrun

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// passedRun is the output of a go test -json run that passed, as a stand-in runtime
// prints it: a pass event for the test the argv names (or for one test, when the
// argv runs the whole package) and a pass event for the package.
func passedRun(argv []string) string {
	name := "TestOne"
	for i, w := range argv {
		if w == "-run" && i+1 < len(argv) {
			name = strings.Trim(argv[i+1], "^$")
		}
	}
	events := []map[string]string{
		{"Action": "run", "Package": "example.com/p", "Test": name},
		{"Action": "output", "Package": "example.com/p", "Test": name, "Output": "--- PASS: " + name + " (0.00s)\n"},
		{"Action": "pass", "Package": "example.com/p", "Test": name},
		{"Action": "output", "Package": "example.com/p", "Output": "ok  \texample.com/p\t0.001s\n"},
		{"Action": "pass", "Package": "example.com/p"},
	}
	var b strings.Builder
	for _, ev := range events {
		line, _ := json.Marshal(ev)
		b.Write(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// TestARunPassesOnlyWhenItsOutputShowsTheTestPassed proves the property of ADR-0042
// that a run is "passed" when the named test ran and passed, and not when the
// process only ended with exit code 0. The mutated field is the output of a run
// that exits 0: the same exit code is a pass with a pass event and is not one without.
func TestARunPassesOnlyWhenItsOutputShowsTheTestPassed(t *testing.T) {
	testRun := []string{"go", "test", "-json", "-run", "^TestPrice$", "./billing"}
	packageRun := []string{"go", "test", "-json", "./billing"}
	cases := []struct {
		name   string
		argv   []string
		output string
		want   bool
	}{
		{"test passed", testRun, passedRun(testRun), true},
		{"package passed", packageRun, passedRun([]string{"-run", "^TestOne$"}), true},
		{"exit 0 with no output (the process ended early)", testRun, "", false},
		{"exit 0 with the plain ok line only", testRun, "ok  \texample.com/p\t0.001s\n", false},
		{"exit 0 with the package passing and no test", testRun,
			`{"Action":"output","Package":"p","Output":"testing: warning: no tests to run\n"}` + "\n" + `{"Action":"pass","Package":"p"}` + "\n", false},
		{"another test passed, the named one did not", testRun,
			`{"Action":"pass","Package":"p","Test":"TestOther"}` + "\n" + `{"Action":"pass","Package":"p"}` + "\n", false},
		{"the named test was skipped", testRun,
			`{"Action":"skip","Package":"p","Test":"TestPrice"}` + "\n" + `{"Action":"pass","Package":"p"}` + "\n", false},
		{"package run with no test to run", packageRun, `{"Action":"pass","Package":"p"}` + "\n", false},
		{"package run with a test and no package pass", packageRun, `{"Action":"pass","Package":"p","Test":"TestA"}` + "\n", false},
		{"a pass event printed as program text is not an event", testRun,
			`{"Action":"output","Package":"p","Test":"TestPrice","Output":"{\"Action\":\"pass\",\"Test\":\"TestPrice\"}\n"}` + "\n", false},
	}
	for _, tc := range cases {
		_, got := judgeProof(tc.argv, tc.output)
		if got != tc.want {
			t.Errorf("%s: judged passed=%v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestAnEarlyExitWithCodeZeroIsNotAPassInTheProofRuns proves, through the container
// runner, that a patch which makes the package end the process with exit code 0 before
// any test starts is not shown as passing: the change+patch run and the package run
// have no pass event, so they do not count as passed (ADR-0042, the proof is the test).
func TestAnEarlyExitWithCodeZeroIsNotAPassInTheProofRuns(t *testing.T) {
	base, change := fixTrees(t)
	c := mustContainer(t, DefaultContainerConfig())
	fail := exitCode(t, 1)
	c.run = func(_ context.Context, argv []string) (string, error) {
		if argv[1] == "rm" {
			return "", nil
		}
		host := mountHost(argv)
		src, err := os.ReadFile(host + "/billing/price.go")
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.Contains(string(src), "os.Exit(0)"):
			// The patched package ends the process at once: exit 0, no test events.
			return "", nil
		case strings.Contains(string(src), "return 80"):
			return passedRun(argv), nil
		default:
			return "FAIL got 60 want 80", fail
		}
	}
	req := fixRequest(base, change)
	// The mutated field: the patch leaves the price wrong and adds an early exit.
	req.Edits[0].New = "return 60 }\n\nfunc init() { os.Exit(0) "

	out, err := c.RunFix(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Runs) < 3 {
		t.Fatalf("%#v", out)
	}
	patched := out.Runs[2]
	if patched.Label != RunPatched || patched.ExitCode != 0 || patched.Result == ResultPassed {
		t.Fatalf("an early exit with code 0 was recorded as %#v", patched)
	}
	if !strings.Contains(patched.Output, "no pass event") {
		t.Fatalf("the record does not say why: %q", patched.Output)
	}
	for _, r := range out.Runs {
		if r.Label == RunPackage {
			t.Fatalf("the package run happened after a patched run that did not pass: %#v", out.Runs)
		}
	}
}

// TestTheProversBaseRunNeedsTheTestsPassEvent proves the Prover's own runs use the
// same judgement: a base run that exits 0 without a pass event does not show the test
// passes on the base, so no finding is proved by it.
func TestTheProversBaseRunNeedsTheTestsPassEvent(t *testing.T) {
	base, change := t.TempDir(), t.TempDir()
	c := mustContainer(t, DefaultContainerConfig())
	fail := exitCode(t, 1)
	c.run = func(_ context.Context, argv []string) (string, error) {
		if argv[1] == "rm" {
			return "", nil
		}
		if strings.Contains(mountOf(argv), change+":/checkout:ro") {
			return "FAIL TestDiscounted", fail
		}
		return "", nil
	}
	out, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: base, ChangeDir: change,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.FailsOnChange || out.PassesOnBase {
		t.Fatalf("fails_on_change=%v passes_on_base=%v: a base run with no pass event passed", out.FailsOnChange, out.PassesOnBase)
	}
}

// TestARecordedPassNeedsAPassLineToBeReplayed proves the sample applies the same
// judgement to a recording: a recorded run marked passed whose output shows no pass
// line is replayed as not run.
func TestARecordedPassNeedsAPassLineToBeReplayed(t *testing.T) {
	base, change := fixTrees(t)
	cmd := "go test -json -run '^TestPrice$' ./billing"
	recorded := func(output string) *Sample {
		return NewSample(nil, RecordedFix{FindingID: "F1", Attempt: 1, Runs: []FixRun{
			{Label: RunBase, Command: cmd, Result: ResultPassed, Output: output},
		}})
	}
	req := fixRequest(base, change)
	req.FindingID = "F1"
	for _, tc := range []struct {
		name, output, want string
	}{
		{"a pass line", "=== RUN   TestPrice\n--- PASS: TestPrice (0.00s)\nPASS\nok  \texample.com/p\t0.001s", ResultPassed},
		{"no pass line", "ok  \texample.com/p\t0.001s", ResultNotRun},
		{"empty output", "", ResultNotRun},
	} {
		out, err := recorded(tc.output).RunFix(context.Background(), req)
		if err != nil || len(out.Runs) != 1 {
			t.Fatalf("%s: %#v %v", tc.name, out, err)
		}
		if out.Runs[0].Result != tc.want {
			t.Errorf("%s: replayed as %q, want %q", tc.name, out.Runs[0].Result, tc.want)
		}
	}
}
