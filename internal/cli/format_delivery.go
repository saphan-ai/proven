package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/review"
)

func deliveryStatus(status string) string {
	switch status {
	case "code_indication":
		return "Indicated by code"
	case "contradiction":
		return "Contradicted"
	default:
		return "Unclear"
	}
}
func deliveryRefs(c delivery.Claim) string {
	var refs []string
	if c.SourceSHA != "" {
		refs = append(refs, "source commit "+c.SourceSHA[:min(12, len(c.SourceSHA))])
	}
	for _, cite := range c.Citations {
		refs = append(refs, fmt.Sprintf("%s:%d", cite.Path, cite.StartLine))
	}
	for _, id := range c.FindingIDs {
		refs = append(refs, "proved counterexample "+id)
	}
	return strings.Join(refs, ", ")
}

// Old signed packs retain their flat claims. Group them only by original source,
// never by similar requirement text or status, and keep every claim intact.
func deliveryAssessments(d *delivery.Result) []delivery.Assessment {
	if len(d.Assessments) != 0 {
		return d.Assessments
	}
	var out []delivery.Assessment
	indices := map[string]int{}
	for _, c := range d.Claims {
		key := fmt.Sprintf("%s:%d", c.SourceKind, c.SourceIndex)
		i, ok := indices[key]
		if !ok {
			i = len(out)
			indices[key] = i
			out = append(out, delivery.Assessment{Source: delivery.Source{ID: key, Kind: c.SourceKind, Index: c.SourceIndex, SHA: c.SourceSHA}, Complete: d.Complete, SemanticStatus: "assessed"})
		}
		out[i].Claims = append(out[i].Claims, c)
	}
	return out
}

func mdSource(s delivery.Source) string {
	label := mdText(mdShort(s.Label(), mdMaxTitle))
	u, err := url.Parse(s.URL)
	if err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && len(s.URL) <= 2048 {
		target := strings.NewReplacer("<", "%3C", ">", "%3E").Replace(u.String())
		label = "[" + label + "](<" + target + ">)"
	}
	if s.Author != "" {
		label += " by " + mdText(mdShort(s.Author, 100))
	}
	return label
}

func mdDelivery(b *strings.Builder, result review.Result) {
	d := result.Delivery
	if d == nil {
		return
	}
	b.WriteString("### Declared scope — each source separately\n\n")
	if d.Mode == "deterministic" {
		b.WriteString("Description checks ran without a model. They check text completeness, not meaning or implementation.\n\n")
	} else {
		b.WriteString("Model assessment with checked citations; separate from proved findings. No acceptance tests are executed by this check.\n\n")
	}
	assessments := deliveryAssessments(d)
	used := 0
	for i, a := range assessments {
		var section strings.Builder
		fmt.Fprintf(&section, "#### %s\n\n", mdSource(a.Source))
		if a.Source.Text != "" {
			fmt.Fprintf(&section, "> %s\n\n", mdText(mdShort(a.Source.Text, 300)))
		}
		if a.Description.Status != "" {
			reason := ""
			if a.Description.Status != "present" {
				reason = " " + mdText(a.Description.Reason)
			}
			fmt.Fprintf(&section, "**Description: %s.**%s\n\n", mdText(a.Description.Status), reason)
		}
		if len(a.Claims) != 0 {
			section.WriteString("| ID | Stated requirement | Assessment | Reason and references |\n| --- | --- | --- | --- |\n")
			for j, c := range a.Claims {
				if j == 6 {
					fmt.Fprintf(&section, "\n%d more findings for this source are in JSON and HTML.\n", len(a.Claims)-j)
					break
				}
				fmt.Fprintf(&section, "| %s | %s | %s | %s |\n", mdCell(c.ID), mdCell(mdShort(c.Requirement, mdMaxTitle)), deliveryStatus(c.Status), mdCell(mdShort(c.Reason+" "+deliveryRefs(c), mdMaxNote)))
			}
			section.WriteString("\n")
		} else if a.SemanticStatus == "assessed" {
			section.WriteString("No assessable delivery claims were extracted from this source.\n\n")
		} else if a.SemanticStatus == "not_assessable" {
			section.WriteString("No semantic assessment: the description is missing, too short or a placeholder.\n\n")
		}
		if a.Issue != "" {
			section.WriteString("**Source incomplete:** " + mdText(mdShort(a.Issue, mdMaxNote)) + "\n\n")
		}
		if used+section.Len() > 12000 {
			fmt.Fprintf(b, "%d more source assessments are in the JSON and HTML reports; the PR comment has a size limit. Sources remain separate in those reports.\n\n", len(assessments)-i)
			break
		}
		b.WriteString(section.String())
		used += section.Len()
	}
	if len(assessments) == 0 {
		b.WriteString("No assessable requirements were extracted; this does not certify delivery.\n\n")
	}
	if d.ContextTruncated {
		b.WriteString("Source code context was bounded; claims outside it may remain unclear.\n\n")
	}
	if d.Issue != "" {
		b.WriteString("Scope check incomplete: " + mdText(mdShort(d.Issue, mdMaxNote)) + "\n\n")
	}
}

