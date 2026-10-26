package evidence

import (
	"archive/tar"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
)

const secretMarker = "SECRET-MARKER-9f3a1c77d2"

// sampleInput is a review result with a test, a query and a receipted model,
// and a trace that holds the matching model event.
func sampleInput(t *testing.T) (Input, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	res := review.Result{
		Findings: []review.Finding{{
			ID: "F1", RuleID: "discount.negative", Title: "negative total", Message: "total < 0 for a 150% discount",
			Severity: review.SeverityError, ProofLevel: review.ProvedByTest,
			Evidence: []review.Evidence{
				{Kind: review.EvidenceTest, Test: &review.TestEvidence{Path: "discount_test.go", Name: "TestNeg", Command: "go test ./...", FailsOnChange: true, PassesOnBase: true, Output: "FAIL: TestNeg a<b"}},
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{SQL: "select 1", Engine: "local-sample", Snapshot: "s1", Columns: []string{"n"}, Rows: [][]string{{"1"}}, Bound: true}},
			},
			Refutation: &review.Refutation{},
		}},
		ProofCounts: map[review.ProofLevel]int{review.ProvedByTest: 1, review.ProvedByTool: 0, review.ProvedByData: 0, review.AgreedNotProven: 0},
		Models:      []review.ModelUse{{Agent: "planner", Route: "nvidia", Model: "nvidia/nemotron", ReceiptID: "rcpt-1", FromReceipt: true}},
	}
	trace := []byte(`{"run_id":"r","time":"t","kind":"claim","agent":"planner","text":"x"}` + "\n" +
		`{"run_id":"r","time":"t","kind":"model","agent":"planner","text":"","model":{"model":"nvidia/nemotron","receipt_id":"rcpt-1","from_receipt":true}}` + "\n")
	return Input{
		Result:   res,
		Trace:    trace,
		Commits:  Commits{Target: ".", Baseline: "main", BaselineResolved: "aaaa", Head: "HEAD", HeadResolved: "bbbb"},
		Settings: map[string]string{"sample": "true"},
		Redactor: redact.New(map[string]string{"METERLESS_RUNTIME_TOKEN": secretMarker}),
		Key:      priv,
		Now:      time.Date(2026, 10, 24, 1, 2, 3, 0, time.UTC),
	}, pub
}

