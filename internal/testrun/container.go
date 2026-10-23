package testrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// The flags below are the ones docker-run(1) from Docker 29.9.1 documents.
// --network none is "No networking in the container". --rm removes the
// container when it exits. --cpus takes a number and 0 means no limit.
// --memory takes a number and a suffix b, k, m or g, and 0 means no limit.
// A volume's ro option mounts it read-only; the default is read-write.
// That manual does not document a flag that stops a missing image from
// being pulled on the host before the container starts.

// The documented defaults of the container settings (ADR-0013, ADR-0025
// point 5). checkoutDir is where the one checkout is mounted inside.
const (
	defaultRuntime = "docker"
	defaultImage   = "golang:1.26"
	defaultCPUs    = "1"
	defaultMemory  = "512m"
	checkoutDir    = "/checkout"
)

// defaultTimeout is the default limit for one side of a test run.
const defaultTimeout = 2 * time.Minute

// cleanupMargin is the longest one side (base or change) may run past
// Timeout, including removal. cleanupBound is the removal's own limit and
// sits inside that margin. A side therefore returns within Timeout plus
// 3 seconds even when every runtime call hangs.
const (
	cleanupMargin = 3 * time.Second
	cleanupBound  = 2 * time.Second
)

// ContainerConfig holds the container's settings (ADR-0025 point 5). There is no
// field that gives the container a network, an environment variable or a mount
// beyond the one checkout.
type ContainerConfig struct {
	Runtime         string        // container CLI name or path; default "docker"
	Image           string        // image the test runs in
	Timeout         time.Duration // limit for one test run, base or change
	CPUs            string        // CPU limit passed to the runtime
	Memory          string        // memory limit passed to the runtime
	MaxWrittenBytes int           // size limit for a test file the prover writes
}

// cpuPattern and memPattern accept only the plain forms docker-run(1)
// documents, so a value can never smuggle in another flag.
var (
	cpuPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	memPattern = regexp.MustCompile(`^([0-9]+)([bkmg])?$`)
)

// DefaultContainerConfig returns the documented defaults.
func DefaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		Runtime:         defaultRuntime,
		Image:           defaultImage,
		Timeout:         defaultTimeout,
		CPUs:            defaultCPUs,
		Memory:          defaultMemory,
		MaxWrittenBytes: DefaultMaxWrittenBytes,
	}
}

// Validate refuses a setting by name with the allowed values (ADR-0013).
func (c ContainerConfig) Validate() error {
	// Runtime and image become argv words; a leading dash would read as a flag.
	if !validToken(c.Runtime) {
		return &ConfigError{
			Field: "Runtime", Value: c.Runtime,
			Reason:  "the runtime is a CLI name or path, with no spaces and no leading dash",
			Allowed: []string{defaultRuntime},
		}
	}
	if !validToken(c.Image) {
		return &ConfigError{
			Field: "Image", Value: c.Image,
			Reason:  "the image is a reference, with no spaces and no leading dash",
			Allowed: []string{defaultImage},
		}
	}
	if c.Timeout <= 0 {
		return &ConfigError{
			Field: "Timeout", Value: c.Timeout.String(),
			Reason:  "the time limit is a positive duration; a run past it is not proof",
			Allowed: []string{defaultTimeout.String()},
		}
	}
	// 0 means no limit to the runtime, so a zero CPU or memory limit is refused:
	// a setting must not be able to lift the container's bounds.
	if !validCPUs(c.CPUs) {
		return &ConfigError{
			Field: "CPUs", Value: c.CPUs,
			Reason:  "the CPU limit is a positive number; 0 means no limit and is refused",
			Allowed: []string{defaultCPUs, "0.5"},
		}
	}
	if !validMemory(c.Memory) {
		return &ConfigError{
			Field: "Memory", Value: c.Memory,
			Reason:  "the memory limit is a positive size with suffix b, k, m or g; 0 is refused",
			Allowed: []string{defaultMemory},
		}
	}
	if c.MaxWrittenBytes <= 0 {
		return &ConfigError{
			Field: "MaxWrittenBytes", Value: strconv.Itoa(c.MaxWrittenBytes),
			Reason:  "the written test size limit is a positive number of bytes",
			Allowed: []string{strconv.Itoa(DefaultMaxWrittenBytes)},
		}
	}
	return nil
}

// validToken accepts one non-empty word that does not start with a dash.
func validToken(s string) bool {
	if s == "" || strings.HasPrefix(s, "-") {
		return false
	}
	return !strings.ContainsAny(s, " \t\n\r")
}

// validCPUs accepts a positive decimal number.
func validCPUs(s string) bool {
	if !cpuPattern.MatchString(s) {
		return false
	}
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && f > 0
}

// validMemory accepts a positive whole number with an optional b, k, m or g suffix.
func validMemory(s string) bool {
	m := memPattern.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	n, err := strconv.ParseUint(m[1], 10, 64)
	return err == nil && n > 0
}

