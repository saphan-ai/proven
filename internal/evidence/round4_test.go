package evidence

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
)

// A configured secret must reach no file of the pack whatever characters it
// holds. The pack writes JSON with HTML escaping off; a value holding a quote,
// an HTML character and a character %q escapes but JSON does not (U+00AD) is
// written by the pack in a spelling that is none of the other needles.
func TestASecretInThePacksOwnJSONSpellingIsRedacted(t *testing.T) {
	secret := "tok\"en<12­34"
	in, pub := sampleInput(t)
	in.Redactor = redact.New(map[string]string{"METERLESS_RUNTIME_TOKEN": secret})
	in.Result.Findings[0].Message = "leaked " + secret
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := marshal(secret)
	if err != nil {
		t.Fatal(err)
	}
	spelled := strings.Trim(strings.TrimSpace(string(enc)), `"`)
	if spelled == secret {
		t.Fatal("the control is void: the pack's spelling equals the raw value")
	}
	for name, body := range files {
		if bytes.Contains(body, []byte(spelled)) || bytes.Contains(body, []byte(secret)) {
			t.Errorf("%s holds the configured secret as %q", name, spelled)
		}
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatal(err)
	}
}

// A value that ends in a quote is written ending in an escaped quote; every
// spelling keeps it, so none is left half in the pack.
func TestASecretEndingInAQuoteIsRedactedWhole(t *testing.T) {
	secret := "tail-of-the-token-ends-in-\""
	in, _ := sampleInput(t)
	in.Redactor = redact.New(map[string]string{"METERLESS_RUNTIME_TOKEN": secret})
	in.Result.Findings[0].Message = "leaked " + secret
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(pack, []byte("tail-of-the-token")) {
		t.Fatal("a part of the secret is in the pack")
	}
}

// A configured secret shorter than the floor is refused by name before any file
// is made; the refusal never holds the value. The mutated field is the length.
func TestAConfiguredSecretShorterThanTheFloorRefusesThePack(t *testing.T) {
	for _, tc := range []struct {
		value   string
		refused bool
	}{
		{"abcde", true},
		{strings.Repeat("k", service.MinSecretBytes-1), true},
		{strings.Repeat("k", service.MinSecretBytes), false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			in, pub := sampleInput(t)
			in.Redactor = redact.New(map[string]string{"TEST_SECRET": tc.value})
			in.Result.Findings[0].Message = "service returned " + tc.value
			pack, err := Build(in)
			if !tc.refused {
				if err != nil {
					t.Fatalf("a secret at the floor must be redacted, not refused: %v", err)
				}
				if bytes.Contains(pack, []byte(tc.value)) {
					t.Fatal("the pack holds the secret")
				}
				if _, err := Verify(pack, pub); err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || pack != nil {
				t.Fatal("a pack was made with a short configured secret")
			}
			if !strings.Contains(err.Error(), "TEST_SECRET") || strings.Contains(err.Error(), tc.value) {
				t.Fatalf("the refusal must name the variable and never the value: %v", err)
			}
		})
	}
}

// The mutated field is the manifest's file list, signed again: a file the pack
// format requires, left out of the list and out of the archive, is a pack
// Verify must refuse, whatever the manifest says.
func TestRequiredPackFilesCannotBeOmitted(t *testing.T) {
	pack, in, pub := builtPack(t)
	for _, omitted := range Required {
		t.Run(omitted, func(t *testing.T) {
			dropped := resign(t, pack, in.Key, func(b []byte) []byte {
				var m Manifest
				if err := json.Unmarshal(b, &m); err != nil {
					t.Fatal(err)
				}
				refs := m.Files[:0]
				for _, ref := range m.Files {
					if ref.Path != omitted {
						refs = append(refs, ref)
					}
				}
				m.Files = refs
				out, err := marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				return out
			})
			// resign kept the file in the archive; take it out too, so the manifest and the archive agree.
			changed := rewrite(t, dropped, func(n string, b []byte) []byte {
				if n == omitted {
					return nil
				}
				return b
			}, nil)
			_, err := Verify(changed, pub)
			wantCheck(t, err, CheckMissing, "requires")
			wantRefusal(t, err, CheckMissing, omitted)
		})
	}
}

