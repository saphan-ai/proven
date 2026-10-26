package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestMarkovConfigRefusesWhatTheSpecDoesNotName proves the shim's partial configuration is read over the spec's
// defaults and that an unknown field or a value that is not a whole number is refused by name, never ignored.
func TestMarkovConfigRefusesWhatTheSpecDoesNotName(t *testing.T) {
	cfg, err := chunkConfig(json.RawMessage(`{"chunkSize":4000,"carryoverTokens":640}`))
	if err != nil || cfg.ChunkSize != 4000 || cfg.CarryoverTokens != 640 || cfg.MaxChunks != 24 {
		t.Fatalf("%+v %v", cfg, err)
	}
	for _, raw := range []string{`{"chunksize":4000}`, `{"chunkSize":4000.5}`, `{"chunkSize":"4000"}`, `{"chunkSize":128}`} {
		if _, err := chunkConfig(json.RawMessage(raw)); err == nil {
			t.Errorf("%s accepted", raw)
		}
	}
}

// TestMarkovRunIsCarriedAcrossRequests proves a run survives the one-process-per-request bridge: the machine an
// answer carries, handed back with the generator's text, moves the run on, and the last answer holds the record.
func TestMarkovRunIsCarriedAcrossRequests(t *testing.T) {
	no := false
	ans := markovOp(request{Op: "start", Goal: "g", Reflect: &no, At: 1})
	if ans["ok"] != true || ans["done"] != false || !strings.Contains(ans["prompt"].(string), "User Request: g") {
		t.Fatalf("%v", ans)
	}
	texts := []string{"first part\n[STATE_CHECKPOINT]\nCOMPLETED: the first part of the work", "second part [TASK_COMPLETE]"}
	for i, text := range texts {
		machine, err := json.Marshal(ans["machine"])
		if err != nil {
			t.Fatal(err)
		}
		ans = markovOp(request{Op: "feed", Machine: machine, Text: text, At: int64(2 + i)})
		if ans["ok"] != true {
			t.Fatalf("%v", ans)
		}
		if i == 0 && !strings.Contains(ans["prompt"].(string), "COMPLETED: the first part of the work") {
			t.Fatalf("step 2 prompt does not carry the state: %v", ans["prompt"])
		}
	}
	if ans["done"] != true {
		t.Fatalf("%v", ans)
	}
	b, _ := json.Marshal(ans["run"])
	if !strings.Contains(string(b), `"status":"completed"`) || !strings.Contains(string(b), `"tokenLabel":"estimated"`) {
		t.Fatal(string(b))
	}
}