// Container is the live test-run adapter. It runs each side of a test in a
// fresh container with no network, no environment and one read-only
// checkout (ADR-0025 points 1 to 4). review.Open builds it from Load.
type Container struct {
	runtime string
	cfg     ContainerConfig
	// run starts one runtime invocation. Production leaves it nil.
	run func(ctx context.Context, argv []string) (string, error)
}

// NewContainer resolves the runtime with lookPath. A nil lookPath uses
// exec.LookPath. When no runtime is found it returns a *service.Down naming
// the service "prover" (ADR-0025 point 6), and the caller runs no code.
func NewContainer(cfg ContainerConfig, lookPath func(string) (string, error)) (*Container, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	// Refuse a bad setting before looking for the runtime, so the refusal names the setting.
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	path, err := lookPath(cfg.Runtime)
	if err != nil || path == "" {
		if err == nil {
			err = errors.New("empty path")
		}
		return nil, &service.Down{
			Service: "prover",
			Reason:  fmt.Sprintf("Prover off: container runtime %q not found: %v", cfg.Runtime, err),
		}
	}
	return &Container{runtime: path, cfg: cfg}, nil
}

// Off runs no code. Reason is why the Prover is off. review.Open uses it when
// NewContainer reports the runtime missing (ADR-0025 point 6).
type Off struct{ Reason string }

// Run reports the Prover down and starts nothing.
func (o Off) Run(ctx context.Context, _ Request) (Outcome, error) {
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	return Outcome{}, &service.Down{Service: "prover", Reason: o.Reason}
}

// Argv is the runtime's argument list for running command in dir: the whole
// sentence handed to the runtime, so a test can read every property of the run.
// Run adds a unique --name so a timed-out container can be removed.
func (c *Container) Argv(dir, command string) []string {
	return c.argv(dir, command, "")
}

// argv builds the run line. Each flag below carries out ADR-0025: --rm and
// --network none, CPU and memory limits from the settings, and the checkout
// mounted read-only as the only volume. No -e or --env flag is ever added.
func (c *Container) argv(dir, command, name string) []string {
	if c == nil {
		return nil
	}
	// The runtime needs an absolute host path for a bind mount.
	host := dir
	if abs, err := filepath.Abs(dir); err == nil && abs != "" {
		host = abs
	}
	args := []string{c.runtime, "run"}
	if name != "" {
		args = append(args, "--name", name)
	}
	args = append(args,
		"--rm",
		"--network", "none",
		"--cpus", c.cfg.CPUs,
		"--memory", c.cfg.Memory,
		"-v", host+":"+checkoutDir+":ro",
		"--workdir", checkoutDir,
		c.cfg.Image,
		"sh", "-c", command,
	)
	return args
}

// runKind is how one side of a test run ended. Only kindPass and kindFail
// are test results; the others mean the side proves nothing.
type runKind int

const (
	kindPass runKind = iota
	kindFail
	kindNotRun
	kindTimeout
	kindCancelled
)

// Run runs req.Command once in req.BaseDir and once in req.ChangeDir, each in
// its own container, and reports whether it fails on the change and passes on
// the base. A run past its time limit is never proof. Exit 125, 126 and 127
// are the runtime's own failures (docker-run(1) Exit Status), not a test result.
func (c *Container) Run(ctx context.Context, req Request) (Outcome, error) {
	if c == nil || c.runtime == "" {
		return Outcome{}, &service.Down{Service: "prover", Reason: "Prover off: no container runtime"}
	}
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	out := Outcome{Path: req.Path, Name: req.Name, Command: req.Command}
	// A test the Prover wrote is checked before anything is copied or started.
	// A refusal is an outcome with Found false, not an error: the review goes on
	// and the trace says why the test did not run.
	if req.Content != "" {
		if reason := RefuseWritten(req.Path, req.Content, req.Command, req.BaseDir, req.ChangeDir, c.cfg.MaxWrittenBytes); reason != "" {
			out.Output = "refused: " + reason
			return out, nil
		}
		// The written test goes into a copy of each checkout, so the caller's
		// trees are never changed and both sides run the same test text.
		changeCopy, err := stageWritten(req.ChangeDir, req.Path, req.Content)
		if err != nil {
			return out, err
		}
		defer os.RemoveAll(changeCopy)
		baseCopy, err := stageWritten(req.BaseDir, req.Path, req.Content)
		if err != nil {
			return out, err
		}
		defer os.RemoveAll(baseCopy)
		req.ChangeDir = changeCopy
		req.BaseDir = baseCopy
	}
	if strings.TrimSpace(req.Command) == "" || req.BaseDir == "" || req.ChangeDir == "" {
		out.Output = "the test could not run: command, base directory and change directory are required"
		return out, nil
	}
	out.Found = true
	// The change runs first; a cancelled side stops the run and returns the
	// cancellation, with any container that may remain named in the error.
	changeText, changeKind := c.side(ctx, req.ChangeDir, req.Command)
	if changeKind == kindCancelled {
		return Outcome{}, withRemoval(ctx.Err(), changeText)
	}
	baseText, baseKind := c.side(ctx, req.BaseDir, req.Command)
	if baseKind == kindCancelled {
		return Outcome{}, withRemoval(ctx.Err(), changeText+"\n"+baseText)
	}
	out.FailsOnChange = changeKind == kindFail
	out.PassesOnBase = baseKind == kindPass
	out.Output = formatSides(changeText, changeKind, baseText, baseKind)
	return out, nil
}