// rewrite returns the pack with edit applied to its file set. A nil result from edit drops the file.
func rewrite(t *testing.T, pack []byte, edit func(name string, body []byte) []byte, add map[string][]byte) []byte {
	t.Helper()
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	for name, body := range files {
		if nb := edit(name, body); nb != nil {
			out[name] = nb
		}
	}
	for name, body := range add {
		out[name] = body
	}
	b, err := archive(out)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func keep(_ string, body []byte) []byte { return body }

func wantRefusal(t *testing.T, err error, check, file string) {
	t.Helper()
	var r *Refusal
	if !errors.As(err, &r) {
		t.Fatalf("want a refusal on %s %s, got %v", check, file, err)
	}
	if r.Check != check || r.File != file {
		t.Fatalf("refusal is %s on %q, want %s on %q (%v)", r.Check, r.File, check, file, err)
	}
	if !strings.Contains(err.Error(), file) {
		t.Fatalf("the refusal does not name the file %q: %v", file, err)
	}
}

func TestAPackVerifies(t *testing.T) {
	in, pub := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Verify(pack, pub)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Findings != 1 || rep.Tests != 1 || rep.Queries != 1 || rep.Receipted != 1 || rep.TraceEvents != 2 {
		t.Fatalf("report: %+v", rep)
	}
	if rep.Commits.HeadResolved != "bbbb" || rep.Created != "2026-10-24T01:02:03Z" || len(rep.Models) != 1 {
		t.Fatalf("report: %+v", rep)
	}
	again, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	// Ed25519 signing is deterministic, so the same input is the same bytes.
	if !bytes.Equal(pack, again) {
		t.Fatal("the same input made two different packs")
	}
}

// The mutated field is the file's content with the manifest untouched.
func TestOneChangedByteIsRefusedNamingTheFile(t *testing.T) {
	in, pub := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{FindingsName, TraceName, TestsName, QueriesName, ReceiptsName, CommitsName, SettingsName} {
		t.Run(name, func(t *testing.T) {
			changed := rewrite(t, pack, func(n string, body []byte) []byte {
				if n != name {
					return body
				}
				b := append([]byte(nil), body...)
				b[len(b)/2] ^= 0x01
				return b
			}, nil)
			_, err := Verify(changed, pub)
			wantRefusal(t, err, CheckHash, name)
		})
	}
}

// The mutated field is the manifest, with the signature untouched.
func TestAChangedManifestIsRefusedBySignature(t *testing.T) {
	in, pub := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	changed := rewrite(t, pack, func(n string, body []byte) []byte {
		if n == ManifestName {
			return bytes.Replace(body, []byte(`"created": "2026`), []byte(`"created": "2027`), 1)
		}
		return body
	}, nil)
	if bytes.Equal(changed, pack) {
		t.Fatal("the manifest was not changed")
	}
	_, err = Verify(changed, pub)
	wantRefusal(t, err, CheckSignature, ManifestName)
}

func TestAMissingAndAnExtraFileAreRefused(t *testing.T) {
	in, pub := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	missing := rewrite(t, pack, func(n string, body []byte) []byte {
		if n == QueriesName {
			return nil
		}
		return body
	}, nil)
	_, err = Verify(missing, pub)
	wantRefusal(t, err, CheckMissing, QueriesName)

	extra := rewrite(t, pack, keep, map[string][]byte{"planted.txt": []byte("x")})
	_, err = Verify(extra, pub)
	wantRefusal(t, err, CheckExtra, "planted.txt")
}

func TestAWrongKeyIsRefusedBySignature(t *testing.T) {
	in, _ := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	other, _, _ := ed25519.GenerateKey(rand.Reader)
	_, err = Verify(pack, other)
	wantRefusal(t, err, CheckSignature, ManifestName)
}

func TestAReceiptThePackCannotShowIsRefused(t *testing.T) {
	in, pub := sampleInput(t)
	in.Trace = []byte(`{"run_id":"r","time":"t","kind":"claim","agent":"planner","text":"x"}` + "\n")
	pack, err := build(in)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(pack, pub)
	wantRefusal(t, err, CheckReceipt, TraceName)
	// Build refuses to hand back a pack its own verification refuses.
	if _, err := Build(in); err == nil {
		t.Fatal("Build returned a pack that does not verify")
	}

	in, pub = sampleInput(t)
	in.Result.Models[0].ReceiptID = ""
	pack, err = build(in)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(pack, pub)
	wantRefusal(t, err, CheckReceipt, ReceiptsName)
}

func TestNoConfiguredSecretReachesThePack(t *testing.T) {
	in, pub := sampleInput(t)
	in.Result.Findings[0].Message = "leaked " + secretMarker
	in.Result.Findings[0].Evidence[0].Test.Output = "token=" + secretMarker
	in.Result.Findings[0].Evidence[1].Query.Rows = [][]string{{secretMarker}}
	in.Trace = append(in.Trace, []byte(`{"run_id":"r","text":"`+secretMarker+`"}`+"\n")...)
	in.Settings["PROVEN_X"] = secretMarker
	in.Commits.Baseline = secretMarker
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(pack, []byte(secretMarker)) {
		t.Fatal("the pack holds the configured secret")
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if bytes.Contains(body, []byte(secretMarker)) {
			t.Fatalf("%s holds the configured secret", name)
		}
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatalf("a redacted pack must still verify: %v", err)
	}
}

func TestThePrivateKeyNeverEntersThePack(t *testing.T) {
	in, _ := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(pack, in.Key.Seed()) || bytes.Contains(pack, []byte(in.Key)) {
		t.Fatal("the pack holds the private key")
	}
}

func TestAnArchiveThatIsNotAFlatFileListIsRefused(t *testing.T) {
	_, pub := sampleInput(t)
	wantDetail := map[string]string{"duplicate": "holds the name twice", "path": "not a plain file name", "symlink": "not a regular file"}
	for name, mk := range map[string]func(*tar.Writer){
		"duplicate": func(tw *tar.Writer) {
			for i := 0; i < 2; i++ {
				tw.WriteHeader(&tar.Header{Name: "a", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg})
				tw.Write([]byte("x"))
			}
		},
		"path": func(tw *tar.Writer) {
			tw.WriteHeader(&tar.Header{Name: "../a", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg})
			tw.Write([]byte("x"))
		},
		"symlink": func(tw *tar.Writer) {
			tw.WriteHeader(&tar.Header{Name: "a", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink})
		},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			mk(tw)
			tw.Close()
			_, err := Verify(buf.Bytes(), pub)
			var r *Refusal
			if !errors.As(err, &r) || r.Check != CheckFormat || !strings.Contains(r.Detail, wantDetail[name]) {
				t.Fatalf("got %v", err)
			}
		})
	}
	_, err := Verify([]byte("not a tar"), pub)
	var r *Refusal
	if !errors.As(err, &r) || r.Check != CheckFormat {
		t.Fatalf("got %v", err)
	}
}

func TestKeyRefusalsNameTheSettingAndNeverTheKey(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadPrivateKey(EnvPackKey, "")
	var ke *KeyError
	if !errors.As(err, &ke) || ke.Setting != EnvPackKey || !strings.Contains(err.Error(), EnvPackKey+": not set") {
		t.Fatalf("unset: %v", err)
	}
	_, err = LoadPrivateKey(EnvPackKey, filepath.Join(dir, "absent.pem"))
	if !errors.As(err, &ke) || !strings.Contains(err.Error(), "absent.pem") {
		t.Fatalf("absent: %v", err)
	}
	bad := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(bad, []byte("KEY-BODY-"+secretMarker), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = LoadPrivateKey(EnvPackKey, bad)
	if !errors.As(err, &ke) || strings.Contains(err.Error(), secretMarker) {
		t.Fatalf("a key's content must never be printed: %v", err)
	}
	// A key of the wrong type is named as such.
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(priv)
	good := filepath.Join(dir, "good.pem")
	if err := os.WriteFile(good, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadPrivateKey(EnvPackKey, good)
	if err != nil || !bytes.Equal(got, priv) {
		t.Fatalf("good key: %v", err)
	}
	pder, _ := x509.MarshalPKIXPublicKey(pub)
	pubFile := filepath.Join(dir, "pub.pem")
	if err := os.WriteFile(pubFile, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pder}), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPublicKey(EnvPackPublicKey, pubFile); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPublicKey(EnvPackPublicKey, good); err == nil {
		t.Fatal("a private key file was accepted as a public key")
	}
}
