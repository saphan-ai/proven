package skeptic_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/agent/skeptic"
	"github.com/saphan-ai/proven/internal/generator"
)

func TestFamilyNamesNemotronAndTheVendor(t *testing.T) {
	if got := skeptic.Family(generator.DefaultSkepticModel); got != "nemotron" {
		t.Fatalf("nemotron %s", got)
	}
	if got := skeptic.Family("anthropic/claude-opus"); got != "anthropic" {
		t.Fatalf("anthropic %s", got)
	}
	if got := skeptic.Family("unknown"); got != "unknown" {
		t.Fatalf("unknown %s", got)
	}
	if got := skeptic.Family(""); got != "unknown" {
		t.Fatalf("empty %s", got)
	}
}

func TestSameFamilySentenceNamesTheFamily(t *testing.T) {
	model := generator.DefaultSkepticModel
	got, ok := skeptic.SameFamilySentence(model, model)
	if !ok {
		t.Fatal("same family")
	}
	want := "skeptic family nemotron is the same as the finder family nemotron (" + model + " and " + model + ")"
	if got != want {
		t.Fatalf("%s", got)
	}
	if _, ok := skeptic.SameFamilySentence(model, "anthropic/claude-opus"); ok {
		t.Fatal("different families must not produce the sentence")
	}
	if _, ok := skeptic.SameFamilySentence("unknown", "unknown"); ok {
		t.Fatal("unknown is not a family to report")
	}
}
