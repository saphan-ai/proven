package lab

// Round 4: the Meterless feature questions and settings, the page line that says which are on, livecheck's use of the
// workflow's own derivation, the detection of Zetaris and Nightshift by their own answers, and the trigger
// configuration that keeps one review per head commit. A guard is tested by mutating exactly the thing it guards in a
// copy and showing the same assertion fail; each control names the field it changes.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// copyInto copies more of the repository into a tree made by copyLab.
func copyInto(t *testing.T, tree string, rels ...string) {
	t.Helper()
	for _, rel := range rels {
		src := filepath.Join(root, rel)
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			r, _ := filepath.Rel(root, path)
			if info.IsDir() {
				return os.MkdirAll(filepath.Join(tree, r), 0o755)
			}
			b, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			return os.WriteFile(filepath.Join(tree, r), b, info.Mode())
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// versionsList reads one space-separated list of lab/versions.env, so the tests hold the real lists.
func versionsList(t *testing.T, name string) string {
	t.Helper()
	m := regexp.MustCompile(`(?m)^` + name + `=(.*)$`).FindStringSubmatch(read(t, "lab", "versions.env"))
	if m == nil {
		t.Fatalf("lab/versions.env has no %s", name)
	}
	return m[1]
}

// ---- the questions and the settings ------------------------------------------------------------------------------

// The owner's 2026-10-26 ruling: all four Meterless engines are on by default, in the lab too.
func TestFeaturesNothingIsAskedAndBothArePassedOnWithNoTerminalAndNoSetting(t *testing.T) {
	r := runUp(t, upOpts{})
	noQuestions(t, r.out)
	if !r.reachedDocker() {
		t.Fatalf("the flow stopped before docker:\n%s", r.out)
	}
	for _, name := range []string{"PROVEN_MARKOVIAN", "PROVEN_LAB_SCOUT_INTENT"} {
		if !strings.Contains(r.dockerEnv, name+"=on\n") {
			t.Errorf("the lab did not pass %s=on with nothing set:\n%s", name, r.dockerEnv)
		}
	}
	// Mutated field: the lab's default. With it gone the same assertion must fail.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", `: "${PROVEN_MARKOVIAN:=on}" "${PROVEN_LAB_SCOUT_INTENT:=on}"`, `:`)
	if m := runUp(t, upOpts{tree: tree}); strings.Contains(m.dockerEnv, "PROVEN_MARKOVIAN=on\n") {
		t.Error("the control did not turn red: with the default removed the lab still passed PROVEN_MARKOVIAN=on")
	}
}

func TestFeaturesEachSettingReachesTheLabInSampleModeToo(t *testing.T) {
	env := []string{"PROVEN_LAB_MODE=sample", "PROVEN_MARKOVIAN=off", "PROVEN_LAB_SCOUT_INTENT=off"}
	check := func(r upResult) string {
		noQuestions(t, r.out)
		for _, w := range []string{"PROVEN_MARKOVIAN=off\n", "PROVEN_LAB_SCOUT_INTENT=off\n"} {
			if !strings.Contains(r.dockerEnv, w) {
				return "the lab did not start with " + w + " it started with: " + r.dockerEnv
			}
		}
		return ""
	}
	if msg := check(runUp(t, upOpts{env: env})); msg != "" {
		t.Error(msg)
	}
	// The negative control. Mutated field: the exemption of the features in flow_drop_keys, so sample mode drops them
	// as it drops a key. The same assertion must fail.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", `*" $fd_name "*) continue ;;`, `*" $fd_name "*) : ;;`)
	if msg := check(runUp(t, upOpts{tree: tree, env: env})); msg == "" {
		t.Error("the control did not turn red: with the exemption removed sample mode still kept the settings")
	}
}

func TestFeaturesAWrongValueIsRefusedByName(t *testing.T) {
	for _, c := range []struct{ name, value, want string }{
		{"PROVEN_MARKOVIAN", "maybe", "up: PROVEN_MARKOVIAN=maybe is not one of: on, off."},
		{"PROVEN_LAB_SCOUT_INTENT", "1", "up: PROVEN_LAB_SCOUT_INTENT=1 is not one of: on, off."},
	} {
		r := runUp(t, upOpts{env: []string{c.name + "=" + c.value}})
		if r.err == nil || r.reachedDocker() || !strings.Contains(r.out, c.want) {
			t.Errorf("%s=%s was not refused by name before docker:\n%s", c.name, c.value, r.out)
		}
	}
}

func TestFeaturesTheQuestionsAreAskedOnATerminalAndEachAnswerIsTheSetting(t *testing.T) {
	// y on tweak, n for Scout Intent, y for Markovian.
	answers := ttyAnswers(t, "sample", "y", "n", "y")
	r := runUp(t, upOpts{tty: answers})
	for _, w := range []string{
		"Do you want to tweak Meterless features? [y/N]",
		"H-MEM (memory of refuted findings) and World Model (claims, evidence and verdicts as a graph) are on by default.",
		"Scout Intent: plans each review before the agents run (on by default; its benefit is not measured yet). Use it?",
		"Markovian: reviews a large change in bounded steps (on by default). Use it?",
	} {
		if !strings.Contains(r.out, w) {
			t.Errorf("the output lacks %q:\n%s", w, r.out)
		}
	}
	if !strings.Contains(r.dockerEnv, "PROVEN_LAB_SCOUT_INTENT=off\n") || !strings.Contains(r.dockerEnv, "PROVEN_MARKOVIAN=on\n") {
		t.Errorf("the answers n and y were not passed as off and on:\n%s", r.dockerEnv)
	}
	// Mutated field: the mapping of n. The same assertion must fail.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", `n) eval "$ff_var=off" ;;`, `n) eval "$ff_var=on" ;;`)
	m := runUp(t, upOpts{tree: tree, tty: ttyAnswers(t, "sample", "y", "n", "y")})
	if strings.Contains(m.dockerEnv, "PROVEN_LAB_SCOUT_INTENT=off\n") {
		t.Error("the control did not turn red: with n mapped to on the lab still got off")
	}
}

func TestFeaturesEnterAndNKeepTheDefaultOn(t *testing.T) {
	for name, lines := range map[string][]string{
		"N to the first question": {"sample", "n"},
		"Enter to the first":      {"sample", ""},
		"Enter to both features":  {"sample", "y", "", ""},
	} {
		r := runUp(t, upOpts{tty: ttyAnswers(t, lines...)})
		if !r.reachedDocker() {
			t.Errorf("%s: the flow stopped before docker:\n%s", name, r.out)
			continue
		}
		if !strings.Contains(r.dockerEnv, "PROVEN_MARKOVIAN=on\n") || !strings.Contains(r.dockerEnv, "PROVEN_LAB_SCOUT_INTENT=on\n") {
			t.Errorf("%s: the default on was not passed for both:\n%s", name, r.dockerEnv)
		}
	}
}

func TestFeaturesASettingGivenInAdvanceIsNotAskedAgain(t *testing.T) {
	r := runUp(t, upOpts{tty: ttyAnswers(t, "sample", "y", "n"), env: []string{"PROVEN_MARKOVIAN=on"}})
	if strings.Contains(r.out, "Markovian: reviews") {
		t.Errorf("Markovian was asked though PROVEN_MARKOVIAN was set:\n%s", r.out)
	}
	if !strings.Contains(r.out, "Scout Intent: plans") || !strings.Contains(r.dockerEnv, "PROVEN_MARKOVIAN=on\n") || !strings.Contains(r.dockerEnv, "PROVEN_LAB_SCOUT_INTENT=off\n") {
		t.Errorf("the one open feature was not asked, or a value was lost:\n%s\n%s", r.out, r.dockerEnv)
	}
	both := runUp(t, upOpts{tty: ttyAnswers(t, "sample"), env: []string{"PROVEN_MARKOVIAN=on", "PROVEN_LAB_SCOUT_INTENT=on"}})
	if strings.Contains(both.out, "tweak Meterless features") || !both.reachedDocker() {
		t.Errorf("the tweak question was asked though both settings were given:\n%s", both.out)
	}
}

func TestFeaturesAnInputThatEndsAtAFeatureQuestionIsAStopByName(t *testing.T) {
	r := runUp(t, upOpts{tty: ttyAnswers(t, "sample", "y")})
	if r.err == nil || r.reachedDocker() || !strings.Contains(r.out, "up: the input ended before an answer came. NEXT — answer in advance with PROVEN_LAB_SCOUT_INTENT.") {
		t.Errorf("the end of input at the Scout Intent question was not a stop that names its setting:\n%s", r.out)
	}
}

// ---- the page, from what the seed stored -------------------------------------------------------------------------

// pageFeatures stores the settings the way the seed does, against a stub of Gitea, renders the start page and returns
// the features line the page shows, and the names the seed stored.
func pageFeatures(t *testing.T, tree string, env ...string) (line string, stored []string) {
	t.Helper()
	needTool(t, "curl")
	srv, calls, mu := stubGitea(t, http.StatusCreated)
	dir := t.TempDir()
	rows := filepath.Join(dir, "rows.html")
	if err := os.WriteFile(rows, []byte("    <tr><td>1</td></tr>\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	pageDir := filepath.Join(dir, "page")
	if err := os.Mkdir(pageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; . "$ROOT/lab/seed/render.sh"; ensure_live_secrets >/dev/null; render_page; cat "$PAGE_DIR/lab.html"`)
	cmd.Env = cleanEnv(t, dir, append([]string{
		"ROOT=" + tree, "GITEA_INTERNAL_URL=" + srv.URL, "LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "API_TOKEN=x",
		"LAB_LIVE_KEYS=" + versionsList(t, "LAB_LIVE_KEYS"), "LAB_FEATURE_KEYS=" + versionsList(t, "LAB_FEATURE_KEYS"),
		"PAGE_TEMPLATE=" + filepath.Join(tree, "lab", "page", "lab.html.tmpl"), "PAGE_DIR=" + pageDir, "PR_ROWS_FILE=" + rows,
		"CLEAN_COUNT=2", "LAB_URL=http://127.0.0.1:3300/", "LAB_PASSWORD_PLACEHOLDER=x", "LAB_ADMIN_PASSWORD=pw-0123456789",
	}, env...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render: %v\n%s", err, out)
	}
	m := regexp.MustCompile(`id="features">[^<]*<b>([^<]*)</b>`).FindStringSubmatch(string(out))
	if m == nil {
		t.Fatalf("the start page has no features line:\n%s", out)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, c := range *calls {
		if c.method == http.MethodPut {
			stored = append(stored, c.name)
		}
	}
	return m[1], stored
}

func TestFeaturesThePageSaysWhichAreOnFromWhatWasStored(t *testing.T) {
	tree := mustAbs(t, root)
	line, stored := pageFeatures(t, tree, "PROVEN_LAB_SCOUT_INTENT=off", "PROVEN_MARKOVIAN=on")
	want := "H-MEM on (always), World Model on (always), Scout Intent off, Markovian on"
	if line != want {
		t.Errorf("the page says %q, want %q", line, want)
	}
	if strings.Join(stored, " ") != "PROVEN_MARKOVIAN PROVEN_LAB_SCOUT_INTENT" {
		t.Errorf("the seed stored %v", stored)
	}
	// With nothing set, the page says Proven's own default stands, and nothing was stored.
	line, stored = pageFeatures(t, tree)
	if line != "H-MEM on (always), World Model on (always), Scout Intent Proven's default, Markovian Proven's default" || len(stored) != 0 {
		t.Errorf("with no setting the page says %q and %v was stored", line, stored)
	}
	// A feature is not a key: with only features set the lab is still a sample lab.
	// Mutated field: the page line ignores the Scout Intent setting.
	mut := copyLab(t)
	copyInto(t, mut, "lab/seed", "lab/page")
	mutate(t, mut, "lab/seed/golive.sh", `scout=$(feature_word "${PROVEN_LAB_SCOUT_INTENT:-}")`, `scout=$(feature_word "")`)
	if line, _ := pageFeatures(t, mut, "PROVEN_LAB_SCOUT_INTENT=off"); strings.Contains(line, "Scout Intent off") {
		t.Error("the control did not turn red: with the setting ignored the page still said off")
	}
}

func TestFeaturesTheSeedNamesWhatItStoredAndNeverAValue(t *testing.T) {
	needTool(t, "curl")
	srv, _, _ := stubGitea(t, http.StatusCreated)
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets`)
	cmd.Env = cleanEnv(t, t.TempDir(), "ROOT="+mustAbs(t, root), "GITEA_INTERNAL_URL="+srv.URL, "LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount",
		"API_TOKEN=x", "LAB_LIVE_KEYS="+versionsList(t, "LAB_LIVE_KEYS"), "LAB_FEATURE_KEYS="+versionsList(t, "LAB_FEATURE_KEYS"),
		"PROVEN_MARKOVIAN=off", "PROVEN_LAB_SCOUT_INTENT=on")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "seed: stored the Meterless feature settings: PROVEN_MARKOVIAN, PROVEN_LAB_SCOUT_INTENT\n") {
		t.Errorf("the seed did not name the stored feature settings: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "seed: stored as Actions secrets") {
		t.Errorf("a feature was listed as a stored key:\n%s", out)
	}
}

func TestFeaturesAFeatureAloneDoesNotMakeTheLabLive(t *testing.T) {
	needTool(t, "curl")
	srv, _, _ := stubGitea(t, http.StatusCreated)
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets >/dev/null; printf 'state=%s\n' "$LAB_LIVE_STATE"`)
	cmd.Env = cleanEnv(t, t.TempDir(), "ROOT="+mustAbs(t, root), "GITEA_INTERNAL_URL="+srv.URL, "LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount",
		"API_TOKEN=x", "LAB_LIVE_KEYS="+versionsList(t, "LAB_LIVE_KEYS"), "LAB_FEATURE_KEYS="+versionsList(t, "LAB_FEATURE_KEYS"), "PROVEN_MARKOVIAN=off")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "state=no key is stored; reviews are sample reviews") {
		t.Errorf("a feature setting alone changed the live state: %v\n%s", err, out)
	}
}

// ---- the review is given the settings ----------------------------------------------------------------------------

// richDocker stands in for docker and for the container: it keeps its arguments, then runs the script it was given with
// only the variables named by -e, in a path of the test's own, on a stand-in `proven` that records the file named by
// PROVEN_CONFIG and the settings it sees.
const richDocker = `#!/bin/sh
printf '%s\n' "$*" > "$FAKE_DIR/docker.args"
envs=""
script=""
while [ $# -gt 0 ]; do
  case "$1" in
    -e) shift; case "$1" in *=*) envs="$envs $1" ;; *) envs="$envs $1=$(printenv "$1")" ;; esac ;;
    -c) shift; script=$1 ;;
  esac
  shift
done
script=$(printf '%s' "$script" | sed "s|/tmp/proven-run|$FAKE_DIR/run|g")
env -i PATH="$PATH" FAKE_DIR="$FAKE_DIR" $envs sh -c "$script"
`

const richProven = `#!/bin/sh
env | grep -E '^(PROVEN_MARKOVIAN|PROVEN_CONFIG|SCOUT_INTENT)=' > "$FAKE_DIR/proven.env"
if [ -n "$PROVEN_CONFIG" ]; then cat "$PROVEN_CONFIG" > "$FAKE_DIR/proven.config"; fi
printf 'review class: standard\nfailed services: none\n'
`

// reviewSees runs the workflow's review step from the given text and returns what the review process saw: its
// proven.yaml (empty when none was given) and the settings in its environment. An unset secret is an empty variable.
func reviewSees(t *testing.T, wf string, env ...string) (config, seen string) {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", richDocker)
	writeExec(t, bin, "proven", richProven)
	work := filepath.Join(tmp, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-e", "-c", workflowStepFrom(t, wf, "Run the review"))
	cmd.Dir = work
	cmd.Env = append([]string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "FAKE_DIR=" + tmp, "HEAD_REF=demo-discount",
		"GATEWAY_ZETARIS_PATH=/zetaris/api/v1.0", "GATEWAY_ZETARIS_ORG_ID=0"}, env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the review step: %v\n%s", err, out)
	}
	c, _ := os.ReadFile(filepath.Join(tmp, "proven.config"))
	s, _ := os.ReadFile(filepath.Join(tmp, "proven.env"))
	return string(c), string(s)
}

func TestFeaturesTheReviewIsGivenEachSettingAsProvensOwn(t *testing.T) {
	wf := read(t, "lab", "workflow", "proven.yml")
	// Sample mode: Scout Intent is written to the file the review is pointed at; Markovian is a live-only setting.
	config, seen := reviewSees(t, wf, "PROVEN_LAB_SCOUT_INTENT=off")
	if strings.TrimSpace(config) != "scout_intent: off" || !strings.Contains(seen, "PROVEN_CONFIG=") {
		t.Errorf("the review's proven.yaml is %q (environment %q), want scout_intent: off", config, seen)
	}
	config, _ = reviewSees(t, wf, "PROVEN_LAB_SCOUT_INTENT=on")
	if strings.TrimSpace(config) != "scout_intent: on" {
		t.Errorf("the review's proven.yaml is %q, want scout_intent: on", config)
	}
	// Live: PROVEN_MARKOVIAN reaches the review by name, with its value.
	_, seen = reviewSees(t, wf, "PROVEN_GENERATOR_KEY=nvapi-testkey-0123456789", "PROVEN_MARKOVIAN=off")
	if !strings.Contains(seen, "PROVEN_MARKOVIAN=off\n") {
		t.Errorf("the live review did not see PROVEN_MARKOVIAN=off: %q", seen)
	}
	// Nothing set in the workflow's environment: it adds no value of its own (make up passes the lab's defaults).
	config, seen = reviewSees(t, wf, "PROVEN_GENERATOR_KEY=nvapi-testkey-0123456789")
	if config != "" || strings.Contains(seen, "PROVEN_MARKOVIAN") || strings.Contains(seen, "PROVEN_CONFIG") {
		t.Errorf("with no setting the review was given config %q and %q", config, seen)
	}
	// Mutated field: the value handed to the review for Scout Intent. Showing the same assertion red.
	mut := strings.Replace(wf, `-e SCOUT_INTENT="$PROVEN_LAB_SCOUT_INTENT"`, `-e SCOUT_INTENT=`, 1)
	if mut == wf {
		t.Fatal("the mutation target is not in the workflow")
	}
	if config, _ := reviewSees(t, mut, "PROVEN_LAB_SCOUT_INTENT=off"); strings.TrimSpace(config) == "scout_intent: off" {
		t.Error("the control did not turn red: with the value dropped the review still got scout_intent: off")
	}
	mutM := strings.Replace(wf, `live_env="$live_env -e PROVEN_MARKOVIAN"`, `:`, 1)
	if mutM == wf {
		t.Fatal("the mutation target is not in the workflow")
	}
	if _, seen := reviewSees(t, mutM, "PROVEN_GENERATOR_KEY=nvapi-testkey-0123456789", "PROVEN_MARKOVIAN=off"); strings.Contains(seen, "PROVEN_MARKOVIAN=off\n") {
		t.Error("the control did not turn red: with Markovian left out of the passed names the review still saw it")
	}
}

func TestFeaturesAScoutIntentValueThatIsNotOnOrOffStopsTheStep(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", richDocker)
	cmd := exec.Command("sh", "-e", "-c", workflowStep(t, "Run the review"))
	cmd.Dir = tmp
	cmd.Env = []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "FAKE_DIR=" + tmp, "HEAD_REF=demo-discount", "PROVEN_LAB_SCOUT_INTENT=on\nx: y"}
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "is not on or off") {
		t.Errorf("a value that is not on or off reached the container's proven.yaml:\n%s", out)
	}
}

// ---- livecheck judges readiness with the workflow's own derivation ------------------------------------------------

// fakeLivecheck answers as the real checker does for the settings it is given: the generator by its mode (nightshift
// when none is given) and the key that mode needs, the guardrail by its mode, the query by Zetaris's address.
const fakeLivecheck = `#!/bin/sh
mode=${PROVEN_GENERATOR_MODE:-nightshift}
bad=0
if [ "$mode" = nvidia ]; then
  if [ -n "$PROVEN_GENERATOR_KEY" ]; then echo "generator: ready: mode nvidia"; else echo "generator: not ready: PROVEN_GENERATOR_KEY is empty"; bad=1; fi
else
  if [ -n "$METERLESS_RUNTIME_TOKEN" ]; then echo "generator: ready: mode nightshift"; else echo "generator: not ready: nightshift: METERLESS_RUNTIME_TOKEN is empty"; bad=1; fi
fi
case "${PROVEN_GUARDRAIL_MODE:-}" in
  off) echo "guardrail: ready: off by setting" ;;
  nvidia) if [ -n "$PROVEN_GUARDRAIL_KEY" ]; then echo "guardrail: ready: verdict safe"; else echo "guardrail: not ready: PROVEN_GUARDRAIL_KEY is empty"; bad=1; fi ;;
  *) echo "guardrail: not ready: PROVEN_GUARDRAIL_KEY is empty"; bad=1 ;;
