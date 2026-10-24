package tool

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/testrun"
)

// TestGovulncheckRunInTheContainerDoesNotInventReach proves a real run in the
// Prover's container never labels a Go dependency "imported but not reached"
// unless govulncheck answered and named that module (ADR-0043).
// Whether the default image contains the binary was not measured here: docker
// is not a command this leg may run. If the binary is missing, the run is
// named and the finding stays "not measured". If the binary answers and does
// not mention the module, the finding is still not measured.
// It runs only with PROVEN_TEST_DOCKER=1.
func TestGovulncheckRunInTheContainerDoesNotInventReach(t *testing.T) {
	if os.Getenv("PROVEN_TEST_DOCKER") != "1" {
		t.Skip("PROVEN_TEST_DOCKER=1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker")
	}
	base := t.TempDir()
	change := t.TempDir()
	const (
		mod = "module demo\n\ngo 1.26\n"
		src = "package demo\n\nfunc Add(a, b int) int { return a + b }\n"
	)
	for _, dir := range []string{base, change} {
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "add.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		// ADR-0025: the container runs as a user that does not own these trees.
		if err := testrun.MakeReadable(dir); err != nil {
			t.Fatal(err)
		}
	}
	box, err := testrun.NewContainer(testrun.DefaultContainerConfig(), exec.LookPath)
	if err != nil {
		t.Fatal(err)
	}
	limit, err := Load(os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	// The database path is inside the mount. The directory is absent on
	// purpose when the binary is missing: the command is still the one the
	// adapter builds, and a missing binary returns before the database is opened.
	out, err := NewRunner(Govulncheck{DB: "file:///checkout/vulndb"}, ContainerHost{Box: box, Timeout: limit}).Run(context.Background(), Request{
		Paths: []string{"add.go"}, BaseDir: base, ChangeDir: change,
	})
	if err != nil {
		t.Fatal(err)
	}
	finding := LabelFindings([]candidate.Candidate{
		{Risk: candidate.Risk{ID: "osv:example.com/quiet", Message: "in range"}, Ecosystem: "Go"},
	}, out.Answered, Symbols(out.ChangeOutput))
	if finding[0].ReachLabel != candidate.ReachNotMeasured {
		t.Fatalf("label %q reason %q\n%s", finding[0].ReachLabel, out.Reason, out.ChangeOutput)
	}
	if strings.Contains(finding[0].Message, "imported but not reached") || finding[0].ReachLabel == candidate.ReachReached {
		t.Fatalf("%#v", finding[0])
	}
	if !out.Answered && !strings.Contains(out.Reason, "govulncheck") {
		t.Fatalf("a miss was not named: %q", out.Reason)
	}
	if out.Answered && !strings.HasPrefix(out.Command, "govulncheck -json -db ") {
		t.Fatalf("command %q", out.Command)
	}
}
