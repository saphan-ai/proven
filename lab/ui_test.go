package lab

// What make up tells a person while it works: the plan right after the questions and before the first Docker call, one
// status line after each phase with its time and an empty line on each side, red naming the phase that failed, and a
// ready block that ends on START HERE and one empty line. Colour only on a terminal; NO_COLOR is respected.
//
// The runs use a stand-in `docker` that succeeds (or fails on one named call), so up.sh goes from the first phase to the
// banner without a container. The pseudo-terminal cases need python3 and a host that opens one; they are skipped by name
// where it cannot, and the painting itself is tested without one.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// okDocker succeeds at everything and prints nothing, except the one id up.sh asks for. A failure is named by FAIL_ON:
// the call whose arguments contain it exits 1. Every call is appended to docker.calls, so a test can see the order.
const okDocker = `#!/bin/sh
case "$1" in ps) exit 0 ;; version) echo linux/amd64; exit 0 ;; esac
echo "DOCKER-CALLED $*" >> "$FAKE_DIR/docker.calls"
if [ -n "$FAIL_ON" ]; then
  case " $* " in *"$FAIL_ON"*) exit 1 ;; esac
fi
case " $* " in
  *" ps -q dind "*) echo dind0001 ;;
esac
exit 0
`

// stubDemoCheck stands in for the wait on the reviews: it prints one progress line per pull request and exits with
// the status the test chose.
func stubDemoCheck(status string) string {
	return "#!/bin/sh\necho 'demo-check: [1/2] #1 demo-discount: defect, as the manifest says'\necho 'demo-check: [2/2] #2 clean-rename: clean, as the manifest says'\nexit " + status + "\n"
}

// phaseTree is a copy of the lab with demo-check.sh replaced, and the pull-request list the plan counts.
func phaseTree(t *testing.T, demoStatus string) string {
	t.Helper()
	tree := copyLab(t)
	for _, rel := range []string{"lab/seed/prs.tsv"} {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(tree, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tree, rel), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tree, "lab", "scripts", "demo-check.sh"), []byte(stubDemoCheck(demoStatus)), 0o755); err != nil {
		t.Fatal(err)
	}
	return tree
}

// fakeGit answers the three questions the image scripts ask of a checkout; the copy of the lab is not one.
const fakeGit = "#!/bin/sh\necho abc1234\n"

func upThrough(t *testing.T, tree string, env ...string) upResult {
	t.Helper()
	return runUp(t, upOpts{tree: tree, bins: map[string]string{"docker": okDocker, "git": fakeGit}, env: append([]string{"PROVEN_LAB_MODE=sample", "FAKE_DIR=" + t.TempDir()}, env...)})
}

// tail is the last n bytes of s, for a failure message.
func tail(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[len(s)-n:]
}

// orderDocker marks the output file at its first call: the marker lands exactly where the output stood at that moment.
const orderDocker = "#!/bin/sh\ncase \"$1\" in ps) exit 0;; esac\necho DOCKER-CALLED >> \"$OUT\"\nexit 97\n"

// upToFirstDocker runs up.sh with its output in a file the stand-in docker also appends to, so the order is the file's.
func upToFirstDocker(t *testing.T, tree string, answers string) string {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", orderDocker)
	out := filepath.Join(tmp, "out.txt")
	cmd := exec.Command("sh", "-c", `sh "$1/lab/scripts/up.sh" >>"$OUT" 2>&1`, "sh", tree)
	cmd.Env = append([]string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp, "FAKE_DIR=" + tmp, "OUT=" + out, "PROVEN_LAB_TTY=" + answers}, labPorts(t)...)
	_ = cmd.Run() // the stand-in docker fails every call, so up.sh stops: the file is what is read
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

var stepLine = regexp.MustCompile(`(?m)^==> \[(\d)/5\] (.+) \((\d+ s|\d+ min \d+ s)\)$`)

// around reports whether the line is preceded and followed by an empty line in out.
func around(out, line string) bool {
	return strings.Contains(out, "\n\n"+line+"\n\n")
}

