package lab

// The Nightshift receipts importer (lab/seed/nightshift-import.sh): only Proven's own calls, prompt_preview dropped, the
// receipts replayed to end at the moment of the import, one push to the lab's own Loki. A stand-in curl records what would
// have been sent.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const importSecret = "PROMPT-TEXT-THAT-MUST-NOT-LEAVE-9c1e"

// stubCurl answers the importer's two requests: the label query says what the test chose, the push is written to a file
// and answered 204. The importer reads no other answer.
const stubCurl = `#!/bin/sh
case " $* " in
  *" -X POST "*)
    for a in "$@"; do
      case $a in @*) cp "${a#@}" "$FAKE_DIR/loki.post" ;; esac
    done
    printf '204'
    ;;
  *)
    printf '%s' "$LABELS_ANSWER"
    ;;
esac
`

const receipts = `{"ts_ms": 1792921200000, "app": "proven", "status": "ok", "model": "nvidia/nemotron-3.5-super-vl-120b-a12b", "provider": "nvidia", "verified": true, "latency_ms": 2100, "input_tokens": 21530, "output_tokens": 105, "request_id": "chatcmpl-1", "prompt_preview": "` + importSecret + ` and a \"quoted\" word, with a comma"}
{"ts_ms": 1792921260000, "app": "proven", "status": "failed", "model": "nvidia/nemotron-3.5-super-vl-120b-a12b", "provider": "nvidia", "reason": "provider_deadline", "verified": true, "request_id": "chatcmpl-2"}
{"ts_ms": 1792921320000, "prompt_preview": "` + importSecret + `", "app": "proven", "status": "failed", "model": "none", "provider": "nvidia", "reason": "empty_response", "verified": false, "request_id": "chatcmpl-3"}
{"ts_ms": 1792921380000, "app": "proven-review", "status": "ok", "model": "nvidia/other", "provider": "nvidia", "verified": true, "request_id": "chatcmpl-fleet", "prompt_preview": "` + importSecret + `"}
not json at all
{"ts_ms": "soon", "app": "proven", "status": "ok"}
`

type pushed struct {
	Streams []struct {
		Stream map[string]string `json:"stream"`
		Values [][]string        `json:"values"`
	} `json:"streams"`
}

func runImport(t *testing.T, tree, input string, labels string, extra ...string) (out string, post []byte, err error) {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	writeExec(t, bin, "curl", stubCurl)
	cmd := exec.Command("sh", filepath.Join(tree, "lab", "seed", "nightshift-import.sh"))
	cmd.Env = append([]string{"PATH=" + bin + ":" + os.Getenv("PATH"), "FAKE_DIR=" + dir, "TMPDIR=" + dir, "LAB_LOKI_URL=http://lgtm:3100", "LABELS_ANSWER=" + labels}, extra...)
	cmd.Stdin = strings.NewReader(input)
	b, err := cmd.CombinedOutput()
	post, _ = os.ReadFile(filepath.Join(dir, "loki.post"))
	return string(b), post, err
}

