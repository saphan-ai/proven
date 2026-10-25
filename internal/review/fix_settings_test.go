package review

import "testing"

// TestFixSettingsDefaultAndOtherValues proves ADR-0042's settings: proposed fixes are on by
// default with two attempts, `fixes: off` turns them off, and fix_attempts takes a number from 1 to 5.
func TestFixSettingsDefaultAndOtherValues(t *testing.T) {
	def := defaultSettings()
	if !def.Fixes || def.FixAttempts != 2 {
		t.Fatalf("defaults: fixes %t attempts %d", def.Fixes, def.FixAttempts)
	}
	off := mustParse(t, "fixes: off\n")
	if off.Fixes || off.FixAttempts != 2 {
		t.Fatalf("fixes off: %t %d", off.Fixes, off.FixAttempts)
	}
	more := mustParse(t, "fix_attempts: 5\n")
	if !more.Fixes || more.FixAttempts != 5 {
		t.Fatalf("fix_attempts 5: %t %d", more.Fixes, more.FixAttempts)
	}
}

// TestFixSettingsRefuseWhatTheyDoNotAllow proves each bad value is refused by its key with the
// allowed values: a count outside 1 to 5 would turn a retry into an open loop or into none at all.
func TestFixSettingsRefuseWhatTheyDoNotAllow(t *testing.T) {
	cases := []struct{ line, field, allowed string }{
		{"fix_attempts: 0\n", "fix_attempts", allowedFixAttempts},
		{"fix_attempts: 6\n", "fix_attempts", allowedFixAttempts},
		{"fix_attempts: two\n", "fix_attempts", allowedFixAttempts},
		{"fix_attempts: -1\n", "fix_attempts", allowedFixAttempts},
		{"fixes: maybe\n", "fixes", "on"},
	}
	for _, tc := range cases {
		_, err := parseConfig([]byte(tc.line))
		var ce *ConfigError
		if !asConfigError(err, &ce) || ce.Field != tc.field || len(ce.Allowed) == 0 || ce.Allowed[0] != tc.allowed {
			t.Fatalf("%q: %v", tc.line, err)
		}
	}
}

// asConfigError reports whether err is a *ConfigError and stores it in target.
func asConfigError(err error, target **ConfigError) bool {
	ce, ok := err.(*ConfigError)
	if ok {
		*target = ce
	}
	return ok
}