// side runs command in one checkout and classifies how it ended.
func (c *Container) side(ctx context.Context, dir, command string) (string, runKind) {
	if err := ctx.Err(); err != nil {
		return "", kindCancelled
	}
	host, err := filepath.Abs(dir)
	// A colon would split the -v host:container:ro value into the wrong parts.
	if err != nil || strings.Contains(host, ":") {
		return "the test could not run: the checkout path cannot be one mount", kindNotRun
	}
	// A unique name lets remove find the container after a timeout or a cancel.
	name := containerName()
	argv := c.argv(dir, command, name)
	runCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()
	text, err := c.invoke(runCtx, argv)
	// The caller's context is checked before the run's own limit: a cancelled
	// review is not reported as a timeout. Either way the container is removed,
	// because killing the CLI does not stop the container it started.
	if ctx.Err() != nil {
		return joinRemoval(text, c.remove(name)), kindCancelled
	}
	if runCtx.Err() != nil {
		return joinRemoval(text, c.remove(name)), kindTimeout
	}
	if err == nil {
		return text, kindPass
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		switch exit.ExitCode() {
		case 125, 126, 127:
			return text, kindNotRun
		default:
			return text, kindFail
		}
	}
	// No exit status at all: the runtime could not be started, so the test did not run.
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + err.Error(), kindNotRun
}

// remove asks the runtime to delete name. Its own context is cleanupBound
// long and is not the run's context: a timed-out run must still be removed,
// and that removal must not add the old separate 10 second wait. A non-nil
// error means the container may still be there; the caller puts that into
// the outcome.
func (c *Container) remove(name string) string {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupBound)
	defer cancel()
	_, err := c.invoke(ctx, []string{c.runtime, "rm", "-f", name})
	if err == nil {
		return ""
	}
	return fmt.Sprintf("container %s may remain: removal failed: %s", name, err.Error())
}

// joinRemoval puts a failed removal's note in front of the side's output.
func joinRemoval(text, removal string) string {
	if removal == "" {
		return text
	}
	// The removal note goes first: trimOutput keeps only the first 8192 bytes of a side's output.
	if text == "" {
		return removal
	}
	return removal + "\n" + text
}

// withRemoval adds a "may remain" note to a cancellation error, so a
// cancelled review still says that a container may be left on the host.
func withRemoval(err error, text string) error {
	if err == nil {
		err = context.Canceled
	}
	if !strings.Contains(text, "may remain") {
		return err
	}
	return fmt.Errorf("%w\n%s", err, strings.TrimSpace(text))
}

// invoke runs one runtime command. Tests set c.run to record the argv
// without a runtime.
func (c *Container) invoke(ctx context.Context, argv []string) (string, error) {
	if c.run != nil {
		return c.run(ctx, argv)
	}
	return commandOutput(ctx, argv)
}

// commandOutput runs argv and returns stdout and stderr together, in the order printed.
func commandOutput(ctx context.Context, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", errors.New("testrun: empty argv")
	}
	// argv carries no -e and no --env, so the container does not receive the
	// process environment. The runtime process itself keeps it, because that
	// is how it finds its own daemon socket.
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// nameSeq numbers the containers of this process, so concurrent runs never share a name.
var nameSeq atomic.Uint64

// containerName returns a name unique to this process and run.
func containerName() string {
	return fmt.Sprintf("proven-%d-%d", os.Getpid(), nameSeq.Add(1))
}

// formatSides writes both sides into Outcome.Output, and says so when a side
// timed out or could not run, so the trace never shows such a run as proof.
func formatSides(changeText string, changeKind runKind, baseText string, baseKind runKind) string {
	var b strings.Builder
	fmt.Fprintf(&b, "change: %s\n%s\nbase: %s\n%s", sideWord(changeKind), trimOutput(changeText), sideWord(baseKind), trimOutput(baseText))
	if changeKind == kindTimeout || baseKind == kindTimeout {
		b.WriteString("\na run past its time limit is not proof")
	}
	if changeKind == kindNotRun || baseKind == kindNotRun {
		b.WriteString("\nthe test could not run")
	}
	return strings.TrimSpace(b.String())
}

// sideWord is how a side's end reads in the output.
func sideWord(k runKind) string {
	switch k {
	case kindPass:
		return "passed"
	case kindFail:
		return "failed"
	case kindTimeout:
		return "passed its time limit and is not proof"
	case kindNotRun:
		return "could not run"
	default:
		return "not run"
	}
}

// trimOutput bounds a side's output, so a noisy test cannot fill the trace.
func trimOutput(s string) string {
	s = strings.TrimSpace(s)
	const max = 8192
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n[truncated]"
}
