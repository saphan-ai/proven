package testrun_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// contractW1bSecret is the value T2 plants in the process environment.
// No argv element may carry it.
const contractW1bSecret = "s3cret-T2"

// contractW1bDialSource is a test that dials 1.1.1.1:53 and fails either way,
// naming which way it went, so the outcome's output shows it. T10 writes it
// into each checkout as dial_test.go, beside contractW1bDialModule.
const contractW1bDialSource = `package dial

import (
	"net"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "1.1.1.1:53", 5*time.Second)
	if err != nil {
		t.Fatalf("dial-failed: %v", err)
	}
	conn.Close()
	t.Fatal("dial-ok")
}
`

// contractW1bDialModule is the module file the dial test needs to run with go test.
const contractW1bDialModule = "module dial\n\ngo 1.22\n"

// TestContractW1bT1 proves the run line turns the network off (ADR-0025 point 2).
func TestContractW1bT1(t *testing.T) {
	runtime := contractW1bRuntime(t, "exit 0\n")
	dir := t.TempDir()
	why := "T1: argv must carry --network none or --network=none"
	argv := contractW1bArgv(t, testrun.DefaultContainerConfig(), runtime, dir, why)
	if !contractW1bNetworkNone(argv) {
		t.Fatalf("%s; got %q", why, argv)
	}
}

// TestContractW1bT2 proves no environment variable of Proven reaches the
// container, even when one holds a secret (ADR-0025 point 3).
func TestContractW1bT2(t *testing.T) {
	t.Setenv("METERLESS_RUNTIME_TOKEN", contractW1bSecret)
	runtime := contractW1bRuntime(t, "exit 0\n")
	dir := t.TempDir()
	why := "T2: argv must pass no Proven environment (no -e, --env, or --env-file, and no element contains " + contractW1bSecret + ")"
	argv := contractW1bArgv(t, testrun.DefaultContainerConfig(), runtime, dir, why)
	if len(argv) == 0 {
		t.Fatalf("%s; argv is empty", why)
	}
	if leak := contractW1bEnvLeak(argv); leak != "" {
		t.Fatalf("%s; found %q in %q", why, leak, argv)
	}
}

// TestContractW1bT3 proves the container mounts the one checkout and no other
// host path (ADR-0025 point 4).
func TestContractW1bT3(t *testing.T) {
	runtime := contractW1bRuntime(t, "exit 0\n")
	dir := t.TempDir()
	why := "T3: argv must mount exactly the given dir and no other host path"
	argv := contractW1bArgv(t, testrun.DefaultContainerConfig(), runtime, dir, why)
	flags, sources := contractW1bMounts(argv)
	if flags != 1 || len(sources) != 1 || filepath.Clean(sources[0]) != filepath.Clean(dir) {
		t.Fatalf("%s; mounts=%d sources=%q argv=%q", why, flags, sources, argv)
	}
	if others := contractW1bOtherHostPaths(argv, dir, runtime); len(others) > 0 {
		t.Fatalf("%s; other host paths %q argv=%q", why, others, argv)
	}
}

// TestContractW1bT4 proves the container is removed on exit and carries the
// configured CPU and memory limits (ADR-0025 point 5).
func TestContractW1bT4(t *testing.T) {
	runtime := contractW1bRuntime(t, "exit 0\n")
	cfg := testrun.DefaultContainerConfig()
	cfg.CPUs = "1"
	cfg.Memory = "512m"
	dir := t.TempDir()
	why := "T4: argv must carry --rm, the CPU limit 1 and the memory limit 512m"
	argv := contractW1bArgv(t, cfg, runtime, dir, why)
	var missing []string
	if !contractW1bHasArg(argv, "--rm") {
		missing = append(missing, "--rm")
	}
	if !contractW1bFlagValue(argv, []string{"--cpus"}, "1") {
		missing = append(missing, "cpu limit 1")
	}
	if !contractW1bFlagValue(argv, []string{"--memory", "-m"}, "512m") {
		missing = append(missing, "memory limit 512m")
	}
	if len(missing) > 0 {
		t.Fatalf("%s; missing %s; argv %q", why, strings.Join(missing, ", "), argv)
	}
}

// TestContractW1bT5 proves a missing runtime turns the Prover off and runs
// nothing on the host (ADR-0025 point 6).
func TestContractW1bT5(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	runtime := contractW1bRuntime(t, "echo ran > "+contractW1bShQuote(t, marker)+"\nexit 0\n")
	_, err := testrun.NewContainer(testrun.DefaultContainerConfig(), func(string) (string, error) {
		return "", errors.New("runtime not on PATH")
	})
	var down *service.Down
	if !errors.As(err, &down) || down == nil || down.Service != "prover" {
		got := ""
		if down != nil {
			got = down.Service
		}
		t.Fatalf("T5: NewContainer must return *service.Down with Service prover when lookPath fails; got %v (service %q)", err, got)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatalf("T5: fake runtime %s was executed", runtime)
	}
}

