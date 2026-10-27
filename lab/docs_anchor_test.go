package lab

// The docs site (docs-site/scripts/guards.mjs) fails its build when a link points at a heading that is not on the page: in
// round 1 docs/LAB.md linked #the-dashboard while the heading is "The dashboards", so the docs image did not build and the
// lab's "Open the docs" was dead. The guard needs Node and the site's content step; this holds the same rule, for the lab's
// pages, with the same anchor rule as docs-site/scripts/lib.mjs (anchorOf, headingAnchors).

import (
	"regexp"
	"strings"
	"testing"
	"unicode"
)

var (
	headingRe    = regexp.MustCompile("^#{1,6}\\s+(.*?)\\s*#*\\s*$")
	mdLinkInHead = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	inPageLinkRe = regexp.MustCompile("\\]\\(#([^)\\s]+)\\)")
)

// anchorOf is anchorOf of docs-site/scripts/lib.mjs: backticks go, a link keeps its text, lower case, letters, digits,
// spaces, underscore and hyphen stay, each space becomes a hyphen.
func anchorOf(h string) string {
	h = strings.ReplaceAll(h, "`", "")
	h = mdLinkInHead.ReplaceAllString(h, "$1")
	h = strings.TrimSpace(strings.ToLower(h))
	var b strings.Builder
	for _, r := range h {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '-':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteRune('-')
		}
	}
	return b.String()
}

// brokenInPageAnchors returns every in-page link of the Markdown whose heading is not on the page.
func brokenInPageAnchors(md string) []string {
	anchors := map[string]bool{}
	seen := map[string]int{}
	fence := false
	for _, l := range strings.Split(md, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		if m := headingRe.FindStringSubmatch(l); m != nil {
			a := anchorOf(m[1])
			k := seen[a]
			seen[a]++
			if k > 0 {
				a += "-" + string(rune('0'+k))
			}
			anchors[a] = true
		}
	}
	var bad []string
	fence = false
	for _, l := range strings.Split(md, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		for _, m := range inPageLinkRe.FindAllStringSubmatch(l, -1) {
			if !anchors[m[1]] {
				bad = append(bad, m[1])
			}
		}
	}
	return bad
}

func TestTheLabDocsLinkOnlyToHeadingsThatExist(t *testing.T) {
	for _, doc := range []string{"LAB.md", "LAB_GO_LIVE_TEST.md", "CONFIGURATION.md"} {
		md := read(t, "docs", doc)
		if bad := brokenInPageAnchors(md); len(bad) > 0 {
			t.Errorf("docs/%s links to headings that are not on the page: %v (the docs site's link guard fails the build)", doc, bad)
		}
	}
	// The negative control. Mutated field: the anchor of the round-1 break, #the-dashboards put back to #the-dashboard.
	lab := read(t, "docs", "LAB.md")
	mut := strings.Replace(lab, "](#the-dashboards)", "](#the-dashboard)", 1)
	if mut == lab {
		t.Fatal("the mutation changed nothing: docs/LAB.md no longer links to #the-dashboards")
	}
	if len(brokenInPageAnchors(mut)) == 0 {
		t.Error("the control did not turn red: the old anchor passed")
	}
}
