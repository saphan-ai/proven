package generator

import "testing"

// Property: the alias table holds the one entry the owner ruled on
// (2026-10-26), so an alias added without a ruling turns this red.
func TestModelAliasesHoldTheOneRuledEntry(t *testing.T) {
	want := map[string]string{"private/nvidia/nemotron-3.5-super-120b-a12b": "nvidia/nemotron-3.5-super-vl-120b-a12b"}
	if len(modelAliases) != len(want) {
		t.Fatalf("aliases %v", modelAliases)
	}
	for k, v := range want {
		if modelAliases[k] != v {
			t.Fatalf("aliases %v", modelAliases)
		}
	}
}
