package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/scout"
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

// TestScoutGuardKeepsSecurityWhenDocsRanksFirst checks that a class order cannot
// turn unverified source under docs into an undoubled documentation review.
func TestScoutGuardKeepsSecurityWhenDocsRanksFirst(t *testing.T) {
	_, err := parseConfig([]byte("class_order: docs, security, backend, mobile, user interface\n"))
	var ce *ConfigError
	if !asConfigError(err, &ce) || ce.Field != "class_order" || !strings.Contains(ce.Reason, "ADR-0053") {
		t.Fatalf("class order refusal = %v, want *ConfigError naming class_order and ADR-0053", err)
	}
	cfg, err := parseConfig([]byte("class_order: mobile, backend, security, docs, user interface\n"))
	if err != nil {
		t.Fatal(err)
	}
	rig := newRig()
	target := trees(t, map[string][2]string{
		"docs/source.go": {"package docs\n", "package docs\n\nvar Open = true\n"},
	})
	res := rig.review(t, cfg, target)
	if res.Class != "security" || !res.Doubled || len(rig.spec.calls) == 0 {
		t.Fatalf("class %s (%s) doubled %t, second reviewer calls %d", res.Class, res.ClassRule, res.Doubled, len(rig.spec.calls))
	}
}

// TestScoutNeverLightensVendoredCode proves the dependency arm of the guard (ADR-0053 point 3): a change of
// only vendored code is class security and doubled, and the second reviewer is called, not the lighter
// dependency plan that omits the Prover. Mutated field: the vendored-path check of notManifestOnly.
func TestScoutNeverLightensVendoredCode(t *testing.T) {
	rig := newRig()
	target := trees(t, map[string][2]string{
		"vendor/example.com/lib/Pipfile": {"[packages]\n", "[packages]\nrequests = \"*\"\n"},
	})
	res := rig.review(t, planned(), target)
	if res.Class != "security" || !res.Doubled || len(rig.spec.calls) == 0 {
		t.Fatalf("class %s (%s) doubled %t, second reviewer calls %d", res.Class, res.ClassRule, res.Doubled, len(rig.spec.calls))
	}
}

// TestScoutNeverLightensAManifestRenamedToSource proves the rename case of the dependency arm (ADR-0053
// point 3): a manifest renamed to a source file keeps the old path's dependency weight in Scout's answer,
// but the new path is code, so the review is class security and doubled. Mutated field: the destination of
// the rename, and nothing else (the source is a manifest in every case). Measured on the default registry:
// Scout answers high-band dependency for Pipfile -> app/x.py, and security for go.mod -> internal/x/x.go
// (go.mod is also a security path), so the second case holds with or without the guard.
func TestScoutNeverLightensAManifestRenamedToSource(t *testing.T) {
	for _, tc := range []struct{ from, to string }{
		{"Pipfile", "app/x.py"},
		{"go.mod", "internal/x/x.go"},
	} {
		t.Run(tc.from+" to "+tc.to, func(t *testing.T) {
			repo := t.TempDir()
			gitCmd(t, repo, "init")
			content := []byte("module example.com/app\n\ngo 1.22\n")
			if err := os.WriteFile(filepath.Join(repo, tc.from), content, 0o644); err != nil {
				t.Fatal(err)
			}
			gitCmd(t, repo, "add", tc.from)
			gitCmd(t, repo, "commit", "-m", "[fixture] note: base")
			base := strings.TrimSpace(string(gitCmd(t, repo, "rev-parse", "HEAD")))
			if err := os.MkdirAll(filepath.Dir(filepath.Join(repo, tc.to)), 0o755); err != nil {
				t.Fatal(err)
			}
			// The rename is made as a write and a remove, because git mv renames on disk and that fails
			// across some file systems; git still reports it as a rename of identical content.
			if err := os.WriteFile(filepath.Join(repo, tc.to), content, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(repo, tc.from)); err != nil {
				t.Fatal(err)
			}
			gitCmd(t, repo, "add", "-A")
			gitCmd(t, repo, "commit", "-m", "[fixture] note: rename")
			head := strings.TrimSpace(string(gitCmd(t, repo, "rev-parse", "HEAD")))
			rig := newRig()
			res := rig.reviewRequest(t, planned(), Request{Target: repo, BaselineCommit: base, HeadCommit: head})
			if res.Class != "security" || !res.Doubled || len(rig.spec.calls) == 0 {
				t.Fatalf("class %s (%s) doubled %t, second reviewer calls %d", res.Class, res.ClassRule, res.Doubled, len(rig.spec.calls))
			}
		})
	}
}

// TestNotManifestOnlyChecksEveryPath proves the closed manifest set: each old and new path must be a
// manifest or lockfile outside vendor directories. Each row changes one field of the change.
func TestNotManifestOnlyChecksEveryPath(t *testing.T) {
	cases := []struct {
		name    string
		changed []scout.ChangedFile
		refused bool
	}{
		{"manifest and lockfile", []scout.ChangedFile{{Status: "M", Path: "go.mod"}, {Status: "M", Path: "go.sum"}}, false},
		{"nested manifest", []scout.ChangedFile{{Status: "M", Path: "web/package.json"}}, false},
		{"rename between manifests", []scout.ChangedFile{{Status: "R", From: "Gemfile.lock", Path: "Gemfile"}}, false},
		{"source file", []scout.ChangedFile{{Status: "M", Path: "internal/x/x.go"}}, true},
		{"manifest renamed to source (new path)", []scout.ChangedFile{{Status: "R", From: "go.mod", Path: "internal/x/x.go"}}, true},
		{"source renamed to manifest (old path)", []scout.ChangedFile{{Status: "R", From: "internal/x/x.go", Path: "go.mod"}}, true},
		{"vendored manifest", []scout.ChangedFile{{Status: "M", Path: "vendor/modules.txt"}}, true},
		{"vendored manifest deeper", []scout.ChangedFile{{Status: "M", Path: "third_party/vendor/go.mod"}}, true},
		{"manifest-like name with a source suffix", []scout.ChangedFile{{Status: "M", Path: "go.mod.go"}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := notManifestOnly(tc.changed) != ""; got != tc.refused {
				t.Fatalf("refused %t, want %t", got, tc.refused)
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
