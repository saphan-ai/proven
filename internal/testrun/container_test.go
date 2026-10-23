package testrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// TestDefaultContainerConfigValidates proves the defaults pass Validate and
// that AGENTS.md documents each of them, so the code and the page cannot drift.
func TestDefaultContainerConfigValidates(t *testing.T) {
	cfg := DefaultContainerConfig()
	if cfg.Runtime != "docker" || cfg.Image != "golang:1.26" || cfg.Timeout != 2*time.Minute || cfg.CPUs != "1" || cfg.Memory != "512m" {
		t.Fatalf("%#v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	root := repoRoot(t)
	doc, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{cfg.Runtime, cfg.Image, cfg.CPUs, cfg.Memory, "2 minutes", "2 seconds", "3 seconds", "may remain"} {
		if !bytes.Contains(doc, []byte(want)) {
			t.Fatalf("AGENTS.md does not document %s", want)
		}
	}
}

// TestValidateNamesTheRefusedSetting proves each bad value is refused by its
// field name with allowed values, and that a fractional CPU limit is accepted.
func TestValidateNamesTheRefusedSetting(t *testing.T) {
	base := DefaultContainerConfig()
	cases := []struct {
		name  string
		edit  func(*ContainerConfig)
		field string
	}{
		{"empty runtime", func(c *ContainerConfig) { c.Runtime = "" }, "Runtime"},
		{"runtime flag", func(c *ContainerConfig) { c.Runtime = "-rf" }, "Runtime"},
		{"empty image", func(c *ContainerConfig) { c.Image = "" }, "Image"},
		{"image flag", func(c *ContainerConfig) { c.Image = "-pull" }, "Image"},
		{"zero timeout", func(c *ContainerConfig) { c.Timeout = 0 }, "Timeout"},
		{"zero cpus", func(c *ContainerConfig) { c.CPUs = "0" }, "CPUs"},
		{"cpu word", func(c *ContainerConfig) { c.CPUs = "one" }, "CPUs"},
		{"zero memory", func(c *ContainerConfig) { c.Memory = "0" }, "Memory"},
		{"memory suffix", func(c *ContainerConfig) { c.Memory = "512M" }, "Memory"},
		{"zero written limit", func(c *ContainerConfig) { c.MaxWrittenBytes = 0 }, "MaxWrittenBytes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.edit(&cfg)
			err := cfg.Validate()
			var se *ConfigError
			if !errors.As(err, &se) || se.Field != tc.field {
				t.Fatalf("%v", err)
			}
			if len(se.Allowed) == 0 {
				t.Fatal("allowed values missing")
			}
		})
	}
	half := base
	half.CPUs = "0.5"
	if err := half.Validate(); err != nil {
		t.Fatal(err)
	}
}

// TestNewContainerDoesNotLookUpAnInvalidSetting proves a bad setting is
// refused before the runtime is looked up, and is not reported as a down service.
func TestNewContainerDoesNotLookUpAnInvalidSetting(t *testing.T) {
	called := false
	_, err := NewContainer(ContainerConfig{}, func(string) (string, error) {
		called = true
		return "/bin/docker", nil
	})
	if err == nil {
		t.Fatal("expected a refused setting")
	}
	if called {
		t.Fatal("lookPath ran")
	}
	if _, ok := service.As(err); ok {
		t.Fatal("a bad setting is not a down service")
	}
}

// TestNewContainerMissingRuntimeIsProverDown proves a missing runtime turns
// the Prover off by name, with the runtime named in the reason (ADR-0025 point 6).
func TestNewContainerMissingRuntimeIsProverDown(t *testing.T) {
	_, err := NewContainer(DefaultContainerConfig(), func(name string) (string, error) {
		if name != "docker" {
			t.Fatalf("looked up %s", name)
		}
		return "", errors.New("missing")
	})
	d, ok := service.As(err)
	if !ok {
		t.Fatal(err)
	}
	if d.Service != "prover" || !strings.Contains(d.Reason, "Prover off") || !strings.Contains(d.Reason, "docker") {
		t.Fatalf("%s %s", d.Service, d.Reason)
	}
}

// TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv proves the run line is
// stable and carries no network, no environment and one read-only mount (ADR-0025).
func TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv(t *testing.T) {
	dir := t.TempDir()
	c := mustContainer(t, DefaultContainerConfig())
	argv := c.Argv(dir, goTest)
	again := c.Argv(dir, goTest)
	if strings.Join(argv, "\x00") != strings.Join(again, "\x00") {
		t.Fatal("Argv is not stable")
	}
	if argv[0] != "/bin/docker" || argv[1] != "run" {
		t.Fatalf("%v", argv)
	}
	assertFlag(t, argv, "--rm")
	assertPair(t, argv, "--network", "none")
	assertPair(t, argv, "--cpus", "1")
	assertPair(t, argv, "--memory", "512m")
	assertPair(t, argv, "--workdir", "/checkout")
	mounts := 0
	for i, a := range argv {
		switch a {
		case "-v", "--volume", "--mount":
			mounts++
			if a == "-v" && !strings.HasPrefix(argv[i+1], dir+":/checkout:ro") && argv[i+1] != dir+":/checkout:ro" {
				abs, _ := filepath.Abs(dir)
				if argv[i+1] != abs+":/checkout:ro" {
					t.Fatalf("mount %s", argv[i+1])
				}
			}
		case "-e", "--env", "--env-file", "--privileged", "--network=host":
			t.Fatalf("forbidden %s in %v", a, argv)
		}
	}
	if mounts != 1 {
		t.Fatalf("mounts %d in %v", mounts, argv)
	}
	// The words after the image are the test command and nothing else: no shell.
	tail := argv[len(argv)-len(goTest)-1:]
	if tail[0] != "golang:1.26" || strings.Join(tail[1:], "\x00") != strings.Join(goTest, "\x00") {
		t.Fatalf("%v", argv)
	}
	for _, a := range argv {
		if a == "sh" || a == "-c" {
			t.Fatalf("a shell in the run line: %v", argv)
		}
	}
	for _, a := range argv {
		if a == "--name" {
			t.Fatal("Argv must stay stable; Run adds --name")
		}
	}
}

// TestArgvUsesTheConfiguredLimits proves the CPU, memory and image settings
// reach the run line unchanged.
func TestArgvUsesTheConfiguredLimits(t *testing.T) {
	cfg := DefaultContainerConfig()
	cfg.CPUs = "0.5"
	cfg.Memory = "256m"
	cfg.Image = "golang:1.26-bookworm"
	c := mustContainer(t, cfg)
	argv := c.Argv(t.TempDir(), goTest)
	assertPair(t, argv, "--cpus", "0.5")
	assertPair(t, argv, "--memory", "256m")
	found := false
	for _, a := range argv {
		if a == "golang:1.26-bookworm" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%v", argv)
	}
}

// TestRunReadsBothCheckoutsAndDoesNotProveATimeout proves a run that times
// out on both sides is not proof, and each side is run once and removed once.
func TestRunReadsBothCheckoutsAndDoesNotProveATimeout(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	cfg := DefaultContainerConfig()
	cfg.Timeout = 20 * time.Millisecond
	c := mustContainer(t, cfg)
	var runs [][]string
	c.run = func(ctx context.Context, argv []string) (string, error) {
		runs = append(runs, append([]string(nil), argv...))
		if len(argv) > 1 && argv[1] == "rm" {
			return "", nil
		}
		<-ctx.Done()
		return "still running", ctx.Err()
	}
	out, err := c.Run(context.Background(), Request{
		BaseDir: base, ChangeDir: change, Name: "TestDiscounted", Path: "billing/discount_test.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Found || out.FailsOnChange || out.PassesOnBase {
		t.Fatalf("%#v", out)
	}
	if !strings.Contains(out.Output, "not proof") {
		t.Fatal(out.Output)
	}
	runCount, rmCount := 0, 0
	for _, argv := range runs {
		if argv[1] == "run" {
			runCount++
			if !sameSentence(stripName(argv), c.Argv(mountHost(argv), goTest)) {
				t.Fatalf("ran %v", argv)
			}
		}
		if argv[1] == "rm" {
			rmCount++
		}
		forbidEnv(t, argv)
	}
	if runCount != 2 || rmCount != 2 {
		t.Fatalf("runs %d rm %d", runCount, rmCount)
	}
}

// T7: Timeout is 1s and every runtime call would sleep 30s. The fake returns
// when its context ends, which is what exec.CommandContext does to a hang.
// It proves Run returns within both sides' limits plus the cleanup margin,
// and names each container that may remain.
func TestRunReturnsWithinTenSecondsWhenTheRuntimeHangs(t *testing.T) {
	if cleanupMargin > 3*time.Second || cleanupBound <= 0 || cleanupBound > cleanupMargin {
		t.Fatalf("cleanup margin %s bound %s", cleanupMargin, cleanupBound)
	}
	base := t.TempDir()
	change := t.TempDir()
	cfg := DefaultContainerConfig()
	cfg.Timeout = time.Second
	c := mustContainer(t, cfg)
	var calls []string
	c.run = func(ctx context.Context, argv []string) (string, error) {
		kind := "run"
		if len(argv) > 1 {
			kind = argv[1]
		}
		start := time.Now()
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			calls = append(calls, fmt.Sprintf("%s %s", kind, time.Since(start).Round(time.Millisecond)))
			return "hung", ctx.Err()
		case <-timer.C:
			calls = append(calls, kind+" slept 30s")
			return "slept", nil
		}
	}
	start := time.Now()
	out, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: base, ChangeDir: change,
	})
	elapsed := time.Since(start)
	t.Logf("elapsed %s calls %v", elapsed.Round(time.Millisecond), calls)
	if err != nil {
		t.Fatal(err)
	}
	limit := 2 * (cfg.Timeout + cleanupMargin)
	if elapsed > limit || elapsed > 10*time.Second {
		t.Fatalf("Run held for %s; limit %s; calls %v; output %s", elapsed.Round(time.Millisecond), limit, calls, out.Output)
	}
	if !out.Found || out.FailsOnChange || out.PassesOnBase {
		t.Fatalf("%#v", out)
	}
	if !strings.Contains(out.Output, "not proof") {
		t.Fatal(out.Output)
	}
	if strings.Count(out.Output, "may remain") < 2 {
		t.Fatalf("a hung removal was silent: %s", out.Output)
	}
}

