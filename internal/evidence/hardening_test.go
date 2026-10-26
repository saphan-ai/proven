package evidence

import (
	"archive/tar"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// resign returns the pack with its manifest edited and signed again by key, so
// the only thing wrong with it is what edit changed.
func resign(t *testing.T, pack []byte, key ed25519.PrivateKey, edit func(manifest []byte) []byte) []byte {
	t.Helper()
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	manifest := edit(files[ManifestName])
	files[ManifestName] = manifest
	files[SignatureName] = []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(key, manifest)) + "\n")
	out, err := archive(files)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// editManifest decodes the manifest, applies edit and writes it back as the writer would.
func editManifest(t *testing.T, edit func(m *Manifest)) func([]byte) []byte {
	return func(b []byte) []byte {
		var m Manifest
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		edit(&m)
		out, err := marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
}

func builtPack(t *testing.T) ([]byte, Input, ed25519.PublicKey) {
	t.Helper()
	in, pub := sampleInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	return pack, in, pub
}

func wantCheck(t *testing.T, err error, check, detail string) {
	t.Helper()
	var r *Refusal
	if !errors.As(err, &r) {
		t.Fatalf("want a %s refusal, got %v", check, err)
	}
	if r.Check != check || !strings.Contains(r.Detail, detail) {
		t.Fatalf("refusal is %s %q, want %s with %q", r.Check, r.Detail, check, detail)
	}
}

// C1. The mutated field is the model use's route and model; the secret must reach no file.
func TestASecretInAModelUseReachesNoFileAndThePackVerifies(t *testing.T) {
	in, pub := sampleInput(t)
	in.Result.Models[0].Route = "https://x.example/" + secretMarker
	in.Result.Models = append(in.Result.Models, review.ModelUse{Agent: "judge", Route: "nvidia", Model: "nvidia/" + secretMarker})
	pack, err := Build(in)
	if err != nil {
		t.Fatalf("a secret in a model use must be redacted, not refused: %v", err)
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
	rep, err := Verify(pack, pub)
	if err != nil {
		t.Fatalf("a redacted pack must verify: %v", err)
	}
	if len(rep.Models) != 2 {
		t.Fatalf("models: %+v", rep.Models)
	}
}

// C2. The mutated field is the key file's mode, nothing else.
func TestAPrivateKeyThatGroupOrOthersCanReachIsRefused(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(priv)
	body := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	for mode, refused := range map[os.FileMode]bool{0o600: false, 0o400: false, 0o644: true, 0o640: true, 0o604: true, 0o660: true, 0o666: true, 0o610: true} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "pack.key")
			if err := os.WriteFile(path, body, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, mode); err != nil {
				t.Fatal(err)
			}
			got, err := LoadPrivateKey(EnvPackKey, path)
			if !refused {
				if err != nil || !bytes.Equal(got, priv) {
					t.Fatalf("an owner-only key must load: %v", err)
				}
				return
			}
			var ke *KeyError
			if !errors.As(err, &ke) || ke.Setting != EnvPackKey || ke.Path != path {
				t.Fatalf("want a KeyError naming the setting and the path, got %v", err)
			}
			if !strings.Contains(err.Error(), "chmod 600") || !strings.Contains(err.Error(), fmt.Sprintf("%04o", mode)) {
				t.Fatalf("the refusal must name the mode and the remedy: %v", err)
			}
			if got != nil {
				t.Fatal("a key was returned with the refusal")
			}
		})
	}
}

