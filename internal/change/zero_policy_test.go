package change

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestEveryUnaskedPolicyIsClosed: a Policy that no one asked for public mode for, however it was
// built (the zero value, new, NewPolicy(true)), runs git with every offline switch, so code that
// builds one without NewPolicy cannot reach a remote. Mutated field: the field the zero Policy
// leaves false (public instead of offline).
func TestEveryUnaskedPolicyIsClosed(t *testing.T) {
	var zero Policy
	for name, p := range map[string]Policy{"zero": zero, "literal": {}, "new": *new(Policy), "NewPolicy(true)": NewPolicy(true, Limits{})} {
		if !p.Offline() {
			t.Errorf("%s: the policy is public", name)
		}
		if args := p.GitCommand("status").Args; !slices.Contains(args, "protocol.allow=never") {
			t.Errorf("%s: git runs without the offline switches: %v", name, args)
		}
	}
	if NewPolicy(false, Limits{}).Offline() {
		t.Error("control: NewPolicy(false) must be public")
	}
}

// TestCensusSeesIndirectAccess: the cases the codex review of 57fccdf named. A zero Policy built
// through new is closed (above); an alias of the process-wide state is named by the in-package
// census; the module census still names every process-default function.
func TestCensusSeesIndirectAccess(t *testing.T) {
	if got := censusOf(t, "package change\nvar processState = &offline\nfunc (p Policy) WorkTree() { _ = processState.Load() }\n"); len(got) == 0 {
		t.Error("an alias of the process-wide state escaped the in-package census")
	}
	path := filepath.Join(t.TempDir(), "source.go")
	src := "package sample\nimport \"github.com/saphan-ai/proven/internal/change\"\nfunc run() { _ = change.GitCommand(\"status\") }\n"
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := reaches(t, path, "source.go"); len(got) == 0 {
		t.Error("control: the module census missed a process-default call")
	}
}