func htmlDelivery(b *strings.Builder, result review.Result) {
	d := result.Delivery
	if d == nil {
		return
	}
	b.WriteString("<section><h2>Declared scope — each source separately</h2>\n")
	if d.Mode == "deterministic" {
		b.WriteString("<p>Description checks ran without a model. Text completeness does not establish meaning or delivery.</p>\n")
	} else {
		b.WriteString("<p>Model assessment with checked citations; separate from proved findings. This check does not execute acceptance tests.</p>\n")
	}
	for i, a := range deliveryAssessments(d) {
		fmt.Fprintf(b, "<section id=\"scope-%d\"><h3>%s</h3>\n", i+1, htmlText(a.Source.Label()))
		if a.Source.Author != "" {
			b.WriteString("<p>Author: " + htmlText(a.Source.Author) + "</p>\n")
		}
		if a.Source.URL != "" {
			b.WriteString("<p>Source: " + htmlText(a.Source.URL) + "</p>\n")
		}
		if a.Source.Text != "" {
			b.WriteString("<blockquote>" + htmlText(a.Source.Text) + "</blockquote>\n")
		}
		if a.Description.Status != "" {
			reason := ""
			if a.Description.Status != "present" {
				reason = " " + htmlText(a.Description.Reason)
			}
			b.WriteString("<p><strong>Description: " + htmlText(a.Description.Status) + ".</strong>" + reason + "</p>\n")
		}
		if len(a.Claims) != 0 {
			b.WriteString("<table><thead><tr><th>ID</th><th>Stated requirement</th><th>Assessment</th><th>Reason and references</th></tr></thead><tbody>\n")
			for _, c := range a.Claims {
				fmt.Fprintf(b, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n", htmlText(c.ID), htmlText(c.Requirement), deliveryStatus(c.Status), htmlText(c.Reason+" "+deliveryRefs(c)))
			}
			b.WriteString("</tbody></table>\n")
			for _, c := range a.Claims {
				b.WriteString("<details><summary>" + htmlText(c.ID) + " — checked quotations</summary>\n<blockquote>" + htmlText(c.SourceQuote) + "</blockquote>\n")
				for _, cite := range c.Citations {
					fmt.Fprintf(b, "<p>%s:%d–%d</p><pre>%s</pre>\n", htmlText(cite.Path), cite.StartLine, cite.EndLine, htmlPre(cite.Quote))
				}
				b.WriteString("</details>\n")
			}
		} else if a.SemanticStatus == "assessed" {
			b.WriteString("<p>No assessable delivery claims were extracted from this source.</p>\n")
		} else if a.SemanticStatus == "not_assessable" {
			b.WriteString("<p>No semantic assessment: the description is missing, too short or a placeholder.</p>\n")
		}
		if a.Issue != "" {
			b.WriteString("<p><strong>Source incomplete:</strong> " + htmlText(a.Issue) + "</p>\n")
		}
		b.WriteString("</section>\n")
	}
	if len(d.Claims) == 0 && len(d.Assessments) == 0 {
		b.WriteString("<p>No assessable requirements were extracted; this does not certify delivery.</p>\n")
	}
	if d.ContextTruncated {
		b.WriteString("<p>Source code context was bounded; claims outside it may remain unclear.</p>\n")
	}
	if d.Issue != "" {
		b.WriteString("<p>Scope check incomplete: " + htmlText(d.Issue) + "</p>\n")
	}
	b.WriteString("</section>\n")
}