// C3. The mutated field is a name from the pack, which no one has signed yet.
func TestARefusalShowsNoControlOrBidirectionalCharacter(t *testing.T) {
	_, pub := sampleInput(t)
	for name, entry := range map[string]string{
		"escape":  "a\x1b]0;owned\x07\x1b[2J",
		"bidi":    "gpj" + string(rune(0x202e)) + "txt.exe",
		"c1":      "a\u009b31m",
		"invalid": "a\xffb",
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			tw.WriteHeader(&tar.Header{Name: entry, Linkname: "x", Typeflag: tar.TypeSymlink})
			tw.Close()
			_, err := Verify(buf.Bytes(), pub)
			var r *Refusal
			if !errors.As(err, &r) || r.File != entry {
				t.Fatalf("want a refusal that carries the name, got %v", err)
			}
			for _, c := range err.Error() {
				if c < 0x20 || c == 0x7f || (c >= 0x80 && c <= 0x9f) || IsBidi(c) {
					t.Fatalf("the refusal holds the raw character %U: %q", c, err.Error())
				}
			}
			if strings.ContainsRune(err.Error(), '�') {
				t.Fatalf("an invalid byte must be shown, not replaced: %q", err.Error())
			}
		})
	}
	rlo := string(rune(0x202e))
	if got := Visible("a\x1bb" + rlo + "c\xffd\ne"); got != `a\x1bb`+`\`+`u202ec\xffd\x0ae` {
		t.Fatalf("Visible: %q", got)
	}
	if got := Visible("plain é 日本"); got != "plain é 日本" {
		t.Fatalf("Visible changed plain text: %q", got)
	}
}

// fixChecksum writes the header checksum of the 512-byte block at off.
func fixChecksum(b []byte, off int) {
	blk := b[off : off+512]
	for i := 148; i < 156; i++ {
		blk[i] = ' '
	}
	sum := 0
	for _, c := range blk {
		sum += int(c)
	}
	copy(blk[148:156], fmt.Sprintf("%06o\x00 ", sum))
}

// C4. Each mutation leaves every hash and the signature true; only the container bytes change.
func TestThePackIsOneByteStringAndAnyOtherIsRefused(t *testing.T) {
	pack, _, pub := builtPack(t)
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	manifestLen := len(files[ManifestName])
	want := func(t *testing.T, mutated []byte) {
		t.Helper()
		if bytes.Equal(mutated, pack) {
			t.Fatal("the pack was not changed")
		}
		_, err := Verify(mutated, pub)
		wantCheck(t, err, CheckFormat, "writer makes")
	}
	t.Run("appended byte", func(t *testing.T) {
		want(t, append(append([]byte(nil), pack...), 0x00))
	})
	t.Run("appended text", func(t *testing.T) {
		want(t, append(append([]byte(nil), pack...), []byte("tail")...))
	})
	t.Run("header metadata", func(t *testing.T) {
		m := append([]byte(nil), pack...)
		copy(m[265:], "someone") // uname of the first entry
		fixChecksum(m, 0)
		want(t, m)
	})
	t.Run("header mtime", func(t *testing.T) {
		m := append([]byte(nil), pack...)
		copy(m[136:], "00000000001\x00") // mtime of the first entry
		fixChecksum(m, 0)
		want(t, m)
	})
	t.Run("padding", func(t *testing.T) {
		m := append([]byte(nil), pack...)
		if manifestLen%512 == 0 {
			t.Skip("the manifest has no padding")
		}
		m[512+manifestLen] = 0x01
		want(t, m)
	})
	t.Run("signature whitespace", func(t *testing.T) {
		m := rewrite(t, pack, func(n string, b []byte) []byte {
			if n == SignatureName {
				return append([]byte(" "), b...)
			}
			return b
		}, nil)
		_, err := Verify(m, pub)
		wantCheck(t, err, CheckSignature, "Ed25519")
	})
	if _, err := Verify(pack, pub); err != nil {
		t.Fatalf("the control pack must verify: %v", err)
	}
}

// C4. One bit flipped at each position of the pack, one position at a time:
// none may verify. Before the byte-equality check, 28.7% of them did.
func TestEveryChangedByteOfThePackIsRefused(t *testing.T) {
	pack, _, pub := builtPack(t)
	accepted := 0
	for i := range pack {
		mutated := append([]byte(nil), pack...)
		mutated[i] ^= 0x01
		if _, err := Verify(mutated, pub); err == nil {
			accepted++
			if accepted <= 3 {
				t.Errorf("byte %d of %d changed and the pack still verifies", i, len(pack))
			}
		}
	}
	if accepted > 0 {
		t.Fatalf("%d of %d single-bit changes still verify", accepted, len(pack))
	}
}

// C5. Each guard below is checked alone: the manifest is signed again, so only the guarded field is wrong.
func TestEachManifestGuardRefusesOnItsOwn(t *testing.T) {
	pack, in, pub := builtPack(t)
	cases := map[string]struct {
		edit   func([]byte) []byte
		check  string
		detail string
	}{
		"version": {editManifest(t, func(m *Manifest) { m.Version = FormatVersion + 1 }), CheckFormat, "version 2"},
		"key id":  {editManifest(t, func(m *Manifest) { m.KeyID = strings.Repeat("0", 64) }), CheckSignature, "names key"},
		"file listed twice": {editManifest(t, func(m *Manifest) {
			m.Files = append(m.Files, m.Files[0])
		}), CheckFormat, "twice"},
		"unknown field": {func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"version": 1,`), []byte(`"version": 1,`+"\n"+`  "extra": true,`), 1)
		}, CheckFormat, "not readable"},
		"receipts differ from the manifest": {editManifest(t, func(m *Manifest) { m.Models[0].Model = "nvidia/other" }), CheckReceipt, "differs"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mutated := resign(t, pack, in.Key, tc.edit)
			if bytes.Equal(mutated, pack) {
				t.Fatal("the manifest was not changed")
			}
			_, err := Verify(mutated, pub)
			wantCheck(t, err, tc.check, tc.detail)
		})
	}
}

// The mutated field is the header's size: a pack that claims a huge file is refused before it is read.
func TestAnEntryOverTheBoundIsRefusedBeforeItIsRead(t *testing.T) {
	_, pub := sampleInput(t)
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "big", Mode: 0o644, Size: maxEntry + 1, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	// No Close: the header is written and its data never is.
	_, err := Verify(buf.Bytes(), pub)
	wantCheck(t, err, CheckFormat, "byte bound")
	// One byte under the bound is not refused for its size.
	buf.Reset()
	tw = tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "big", Mode: 0o644, Size: maxEntry, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	_, err = Verify(buf.Bytes(), pub)
	var r *Refusal
	if errors.As(err, &r) && strings.Contains(r.Detail, "byte bound") {
		t.Fatalf("an entry at the bound was refused for its size: %v", err)
	}
}
