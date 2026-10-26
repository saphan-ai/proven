package review

import (
	"testing"

	"github.com/saphan-ai/proven/internal/markov"
)

// TestScoutNeverLightensANonProseChange proves ADR-0053 point 3: when Scout answers docs for a change whose
// files are not verified prose, the review is class security and doubled, so the second reviewer runs. The
// names are ones Scout's registry cannot read as security: credentials-handling code, a Polish name and a Thai
// name, all under docs/, which is why Scout calls them documentation. Remove the guard and each case comes back
// as class docs, undoubled (the blind set showed this with three real changes).
func TestScoutNeverLightensANonProseChange(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"credentials code under docs", "docs/credentials.go"},
		{"Polish name", "docs/uwierzytelnianie_uzytkownika.go"},
		{"Thai name", "docs/รหัสผ่านผู้ใช้.go"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rig := newRig()
			target := trees(t, map[string][2]string{tc.path: {"package docs\n", "package docs\n\nvar Open = true\n"}})
			res := rig.review(t, planned(), target)
			if res.Class != "security" || !res.Doubled || len(rig.spec.calls) == 0 {
				t.Fatalf("class %s (%s) doubled %t, second reviewer calls %d", res.Class, res.ClassRule, res.Doubled, len(rig.spec.calls))
			}
		})
	}
}

// TestScoutStillLightensVerifiedProse proves the guard's other side (ADR-0053 point 3): a change whose every
// file is verified prose keeps the documentation plan, with no model call. The guard must not turn every docs
// change into a security review.
func TestScoutStillLightensVerifiedProse(t *testing.T) {
	rig := newRig()
	res := rig.review(t, planned(), docsOnly(t))
	if res.Class != "docs" || res.Doubled || rig.calls() != 0 {
		t.Fatalf("class %s doubled %t, model calls %d", res.Class, res.Doubled, rig.calls())
	}
}

// TestMeterlessEnginesAreOnByDefault proves ADR-0053 point 2: Markovian and Scout Intent are on in the
// defaults and in an empty proven.yaml, and each is switched off by its own setting. Mutated field: the
// switch of the one engine under test.
func TestMeterlessEnginesAreOnByDefault(t *testing.T) {
	if !defaultSettings().ScoutIntent {
		t.Fatal("scout_intent is off by default")
	}
	cfg, err := parseConfig([]byte("timeout: 0s\n"))
	if err != nil || !cfg.ScoutIntent {
		t.Fatalf("empty file: %v %t", err, cfg.ScoutIntent)
	}
	cfg, err = parseConfig([]byte("scout_intent: off\n"))
	if err != nil || cfg.ScoutIntent {
		t.Fatalf("scout_intent: off: %v %t", err, cfg.ScoutIntent)
	}
	none := func(string) string { return "" }
	s, err := markov.Load(none)
	if err != nil || !s.On {
		t.Fatalf("unset PROVEN_MARKOVIAN: %v %t", err, s.On)
	}
	s, err = markov.Load(func(k string) string {
		if k == markov.EnvSwitch {
			return "off"
		}
		return ""
	})
	if err != nil || s.On {
		t.Fatalf("PROVEN_MARKOVIAN=off: %v %t", err, s.On)
	}
}

// TestScoutSwitchedOffRunsTheFullPlanForANonProseChange proves that switching Scout Intent off restores the
// review as it was: the configured class and the full plan, as the rules give them.
func TestScoutSwitchedOffRunsTheFullPlanForANonProseChange(t *testing.T) {
	rig := newRig()
	cfg := defaultSettings()
	cfg.ScoutIntent = false
	target := trees(t, map[string][2]string{"docs/credentials.go": {"package docs\n", "package docs\n\nvar Open = true\n"}})
	res := rig.review(t, cfg, target)
	if res.Class != "docs" || res.Doubled {
		t.Fatalf("class %s doubled %t", res.Class, res.Doubled)
	}
}