// TestRunNamesTheContainerWhenRemovalFails proves a failed removal names the
// container and the runtime's error in the output.
func TestRunNamesTheContainerWhenRemovalFails(t *testing.T) {
	cfg := DefaultContainerConfig()
	cfg.Timeout = 20 * time.Millisecond
	c := mustContainer(t, cfg)
	var names []string
	c.run = func(ctx context.Context, argv []string) (string, error) {
		if len(argv) > 1 && argv[1] == "rm" {
			if len(argv) < 4 || argv[2] != "-f" {
				t.Fatalf("rm argv %v", argv)
			}
			names = append(names, argv[3])
			return "Error response from daemon: removal failed", errors.New("device or resource busy")
		}
		<-ctx.Done()
		return "still running", ctx.Err()
	}
	out, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Found || out.FailsOnChange || out.PassesOnBase || !strings.Contains(out.Output, "not proof") {
		t.Fatalf("%#v", out)
	}
	if len(names) != 2 {
		t.Fatalf("removals %v", names)
	}
	for _, name := range names {
		if !strings.Contains(out.Output, name) || !strings.Contains(out.Output, "may remain") || !strings.Contains(out.Output, "device or resource busy") {
			t.Fatalf("output does not name %s: %s", name, out.Output)
		}
	}
}

