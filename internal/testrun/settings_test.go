package testrun

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestLoadNilEnvUsesDefaults proves an empty environment gives exactly the documented defaults.
func TestLoadNilEnvUsesDefaults(t *testing.T) {
	cfg, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg != DefaultContainerConfig() {
		t.Fatalf("%#v", cfg)
	}
}

// TestLoadRuntime proves PROVEN_TEST_RUNTIME is read, and a flag-like value is refused by name.
func TestLoadRuntime(t *testing.T) {
	testStringSetting(t, EnvRuntime, "docker", "/usr/bin/podman", "-rf")
}

// Mutated field: PROVEN_TEST_IMAGE. Ignoring that variable turns "other" and
// "refused" red.
// It proves PROVEN_TEST_IMAGE is read, and a value with a space is refused by name.
func TestLoadImage(t *testing.T) {
	testStringSetting(t, EnvImage, "golang:1.26", "alpine:3", "bad image")
}

// TestLoadCPUs proves PROVEN_TEST_CPUS is read, and a word is refused by name.
func TestLoadCPUs(t *testing.T) {
	testStringSetting(t, EnvCPUs, "1", "2", "one")
}

// TestLoadMemory proves PROVEN_TEST_MEMORY is read, and a word is refused by name.
func TestLoadMemory(t *testing.T) {
	testStringSetting(t, EnvMemory, "512m", "256m", "lots")
}

// TestLoadTimeout proves PROVEN_TEST_TIMEOUT is read as a duration, and zero
// or a word is refused by name.
func TestLoadTimeout(t *testing.T) {
	cases := []struct {
		name  string
		set   bool
		value string
		want  time.Duration
		bad   bool
	}{
		{name: "default", want: 2 * time.Minute},
		{name: "other", set: true, value: "30s", want: 30 * time.Second},
		{name: "refused", set: true, value: "0s", bad: true},
		{name: "refused word", set: true, value: "lots", bad: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(settingEnv(EnvTimeout, tc.set, tc.value))
			if tc.bad {
				assertRefused(t, err, EnvTimeout, tc.value)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := DefaultContainerConfig()
			want.Timeout = tc.want
			if cfg != want {
				t.Fatalf("%#v", cfg)
			}
		})
	}
}

// TestLoadMaxWrittenBytes proves PROVEN_TEST_MAX_WRITTEN_BYTES is read, and
// zero or a word is refused by name.
func TestLoadMaxWrittenBytes(t *testing.T) {
	cases := []struct {
		name  string
		set   bool
		value string
		want  int
		bad   bool
	}{
		{name: "default", want: DefaultMaxWrittenBytes},
		{name: "other", set: true, value: "1024", want: 1024},
		{name: "refused", set: true, value: "0", bad: true},
		{name: "refused word", set: true, value: "lots", bad: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(settingEnv(EnvMaxWrittenBytes, tc.set, tc.value))
			if tc.bad {
				assertRefused(t, err, EnvMaxWrittenBytes, tc.value)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := DefaultContainerConfig()
			want.MaxWrittenBytes = tc.want
			if cfg != want {
				t.Fatalf("%#v", cfg)
			}
		})
	}
}

// testStringSetting checks one string setting three ways: unset gives def,
// set gives other, and bad is refused by the variable's name.
func testStringSetting(t *testing.T, key, def, other, bad string) {
	t.Helper()
	cases := []struct {
		name  string
		set   bool
		value string
		want  string
		bad   bool
	}{
		{name: "default", want: def},
		{name: "other", set: true, value: other, want: other},
		{name: "refused", set: true, value: bad, bad: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(settingEnv(key, tc.set, tc.value))
			if tc.bad {
				assertRefused(t, err, key, tc.value)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := settingValue(cfg, key)
			if got != tc.want {
				t.Fatalf("%s %q", key, got)
			}
			want := DefaultContainerConfig()
			setSettingValue(&want, key, tc.want)
			if cfg != want {
				t.Fatalf("%#v", cfg)
			}
		})
	}
}

func settingEnv(key string, set bool, value string) func(string) string {
	if !set {
		return func(string) string { return "" }
	}
	return func(k string) string {
		if k == key {
			return value
		}
		return ""
	}
}

func settingValue(cfg ContainerConfig, key string) string {
	switch key {
	case EnvRuntime:
		return cfg.Runtime
	case EnvImage:
		return cfg.Image
	case EnvCPUs:
		return cfg.CPUs
	case EnvMemory:
		return cfg.Memory
	default:
		return ""
	}
}

func setSettingValue(cfg *ContainerConfig, key, value string) {
	switch key {
	case EnvRuntime:
		cfg.Runtime = value
	case EnvImage:
		cfg.Image = value
	case EnvCPUs:
		cfg.CPUs = value
	case EnvMemory:
		cfg.Memory = value
	}
}

// assertRefused checks that err is a *ConfigError naming field and value.
func assertRefused(t *testing.T, err error, field, value string) {
	t.Helper()
	var cfg *ConfigError
	if !errors.As(err, &cfg) {
		t.Fatalf("got %v", err)
	}
	if cfg.Field != field || cfg.Value != value {
		t.Fatalf("%#v", cfg)
	}
	if len(cfg.Allowed) == 0 || !strings.Contains(err.Error(), "allowed") {
		t.Fatalf("%v", err)
	}
	if !strings.Contains(err.Error(), field) {
		t.Fatalf("%v", err)
	}
}
