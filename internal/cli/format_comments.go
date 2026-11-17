package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

func commentLabel(s string) string {
	switch s {
	case "missing_api_doc":
		return "Missing API documentation"
	case "stale_param_tag":
		return "Obsolete parameter name"
	case "placeholder_comment":
		return "Placeholder comment"
	case "doc_name":
		return "Documentation naming convention"
	case "needs_attention":
		return "Comment needs attention"
	case "consistent":
		return "Consistent with supplied code"
	case "unclear":
		return "Unclear from supplied code"
	case "not_assessed":
		return "Not assessed"
	}
	return s
}

func textComments(w io.Writer, r review.Result) error {
	c := r.CodeComments
	if c == nil {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Code comments: %s, %d changed units, %d structural issues\n", visible(c.Mode), len(c.Units), len(c.Issues))
	for _, i := range c.Issues {
		fmt.Fprintf(&b, "  %s %s: %s\n", visible(i.UnitID), visible(i.Rule), visible(i.Message))
	}
	for _, a := range c.Assessments {
		fmt.Fprintf(&b, "  %s %s (model assessment): %s\n", visible(a.UnitID), visible(a.Status), visible(a.Reason))
	}
	if c.Issue != "" {
		fmt.Fprintln(&b, visible(c.Issue))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func mdComments(b *strings.Builder, r review.Result) {
	c := r.CodeComments
	if c == nil {
		return
	}
	fmt.Fprintf(b, "### Code comments\n\n%d changed Go declarations/comments; %d structural issues. Semantic results are model assessments, separate from proved code defects.\n\n", len(c.Units), len(c.Issues))
	b.WriteString("| Source | Check | Result |\n| --- | --- | --- |\n")
	n := 0
	for _, i := range c.Issues {
		if n >= 20 {
			break
		}
		fmt.Fprintf(b, "| %s | %s | %s |\n", mdCell(mdShort(i.UnitID, mdMaxTitle)), mdCell(commentLabel(i.Rule)), mdCell(mdShort(i.Message, mdMaxNote)))
		n++
	}
	for _, a := range c.Assessments {
		if n >= 20 {
			break
		}
		fmt.Fprintf(b, "| %s | %s (model) | %s |\n", mdCell(mdShort(a.UnitID, mdMaxTitle)), mdCell(commentLabel(a.Status)), mdCell(mdShort(a.Reason, mdMaxNote)))
		n++
	}
	if len(c.Issues)+len(c.Assessments) > n {
		fmt.Fprintf(b, "\n%d more comment results are in JSON and HTML.\n", len(c.Issues)+len(c.Assessments)-n)
	}
	if len(c.Unsupported) > 0 {
		fmt.Fprintf(b, "\n%d selected source paths have no comment-language adapter; they were not checked.\n", len(c.Unsupported))
	}
	if c.Issue != "" {
		b.WriteString("\n" + mdText(mdShort(c.Issue, mdMaxNote)) + "\n")
	}
	b.WriteString("\n")
}
func htmlComments(b *strings.Builder, r review.Result) {
	c := r.CodeComments
	if c == nil {
		return
	}
	fmt.Fprintf(b, "<section><h2>Code comments</h2><p>%d changed Go declarations/comments; %d structural issues. Semantic results are model assessments, separate from proved code defects.</p>\n", len(c.Units), len(c.Issues))
	for _, u := range c.Units {
		fmt.Fprintf(b, "<details><summary>%s:%d %s</summary>\n", htmlText(u.Path), u.Line, htmlText(u.Symbol))
		if u.Text != "" {
			b.WriteString("<pre>" + htmlPre(u.Text) + "</pre>\n")
		} else {
			b.WriteString("<p>No documentation comment.</p>\n")
		}
		for _, i := range c.Issues {
			if i.UnitID == u.ID {
				b.WriteString("<p>" + htmlText(commentLabel(i.Rule)+": "+i.Message) + "</p>\n")
			}
		}
		for _, a := range c.Assessments {
			if a.UnitID == u.ID {
				b.WriteString("<p>Model assessment: " + htmlText(commentLabel(a.Status)+" — "+a.Reason) + "</p>\n")
				if a.Quote != "" {
					b.WriteString("<pre>" + htmlPre(a.Quote) + "</pre>\n")
				}
			}
		}
		b.WriteString("</details>\n")
	}
	for _, p := range c.Unsupported {
		b.WriteString("<p>Not checked — no comment adapter: " + htmlText(p) + "</p>\n")
	}
	if c.Issue != "" {
		b.WriteString("<p>" + htmlText(c.Issue) + "</p>\n")
	}
	b.WriteString("</section>\n")
}
