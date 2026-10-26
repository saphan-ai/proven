package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestScoutDecideAnswersInTheHarnessShape proves a decide call over the shipped change registry answers
// with the fields the eval runner reads: the contract's primary intent, the candidates, the risk level
// and the tool plan's capabilities. Each field is read back through JSON, as the runner reads it.
func TestScoutDecideAnswersInTheHarnessShape(t *testing.T) {
	reg := filepath.Join("..", "..", "internal", "scout", "changes.json")
	line, err := json.Marshal(map[string]any{
		"id": 7, "engine": "scout", "op": "decide", "registry": reg, "policyPack": "default",
		"prompt": "M go.mod\nM go.sum", "surface": "change",
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(handle(line))
	if err != nil {
		t.Fatal(err)
	}
	var ans struct {
		ID       int  `json:"id"`
		OK       bool `json:"ok"`
		Decision struct {
			Candidates []struct {
				Intent string `json:"intent"`
			} `json:"candidates"`
			Risk struct {
				Level string `json:"level"`
			} `json:"risk"`
			ToolPlan []struct {
				Capability string `json:"capability"`
			} `json:"toolPlan"`
			Clarification     *json.RawMessage `json:"clarification"`
			ExecutionContract struct {
				Intent struct {
					Primary struct {
						ID string `json:"id"`
					} `json:"primary"`
				} `json:"intent"`
			} `json:"executionContract"`
		} `json:"decision"`
	}
	if err := json.Unmarshal(raw, &ans); err != nil {
		t.Fatal(err)
	}
	d := ans.Decision
	if !ans.OK || ans.ID != 7 || d.ExecutionContract.Intent.Primary.ID != "change.dependency" || d.Clarification != nil {
		t.Fatalf("%s", raw)
	}
	if len(d.Candidates) == 0 || d.Candidates[0].Intent != "change.dependency" || d.Risk.Level != "medium" {
		t.Fatalf("%s", raw)
	}
	if len(d.ToolPlan) != 1 || d.ToolPlan[0].Capability != "plan.dependency" {
		t.Fatalf("%s", raw)
	}
}

// TestScoutRefusesWhatItDoesNotKeep proves replan is refused by name, and that a registry that does not
// load is an answer with ok false rather than a crash.
func TestScoutRefusesWhatItDoesNotKeep(t *testing.T) {
	ans := handle([]byte(`{"id":1,"engine":"scout","op":"replan","registry":"x"}`))
	if ans["ok"] != false || ans["error"] == "" {
		t.Fatalf("%v", ans)
	}
	bad := filepath.Join(t.TempDir(), "r.json")
	if err := os.WriteFile(bad, []byte(`{"intents":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	ans = handle([]byte(`{"id":2,"engine":"scout","op":"decide","registry":"` + bad + `","prompt":"x"}`))
	if ans["ok"] != false {
		t.Fatalf("%v", ans)
	}
}
