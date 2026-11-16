package cli

import (
	"fmt"
	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/review"
	"strings"
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
		refs = append(refs, "source commit "+c.SourceSHA)
	}
	for _, cite := range c.Citations {
		refs = append(refs, fmt.Sprintf("%s:%d", cite.Path, cite.StartLine))
	}
	for _, id := range c.FindingIDs {
		refs = append(refs, "proved counterexample "+id)
	}
	return strings.Join(refs, ", ")
}
func mdDelivery(b *strings.Builder, result review.Result) {
	d := result.Delivery
	if d == nil {
		return
	}
	b.WriteString("### PR delivery\n\nModel assessment with checked citations; separate from proved findings. This check does not execute acceptance tests.\n\n")
	b.WriteString("| Stated requirement | Assessment | Reason and references |\n| --- | --- | --- |\n")
	for i, c := range d.Claims {
		if i == mdMaxNotes {
			fmt.Fprintf(b, "\n%d more claims are in the JSON report.\n", len(d.Claims)-i)
			break
		}
		fmt.Fprintf(b, "| %s | %s | %s |\n", mdCell(mdShort(c.Requirement, mdMaxTitle)), deliveryStatus(c.Status), mdCell(mdShort(c.Reason+" "+deliveryRefs(c), mdMaxNote)))
	}
	if len(d.Claims) == 0 {
		b.WriteString("\nNo assessable requirements were extracted; this does not certify delivery.\n")
	}
	if d.ContextTruncated {
		b.WriteString("\nSource context was bounded; claims outside it may remain unclear.\n")
	}
	if d.Issue != "" {
		b.WriteString("\nDelivery check incomplete: " + mdText(d.Issue) + "\n")
	}
	b.WriteString("\n")
}
func htmlDelivery(b *strings.Builder, result review.Result) {
	d := result.Delivery
	if d == nil {
		return
	}
	b.WriteString("<section><h2>PR delivery</h2><p>Model assessment with checked citations; separate from proved findings. This check does not execute acceptance tests.</p>\n<table><thead><tr><th>Stated requirement</th><th>Assessment</th><th>Reason and references</th></tr></thead><tbody>\n")
	for _, c := range d.Claims {
		fmt.Fprintf(b, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n", htmlText(c.Requirement), deliveryStatus(c.Status), htmlText(c.Reason+" "+deliveryRefs(c)))
	}
	b.WriteString("</tbody></table>\n")
	if len(d.Claims) == 0 {
		b.WriteString("<p>No assessable requirements were extracted; this does not certify delivery.</p>\n")
	}
	if d.ContextTruncated {
		b.WriteString("<p>Source context was bounded; claims outside it may remain unclear.</p>\n")
	}
	if d.Issue != "" {
		b.WriteString("<p>Delivery check incomplete: " + htmlText(d.Issue) + "</p>\n")
	}
	b.WriteString("</section>\n")
}