// The signature file is the base64 and one newline, byte for byte. Each
// spelling below decodes to the same 64 signature bytes, so only the
// byte-for-byte clause can refuse it; the archive is rebuilt from the changed
// file, so the canonical-archive check passes.
func TestANonCanonicalSignatureSpellingIsRefused(t *testing.T) {
	pack, _, pub := builtPack(t)
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	body := strings.TrimSuffix(string(files[SignatureName]), "\n")
	last := body[len(body)-3] // the char before "==": its 4 low bits are padding
	alt := strings.IndexByte(alphabet, last) + 1
	for name, s := range map[string]string{
		"carriage return": body + "\r\n",
		"inner newline":   body[:10] + "\n" + body[10:] + "\n",
		"padding bits":    body[:len(body)-3] + string(alphabet[alt]) + "==\n",
	} {
		t.Run(name, func(t *testing.T) {
			m := rewrite(t, pack, func(n string, b []byte) []byte {
				if n == SignatureName {
					return []byte(s)
				}
				return b
			}, nil)
			_, err := Verify(m, pub)
			wantCheck(t, err, CheckSignature, "Ed25519")
		})
	}
}

// The mutated field is the manifest's model count, signed again: one model use
// more than receipts.json holds must be refused by the count check, and one
// fewer must be refused, not panic.
func TestAModelCountThatDiffersFromTheReceiptsIsRefused(t *testing.T) {
	pack, in, pub := builtPack(t)
	for name, edit := range map[string]func(m *Manifest){
		"one more":  func(m *Manifest) { m.Models = append(m.Models, m.Models[0]) },
		"one fewer": func(m *Manifest) { m.Models = m.Models[:0] },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := resign(t, pack, in.Key, editManifest(t, edit))
			_, err := Verify(mutated, pub)
			wantCheck(t, err, CheckReceipt, "the manifest names")
		})
	}
}

// The manifest's commits are redacted field by field, apart from commits.json.
// The target and the head are the fields this test mutates.
func TestASecretInTheTargetOrHeadReachesNoFile(t *testing.T) {
	in, pub := sampleInput(t)
	in.Commits.Target = "/srv/" + secretMarker
	in.Commits.Head = "refs/heads/" + secretMarker
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if bytes.Contains(body, []byte(secretMarker)) {
			t.Errorf("%s holds the configured secret", name)
		}
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatal(err)
	}
}

// The mutated field is a listed file's size, with the hash true: the size the
// manifest signs is the size the file has.
func TestAListedSizeThatIsNotTheFilesIsRefused(t *testing.T) {
	pack, in, pub := builtPack(t)
	mutated := resign(t, pack, in.Key, editManifest(t, func(m *Manifest) { m.Files[0].Size++ }))
	_, err := Verify(mutated, pub)
	wantCheck(t, err, CheckHash, "size is")
}

// VerifyFile reads a pack from a path: the pack it reads is the one Verify
// accepts, and a path that is not a file is refused as a format failure.
func TestVerifyFileReadsAPackAndRefusesAMissingPath(t *testing.T) {
	pack, _, pub := builtPack(t)
	path := filepath.Join(t.TempDir(), "p.tar")
	if err := os.WriteFile(path, pack, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyFile(path, pub); err != nil {
		t.Fatal(err)
	}
	_, err := VerifyFile(filepath.Join(t.TempDir(), "absent"), pub)
	wantCheck(t, err, CheckFormat, "cannot be read")
}

// Required is the whole payload set the writer makes: a file the writer adds
// without naming it in Required would be a file Verify does not insist on.
func TestRequiredIsEveryPayloadFileTheWriterMakes(t *testing.T) {
	pack, _, _ := builtPack(t)
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(Required)+2 {
		t.Fatalf("the pack holds %d files, Required names %d and the manifest and signature make two more", len(files), len(Required))
	}
	for _, name := range Required {
		if _, ok := files[name]; !ok {
			t.Fatalf("the writer makes no %s", name)
		}
	}
}
