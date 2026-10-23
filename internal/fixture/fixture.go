// Package fixture loads a sample change: trees, recorded answers, recorded tests, sample tables.
//
// Sample mode (--sample) runs the whole review with no network. A fixture
// stands in for every outside service: the recorded answers stand in for the
// model calls of the Planner, Investigator, Prover and Skeptic, the recorded
// tests stand in for the Prover's container runs, and tables.sql stands in for
// the Zetaris tables the Investigator queries. The Judge and the Writer run as
// they do live. The pipeline loads a fixture when the request asks for a sample.
package fixture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Fixture is one sample change. Root holds the base/ and change/ trees,
// Snapshot names the sample data the evidence cites, Answers are the recorded
// model answers, Tests the recorded test outcomes, and Tables the path of the
// SQL file that builds the sample tables.
type Fixture struct {
	Root     string
	Snapshot string
	Answers  []generator.Response
	Tests    []testrun.Recorded
	Tables   string
}

// file is the shape of fixture.json on disk.
type file struct {
	Snapshot            string          `json:"snapshot"`
	PlannerPromptSHA256 string          `json:"planner_prompt_sha256"`
	PlannerResponse     json.RawMessage `json:"planner_response"`
	Responses           []struct {
		Agent    string `json:"agent"`
		Prompt   string `json:"prompt"`
		Response string `json:"response"`
	} `json:"responses"`
	RecordedTests []struct {
		FindingID     string `json:"finding_id"`
		Path          string `json:"path"`
		Name          string `json:"name"`
		Command       string `json:"command"`
		Output        string `json:"output"`
		FailsOnChange bool   `json:"fails_on_change"`
		PassesOnBase  bool   `json:"passes_on_base"`
	} `json:"recorded_tests"`
}

// Load reads fixture.json and checks the planner prompt still matches the trees.
// A fixture whose trees changed after its answers were recorded is refused, so a
// sample review never replays answers to a prompt it did not ask.
func Load(root string) (*Fixture, error) {
	b, err := os.ReadFile(filepath.Join(root, "fixture.json"))
	if err != nil {
		return nil, err
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	// Every piece of evidence names its snapshot, so a fixture without one is
	// refused.
	if f.Snapshot == "" {
		return nil, fmt.Errorf("fixture: snapshot is empty")
	}
	files, err := change.Read(root)
	if err != nil {
		return nil, err
	}
	// The planner answer is keyed by a hash of its prompt, not by the prompt
	// text, so an edit to the trees or to the planner prompt shows up here.
	prompt := planner.Prompt(files)
	sum := sha256.Sum256([]byte(prompt))
	got := hex.EncodeToString(sum[:])
	if got != f.PlannerPromptSHA256 {
		return nil, fmt.Errorf("fixture: planner prompt sha256 is %s", got)
	}
	if len(f.PlannerResponse) == 0 {
		return nil, fmt.Errorf("fixture: planner response is empty")
	}
	// The planner answer goes first and carries the prompt rebuilt here, so a
	// recorded generator finds it by the same prompt the Planner will send.
	answers := []generator.Response{{
		Agent: "planner", Prompt: prompt, Response: string(f.PlannerResponse),
	}}
	for _, r := range f.Responses {
		answers = append(answers, generator.Response{Agent: r.Agent, Prompt: r.Prompt, Response: r.Response})
	}
	tests := make([]testrun.Recorded, 0, len(f.RecordedTests))
	for _, r := range f.RecordedTests {
		tests = append(tests, testrun.Recorded{
			FindingID: r.FindingID, Path: r.Path, Name: r.Name, Command: r.Command,
			Output: r.Output, FailsOnChange: r.FailsOnChange, PassesOnBase: r.PassesOnBase,
		})
	}
	// The sample tables must exist now, not fail later in the Investigator.
	tables := filepath.Join(root, "tables.sql")
	if _, err := os.Stat(tables); err != nil {
		return nil, err
	}
	return &Fixture{Root: root, Snapshot: f.Snapshot, Answers: answers, Tests: tests, Tables: tables}, nil
}
