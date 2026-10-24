package testrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/saphan-ai/proven/internal/childenv"
	"github.com/saphan-ai/proven/internal/service"
)

// The flags below are the ones docker-run(1) from Docker 29.9.1 documents.
// --network none is "No networking in the container". --rm removes the
// container when it exits. --cpus takes a number and 0 means no limit.
// --memory takes a number and a suffix b, k, m or g, and 0 means no limit.
// A volume's ro option mounts it read-only; the default is read-write.
// --pull never stops a missing image from being pulled on the host before
// the container starts; the run then fails and Run reports the Prover down.

// The documented defaults of the container settings (ADR-0013, ADR-0025
// point 5). checkoutDir is where the one checkout is mounted inside.
const (
	defaultRuntime = "docker"
	defaultImage   = "golang:1.26"
	defaultCPUs    = "1"
	defaultMemory  = "512m"
	checkoutDir    = "/checkout"

	// defaultUser is "nobody" and its group, a fixed unprivileged identity that
	// exists in common images and owns nothing in the checkout (ADR-0025).
	defaultUser = "65534:65534"
	// defaultPids bounds the processes of one container, so a runaway test
	// cannot exhaust the host's process table.
	defaultPids = 256
	// defaultTmpSize bounds each writable scratch mount. A Go test build
	// measured about 135 MB on a cold cache (compiled packages, the build's
	// work directory and the test binary), so the default leaves a margin of
	// about four times; maxTmpBytes stays the upper bound.
	defaultTmpSize = "512m"

	// maxPidsLimit and maxTmpBytes are the upper bounds of the PidsLimit and
	// TmpSize settings, so no setting can turn a bound into a practically open
	// one (ADR-0025 point 5). The scratch bound is 1 GiB per mount.
	maxPidsLimit = 4096
	maxTmpBytes  = 1 << 30
)

// scratchDirs are the only writable places in the container beside the
// checkout's own read-only mount. A Go test builds and runs its binary under
// /tmp, and its build cache lives there too (see containerEnv).
var scratchDirs = []string{"/tmp"}

// containerEnv is the whole environment the container receives, passed as -e.
// The values are constants, never read from Proven's environment, and none is
// a secret. They put the home and the Go build cache on scratch, because the
// root is read-only and the image's own home for the unprivileged user (for the
// default image, a path that does not exist) is not writable. This amends the
// "no -e" contract of ADR-0025 point 3: no secret and no inherited value is
// passed; these two fixed values are.
var containerEnv = []string{"HOME=/tmp", "GOCACHE=/tmp/go-build"}

// runtimeEnvNames are the variables, beyond PATH and HOME, the runtime process
// may inherit. They only say where the container daemon is; none holds a
// credential (ADR-0025).
var runtimeEnvNames = []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "XDG_RUNTIME_DIR"}

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
	User            string        // fixed unprivileged user and group, "uid:gid"
	PidsLimit       int           // most processes one container may run
	TmpSize         string        // size bound of each writable scratch mount
}

// cpuPattern and memPattern accept only the plain forms docker-run(1)
// documents, so a value can never smuggle in another flag.
var (
	cpuPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	memPattern = regexp.MustCompile(`^([0-9]+)([bkmg])?$`)
	// userPattern accepts a numeric uid and gid only; root is refused by validUser.
	userPattern = regexp.MustCompile(`^([0-9]+):([0-9]+)$`)
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
		User:            defaultUser,
		PidsLimit:       defaultPids,
		TmpSize:         defaultTmpSize,
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
	// A setting must not be able to hand the test root, lift the process bound
	// or make the scratch unbounded (ADR-0025).
	if !validUser(c.User) {
		return &ConfigError{
			Field: "User", Value: c.User,
			Reason:  "the user is a numeric uid:gid that is not root",
			Allowed: []string{defaultUser},
		}
	}
	if c.PidsLimit <= 0 || c.PidsLimit > maxPidsLimit {
		return &ConfigError{
			Field: "PidsLimit", Value: strconv.Itoa(c.PidsLimit),
			Reason:  "the process limit is a positive number of at most " + strconv.Itoa(maxPidsLimit) + "; 0 and -1 mean no limit and are refused",
			Allowed: []string{strconv.Itoa(defaultPids)},
		}
	}
	if !validMemory(c.TmpSize) || sizeBytes(c.TmpSize) > maxTmpBytes {
		return &ConfigError{
			Field: "TmpSize", Value: c.TmpSize,
			Reason:  "the scratch size is a positive size of at most 1g with suffix b, k, m or g; 0 is refused",
			Allowed: []string{defaultTmpSize},
		}
	}
	return nil
}