func TestUpPrintsThePlanAfterTheQuestionsAndBeforeTheFirstDockerCall(t *testing.T) {
	tree := phaseTree(t, "0")
	out := upToFirstDocker(t, tree, ttyAnswers(t, "", ""))
	plan := strings.Index(out, "What will happen next:")
	if plan < 0 {
		t.Fatalf("no plan:\n%s", out)
	}
	if q := strings.Index(out, "[SAMPLE/live]"); q < 0 || q > plan {
		t.Errorf("the plan does not come after the questions: question at %d, plan at %d", q, plan)
	}
	if d := strings.Index(out, "DOCKER-CALLED"); d < plan {
		t.Errorf("the plan does not come before the first Docker call: plan at %d, docker at %d\n%s", plan, d, out)
	}
	for _, w := range []string{"1. Build Proven's image (about 1 min", "2. Start Gitea, Grafana and the lab's Docker", "3. Create the demo repository and 7 pull requests",
		"4. Proven reviews each pull request (a few minutes)", "5. Show you where to start"} {
		if !strings.Contains(out, w) {
			t.Errorf("the plan lacks %q:\n%s", w, out)
		}
	}
	// The negative control: the plan is printed after the Docker call. Mutated field: where up.sh calls ui_plan.
	mut := phaseTree(t, "0")
	mutate(t, mut, "lab/scripts/up.sh", "\nui_plan\nui_begin 1", "\nui_begin 1")
	mutate(t, mut, "lab/scripts/up.sh", "sh \"$ROOT/lab/scripts/image.sh\"", "sh \"$ROOT/lab/scripts/image.sh\"\nui_plan")
	m := upToFirstDocker(t, mut, ttyAnswers(t, "", ""))
	if d, p := strings.Index(m, "DOCKER-CALLED"), strings.Index(m, "What will happen next:"); p >= 0 && d >= p {
		t.Error("the control did not turn red: with ui_plan moved after the first Docker call the plan still came first")
	}
}

func TestUpPrintsOneGreenLineAfterEachPhaseWithAnEmptyLineOnEachSide(t *testing.T) {
	r := upThrough(t, phaseTree(t, "0"))
	if r.err != nil {
		t.Fatalf("up failed: %v\n%s", r.err, r.out)
	}
	got := stepLine.FindAllStringSubmatch(r.out, -1)
	if len(got) != 4 {
		t.Fatalf("want four step lines, got %d:\n%s", len(got), r.out)
	}
	for i, m := range got {
		if m[1] != string(rune('1'+i)) {
			t.Errorf("step line %d is numbered %s", i+1, m[1])
		}
		if !around(r.out, m[0]) {
			t.Errorf("the step line is not set off by an empty line before and after: %q", m[0])
		}
	}
	for _, w := range []string{"Proven's image is built", "Gitea, Grafana and the lab's Docker are up", "The demo repository and its 7 pull requests are created", "Proven reviewed every pull request"} {
		if !strings.Contains(r.out, w) {
			t.Errorf("no step line says %q", w)
		}
	}
	if strings.Contains(r.out, "\x1b") {
		t.Errorf("a control character was printed with no terminal:\n%q", r.out)
	}
	if strings.Contains(r.out, "FAILED") {
		t.Errorf("a failure was named on a run that succeeded:\n%s", r.out)
	}
	// One progress line per pull request reached the screen.
	if strings.Count(r.out, "demo-check: [") != 2 {
		t.Errorf("the wait's progress lines are not shown:\n%s", r.out)
	}
}

