package baseline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
)

// Modes of a run. A manifest-only run reads the manifests and asks no store.
const (
	ModeEvidence     = "evidence"
	ModeManifestOnly = "manifest-only"
)

// The two files a run writes beside the reference.
const (
	FindingsFile = "proven-findings.json"
	ReportFile   = "REPORT.md"
)

// LoadFindings reads a findings file a run wrote.
func LoadFindings(path string) (Findings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Findings{}, err
	}
	var f Findings
	if err := json.Unmarshal(b, &f); err != nil {
		return Findings{}, fmt.Errorf("baseline: %s: %w", path, err)
	}
	return f, nil
}

// evidenceIDLen is how many hex digits of the digest an evidence id keeps.
const evidenceIDLen = 12

// Findings is what one run of the dependency check kept. It is the input of
// Report and is committed beside it.
type Findings struct {
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	Mode       string `json:"mode"`
	// Evidence says in one phrase what the verdicts rest on.
	Evidence string `json:"evidence"`
	// Store is the evidence store that answered: zetaris, local or couchdb.
	// It is empty on a manifest-only run.
	Store string `json:"store"`
	// Manifests is every manifest or lock file Proven reads at the commit.
	Manifests []string `json:"manifests"`
	// ReadsTotal is how many package versions the manifests pin.
	ReadsTotal int `json:"reads_total"`
	// Reads is the pins of the packages the reference names, in the files it names.
	Reads []Read `json:"reads"`
	// Matches is the packages whose query returned a row and a version
	// inside a range, and the packages the reference names whatever came back.
	Matches []Match `json:"matches"`
}

// Read is one package version a manifest pins.
type Read struct {
	File      string `json:"file"`
	Ecosystem string `json:"ecosystem"`
	Package   string `json:"package"`
	Version   string `json:"version"`
}

// Match is one package version and the rows the OSV query returned for it.
type Match struct {
	File       string     `json:"file"`
	Ecosystem  string     `json:"ecosystem"`
	Package    string     `json:"package"`
	Version    string     `json:"version"`
	EvidenceID string     `json:"evidence_id"`
	Engine     string     `json:"engine"`
	Snapshot   string     `json:"snapshot"`
	Advisories []Advisory `json:"advisories"`
	// KEV is the CVE aliases of an affecting advisory that the KEV view holds.
	KEV []string `json:"kev,omitempty"`
}

// Advisory is one OSV row and Proven's comparison of the version with it.
type Advisory struct {
	ID           string `json:"id"`
	Aliases      string `json:"aliases"`
	Introduced   string `json:"introduced"`
	Fixed        string `json:"fixed"`
	LastAffected string `json:"last_affected"`
	Affected     bool   `json:"affected"`
}

// Options are the settings of one run.
type Options struct {
	// Commit is the commit the tree was exported from.
	Commit string
	// Evidence is the phrase every verdict line carries.
	Evidence string
	// Store names the evidence store that run answers; empty for manifest-only.
	Store string
	// Table and KEVView name the OSV table and the KEV view; empty is the default.
	Table   string
	KEVView string
}

// Export writes the manifests of rev in repo into dir as a change from an
// empty base: dir/base is empty and dir/change holds every manifest and lock
// file a registered adapter reads. Only those files are kept, so the check
// reads the dependencies and nothing else. It returns their paths, sorted.
// Why a change from an empty base: the dependency check reads the packages a
// change adds or raises, and a first commit of a project adds all of them.
func Export(repo, rev, dir string) ([]string, error) {
	scratch, err := os.MkdirTemp(dir, "export-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch) // the export is a copy; a failed removal leaves only scratch files
	if _, err := change.ExportCommit(repo, rev, scratch); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, "base"), 0o755); err != nil {
		return nil, err
	}
	var kept []string
	err = filepath.Walk(scratch, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(scratch, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !readsManifest(rel) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, "change", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		kept = append(kept, rel)
		return os.WriteFile(dst, b, 0o644)
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(kept)
	return kept, nil
}

// readsManifest reports whether a registered adapter reads the path.
func readsManifest(path string) bool {
	for _, r := range manifest.Readers() {
		if r.Match != nil && r.Match(path) {
			return true
		}
	}
	return false
}

// Analyze runs the dependency check over the tree Export wrote and keeps what
// it returned. A nil run is a manifest-only run: the manifests are read and no
// advisory is asked for. A store that is down refuses the run, because a
// missing answer is not "missed".
func Analyze(ctx context.Context, tree string, manifests []string, ref Reference, run query.Runner, opt Options) (Findings, error) {
	files, err := change.Read(tree)
	if err != nil {
		return Findings{}, err
	}
	hits, _, err := manifest.Collect(files, manifest.Readers())
	if err != nil {
		return Findings{}, err
	}
	f := Findings{
		Repository: ref.Repository, Commit: opt.Commit, Mode: ModeManifestOnly,
		Evidence: opt.Evidence, Store: opt.Store, Manifests: append([]string{}, manifests...),
		ReadsTotal: len(hits), Reads: []Read{}, Matches: []Match{},
	}
	for _, h := range hits {
		if named(ref, h.File, h.Package) {
			f.Reads = append(f.Reads, Read{File: h.File, Ecosystem: h.Ecosystem, Package: h.Package, Version: h.Version})
		}
	}
	if run == nil {
		return f, nil
	}
	f.Mode = ModeEvidence
	if opt.KEVView != "" {
		ctx = osv.WithKEVView(ctx, opt.KEVView)
	}
	out, err := osv.Check(ctx, run, tree, opt.Table, "")
	if err != nil {
		return Findings{}, err
	}
	if out.Down != nil {
		return Findings{}, fmt.Errorf("baseline: evidence service %s is down: %s", out.Down.Service, out.Down.Reason)
	}
	for _, m := range out.Matches {
		if !affects(m) && !named(ref, m.File, m.Package) {
			continue
		}
		f.Matches = append(f.Matches, toMatch(m))
	}
	return f, nil
}

// named reports whether the reference has an alert for the package in the file.
func named(ref Reference, file, pkg string) bool {
	for _, a := range ref.Alerts {
		if a.Manifest == file && a.Package == pkg {
			return true
		}
	}
	return false
}

// affects reports whether the version lies inside at least one returned range.
func affects(m osv.Match) bool {
	for _, a := range m.Advisories {
		if a.Affected {
			return true
		}
	}
	return false
}

func toMatch(m osv.Match) Match {
	out := Match{
		File: m.File, Ecosystem: m.Ecosystem, Package: m.Package, Version: m.Version,
		EvidenceID: evidenceID(m), Engine: m.Engine, Snapshot: m.Snapshot,
		Advisories: make([]Advisory, 0, len(m.Advisories)), KEV: m.KEV,
	}
	for _, a := range m.Advisories {
		out.Advisories = append(out.Advisories, Advisory{
			ID: a.Row.ID, Aliases: a.Row.Aliases, Introduced: a.Row.Introduced, Fixed: a.Row.Fixed,
			LastAffected: a.Row.LastAffected, Affected: a.Affected,
		})
	}
	return out
}

// evidenceID names the evidence of one package: a digest of the query as the
// store ran it and the rows it returned, so the same answer has the same id on
// every run and a different answer does not.
func evidenceID(m osv.Match) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\n%s\n", m.Form, m.Query)
	for _, a := range m.Advisories {
		h.Write([]byte(strings.Join(a.Row.Fields(), "\x1f") + "\n"))
	}
	return "ev-" + hex.EncodeToString(h.Sum(nil))[:evidenceIDLen]
}
