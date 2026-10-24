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
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"
)

// maxEntry bounds one file read from a pack, so a pack that claims a huge
// file is refused before it is held in memory.
const maxEntry = 512 << 20

// maxPack bounds the whole file VerifyFile reads: nine files of at most
// maxEntry, with the headers and padding, fit in it.
const maxPack = 9*maxEntry + 1<<20

// The checks Verify makes, in order. A Refusal names one.
const (
	CheckFormat    = "format"
	CheckSignature = "signature"
	CheckHash      = "hash"
	CheckMissing   = "missing"
	CheckExtra     = "extra"
	CheckReceipt   = "receipt"
)

// Refusal is a pack Verify did not accept: which check failed and on which file.
type Refusal struct {
	Check  string
	File   string
	Detail string
}

// Error is one line a terminal can show. The file name comes from the pack,
// which nobody has signed yet when a refusal is made, so it is escaped.
func (r *Refusal) Error() string {
	msg := "evidence pack refused: " + r.Check + " check failed"
	if r.File != "" {
		msg += " on " + Visible(r.File)
	}
	return msg + ": " + Visible(r.Detail)
}

// Visible returns s with every character that could act on a terminal shown
// as a visible escape: C0, DEL and C1 controls as \xNN, bidirectional
// controls and invisible marks as \uNNNN, and a byte that is not valid UTF-8
// as \xNN of that byte. Everything else is written unchanged.
func Visible(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			fmt.Fprintf(&b, `\x%02x`, r)
		case IsBidi(r):
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}

// IsBidi reports a bidirectional control or an invisible mark that can
// reorder or hide the text around it: U+061C, U+200B to U+200F, U+2028 to
// U+202E, U+2060 to U+2069 and U+FEFF.
func IsBidi(r rune) bool {
	return r == 0x061c || (r >= 0x200b && r <= 0x200f) || (r >= 0x2028 && r <= 0x202e) || (r >= 0x2060 && r <= 0x2069) || r == 0xfeff
}

func refuse(check, file, format string, args ...any) *Refusal {
	return &Refusal{Check: check, File: file, Detail: fmt.Sprintf(format, args...)}
}

// Report is what a verified pack says about the review.
type Report struct {
	Created      string
	KeyID        string
	Commits      Commits
	Models       []Model
	Files        []FileRef
	Findings     int
	ProofCounts  map[string]int
	Tests        int
	Queries      int
	Receipted    int // model uses that carry a provider receipt id, each found in the trace
	Unreceipted  int // model uses with no provider receipt
	TraceEvents  int
	FailedSvcs   []string
	NotChecked   []string
	ChecksPassed []string
}

// NotChecked is what Verify cannot check offline. It is printed with every report.
var notChecked = []string{
	"the provider's own signature on each receipt: a pack keeps the receipt id, route and model, not the signed body",
	"that the review's reading of the code was right: a pack proves what was recorded, not that it was true",
}

// VerifyFile reads the pack at path and verifies it against pub.
func VerifyFile(packPath string, pub ed25519.PublicKey) (*Report, error) {
	f, err := os.Open(packPath)
	if err != nil {
		return nil, refuse(CheckFormat, "", "the pack cannot be read: %v", err)
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, maxPack+1))
	if err != nil {
		return nil, refuse(CheckFormat, "", "the pack cannot be read: %v", err)
	}
	if len(body) > maxPack {
		return nil, refuse(CheckFormat, "", "larger than the %d byte bound", maxPack)
	}
	return Verify(body, pub)
}

