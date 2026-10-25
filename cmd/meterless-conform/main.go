// Command meterless-conform exposes the H-MEM and world stores, and Scout Intent, on stdin.
// One JSON request per line, one JSON answer per line. No network.
//
// It lets Meterless's own conformance suites test the Go engines through the
// TypeScript bridge in test/conformance/meterless (ADR-0003, point 4; ADR-0018,
// point 7). The bridge starts one process per request, so a session is a
// directory under the temp dir, and the next process opens the same store.
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/scout"
	"github.com/saphan-ai/proven/internal/world"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

// run answers each line in turn. A bad request gets an answer with ok false;
// only a failure to read or to write stops the process.
func run(in *os.File, out *os.File, errOut *os.File) int {
	sc := bufio.NewScanner(in)
	// A line may be up to 8 MiB: an ingest request carries all its observations
	// on one line.
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	for sc.Scan() {
		if err := enc.Encode(handle(sc.Bytes())); err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

// request holds the fields of every operation; each one reads those it
// needs. At is the time, in Unix milliseconds, that an H-MEM call runs at.
type request struct {
	ID           int               `json:"id"`
	Engine       string            `json:"engine"`
	Op           string            `json:"op"`
	Session      string            `json:"session"`
	At           int64             `json:"at"`
	Input        json.RawMessage   `json:"input"`
	Query        string            `json:"query"`
	TopN         int               `json:"topN"`
	Threshold    float64           `json:"threshold"`
	MemoryID     string            `json:"memoryId"`
	By           string            `json:"by"`
	Kind         string            `json:"kind"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Similarity   float64           `json:"similarity"`
	Observations []json.RawMessage `json:"observations"`
	// Registry, PolicyPack, Prompt, User and Surface are Scout's (ADR-0049 point 4).
	Registry   string          `json:"registry"`
	PolicyPack json.RawMessage `json:"policyPack"`
	Prompt     string          `json:"prompt"`
	User       scout.User      `json:"user"`
	Surface    string          `json:"surface"`
}

// handle decodes one line and sends it to its engine. A line that does not
// decode is answered with id 0, because its id is unknown.
func handle(line []byte) map[string]any {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		return fail(0, err.Error())
	}
	switch req.Engine {
	case "hmem":
		return hmemOp(req)
	case "world":
		return worldOp(req)
	case "scout":
		return scoutOp(req)
	default:
		return fail(req.ID, "unknown engine "+req.Engine)
	}
}

// hmemOp runs one H-MEM operation.
func hmemOp(req request) map[string]any {
	switch req.Op {
	case "open":
		id, err := createSession("h")
		if err != nil {
			return fail(req.ID, err.Error())
		}
		if _, err := openH(id, req.At); err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"session": id})
	// ADR-0018, point 1: these belong to the parts built after the minimum, so
	// they are refused by name, not faked.
	case "sleep", "dream", "conflicts", "ledger":
		return fail(req.ID, "hmem: "+req.Op+" is deferred (ADR-0018)")
	}
	st, err := openH(req.Session, req.At)
	if err != nil {
		return fail(req.ID, err.Error())
	}
	switch req.Op {
	case "add":
		rec, err := addH(st, req.At, req.Input)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"record": rec})
	case "supersede":
		if err := st.MarkSuperseded(req.MemoryID, req.By); err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{})
	case "feedback":
		if err := st.Feedback(req.MemoryID, req.Kind); err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{})
	case "get":
		rec, found := st.Get(req.MemoryID)
		if !found {
			return ok(req.ID, map[string]any{"record": nil})
		}
		return ok(req.ID, map[string]any{"record": rec})
	case "all":
		recs := st.All()
		if recs == nil {
			recs = []hmem.Record{}
		}
		return ok(req.ID, map[string]any{"records": recs})
	case "query":
		q, err := st.Query(req.Query, req.TopN, req.Threshold)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		hits := make([]map[string]any, 0, len(q.Hits))
		for _, h := range q.Hits {
			hits = append(hits, map[string]any{"memory": h.Record, "relevance": h.Relevance})
		}
		return ok(req.ID, map[string]any{"result": map[string]any{"memories": hits}})
	default:
		return fail(req.ID, "hmem: unknown op "+req.Op)
	}
}

// worldOp runs one World Model operation. normalize, entityId and decide are
// pure functions and need no session.
func worldOp(req request) map[string]any {
	switch req.Op {
	case "normalize":
		return ok(req.ID, map[string]any{"name": world.NormalizeName(req.Name)})
	case "entityId":
		return ok(req.ID, map[string]any{"entityId": world.EntityIDFromName(req.Type, req.Name)})
	case "decide":
		return ok(req.ID, map[string]any{"decision": world.Decide(req.Similarity)})
	case "open":
		id, err := createSession("w")
		if err != nil {
			return fail(req.ID, err.Error())
		}
		if _, err := openW(id); err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"session": id})
	// ADR-0018, point 1: the graph is outside the minimum, so it is refused by
	// name.
	case "relate", "neighbors":
		return fail(req.ID, "world: bitemporal graph is outside the minimum (ADR-0018)")
	}
	st, err := openW(req.Session)
	if err != nil {
		return fail(req.ID, err.Error())
	}
	switch req.Op {
	case "upsert":
		ent, err := upsertW(st, req.Input)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"entity": ent})
	case "ingest":
		rep, err := ingestW(st, req.Observations)
		if err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"changed": rep.Changed})
	case "canonical":
		return ok(req.ID, map[string]any{"canonical": st.Canonical()})
	case "audit":
		// The audit is the World Model's own JSONL log (ADR-0018, point 3).
		rows, err := readAudit(filepath.Join(sessionBase(req.Session), "logs", "world.jsonl"))
		if err != nil {
			return fail(req.ID, err.Error())
		}
		return ok(req.ID, map[string]any{"audit": rows})
	default:
		return fail(req.ID, "world: unknown op "+req.Op)
	}
}

// createSession makes a new random session directory with the folders both
// stores write to. The prefix h or w says which engine it is for.
func createSession(prefix string) (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	id := prefix + hex.EncodeToString(buf[:])
	base := sessionBase(id)
	if err := os.MkdirAll(filepath.Join(base, "state"), 0o700); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(base, "logs"), 0o700); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(base, "hmem"), 0o700); err != nil {
		return "", err
	}
	return id, nil
}

// sessionBase is a session's directory under the temp dir.
func sessionBase(id string) string {
	return filepath.Join(os.TempDir(), "proven-meterless-conform", id)
}

// knownSession accepts only an id this command could have made, a prefix and
// hex digits, for a directory that exists. The id is joined into a path, so
// this also keeps a request from naming a path outside the sessions.
func knownSession(id string) error {
	if len(id) < 2 || id[0] != 'h' && id[0] != 'w' {
		return fmt.Errorf("unknown session")
	}
	for _, r := range id[1:] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return fmt.Errorf("unknown session")
		}
	}
	st, err := os.Stat(sessionBase(id))
	if err != nil || !st.IsDir() {
		return fmt.Errorf("unknown session")
	}
	return nil
}

// openH opens a session's H-MEM store with the request's time as its clock,
// so the suite controls time.
func openH(id string, at int64) (*hmem.Store, error) {
	if err := knownSession(id); err != nil {
		return nil, fmt.Errorf("hmem: %w", err)
	}
	base := sessionBase(id)
	return hmem.Open(hmem.Options{
		Dir:    filepath.Join(base, "state"),
		LogDir: filepath.Join(base, "logs"),
		Clock:  func() time.Time { return time.UnixMilli(at) },
	})
}

// addH decodes an add request's input and adds the memory at the request's
// time. An empty or null input is passed on as empty fields, and the store
// decides.
func addH(st *hmem.Store, at int64, raw json.RawMessage) (hmem.Record, error) {
	var body struct {
		Content    string   `json:"content"`
		Type       string   `json:"type"`
		Layer      string   `json:"layer"`
		Tags       []string `json:"tags"`
		Confidence *float64 `json:"confidence"`
		Source     string   `json:"source"`
		Origin     string   `json:"origin"`
		Label      string   `json:"label"`
		ChatID     string   `json:"chatId"`
	}
	if len(bytes.TrimSpace(raw)) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &body); err != nil {
			return hmem.Record{}, err
		}
	}
	return st.Add(hmem.AddInput{
		Content:    body.Content,
		Type:       body.Type,
		Layer:      body.Layer,
		Tags:       body.Tags,
		Confidence: body.Confidence,
		Source:     body.Source,
		Origin:     body.Origin,
		Label:      body.Label,
		ChatID:     body.ChatID,
		At:         at,
	})
}

// openW opens a session's World Model, backed by the session's own H-MEM
// store (ADR-0018, point 4).
func openW(id string) (*world.Store, error) {
	if err := knownSession(id); err != nil {
		return nil, fmt.Errorf("world: %w", err)
	}
	base := sessionBase(id)
	mem, err := hmem.Open(hmem.Options{
		Dir:    filepath.Join(base, "hmem"),
		LogDir: filepath.Join(base, "logs"),
	})
	if err != nil {
		return nil, err
	}
	return world.Open(world.Options{
		Dir:    filepath.Join(base, "state"),
		LogDir: filepath.Join(base, "logs"),
		Memory: mem,
	})
}

// upsertW decodes an entity and upserts it. An external key gives its id;
// otherwise the id comes from the type and the name.
func upsertW(st *world.Store, raw json.RawMessage) (world.Entity, error) {
	var in struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		State       string `json:"state"`
		ExternalKey *struct {
			System string `json:"system"`
			ID     string `json:"id"`
		} `json:"externalKey"`
		Source *struct {
			Kind string `json:"kind"`
			At   string `json:"at"`
			By   string `json:"by"`
		} `json:"source"`
	}
	if len(bytes.TrimSpace(raw)) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &in); err != nil {
			return world.Entity{}, err
		}
	}
	e := world.Entity{Type: in.Type, Name: in.Name, State: in.State}
	if in.ExternalKey != nil {
		e.ID = world.ExternalID(in.Type, in.ExternalKey.System, in.ExternalKey.ID)
	} else if in.Name != "" {
		e.ID = world.EntityIDFromName(in.Type, in.Name)
	}
	if in.Source != nil {
		e.Source = world.Provenance{Kind: in.Source.Kind, At: in.Source.At, By: in.Source.By}
	}
	if err := st.UpsertEntity(e); err != nil {
		return world.Entity{}, err
	}
	return e, nil
}

// ingestW writes the observations as JSONL and ingests them. A blank
// observation is skipped.
func ingestW(st *world.Store, obs []json.RawMessage) (world.Report, error) {
	var buf bytes.Buffer
	for _, raw := range obs {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			continue
		}
		buf.Write(raw)
		buf.WriteByte('\n')
	}
	return st.Ingest(&buf, 0)
}

// readAudit reads a JSONL log and numbers its rows from 1. A log not yet
// written is an empty audit.
func readAudit(path string) ([]map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	out := []map[string]any{}
	seq := 0
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		seq++
		row["seq"] = seq
		out = append(out, row)
	}
	return out, nil
}

// ok marks an answer as successful and adds its id.
func ok(id int, fields map[string]any) map[string]any {
	fields["id"] = id
	fields["ok"] = true
	return fields
}

// fail is an answer with ok false and the error.
func fail(id int, msg string) map[string]any {
	return map[string]any{"id": id, "ok": false, "error": msg}
}
