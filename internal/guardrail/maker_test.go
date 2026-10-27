package guardrail

import "testing"

// Property: Maker names NVIDIA only from the configuration: NVIDIA's hosted API, or
// one of the NVIDIA content-safety models Proven ships as its defaults. Any other
// model on a server gets no maker. Mutated line for the negative control: Maker's
// condition; returning "" always turns the first three cases red.
func TestMakerNamesNVIDIAOnlyFromTheConfiguration(t *testing.T) {
	for _, c := range []struct{ backend, model, want string }{
		{ModeNVIDIA, defaultNVIDIAModel, "NVIDIA"},
		{ModeNVIDIA, "nvidia/another-model", "NVIDIA"},
		{ModeServer, defaultServerModel, "NVIDIA"},
		{ModeServer, "acme-safety", ""},
		{ModeServer, "nemotron-lookalike", ""},
	} {
		if got := Maker(c.backend, c.model); got != c.want {
			t.Errorf("Maker(%q, %q) = %q, want %q", c.backend, c.model, got, c.want)
		}
	}
}
