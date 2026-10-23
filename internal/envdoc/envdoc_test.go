package envdoc

import (
	"strings"
	"testing"
	"time"
)

// TestSecondSpecKeepsWholeSeconds proves whole seconds are written as 60s,
// not 1m0s, and a fraction keeps Go's own form.
func TestSecondSpecKeepsWholeSeconds(t *testing.T) {
	if got := SecondSpec(60 * time.Second); got != "60s" {
		t.Fatalf("60s rendered as %s", got)
	}
	if got := SecondSpec(5 * time.Second); got != "5s" {
		t.Fatalf("5s rendered as %s", got)
	}
	if got := SecondSpec(1500 * time.Millisecond); got != "1.5s" {
		t.Fatalf("1.5s rendered as %s", got)
	}
}

// TestDriftNamesTheExtraSetting proves a row present only in the render is
// named as the drift.
func TestDriftNamesTheExtraSetting(t *testing.T) {
	base := []Section{{
		Heading: "One",
		Rows:    []Setting{{Name: "PROVEN_A", Default: Tick("1"), Description: "a"}},
	}}
	extra := []Section{{
		Heading: "One",
		Rows: []Setting{
			{Name: "PROVEN_A", Default: Tick("1"), Description: "a"},
			{Name: "PROVEN_FAKE_SETTING", Default: Tick("nope"), Description: "not a setting"},
		},
	}}
	if name := DriftSetting(Render(base), Render(extra)); name != "PROVEN_FAKE_SETTING" {
		t.Fatalf("drift named %q", name)
	}
}

// TestApplyIsIdempotent proves the first Apply on a page without markers
// replaces the hand-written sections, a second Apply changes nothing, and the
// prose outside the markers stays.
func TestApplyIsIdempotent(t *testing.T) {
	page := "intro\n\n" + startHeading + "\n\nold\n\n" + endHeading + "\nrest\n"
	rendered := Render([]Section{{
		Heading: strings.TrimPrefix(startHeading, "## "),
		Rows:    []Setting{{Name: "PROVEN_A", Default: Tick("1"), Description: "a"}},
	}})
	once, err := Apply(page, rendered)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := Apply(once, rendered)
	if err != nil {
		t.Fatal(err)
	}
	if once != twice {
		t.Fatalf("second apply changed the page\n%s", twice)
	}
	got, err := Extract(twice)
	if err != nil {
		t.Fatal(err)
	}
	if got != rendered {
		t.Fatalf("extracted\n%s\nwant\n%s", got, rendered)
	}
	if !strings.Contains(twice, endHeading) || !strings.Contains(twice, "intro\n") {
		t.Fatalf("prose outside the markers moved\n%s", twice)
	}
}

// TestCheckRejectsAPipe proves a pipe in a cell, which would break the
// table, is refused with the setting's name.
func TestCheckRejectsAPipe(t *testing.T) {
	err := Check([]Section{{
		Heading: "One",
		Rows:    []Setting{{Name: "PROVEN_A", Default: "a|b", Description: "a"}},
	}})
	if err == nil || !strings.Contains(err.Error(), "PROVEN_A") {
		t.Fatalf("error %v", err)
	}
}
