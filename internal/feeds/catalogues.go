package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

type kevConnector struct{}

func (kevConnector) Name() string { return "kev" }
func (kevConnector) Fetch(ctx context.Context, s *Session) []Snapshot {
	origin := s.Config.KEV
	d, r := s.Client.Downloaded, s.Client.Reused
	changed := s.Client.Changed
	b, err := s.Client.JSON(ctx, "GET", origin, nil)
	if err != nil {
		return []Snapshot{s.failure("kev", origin, err)}
	}
	var doc struct {
		DateReleased    string `json:"dateReleased"`
		Count           *int   `json:"count"`
		Vulnerabilities []struct {
			ID         string `json:"cveID"`
			Vendor     string `json:"vendorProject"`
			Product    string `json:"product"`
			Added      string `json:"dateAdded"`
			Ransomware string `json:"knownRansomwareCampaignUse"`
		} `json:"vulnerabilities"`
	}
	if json.Unmarshal(b, &doc) != nil || doc.DateReleased == "" || doc.Count == nil || *doc.Count != len(doc.Vulnerabilities) {
		return []Snapshot{s.failure("kev", origin, fmt.Errorf("KEV catalogue omitted its date, count or complete entries"))}
	}
	table := store.Table{Name: querytpl.DefaultKEVView, Columns: []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}, Rows: [][]string{}}
	for _, v := range doc.Vulnerabilities {
		if !cveID.MatchString(v.ID) {
			return []Snapshot{s.failure("kev", origin, fmt.Errorf("KEV entry omitted its CVE identity"))}
		}
		table.Rows = append(table.Rows, []string{v.ID, v.Vendor, v.Product, v.Added, v.Ransomware})
	}
	status := s.status("kev", origin, d, r)
	status.UpstreamAt = doc.DateReleased
	status.Rows = len(table.Rows)
	if s.Client.Changed == changed {
		status.Status = "unchanged"
	}
	return []Snapshot{{Table: table, Refresh: status}}
}

type epssConnector struct{}

func (epssConnector) Name() string { return "epss" }
func (epssConnector) Fetch(ctx context.Context, s *Session) []Snapshot {
	origin := s.Config.EPSS
	d, r := s.Client.Downloaded, s.Client.Reused
	changed := s.Client.Changed
	ids := make([]string, 0, len(s.CVEs))
	for id := range s.CVEs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	table := store.Table{Name: querytpl.DefaultEPSSView, Columns: []string{"cve_id", "epss", "percentile", "score_date"}, Rows: [][]string{}}
	latest := ""
	for len(ids) > 0 {
		count := min(50, len(ids))
		batch := ids[:count]
		ids = ids[count:]
		b, err := s.Client.JSON(ctx, "GET", origin+"/epss?cve="+url.QueryEscape(strings.Join(batch, ",")), nil)
		if err != nil {
			return []Snapshot{s.failure("epss", origin, err)}
		}
		var doc struct {
			Status string                                         `json:"status"`
			Total  *int                                           `json:"total"`
			Data   []struct{ CVE, EPSS, Percentile, Date string } `json:"data"`
		}
		if json.Unmarshal(b, &doc) != nil || doc.Status != "OK" || doc.Total == nil || *doc.Total != len(doc.Data) {
			return []Snapshot{s.failure("epss", origin, fmt.Errorf("EPSS returned an incomplete score page"))}
		}
		allowed := map[string]bool{}
		for _, id := range batch {
			allowed[id] = true
		}
		for _, v := range doc.Data {
			score, scoreErr := strconv.ParseFloat(v.EPSS, 64)
			percentile, percentileErr := strconv.ParseFloat(v.Percentile, 64)
			_, dateErr := time.Parse("2006-01-02", v.Date)
			if !allowed[v.CVE] || dateErr != nil || scoreErr != nil || percentileErr != nil || math.IsNaN(score) || math.IsNaN(percentile) || score < 0 || score > 1 || percentile < 0 || percentile > 1 {
				return []Snapshot{s.failure("epss", origin, fmt.Errorf("EPSS returned an unrequested CVE or undated score"))}
			}
			if v.Date > latest {
				latest = v.Date
			}
			table.Rows = append(table.Rows, []string{v.CVE, v.EPSS, v.Percentile, v.Date})
		}
	}
	status := s.status("epss", origin, d, r)
	status.UpstreamAt = latest
	status.Rows = len(table.Rows)
	if len(s.CVEs) == 0 {
		status.Status = "not_needed"
	} else if s.Client.Changed == changed {
		status.Status = "unchanged"
	}
	return []Snapshot{{Table: table, Refresh: status}}
}
