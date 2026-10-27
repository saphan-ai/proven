package lab

// These tests hold the lab's go-live wiring together without Docker, a network or a key (docs/LAB.md, "Go live").
// Their names carry GoLive so `make check-golive` runs exactly these. A test that needs a tool this host lacks says
// so with t.Skip and the tool's name; it is never a silent pass.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

// cleanEnv is the environment of a script under test: a path, a home and a private temp directory, and what the test
// adds. Nothing else of the test runner's environment (a real key in the developer's shell) can reach the script.
func cleanEnv(t *testing.T, tmp string, extra ...string) []string {
	t.Helper()
	return append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp}, extra...)
}

// secretsFile writes the credentials file make up hands the seed (NAME=value lines, mode 0600) and returns its path.
func secretsFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "secrets")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func needTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is not on this host's path; this test needs it", name)
	}
}

// walkFor returns every file under dir whose bytes contain needle.
func walkFor(t *testing.T, dir, needle string) []string {
	t.Helper()
	var hits []string
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr == nil && strings.Contains(string(b), needle) {
			hits = append(hits, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hits
}

type secretCall struct{ method, name, auth, data string }

// stubGitea records the calls to the repository's Actions secrets, as Gitea's API takes them.
func stubGitea(t *testing.T, wantStatus int) (*httptest.Server, *[]secretCall, *sync.Mutex) {
	t.Helper()
	var mu sync.Mutex
	var calls []secretCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/api/v1/repos/proven/discount/actions/secrets/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var parsed struct{ Data string }
		_ = json.Unmarshal(body, &parsed)
		mu.Lock()
		calls = append(calls, secretCall{r.Method, strings.TrimPrefix(r.URL.Path, prefix), r.Header.Get("Authorization"), parsed.Data})
		mu.Unlock()
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound) // there was none to remove
			return
		}
		w.WriteHeader(wantStatus)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls, &mu
}

func TestGoLiveEachKeyBecomesASecretAndAnEmptyOneIsNotStored(t *testing.T) {
	needTool(t, "curl")
	srv, calls, mu := stubGitea(t, http.StatusCreated)
	tmp := t.TempDir()
	// Values with the characters a JSON string and a shell both care about.
	generator := `nvapi-aaaa"quote\back&amp|pipe-1234567890`
	token := "tok-0123456789abcdef"
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets; printf 'state=%s\n' "$LAB_LIVE_STATE"`)
	// The credentials arrive in the state volume's file, as make up writes it: NAME=value lines. ZETARIS_API_KEY is set
	// and empty there: not stored.
	secrets := filepath.Join(t.TempDir(), "secrets")
	if err := os.WriteFile(secrets, []byte("PROVEN_GENERATOR_KEY="+generator+"\nMETERLESS_RUNTIME_TOKEN="+token+"\nZETARIS_API_KEY=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd.Env = cleanEnv(t, tmp,
		"ROOT="+mustAbs(t, root),
		"GITEA_INTERNAL_URL="+srv.URL,
		"LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "API_TOKEN=lab-api-token",
		"LAB_LIVE_KEYS=PROVEN_GENERATOR_KEY METERLESS_RUNTIME_TOKEN ZETARIS_API_KEY",
		"LAB_SECRET_KEYS=PROVEN_GENERATOR_KEY METERLESS_RUNTIME_TOKEN ZETARIS_API_KEY",
		"LAB_SECRETS_FILE="+secrets,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ensure_live_secrets: %v\n%s", err, out)
	}
	mu.Lock()
	defer mu.Unlock()
	got := map[string]secretCall{}
	for _, c := range *calls {
		got[c.method+" "+c.name] = c
	}
	if c, ok := got["PUT PROVEN_GENERATOR_KEY"]; !ok || c.data != generator {
		t.Errorf("the generator key was not stored intact: %+v", c)
	}
	if c, ok := got["PUT METERLESS_RUNTIME_TOKEN"]; !ok || c.data != token {
		t.Errorf("the Nightshift token was not stored: %+v", c)
	}
	if _, ok := got["PUT ZETARIS_API_KEY"]; ok {
		t.Error("an empty key was stored")
	}
	if _, ok := got["DELETE ZETARIS_API_KEY"]; !ok {
		t.Error("an empty key was not removed from the repository, so a key taken out of lab/.env would stay live")
	}
	for _, c := range *calls {
		if c.auth != "token lab-api-token" {
			t.Errorf("a call was made without the lab's API token: %+v", c)
		}
	}
	// Nothing prints a value: the log carries names.
	for _, v := range []string{generator, token} {
		if strings.Contains(string(out), v) {
			t.Error("the output carries a key's value")
		}
	}
	if !strings.Contains(string(out), "state=keys stored: PROVEN_GENERATOR_KEY, METERLESS_RUNTIME_TOKEN") {
		t.Errorf("the page state does not name the stored keys:\n%s", out)
	}
	// Nothing is written to a file: the script's temp directory holds no value.
	for _, v := range []string{generator, token} {
		if hits := walkFor(t, tmp, v); len(hits) > 0 {
			t.Errorf("a key's value was written to %v", hits)
		}
	}
}

func TestGoLiveNoKeyAtAllLeavesSampleMode(t *testing.T) {
	needTool(t, "curl")
	srv, calls, mu := stubGitea(t, http.StatusCreated)
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets; printf 'state=%s\n' "$LAB_LIVE_STATE"`)
	cmd.Env = cleanEnv(t, t.TempDir(),
		"ROOT="+mustAbs(t, root), "GITEA_INTERNAL_URL="+srv.URL,
		"LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "API_TOKEN=x",
		"LAB_LIVE_KEYS=PROVEN_GENERATOR_KEY METERLESS_RUNTIME_TOKEN")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, c := range *calls {
		if c.method == http.MethodPut {
			t.Errorf("a secret was stored with no key set: %+v", c)
		}
	}
	if !strings.Contains(string(out), "state=no key is stored; reviews are sample reviews") {
		t.Errorf("the page state is wrong with no key:\n%s", out)
	}
}

