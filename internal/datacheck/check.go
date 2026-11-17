// Package datacheck applies registered source rules to changed dependencies.
// The existing source and query adapters decide semantics; no model is called.
package datacheck

import (
	"context"
	"fmt"
	"regexp"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/source"
)

var cves = regexp.MustCompile(`CVE-[0-9]{4}-[0-9]{4,}`)

func Check(ctx context.Context, files []change.File, matches []osv.Match, q query.Runner, rel querytpl.Relations) []source.Observation {
	hits, _, err := manifest.Collect(files, manifest.Readers())
	if err != nil {
		return []source.Observation{{Issue: "dependency manifests could not be read"}}
	}
	var out []source.Observation
	for _, hit := range hits {
		var ids, advisories []string
		seen := map[string]bool{}
		for _, m := range matches {
			if m.Ecosystem == hit.Ecosystem && m.Package == hit.Package && m.Version == hit.Version {
				for _, a := range m.Advisories {
					if a.Affected {
						advisories = append(advisories, a.Row.ID)
						for _, id := range cves.FindAllString(a.Row.ID+" "+a.Row.Aliases, -1) {
							if !seen[id] {
								seen[id] = true
								ids = append(ids, id)
							}
						}
					}
				}
			}
		}
		for _, adapter := range source.All() {
			// These three already run in OSV comparison and finding enrichment.
			if adapter.ID() == "osv" || adapter.ID() == "kev" || adapter.ID() == "epss" {
				continue
			}
			for _, spec := range adapter.Templates() {
				params := map[string]string{}
				identities := []string{""}
				binding := false
				known := true
				identityParam := ""
				for _, p := range spec.Template.Params {
					switch p.Name {
					case querytpl.ParamPackage:
						params[p.Name] = hit.Package
						binding = true
					case querytpl.ParamEcosystem:
						params[p.Name] = hit.Ecosystem
					case "version":
						params[p.Name] = hit.Version
					case querytpl.ParamCVEID:
						identities = ids
						identityParam = p.Name
						binding = true
					case querytpl.ParamAdvisoryID:
						identities = advisories
						identityParam = p.Name
						binding = true
					default:
						known = false
					}
				}
				if !known || !binding {
					continue
				}
				for _, id := range identities {
					o := source.Observation{Adapter: adapter.ID(), Template: spec.Template.Name, Ecosystem: hit.Ecosystem, Package: hit.Package, Version: hit.Version, Path: hit.File, Line: hit.Line}
					if len(out) >= 1000 {
						out = append(out, source.Observation{Issue: "deterministic data checks reached their 1000-query limit; remaining work was not checked"})
						return out
					}
					if ctx.Err() != nil {
						o.Issue = "deterministic data check did not run before the deadline"
						out = append(out, o)
						continue
					}
					if identityParam != "" {
						params[identityParam] = id
					}
					res, err := q.Run(ctx, querytpl.Request{Template: spec.Template.Name, Params: params}, rel)
					if err != nil {
						o.Issue = fmt.Sprintf("%s did not return usable evidence", adapter.ID())
						out = append(out, o)
						continue
					}
					o.SQL, o.Form, o.Engine, o.Snapshot = res.SQL, res.Form, res.Engine, res.Snapshot
					kept, matched := adapter.Rule(source.Evidence{SQL: res.SQL, Form: res.Form, Engine: res.Engine, Snapshot: res.Snapshot, Columns: res.Columns, Rows: res.Rows})
					o.Complete, o.Matched = true, matched
					if matched {
						o.Columns, o.Rows = kept.Columns, kept.Rows
					}
					out = append(out, o)
				}
			}
		}
	}
	return out
}
