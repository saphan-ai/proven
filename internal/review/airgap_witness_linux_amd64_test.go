//go:build linux && amd64

package review

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/testrun"
)

// witnessSpec is what the test hands the child that runs the review under the witness.
type witnessSpec struct {
	Gen, Guard, Zetaris, Daemon, Canary int // loopback ports of the services the review may reach
	Home                                string
	Fixture                             string
	Full, Partial                       witnessRepo
}

type witnessRepo struct{ Dir, Base, Head string }

// The child of TestAirGappedWitnessSeesEveryConnection. The filter cannot be taken off, so the review
// runs in this binary started again, the way internal/cli does it, and the test only reads its report.
func TestMain(m *testing.M) {
	if raw := os.Getenv("PROVEN_WITNESS_SPEC"); raw != "" {
		os.Exit(witnessChild(raw))
	}
	os.Exit(m.Run())
}

func witnessChild(raw string) int {
	var spec witnessSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	local := func(port int) string { return net.JoinHostPort("127.0.0.1", strconv.Itoa(port)) }
	allowed := map[string]bool{local(spec.Gen): true, local(spec.Guard): true, local(spec.Zetaris): true, local(spec.Daemon): true, local(spec.Canary): true}
	w, err := installWitness(func(target string) bool { return allowed[target] || strings.HasPrefix(target, "unix:") })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	step := func(name string, err error) {
		if err != nil {
			fmt.Fprintf(out, "step %s FAIL %s\n", name, strings.ReplaceAll(err.Error(), "\n", " "))
			return
		}
		fmt.Fprintf(out, "step %s ok\n", name)
	}
	fmt.Fprintf(out, "pid %d\n", os.Getpid())

	// The container runtime reads its daemon address from the process environment, so the child has it too.
	env := map[string]string{
		"PROVEN_NETWORK":            "local",
		"METERLESS_RUNTIME_TOKEN":   "test-token",
		"PROVEN_GENERATOR_BASE_URL": "http://localhost:" + strconv.Itoa(spec.Gen),
		"PROVEN_GUARDRAIL_BASE_URL": "http://" + local(spec.Guard),
		"PROVEN_GUARDRAIL_KEY":      "guard-key",
		"ZETARIS_REST_URL":          "http://" + local(spec.Zetaris),
		"ZETARIS_API_KEY":           "zetaris-key",
		"ZETARIS_ORG_ID":            "1",
		"PROVEN_GENERATOR_TIMEOUT":  "10s",
		"DOCKER_HOST":               "tcp://" + local(spec.Daemon),
		"HOME":                      spec.Home,
	}
	_ = os.Setenv("DOCKER_HOST", env["DOCKER_HOST"])
	lookup := func(k string) string { return env[k] }
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Canary: a process of its own that connects. The witness must see it, or it sees no subprocess at all.
	canary := exec.CommandContext(ctx, "git", "ls-remote", "git://"+local(spec.Canary)+"/canary.git")
	_ = canary.Run()
	step("canary", nil)

	// The container runtime: asked whether it answers and has the image, it is the fake on PATH, which
	// connects to the daemon address like the real client. Its run line carries the two flags.
	cfg, err := testrun.Load(lookup)
	if err == nil {
		var box *testrun.Container
		if box, err = testrun.NewContainer(cfg, nil); err == nil {
			err = box.Probe(ctx)
			line := strings.Join(box.Argv(spec.Home, []string{"go", "test"}), " ")
			if err == nil && (!strings.Contains(line, "--pull never") || !strings.Contains(line, "--network none")) {
				err = fmt.Errorf("run line %q", line)
			}
		}
	}
	step("runtime", err)

	review := func(req Request) error {
		rev, err := Open(Options{LogDir: filepath.Join(spec.Home, "logs", strconv.Itoa(len(req.Target))+filepath.Base(req.Target)), Env: lookup})
		if err != nil {
			return err
		}
		res, err := rev.Review(ctx, req)
		if err != nil {
			return err
		}
		for _, s := range res.FailedServices {
			if s.Service == "nightshift" || s.Service == "guardrail" || s.Service == "zetaris" {
				return fmt.Errorf("service failed: %#v", res.FailedServices)
			}
		}
		if res.Partial {
			return fmt.Errorf("partial result")
		}
		return nil
	}
	step("review-fixture", review(Request{Target: spec.Fixture}))
	step("review-full-clone", review(Request{Target: spec.Full.Dir, BaselineCommit: spec.Full.Base, HeadCommit: spec.Full.Head}))
	perr := review(Request{Target: spec.Partial.Dir, BaselineCommit: spec.Partial.Base, HeadCommit: spec.Partial.Head})
	if perr == nil || !strings.Contains(perr.Error(), "PROVEN_NETWORK=local") {
		step("review-partial-clone", fmt.Errorf("a missing object was not refused by name: %v", perr))
	} else {
		fmt.Fprintf(out, "step review-partial-clone ok refused: %s\n", strings.ReplaceAll(perr.Error(), "\n", " "))
	}

	// A public address asked for by hand, through the policy and through a client that carries it.
	pol, err := network.Load(lookup)
	if err != nil {
		step("probe", err)
		return 1
	}
	var probeErr error
	for _, base := range []string{"https://203.0.113.9:443/v1", "https://integrate.api.nvidia.com/v1"} {
		g := generator.Settings{BaseURL: base, App: "proven", Model: "meterless/auto", Token: "test-token", Timeout: 5 * time.Second, Network: pol}
		if _, err := generator.NewLive(g, nil).Complete(ctx, "planner", "probe"); err == nil {
			probeErr = fmt.Errorf("a call to %s succeeded", base)
		}
	}
	for _, addr := range []string{"203.0.113.9:443", "8.8.8.8:53", "integrate.api.nvidia.com:443"} {
		if conn, err := pol.DialContext(ctx, "tcp", addr); err == nil {
			conn.Close()
			probeErr = fmt.Errorf("a dial to %s opened", addr)
		}
	}
	step("probe-public", probeErr)

	// The witness's own account, in the order the calls were made.
	for _, r := range w.records() {
		fmt.Fprintf(out, "connect pid=%d allowed=%v %s\n", r.Pid, r.Allowed, r.Target)
	}
	fmt.Fprintln(out, "done")
	return 0
}

