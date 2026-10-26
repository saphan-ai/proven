package lab

// Tests of the start-up flow (lab/scripts/flow.sh) through the whole of lab/scripts/up.sh, with a stand-in `docker`
// that records every call and the lab's settings it was handed, and then fails. The flow runs before any container,
// so a refusal leaves no docker call, and a path that was taken shows in the settings the first call received.
// Each run is a session of its own (no controlling terminal), so nothing here can read a developer's terminal.
//
// A guard is tested by mutating exactly the thing it guards in a copy of the scripts: the same assertion on the copy
// must fail, and the test says which field it changed.

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const (
	goodToken = "judge-token-0123456789"
	badToken  = "stale-token-0123456789"
)

type upResult struct {
	out, calls, dockerEnv string
	err                   error
}

func (r upResult) reachedDocker() bool { return r.calls != "" }

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return fmt.Sprint(l.Addr().(*net.TCPAddr).Port)
}

// labPorts gives every lab port a number nothing holds, so a lab already running on this host cannot refuse a test.
func labPorts(t *testing.T) []string {
	t.Helper()
	return []string{"GITEA_HTTP_PORT=" + freePort(t), "LGTM_GRAFANA_PORT=" + freePort(t), "LGTM_OTLP_PORT=" + freePort(t),
		"LAB_PAGE_PORT=" + freePort(t), "LAB_DOCS_PORT=" + freePort(t)}
}

