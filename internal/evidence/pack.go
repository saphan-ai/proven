// Package evidence writes and verifies the signed evidence pack of ADR-0035.
//
// A pack is one tar archive per review. It holds the findings, the trace of
// the run, every test and tool command with its output, every query with its
// rows, the model receipts, the reviewed commits and the settings used, and a
// MANIFEST.json that names each file with its SHA-256. MANIFEST.sig is an
// Ed25519 signature over the exact bytes of the manifest, made with a key the
// organisation configures. Verify checks the signature first, then every
// hash, then that no file is missing or extra, then the receipts.
//
// The archive is a tar and not a zip. A tar has no central directory to
// disagree with the entries, so an entry list read once is the whole list,
// and the writer can make the same bytes every time from the same input:
// sorted entries, fixed times, fixed owners.
//
// ADR-0005: every file passes through the review's Redactor before it is
// hashed, so no configured secret reaches the pack, and the private key never
// does. A pack holds source excerpts and query rows; where it is kept is the
// organisation's decision (docs/EVIDENCE_PACK.md).
package evidence

import (
	"archive/tar"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/service"
)

// Required lists the files every pack holds, besides the manifest and its
// signature. Verify refuses a pack that lacks one, whatever the signed
// manifest lists: a manifest that leaves out settings.json is signed and
// still not a pack.
var Required = []string{FindingsName, TraceName, TestsName, QueriesName, ReceiptsName, CommitsName, SettingsName}

// File names inside a pack. The manifest and its signature are not listed in
// the manifest: a file cannot hold its own hash.
const (
	ManifestName  = "MANIFEST.json"
	SignatureName = "MANIFEST.sig"
	FindingsName  = "findings.json"
	TraceName     = "trace.jsonl"
	TestsName     = "tests.json"
	QueriesName   = "queries.json"
	ReceiptsName  = "receipts.json"
	CommitsName   = "commits.json"
	SettingsName  = "settings.json"
)

// FormatVersion is the manifest's version. A manifest of another version is refused.
const FormatVersion = 1

// Input is everything one pack is made from.
type Input struct {
	Result review.Result
	// Trace is the JSONL of this run only. The writer redacts it again.
	Trace []byte
	// Commits are the commits the review was asked about, as the shell named them and as git resolved them.
	Commits Commits
	// Settings are the settings the run used, secrets already reduced to "set" or "unset".
	Settings map[string]string
	Redactor *redact.Redactor
	Key      ed25519.PrivateKey
	Now      time.Time
}

// Commits names the reviewed change. Resolved fields are empty when git could not resolve the name.
type Commits struct {
	Target           string `json:"target"`
	Baseline         string `json:"baseline"`
	BaselineResolved string `json:"baseline_resolved"`
	Head             string `json:"head"`
	HeadResolved     string `json:"head_resolved"`
}

// Manifest is MANIFEST.json.
type Manifest struct {
	Version int       `json:"version"`
	Created string    `json:"created"`
	KeyID   string    `json:"key_id"`
	Commits Commits   `json:"commits"`
	Models  []Model   `json:"models"`
	Files   []FileRef `json:"files"`
}

// Model is one model use as the provider's receipt names it.
type Model struct {
	Agent       string `json:"agent"`
	Route       string `json:"route"`
	Model       string `json:"model"`
	ReceiptID   string `json:"receipt_id"`
	FromReceipt bool   `json:"from_receipt"`
}

// FileRef is one file of the pack with its hash.
type FileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

// KeyID is the SHA-256 of a public key's raw bytes, in hex. It names the key
// a pack was signed with and never reveals the private half.
func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])
}

// Build returns the archive's bytes. It refuses an Input with no key, and it
// refuses a pack that its own verification does not accept: a pack no
// auditor could verify is not handed back with exit 0.
func Build(in Input) ([]byte, error) {
	pack, err := build(in)
	if err != nil {
		return nil, err
	}
	if _, err := Verify(pack, in.Key.Public().(ed25519.PublicKey)); err != nil {
		return nil, fmt.Errorf("evidence pack: the pack just built does not verify: %w", err)
	}
	return pack, nil
}

// build makes the pack without verifying it. Tests use it to make a pack
// that Verify must refuse.
func build(in Input) ([]byte, error) {
	if len(in.Key) != ed25519.PrivateKeySize {
		return nil, errors.New("evidence pack: no signing key")
	}
	// The redactor replaces every configured value, but a value this short would
	// also replace ordinary text in the pack, and a pack is evidence: no short
	// secret is kept out of it by luck, so none is accepted. The refusal names
	// the variables and never their values.
	if short := in.Redactor.Shorter(service.MinSecretBytes); len(short) > 0 {
		return nil, fmt.Errorf("evidence pack: not written, because the configured secret %s is shorter than %d bytes and cannot be kept out of a pack; set a longer value", strings.Join(short, ", "), service.MinSecretBytes)
	}
	files, err := collect(in)
	if err != nil {
		return nil, err
	}
	// The manifest's model list is read back from the redacted receipts.json,
	// so the two cannot differ and no secret reaches the signed manifest.
	var redactedModels []Model
	if err := json.Unmarshal(files[ReceiptsName], &redactedModels); err != nil {
		return nil, fmt.Errorf("evidence pack: %s after redaction: %w", ReceiptsName, err)
	}
	m := Manifest{
		Version: FormatVersion,
		Created: in.Now.UTC().Format(time.RFC3339),
		KeyID:   KeyID(in.Key.Public().(ed25519.PublicKey)),
		Commits: Commits{
			Target:           in.Redactor.Text(in.Commits.Target),
			Baseline:         in.Redactor.Text(in.Commits.Baseline),
			BaselineResolved: in.Commits.BaselineResolved,
			Head:             in.Redactor.Text(in.Commits.Head),
			HeadResolved:     in.Commits.HeadResolved,
		},
		Models: redactedModels,
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		m.Files = append(m.Files, FileRef{Path: name, SHA256: hex.EncodeToString(sum[:]), Size: len(files[name])})
	}
	manifest, err := marshal(m)
	if err != nil {
		return nil, err
	}
	// The signature covers the manifest's bytes exactly as written, so a verifier needs no canonical form.
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(in.Key, manifest)) + "\n"
	files[ManifestName] = manifest
	files[SignatureName] = []byte(sig)
	return archive(files)
}

