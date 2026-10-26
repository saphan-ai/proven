package testrun

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/network"
)

// Environment variables of the live container. The set is closed: Load reads
// these names and no others. An empty value means the default from
// DefaultContainerConfig, which is the one list of those defaults.
const (
	EnvRuntime         = "PROVEN_TEST_RUNTIME"
	EnvImage           = "PROVEN_TEST_IMAGE"
	EnvCPUs            = "PROVEN_TEST_CPUS"
	EnvMemory          = "PROVEN_TEST_MEMORY"
	EnvTimeout         = "PROVEN_TEST_TIMEOUT"
	EnvMaxWrittenBytes = "PROVEN_TEST_MAX_WRITTEN_BYTES"
	EnvUser            = "PROVEN_TEST_USER"
	EnvPidsLimit       = "PROVEN_TEST_PIDS"
	EnvTmpSize         = "PROVEN_TEST_TMP_SIZE"
)

// ConfigError names a refused container setting and the values that are allowed.
// Field is the Go field when Validate refuses the struct, and the environment
// variable when Load refuses a setting. A refusal does not start a container.
// There is no setting that gives the container a network or a secret (ADR-0013).
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the environment variable, the refused value, the reason and the allowed values.
func (e *ConfigError) Error() string {
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

// Load reads the live container settings (ADR-0013). A wrong value is refused.
// An empty lookup uses DefaultContainerConfig. Validate has accepted the result.
// A refusal is a *ConfigError that names the environment variable, so the
// shell can tell the user which setting to fix; no container is started.
func Load(env func(string) string) (ContainerConfig, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	cfg := DefaultContainerConfig()
	// A wrong PROVEN_NETWORK is refused before a container starts. The image is
	// never pulled: --pull never is on every run, so there is no PullNever switch.
	policy, perr := network.Load(env)
	if perr != nil {
		if ne, ok := network.As(perr); ok {
			return ContainerConfig{}, &ConfigError{Field: ne.Field, Value: ne.Value, Reason: ne.Reason, Allowed: ne.Allowed}
		}
		return ContainerConfig{}, perr
	}
	cfg.local = policy.Local()
	// The daemon the runtime talks to is a connection the HTTP dial check never sees.
	if derr := checkDaemon(policy, env); derr != nil {
		return ContainerConfig{}, derr
	}
	if raw := env(EnvRuntime); raw != "" {
		cfg.Runtime = raw
	}
	// Each runtime finds its daemon its own way. One Proven cannot judge is refused.
	if rerr := checkRuntime(policy, env, cfg.Runtime); rerr != nil {
		return ContainerConfig{}, rerr
	}
	if raw := env(EnvImage); raw != "" {
		cfg.Image = raw
	}
	if raw := env(EnvCPUs); raw != "" {
		cfg.CPUs = raw
	}
	if raw := env(EnvMemory); raw != "" {
		cfg.Memory = raw
	}
	if raw := env(EnvUser); raw != "" {
		cfg.User = raw
	}
	if raw := env(EnvTmpSize); raw != "" {
		cfg.TmpSize = raw
	}
	// Timeout and size are parsed here because the type changes; a value that
	// parses but is not positive is refused by Validate below.
	if raw := env(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return ContainerConfig{}, refuse(EnvTimeout, clip(raw), "not a Go duration greater than zero", timeoutAllowed())
		}
		cfg.Timeout = d
	}
	if raw := env(EnvMaxWrittenBytes); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return ContainerConfig{}, refuse(EnvMaxWrittenBytes, clip(raw), "not a positive number of bytes", bytesAllowed())
		}
		cfg.MaxWrittenBytes = n
	}
	if raw := env(EnvPidsLimit); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return ContainerConfig{}, refuse(EnvPidsLimit, clip(raw), "not a positive number of processes", pidsAllowed())
		}
		cfg.PidsLimit = n
	}
	if err := cfg.Validate(); err != nil {
		return ContainerConfig{}, fromSetting(err)
	}
	return cfg, nil
}

// timeoutAllowed and bytesAllowed describe the allowed form of a value that did not parse.
func timeoutAllowed() string {
	shown := defaultTimeout.String()
	return "a Go duration greater than zero, for example " + shown + "; empty means " + shown
}

func bytesAllowed() string {
	return fmt.Sprintf("a positive integer; empty means %d", DefaultMaxWrittenBytes)
}

func pidsAllowed() string {
	return fmt.Sprintf("a positive integer of at most %d; empty means %d", maxPidsLimit, defaultPids)
}

// refuse builds a ConfigError with one allowed form.
func refuse(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

// fromSetting renames a ConfigError's field to its environment variable,
// so the user sees the name they set, not the Go field.
func fromSetting(err error) error {
	var se *ConfigError
	if !errors.As(err, &se) || se == nil {
		return err
	}
	return &ConfigError{
		Field:   envName(se.Field),
		Value:   clip(se.Value),
		Reason:  se.Reason,
		Allowed: append([]string(nil), se.Allowed...),
	}
}

// envName maps a ContainerConfig field to the variable that sets it.
func envName(field string) string {
	switch field {
	case "Runtime":
		return EnvRuntime
	case "Image":
		return EnvImage
	case "Timeout":
		return EnvTimeout
	case "CPUs":
		return EnvCPUs
	case "Memory":
		return EnvMemory
	case "MaxWrittenBytes":
		return EnvMaxWrittenBytes
	case "User":
		return EnvUser
	case "PidsLimit":
		return EnvPidsLimit
	case "TmpSize":
		return EnvTmpSize
	default:
		return field
	}
}

// clip bounds a refused value at 128 bytes, so a huge value cannot flood the error.
func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}

// Environment lists the live container settings in the order the page shows.
// Each Name is the Env constant above; each Default is DefaultContainerConfig's.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{Name: EnvRuntime, Default: envdoc.Tick(defaultRuntime), Description: "the container command line, by name or absolute path"},
		{Name: EnvImage, Default: envdoc.Tick(defaultImage), Description: "the image each test runs in"},
		{Name: EnvCPUs, Default: envdoc.Tick(defaultCPUs), Description: "the CPU limit of one test run"},
		{Name: EnvMemory, Default: envdoc.Tick(defaultMemory), Description: "the memory limit of one test run; a positive size with suffix b, k, m or g"},
		{Name: EnvTimeout, Default: envdoc.Tick(envdoc.SecondSpec(defaultTimeout)), Description: "the limit for one test run, base or change"},
		{Name: EnvMaxWrittenBytes, Default: envdoc.Tick(strconv.Itoa(DefaultMaxWrittenBytes)), Description: "the largest test file the Prover may write"},
		{Name: EnvUser, Default: envdoc.Tick(defaultUser), Description: "the unprivileged user and group the test runs as, as numbers; root is refused"},
		{Name: EnvPidsLimit, Default: envdoc.Tick(strconv.Itoa(defaultPids)), Description: "the most processes one test run may start; at most 4096"},
		{Name: EnvTmpSize, Default: envdoc.Tick(defaultTmpSize), Description: "the size limit of the writable scratch mount; a positive size of at most 1g with suffix b, k, m or g"},
	}
}