func TestGoLiveARefusedStoreStopsTheSeedByName(t *testing.T) {
	needTool(t, "curl")
	srv, _, _ := stubGitea(t, http.StatusForbidden)
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets`)
	cmd.Env = cleanEnv(t, t.TempDir(),
		"ROOT="+mustAbs(t, root), "GITEA_INTERNAL_URL="+srv.URL,
		"LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "API_TOKEN=x",
		"LAB_LIVE_KEYS=PROVEN_GENERATOR_KEY", "LAB_SECRET_KEYS=PROVEN_GENERATOR_KEY", "LAB_SECRETS_FILE="+secretsFile(t, "PROVEN_GENERATOR_KEY=secret-value-12345\n"))
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("a refused store did not stop the seed")
	}
	if !strings.Contains(string(out), "could not store the secret PROVEN_GENERATOR_KEY (HTTP 403)") {
		t.Errorf("the refusal does not name the secret and the status:\n%s", out)
	}
	if strings.Contains(string(out), "secret-value-12345") {
		t.Error("the refusal prints the value")
	}
}

// workflowStep returns the script of the step of lab/workflow/proven.yml whose name is name.
func workflowStep(t *testing.T, name string) string {
	t.Helper()
	return workflowStepFrom(t, read(t, "lab", "workflow", "proven.yml"), name)
}

// workflowStepFrom is workflowStep on a workflow text the caller holds, so a test can mutate the text first.
func workflowStepFrom(t *testing.T, text, name string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	var out []string
	in, collecting := false, false
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "      - name: "):
			in = strings.TrimPrefix(l, "      - name: ") == name
			collecting = false
		case in && strings.HasPrefix(l, "        run: |"):
			collecting = true
		case in && collecting:
			if l != "" && !strings.HasPrefix(l, "          ") {
				collecting = false
				continue
			}
			out = append(out, strings.TrimPrefix(l, "          "))
		}
	}
	if len(out) == 0 {
		t.Fatalf("lab/workflow/proven.yml has no step %q with a run script", name)
	}
	return strings.Join(out, "\n")
}

// runWorkflowReview runs the first step with a stand-in `docker` that records its arguments and the variables it
// received, and prints a review. It returns mode.txt, the recorded arguments and the recorded environment.
func runWorkflowReview(t *testing.T, env ...string) (mode, args, dockerEnv string) {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$FAKE_DIR/docker.args\"\nenv > \"$FAKE_DIR/docker.env\"\nprintf 'review class: standard\\nfailed services: none\\n----- proven: markdown -----\\n**Proven: clean**\\n'\n"
	if err := os.WriteFile(filepath.Join(bin, "docker"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(tmp, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}
	script := workflowStep(t, "Run the review")
	cmd := exec.Command("sh", "-e", "-c", script)
	cmd.Dir = work
	cmd.Env = append([]string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "FAKE_DIR=" + tmp,
		"HEAD_REF=demo-discount", "GATEWAY_ZETARIS_PATH=/zetaris/api/v1.0", "GATEWAY_ZETARIS_ORG_ID=0"}, env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the review step: %v\n%s", err, out)
	}
	m, err := os.ReadFile(filepath.Join(work, "mode.txt"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(filepath.Join(tmp, "docker.args"))
	e, _ := os.ReadFile(filepath.Join(tmp, "docker.env"))
	return strings.TrimSpace(string(m)), string(a), string(e)
}

func TestGoLiveTheWorkflowRunsSampleWithNoKeyAndLiveWithOne(t *testing.T) {
	const nvKey = "nvapi-testkey-0123456789"
	const nsToken = "ns-token-0123456789"
	const gwToken = "judge-token-0123456789"
	const zKey = "zk-0123456789"
	cases := []struct {
		name      string
		env       []string
		wantMode  string
		wantSamp  string
		wantNames []string
		notNames  []string
		wantEnv   []string
	}{
		{"no key", nil, "sample: no keys", "SAMPLE_MODE=true", nil,
			[]string{"PROVEN_GENERATOR_KEY", "METERLESS_RUNTIME_TOKEN", "ZETARIS_API_KEY"}, nil},
		{"an NVIDIA key alone", []string{"PROVEN_GENERATOR_KEY=" + nvKey}, "live: NVIDIA direct, Zetaris off", "SAMPLE_MODE=false",
			[]string{"PROVEN_GENERATOR_KEY", "PROVEN_GENERATOR_MODE", "PROVEN_GUARDRAIL_MODE", "PROVEN_GUARDRAIL_KEY"}, []string{"METERLESS_RUNTIME_TOKEN"},
			[]string{"PROVEN_GENERATOR_MODE=nvidia", "PROVEN_GUARDRAIL_MODE=nvidia", "PROVEN_GUARDRAIL_KEY=" + nvKey}},
		{"Nightshift and Zetaris", []string{"METERLESS_RUNTIME_TOKEN=" + nsToken, "ZETARIS_REST_URL=http://zetaris-api:8888/api/v1.0", "ZETARIS_API_KEY=" + zKey, "ZETARIS_ORG_ID=7"},
			"live: NVIDIA via Nightshift, Zetaris on", "SAMPLE_MODE=false",
			[]string{"METERLESS_RUNTIME_TOKEN", "ZETARIS_API_KEY", "ZETARIS_REST_URL", "ZETARIS_ORG_ID"}, nil,
			[]string{"PROVEN_GENERATOR_MODE=nightshift"}},
		{"Nightshift without Zetaris", []string{"METERLESS_RUNTIME_TOKEN=" + nsToken}, "live: NVIDIA via Nightshift, Zetaris off", "SAMPLE_MODE=false",
			[]string{"METERLESS_RUNTIME_TOKEN"}, []string{"ZETARIS_API_KEY"}, nil},
		{"the judges' gateway", []string{"JUDGES_GATEWAY_URL=https://gw.example.test/", "JUDGES_GATEWAY_TOKEN=" + gwToken},
			"live: NVIDIA via the judges' gateway, Zetaris on", "SAMPLE_MODE=false",
			[]string{"METERLESS_RUNTIME_TOKEN", "ZETARIS_API_KEY", "ZETARIS_REST_URL", "PROVEN_GENERATOR_BASE_URL"}, []string{"JUDGES_GATEWAY_TOKEN"},
			[]string{"PROVEN_GENERATOR_BASE_URL=https://gw.example.test", "ZETARIS_REST_URL=https://gw.example.test/zetaris/api/v1.0", "METERLESS_RUNTIME_TOKEN=" + gwToken}},
		{"the judges' gateway by an Access sign-in", []string{"JUDGES_GATEWAY_URL=https://gw.example.test/", "JUDGES_GATEWAY_ACCESS_TOKEN=" + gwToken},
			"live: NVIDIA via the judges' gateway (Access sign-in), Zetaris on", "SAMPLE_MODE=false",
			[]string{"JUDGES_GATEWAY_ACCESS_TOKEN", "METERLESS_RUNTIME_TOKEN", "ZETARIS_API_KEY"}, []string{"JUDGES_GATEWAY_TOKEN"},
			[]string{"PROVEN_GENERATOR_BASE_URL=https://gw.example.test", "JUDGES_GATEWAY_ACCESS_TOKEN=" + gwToken}},
		{"a model name from the organisers' guide", []string{"PROVEN_GENERATOR_KEY=" + nvKey, "PROVEN_GENERATOR_MODEL=some/model"}, "live: NVIDIA direct, Zetaris off", "SAMPLE_MODE=false",
			[]string{"PROVEN_GENERATOR_MODEL"}, nil, []string{"PROVEN_GENERATOR_MODEL=some/model"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mode, args, dockerEnv := runWorkflowReview(t, c.env...)
			if mode != c.wantMode {
				t.Errorf("the first line of the comment is %q, want %q", mode, c.wantMode)
			}
			if !strings.Contains(args, c.wantSamp) {
				t.Errorf("the container is not told %s: %s", c.wantSamp, args)
			}
			for _, n := range c.wantNames {
				if !strings.Contains(args, "-e "+n+" ") && !strings.HasSuffix(strings.TrimSpace(args), "-e "+n) {
					t.Errorf("the container is not given %s by name: %s", n, args)
				}
			}
			for _, n := range c.notNames {
				if strings.Contains(args, "-e "+n+" ") {
					t.Errorf("the container is given %s though it is not set: %s", n, args)
				}
			}
			for _, w := range c.wantEnv {
				if !strings.Contains(dockerEnv, w+"\n") {
					t.Errorf("the container's environment lacks %s", w)
				}
			}
			// A value is on no command line, only the name is.
			for _, v := range []string{nvKey, nsToken, gwToken, zKey} {
				if strings.Contains(args, v) {
					t.Errorf("a key's value is on the docker command line: %s", args)
				}
			}
		})
	}
}

func TestGoLiveTheCommentStartsWithTheMarkerThenHowItWasMade(t *testing.T) {
	needTool(t, "awk")
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	// A stand-in wget answers like Gitea does and keeps what was posted.
	fake := "#!/bin/sh\nwhile [ $# -gt 0 ]; do case \"$1\" in --post-file=*) cp \"${1#--post-file=}\" \"$FAKE_DIR/posted.json\" ;; -O) shift; out=\"$1\" ;; esac; shift; done\nprintf '{\"id\":1}' > \"$out\"\n"
	if err := os.WriteFile(filepath.Join(bin, "wget"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "comment-http"), []byte(fakeCommentHTTP), 0o755); err != nil {
		t.Fatal(err)
	}
	// The comment carries the review's Markdown; the text form decides the clean or incomplete line above it.
	for _, tc := range []struct{ mode, review, md, wantSecond, wantThird string }{
		{"sample: no keys", "review class: standard\nfailed services: none\n", "**Proven: clean**\n", "sample: no keys", "clean: Proven kept no finding"},
		{"live: NVIDIA via Nightshift, Zetaris on", "review class: standard\nfinding 1: x\n", "**Proven: blocked** — 1 finding\n", "live: NVIDIA via Nightshift, Zetaris on", "**Proven: blocked**"},
		// No finding while the generator failed is not clean: the line says the review is incomplete.
		{"live: NVIDIA direct, Zetaris off", "review class: standard\nfailed service: nvidia: http 500\n", "**Proven: clean**\n", "live: NVIDIA direct, Zetaris off", "incomplete: Proven kept no finding"},
	} {
		work := filepath.Join(tmp, strings.ReplaceAll(tc.mode, " ", "_"))
		if err := os.Mkdir(work, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range map[string]string{"mode.txt": tc.mode + "\n", "review.txt": tc.review, "review.md": tc.md} {
			if err := os.WriteFile(filepath.Join(work, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command("sh", "-c", workflowStep(t, "Post the review on the pull request"))
		cmd.Dir = work
		cmd.Env = []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "FAKE_DIR=" + tmp,
			"PROVEN_TOKEN=t", "API_URL=http://x/api/v1", "REPO=proven/discount", "PR=1", "RUN_NUMBER=4", "HEAD_SHA=abc123", "COMMENT_HTTP=" + filepath.Join(bin, "comment-http")}
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("the post step: %v\n%s", err, out)
		}
		comment := strings.Split(string(mustRead(t, filepath.Join(work, "comment.txt"))), "\n")
		if comment[0] != "<!-- proven-review run=4 sha=abc123 -->" {
			t.Errorf("the marker is not the first line (e2e reads it there): %q", comment[0])
		}
		if comment[1] != tc.wantSecond {
			t.Errorf("the first visible line is %q, want %q", comment[1], tc.wantSecond)
		}
		// An empty line keeps the mode line its own paragraph when the comment is read as Markdown.
		if comment[2] != "" || !strings.HasPrefix(comment[3], tc.wantThird) {
			t.Errorf("the lines after the mode are %q, %q; want an empty line, then one starting %q", comment[2], comment[3], tc.wantThird)
		}
	}
}

func TestGoLiveEveryNameTheSettingsListIsHeldTogether(t *testing.T) {
	env := read(t, "lab", "versions.env")
	m := regexp.MustCompile(`(?m)^LAB_LIVE_KEYS=(.*)$`).FindStringSubmatch(env)
	if m == nil {
		t.Fatal("lab/versions.env has no LAB_LIVE_KEYS")
	}
	keys := strings.Fields(m[1])
	sort.Strings(keys)
	compose := read(t, "lab", "compose.yaml")
	wf := read(t, "lab", "workflow", "proven.yml")
	example := read(t, "lab", ".env.example")
	secret := strings.Fields(versionsList(t, "LAB_SECRET_KEYS"))
	for _, k := range keys {
		isSecret := false
		for _, s := range secret {
			isSecret = isSecret || s == k
		}
		// A credential is named nowhere in Compose (lab/secrets_test.go); a setting is passed by name.
		if !isSecret && !regexp.MustCompile(`(?m)^      `+k+`:$`).MatchString(compose) {
			t.Errorf("the seed service does not pass %s by name (lab/compose.yaml)", k)
		}
		if !strings.Contains(wf, "secrets."+k+" }}") {
			t.Errorf("the workflow does not read the secret %s", k)
		}
		if !regexp.MustCompile(`(?m)^` + k + `=$`).MatchString(example) {
			t.Errorf("lab/.env.example does not list %s, empty", k)
		}
	}
	// The other way: the seed passes no name the list lacks, so nothing travels unlisted.
	seed := compose[strings.Index(compose, "  seed:"):]
	seed = seed[:strings.Index(seed, "    volumes:")]
	for _, m := range regexp.MustCompile(`(?m)^      ([A-Z][A-Z0-9_]*):$`).FindAllStringSubmatch(seed, -1) {
		found := false
		for _, k := range keys {
			found = found || k == m[1]
		}
		if !found {
			t.Errorf("the seed passes %s by name and LAB_LIVE_KEYS does not list it", m[1])
		}
	}
}

func TestGoLiveTheExampleHoldsNoValue(t *testing.T) {
	for n, line := range strings.Split(read(t, "lab", ".env.example"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasSuffix(line, "=") {
			t.Errorf("lab/.env.example line %d carries a value; a secret or a default belongs in the judge's own lab/.env: %q", n+1, line)
		}
	}
}

func TestGoLiveThePageIsLoopbackOnly(t *testing.T) {
	compose := read(t, "lab", "compose.yaml")
	if !strings.Contains(compose, `"127.0.0.1:${LAB_PAGE_PORT}:${LAB_PAGE_CONTAINER_PORT}"`) {
		t.Error("the page is not published on 127.0.0.1")
	}
	if strings.Contains(compose, "${LAB_PAGE_BIND") || strings.Contains(compose, "${LAB_BIND") {
		t.Error("lab/compose.yaml interpolates an address setting: no setting may move a published port")
	}
	lib := mustAbs(t, filepath.Join(root, "lab", "scripts", "lib.sh"))
	for _, bind := range []string{"127.0.0.1", "::1", "localhost"} {
		cmd := exec.Command("sh", "-c", `ROOT="$1"; export ROOT; . "$2"; lab_require_loopback LAB_PAGE_BIND "$3"`, "sh", mustAbs(t, root), lib, bind)
		cmd.Env = cleanEnv(t, t.TempDir())
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("a loopback bind %s was refused: %v\n%s", bind, err, out)
		}
	}
	// The negative control: change exactly the bind value, and the refusal names it.
	for _, bind := range []string{"0.0.0.0", "192.168.1.20", "::", ""} {
		cmd := exec.Command("sh", "-c", `ROOT="$1"; export ROOT; . "$2"; lab_require_loopback LAB_PAGE_BIND "$3"`, "sh", mustAbs(t, root), lib, bind)
		cmd.Env = cleanEnv(t, t.TempDir())
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("a non-loopback bind %q was accepted", bind)
		}
		if !strings.Contains(string(out), "refused: LAB_PAGE_BIND="+bind+" is not a loopback address") {
			t.Errorf("the refusal for %q does not name the setting and the value: %s", bind, out)
		}
	}
}

// make up refuses a non-loopback page bind BEFORE it starts anything: a stand-in docker records every call.
func TestGoLiveUpRefusesAPageBindThatIsNotLoopbackBeforeItStartsAnything(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := "#!/bin/sh\necho called >> \"$FAKE_DIR/docker.calls\"\nexit 97\n"
	if err := os.WriteFile(filepath.Join(bin, "docker"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(bind string) (string, error, bool) {
		cmd := exec.Command("sh", filepath.Join(mustAbs(t, root), "lab", "scripts", "up.sh"))
		cmd.Env = []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp, "FAKE_DIR=" + tmp, "LAB_PAGE_BIND=" + bind}
		out, err := cmd.CombinedOutput()
		_, statErr := os.Stat(filepath.Join(tmp, "docker.calls"))
		os.Remove(filepath.Join(tmp, "docker.calls"))
		return string(out), err, statErr == nil
	}
	out, err, called := run("0.0.0.0")
	if err == nil || !strings.Contains(out, "refused: LAB_PAGE_BIND=0.0.0.0") {
		t.Errorf("make up did not refuse a bind of 0.0.0.0 by name: %v\n%s", err, out)
	}
	if called {
		t.Error("make up called docker before refusing the bind: the page could be published first")
	}
	out, _, called = run("127.0.0.1")
	if strings.Contains(out, "refused:") || !called {
		t.Errorf("make up with the loopback default was refused or never got to docker (called=%v):\n%s", called, out)
	}
}

// lab/.env is read, never executed: an empty value is left unset, a value already in the environment wins, a line that
// is not NAME=value is refused by its number and never by its content, and a name that would change the shell, the
// loader or docker is refused.
func TestGoLiveTheEnvFileIsReadNotExecuted(t *testing.T) {
	lib := mustAbs(t, filepath.Join(root, "lab", "scripts", "lib.sh"))
	run := func(envBody string, extra ...string) (string, error) {
		tmp := t.TempDir()
		// A copy of the lab layout is not needed: ROOT points at a directory that holds only lab/.env and versions.env.
		if err := os.MkdirAll(filepath.Join(tmp, "lab"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "lab", ".env"), []byte(envBody), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "lab", "versions.env"), []byte("SOMETHING_ELSE=1\nZETARIS_COMPOSE_FILE=\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("sh", "-c", `ROOT="$1"; export ROOT; . "$2"; load_lab_env || exit 9; printf 'gen=%s|zet=%s|file=%s|empty=%s\n' "${PROVEN_GENERATOR_KEY-unset}" "${ZETARIS_API_KEY-unset}" "$ZETARIS_COMPOSE_FILE" "${PROVEN_NVIDIA_BASE_URL-unset}"`, "sh", tmp, lib)
		cmd.Env = cleanEnv(t, tmp, extra...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	out, err := run("# a comment\nPROVEN_GENERATOR_KEY=abc$(touch /tmp/never)def\nZETARIS_API_KEY=\nZETARIS_COMPOSE_FILE=/bundle/docker-compose.yml\n\nPROVEN_NVIDIA_BASE_URL=\n", "ZETARIS_API_KEY=from-environment")
	if err != nil {
		t.Fatalf("a good file was refused: %v\n%s", err, out)
	}
	if !strings.Contains(out, "gen=abc$(touch /tmp/never)def|") {
		t.Errorf("a value was executed or changed: %s", out)
	}
	if !strings.Contains(out, "zet=from-environment|") {
		t.Errorf("the environment did not win over an empty line: %s", out)
	}
	if !strings.Contains(out, "file=/bundle/docker-compose.yml|") {
		t.Errorf("a setting that versions.env leaves empty was not filled from lab/.env: %s", out)
	}
	if !strings.Contains(out, "empty=unset") {
		t.Errorf("an empty value was exported: %s", out)
	}
	for _, bad := range []string{"not a setting\n", "PATH=/tmp/evil\n", "LD_PRELOAD=/tmp/x.so\n", "DOCKER_HOST=tcp://elsewhere\n"} {
		out, err := run("PROVEN_GENERATOR_KEY=ok\n" + bad)
		if err == nil || !strings.Contains(out, "lab/.env: refused line 2") {
			t.Errorf("line %q was not refused by number: %v\n%s", strings.TrimSpace(bad), err, out)
		}
		if strings.Contains(out, "evil") || strings.Contains(out, "elsewhere") {
			t.Errorf("the refusal repeats the content of the line: %s", out)
		}
	}
}

func TestGoLiveNightshiftOnAHostThatIsNotAMacSaysSoAndExitsClean(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("this host is a Mac: make nightshift would start a VM; the Linux answer is not testable here")
	}
	cmd := exec.Command("sh", filepath.Join(mustAbs(t, root), "lab", "scripts", "nightshift.sh"))
	cmd.Env = cleanEnv(t, t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make nightshift exited non-zero on a host without Tart: %v\n%s", err, out)
	}
	for _, want := range []string{"macOS on Apple silicon only", "path A", "path B"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("the answer lacks %q:\n%s", want, out)
		}
	}
}

func TestGoLiveZetarisRunsOnLoopbackAndAsLittleAsTheImageAllows(t *testing.T) {
	override := read(t, "lab", "zetaris", "override.yml")
	for _, line := range strings.Split(override, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, `- "`) && strings.Contains(trim, ":") && !strings.HasPrefix(trim, `- "127.0.0.1:`) {
			t.Errorf("a Zetaris port in the override is not on 127.0.0.1: %s", trim)
		}
	}
	if !strings.Contains(override, "ports: !override") {
		t.Error("the override does not replace the sponsor's published ports")
	}
	compose := read(t, "lab", "compose.yaml")
	svc := compose[strings.Index(compose, "  zetaris-register:"):]
	svc = svc[:strings.Index(svc, "\nnetworks:")]
	for _, want := range []string{`profiles: ["zetaris"]`, "cap_drop:", "read_only: true", "no-new-privileges:true", `user: "65534:65534"`} {
		if !strings.Contains(svc, want) {
			t.Errorf("the zetaris-register service lacks %s", want)
		}
	}
	if strings.Contains(svc, "ports:") {
		t.Error("the zetaris-register service publishes a port")
	}
	// The terms question is answered in the script, in words.
	if !strings.Contains(read(t, "lab", "scripts", "zetaris.sh"), "YOUR sign-in and YOUR bundle") {
		t.Error("zetaris.sh does not say whose sign-in it uses")
	}
}

func TestGoLiveRegisterSendsTheDefinitionsAndTheGrantsAsTheAdministrator(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("register.sh reads /proc/sys/kernel/random/uuid, which only Linux has")
	}
	needTool(t, "curl")
	var mu sync.Mutex
	var sqls []string
	var auth, org string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var q struct{ Sql string }
		_ = json.Unmarshal(body, &q)
		mu.Lock()
		sqls = append(sqls, q.Sql)
		auth, org = r.Header.Get("Authorization"), r.Header.Get("X-Org-ID")
		mu.Unlock()
		if r.URL.Path != "/api/v1.0/sql-editor/sqls/run" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprint(w, `{"msg":"ok"}`)
	}))
	defer srv.Close()
	run := func(stdin string, extra ...string) (string, error) {
		cmd := exec.Command("sh", filepath.Join(mustAbs(t, root), "lab", "zetaris", "register.sh"))
		cmd.Stdin = strings.NewReader(stdin)
		cmd.Env = cleanEnv(t, t.TempDir(), append([]string{
			"ZETARIS_REGISTER_URL=" + srv.URL + "/api/v1.0",
			"ZETARIS_EVIDENCE_SQL=" + mustAbs(t, filepath.Join(root, "lab", "zetaris", "evidence.sql")),
		}, extra...)...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	// The negative control: leave out exactly the administrator's key, and nothing is sent.
	out, err := run("", "ZETARIS_ORG_ID=7", "ZETARIS_PROVEN_USER=proven_user")
	if err == nil || !strings.Contains(out, "no administrator key arrived") {
		t.Errorf("a missing administrator key was not refused by name: %v\n%s", err, out)
	}
	mu.Lock()
	if len(sqls) != 0 {
		t.Error("a request was sent without the administrator's key")
	}
	mu.Unlock()
	out, err = run("admin-key-0123456789", "ZETARIS_ORG_ID=7", "ZETARIS_PROVEN_USER=proven_user")
	if err != nil {
		t.Fatalf("register: %v\n%s", err, out)
	}
	if strings.Contains(out, "admin-key-0123456789") {
		t.Error("register printed the administrator's key")
	}
	mu.Lock()
	defer mu.Unlock()
	if auth != "Bearer admin-key-0123456789" || org != "7" {
		t.Errorf("the calls carry the wrong key or organisation: %q %q", auth, org)
	}
	if len(sqls) < 4 || !strings.HasPrefix(sqls[0], "CREATE LIGHTNING DATABASE") {
		t.Fatalf("the definitions were not sent first and in order: %v", sqls)
	}
	last := sqls[len(sqls)-3:]
	for i, want := range []string{"ASSIGN LIGHTNING DATABASE", "GRANT SELECT ON LIGHTNING DATABASE NAMESPACE", "GRANT SELECT ON VIEW NAMESPACE"} {
		if !strings.HasPrefix(last[i], want) || !strings.HasSuffix(last[i], "TO USER 'proven_user'") {
			t.Errorf("grant %d is %q, want %s ... TO USER 'proven_user'", i, last[i], want)
		}
	}
}