func TestUpNamesTheFailedPhaseInRedAndTheLogToRead(t *testing.T) {
	cases := []struct {
		name, failOn, want, log string
	}{
		{"the build", "buildx build", "==> [1/5] FAILED: Build Proven's image (after ", "Read: the build output above"},
		{"Gitea and Grafana", " up -d gitea", "==> [2/5] FAILED: Start Gitea, Grafana and the lab's Docker (after ", "Read: make logs SERVICE=gitea"},
		{"the seed", " run --rm -T seed", "==> [3/5] FAILED: Create the demo repository and its pull requests (after ", "Read: make logs SERVICE=seed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := upThrough(t, phaseTree(t, "0"), "FAIL_ON="+c.failOn)
			if r.err == nil {
				t.Fatalf("up did not fail:\n%s", r.out)
			}
			if !strings.Contains(r.out, c.want) || !strings.Contains(r.out, c.log) {
				t.Errorf("the red line does not name the phase and the log (%q, %q):\n%s", c.want, c.log, r.out)
			}
			if strings.Contains(r.out, "START HERE") {
				t.Error("the ready banner was printed after a failure")
			}
		})
	}
	// The negative control: the trap no longer names the phase. Mutated field: ui_exit in the EXIT trap of up.sh.
	tree := phaseTree(t, "0")
	mutate(t, tree, "lab/scripts/up.sh", `trap 'ui_exit $?; rm`, `trap 'rm`)
	if m := upThrough(t, tree, "FAIL_ON=buildx build"); strings.Contains(m.out, "FAILED:") {
		t.Error("the control did not turn red: with ui_exit out of the trap the failure was still named")
	}
}

func TestUpEndsOnStartHereAndOneEmptyLine(t *testing.T) {
	r := upThrough(t, phaseTree(t, "0"))
	startRe := regexp.MustCompile(`>>>>>>>>  START HERE:  http://127.0.0.1:[0-9]+/lab  <<<<<<<<\n\n$`)
	endsRight := func(out string) bool { return startRe.MatchString(out) && !strings.HasSuffix(out, "\n\n\n") }
	if !endsRight(r.out) {
		t.Errorf("the output does not end with START HERE and exactly one empty line:\n%q", tail(r.out, 200))
	}
	block := r.out[strings.Index(r.out, "####    Your proven lab is ready"):]
	for _, w := range []string{"Grafana dashboard: http://127.0.0.1:", "Wait for every review: make demo-check", "Help: https://github.com/saphan-ai/proven/blob/main/docs/WALKTHROUGH.md"} {
		if !strings.Contains(block, w) || strings.Index(block, w) > strings.Index(block, "START HERE") {
			t.Errorf("the ready block lacks %q before START HERE", w)
		}
	}
	// The negative control: a second empty line after the START HERE line. Mutated field: the newlines after it.
	tree := phaseTree(t, "0")
	mutate(t, tree, "lab/scripts/banner.sh", `printf '\n\n'`, `printf '\n\n\n'`)
	if m := upThrough(t, tree); endsRight(m.out) {
		t.Error("the control did not turn red: the banner did not change when its last newlines did")
	}
}

func TestUpStillShowsTheLabWhenAReviewDoesNotArrive(t *testing.T) {
	r := upThrough(t, phaseTree(t, "1"))
	if r.err == nil {
		t.Error("up exited 0 though a review did not arrive")
	}
	if !strings.Contains(r.out, "==> [4/5] FAILED: Wait for Proven to review every pull request: the lab is up, but not every review is there yet (after ") ||
		!strings.Contains(r.out, "Read: make logs SERVICE=runner, then make demo-check") {
		t.Errorf("the wait's failure is not named in red with its log:\n%s", r.out)
	}
	if !strings.HasSuffix(r.out, "<<<<<<<<\n\n") {
		t.Errorf("the lab was not shown after a late review:\n%q", tail(r.out, 200))
	}
}

func TestUpDoesNotWaitWhenToldNotTo(t *testing.T) {
	r := upThrough(t, phaseTree(t, "1"), "PROVEN_LAB_WAIT=no")
	if r.err != nil || strings.Contains(r.out, "demo-check: [") || !strings.Contains(r.out, "Not waiting for the reviews (PROVEN_LAB_WAIT=no)") {
		t.Errorf("PROVEN_LAB_WAIT=no did not skip the wait:\n%v\n%s", r.err, r.out)
	}
}

// shUI runs a function of ui.sh in a shell of its own, with the terminal test answered the way the case needs.
func shUI(t *testing.T, script string, env ...string) (stdout, stderr string) {
	t.Helper()
	cmd := exec.Command("sh", "-c", `ROOT="$1"; . "$ROOT/lab/scripts/ui.sh"; `+script, "sh", mustAbs(t, root))
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH")}, env...)
	var so, se strings.Builder
	cmd.Stdout, cmd.Stderr = &so, &se
	if err := cmd.Run(); err != nil {
		t.Fatalf("%v\n%s", err, se.String())
	}
	return so.String(), se.String()
}

