package lab

// The docs site builds only while its generated pages equal what the generator writes (docs-site/scripts/guards.mjs, the
// stale-page guard, run by the site's image build). A page left stale once stopped the site from building and left the start
// page's "Open the docs" dead. This test runs the same generator and compares, so the Go suite turns red on a stale page
// without Node.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// stalePages returns the names of the pages in committed that differ from the ones in fresh, or are missing from either.
func stalePages(committed, fresh string) []string {
	var bad []string
	seen := map[string]bool{}
	for _, dir := range []string{committed, fresh} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return []string{dir + ": " + err.Error()}
		}
		for _, e := range entries {
			if seen[e.Name()] {
				continue
			}
			seen[e.Name()] = true
			a, errA := os.ReadFile(filepath.Join(committed, e.Name()))
			b, errB := os.ReadFile(filepath.Join(fresh, e.Name()))
			if errA != nil || errB != nil || !bytes.Equal(a, b) {
				bad = append(bad, e.Name())
			}
		}
	}
	return bad
}

func TestTheDocsSiteGeneratedPagesAreWhatTheGeneratorWrites(t *testing.T) {
	needTool(t, "go")
	fresh := t.TempDir()
	cmd := exec.Command("sh", filepath.Join(root, "docs-site", "scripts", "generate.sh"), fresh)
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("the generator did not run here (it builds the module): %v\n%s", err, out)
	}
	committed := filepath.Join(root, "docs-site", "generated")
	if bad := stalePages(committed, fresh); len(bad) > 0 {
		t.Errorf("stale generated pages %v: run sh docs-site/scripts/generate.sh docs-site/generated and commit", bad)
	}
	// The negative control: one page differs by a byte. Mutated field: the bytes of settings.md in a copy.
	copyDir := t.TempDir()
	entries, _ := os.ReadDir(fresh)
	for _, e := range entries {
		b, _ := os.ReadFile(filepath.Join(fresh, e.Name()))
		if e.Name() == "settings.md" {
			b = append(b, '\n', '#')
		}
		if err := os.WriteFile(filepath.Join(copyDir, e.Name()), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if bad := stalePages(committed, copyDir); len(bad) != 1 || bad[0] != "settings.md" {
		t.Errorf("the control did not turn red: a changed settings.md gave %v", bad)
	}
}
