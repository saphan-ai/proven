// Package baseline compares Proven's dependency check with a frozen set of
// another tool's alerts, alert by alert.
//
// The reference is test/corpus/reference/dependabot-fbdf7d3/alerts.json: the
// alerts Dependabot raised on this repository, frozen with the commit they
// belong to. The repository at that commit is read as a change from an empty
// base to its manifests (Export), the same shape a first review of a new
// project has, so the dependency check reads every package the manifests pin
// and not only a diff. Analyze runs that check (internal/osv) on the evidence
// store the caller names and keeps what it returned. Report writes the
// comparison. Report is a pure function of the reference and the findings, so
// a test can regenerate it and compare it with the file in the repository.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Alert is one Dependabot alert as the reference keeps it.
type Alert struct {
	Number          int     `json:"number"`
	Severity        string  `json:"severity"`
	GHSA            string  `json:"ghsa"`
	CVE             string  `json:"cve"`
	Ecosystem       string  `json:"ecosystem"`
	Package         string  `json:"package"`
	Manifest        string  `json:"manifest"`
	VulnerableRange string  `json:"vulnerable_range"`
	Patched         string  `json:"patched"`
	State           string  `json:"state"`
	CreatedAt       string  `json:"created_at"`
	FixedAt         *string `json:"fixed_at"`
}

// Reference is the frozen alert set and the commit it belongs to.
type Reference struct {
	Repository    string  `json:"repository"`
	PinnedCommit  string  `json:"pinned_commit"`
	SnapshotTaken string  `json:"snapshot_taken"`
	Source        string  `json:"source"`
	Note          string  `json:"note"`
	Alerts        []Alert `json:"alerts"`
}

// LoadReference reads the reference file. A file with no alerts, or with an
// alert that names no package or manifest, is refused: a comparison against
// it would say nothing.
func LoadReference(path string) (Reference, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Reference{}, err
	}
	var ref Reference
	if err := json.Unmarshal(b, &ref); err != nil {
		return Reference{}, fmt.Errorf("baseline: %s: %w", path, err)
	}
	if ref.PinnedCommit == "" || len(ref.Alerts) == 0 {
		return Reference{}, fmt.Errorf("baseline: %s names no pinned commit or no alerts", path)
	}
	for _, a := range ref.Alerts {
		if a.Package == "" || a.Manifest == "" || (a.GHSA == "" && a.CVE == "") {
			return Reference{}, fmt.Errorf("baseline: %s: alert %d lacks a package, a manifest or an advisory id", path, a.Number)
		}
	}
	return ref, nil
}

// sameEcosystem compares the reference's ecosystem word with Proven's. The
// reference says npm and go; Proven says npm and Go.
func sameEcosystem(a, b string) bool { return strings.EqualFold(a, b) }

// ids is the advisory ids of the alert: its GHSA id and its CVE id.
func (a Alert) ids() []string {
	var out []string
	for _, id := range []string{a.GHSA, a.CVE} {
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}
