package change

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCensusNamesAFieldWrite: a function of this package other than NewPolicy and
// ProcessPolicy that sets a Policy's public field turns a policy no caller asked public mode for
// into a public one, as surely as a literal does, so the in-package census names it. Mutated
// field: the receiver's public field, set in WorkTree's body (`p.public = true`).
func TestCensusNamesAFieldWrite(t *testing.T) {
	for _, src := range []string{
		"package change\nfunc (p Policy) WorkTree(t string) { p.public = true; _ = p.GitCommand(\"-C\", t) }\n",
		"package change\nfunc (p *Policy) open() { p.public = true }\n",
		"package change\nfunc (p Policy) WorkTree(t string) { q := &p; q.public = !q.public; _ = q.GitCommand(\"-C\", t) }\n",
	} {
		if got := censusOf(t, src); len(got) == 0 {
			t.Errorf("the census accepted a write to Policy.public outside NewPolicy and ProcessPolicy:\n%s", src)
		}
	}
	if got := censusOf(t, "package change\nfunc (p Policy) Offline2() bool { return !p.public }\n"); len(got) != 0 {
		t.Errorf("control: a read of the field was named: %v", got)
	}
}

// TestCensusNamesAStrayPolicy pins the literal branches of both censuses, which
// no permanent test feeds. Mutated field: the CompositeLit case of censusFile, and the CompositeLit
// and ValueSpec cases of reaches (each removed, this test fails).
func TestCensusNamesAStrayPolicy(t *testing.T) {
	if got := censusOf(t, "package change\nfunc (p Policy) WorkTree(t string) { q := Policy{public: true}; _ = q.GitCommand(\"-C\", t) }\n"); len(got) == 0 {
		t.Error("the in-package census accepted a public Policy literal in a method")
	}
	for _, src := range []string{
		"package sample\nimport \"github.com/saphan-ai/proven/internal/change\"\nfunc run() { _ = change.Policy{} }\n",
		"package sample\nimport \"github.com/saphan-ai/proven/internal/change\"\nvar p change.Policy\n",
	} {
		path := filepath.Join(t.TempDir(), "source.go")
		if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := reaches(t, path, "source.go"); len(got) == 0 {
			t.Errorf("the module census accepted a Policy built outside NewPolicy:\n%s", src)
		}
	}
}