// TestContractW1bT6 proves a test that fails on the change and passes on the
// base is proof, from two runs of the runtime.
func TestContractW1bT6(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	if strings.Contains(change, base) || strings.Contains(base, change) {
		t.Fatalf("T6: checkout paths overlap %q and %q", base, change)
	}
	record := t.TempDir()
	runtime := contractW1bRuntime(t, contractW1bRecordScript(t, record, contractW1bExitForCheckouts(t, base, change)))
	why := "T6: a test that fails on the change and passes on the base is proved, and the runtime runs twice"
	c := contractW1bContainer(t, testrun.DefaultContainerConfig(), runtime, why)
	out, err := c.Run(context.Background(), contractW1bRequest("F-T6", "TestProved", base, change))
	runs := contractW1bRuns(t, record)
	if err != nil || !out.Found || !out.FailsOnChange || !out.PassesOnBase || len(runs) != 2 {
		t.Fatalf("%s; err=%v outcome=%+v runs=%d", why, err, out, len(runs))
	}
}

// TestContractW1bT7 proves a hung runtime returns within 10 seconds and the
// timed-out run is not proof.
func TestContractW1bT7(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "pids")
	runtime := contractW1bRuntime(t, "echo $$ >> "+contractW1bShQuote(t, pidfile)+"\nexec sleep 30\n")
	t.Cleanup(func() { contractW1bKillPids(pidfile) })
	cfg := testrun.DefaultContainerConfig()
	cfg.Timeout = time.Second
	why := "T7: a run past its time limit returns within 10s and is not proof, or the error names the time limit"
	c := contractW1bContainer(t, cfg, runtime, why)
	base := t.TempDir()
	change := t.TempDir()
	var (
		out testrun.Outcome
		err error
	)
	done := make(chan struct{})
	go func() {
		defer close(done)
		out, err = c.Run(context.Background(), contractW1bRequest("F-T7", "TestHung", base, change))
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("%s; Run did not return within 10s", why)
	}
	if out.FailsOnChange && out.PassesOnBase && !contractW1bNamesTimeLimit(err) {
		t.Fatalf("%s; outcome is proved (%+v) and error %v does not name the time limit", why, out, err)
	}
}

// TestContractW1bT8 proves an empty image and a zero time limit are refused
// by name, and the defaults carry both (ADR-0013).
func TestContractW1bT8(t *testing.T) {
	noImage := testrun.DefaultContainerConfig()
	noImage.Image = ""
	if err := noImage.Validate(); err == nil || !strings.Contains(err.Error(), "Image") {
		t.Fatalf("T8: Validate must refuse an empty Image by name; got %v", err)
	}
	noTimeout := testrun.DefaultContainerConfig()
	noTimeout.Timeout = 0
	if err := noTimeout.Validate(); err == nil || !strings.Contains(err.Error(), "Timeout") {
		t.Fatalf("T8: Validate must refuse a zero Timeout by name; got %v", err)
	}
	cfg := testrun.DefaultContainerConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("T8: DefaultContainerConfig must validate; got %v", err)
	}
	if cfg.Image == "" || cfg.Timeout <= 0 {
		t.Fatalf("T8: defaults must carry a non-empty Image and a positive Timeout; got Image=%q Timeout=%s", cfg.Image, cfg.Timeout)
	}
}

// TestContractW1bT9 proves the base run mounts the base and the change run
// mounts the change, one checkout each.
func TestContractW1bT9(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	record := t.TempDir()
	runtime := contractW1bRuntime(t, contractW1bRecordScript(t, record, "exit 0\n"))
	why := "T9: the two runs mount BaseDir and ChangeDir, one each"
	c := contractW1bContainer(t, testrun.DefaultContainerConfig(), runtime, why)
	_, err := c.Run(context.Background(), contractW1bRequest("F-T9", "TestCheckout", base, change))
	runs := contractW1bRuns(t, record)
	var sources []string
	for _, argv := range runs {
		flags, src := contractW1bMounts(argv)
		if flags != 1 || len(src) != 1 {
			t.Fatalf("%s; a run mounted %d sources %q argv %q", why, flags, src, argv)
		}
		sources = append(sources, filepath.Clean(src[0]))
	}
	want := []string{filepath.Clean(base), filepath.Clean(change)}
	if err != nil || !contractW1bSamePaths(sources, want) {
		t.Fatalf("%s; err=%v sources=%q want %q", why, err, sources, want)
	}
}