esac
if [ -n "$ZETARIS_REST_URL" ] && [ -n "$ZETARIS_API_KEY" ]; then echo "query: ready: rows 1"; else echo "query: not ready: ZETARIS_REST_URL is empty"; bad=1; fi
echo "container: ready"
[ "$bad" -eq 0 ] && exit 0
exit 1
`

func livecheckRun(t *testing.T, tree, labEnv string) (out string, rc int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(tree, "lab", ".env"), []byte(labEnv), 0o600); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	writeExec(t, tmp, "checker", fakeLivecheck)
	cmd := exec.Command("sh", filepath.Join(tree, "lab", "scripts", "livecheck.sh"))
	cmd.Env = cleanEnv(t, tmp, "LIVECHECK_BIN="+filepath.Join(tmp, "checker"))
	b, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(b), ee.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(b), 0
}

// livecheckPathA is the one assertion of the control: with only the NVIDIA key in lab/.env, livecheck names the
// generator and the guardrail ready, exactly as the workflow will run them, and does not fail for Zetaris.
func livecheckPathA(t *testing.T, tree string) string {
	out, rc := livecheckRun(t, tree, "PROVEN_GENERATOR_KEY=nvapi-test-0123456789\n")
	for _, w := range []string{"the workflow will run: live: NVIDIA direct, Zetaris off", "generator: ready", "guardrail: ready", "remark: the query line is not needed"} {
		if !strings.Contains(out, w) {
			return fmt.Sprintf("the output lacks %q:\n%s", w, out)
		}
	}
	if rc != 0 {
		return fmt.Sprintf("exit %d, want 0:\n%s", rc, out)
	}
	return ""
}

func livecheckTree(t *testing.T) string {
	t.Helper()
	tree := copyLab(t)
	copyInto(t, tree, "lab/workflow")
	return tree
}

func TestLivecheckJudgesPathAWithTheWorkflowsOwnDerivation(t *testing.T) {
	if msg := livecheckPathA(t, livecheckTree(t)); msg != "" {
		t.Error(msg)
	}
	// The negative control. Mutated field: livecheck reads its own defaults instead of running the workflow's lines.
	mut := livecheckTree(t)
	mutate(t, mut, "lab/scripts/livecheck.sh", `eval "$derive"`, `sample=false; zetaris=off; state=own`)
	if msg := livecheckPathA(t, mut); msg == "" {
		t.Error("the control did not turn red: with its own defaults livecheck still agreed with the workflow")
	}
}

func TestLivecheckOtherSetupsAgreeWithTheWorkflow(t *testing.T) {
	// No key: the workflow runs a sample review, so there is no live service to ask, and the exit says not live.
	out, rc := livecheckRun(t, livecheckTree(t), "")
	if rc != 1 || !strings.Contains(out, "runs a SAMPLE review") || strings.Contains(out, "generator:") {
		t.Errorf("no key: exit %d\n%s", rc, out)
	}
	// Nightshift token alone: the generator is Nightshift. The guardrail is the workflow's default, which needs its
	// own key, so the check fails on it and says so: the same state a live review would be in.
	out, rc = livecheckRun(t, livecheckTree(t), "METERLESS_RUNTIME_TOKEN=ns-0123456789\n")
	if !strings.Contains(out, "generator: ready: mode nightshift") || !strings.Contains(out, "guardrail: not ready") || rc != 1 {
		t.Errorf("Nightshift token: exit %d\n%s", rc, out)
	}
	// With Zetaris set, the query line counts: a failing one fails the check.
	out, rc = livecheckRun(t, livecheckTree(t), "PROVEN_GENERATOR_KEY=nvapi-test-0123456789\nZETARIS_REST_URL=http://z.example.test/api\nZETARIS_API_KEY=zk-0123456789\nZETARIS_ORG_ID=7\n")
	if rc != 0 || !strings.Contains(out, "the workflow will run: live: NVIDIA direct, Zetaris on") || strings.Contains(out, "remark:") {
		t.Errorf("Zetaris on: exit %d\n%s", rc, out)
	}
}

func TestLivecheckRefusesWhenTheWorkflowHasNoDerivationBlock(t *testing.T) {
	tree := livecheckTree(t)
	mutate(t, tree, "lab/workflow/proven.yml", "# derive-begin", "# not-the-marker")
	out, rc := livecheckRun(t, tree, "PROVEN_GENERATOR_KEY=nvapi-test-0123456789\n")
	if rc == 0 || !strings.Contains(out, "has no derive-begin ... derive-end block") {
		t.Errorf("a workflow without the block was read as a sample or a live setup: exit %d\n%s", rc, out)
	}
}

// ---- Zetaris and Nightshift are present only when they answer in their own shape ---------------------------------

func probeHost(t *testing.T, tree, kind, url string) bool {
	t.Helper()
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/scripts/flow.sh"; flow_probe "$KIND" "$URL"`)
	cmd.Env = cleanEnv(t, t.TempDir(), "ROOT="+tree, "KIND="+kind, "URL="+url)
	return cmd.Run() == nil
}

