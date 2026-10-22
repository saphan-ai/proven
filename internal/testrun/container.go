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

const (
	defaultRuntime = "docker"
	defaultImage   = "golang:1.26"
	defaultCPUs    = "1"
	defaultMemory  = "512m"
	checkoutDir    = "/checkout"
)

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
	Runtime string        // container CLI name or path; default "docker"
	Image   string        // image the test runs in
	Timeout time.Duration // limit for one test run, base or change
	CPUs    string        // CPU limit passed to the runtime
	Memory  string        // memory limit passed to the runtime
}

var (
	cpuPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	memPattern = regexp.MustCompile(`^([0-9]+)([bkmg])?$`)
)

// SettingError is a refused container setting (ADR-0013). There is no
// setting that gives the container a network or a secret.
type SettingError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

func (e *SettingError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := fmt.Sprintf("configuration refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// DefaultContainerConfig returns the documented defaults.
func DefaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		Runtime: defaultRuntime,
		Image:   defaultImage,
		Timeout: defaultTimeout,
		CPUs:    defaultCPUs,
		Memory:  defaultMemory,
	}
}

// Validate refuses a setting by name with the allowed values (ADR-0013).
func (c ContainerConfig) Validate() error {
	if !validToken(c.Runtime) {
		return &SettingError{
			Field: "Runtime", Value: c.Runtime,
			Reason:  "the runtime is a CLI name or path, with no spaces and no leading dash",
			Allowed: []string{defaultRuntime},
		}
	}
	if !validToken(c.Image) {
		return &SettingError{
			Field: "Image", Value: c.Image,
			Reason:  "the image is a reference, with no spaces and no leading dash",
			Allowed: []string{defaultImage},
		}
	}
	if c.Timeout <= 0 {
		return &SettingError{
			Field: "Timeout", Value: c.Timeout.String(),
			Reason:  "the time limit is a positive duration; a run past it is not proof",
			Allowed: []string{defaultTimeout.String()},
		}
	}
	if !validCPUs(c.CPUs) {
		return &SettingError{
			Field: "CPUs", Value: c.CPUs,
			Reason:  "the CPU limit is a positive number; 0 means no limit and is refused",
			Allowed: []string{defaultCPUs, "0.5"},
		}
	}
	if !validMemory(c.Memory) {
		return &SettingError{
			Field: "Memory", Value: c.Memory,
			Reason:  "the memory limit is a positive size with suffix b, k, m or g; 0 is refused",
			Allowed: []string{defaultMemory},
		}
	}
	return nil
}

func validToken(s string) bool {
	if s == "" || strings.HasPrefix(s, "-") {
		return false
	}
	return !strings.ContainsAny(s, " \t\n\r")
}

func validCPUs(s string) bool {
	if !cpuPattern.MatchString(s) {
		return false
	}
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && f > 0
}

func validMemory(s string) bool {
	m := memPattern.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	n, err := strconv.ParseUint(m[1], 10, 64)
	return err == nil && n > 0
}

// Container is the live test-run adapter.
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

// Off runs no code. Reason is why the Prover is off.
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

func (c *Container) argv(dir, command, name string) []string {
	if c == nil {
		return nil
	}
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
	if strings.TrimSpace(req.Command) == "" || req.BaseDir == "" || req.ChangeDir == "" {
		out.Output = "the test could not run: command, base directory and change directory are required"
		return out, nil
	}
	out.Found = true
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

func (c *Container) side(ctx context.Context, dir, command string) (string, runKind) {
	if err := ctx.Err(); err != nil {
		return "", kindCancelled
	}
	host, err := filepath.Abs(dir)
	if err != nil || strings.Contains(host, ":") {
		return "the test could not run: the checkout path cannot be one mount", kindNotRun
	}
	name := containerName()
	argv := c.argv(dir, command, name)
	runCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()
	text, err := c.invoke(runCtx, argv)
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

func joinRemoval(text, removal string) string {
	if removal == "" {
		return text
	}
	// The removal note goes first: the outcome keeps only the first 8192 bytes of output.
	if text == "" {
		return removal
	}
	return removal + "\n" + text
}

func withRemoval(err error, text string) error {
	if err == nil {
		err = context.Canceled
	}
	if !strings.Contains(text, "may remain") {
		return err
	}
	return fmt.Errorf("%w\n%s", err, strings.TrimSpace(text))
}

func (c *Container) invoke(ctx context.Context, argv []string) (string, error) {
	if c.run != nil {
		return c.run(ctx, argv)
	}
	return commandOutput(ctx, argv)
}

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

var nameSeq atomic.Uint64

func containerName() string {
	return fmt.Sprintf("proven-%d-%d", os.Getpid(), nameSeq.Add(1))
}

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

func trimOutput(s string) string {
	s = strings.TrimSpace(s)
	const max = 8192
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n[truncated]"
}