// TestContractW1bT10 proves, in a real container, that a dial out fails.
// It needs a runtime and the image, so it runs only with PROVEN_TEST_DOCKER=1.
func TestContractW1bT10(t *testing.T) {
	if os.Getenv("PROVEN_TEST_DOCKER") != "1" {
		t.Skip("PROVEN_TEST_DOCKER=1")
	}
	c, err := testrun.NewContainer(testrun.DefaultContainerConfig(), exec.LookPath)
	if err != nil || c == nil {
		t.Fatalf("T10: the dial must fail inside a real container; NewContainer: %v", err)
	}
	base := t.TempDir()
	change := t.TempDir()
	for _, dir := range []string{base, change} {
		if writeErr := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(contractW1bDialModule), 0o644); writeErr != nil {
			t.Fatalf("T10: write go.mod: %v", writeErr)
		}
		if writeErr := os.WriteFile(filepath.Join(dir, "dial_test.go"), []byte(contractW1bDialSource), 0o644); writeErr != nil {
			t.Fatalf("T10: write dial_test.go: %v", writeErr)
		}
	}
	var out testrun.Outcome
	done := make(chan struct{})
	go func() {
		defer close(done)
		out, err = c.Run(context.Background(), contractW1bRequest("F-T10", "TestDial", base, change))
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Minute):
		t.Fatal("T10: Run did not return within 2m; the dial inside the container did not finish")
	}
	if err != nil || out.PassesOnBase || strings.Contains(out.Output, "dial-ok") || !strings.Contains(out.Output, "dial-failed") {
		t.Fatalf("T10: the dial must fail inside the container; err=%v outcome=%+v", err, out)
	}
}

func contractW1bContainer(t *testing.T, cfg testrun.ContainerConfig, runtime, why string) *testrun.Container {
	t.Helper()
	c, err := testrun.NewContainer(cfg, contractW1bFound(runtime))
	if err != nil || c == nil {
		t.Fatalf("%s; NewContainer: container=%v err=%v", why, c != nil, err)
	}
	return c
}

func contractW1bArgv(t *testing.T, cfg testrun.ContainerConfig, runtime, dir, why string) []string {
	t.Helper()
	c := contractW1bContainer(t, cfg, runtime, why)
	return c.Argv(dir, []string{"go", "test", "-run", "^TestOne$", "./pkg"})
}

func contractW1bFound(runtime string) func(string) (string, error) {
	return func(string) (string, error) {
		return runtime, nil
	}
}

func contractW1bRequest(id, name, base, change string) testrun.Request {
	return testrun.Request{
		FindingID: id,
		Path:      "contract_test.go",
		Name:      name,
		BaseDir:   base,
		ChangeDir: change,
	}
}

// contractW1bRuntime writes a fake runtime script with body and returns its path.
func contractW1bRuntime(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-runtime")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("contractW1b: write runtime: %v", err)
	}
	return path
}

func contractW1bShQuote(t *testing.T, s string) string {
	t.Helper()
	if strings.ContainsAny(s, "'\n") {
		t.Fatalf("contractW1b: path %q cannot be quoted", s)
	}
	return "'" + s + "'"
}

// contractW1bRecordScript is a fake runtime body that saves its argv in
// recordDir, then runs after.
func contractW1bRecordScript(t *testing.T, recordDir, after string) string {
	t.Helper()
	q := contractW1bShQuote(t, recordDir)
	return "rec=" + q + "\n" +
		"mkdir -p \"$rec\"\n" +
		"id=\"$rec/$(date +%s%N)-$$\"\n" +
		"printf '%s\\n' '---' > \"$id\"\n" +
		"for a in \"$@\"; do\n" +
		"  printf '%s\\n' \"$a\" >> \"$id\"\n" +
		"done\n" +
		after
}

// contractW1bExitForCheckouts is a script tail that exits 1 when the change
// is mounted, 0 when the base is, and 2 when neither is.
func contractW1bExitForCheckouts(t *testing.T, base, change string) string {
	t.Helper()
	return "base=" + contractW1bShQuote(t, base) + "\n" +
		"change=" + contractW1bShQuote(t, change) + "\n" +
		"hit_change=0\n" +
		"hit_base=0\n" +
		"for a in \"$@\"; do\n" +
		"  case \"$a\" in\n" +
		"    *\"$change\"*) hit_change=1 ;;\n" +
		"  esac\n" +
		"  case \"$a\" in\n" +
		"    *\"$base\"*) hit_base=1 ;;\n" +
		"  esac\n" +
		"done\n" +
		"if [ \"$hit_change\" -eq 1 ]; then\n" +
		"  exit 1\n" +
		"fi\n" +
		"if [ \"$hit_base\" -eq 1 ]; then\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 2\n"
}