func writeExec(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

// fakeDocker records its arguments, and the settings of the lab that reached it, then fails: nothing is started.
const fakeDocker = `#!/bin/sh
case "$1" in ps) echo "$*" >> "$FAKE_DIR/docker.ps"; exit 0 ;; esac
echo "$*" >> "$FAKE_DIR/docker.calls"
env | grep -E '^(LAB_MODE|JUDGES_GATEWAY_[A-Z_]+|ZETARIS|ZETARIS_REST_URL|PROVEN_GENERATOR_[A-Z_]+|METERLESS_[A-Z_]+)=' >> "$FAKE_DIR/docker.env"
exit 97
`

// copyLab copies what up.sh reads into a temporary tree, so a test can change one line of it.
func copyLab(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	for _, p := range []string{"lab/scripts", "lab/versions.env", "lab/runner/config.yaml", "go.mod"} {
		src := filepath.Join(root, p)
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			if info.IsDir() {
				return os.MkdirAll(filepath.Join(dst, rel), 0o755)
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(dst, rel)), 0o755); err != nil {
				return err
			}
			b, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			return os.WriteFile(filepath.Join(dst, rel), b, info.Mode())
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// mutate changes exactly one thing in a copy of the lab and fails the test when the thing is not there to change,
// so a mutation can never pass by changing nothing.
func mutate(t *testing.T, tree, rel, old, repl string) {
	t.Helper()
	p := filepath.Join(tree, rel)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), old) {
		t.Fatalf("the mutation target %q is not in %s", old, rel)
	}
	if err := os.WriteFile(p, []byte(strings.Replace(string(b), old, repl, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

type upOpts struct {
	tree   string // "" is the real tree
	env    []string
	bins   map[string]string // name -> script, placed first on the path
	tty    string            // a file or FIFO standing for the terminal; "" means no terminal
	stdin  io.Reader
	during func(cmd *exec.Cmd)
}

// runUp runs lab/scripts/up.sh in a clean environment of its own, as a session leader with no controlling terminal.
func runUp(t *testing.T, o upOpts) upResult {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", fakeDocker)
	for name, body := range o.bins {
		writeExec(t, bin, name, body)
	}
	tree := o.tree
	if tree == "" {
		tree = mustAbs(t, root)
	}
	cmd := exec.Command("sh", filepath.Join(tree, "lab", "scripts", "up.sh"))
	env := append([]string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp, "FAKE_DIR=" + tmp}, labPorts(t)...)
	if o.tty != "" {
		env = append(env, "PROVEN_LAB_TTY="+o.tty)
	}
	cmd.Env = append(env, o.env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = o.stdin
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if o.during != nil {
		o.during(cmd)
	}
	err := cmd.Wait()
	calls, _ := os.ReadFile(filepath.Join(tmp, "docker.calls"))
	de, _ := os.ReadFile(filepath.Join(tmp, "docker.env"))
	return upResult{out: out.String(), calls: string(calls), dockerEnv: string(de), err: err}
}

// stubGateway answers the way the judges' gateway is specified to: 400 not_a_template_request to credentials it
// accepts, 401 to others. It records the path of each request and whether the right header carried the credential.
func stubGateway(t *testing.T, delay time.Duration, header, want string) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		got := r.Header.Get(header)
		mu.Lock()
		seen = append(seen, fmt.Sprintf("%s %s credential-ok=%v", r.Method, r.URL.Path, got == want))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if got == want {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"not_a_template_request"}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"token_expired"}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

func liveOurs(extra ...string) []string {
	return append([]string{"PROVEN_LAB_MODE=live", "PROVEN_LAB_JUDGE=yes", "PROVEN_LAB_BACKEND=ours"}, extra...)
}

func fakeJWT(exp int64) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte(fmt.Sprintf(`{"aud":["x"],"exp":%d}`, exp))) + ".c2lnbmF0dXJl"
}

func noRealCloudflared(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("cloudflared"); err == nil {
		t.Skip("a real cloudflared is on this host's path; this test needs its absence")
	}
}

var questions = []string{"[sample/live]", "[y/n]", "[ours/own]", "Do you want to see Proven in SAMPLE mode"}

func noQuestions(t *testing.T, out string) {
	t.Helper()
	for _, q := range questions {
		if strings.Contains(out, q) {
			t.Errorf("a question was printed with no terminal: %q\n%s", q, out)
		}
	}
}

func TestFlowWithNoTerminalAndNoSettingIsSampleAndAsksNothing(t *testing.T) {
	// The mutated field is the default of the mode: a lab with a key in the environment and no setting.
	r := runUp(t, upOpts{env: []string{"PROVEN_GENERATOR_KEY=nvapi-0123456789"}})
	noQuestions(t, r.out)
	if !strings.Contains(r.out, "up: sample mode") {
		t.Errorf("no sample mode line:\n%s", r.out)
	}
	if !strings.Contains(r.dockerEnv, "LAB_MODE=sample\n") {
		t.Errorf("the lab was not started in sample mode: %q", r.dockerEnv)
	}
	if strings.Contains(r.dockerEnv, "PROVEN_GENERATOR_KEY") || !strings.Contains(r.out, "sample mode does not use these settings: PROVEN_GENERATOR_KEY") {
		t.Errorf("sample mode kept a key, or did not say it dropped one:\n%s\n%s", r.out, r.dockerEnv)
	}
	if !r.reachedDocker() {
		t.Errorf("the flow stopped before docker:\n%s", r.out)
	}
	// The negative control: the default with no terminal and no setting becomes live. Mutated field: that default.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", "      fl_mode=sample\n", "      fl_mode=live\n")
	m := runUp(t, upOpts{tree: tree})
	if strings.Contains(m.out, "up: sample mode") || strings.Contains(m.dockerEnv, "LAB_MODE=sample\n") {
		t.Error("the control did not turn red: with the default changed the lab still started in sample mode")
	}
}

func TestFlowEachPathRunsFromItsSettingsAndPrintsNoQuestion(t *testing.T) {
	jwt := fakeJWT(1893456000) // 2030-01-01 00:00 UTC
	tokenGW, _ := stubGateway(t, 0, "Authorization", "Bearer "+goodToken)
	accessGW, accessSeen := stubGateway(t, 0, "cf-access-token", jwt)
	zet := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) }))
	t.Cleanup(zet.Close)
	nsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"data":[]}`) }))
	t.Cleanup(nsrv.Close)
	cf := "#!/bin/sh\ncase \"$1 $2\" in\n\"access login\") echo \"Successfully fetched your token:\"; echo LOGINPRINTEDTOKEN; exit 0;;\n\"access token\") [ \"$3\" = \"-app=$STUB_APP\" ] || exit 3; echo \"$STUB_ACCESS_JWT\"; exit 0;;\nesac\nexit 2\n"
	zhost := strings.TrimPrefix(zet.URL, "http://")
	nhost := strings.TrimPrefix(nsrv.URL, "http://")
	_, zport, _ := net.SplitHostPort(zhost)
	_, nport, _ := net.SplitHostPort(nhost)
	cases := []struct {
		name     string
		env      []string
		bins     map[string]string
		wantOut  []string
		wantEnv  []string
		wantDock bool
	}{
		{"sample", []string{"PROVEN_LAB_MODE=sample"}, nil, []string{"up: sample mode"}, []string{"LAB_MODE=sample\n"}, true},
		{"live, judge, ours, by token", liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL="+tokenGW.URL, "JUDGES_GATEWAY_TOKEN="+goodToken), nil,
			[]string{"the gateway accepted your credentials: ready.", "live mode: live via our gateway."}, []string{"LAB_MODE=live via our gateway\n", "JUDGES_GATEWAY_TOKEN=" + goodToken + "\n"}, true},
		{"live, judge, ours, by Access sign-in", liveOurs("JUDGES_GATEWAY_URL="+accessGW.URL, "STUB_ACCESS_JWT="+jwt, "STUB_APP="+accessGW.URL), map[string]string{"cloudflared": cf},
			[]string{"the gateway accepted your credentials: ready.", "The token expires 2030-01-01 00:00 UTC", "run make up again"},
			[]string{"LAB_MODE=live via our gateway\n", "JUDGES_GATEWAY_ACCESS_TOKEN=" + jwt + "\n"}, true},
		{"live, judge, own, no automation", []string{"PROVEN_LAB_MODE=live", "PROVEN_LAB_JUDGE=yes", "PROVEN_LAB_BACKEND=own", "PROVEN_LAB_OWN_SETUP=no", "DETECT_ZETARIS_URL=http://127.0.0.1:1/x", "DETECT_NIGHTSHIFT_URL=http://127.0.0.1:1/y"}, nil,
			[]string{"docs/LAB.md", "We cannot vouch for results on a setup we do not control: this is best effort.", "meterless.ai"}, []string{"LAB_MODE=live on your own setup\n"}, true},
		{"live, not a judge", []string{"PROVEN_LAB_MODE=live", "PROVEN_LAB_JUDGE=no", "PROVEN_LAB_OWN_SETUP=no", "DETECT_ZETARIS_URL=http://127.0.0.1:1/x", "DETECT_NIGHTSHIFT_URL=http://127.0.0.1:1/y"}, nil,
			[]string{"this is best effort."}, []string{"LAB_MODE=live on your own setup\n"}, true},
		{"live, own, what already answers is used", []string{"PROVEN_LAB_MODE=live", "PROVEN_LAB_JUDGE=no", "PROVEN_LAB_OWN_SETUP=no", "PROVEN_LAB_USE_DETECTED=yes",
			"DETECT_ZETARIS_URL=" + zet.URL + "/api/v1.0", "DETECT_NIGHTSHIFT_URL=" + nsrv.URL + "/v1/models"}, nil,
			[]string{"Zetaris already answers on this machine", "Nightshift already answers on this machine", "keys stay yours"},
			[]string{"ZETARIS_REST_URL=http://host.docker.internal:" + zport + "/api/v1.0\n", "PROVEN_GENERATOR_BASE_URL=http://host.docker.internal:" + nport + "\n"}, true},
		{"live, own, the guide's names", []string{"PROVEN_LAB_MODE=live", "PROVEN_LAB_JUDGE=no", "PROVEN_LAB_OWN_SETUP=no", "METERLESS_BASE_URL=http://nightshift.example.test:8088", "METERLESS_MODEL=meterless/auto",
			"DETECT_ZETARIS_URL=http://127.0.0.1:1/x", "DETECT_NIGHTSHIFT_URL=http://127.0.0.1:1/y"}, nil, nil,
			[]string{"PROVEN_GENERATOR_BASE_URL=http://nightshift.example.test:8088\n", "PROVEN_GENERATOR_MODEL=meterless/auto\n"}, true},
	}
	if runtime.GOOS != "darwin" {
		cases = append(cases, struct {
			name     string
			env      []string
			bins     map[string]string
			wantOut  []string
			wantEnv  []string
			wantDock bool
		}{"live, own, the automated setup", []string{"PROVEN_LAB_MODE=live", "PROVEN_LAB_JUDGE=no", "PROVEN_LAB_OWN_SETUP=yes", "DETECT_ZETARIS_URL=http://127.0.0.1:1/x", "DETECT_NIGHTSHIFT_URL=http://127.0.0.1:1/y"}, nil,
			[]string{"running the automated setup", "Tart runs Nightshift's macOS VM"}, []string{"ZETARIS=local\n"}, true})
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := runUp(t, upOpts{env: c.env, bins: c.bins})
			noQuestions(t, r.out)
			for _, w := range c.wantOut {
				if !strings.Contains(r.out, w) {
					t.Errorf("the output lacks %q:\n%s", w, r.out)
				}
			}
			for _, w := range c.wantEnv {
				if !strings.Contains(r.dockerEnv, w) {
					t.Errorf("the lab did not start with %q; it started with:\n%s", w, r.dockerEnv)
				}
			}
			if r.reachedDocker() != c.wantDock {
				t.Errorf("reached docker = %v, want %v:\n%s", r.reachedDocker(), c.wantDock, r.out)
			}
			for _, secret := range []string{goodToken, jwt, "LOGINPRINTEDTOKEN"} {
				if strings.Contains(r.out, secret) {
					t.Errorf("a credential is in the output")
				}
			}
		})
	}
	if len(*accessSeen) != 1 || !strings.HasSuffix((*accessSeen)[0], "credential-ok=true") {
		t.Errorf("the Access path must check the gateway exactly once, with the Access token in cf-access-token: %v", *accessSeen)
	}
}

func TestFlowChecksTheGatewayOnceAndNamesARefusal(t *testing.T) {
	// The mutated field is the credential: one byte of the token differs, and the gateway answers 401.
	gw, seen := stubGateway(t, 0, "Authorization", "Bearer "+goodToken)
	r := runUp(t, upOpts{env: liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL="+gw.URL, "JUDGES_GATEWAY_TOKEN="+badToken)})
	if r.err == nil || r.reachedDocker() {
		t.Errorf("a refused credential must stop before any container (err=%v, docker=%v):\n%s", r.err, r.reachedDocker(), r.out)
	}
	if !strings.Contains(r.out, "the gateway refused your credentials: HTTP 401 (token_expired)") {
		t.Errorf("the refusal is not named:\n%s", r.out)
	}
	if strings.Contains(r.out, badToken) || strings.Contains(r.out, goodToken) {
		t.Errorf("a token is in the output")
	}
	if len(*seen) != 1 {
		t.Errorf("the gateway was asked %d times, want once: %v", len(*seen), *seen)
	}
	// A gateway nobody answers at is named too, and so is the missing address.
	r = runUp(t, upOpts{env: liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL=http://127.0.0.1:1", "JUDGES_GATEWAY_TOKEN="+goodToken)})
	if r.err == nil || !strings.Contains(r.out, "the gateway did not answer") || r.reachedDocker() {
		t.Errorf("an unreachable gateway is not named:\n%s", r.out)
	}
	r = runUp(t, upOpts{env: liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_TOKEN="+goodToken)})
	if r.err == nil || !strings.Contains(r.out, "the judges' gateway address is not set") || r.reachedDocker() {
		t.Errorf("a missing address is not named:\n%s", r.out)
	}
	r = runUp(t, upOpts{env: liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL="+gw.URL)})
	if r.err == nil || !strings.Contains(r.out, "the judge token is not set and there is no terminal") || r.reachedDocker() {
		t.Errorf("a missing token with no terminal is not named:\n%s", r.out)
	}
}

func TestFlowAccessPathWithoutCloudflaredSaysHowToInstallItAndStops(t *testing.T) {
	noRealCloudflared(t)
	r := runUp(t, upOpts{env: liveOurs("JUDGES_GATEWAY_URL=https://gw.example.test")})
	if r.err == nil || r.reachedDocker() {
		t.Errorf("must stop before any container (err=%v):\n%s", r.err, r.out)
	}
	for _, w := range []string{"cloudflared", "brew install cloudflared", "PROVEN_LAB_GATEWAY_AUTH=token", "Nothing was started"} {
		if !strings.Contains(r.out, w) {
			t.Errorf("the message lacks %q:\n%s", w, r.out)
		}
	}
}

// refusesBadSetting is the assertion the negative control for the setting check runs twice.
func refusesBadSetting(t *testing.T, tree string) bool {
	t.Helper()
	r := runUp(t, upOpts{tree: tree, env: []string{"PROVEN_LAB_MODE=maybe"}})
	return r.err != nil && !r.reachedDocker() && strings.Contains(r.out, "up: PROVEN_LAB_MODE=maybe is not one of: sample, live.")
}

func TestFlowRefusesABadSettingByNameBeforeAnyContainer(t *testing.T) {
	for _, bad := range []string{"PROVEN_LAB_MODE=maybe", "PROVEN_LAB_JUDGE=maybe", "PROVEN_LAB_BACKEND=mine", "PROVEN_LAB_GATEWAY_AUTH=password",
		"PROVEN_LAB_OWN_SETUP=sure", "PROVEN_LAB_USE_DETECTED=y"} {
		name := bad[:strings.Index(bad, "=")]
		r := runUp(t, upOpts{env: []string{bad}})
		if r.err == nil || r.reachedDocker() || !strings.Contains(r.out, "up: "+bad+" is not one of: ") || !strings.Contains(r.out, "Nothing was started") {
			t.Errorf("%s was not refused by name before any container (docker=%v):\n%s", name, r.reachedDocker(), r.out)
		}
	}
	if !refusesBadSetting(t, "") {
		t.Fatal("the real flow does not refuse a bad mode")
	}
	// The negative control: flow_check_enum accepts any value. Mutated field: the comparison of the value to the list.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", `[ "$fe_val" = "$fe_ok" ] && return 0`, `return 0`)
	if refusesBadSetting(t, tree) {
		t.Error("the control did not turn red: the mutated flow still refused a bad mode")
	}
}

// refusesTakenPort is the assertion the negative control for the port check runs twice.
func refusesTakenPort(t *testing.T, tree string) (bool, string) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	port := fmt.Sprint(l.Addr().(*net.TCPAddr).Port)
	r := runUp(t, upOpts{tree: tree, env: []string{"PROVEN_LAB_MODE=sample", "LAB_PAGE_PORT=" + port}})
	ok := r.err != nil && !r.reachedDocker() && strings.Contains(r.out, "up: port "+port+" (LAB_PAGE_PORT) is already in use by ")
	return ok, r.out
}

func TestFlowRefusesATakenPortByNameBeforeAnyContainer(t *testing.T) {
	needTool(t, "ss")
	ok, out := refusesTakenPort(t, "")
	if !ok {
		t.Fatalf("a taken port was not refused by name before docker:\n%s", out)
	}
	if !strings.Contains(out, "Nothing was started") {
		t.Errorf("the refusal does not say nothing was started:\n%s", out)
	}
	if strings.Contains(out, "is already in use by a program") && !strings.Contains(out, "users:") {
		t.Logf("this host did not name the holder, so the line says a program: %s", out)
	}
	// Two taken ports are both named, in one refusal.
	a, _ := net.Listen("tcp", "127.0.0.1:0")
	b, _ := net.Listen("tcp", "127.0.0.1:0")
	defer a.Close()
	defer b.Close()
	pa, pb := fmt.Sprint(a.Addr().(*net.TCPAddr).Port), fmt.Sprint(b.Addr().(*net.TCPAddr).Port)
	r := runUp(t, upOpts{env: []string{"PROVEN_LAB_MODE=sample", "GITEA_HTTP_PORT=" + pa, "LGTM_OTLP_PORT=" + pb}})
	if !strings.Contains(r.out, "port "+pa+" (GITEA_HTTP_PORT)") || !strings.Contains(r.out, "port "+pb+" (LGTM_OTLP_PORT)") || r.reachedDocker() {
		t.Errorf("two taken ports were not both named:\n%s", r.out)
	}
	// A port setting that is not a number is refused by name.
	r = runUp(t, upOpts{env: []string{"PROVEN_LAB_MODE=sample", "LAB_DOCS_PORT=http"}})
	if r.err == nil || !strings.Contains(r.out, "LAB_DOCS_PORT=http is not a port number") || r.reachedDocker() {
		t.Errorf("a bad port number was not refused by name:\n%s", r.out)
	}
	// The negative control: the flow no longer checks the ports. Mutated field: the call that checks them.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", "  flow_check_lab_ports\n", "")
	if ok, _ := refusesTakenPort(t, tree); ok {
		t.Error("the control did not turn red: the flow with no port check still refused a taken port")
	}
}

// procArgv samples every process's command line while fn runs, and returns each line that holds needle.
func procArgv(needle string, stop <-chan struct{}, hits chan<- string) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		entries, _ := filepath.Glob("/proc/[0-9]*/cmdline")
		for _, e := range entries {
			b, err := os.ReadFile(e)
			if err == nil && strings.Contains(string(b), needle) {
				select {
				case hits <- strings.ReplaceAll(string(b), "\x00", " "):
				default:
				}
			}
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// credentialReach runs the token path against a slow gateway and returns the output, every file the token reached
// under the run's own directories, and the command lines that held it while the check was in flight.
func credentialReach(t *testing.T, tree string) (out string, files []string, argv []string) {
	t.Helper()
	gw, _ := stubGateway(t, 600*time.Millisecond, "Authorization", "Bearer "+goodToken)
	stop := make(chan struct{})
	hits := make(chan string, 64)
	done := make(chan struct{})
	go func() {
		procArgv(goodToken, stop, hits)
		close(done)
	}()
	r := runUp(t, upOpts{tree: tree, env: liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL="+gw.URL, "JUDGES_GATEWAY_TOKEN="+goodToken)})
	close(stop)
	<-done
	close(hits)
	for h := range hits {
		argv = append(argv, h)
	}
	// Every file of the lab, but not the tests, which hold the token as a constant.
	for _, f := range walkFor(t, filepath.Join(mustAbs(t, root), "lab"), goodToken) {
		if !strings.HasSuffix(f, "_test.go") {
			files = append(files, f)
		}
	}
	return r.out, files, argv
}

func TestFlowTheCredentialIsInNoOutputNoFileAndNoCommandLine(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the command lines are sampled from /proc, which only Linux has; the master's live run samples them with ps on the other hosts")
	}
	out, files, argv := credentialReach(t, "")
	if !strings.Contains(out, "ready") {
		t.Fatalf("the path did not reach the check:\n%s", out)
	}
	if strings.Contains(out, goodToken) {
		t.Error("the token is in the output")
	}
	for _, f := range files {
		t.Errorf("the token is in a file of the lab: %s", f)
	}
	if len(argv) > 0 {
		t.Errorf("the token was on a command line: %v", argv)
	}
	// The negative controls, one field each. (1) The success line prints the token. (2) The credential goes to curl
	// as an argument instead of through the pipe.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", `echo "up: the gateway accepted your credentials: ready."`, `echo "up: the gateway accepted $JUDGES_GATEWAY_TOKEN: ready."`)
	if out, _, _ := credentialReach(t, tree); !strings.Contains(out, goodToken) {
		t.Error("the control did not turn red: the token printed on purpose was not seen in the output")
	}
	tree = copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", "curl -sS --max-time", `curl -sS -H "$FLOW_CRED_HEADER: $FLOW_CRED_VALUE" --max-time`)
	if _, _, argv := credentialReach(t, tree); len(argv) == 0 {
		t.Error("the control did not turn red: a token placed on curl's command line was not seen")
	}
}

// ttyAnswers makes a file that stands for the terminal, holding the answers one per line.
func ttyAnswers(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tty")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFlowOnATerminalPrintsTheOwnersBannerAndQuestionsWordForWord(t *testing.T) {
	r := runUp(t, upOpts{tty: ttyAnswers(t, "maybe", "SAMPLE")})
	for _, w := range []string{
		"Proven: a pull-request review companion that only says what it can prove (and admits what it could not check)",
		"Team Saphan.ai: built in public by one person and a fleet of AI agents.",
		"Do you want to see Proven in SAMPLE mode (no keys, nothing leaves this machine) or LIVE mode (real models and live data)? [sample/live]",
		"That is not one of: sample live. Again, please.",
		"up: sample mode",
	} {
		if !strings.Contains(r.out, w) {
			t.Errorf("the output lacks %q:\n%s", w, r.out)
		}
	}
	if strings.Index(r.out, "Proven: a pull-request") > strings.Index(r.out, "Do you want to see") {
		t.Error("the banner does not come first")
	}
	if !strings.Contains(r.dockerEnv, "LAB_MODE=sample\n") {
		t.Errorf("the lab did not start in sample mode: %q", r.dockerEnv)
	}
	// The judge's questions, in the owner's words, then both answers.
	gw, _ := stubGateway(t, 0, "Authorization", "Bearer "+goodToken)
	r = runUp(t, upOpts{tty: ttyAnswers(t, "live", "Y", "ours", goodToken), env: []string{"PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL=" + gw.URL}})
	for _, w := range []string{
		"Are you a judge of the Open Agent Hackathon? [y/n]",
		"For Proven to prove anything live, it needs Zetaris and Meterless Nightshift. Connect your own instances, or use the ones Proven was built with, in Hua Hin, Thailand, under Marcin's desk.",
		"[ours/own]",
		"ready.",
	} {
		if !strings.Contains(r.out, w) {
			t.Errorf("the output lacks %q:\n%s", w, r.out)
		}
	}
	if !strings.Contains(r.dockerEnv, "LAB_MODE=live via our gateway\n") || strings.Contains(r.out, goodToken) {
		t.Errorf("the judge's path went wrong, or the token was printed:\n%s\n%s", r.out, r.dockerEnv)
	}
}

func TestFlowAnswersGivenInAdvanceAreNotAskedAgain(t *testing.T) {
	r := runUp(t, upOpts{tty: ttyAnswers(t), env: []string{"PROVEN_LAB_MODE=sample"}})
	if strings.Contains(r.out, "[sample/live]") || !strings.Contains(r.out, "up: sample mode") {
		t.Errorf("a question was asked though a setting answered it:\n%s", r.out)
	}
}

func TestFlowStopsByNameAfterThreeWrongAnswersAndWhenTheInputEnds(t *testing.T) {
	r := runUp(t, upOpts{tty: ttyAnswers(t, "maybe", "perhaps", "nope", "sample")})
	if r.err == nil || r.reachedDocker() || !strings.Contains(r.out, "up: three answers in a row were not one of: sample, live. NEXT — answer in advance with PROVEN_LAB_MODE.") {
		t.Errorf("three wrong answers did not stop by name:\n%s", r.out)
	}
	if strings.Count(r.out, "[sample/live]") != 3 {
		t.Errorf("the question was asked %d times, want exactly 3", strings.Count(r.out, "[sample/live]"))
	}
	empty := filepath.Join(t.TempDir(), "tty")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	r = runUp(t, upOpts{tty: empty})
	if r.err == nil || r.reachedDocker() || !strings.Contains(r.out, "the input ended before an answer came") {
		t.Errorf("an input that ended did not stop by name:\n%s", r.out)
	}
}

// A stand-in stty logs each call. `stty -g` answers with a fixed string, as a terminal's saved state.
const fakeStty = `#!/bin/sh
echo "$*" >> "$FAKE_DIR/stty.log"
case "$1" in -g) echo SAVEDSTATE ;; esac
exit 0
`

// sttyAfterInterrupt starts the token prompt, interrupts it while it waits, and returns the calls stty saw.
func sttyAfterInterrupt(t *testing.T, tree string) []string {
	t.Helper()
	fifo := filepath.Join(t.TempDir(), "tty")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("this host makes no FIFO: %v", err)
	}
	logdir := t.TempDir()
	var w *os.File
	r := runUp(t, upOpts{tree: tree, tty: fifo, bins: map[string]string{"stty": fakeStty},
		env: liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL=http://127.0.0.1:1", "FAKE_DIR="+logdir),
		during: func(cmd *exec.Cmd) {
			var err error
			// Opening the write end releases the flow's open of the read end; the flow then waits in read.
			w, err = os.OpenFile(fifo, os.O_WRONLY, 0)
			if err != nil {
				t.Error(err)
				return
			}
			// The prompt is printed before the read; the stty -echo call is logged just before it. Wait for it.
			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				b, _ := os.ReadFile(filepath.Join(logdir, "stty.log"))
				if strings.Contains(string(b), "-echo") {
					break
				}
				time.Sleep(20 * time.Millisecond)
			}
			time.Sleep(100 * time.Millisecond)
			_ = cmd.Process.Signal(syscall.SIGINT)
		}})
	if w != nil {
		w.Close()
	}
	_ = r
	b, _ := os.ReadFile(filepath.Join(logdir, "stty.log"))
	return strings.Fields(strings.ReplaceAll(string(b), "\n", " \n "))
}

func TestFlowRestoresTheTerminalWhenTheTokenPromptIsInterrupted(t *testing.T) {
	log := strings.Join(sttyAfterInterrupt(t, ""), " ")
	iEcho := strings.Index(log, "-echo")
	iBack := strings.LastIndex(log, "SAVEDSTATE")
	if iEcho < 0 {
		t.Fatalf("echo was never turned off, so the prompt was never reached: %q", log)
	}
	if iBack < iEcho {
		t.Errorf("the saved terminal state was not restored after echo went off: %q", log)
	}
	// The negative control: the interrupt no longer restores the terminal. Mutated field: the INT trap of the prompt.
	tree := copyLab(t)
	mutate(t, tree, "lab/scripts/flow.sh", "trap 'flow_restore_echo; echo; exit 130' INT", "trap - INT")
	log = strings.Join(sttyAfterInterrupt(t, tree), " ")
	if strings.Index(log, "-echo") < 0 {
		t.Fatalf("the mutated run never reached the prompt: %q", log)
	}
	if strings.LastIndex(log, "SAVEDSTATE") > strings.Index(log, "-echo") {
		t.Error("the control did not turn red: with the INT trap removed the terminal was still restored")
	}
}