// Verify checks a pack: the signature over the manifest, the hash of every
// file, that no file is missing or extra, and every receipt. The first
// failure is returned as a *Refusal naming the check and the file.
func Verify(pack []byte, pub ed25519.PublicKey) (*Report, error) {
	files, err := readArchive(pack)
	if err != nil {
		return nil, err
	}
	// The pack is one byte string: the archive the writer would make from these
	// files, and nothing else. Header metadata, padding and trailing bytes lie
	// outside every hash, so they are held to the canonical form here.
	canonical, err := archive(files)
	if err != nil {
		return nil, refuse(CheckFormat, "", "the archive cannot be rebuilt: %v", err)
	}
	if !bytes.Equal(canonical, pack) {
		return nil, refuse(CheckFormat, "", "the archive is not the one the writer makes from its files (header metadata, padding or trailing bytes differ)")
	}
	manifestBytes, ok := files[ManifestName]
	if !ok {
		return nil, refuse(CheckMissing, ManifestName, "the pack has no manifest")
	}
	sigBytes, ok := files[SignatureName]
	if !ok {
		return nil, refuse(CheckMissing, SignatureName, "the pack has no signature")
	}
	// The signature comes first: every later check reads the manifest, and a
	// manifest nobody signed proves nothing about the files it lists.
	// The signature file is the base64 and one newline, byte for byte: no
	// hash covers it, so no whitespace is trimmed from it.
	sig, err := base64.StdEncoding.DecodeString(string(bytes.TrimSuffix(sigBytes, []byte("\n"))))
	if err != nil || len(sig) != ed25519.SignatureSize || string(sigBytes) != base64.StdEncoding.EncodeToString(sig)+"\n" {
		return nil, refuse(CheckSignature, SignatureName, "not an Ed25519 signature")
	}
	if len(pub) != ed25519.PublicKeySize || !ed25519.Verify(pub, manifestBytes, sig) {
		return nil, refuse(CheckSignature, ManifestName, "the signature does not match the manifest under key %s", KeyID(pub))
	}
	var m Manifest
	dec := json.NewDecoder(bytes.NewReader(manifestBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return nil, refuse(CheckFormat, ManifestName, "the manifest is not readable: %v", err)
	}
	if m.Version != FormatVersion {
		return nil, refuse(CheckFormat, ManifestName, "version %d, this build reads version %d", m.Version, FormatVersion)
	}
	if m.KeyID != KeyID(pub) {
		return nil, refuse(CheckSignature, ManifestName, "the manifest names key %s, verified with key %s", m.KeyID, KeyID(pub))
	}
	listed := map[string]bool{}
	for _, ref := range m.Files {
		if listed[ref.Path] {
			return nil, refuse(CheckFormat, ref.Path, "the manifest lists the file twice")
		}
		listed[ref.Path] = true
		data, present := files[ref.Path]
		if !present {
			return nil, refuse(CheckMissing, ref.Path, "the manifest lists it and the pack does not hold it")
		}
		if ref.Size != len(data) {
			return nil, refuse(CheckHash, ref.Path, "size is %d, the manifest says %d", len(data), ref.Size)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != ref.SHA256 {
			return nil, refuse(CheckHash, ref.Path, "sha256 is %s, the manifest says %s", got, ref.SHA256)
		}
	}
	for _, name := range Required {
		if !listed[name] {
			return nil, refuse(CheckMissing, name, "the pack format requires it and the manifest does not list it")
		}
	}
	var extra []string
	for name := range files {
		if name != ManifestName && name != SignatureName && !listed[name] {
			extra = append(extra, name)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		return nil, refuse(CheckExtra, extra[0], "the pack holds a file the manifest does not list (%d extra)", len(extra))
	}
	rep := &Report{
		Created: m.Created, KeyID: m.KeyID, Commits: m.Commits, Models: m.Models, Files: m.Files,
		NotChecked:   append([]string(nil), notChecked...),
		ChecksPassed: []string{"signature", "hash of every file", "no file missing or extra", "receipts"},
	}
	if err := checkReceipts(files, m, rep); err != nil {
		return nil, err
	}
	if err := describe(files, rep); err != nil {
		return nil, err
	}
	return rep, nil
}

// readArchive reads every entry of the tar. A duplicate name, a name that is
// not a plain file name, and an entry that is not a regular file are refused:
// the pack is a flat list of files, and an archive that says otherwise is not one.
func readArchive(pack []byte) (map[string][]byte, error) {
	files := map[string][]byte{}
	tr := tar.NewReader(bytes.NewReader(pack))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, refuse(CheckFormat, "", "not a readable tar archive: %v", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			return nil, refuse(CheckFormat, hdr.Name, "an entry that is not a regular file")
		}
		if hdr.Name == "" || hdr.Name != path.Base(hdr.Name) || hdr.Name == "." || hdr.Name == ".." {
			return nil, refuse(CheckFormat, hdr.Name, "the name is not a plain file name")
		}
		if _, dup := files[hdr.Name]; dup {
			return nil, refuse(CheckFormat, hdr.Name, "the archive holds the name twice")
		}
		if hdr.Size > maxEntry {
			return nil, refuse(CheckFormat, hdr.Name, "larger than the %d byte bound", maxEntry)
		}
		data, err := io.ReadAll(io.LimitReader(tr, maxEntry+1))
		if err != nil {
			return nil, refuse(CheckFormat, hdr.Name, "unreadable: %v", err)
		}
		files[hdr.Name] = data
	}
	if len(files) == 0 {
		return nil, refuse(CheckFormat, "", "the archive holds no file")
	}
	return files, nil
}

// checkReceipts checks what a pack can show about its receipts offline: that
// receipts.json is the manifest's own model list, that a use named as taken
// from a receipt carries an id and a model, and that the trace holds a model
// event with that id and that model. A receipt's own signature is not in the pack.
func checkReceipts(files map[string][]byte, m Manifest, rep *Report) error {
	var receipts []Model
	if err := json.Unmarshal(files[ReceiptsName], &receipts); err != nil {
		return refuse(CheckReceipt, ReceiptsName, "not readable: %v", err)
	}
	if len(receipts) != len(m.Models) {
		return refuse(CheckReceipt, ReceiptsName, "holds %d receipts, the manifest names %d model uses", len(receipts), len(m.Models))
	}
	for i := range receipts {
		if receipts[i] != m.Models[i] {
			return refuse(CheckReceipt, ReceiptsName, "receipt %d differs from the manifest's model use", i+1)
		}
	}
	inTrace := map[string][]string{}
	sc := bytes.Split(files[TraceName], []byte("\n"))
	for _, line := range sc {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var ev struct {
			Model *struct {
				Model     string `json:"model"`
				ReceiptID string `json:"receipt_id"`
			} `json:"model"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			return refuse(CheckFormat, TraceName, "a line is not JSON: %v", err)
		}
		rep.TraceEvents++
		if ev.Model != nil && ev.Model.ReceiptID != "" {
			inTrace[ev.Model.ReceiptID] = append(inTrace[ev.Model.ReceiptID], ev.Model.Model)
		}
	}
	for _, r := range receipts {
		if !r.FromReceipt {
			rep.Unreceipted++
			continue
		}
		if r.ReceiptID == "" || r.Model == "" || r.Model == "unknown" {
			return refuse(CheckReceipt, ReceiptsName, "the %s use is marked as taken from a receipt and has no receipt id or no model", r.Agent)
		}
		found := false
		for _, model := range inTrace[r.ReceiptID] {
			if model == r.Model {
				found = true
			}
		}
		if !found {
			return refuse(CheckReceipt, TraceName, "receipt %s (%s, %s) has no matching model event in the trace", r.ReceiptID, r.Agent, r.Model)
		}
		rep.Receipted++
	}
	return nil
}

// describe fills the report's account of what was reviewed and on what evidence.
func describe(files map[string][]byte, rep *Report) error {
	var res struct {
		Findings       []json.RawMessage
		ProofCounts    map[string]int
		FailedServices []struct{ Service, Reason string }
	}
	if err := json.Unmarshal(files[FindingsName], &res); err != nil {
		return refuse(CheckFormat, FindingsName, "not readable: %v", err)
	}
	rep.Findings = len(res.Findings)
	rep.ProofCounts = res.ProofCounts
	for _, f := range res.FailedServices {
		rep.FailedSvcs = append(rep.FailedSvcs, f.Service)
	}
	var tests, queries []json.RawMessage
	if err := json.Unmarshal(files[TestsName], &tests); err != nil {
		return refuse(CheckFormat, TestsName, "not readable: %v", err)
	}
	if err := json.Unmarshal(files[QueriesName], &queries); err != nil {
		return refuse(CheckFormat, QueriesName, "not readable: %v", err)
	}
	rep.Tests, rep.Queries = len(tests), len(queries)
	return nil
}