// validUser accepts a numeric uid:gid where neither is 0, the root identity.
func validUser(s string) bool {
	m := userPattern.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	uid, uerr := strconv.ParseUint(m[1], 10, 32)
	gid, gerr := strconv.ParseUint(m[2], 10, 32)
	return uerr == nil && gerr == nil && uid > 0 && gid > 0
}

// sizeBytes returns the bytes of a size that validMemory accepted, and
// math.MaxInt64 for one too large to count, so the caller's upper bound
// refuses it. A bare number is bytes, as in docker-run(1).
func sizeBytes(s string) int64 {
	m := memPattern.FindStringSubmatch(s)
	if m == nil {
		return math.MaxInt64
	}
	n, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return math.MaxInt64
	}
	unit := uint64(1)
	switch m[2] {
	case "k":
		unit = 1 << 10
	case "m":
		unit = 1 << 20
	case "g":
		unit = 1 << 30
	}
	if n > math.MaxInt64/unit {
		return math.MaxInt64
	}
	return int64(n * unit)
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
// fresh container with no network, a fixed non-secret environment, an unprivileged user, a
// read-only root and one read-only checkout (ADR-0025 points 1 to 4). review.Open builds it from Load.
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

// RunFix reports the Prover down and starts nothing, as Run does.
func (o Off) RunFix(ctx context.Context, _ FixRequest) (FixOutcome, error) {
	if err := ctx.Err(); err != nil {
		return FixOutcome{}, err
	}
	return FixOutcome{}, &service.Down{Service: "prover", Reason: o.Reason}
}

// Argv is the runtime's argument list for running the test command words in
// dir: the whole sentence handed to the runtime, so a test can read every
// property of the run. Run adds a unique --name so a timed-out container can
// be removed.
func (c *Container) Argv(dir string, test []string) []string {
	return c.argv(dir, test, "")
}

// How one RunCommand ended. A timeout or a side that did not run is not a
// result of the command (ADR-0040: a tool that runs out of time is named).
const (
	CommandPass      = "pass"
	CommandFail      = "fail"
	CommandNotRun    = "not-run"
	CommandTimeout   = "timeout"
	CommandCancelled = "cancelled"
)

// CommandResult is one run of a command Proven built, in one checkout.
type CommandResult struct {
	Output string
	Status string
}

// WithTimeout returns a copy whose limit for one side is d.
// ADR-0040: a tool's limit is its own setting and does not change the test
// runner that shares this container. A non-positive d is refused: a zero
// limit would make every run look timed out.
func (c *Container) WithTimeout(d time.Duration) (*Container, error) {
	if c == nil {
		return nil, errors.New("testrun: no container")
	}
	if d <= 0 {
		return nil, &ConfigError{
			Field:   "Timeout",
			Value:   d.String(),
			Reason:  "the time limit is a positive duration; a run past it is not proof",
			Allowed: []string{defaultTimeout.String()},
		}
	}
	cp := *c
	cp.cfg.Timeout = d
	return &cp, nil
}

// RunCommand runs argv inside the locked-down container, with dir as the one
// checkout. argv is the words after the image. There is no shell (ADR-0025).
// A cancelled review returns the context error. A timeout is a result, not an
// error: the caller names it and does not treat it as a clean run (ADR-0040).
// An absent image is the Prover down.
func (c *Container) RunCommand(ctx context.Context, dir string, argv []string) (CommandResult, error) {
	if c == nil || c.runtime == "" {
		return CommandResult{}, &service.Down{Service: "prover", Reason: "Prover off: no container runtime"}
	}
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	if len(argv) == 0 {
		return CommandResult{Status: CommandNotRun, Output: "refused: empty command"}, nil
	}
	if dir == "" {
		return CommandResult{Status: CommandNotRun, Output: "the command could not run: a checkout directory is required"}, nil
	}
	// ADR-0025: a checkout that is a link is not a mount source.
	if info, err := os.Lstat(filepath.Clean(dir)); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return CommandResult{Status: CommandNotRun, Output: "the command could not run: a checkout is a symbolic link and is not mounted"}, nil
	}
	text, kind := c.side(ctx, dir, argv)
	if kind == kindCancelled {
		return CommandResult{}, withRemoval(ctx.Err(), text)
	}
	if down := c.absentImage(text, kind); down != nil {
		return CommandResult{}, down
	}
	return CommandResult{Output: text, Status: commandStatus(kind)}, nil
}

