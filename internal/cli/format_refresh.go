package cli

import (
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

func mdRefresh(b *strings.Builder, r review.Result) {
	if len(r.SourceRefresh) == 0 {
		mdDataChecks(b, r)
		return
	}
	b.WriteString("### Source refresh\n\n| Source | Refresh | Checked at (UTC) | Upstream observation | Rows |\n| --- | --- | --- | --- | --- |\n")
	for _, s := range r.SourceRefresh {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %d |\n", mdCell(s.ID), mdCell(s.Status), mdCell(s.CheckedAt), mdCell(s.UpstreamAt), s.Rows)
	}
	for _, s := range r.SourceRefresh {
		if s.Issue != "" {
			fmt.Fprintf(b, "\n%s: %s\n\n", mdText(s.ID), mdText(mdShort(s.Issue, mdMaxNote)))
		}
	}
	b.WriteString("\nChecked now means the provider was contacted or confirmed a cached response; its observation date can be older.\n\n")
	mdDataChecks(b, r)
}
func htmlRefresh(b *strings.Builder, r review.Result) {
	if len(r.SourceRefresh) == 0 {
		htmlDataChecks(b, r)
		return
	}
	b.WriteString("<section><h2>Source refresh</h2><table><thead><tr><th>Source</th><th>Refresh</th><th>Checked at (UTC)</th><th>Upstream observation</th><th>Rows</th></tr></thead><tbody>\n")
	for _, s := range r.SourceRefresh {
		fmt.Fprintf(b, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td></tr>\n", htmlText(s.ID), htmlText(s.Status), htmlText(s.CheckedAt), htmlText(s.UpstreamAt), s.Rows)
	}
	b.WriteString("</tbody></table>\n")
	for _, s := range r.SourceRefresh {
		if s.Issue != "" {
			b.WriteString("<p>" + htmlText(s.ID) + ": " + htmlText(s.Issue) + "</p>\n")
		}
	}
	b.WriteString("<p>Checked now means the provider was contacted or confirmed a cached response; its observation date can be older.</p></section>\n")
	htmlDataChecks(b, r)
}

func mdDataChecks(b *strings.Builder, r review.Result) {
	if len(r.SourceChecks) == 0 {
		return
	}
	b.WriteString("### Deterministic data checks\n\nRegistered source rules, run without a model. A data observation is separate from a proved code defect.\n\n| Source | Package/version | Rule result |\n| --- | --- | --- |\n")
	for i, o := range r.SourceChecks {
		if i == 20 {
			fmt.Fprintf(b, "\n%d more source checks are in JSON and HTML.\n", len(r.SourceChecks)-i)
			break
		}
		status := "no matching rule"
		if o.Matched {
			status = "rule matched"
		}
		if !o.Complete {
			status = "unavailable: " + o.Issue
		}
		fmt.Fprintf(b, "| %s | %s | %s |\n", mdCell(o.Adapter), mdCell(mdShort(o.Package+"@"+o.Version, mdMaxTitle)), mdCell(mdShort(status, mdMaxNote)))
	}
	b.WriteString("\n")
}
func htmlDataChecks(b *strings.Builder, r review.Result) {
	if len(r.SourceChecks) == 0 {
		return
	}
	b.WriteString("<section><h2>Deterministic data checks</h2><p>Registered source rules, run without a model. A data observation is separate from a proved code defect.</p><table><thead><tr><th>Source</th><th>Package/version</th><th>Rule result</th><th>Location</th></tr></thead><tbody>\n")
	for _, o := range r.SourceChecks {
		status := "no matching rule"
		if o.Matched {
			status = "rule matched"
		}
		if !o.Complete {
			status = "unavailable: " + o.Issue
		}
		fmt.Fprintf(b, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s:%d</td></tr>\n", htmlText(o.Adapter), htmlText(o.Package+"@"+o.Version), htmlText(status), htmlText(o.Path), o.Line)
	}
	b.WriteString("</tbody></table>\n")
	for _, o := range r.SourceChecks {
		if o.Matched {
			b.WriteString("<details><summary>" + htmlText(o.Adapter+" "+o.Package) + " — source evidence</summary><pre>" + htmlText(o.SQL) + "</pre>\n")
			for _, row := range o.Rows {
				b.WriteString("<pre>" + htmlText(strings.Join(row, " | ")) + "</pre>\n")
			}
			b.WriteString("<p>Snapshot: " + htmlText(o.Snapshot) + "</p></details>\n")
		}
	}
	b.WriteString("</section>\n")
}