// contractW1bRuns reads back the argv of every recorded runtime call, in order.
func contractW1bRuns(t *testing.T, dir string) [][]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("contractW1b: read %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	runs := make([][]string, 0, len(names))
	for _, name := range names {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("contractW1b: read run: %v", err)
		}
		var args []string
		for _, line := range strings.Split(string(body), "\n") {
			if line == "" || line == "---" {
				continue
			}
			args = append(args, line)
		}
		runs = append(runs, args)
	}
	return runs
}

func contractW1bSamePaths(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	left := append([]string(nil), got...)
	right := append([]string(nil), want...)
	sort.Strings(left)
	sort.Strings(right)
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func contractW1bNetworkNone(argv []string) bool {
	for i, arg := range argv {
		if arg == "--network=none" {
			return true
		}
		if arg == "--network" && i+1 < len(argv) && argv[i+1] == "none" {
			return true
		}
	}
	return false
}

// contractW1bEnvLeak returns the first argv element that passes environment
// to the container or carries the planted secret.
func contractW1bEnvLeak(argv []string) string {
	for _, arg := range argv {
		switch arg {
		case "-e", "--env", "--env-file":
			return arg
		}
		if strings.HasPrefix(arg, "--env=") || strings.HasPrefix(arg, "--env-file=") {
			return arg
		}
		if strings.Contains(arg, contractW1bSecret) {
			return arg
		}
	}
	return ""
}

// contractW1bMounts counts the mount flags in argv, in every spelling the
// runtime accepts, and returns their host sources.
func contractW1bMounts(argv []string) (int, []string) {
	flags := 0
	var sources []string
	skip := false
	for i, arg := range argv {
		if skip {
			skip = false
			continue
		}
		switch {
		case arg == "-v" || arg == "--volume" || arg == "--mount":
			flags++
			if i+1 < len(argv) {
				if arg == "--mount" {
					sources = append(sources, contractW1bMountSource(argv[i+1]))
				} else {
					sources = append(sources, contractW1bVolumeSource(argv[i+1]))
				}
				skip = true
			}
		case strings.HasPrefix(arg, "--volume="):
			flags++
			sources = append(sources, contractW1bVolumeSource(strings.TrimPrefix(arg, "--volume=")))
		case strings.HasPrefix(arg, "--mount="):
			flags++
			sources = append(sources, contractW1bMountSource(strings.TrimPrefix(arg, "--mount=")))
		}
	}
	return flags, sources
}

func contractW1bVolumeSource(spec string) string {
	source, _, _ := strings.Cut(spec, ":")
	return source
}

func contractW1bMountSource(spec string) string {
	for _, field := range strings.Split(spec, ",") {
		if rest, ok := strings.CutPrefix(field, "source="); ok {
			return rest
		}
		if rest, ok := strings.CutPrefix(field, "src="); ok {
			return rest
		}
	}
	return ""
}

// contractW1bOtherHostPaths returns host paths in argv other than the
// allowed checkout and the runtime itself.
func contractW1bOtherHostPaths(argv []string, allowed, runtime string) []string {
	allowed = filepath.Clean(allowed)
	runtime = filepath.Clean(runtime)
	home, _ := os.UserHomeDir()
	if home != "" {
		home = filepath.Clean(home)
	}
	var bad []string
	seen := map[string]bool{}
	add := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		bad = append(bad, path)
	}
	for _, arg := range argv {
		for _, token := range contractW1bPathTokens(arg) {
			if !filepath.IsAbs(token) {
				continue
			}
			clean := filepath.Clean(token)
			if clean == allowed || clean == runtime {
				continue
			}
			if home != "" && clean == home {
				add(clean)
				continue
			}
			if _, err := os.Stat(clean); err != nil {
				continue
			}
			add(clean)
		}
	}
	return bad
}

func contractW1bPathTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '=' || r == ':'
	})
}

func contractW1bHasArg(argv []string, want string) bool {
	for _, arg := range argv {
		if arg == want {
			return true
		}
	}
	return false
}

func contractW1bFlagValue(argv, flags []string, value string) bool {
	for i, arg := range argv {
		for _, flag := range flags {
			if arg == flag+"="+value {
				return true
			}
			if arg == flag && i+1 < len(argv) && argv[i+1] == value {
				return true
			}
		}
	}
	return false
}

// contractW1bNamesTimeLimit reports whether err says the run hit its time limit.
func contractW1bNamesTimeLimit(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "timed out") ||
		strings.Contains(msg, "time limit") ||
		strings.Contains(msg, "deadline")
}

// contractW1bKillPids kills the fake runtimes T7 left sleeping.
func contractW1bKillPids(path string) {
	body, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil || pid <= 0 {
			continue
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		_ = proc.Kill()
	}
}
