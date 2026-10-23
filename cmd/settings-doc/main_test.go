package main

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// TestCheckRefusesBeforeWrite proves a cell Check rejects is not written
// into the page.
// Mutated field: the default cell. A pipe in it is the refusal.
func TestCheckRefusesBeforeWrite(t *testing.T) {
	page := "## The model generator: Meterless Nightshift\n\n## The Prover's container\n"
	bad := []envdoc.Section{{
		Heading: "H",
		Rows:    []envdoc.Setting{{Name: "PROVEN_X", Default: "a|b", Description: "d"}},
	}}
	got, err := applyCatalog(bad, page)
	if err == nil {
		t.Fatalf("a pipe in a cell was accepted:\n%s", got)
	}
	if !strings.Contains(err.Error(), "PROVEN_X") {
		t.Fatal(err)
	}
	ok := []envdoc.Section{{
		Heading: "H",
		Rows:    []envdoc.Setting{{Name: "PROVEN_X", Default: "`5s`", Description: "the bound"}},
	}}
	got, err = applyCatalog(ok, page)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "PROVEN_X") || strings.Contains(got, "a|b") {
		t.Fatalf("page:\n%s", got)
	}
}
