// Package fixture loads a sample change: trees, recorded answers, recorded tests, sample tables.
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
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Fixture is one sample change.
type Fixture struct {
	Root      string
	Snapshot  string
	Answers   []generator.Response
	Tests     []testrun.Recorded
	Tables    string
	Guardrail []guardrail.Record // empty means the sample review answers safe
}

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
	// Guardrail is optional. agent "planner" with an empty prompt binds the
	// planner prompt of these trees. A fixture with no records answers safe.
	Guardrail []struct {
		Agent      string   `json:"agent"`
		Prompt     string   `json:"prompt"`
		Answer     string   `json:"answer"`
		User       string   `json:"user"`
		Response   string   `json:"response"`
		Categories []string `json:"categories"`
	} `json:"guardrail"`
}

// Load reads fixture.json and checks the planner prompt still matches the trees.
func Load(root string) (*Fixture, error) {
	b, err := os.ReadFile(filepath.Join(root, "fixture.json"))
	if err != nil {
		return nil, err
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if f.Snapshot == "" {
		return nil, fmt.Errorf("fixture: snapshot is empty")
	}
	files, err := change.Read(root)
	if err != nil {
		return nil, err
	}
	plannerPrompt := planner.Prompt(files)
	sum := sha256.Sum256([]byte(plannerPrompt))
	got := hex.EncodeToString(sum[:])
	if got != f.PlannerPromptSHA256 {
		return nil, fmt.Errorf("fixture: planner prompt sha256 is %s", got)
	}
	if len(f.PlannerResponse) == 0 {
		return nil, fmt.Errorf("fixture: planner response is empty")
	}
	answers := []generator.Response{{
		Agent: "planner", Prompt: plannerPrompt, Response: string(f.PlannerResponse),
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
	tables := filepath.Join(root, "tables.sql")
	if _, err := os.Stat(tables); err != nil {
		return nil, err
	}
	rails := make([]guardrail.Record, 0, len(f.Guardrail))
	for _, row := range f.Guardrail {
		text := row.Prompt
		if text == "" && row.Agent == "planner" {
			text = plannerPrompt
		}
		if text == "" {
			return nil, fmt.Errorf("fixture: a guardrail record has no prompt")
		}
		rails = append(rails, guardrail.Record{
			Prompt: text, Answer: row.Answer, User: row.User, Response: row.Response,
			Categories: append([]string(nil), row.Categories...),
		})
	}
	return &Fixture{
		Root: root, Snapshot: f.Snapshot, Answers: answers, Tests: tests, Tables: tables, Guardrail: rails,
	}, nil
}
