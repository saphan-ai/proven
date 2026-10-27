package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeGit puts a git first on PATH that records its argv and its environment, one file per call,
// and answers merge-base with a commit name. It returns the directory the records go to.
func fakeGit(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	rec := t.TempDir()
	script := "#!/bin/sh\nn=$(ls '" + rec + "' | wc -l | tr -d ' ')\n" +
		"printf '%s\\n' \"$@\" > '" + rec + "'/argv-$n\n/usr/bin/env > '" + rec + "'/env-$n\n" +
		"echo 0123456789abcdef0123456789abcdef01234567\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return rec
}

// runCI runs `proven ci` with the given environment and returns the exit code.
func runCI(t *testing.T, env map[string]string) int {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string { return env[k] }, &stdout, &stderr, openCapture(&captureReviewer{}))
	if code == exitFatal {
		t.Fatalf("ci could not review: %s", stderr.String())
	}
	return code
}

// TestCIGitCannotFetchInLocalMode proves the wiring of applyCI: the merge base of `ci` is the one git
// read made before review.Open loads the mode, so ci tells git itself. With PROVEN_NETWORK=local the
// merge-base call carries protocol.allow=never and GIT_NO_LAZY_FETCH=1; with the mode unset it carries
// neither, so the switches come from the mode and nothing else.
// Mutated field: the offline policy applyCI builds for the merge base (a public one leaves both switches off).
func TestCIGitCannotFetchInLocalMode(t *testing.T) { ciGitSwitchesFollowTheMode(t) }

// TestCIGitCannotFetchInLocalModeWithPaddedWC is the same proof on a host whose wc pads its count
// with spaces, as BSD wc does. fakeGit counts its own calls with wc, so a padded count that is not
// stripped names the record file "argv-<spaces>0" and both subtests read no record.
// Mutated field: the `tr -d ' '` of fakeGit's count (removed, both subtests fail).
func TestCIGitCannotFetchInLocalModeWithPaddedWC(t *testing.T) {
	realWC, err := exec.LookPath("wc")
	if err != nil {
		t.Fatal(err)
	}
	shim := t.TempDir()
	body := "#!/bin/sh\n'" + realWC + "' \"$@\" | sed 's/^/        /'\n"
	if err := os.WriteFile(filepath.Join(shim, "wc"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shim+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := exec.Command("sh", "-c", "ls / | wc -l").Output()
	if err != nil || !strings.HasPrefix(string(out), " ") {
		t.Fatalf("the shim does not pad wc's count (%q, %v), so this test would prove nothing", out, err)
	}
	ciGitSwitchesFollowTheMode(t)
}

func ciGitSwitchesFollowTheMode(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		mode string
		want bool
	}{{"local", true}, {"", false}} {
		t.Run("mode="+tc.mode, func(t *testing.T) {
			rec := fakeGit(t)
			runCI(t, map[string]string{
				"PROVEN_NETWORK": tc.mode, "GITHUB_BASE_REF": "main", "GITHUB_SHA": "fedcba9876543210fedcba9876543210fedcba98",
			})
			argv, err := os.ReadFile(filepath.Join(rec, "argv-0"))
			if err != nil {
				t.Fatalf("git was not run: %v", err)
			}
			env, _ := os.ReadFile(filepath.Join(rec, "env-0"))
			if !strings.Contains(string(argv), "merge-base") {
				t.Fatalf("the first git call is not merge-base:\n%s", argv)
			}
			hasArg := strings.Contains(string(argv), "protocol.allow=never")
			hasEnv := strings.Contains(string(env), "GIT_NO_LAZY_FETCH=1")
			if hasArg != tc.want || hasEnv != tc.want {
				t.Fatalf("mode %q: protocol.allow=never %v, GIT_NO_LAZY_FETCH=1 %v; want both %v\nargv:\n%s", tc.mode, hasArg, hasEnv, tc.want, argv)
			}
		})
	}
}
