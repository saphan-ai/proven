package markov

import (
	"errors"
	"strings"
	"testing"
)

// TestLoadDefaultsAreOffAndBadValuesAreRefusedByName proves ADR-0013 for the four settings: unset is the engine off
// with the documented defaults, each other allowed value is taken, and a value that is not allowed is refused
// with the variable named and the allowed values listed, never replaced by the default.
func TestLoadDefaultsAreOffAndBadValuesAreRefusedByName(t *testing.T) {
	s, err := Load(nil)
	if err != nil || s != (Settings{On: false, Threshold: 8000, Unit: UnitFile, Carry: 800}) {
		t.Fatalf("%+v %v", s, err)
	}
	env := map[string]string{EnvSwitch: "on", EnvThreshold: "20000", EnvUnit: "hunk", EnvCarry: "1024"}
	s, err = Load(func(k string) string { return env[k] })
	if err != nil || s != (Settings{On: true, Threshold: 20000, Unit: UnitHunk, Carry: 1024}) {
		t.Fatalf("%+v %v", s, err)
	}
	bad := []struct{ key, value string }{
		{EnvSwitch, "yes"}, {EnvSwitch, "true"},
		{EnvThreshold, "0"}, {EnvThreshold, "-5"}, {EnvThreshold, "8k"}, {EnvThreshold, "10000001"},
		{EnvUnit, "line"}, {EnvUnit, "File"},
		{EnvCarry, "127"}, {EnvCarry, "32769"}, {EnvCarry, "lots"},
	}
	for _, b := range bad {
		_, err := Load(func(k string) string {
			if k == b.key {
				return b.value
			}
			return ""
		})
		var ce *ConfigError
		if !errors.As(err, &ce) || ce.Field != b.key || ce.Value != b.value || len(ce.Allowed) == 0 {
			t.Errorf("%s=%s: %v", b.key, b.value, err)
		}
	}
}

// TestEnvironmentDeclaresEverySettingLoadReads proves the configuration page lists exactly the four variables Load
// reads, with off as the switch's default.
func TestEnvironmentDeclaresEverySettingLoadReads(t *testing.T) {
	read := map[string]bool{}
	if _, err := Load(func(k string) string { read[k] = true; return "" }); err != nil {
		t.Fatal(err)
	}
	rows := Environment()
	if len(rows) != len(read) {
		t.Fatalf("%d rows, %d variables read", len(rows), len(read))
	}
	for _, r := range rows {
		if !read[r.Name] {
			t.Errorf("%s is declared and not read", r.Name)
		}
	}
	if rows[0].Name != EnvSwitch || rows[0].Default != "`off`" {
		t.Fatalf("%+v", rows[0])
	}
}

// TestFitBoundsTheStateAndDropsTheOldestDecisionFirst proves the carried-state bound: a block over the budget is cut
// to fit, the established decisions go before the context, and the awaiting claim, carried last, survives longest.
func TestFitBoundsTheStateAndDropsTheOldestDecisionFirst(t *testing.T) {
	r := ReviewState{
		Established: []string{strings.TrimSpace(strings.Repeat("old established fact ", 15)), "newer established fact"},
		Suspected:   []string{"a suspicion"},
		Awaiting:    []string{"a claim that needs file 20"},
	}
	full := r.Block(3, 10)
	if !strings.Contains(full.Render(), "- awaiting evidence: a claim that needs file 20") || !strings.Contains(full.Render(), "COMPLETED: 2 parts") {
		t.Fatal(full.Render())
	}
	budget := full.Tokens() - 50
	got, dropped := full.Fit(budget)
	if got.Tokens() > budget || dropped != 1 || strings.Contains(got.Render(), "old established") || !strings.Contains(got.Render(), "a claim that needs file 20") {
		t.Fatalf("dropped %d:\n%s", dropped, got.Render())
	}
	tiny, _ := full.Fit(30)
	if len(tiny.Decisions)+len(tiny.Context) != 0 || tiny.Completed == "" {
		t.Fatalf("%+v", tiny)
	}
	if same, d := full.Fit(full.Tokens()); d != 0 || same.Render() != full.Render() {
		t.Fatal("a block within its budget was changed")
	}
}

// TestFixBlockCarriesFailuresOnly proves the fix loop's state: the latest failure and its failing run are context,
// earlier failures are decisions, and the counts are the attempts'.
func TestFixBlockCarriesFailuresOnly(t *testing.T) {
	b := FixBlock([]FixFailure{
		{Attempt: 1, Reason: "the patch was refused: it edits a test"},
		{Attempt: 2, Reason: "the patch breaks another test of the package", Run: "package, command go test ./..., exit code 1: FAIL TestOther"},
	}, 3).Render()
	want := "COMPLETED: 2 attempts, none proved\nREMAINING: 1 of 3 attempts, this one included\nDECISIONS:\n- attempt 1 failed: the patch was refused: it edits a test\n" +
		"CONTEXT:\n- attempt 2 failed: the patch breaks another test of the package\n- failing run of attempt 2: package, command go test ./..., exit code 1: FAIL TestOther"
	if b != want {
		t.Fatalf("got:\n%s\nwant:\n%s", b, want)
	}
}

// TestHunksAreUnifiedDiffRegions proves the hunk unit's diff: two changes far apart are two hunks with three lines
// of context and correct line numbers, an added file is one hunk, and equal texts have none.
func TestHunksAreUnifiedDiffRegions(t *testing.T) {
	var lines []string
	for i := 1; i <= 30; i++ {
		lines = append(lines, "l"+string(rune('a'+i%26)))
	}
	base := strings.Join(lines, "\n") + "\n"
	chg := strings.Replace(strings.Replace(base, "lc\n", "lc changed\n", 1), "lz\n", "", 1)
	h := Hunks(base, chg)
	if len(h) != 2 || !strings.HasPrefix(h[0], "@@ -1,5 +1,5 @@\n") || !strings.Contains(h[0], "-lc\n+lc changed\n") || !strings.HasPrefix(h[1], "@@ -22,7 +22,6 @@\n") {
		t.Fatalf("%q", h)
	}
	if add := Hunks("", "x\ny\n"); len(add) != 1 || add[0] != "@@ -1,0 +1,2 @@\n+x\n+y\n" {
		t.Fatalf("%q", add)
	}
	if Hunks(base, base) != nil {
		t.Fatal("equal texts gave a hunk")
	}
}