func serve(t *testing.T, code int, ctype, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(code)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestProbeCountsAServiceOnlyWhenItAnswersInItsOwnShape(t *testing.T) {
	needTool(t, "curl")
	tree := mustAbs(t, root)
	stray404 := serve(t, 404, "text/plain", "404 page not found")
	strayHTML := serve(t, 200, "text/html", "<html><body>It works</body></html>")
	strayHTMLData := serve(t, 200, "text/html", `<html><body><pre>{"data":[]}</pre></body></html>`)
	strayJSON401 := serve(t, 401, "text/plain", "Unauthorized")
	nsList := serve(t, 200, "application/json", `{"object":"list","data":[{"id":"meterless/auto"}]}`)
	ns401 := serve(t, 401, "application/json", `{"error":"unauthorized"}`)
	zJSON := serve(t, 401, "application/json", `{"error":"unauthorized"}`)
	for _, c := range []struct {
		name, kind, url string
		want            bool
	}{
		{"a stray server answering 404 is not Nightshift", "nightshift", stray404, false},
		{"a stray server answering 200 HTML is not Nightshift", "nightshift", strayHTML, false},
		{"an HTML page quoting a model list is not Nightshift", "nightshift", strayHTMLData, false},
		{"a stray 401 with no JSON is not Nightshift", "nightshift", strayJSON401, false},
		{"nothing listening is not Nightshift", "nightshift", "http://127.0.0.1:1/", false},
		{"a JSON list of models is Nightshift", "nightshift", nsList, true},
		{"a 401 JSON from its runtime is Nightshift", "nightshift", ns401, true},
		{"a stray server answering 404 is not Zetaris", "zetaris", stray404, false},
		{"a stray server answering 200 HTML is not Zetaris", "zetaris", strayHTML, false},
		{"a JSON root is Zetaris", "zetaris", zJSON, true},
	} {
		if got := probeHost(t, tree, c.kind, c.url); got != c.want {
			t.Errorf("%s: flow_probe says %v", c.name, got)
		}
	}
	// The negative control. Mutated field: the shape test, so any status that is not 000 counts, as before round 4.
	mut := copyLab(t)
	mutate(t, mut, "lab/scripts/flow.sh", "  return 1\n}\n\nflow_detect_own()", "  [ \"$fq_code\" != 000 ] && return 0\n  return 1\n}\n\nflow_detect_own()")
	if !probeHost(t, mut, "nightshift", stray404) {
		t.Error("the control did not turn red: with the shape test removed a stray 404 was still not taken for Nightshift")
	}
}

// ---- the end of input at the gateway address ----------------------------------------------------------------------

// emptyTTY is a terminal whose input has already ended: a file with no bytes, not a file with one empty line.
func emptyTTY(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tty")
	if err := os.WriteFile(p, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFlowAnInputThatEndsAtTheGatewayAddressSaysSo(t *testing.T) {
	env := liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "PROVEN_MARKOVIAN=off", "PROVEN_LAB_SCOUT_INTENT=off")
	check := func(r upResult) string {
		if r.err == nil || r.reachedDocker() {
			return "the flow did not stop:\n" + r.out
		}
		if !strings.Contains(r.out, "up: the input ended before the gateway address came. NEXT — put it in lab/.env as JUDGES_GATEWAY_URL.") {
			return "the stop does not say the input ended:\n" + r.out
		}
		return ""
	}
	if msg := check(runUp(t, upOpts{tty: emptyTTY(t), env: env})); msg != "" {
		t.Error(msg)
	}
	// Mutated field: the end-of-input test of the address prompt. The same assertion must fail.
	mut := copyLab(t)
	mutate(t, mut, "lab/scripts/flow.sh", `if [ "$fg_rc" -ne 0 ] && [ -z "$JUDGES_GATEWAY_URL" ]; then`, `if false; then`)
	if msg := check(runUp(t, upOpts{tree: mut, tty: emptyTTY(t), env: env})); msg == "" {
		t.Error("the control did not turn red: with the end-of-input test removed the stop still said the input ended")
	}
}

// ---- one review per head commit ---------------------------------------------------------------------------------

func TestWorkflowOnlyThePullRequestEventStartsItAndRunsOfOneHeadQueue(t *testing.T) {
	wf := read(t, "lab", "workflow", "proven.yml")
	on := wf[strings.Index(wf, "\non:\n"):]
	on = on[:strings.Index(on, "\npermissions:")]
	if strings.Contains(on, "push:") || strings.Contains(on, "pull_request_target") || strings.Count(on, "pull_request:") != 1 {
		t.Errorf("the workflow has a trigger besides pull_request:%s", on)
	}
	if !regexp.MustCompile(`types: \[opened, synchronize, reopened\]`).MatchString(on) {
		t.Errorf("the pull_request types are not named:%s", on)
	}
	if !regexp.MustCompile(`(?m)^concurrency:\n  group: proven-review-\$\{\{ github\.event\.pull_request\.number \}\}-\$\{\{ github\.event\.pull_request\.head\.sha \}\}\n  cancel-in-progress: false$`).MatchString(wf) {
		t.Error("runs of one head commit are not grouped by pull request and head sha")
	}
}

// fakeWgetSeen answers the lookup of the pull request's comments with FAKE_COMMENTS, and records any post.
const fakeWgetSeen = `#!/bin/sh
out=""
post=""
while [ $# -gt 0 ]; do case "$1" in --post-file=*) post="${1#--post-file=}" ;; -O) shift; out="$1" ;; esac; shift; done
if [ -n "$post" ]; then cp "$post" "$FAKE_DIR/posted.json"; printf '{"id":1}' > "$out"; exit 0; fi
printf '%s' "$FAKE_COMMENTS" > "$out"
`

// runBothSteps runs the review step and the post step of a workflow text for one head commit, against a pull request
// whose comments are given, and returns whether the review ran and whether a comment was posted.
func runBothSteps(t *testing.T, wf, comments string) (reviewed, posted bool) {
	t.Helper()
	needTool(t, "awk")
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", "#!/bin/sh\necho ran > \"$FAKE_DIR/docker.ran\"\nprintf 'review class: standard\\nfailed services: none\\n'\n")
	writeExec(t, bin, "wget", fakeWgetSeen)
	work := filepath.Join(tmp, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}
	env := []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "FAKE_DIR=" + tmp, "FAKE_COMMENTS=" + comments, "HEAD_REF=demo-discount",
		"PROVEN_TOKEN=tok-0123456789", "API_URL=http://gitea.test/api/v1", "REPO=proven/discount", "PR=7", "HEAD_SHA=0123456789abcdef0123456789abcdef01234567", "RUN_NUMBER=19"}
	for _, step := range []string{"Run the review", "Post the review on the pull request"} {
		cmd := exec.Command("sh", "-e", "-c", workflowStepFrom(t, wf, step))
		cmd.Dir = work
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", step, err, out)
		}
	}
	_, e1 := os.Stat(filepath.Join(tmp, "docker.ran"))
	_, e2 := os.Stat(filepath.Join(tmp, "posted.json"))
	return e1 == nil, e2 == nil
}

