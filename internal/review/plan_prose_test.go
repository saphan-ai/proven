package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setMode changes the mode of a file written by trees, so a test can give a change side an executable bit.
func setMode(t *testing.T, dir, rel string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(filepath.Join(dir, "change", filepath.FromSlash(rel)), mode); err != nil {
		t.Fatal(err)
	}
}

// requireFullPlan fails unless the review ran the configured class in full (planner, prover and skeptic
// were all called) and its plan line names why the change is not documentation-only. The class stays what
// the path rules chose, which for a file under docs/ is docs: the class is not the plan, and the plan is
// what decides whether a model runs.
func requireFullPlan(t *testing.T, rig *planRig, res Result, reason string) {
	t.Helper()
	if !called(rig.def, "planner") || !called(rig.def, "prover") || !called(rig.def, "skeptic") {
		t.Fatalf("the full plan did not run (%d model calls): %+v", rig.calls(), rig.def.calls)
	}
	lines := planLines(t, res.TracePath)
	if len(lines) == 0 || !strings.Contains(lines[0], reason) || !strings.Contains(lines[0], "not documentation-only") {
		t.Fatalf("the plan line does not name %q: %q", reason, lines)
	}
}

// TestPlanSourceFileUnderDocsRunsTheFullPlan proves ADR-0049's prose guard: a sole docs/run.go matches the
// documentation trigger by its directory, but it is Go source, so the full configured plan runs with model
// calls and the trace says by name that the change is not documentation-only. It fails if the guard that
// requires every changed file to be prose is removed (the unplug control of the review of 23edde3).
func TestPlanSourceFileUnderDocsRunsTheFullPlan(t *testing.T) {
	if *scoutOff {
		t.Skip("the guard is Scout Intent's; with it off the full plan runs whatever the file is")
	}
	rig := newRig()
	res := rig.review(t, planned(), trees(t, map[string][2]string{
		"docs/run.go": {"package docs\n", "package docs\n\nvar N = 1\n"},
	}))
	requireFullPlan(t, rig, res, "docs/run.go has no prose suffix")
}

// TestPlanExecutableScriptUnderDocsRunsTheFullPlan proves the same guard for a file whose only fault is its
// suffix and mode: docs/build.sh has an executable bit and a script suffix. The next test holds the mode
// and the content guards apart from the suffix guard.
func TestPlanExecutableScriptUnderDocsRunsTheFullPlan(t *testing.T) {
	if *scoutOff {
		t.Skip("the guard is Scout Intent's; with it off the full plan runs whatever the file is")
	}
	dir := trees(t, map[string][2]string{"docs/build.sh": {"echo old\n", "echo new\n"}})
	setMode(t, dir, "docs/build.sh", 0o755)
	rig := newRig()
	res := rig.review(t, planned(), dir)
	requireFullPlan(t, rig, res, "docs/build.sh has no prose suffix")
}

// TestPlanExecutableProseSuffixRunsTheFullPlan proves the mode and the shebang guard on their own: a file
// with a suffix on the prose list is still not prose when it carries an executable bit, or starts with a
// shebang line. Mutated field, one per case: the mode bit, then the first line.
func TestPlanExecutableProseSuffixRunsTheFullPlan(t *testing.T) {
	if *scoutOff {
		t.Skip("the guard is Scout Intent's; with it off the full plan runs whatever the file is")
	}
	for name, tc := range map[string]struct {
		change string
		mode   os.FileMode
		reason string
	}{
		"executable bit": {"new\n", 0o755, "docs/tool.md is executable"},
		"shebang":        {"#!/bin/sh\necho new\n", 0o644, "docs/tool.md starts with a shebang line"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := trees(t, map[string][2]string{"docs/tool.md": {"old\n", tc.change}})
			setMode(t, dir, "docs/tool.md", tc.mode)
			rig := newRig()
			res := rig.review(t, planned(), dir)
			requireFullPlan(t, rig, res, tc.reason)
		})
	}
}

// TestPlanUnknownSuffixRunsTheFullPlan proves a suffix that is not on the prose list is not prose: a
// template under docs/ is built into something, and an unknown suffix is not known to be inert.
func TestPlanUnknownSuffixRunsTheFullPlan(t *testing.T) {
	if *scoutOff {
		t.Skip("the guard is Scout Intent's; with it off the full plan runs whatever the file is")
	}
	rig := newRig()
	res := rig.review(t, planned(), trees(t, map[string][2]string{
		"docs/page.md.tmpl": {"old\n", "new {{ .X }}\n"},
	}))
	requireFullPlan(t, rig, res, "docs/page.md.tmpl has no prose suffix")
}

// TestPlanProseOnlyStillRunsNoModel proves the guard does not cost the saving it protects: a sole
// docs/guide.md (a plain file with a prose suffix, no executable bit, no shebang) still runs no model call
// and states class docs.
func TestPlanProseOnlyStillRunsNoModel(t *testing.T) {
	rig := newRig()
	res := rig.review(t, planned(), trees(t, map[string][2]string{
		"docs/guide.md": {"old\n", "new\n"},
	}))
	if n := rig.calls(); n != 0 {
		t.Fatalf("a prose-only change made %d model calls: %+v", n, rig.def.calls)
	}
	if res.Class != "docs" {
		t.Fatalf("class %s", res.Class)
	}
}

// TestProseSuffixesAreCheckedOnLoad proves the prose list is read whole, lower-cased and refused by name
// when one item is wrong. Mutated field, one per case: a missing dot, a second dot, a repeat, an empty item.
func TestProseSuffixesAreCheckedOnLoad(t *testing.T) {
	cfg, err := parseConfig([]byte("prose_suffixes: .MD, .rst\n"))
	if err != nil || !same(cfg.ProseSuffixes, []string{".md", ".rst"}) {
		t.Fatalf("%v %v", err, cfg.ProseSuffixes)
	}
	if !same(defaultSettings().ProseSuffixes, []string{".md", ".markdown", ".rst", ".txt", ".adoc"}) {
		t.Fatalf("default list %v", defaultSettings().ProseSuffixes)
	}
	for name, val := range map[string]string{
		"missing dot": "md",
		"second dot":  ".md.tmpl",
		"repeat":      ".md, .MD",
		"empty item":  ".md,,.rst",
		"bare dot":    ".",
	} {
		_, err := parseConfig([]byte("prose_suffixes: " + val + "\n"))
		var ce *ConfigError
		if err == nil || !asConfigError(err, &ce) || ce.Field != "prose_suffixes" {
			t.Errorf("%s: %v", name, err)
		}
	}
}