// commandStatus is the public name of one side's ending.
func commandStatus(k runKind) string {
	switch k {
	case kindPass:
		return CommandPass
	case kindFail:
		return CommandFail
	case kindTimeout:
		return CommandTimeout
	case kindCancelled:
		return CommandCancelled
	default:
		return CommandNotRun
	}
}

// argv builds the run line. Each flag below carries out ADR-0025: --rm and
// --network none, CPU and memory limits from the settings, and the checkout
// mounted read-only as the only volume. The only -e flags are the fixed,
// non-secret values of containerEnv; nothing of Proven's environment is passed.
// The container also runs as a fixed unprivileged user with a process limit,
// no capabilities, no privilege gain and a read-only root; the bounded tmpfs
// mounts are the only other place it may write.
func (c *Container) argv(dir string, test []string, name string) []string {
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
		// ADR-0025: the review never fetches an image at test time. An absent image is the
		// Prover down, and Run names the pull the owner does by hand.
		"--pull", "never",
		"--cpus", c.cfg.CPUs,
		"--memory", c.cfg.Memory,
		"--user", c.cfg.User,
		"--pids-limit", strconv.Itoa(c.cfg.PidsLimit),
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--read-only",
	)
	for _, dir := range scratchDirs {
		// The runtimes mount a tmpfs noexec unless told otherwise, so exec is
		// named: a Go test runs the binary it builds on scratch. nosuid and nodev
		// keep a file written there from gaining privilege or acting as a device.
		// Mode 1777 lets the unprivileged user write to a mount the runtime
		// creates as root.
		args = append(args, "--tmpfs", dir+":rw,exec,nosuid,nodev,mode=1777,size="+c.cfg.TmpSize)
	}
	for _, e := range containerEnv {
		args = append(args, "-e", e)
	}
	args = append(args,
		"-v", host+":"+checkoutDir+":ro",
		"--workdir", checkoutDir,
		c.cfg.Image,
	)
	// ADR-0025: the words after the image are the test command Proven built.
	// They stay separate argv words and no shell reads them.
	return append(args, test...)
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

// Run runs the test req names once in req.BaseDir and once in req.ChangeDir, each in
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
	out := Outcome{Path: req.Path, Name: req.Name}
	// A refusal is an outcome with Found false, not an error: the review goes on
	// and the trace says why the test did not run.
	// A test the Prover wrote is checked against both checkouts before anything
	// is copied or started.
	if req.Content != "" {
		if reason := RefuseWritten(req.Path, req.Content, req.Name, req.BaseDir, req.ChangeDir, c.cfg.MaxWrittenBytes); reason != "" {
			out.Output = "refused: " + reason
			return out, nil
		}
	}
	// ADR-0025: Proven builds the command from the test's file and name. A name
	// or path that fails the checks gets no command, so no container is started.
	test, reason := TestArgv(req.Path, req.Name)
	if reason != "" {
		out.Output = "refused: " + reason
		return out, nil
	}
	out.Command = displayCommand(test)
	// ADR-0025: a checkout that is a link is not a mount source. The runtime would follow it to
	// whatever it names, which is not the tree that was prepared.
	if linkedCheckout(req.BaseDir, req.ChangeDir) {
		out.Output = "the test could not run: a checkout is a symbolic link and is not mounted"
		return out, nil
	}
	if req.Content != "" {
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
	if req.BaseDir == "" || req.ChangeDir == "" {
		out.Output = "the test could not run: base directory and change directory are required"
		return out, nil
	}
	out.Found = true
	// The change runs first; a cancelled side stops the run and returns the
	// cancellation, with any container that may remain named in the error.
	changeText, changeKind := c.side(ctx, req.ChangeDir, test)
	if changeKind == kindCancelled {
		return Outcome{}, withRemoval(ctx.Err(), changeText)
	}
	if down := c.absentImage(changeText, changeKind); down != nil {
		return Outcome{}, down
	}
	baseText, baseKind := c.side(ctx, req.BaseDir, test)
	if baseKind == kindCancelled {
		return Outcome{}, withRemoval(ctx.Err(), changeText+"\n"+baseText)
	}
	out.FailsOnChange = changeKind == kindFail
	out.PassesOnBase = baseKind == kindPass
	out.Output = formatSides(changeText, changeKind, baseText, baseKind)
	return out, nil
}

