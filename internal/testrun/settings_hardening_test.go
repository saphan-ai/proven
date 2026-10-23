package testrun

import "testing"

// settingFrom is an environment lookup that holds one variable.
func settingFrom(key, value string) func(string) string {
	return func(k string) string {
		if k == key {
			return value
		}
		return ""
	}
}

// TestLoadHardeningSettingsAreReadAndRefused proves PROVEN_TEST_USER,
// PROVEN_TEST_PIDS and PROVEN_TEST_TMP_SIZE are read, and that the values
// that would lift a bound (root, no process limit, no scratch size) are
// refused by variable name (ADR-0025, ADR-0013). Mutated field: each variable
// in turn; ignoring one turns its "read" or "refused" case red.
func TestLoadHardeningSettingsAreReadAndRefused(t *testing.T) {
	cfg, err := Load(settingFrom(EnvUser, "1000:1000"))
	if err != nil || cfg.User != "1000:1000" {
		t.Fatalf("user: %#v %v", cfg, err)
	}
	cfg, err = Load(settingFrom(EnvPidsLimit, "64"))
	if err != nil || cfg.PidsLimit != 64 {
		t.Fatalf("pids: %#v %v", cfg, err)
	}
	cfg, err = Load(settingFrom(EnvTmpSize, "128m"))
	if err != nil || cfg.TmpSize != "128m" {
		t.Fatalf("tmp size: %#v %v", cfg, err)
	}
	for _, c := range []struct{ key, bad string }{
		{EnvUser, "0:0"}, {EnvUser, "root"}, {EnvUser, "1000:0"},
		{EnvPidsLimit, "0"}, {EnvPidsLimit, "-1"}, {EnvPidsLimit, "many"},
		{EnvTmpSize, "0"}, {EnvTmpSize, "lots"},
	} {
		_, err := Load(settingFrom(c.key, c.bad))
		assertRefused(t, err, c.key, c.bad)
	}
}
