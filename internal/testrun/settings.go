package testrun

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
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
)

// ConfigError names a refused container setting and the values that are allowed.
// Field is the environment variable. A refusal does not start a container.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

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

// Load reads the live container settings. A wrong value is refused.
// An empty lookup uses DefaultContainerConfig. Validate has accepted the result.
func Load(env func(string) string) (ContainerConfig, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	cfg := DefaultContainerConfig()
	if raw := env(EnvRuntime); raw != "" {
		cfg.Runtime = raw
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
	if err := cfg.Validate(); err != nil {
		return ContainerConfig{}, fromSetting(err)
	}
	return cfg, nil
}

func timeoutAllowed() string {
	shown := defaultTimeout.String()
	return "a Go duration greater than zero, for example " + shown + "; empty means " + shown
}

func bytesAllowed() string {
	return fmt.Sprintf("a positive integer; empty means %d", DefaultMaxWrittenBytes)
}

func refuse(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

func fromSetting(err error) error {
	var se *SettingError
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
	default:
		return field
	}
}

func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