func TestNightshiftImportLoadsOnlyProvenOwnCallsAndDropsThePrompt(t *testing.T) {
	tree := mustAbs(t, root)
	before := time.Now()
	out, post, err := runImport(t, tree, receipts, `{"status":"success","data":[]}`)
	if err != nil {
		t.Fatalf("the import failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "3 loaded, 1 of other apps left out, 2 refused") {
		t.Errorf("the counts are wrong:\n%s", out)
	}
	if strings.Contains(string(post), "prompt_preview") || strings.Contains(string(post), importSecret) || strings.Contains(out, importSecret) {
		t.Errorf("a prompt reached the push or the output:\n%s", post)
	}
	var p pushed
	if err := json.Unmarshal(post, &p); err != nil {
		t.Fatalf("the push is not JSON: %v\n%s", err, post)
	}
	if len(p.Streams) != 3 {
		t.Fatalf("%d streams pushed, want the 3 receipts of app proven", len(p.Streams))
	}
	wantLabels := []map[string]string{
		{"job": "nightshift", "app": "proven", "status": "ok", "reason": "none", "verified": "true", "provider": "nvidia", "model": "nvidia/nemotron-3.5-super-vl-120b-a12b"},
		{"job": "nightshift", "app": "proven", "status": "failed", "reason": "provider_deadline", "verified": "true", "provider": "nvidia", "model": "nvidia/nemotron-3.5-super-vl-120b-a12b"},
		{"job": "nightshift", "app": "proven", "status": "failed", "reason": "empty_response", "verified": "false", "provider": "nvidia", "model": "none"},
	}
	var stamps []int64
	for i, s := range p.Streams {
		for k, v := range wantLabels[i] {
			if s.Stream[k] != v {
				t.Errorf("stream %d label %s = %q, want %q", i, k, s.Stream[k], v)
			}
		}
		ns, err := strconv.ParseInt(s.Values[0][0], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		stamps = append(stamps, ns)
		// The line is the receipt without its prompt, and still a JSON object the dashboard can read.
		var line map[string]any
		if err := json.Unmarshal([]byte(s.Values[0][1]), &line); err != nil {
			t.Errorf("stream %d's line is not JSON: %v\n%s", i, err, s.Values[0][1])
		}
		if _, ok := line["ts_ms"]; !ok {
			t.Errorf("stream %d's line lost its own ts_ms", i)
		}
	}
	// Replayed: the newest ends at the moment of the import, and the spacing is the receipts' own (60 s).
	if newest := time.Unix(0, stamps[2]); newest.Before(before.Add(-2*time.Second)) || newest.After(time.Now().Add(2*time.Second)) {
		t.Errorf("the newest receipt is stamped %v, not the moment of the import", newest)
	}
	if stamps[2]-stamps[1] != 60e9 || stamps[1]-stamps[0] != 60e9 {
		t.Errorf("the spacing of the receipts was not kept: %v", stamps)
	}
	// The negative control: the prompt is not dropped, and the check that refuses it is gone. Mutated field: both.
	mut := copyLab(t)
	copyInto(t, mut, "lab/seed")
	mutate(t, mut, "lab/seed/nightshift-import.sh", `gsub(/,[ \t]*"prompt_preview"`, `gsub(/,[ \t]*"prompt_preview_never"`)
	mutate(t, mut, "lab/seed/nightshift-import.sh", `gsub(/"prompt_preview"[ \t]*:`, `gsub(/"prompt_preview_never"[ \t]*:`)
	mutate(t, mut, "lab/seed/nightshift-import.sh", `if (index(line, "prompt_preview") > 0) { refused++; next }`, ``)
	_, mpost, _ := runImport(t, mut, receipts, `{"status":"success","data":[]}`)
	if !strings.Contains(string(mpost), importSecret) {
		t.Error("the control did not turn red: with the drop and the refusal gone the prompt still did not reach the push")
	}
}

func TestNightshiftImportRefusesWhatItCannotTrust(t *testing.T) {
	tree := mustAbs(t, root)
	// Nothing of Proven's own: nothing is pushed, and it says so.
	out, post, err := runImport(t, tree, `{"ts_ms": 1, "app": "proven-review", "status": "ok"}`+"\n", `{"data":[]}`)
	if err == nil || len(post) != 0 || !strings.Contains(out, "no receipt of Proven's own") {
		t.Errorf("an input with no receipt of Proven's own was not refused:\n%v\n%s", err, out)
	}
	// Loki somewhere else: refused before any request.
	out, post, err = runImport(t, tree, receipts, `{"data":[]}`, "LAB_LOKI_URL=http://loki.example.test:3100")
	if err == nil || len(post) != 0 || !strings.Contains(out, "not the lab's own Loki") {
		t.Errorf("a Loki outside the lab was not refused:\n%v\n%s", err, out)
	}
	// Already loaded: not loaded twice, unless forced.
	out, post, err = runImport(t, tree, receipts, `{"status":"success","data":["nightshift"]}`)
	if err != nil || len(post) != 0 || !strings.Contains(out, "already loaded") {
		t.Errorf("a set already in Loki was loaded again:\n%v\n%s", err, out)
	}
	if _, post, err = runImport(t, tree, receipts, `{"status":"success","data":["nightshift"]}`, "FORCE=1"); err != nil || len(post) == 0 {
		t.Errorf("FORCE=1 did not load the receipts again: %v", err)
	}
}