func TestUiPaintsGreenAndRedOnlyWhereColourIsAllowed(t *testing.T) {
	forced := `ui_color() { return 0; }; ui_begin 2 "Start it" "make logs"; ui_done "It is up"; ui_begin 3 "Seed it" "make logs SERVICE=seed"; ui_fail`
	so, se := shUI(t, forced)
	if !strings.Contains(so, "\x1b[32m==> [2/5] It is up (0 s)\x1b[0m") {
		t.Errorf("the success line is not green: %q", so)
	}
	if !strings.Contains(se, "\x1b[31m==> [3/5] FAILED: Seed it (after 0 s). Read: make logs SERVICE=seed\x1b[0m") {
		t.Errorf("the failure line is not red: %q", se)
	}
	// Under a pipe the same words carry no control character: the stream is not a terminal.
	so, se = shUI(t, `ui_begin 2 "Start it" "make logs"; ui_done "It is up"; ui_begin 3 "Seed it" "x"; ui_fail`)
	if strings.Contains(so+se, "\x1b") || !strings.Contains(so, "==> [2/5] It is up (0 s)") {
		t.Errorf("colour or the words are wrong with no terminal: %q %q", so, se)
	}
	// The duration reads as seconds, then as minutes and seconds.
	if got, _ := shUI(t, `ui_duration 32; echo; ui_duration 65`); got != "32 s\n1 min 5 s" {
		t.Errorf("ui_duration: %q", got)
	}
}

// uiOnTerminal drives the same functions under a pseudo-terminal and returns what the terminal showed.
const uiPtyDriver = `import os, pty, sys, select, time
script = sys.argv[1]
try:
    pid, fd = pty.fork()
except OSError:
    sys.exit(77)
if pid == 0:
    os.execvp("sh", ["sh", "-c", script])
out = b""
end = time.time() + 15
while time.time() < end:
    r, _, _ = select.select([fd], [], [], 0.3)
    if not r:
        continue
    try:
        d = os.read(fd, 4096)
    except OSError:
        break
    if not d:
        break
    out += d
sys.stdout.buffer.write(out)
`

func uiOnTerminal(t *testing.T, tree string, env ...string) string {
	t.Helper()
	needTool(t, "python3")
	driver := filepath.Join(t.TempDir(), "drive.py")
	if err := os.WriteFile(driver, []byte(uiPtyDriver), 0o600); err != nil {
		t.Fatal(err)
	}
	script := `ROOT="` + mustAbs(t, tree) + `"; . "$ROOT/lab/scripts/ui.sh"; ui_begin 1 "Build" "log"; ui_done "Built"`
	cmd := exec.Command("python3", "-I", driver, script)
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH")}, env...)
	out, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 77 {
		t.Skip("this host opens no pseudo-terminal, so the terminal cases are not run here; they run wherever one opens")
	}
	if err != nil {
		t.Fatalf("the driver: %v", err)
	}
	return string(out)
}

func TestUiColourOnATerminalHonoursNoColorAndDumbTerminals(t *testing.T) {
	if out := uiOnTerminal(t, root, "TERM=xterm"); !strings.Contains(out, "\x1b[32m==> [1/5] Built") {
		t.Errorf("a terminal did not get the green line: %q", out)
	}
	for _, env := range [][]string{{"TERM=xterm", "NO_COLOR=1"}, {"TERM=dumb"}} {
		if out := uiOnTerminal(t, root, env...); strings.Contains(out, "\x1b") || !strings.Contains(out, "==> [1/5] Built") {
			t.Errorf("%v: colour was printed, or the line was lost: %q", env, out)
		}
	}
	// The negative control: NO_COLOR is no longer read. Mutated field: the NO_COLOR line of ui_color.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/ui.sh", `[ -z "${NO_COLOR:-}" ] || return 1`, `:`)
	if out := uiOnTerminal(t, tree, "TERM=xterm", "NO_COLOR=1"); !strings.Contains(out, "\x1b") {
		t.Error("the control did not turn red: with NO_COLOR ignored the line was still plain")
	}
}
