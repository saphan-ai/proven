package feeds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

type depsConnector struct{}

func (depsConnector) Name() string { return "deps.dev" }
func (depsConnector) Fetch(ctx context.Context, s *Session) []Snapshot {
	origin := s.Config.Deps
	d, r := s.Client.Downloaded, s.Client.Reused
	changed := s.Client.Changed
	names := []string{"spdx", "scorecard", "cvss"}
	fail := func(err error) []Snapshot {
		var out []Snapshot
		for _, id := range names {
			out = append(out, s.failure(id, origin, err))
		}
		return out
	}
	tables := []store.Table{
		{Name: querytpl.DefaultSPDXTable, Columns: []string{"ecosystem", "package", "licence"}, Rows: [][]string{}},
		{Name: querytpl.DefaultScorecardTable, Columns: []string{"ecosystem", "package", "maintained", "signed", "release_age_days"}, Rows: [][]string{}},
		{Name: querytpl.DefaultCVSSTable, Columns: []string{"cve_id", "cvss_score", "severity", "source"}, Rows: [][]string{}},
	}
	missingLicense, missingScore := 0, 0
	oldestScore := ""
	projects := map[string][]byte{}
	versions := map[string]int{}
	for _, p := range s.Packages {
		versions[p.Ecosystem+"\x00"+p.Name]++
	}
	for _, p := range s.Packages {
		// Existing package-level licence/health tables cannot express two
		// conflicting selected versions. Omit that ambiguous enrichment.
		if versions[p.Ecosystem+"\x00"+p.Name] > 1 {
			missingLicense++
			missingScore++
			continue
		}
		system := map[string]string{"Go": "go", "npm": "npm", "PyPI": "pypi", "Maven": "maven", "crates.io": "cargo"}[p.Ecosystem]
		if system == "" {
			missingLicense++
			missingScore++
			continue
		}
		address := origin + "/systems/" + system + "/packages/" + url.PathEscape(p.Name) + "/versions/" + url.PathEscape(p.Version)
		b, err := s.Client.JSON(ctx, "GET", address, nil)
		if err != nil {
			var status *HTTPStatusError
			if errors.As(err, &status) && status.Status == 404 {
				missingLicense++
				missingScore++
				continue
			}
			return fail(err)
		}
		var version struct {
			VersionKey      struct{ Name, Version, System string }
			PublishedAt     string
			Licenses        []string
			RelatedProjects []struct {
				ProjectKey                       struct{ ID string }
				RelationType, RelationProvenance string
			}
		}
		if json.Unmarshal(b, &version) != nil || version.VersionKey.Name == "" || version.VersionKey.Version == "" {
			return fail(fmt.Errorf("deps.dev version response omitted its identity"))
		}
		sameVersion := version.VersionKey.Version == p.Version
		for _, reader := range manifest.Readers() {
			if reader.Ecosystem == p.Ecosystem && reader.Compare != nil {
				order, ok := reader.Compare(version.VersionKey.Version, p.Version)
				sameVersion = sameVersion || (ok && order == 0)
			}
		}
		if !strings.EqualFold(version.VersionKey.System, system) || !strings.EqualFold(version.VersionKey.Name, p.Name) || !sameVersion {
			return fail(fmt.Errorf("deps.dev returned another package/version identity"))
		}
		if len(version.Licenses) > 0 {
			tables[0].Rows = append(tables[0].Rows, []string{p.Ecosystem, p.Name, strings.Join(version.Licenses, "; ")})
		} else {
			missingLicense++
		}
		age := ""
		if published, err := time.Parse(time.RFC3339, version.PublishedAt); err == nil && !published.After(s.Now) {
			age = strconv.Itoa(int(s.Now.Sub(published).Hours() / 24))
		}
		maintained, signed := "", ""
		measured := false
		for _, project := range version.RelatedProjects {
			if project.RelationType != "SOURCE_REPO" {
				continue
			}
			id := project.ProjectKey.ID
			if !strings.HasPrefix(id, "github.com/") && !strings.HasPrefix(id, "gitlab.com/") && !strings.HasPrefix(id, "bitbucket.org/") {
				continue
			}
			b, ok := projects[id]
			if !ok {
				b, err = s.Client.JSON(ctx, "GET", origin+"/projects/"+url.PathEscape(id), nil)
				if err != nil {
					var status *HTTPStatusError
					if errors.As(err, &status) && status.Status == 404 {
						continue
					}
					return fail(err)
				}
				projects[id] = b
			}
			var result struct {
				ProjectKey struct{ ID string }
				Scorecard  *struct {
					Date   string
					Checks []struct {
						Name  string
						Score int
					}
				}
			}
			if json.Unmarshal(b, &result) != nil || result.ProjectKey.ID == "" {
				return fail(fmt.Errorf("deps.dev project response omitted its identity"))
			}
			if !strings.EqualFold(result.ProjectKey.ID, id) {
				return fail(fmt.Errorf("deps.dev returned another source-project identity"))
			}
			if result.Scorecard == nil {
				continue
			}
			for _, check := range result.Scorecard.Checks {
				if check.Score < 0 || check.Score > 10 {
					continue
				}
				switch check.Name {
				case "Maintained":
					maintained = strconv.Itoa(check.Score)
					measured = true
				case "Signed-Releases":
					if check.Score == 0 {
						signed = "false"
						measured = true
					} else if check.Score == 10 {
						signed = "true"
						measured = true
					}
				}
			}
			if date := result.Scorecard.Date; date != "" && (oldestScore == "" || date < oldestScore) {
				oldestScore = date
			}
			// One source repository's measurements are not silently combined
			// with another repository's release-signing or maintenance result.
			if measured {
				break
			}
		}
		if !measured {
			missingScore++
		}
		if measured || age != "" {
			tables[1].Rows = append(tables[1].Rows, []string{p.Ecosystem, p.Name, maintained, signed, age})
		}
	}
	ids := make([]string, 0, len(s.Advisories))
	for id := range s.Advisories {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	missingCVSS := 0
	for _, id := range ids {
		var record struct{ Aliases []string }
		json.Unmarshal(s.Advisories[id], &record)
		var cves []string
		for _, alias := range append(record.Aliases, id) {
			if cveID.MatchString(alias) {
				cves = append(cves, alias)
			}
		}
		if len(cves) == 0 {
			continue
		}
		address := origin + "/advisories/" + url.PathEscape(id)
		b, err := s.Client.JSON(ctx, "GET", address, nil)
		if err != nil {
			var status *HTTPStatusError
			if errors.As(err, &status) && status.Status == 404 {
				missingCVSS++
				continue
			}
			return fail(err)
		}
		var advisory struct {
			AdvisoryKey struct{ ID string }
			CVSS3Score  *float64
			CVSS3Vector string
			Aliases     []string
		}
		if json.Unmarshal(b, &advisory) != nil || advisory.AdvisoryKey.ID == "" {
			return fail(fmt.Errorf("deps.dev advisory response omitted its identity"))
		}
		if advisory.CVSS3Score == nil || !strings.HasPrefix(advisory.CVSS3Vector, "CVSS:3.") {
			missingCVSS++
			continue
		}
		score := *advisory.CVSS3Score
		if score < 0 || score > 10 {
			return fail(fmt.Errorf("deps.dev returned an invalid CVSS score"))
		}
		severity := "NONE"
		switch {
		case score >= 9:
			severity = "CRITICAL"
		case score >= 7:
			severity = "HIGH"
		case score >= 4:
			severity = "MEDIUM"
		case score > 0:
			severity = "LOW"
		}
		for _, cve := range cves {
			linked := advisory.AdvisoryKey.ID == cve
			for _, alias := range advisory.Aliases {
				linked = linked || alias == cve
			}
			if !linked {
				missingCVSS++
				continue
			}
			tables[2].Rows = append(tables[2].Rows, []string{cve, strconv.FormatFloat(score, 'f', 1, 64), severity, "deps.dev " + address})
		}
	}
	missing := []int{missingLicense, missingScore, missingCVSS}
	var out []Snapshot
	for i, name := range names {
		status := s.status(name, origin, d, r)
		status.Rows = len(tables[i].Rows)
		if name == "scorecard" {
			status.UpstreamAt = oldestScore
		}
		if missing[i] > 0 {
			status.Status = "partial"
			status.Issue = fmt.Sprintf("%d selected packages/advisories have no supported measurement; absence is not a healthy or permissive result", missing[i])
		} else if len(s.Packages) == 0 {
			status.Status = "not_needed"
		} else if s.Client.Changed == changed {
			status.Status = "unchanged"
		}
		out = append(out, Snapshot{Table: tables[i], Refresh: status})
	}
	return out
}
