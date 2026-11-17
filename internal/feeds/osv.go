package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

var advisoryID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,199}$`)
var cveID = regexp.MustCompile(`^CVE-[0-9]{4}-[0-9]{4,}$`)

type osvConnector struct{}

func (osvConnector) Name() string { return "osv" }

type osvQuery struct {
	Package struct {
		Name      string `json:"name"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
	Version   string `json:"version"`
	PageToken string `json:"page_token,omitempty"`
}
type osvRef struct {
	ID       string `json:"id"`
	Modified string `json:"modified"`
}

func (osvConnector) Fetch(ctx context.Context, s *Session) []Snapshot {
	origin := s.Config.OSV
	d, r := s.Client.Downloaded, s.Client.Reused
	changed := s.Client.Changed
	fail := func(err error) []Snapshot {
		return []Snapshot{s.failure("osv", origin, err), s.failure("mal", origin, err)}
	}
	queries := make([]osvQuery, len(s.Packages))
	for i, p := range s.Packages {
		queries[i].Package.Name = p.Name
		queries[i].Package.Ecosystem = p.Ecosystem
		queries[i].Version = p.Version
	}
	refs := map[string]string{}
	for page := 0; len(queries) > 0; page++ {
		if page >= 20 {
			return fail(fmt.Errorf("OSV pagination limit reached"))
		}
		body, _ := json.Marshal(struct {
			Queries []osvQuery `json:"queries"`
		}{queries})
		b, err := s.Client.JSON(ctx, "POST", origin+"/querybatch", body)
		if err != nil {
			return fail(err)
		}
		var reply struct {
			Results []struct {
				Vulns []osvRef `json:"vulns"`
				Next  string   `json:"next_page_token"`
			} `json:"results"`
		}
		if json.Unmarshal(b, &reply) != nil || len(reply.Results) != len(queries) {
			return fail(fmt.Errorf("OSV batch response did not match the requested packages"))
		}
		var next []osvQuery
		for i, result := range reply.Results {
			for _, v := range result.Vulns {
				if len(refs) >= 10000 {
					return fail(fmt.Errorf("OSV advisory budget exceeded 10000 identities"))
				}
				if !advisoryID.MatchString(v.ID) || v.Modified == "" {
					return fail(fmt.Errorf("OSV omitted advisory identity or modified time"))
				}
				refs[v.ID] = v.Modified
			}
			if result.Next != "" {
				if result.Next == queries[i].PageToken {
					return fail(fmt.Errorf("OSV pagination repeated a token"))
				}
				q := queries[i]
				q.PageToken = result.Next
				next = append(next, q)
			}
		}
		queries = next
	}
	tables := []store.Table{{Name: querytpl.DefaultOSVTable, Columns: osv.Columns, Rows: [][]string{}}, {Name: querytpl.DefaultMalTable, Columns: []string{"id", "ecosystem", "package", "summary"}, Rows: [][]string{}}}
	ids := make([]string, 0, len(refs))
	for id := range refs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		key := "OSV record " + origin + "/" + id
		old, have := s.Client.cached(key)
		var b []byte
		if have && old.Modified == refs[id] {
			if err := s.Client.consume(len(old.Body)); err != nil {
				return fail(err)
			}
			b = old.Body
			s.Client.Reused++
		} else {
			var err error
			b, err = s.Client.JSON(ctx, "GET", origin+"/vulns/"+url.PathEscape(id), nil)
			if err != nil {
				return fail(err)
			}
			if err = s.Client.write(key, cacheEntry{Modified: refs[id], Body: b}); err != nil {
				return fail(fmt.Errorf("OSV record cache could not be written"))
			}
		}
		var doc struct {
			ID, Modified, Withdrawn, Summary, Published string
			Aliases                                     []string
			Affected                                    []struct {
				Package struct{ Name, Ecosystem string }
			}
		}
		if json.Unmarshal(b, &doc) != nil || doc.ID != id || !sameOSVRevision(refs[id], doc.Modified) {
			return fail(fmt.Errorf("OSV record identity or modified time differs from the batch response"))
		}
		if doc.Withdrawn != "" {
			continue
		}
		s.Advisories[id] = b
		for _, alias := range append(doc.Aliases, id) {
			if cveID.MatchString(alias) {
				s.CVEs[alias] = true
			}
		}
		rows, err := osv.ParseJSON(b, id)
		if err != nil {
			return fail(fmt.Errorf("OSV record could not be normalized"))
		}
		for _, row := range rows {
			tables[0].Rows = append(tables[0].Rows, row.Fields())
		}
		if strings.HasPrefix(id, "MAL-") {
			for _, a := range doc.Affected {
				tables[1].Rows = append(tables[1].Rows, []string{id, a.Package.Ecosystem, a.Package.Name, doc.Summary})
			}
		}
	}
	var out []Snapshot
	index := store.Table{Name: querytpl.DefaultOSVIDView, Columns: []string{"id", "alias", "published", "severity"}, Rows: [][]string{}}
	for _, id := range ids {
		if raw, ok := s.Advisories[id]; ok {
			var doc struct {
				Aliases   []string
				Published string
				Severity  []struct{ Score string }
			}
			json.Unmarshal(raw, &doc)
			severity := ""
			if len(doc.Severity) > 0 {
				severity = doc.Severity[0].Score
			}
			index.Rows = append(index.Rows, []string{id, strings.Join(doc.Aliases, ","), doc.Published, severity})
		}
	}
	for i, name := range []string{"osv", "mal"} {
		status := s.status(name, origin, d, r)
		status.Rows = len(tables[i].Rows)
		if s.Client.Changed == changed {
			status.Status = "unchanged"
		}
		snapshot := Snapshot{Table: tables[i], Refresh: status}
		if i == 0 {
			snapshot.Additional = []store.Table{index}
		}
		out = append(out, snapshot)
	}
	return out
}

// Querybatch's index can expose microseconds while the full OSV record keeps
// nanoseconds (observed against the public API). Permit that precision loss,
// not a different microsecond, ID or revision. Keep the full record unchanged.
func sameOSVRevision(batch, record string) bool {
	a, err := time.Parse(time.RFC3339Nano, batch)
	if err != nil {
		return false
	}
	b, err := time.Parse(time.RFC3339Nano, record)
	if err != nil {
		return false
	}
	return a.Equal(b) || a.Equal(b.Truncate(time.Microsecond))
}
