package osv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Columns is the CSV header, one flat row per affected range. It is also
// the column list the dependency check selects and reads by name.
// docs/OSV_TABLE.md says what each column holds.
var Columns = []string{
	"id", "aliases", "ecosystem", "package", "introduced", "fixed", "last_affected", "severity", "published",
}

// Row is one affected range of one advisory: one row of the OSV table.
type Row struct {
	ID           string
	Aliases      string
	Ecosystem    string
	Package      string
	Introduced   string
	Fixed        string
	LastAffected string
	Severity     string
	Published    string
}

// Fields is the CSV record in Columns order.
func (r Row) Fields() []string {
	return []string{
		r.ID, r.Aliases, r.Ecosystem, r.Package, r.Introduced, r.Fixed, r.LastAffected, r.Severity, r.Published,
	}
}

// WriteCSV writes the header and the rows.
func WriteCSV(w io.Writer, rows []Row) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(Columns); err != nil {
		return err
	}
	for _, row := range rows {
		if err := cw.Write(row.Fields()); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// ReadDir reads one JSON advisory per file from dir. It does not walk
// subdirectories. A file whose name does not end in .json is skipped.
// Rows are ordered by file name, then by the order of the affected entries.
func ReadDir(dir string) ([]Row, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	var out []Row
	for _, name := range names {
		rows, err := readFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

// advisory is the part of an OSV JSON record this converter reads.
type advisory struct {
	ID        string     `json:"id"`
	Aliases   []string   `json:"aliases"`
	Published string     `json:"published"`
	Severity  []sevScore `json:"severity"`
	Affected  []affected `json:"affected"`
}

// sevScore is one entry of an advisory's severity list.
type sevScore struct {
	Score string `json:"score"`
}

// affected is one package an advisory covers, with its ranges or versions.
type affected struct {
	Package struct {
		Ecosystem string `json:"ecosystem"`
		Name      string `json:"name"`
	} `json:"package"`
	Ranges []struct {
		Events []event `json:"events"`
	} `json:"ranges"`
	Versions          []string        `json:"versions"`
	EcosystemSpecific json.RawMessage `json:"ecosystem_specific"`
	DatabaseSpecific  json.RawMessage `json:"database_specific"`
}

// event is one point of an OSV range. Exactly one field is set.
type event struct {
	Introduced   string `json:"introduced"`
	Fixed        string `json:"fixed"`
	LastAffected string `json:"last_affected"`
}

// readFile turns one advisory into rows. An advisory with no id, or an
// affected entry with no package, refuses the whole run, so a broken file
// does not quietly lose ranges from the table.
func readFile(path string) ([]Row, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc advisory
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("osv: %s: %w", filepath.Base(path), err)
	}
	if strings.TrimSpace(doc.ID) == "" {
		return nil, fmt.Errorf("osv: %s: advisory has no id", filepath.Base(path))
	}
	aliases := joinAliases(doc.Aliases)
	published := strings.TrimSpace(doc.Published)
	top := ""
	if len(doc.Severity) > 0 {
		top = strings.TrimSpace(doc.Severity[0].Score)
	}
	var out []Row
	for i, aff := range doc.Affected {
		name := strings.TrimSpace(aff.Package.Name)
		eco := strings.TrimSpace(aff.Package.Ecosystem)
		if name == "" || eco == "" {
			return nil, fmt.Errorf("osv: %s: affected entry %d has no package", filepath.Base(path), i)
		}
		// The most specific severity wins: the database's own for this package,
		// then the ecosystem's, then the advisory's first score.
		sev := specificSeverity(aff.DatabaseSpecific)
		if sev == "" {
			sev = specificSeverity(aff.EcosystemSpecific)
		}
		if sev == "" {
			sev = top
		}
		var ranges []span
		for _, rg := range aff.Ranges {
			ranges = append(ranges, eventsToSpans(rg.Events)...)
		}
		// An entry with only a versions list becomes one row per version, each a
		// range of one version: introduced and last_affected are that version.
		if len(ranges) == 0 {
			for _, v := range aff.Versions {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				ranges = append(ranges, span{Introduced: v, LastAffected: v})
			}
		}
		for _, sp := range ranges {
			out = append(out, Row{
				ID: doc.ID, Aliases: aliases, Ecosystem: eco, Package: name,
				Introduced: sp.Introduced, Fixed: sp.Fixed, LastAffected: sp.LastAffected,
				Severity: sev, Published: published,
			})
		}
	}
	return out, nil
}

// span is one affected range before it becomes a Row.
type span struct {
	Introduced   string
	Fixed        string
	LastAffected string
}

// eventsToSpans turns the events of one OSV range into spans, in order. An
// introduced opens a span; a fixed or a last_affected closes it. A span
// still open when the next introduced comes, or at the end, has no upper
// bound: every version from introduced on is affected.
func eventsToSpans(events []event) []span {
	var out []span
	var cur *span
	flush := func() {
		if cur == nil {
			return
		}
		out = append(out, *cur)
		cur = nil
	}
	for _, e := range events {
		switch {
		case e.Introduced != "":
			if cur != nil {
				flush()
			}
			cur = &span{Introduced: e.Introduced}
		case e.Fixed != "":
			if cur == nil {
				cur = &span{}
			}
			cur.Fixed = e.Fixed
			flush()
		case e.LastAffected != "":
			if cur == nil {
				cur = &span{}
			}
			cur.LastAffected = e.LastAffected
			flush()
		}
	}
	flush()
	return out
}

// specificSeverity reads a "severity" string from a database_specific or
// ecosystem_specific object. Any other shape gives none, not an error,
// because those objects have no fixed schema.
func specificSeverity(raw json.RawMessage) string {
	if len(bytes.TrimSpace(raw)) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	s, _ := m["severity"].(string)
	return strings.TrimSpace(s)
}

// joinAliases trims, de-duplicates and sorts the aliases, so the column is
// the same however the source ordered them.
func joinAliases(in []string) string {
	seen := map[string]struct{}{}
	var out []string
	for _, a := range in {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if _, ok := seen[a]; ok {
			continue
		}
		seen[a] = struct{}{}
		out = append(out, a)
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}