// linkedCheckout reports whether any named checkout is a symbolic link. The runtime
// would follow it to whatever it names, which is not the tree that was prepared.
// An empty name is skipped: the caller refuses a missing directory on its own.
func linkedCheckout(dirs ...string) bool {
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		// The cleaned path is the one checked: a trailing slash makes Lstat follow
		// the link, and the cleaned form is what names the mount source.
		if info, err := os.Lstat(filepath.Clean(dir)); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return true
		}
	}
	return false
}

// absentImage turns a run the runtime refused for lack of the image into the Prover down, with the
// pull as its next act, the way Probe does. It returns nil for any other outcome. Docker says
// "Unable to find image" and podman says "image not known"; --pull never makes both refuse
// instead of fetching (ADR-0025).
func (c *Container) absentImage(text string, kind runKind) error {
	if kind != kindNotRun {
		return nil
	}
	lower := strings.ToLower(text)
	if !strings.Contains(lower, "unable to find image") && !strings.Contains(lower, "image not known") {
		return nil
	}
	return &service.Down{
		Service: "prover",
		Reason:  "image " + c.cfg.Image + " is not present; next: " + filepath.Base(c.runtime) + " pull " + c.cfg.Image,
	}
}

// side runs the test command words in one checkout and classifies how it ended.
func (c *Container) side(ctx context.Context, dir string, test []string) (string, runKind) {
	text, kind, _ := c.sideExit(ctx, dir, test)
	return text, kind
}

// noExit is the exit code reported for a side that ended without one: it timed
// out, was cancelled, or the runtime could not be started.
const noExit = -1

// sideExit is side with the exit code of the command, which a fix's record keeps
// so a reader can compare it with a run of their own (ADR-0042 point 7).
func (c *Container) sideExit(ctx context.Context, dir string, test []string) (string, runKind, int) {
	if err := ctx.Err(); err != nil {
		return "", kindCancelled, noExit
	}
	host, err := filepath.Abs(dir)
	// A colon would split the -v host:container:ro value into the wrong parts.
	if err != nil || strings.Contains(host, ":") {
		return "the test could not run: the checkout path cannot be one mount", kindNotRun, noExit
	}
	// A unique name lets remove find the container after a timeout or a cancel.
	name := containerName()
	argv := c.argv(dir, test, name)
	runCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()
	text, err := c.invoke(runCtx, argv)
	// The caller's context is checked before the run's own limit: a cancelled
	// review is not reported as a timeout. Either way the container is removed,
	// because killing the CLI does not stop the container it started.
	if ctx.Err() != nil {
		return joinRemoval(text, c.remove(name)), kindCancelled, noExit
	}
	if runCtx.Err() != nil {
		return joinRemoval(text, c.remove(name)), kindTimeout, noExit
	}
	proof := isProofArgv(test)
	if err == nil {
		if !proof {
			return text, kindPass, 0
		}
		// ADR-0042: exit code 0 is not enough; the output must show the test ran and passed.
		shown, passed := judgeProof(test, text)
		if !passed {
			return shown, kindNotRun, 0
		}
		return shown, kindPass, 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		switch exit.ExitCode() {
		case 125, 126, 127:
			return text, kindNotRun, exit.ExitCode()
		default:
			if proof {
				_, text = parseEvents(text)
			}
			return text, kindFail, exit.ExitCode()
		}
	}
	// No exit status at all: the runtime could not be started, so the test did not run.
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + err.Error(), kindNotRun, noExit
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
	// argv carries only the fixed -e values of containerEnv, so the container
	// does not receive the process environment. The runtime process gets an explicit minimal one
	// too (ADR-0025): PATH, HOME and the names that locate its daemon, never
	// a key, because its output can reach the trace.
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = childenv.Minimal(runtimeEnvNames...)
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
