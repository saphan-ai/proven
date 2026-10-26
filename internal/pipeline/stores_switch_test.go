package pipeline

import (
	"github.com/saphan-ai/proven/internal/candidate"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// exists reports whether path is present; any error other than absence fails the test.
func exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	t.Fatal(err)
	return false
}

// requireJudged proves the Judge still ran over every candidate of a review whose store was switched off: F1
// is dropped on its own refuting test and F2 is kept. A switch that skipped the Judge with the store would
// leave F1 undecided and fail here.
func requireJudged(t *testing.T, d candidate.Draft) {
	t.Helper()
	if _, ok := kept(d, "F1"); ok {
		t.Fatal("F1 was kept: the Judge did not decide it")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 was not kept")
	}
	if !traceSays(t, d.TracePath, "F1", "verdict", "dropped: refuted: test TestRefute passes on the change") {
		t.Fatal("F1 was not judged on its own evidence")
	}
}

// TestStoresAreOnByDefault proves ADR-0053 for H-MEM and the World Model: with the zero StoreSettings both
// stores are opened and written, and the second review of the same change is told by H-MEM that F1 was
// refuted. Mutated field: none; this is the default the switches below turn off.
func TestStoresAreOnByDefault(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	reviewWithRefutingTest(t, filepath.Join(root, "logs1"), state)
	if !exists(t, filepath.Join(state, "hmem", "state.json")) || !exists(t, filepath.Join(state, "world", "state.json")) {
		t.Fatal("a default review left a store unwritten")
	}
	second := reviewWithRefutingTest(t, filepath.Join(root, "logs2"), state)
	if !strings.HasPrefix(memoryID(t, second.TracePath, "F1"), "mem_") {
		t.Fatal("H-MEM did not suppress the finding it was told was refuted")
	}
}

// TestHMemOffNeitherOpensReadsNorWrites proves the H-MEM switch (ADR-0053) on every path of the pipeline:
// begin does not open the store (no directory), the Judge's read suppresses nothing on the second review,
// and the fold writes no refutation. The World Model stays on and folds. Mutated field: HMemOff only.
func TestHMemOffNeitherOpensReadsNorWrites(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	off := StoreSettings{HMemOff: true}
	reviewWithRefutingTestStores(t, filepath.Join(root, "logs1"), state, off)
	second := reviewWithRefutingTestStores(t, filepath.Join(root, "logs2"), state, off)
	if exists(t, filepath.Join(state, "hmem")) {
		t.Fatal("H-MEM was opened with its switch off")
	}
	if id := memoryID(t, second.TracePath, "F1"); id != "" {
		t.Fatalf("H-MEM was read with its switch off: %s", id)
	}
	requireJudged(t, second)
	if !exists(t, filepath.Join(state, "world", "state.json")) {
		t.Fatal("the World Model did not fold while only H-MEM was off")
	}
	world, err := os.ReadFile(filepath.Join(state, "world", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(world), "mem_") {
		t.Fatalf("the World Model recorded a memory id with H-MEM off: %s", world)
	}
}

// TestWorldOffNeitherOpensNorFolds proves the World Model switch (ADR-0053): begin does not open it (no
// directory, no audit log), and with no fold the trace teaches H-MEM nothing, so the second review is not
// told F1 was refuted. H-MEM stays on and is opened. Mutated field: WorldOff only.
func TestWorldOffNeitherOpensNorFolds(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	off := StoreSettings{WorldOff: true}
	reviewWithRefutingTestStores(t, filepath.Join(root, "logs1"), state, off)
	second := reviewWithRefutingTestStores(t, filepath.Join(root, "logs2"), state, off)
	if exists(t, filepath.Join(state, "world")) || exists(t, filepath.Join(root, "logs1", "world.jsonl")) {
		t.Fatal("the World Model was opened or written with its switch off")
	}
	if !exists(t, filepath.Join(state, "hmem")) {
		t.Fatal("H-MEM was not opened while only the World Model was off")
	}
	if id := memoryID(t, second.TracePath, "F1"); id != "" {
		t.Fatalf("a refutation reached H-MEM with the fold off: %s", id)
	}
	requireJudged(t, second)
}

// TestBothStoresOffLeaveNoState proves both switches together: a review writes no state at all.
func TestBothStoresOffLeaveNoState(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	reviewWithRefutingTestStores(t, filepath.Join(root, "logs1"), state, StoreSettings{HMemOff: true, WorldOff: true})
	if exists(t, filepath.Join(state, "hmem")) || exists(t, filepath.Join(state, "world")) {
		t.Fatal("a store was opened with both switches off")
	}
}

// TestLoadStores proves the switch values: empty and on leave a store on, off switches it off, and anything
// else is refused by name, never replaced by the default. Each row mutates one variable.
func TestLoadStores(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	cases := []struct {
		name string
		env  map[string]string
		want StoreSettings
		bad  string
	}{
		{"unset", nil, StoreSettings{}, ""},
		{"on", map[string]string{EnvHMem: "on", EnvWorldModel: "on"}, StoreSettings{}, ""},
		{"hmem off", map[string]string{EnvHMem: "off"}, StoreSettings{HMemOff: true}, ""},
		{"world off", map[string]string{EnvWorldModel: " off "}, StoreSettings{WorldOff: true}, ""},
		{"hmem refused", map[string]string{EnvHMem: "no"}, StoreSettings{}, EnvHMem},
		{"world refused", map[string]string{EnvWorldModel: "0"}, StoreSettings{}, EnvWorldModel},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LoadStores(env(tc.env))
			if tc.bad != "" {
				ce, ok := err.(*StoreConfigError)
				if !ok || ce.Field != tc.bad {
					t.Fatalf("error %v, want a refusal naming %s", err, tc.bad)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
	if got, err := LoadStores(nil); err != nil || got != (StoreSettings{}) {
		t.Fatalf("nil env: %+v %v", got, err)
	}
}