func TestWorkflowAHeadThatIsAlreadyReviewedIsNotReviewedOrCommentedAgain(t *testing.T) {
	wf := read(t, "lab", "workflow", "proven.yml")
	const sha = "0123456789abcdef0123456789abcdef01234567"
	already := `[{"id":1,"body":"\u003c!-- proven-review run=18 sha=` + sha + ` --\u003e\nlive: NVIDIA direct, Zetaris off\n"}]`
	otherHead := `[{"id":1,"body":"\u003c!-- proven-review run=18 sha=fedcba9876543210fedcba9876543210fedcba98 --\u003e\nlive\n"}]`
	assert := func(wf string) string {
		if reviewed, posted := runBothSteps(t, wf, already); reviewed || posted {
			return fmt.Sprintf("a head with a review was reviewed=%v posted=%v again", reviewed, posted)
		}
		if reviewed, posted := runBothSteps(t, wf, "[]"); !reviewed || !posted {
			return fmt.Sprintf("a head with no review was reviewed=%v posted=%v", reviewed, posted)
		}
		if reviewed, posted := runBothSteps(t, wf, otherHead); !reviewed || !posted {
			return fmt.Sprintf("a review of another head stopped this one: reviewed=%v posted=%v", reviewed, posted)
		}
		return ""
	}
	if msg := assert(wf); msg != "" {
		t.Error(msg)
	}
	// Mutated field: the head commit the guard looks for. The same assertion must fail.
	mut := strings.Replace(wf, `grep -q "proven-review run=[0-9]* sha=$HEAD_SHA" seen.json; then`, `grep -q "proven-review run=[0-9]* sha=none" seen.json; then`, 1)
	if mut == wf {
		t.Fatal("the mutation target is not in the workflow")
	}
	if msg := assert(mut); msg == "" {
		t.Error("the control did not turn red: with the guard looking for another head the duplicate was still skipped")
	}
}