// TestRunSetsProofOnlyWhenChangeFailsAndBasePasses proves the proof needs the
// change side to fail and the base side to pass, each in its own checkout.
func TestRunSetsProofOnlyWhenChangeFailsAndBasePasses(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	c := mustContainer(t, DefaultContainerConfig())
	fail := exitCode(t, 1)
	var sawChange, sawBase bool
	c.run = func(ctx context.Context, argv []string) (string, error) {
		if argv[1] == "rm" {
			return "", nil
		}
		mount := mountOf(argv)
		switch {
		case strings.Contains(mount, change+":/checkout:ro"):
			sawChange = true
			return "FAIL TestDiscounted", fail
		case strings.Contains(mount, base+":/checkout:ro"):
			sawBase = true
			return "ok", nil
		default:
			t.Fatalf("mount %s", mount)
			return "", nil
		}
	}
	out, err := c.Run(context.Background(), Request{Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: base, ChangeDir: change})
	if err != nil {
		t.Fatal(err)
	}
	if !sawChange || !sawBase || !out.Found || !out.FailsOnChange || !out.PassesOnBase {
		t.Fatalf("saw change %t base %t %#v", sawChange, sawBase, out)
	}
}

// TestRunDockerFailureIsNotProof proves exit 125, the runtime's own failure,
// is reported as a test that could not run, not as a failing test.
func TestRunDockerFailureIsNotProof(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	code := exitCode(t, 125)
	c.run = func(ctx context.Context, argv []string) (string, error) {
		if argv[1] == "rm" {
			return "", nil
		}
		return "docker: Error response from daemon", code
	}
	out, err := c.Run(context.Background(), Request{Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if out.FailsOnChange || out.PassesOnBase || !strings.Contains(out.Output, "could not run") {
		t.Fatalf("%#v", out)
	}
}

// TestRunDoesNotStartWhenTheRequestIsIncomplete proves a request without a
// command or checkouts starts no container and proves nothing.
func TestRunDoesNotStartWhenTheRequestIsIncomplete(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	c.run = func(context.Context, []string) (string, error) {
		t.Fatal("started")
		return "", nil
	}
	out, err := c.Run(context.Background(), Request{Name: "TestDiscounted"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Found || out.FailsOnChange || out.PassesOnBase {
		t.Fatalf("%#v", out)
	}
}

// TestRunCancelledContextStartsNothing proves a cancelled review starts no
// container and returns the cancellation.
func TestRunCancelledContextStartsNothing(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	c.run = func(context.Context, []string) (string, error) {
		t.Fatal("started")
		return "", nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Run(ctx, Request{Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: t.TempDir()})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// TestOffRunsNothing proves the Off adapter reports the Prover off by name.
func TestOffRunsNothing(t *testing.T) {
	_, err := (Off{Reason: "Prover off: container runtime \"docker\" not found"}).Run(context.Background(), Request{Path: "billing/discount_test.go", Name: "TestDiscounted"})
	d, ok := service.As(err)
	if !ok || d.Service != "prover" || !strings.Contains(d.Reason, "Prover off") {
		t.Fatal(err)
	}
}

// mustContainer builds a Container whose runtime lookup always succeeds.
func mustContainer(t *testing.T, cfg ContainerConfig) *Container {
	t.Helper()
	c, err := NewContainer(cfg, func(string) (string, error) { return "/bin/docker", nil })
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// exitCode returns a real *exec.ExitError with code, as the runtime would.
func exitCode(t *testing.T, code int) error {
	t.Helper()
	cmd := exec.Command("sh", "-c", "exit "+strconv.Itoa(code))
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected a non-zero exit")
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != code {
		t.Fatal(err)
	}
	return err
}

func assertFlag(t *testing.T, argv []string, flag string) {
	t.Helper()
	for _, a := range argv {
		if a == flag {
			return
		}
	}
	t.Fatalf("missing %s in %v", flag, argv)
}

func assertPair(t *testing.T, argv []string, flag, value string) {
	t.Helper()
	for i, a := range argv {
		if a == flag && i+1 < len(argv) && argv[i+1] == value {
			return
		}
	}
	t.Fatalf("missing %s %s in %v", flag, value, argv)
}

func forbidEnv(t *testing.T, argv []string) {
	t.Helper()
	for _, a := range argv {
		if a == "-e" || a == "--env" || a == "--env-file" || strings.HasPrefix(a, "--env=") {
			t.Fatalf("env in %v", argv)
		}
	}
}

// stripName drops the per-run --name, so a run line can be compared with Argv.
func stripName(argv []string) []string {
	out := make([]string, 0, len(argv))
	for i := 0; i < len(argv); i++ {
		if argv[i] == "--name" {
			i++
			continue
		}
		out = append(out, argv[i])
	}
	return out
}

func mountOf(argv []string) string {
	for i, a := range argv {
		if a == "-v" && i+1 < len(argv) {
			return argv[i+1]
		}
	}
	return ""
}

// mountHost returns the host side of the run line's one mount.
func mountHost(argv []string) string {
	mount := mountOf(argv)
	if i := strings.Index(mount, ":/checkout:ro"); i >= 0 {
		return mount[:i]
	}
	return mount
}

func sameSentence(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// repoRoot finds the module root, where AGENTS.md lives.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

// TestRemovalFailureSurvivesLongOutput proves a removal failure stays in the
// output when the side's own output is longer than the output limit.
func TestRemovalFailureSurvivesLongOutput(t *testing.T) {
	cfg := DefaultContainerConfig()
	cfg.Timeout = 20 * time.Millisecond
	c := mustContainer(t, cfg)
	long := strings.Repeat("x", 20000)
	c.run = func(ctx context.Context, argv []string) (string, error) {
		if len(argv) > 1 && argv[1] == "rm" {
			return "", errors.New("device or resource busy")
		}
		<-ctx.Done()
		return long, ctx.Err()
	}
	out, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.Output, "may remain") < 2 {
		t.Fatalf("a removal failure was cut off by the output limit: %d bytes, %d notes", len(out.Output), strings.Count(out.Output, "may remain"))
	}
}