// accepting listens on a loopback port, takes each connection and closes it. It is a service that
// exists, so an allowed connect has somewhere to land.
func accepting(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port
}

// serving starts the in-memory service of one port (see memoryService) on a real loopback socket.
func serving(t *testing.T, port string) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: memoryService(port)}
	t.Cleanup(func() { srv.Close() })
	go srv.Serve(ln)
	return ln.Addr().(*net.TCPAddr).Port
}

// partialCloneOf returns a clone of repo that holds the commits and no blob, whose origin is a public
// address. A read of a file makes git fetch it from there, unless the mode forbids it.
func partialCloneOf(t *testing.T, repo string) string {
	t.Helper()
	gitCmd(t, repo, "config", "uploadpack.allowFilter", "true")
	gitCmd(t, repo, "config", "uploadpack.allowAnySHA1InWant", "true")
	clone := filepath.Join(t.TempDir(), "clone")
	gitCmd(t, filepath.Dir(clone), "clone", "--no-checkout", "--filter=blob:none", "file://"+repo, clone)
	gitCmd(t, clone, "remote", "set-url", "origin", "git://127.0.0.2/proven.git")
	return clone
}

// fakeRuntime writes a container command that answers version and image inspect like a healthy
// runtime, and connects to DOCKER_HOST first, as the real client does.
func fakeRuntime(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/bash\n" +
		"if [ -n \"$DOCKER_HOST\" ]; then hp=${DOCKER_HOST#tcp://}; exec 3<>\"/dev/tcp/${hp%%:*}/${hp##*:}\"; fi\n" +
		"exit 0\n"
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestAirGappedWitnessSeesEveryConnection proves ADR-0034 with a witness that stands under the whole
// process: a seccomp filter on the connect syscall, inherited by git and the container runtime. A live
// review in PROVEN_NETWORK=local, with real loopback services, completes; every connect it made, its own
// and its subprocesses', went to an address the test allows; a public address asked for by hand and a
// name off the list reached no connect and no DNS port; and a review of a partial clone whose origin is
// public fails by name without a connect to that origin.
// Mutated fields: the guard of change.GitCommand (a connect to 127.0.0.2:9418), the dial check of the
// generator's client (a connect to 203.0.113.9:443) and the name check of Lookup (a connect to port 53).
func TestAirGappedWitnessSeesEveryConnection(t *testing.T) {
	root := moduleRootForWitness(t)
	full, fullBase, fullHead := twoCommits(t)
	src, srcBase, srcHead := twoCommits(t)
	spec := witnessSpec{
		Gen: serving(t, "8088"), Guard: serving(t, "8081"), Zetaris: serving(t, "9"),
		Daemon: accepting(t), Canary: accepting(t),
		Home:    t.TempDir(),
		Fixture: filepath.Join(root, "testdata", "fixture", "discount"),
		Full:    witnessRepo{full, fullBase, fullHead},
		Partial: witnessRepo{partialCloneOf(t, src), srcBase, srcHead},
	}
	raw, _ := json.Marshal(spec)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = []string{
		"PROVEN_WITNESS_SPEC=" + string(raw),
		"PATH=" + fakeRuntime(t) + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + spec.Home,
		"TMPDIR=" + spec.Home,
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	report := string(stdout)
	if err != nil {
		t.Fatalf("child: %v\nstderr %s\nstdout %s", err, stderr.String(), report)
	}
	if !strings.Contains(report, "\ndone\n") {
		t.Fatalf("the child did not finish\n%s", report)
	}
	var childPid string
	targets := map[string]string{} // target -> pid of its first connect
	for _, line := range strings.Split(report, "\n") {
		switch {
		case strings.HasPrefix(line, "pid "):
			childPid = strings.TrimPrefix(line, "pid ")
		case strings.HasPrefix(line, "step ") && strings.Contains(line, " FAIL "):
			t.Errorf("%s", line)
		case strings.HasPrefix(line, "connect "):
			f := strings.Fields(line)
			if len(f) != 4 {
				t.Errorf("record %q", line)
				continue
			}
			if f[2] != "allowed=true" {
				t.Errorf("a connection the witness refused: %s", line)
			}
			if _, ok := targets[f[3]]; !ok {
				targets[f[3]] = strings.TrimPrefix(f[1], "pid=")
			}
			host, port, _ := net.SplitHostPort(f[3])
			if host != "127.0.0.1" || port == "53" {
				t.Errorf("a connect off the allowed set: %s", line)
			}
		}
	}
	t.Logf("recorded connect targets: %v", targets)
	for name, port := range map[string]int{"generator": spec.Gen, "guardrail": spec.Guard, "zetaris": spec.Zetaris} {
		addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		if pid, ok := targets[addr]; !ok || pid != childPid {
			t.Errorf("no connect by the review itself to the %s at %s (pid %q, child %s)", name, addr, pid, childPid)
		}
	}
	// A subprocess made each of these, so the filter reached it: the container runtime (the fake, a
	// bash script that connects to the daemon address) and git (the canary).
	for name, port := range map[string]int{"container runtime": spec.Daemon, "git": spec.Canary} {
		addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		if pid, ok := targets[addr]; !ok || pid == childPid {
			t.Errorf("the %s was not seen at %s (pid %q, child %s): the filter did not reach a subprocess", name, addr, pid, childPid)
		}
	}
	if !strings.Contains(report, "step review-partial-clone ok refused:") {
		t.Errorf("the partial clone was not refused by name\n%s", report)
	}
}

// moduleRootForWitness is the repository root, two directories above this package.
func moduleRootForWitness(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