// WriteFile writes the pack to path through a temporary file, so a failed write leaves no half pack.
func WriteFile(path string, pack []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".evidence-pack-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(pack); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Chmod(name, 0o600); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

// collect makes every payload file of the pack and redacts it.
func collect(in Input) (map[string][]byte, error) {
	r := in.Redactor
	out := map[string][]byte{}
	put := func(name string, v any) error {
		b, err := marshal(v)
		if err != nil {
			return fmt.Errorf("evidence pack: %s: %w", name, err)
		}
		out[name] = []byte(r.Text(string(b)))
		return nil
	}
	if err := put(FindingsName, in.Result); err != nil {
		return nil, err
	}
	if err := put(TestsName, tests(in.Result)); err != nil {
		return nil, err
	}
	if err := put(QueriesName, queries(in.Result)); err != nil {
		return nil, err
	}
	if err := put(ReceiptsName, models(in.Result)); err != nil {
		return nil, err
	}
	if err := put(CommitsName, in.Commits); err != nil {
		return nil, err
	}
	settings := in.Settings
	if settings == nil {
		settings = map[string]string{}
	}
	if err := put(SettingsName, settings); err != nil {
		return nil, err
	}
	out[TraceName] = []byte(r.Text(string(in.Trace)))
	return out, nil
}

// marshal is indented JSON that keeps <, > and & as written: code and SQL are
// evidence, and an escaped copy would not match what was run.
func marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// archive writes the files as one tar with sorted entries and fixed metadata.
func archive(files map[string][]byte) ([]byte, error) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, name := range names {
		hdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(files[name])),
			ModTime:  time.Unix(0, 0),
			Format:   tar.FormatPAX,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(files[name]); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// models copies the review's model uses; a nil list is an empty one.
func models(res review.Result) []Model {
	out := []Model{}
	for _, m := range res.Models {
		out = append(out, Model{Agent: m.Agent, Route: m.Route, Model: m.Model, ReceiptID: m.ReceiptID, FromReceipt: m.FromReceipt})
	}
	return out
}

// TestRun is one test or tool run with its command and output.
type TestRun struct {
	FindingID     string `json:"finding_id"`
	Kind          string `json:"kind"`
	Path          string `json:"path,omitempty"`
	Name          string `json:"name,omitempty"`
	Tool          string `json:"tool,omitempty"`
	Version       string `json:"version,omitempty"`
	Command       string `json:"command"`
	FailsOnChange bool   `json:"fails_on_change"`
	PassesOnBase  bool   `json:"passes_on_base"`
	Output        string `json:"output"`
	BaseOutput    string `json:"base_output,omitempty"`
}

// QueryRun is one query with the rows it returned.
type QueryRun struct {
	FindingID string     `json:"finding_id"`
	SQL       string     `json:"sql"`
	Engine    string     `json:"engine"`
	Snapshot  string     `json:"snapshot"`
	Columns   []string   `json:"columns"`
	Rows      [][]string `json:"rows"`
	// Bound is whether Proven tied the rows to the change; the Judge counts rows only then.
	Bound bool `json:"bound"`
}

func tests(res review.Result) []TestRun {
	out := []TestRun{}
	for _, f := range res.Findings {
		for _, e := range f.Evidence {
			switch {
			case e.Test != nil:
				t := e.Test
				out = append(out, TestRun{FindingID: f.ID, Kind: e.Kind, Path: t.Path, Name: t.Name, Command: t.Command,
					FailsOnChange: t.FailsOnChange, PassesOnBase: t.PassesOnBase, Output: t.Output})
			case e.Tool != nil:
				t := e.Tool
				out = append(out, TestRun{FindingID: f.ID, Kind: e.Kind, Tool: t.Tool, Version: t.Version, Command: t.Command,
					FailsOnChange: t.PresentOnChange, PassesOnBase: !t.PresentOnBase, Output: t.Output, BaseOutput: t.BaseOutput})
			}
		}
	}
	return out
}

func queries(res review.Result) []QueryRun {
	out := []QueryRun{}
	for _, f := range res.Findings {
		for _, e := range f.Evidence {
			if q := e.Query; q != nil {
				out = append(out, QueryRun{FindingID: f.ID, SQL: q.SQL, Engine: q.Engine, Snapshot: q.Snapshot, Columns: q.Columns, Rows: q.Rows, Bound: q.Bound})
			}
		}
	}
	return out
}
