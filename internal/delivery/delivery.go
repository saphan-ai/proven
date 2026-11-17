// Package delivery holds an advisory assessment of stated PR scope. It has no
// I/O or model dependency and is separate from proved findings.
package delivery

import (
	"fmt"
	"strings"
)

type Commit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

// Comment is one GitHub discussion, review, inline or commit comment. Its body
// stays separate from every other declaration, including replies in a thread.
type Comment struct {
	Kind      string `json:"kind"`
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	URL       string `json:"url,omitempty"`
	Author    string `json:"author,omitempty"`
	SHA       string `json:"sha,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
type Metadata struct {
	Context     string    `json:"context,omitempty"`
	URL         string    `json:"url,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Commits     []Commit  `json:"commits"`
	Comments    []Comment `json:"comments,omitempty"`
	Issues      []string  `json:"issues,omitempty"`
}

// Source is assigned by Proven, not the model. Index addresses the original
// commits/comments array; ID lets a flat claim retain its separate source.
type Source struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Index     int    `json:"index"`
	CommentID int64  `json:"comment_id,omitempty"`
	Text      string `json:"text"`
	URL       string `json:"url,omitempty"`
	Author    string `json:"author,omitempty"`
	SHA       string `json:"sha,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
type Assessment struct {
	Source         Source           `json:"source"`
	Description    DescriptionCheck `json:"description"`
	SemanticStatus string           `json:"semantic_status"`
	Complete       bool             `json:"complete"`
	Claims         []Claim          `json:"claims"`
	Issue          string           `json:"issue,omitempty"`
}

func CommentKind(kind string) bool {
	switch kind {
	case "pr_comment", "review", "review_comment", "commit_comment":
		return true
	}
	return false
}

func Sources(m Metadata) []Source {
	var out []Source
	add := func(s Source) {
		if strings.TrimSpace(s.Text) != "" {
			out = append(out, s)
		}
	}
	add(Source{ID: "title", Kind: "title", Text: m.Title, URL: m.URL})
	// An absent PR description is itself observable, so keep that source.
	if m.Context != "git" {
		out = append(out, Source{ID: "description", Kind: "description", Text: m.Description, URL: m.URL})
	}
	for i, c := range m.Commits {
		out = append(out, Source{ID: fmt.Sprintf("commit:%d:%s", i, c.SHA), Kind: "commit", Index: i, Text: c.Message, SHA: c.SHA, URL: c.URL})
	}
	for i, c := range m.Comments {
		add(Source{ID: fmt.Sprintf("%s:%d:%d", c.Kind, i, c.ID), Kind: c.Kind, Index: i, CommentID: c.ID, Text: c.Body, URL: c.URL, Author: c.Author, SHA: c.SHA, UpdatedAt: c.UpdatedAt})
	}
	return out
}

func (s Source) Label() string {
	switch s.Kind {
	case "title":
		return "PR title"
	case "description":
		return "PR description"
	case "commit":
		if s.SHA != "" {
			return "Commit " + s.SHA[:min(12, len(s.SHA))]
		}
		return fmt.Sprintf("Commit message %d", s.Index+1)
	case "pr_comment":
		return fmt.Sprintf("PR comment #%d", s.CommentID)
	case "review":
		return fmt.Sprintf("Review #%d", s.CommentID)
	case "review_comment":
		return fmt.Sprintf("Inline comment #%d", s.CommentID)
	case "commit_comment":
		return fmt.Sprintf("Commit comment #%d", s.CommentID)
	}
	return s.ID
}

type Citation struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Quote     string `json:"quote"`
}
type Claim struct {
	ID          string     `json:"id,omitempty"`
	SourceID    string     `json:"source_id,omitempty"`
	Requirement string     `json:"requirement"`
	SourceKind  string     `json:"source_kind"`
	SourceIndex int        `json:"source_index"`
	SourceSHA   string     `json:"source_sha,omitempty"`
	SourceQuote string     `json:"source_quote"`
	Status      string     `json:"status"`
	Reason      string     `json:"reason"`
	Citations   []Citation `json:"citations"`
	FindingIDs  []string   `json:"finding_ids,omitempty"`
}
type Result struct {
	Mode             string       `json:"mode,omitempty"`
	Complete         bool         `json:"complete"`
	Snapshot         string       `json:"snapshot"`
	MetadataSHA256   string       `json:"metadata_sha256"`
	URL              string       `json:"url,omitempty"`
	ContextTruncated bool         `json:"context_truncated"`
	Claims           []Claim      `json:"claims"`
	Assessments      []Assessment `json:"assessments,omitempty"`
	Issue            string       `json:"issue,omitempty"`
}

// Counterexample was accepted by ordinary evidence gates. It supports a
// contradiction; it does not prove the positive delivery of another feature.
type Counterexample struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title"`
}

func (r *Result) Redacted(text func(string) string) *Result {
	if r == nil {
		return nil
	}
	out := *r
	out.URL, out.Snapshot, out.Issue = text(r.URL), text(r.Snapshot), text(r.Issue)
	out.Claims = make([]Claim, len(r.Claims))
	redactClaim := func(c Claim) Claim {
		c.ID, c.SourceID, c.SourceSHA = text(c.ID), text(c.SourceID), text(c.SourceSHA)
		c.Requirement, c.SourceQuote, c.Reason = text(c.Requirement), text(c.SourceQuote), text(c.Reason)
		c.Citations = append([]Citation(nil), c.Citations...)
		for j := range c.Citations {
			c.Citations[j].Path = text(c.Citations[j].Path)
			c.Citations[j].Quote = text(c.Citations[j].Quote)
		}
		c.FindingIDs = append([]string(nil), c.FindingIDs...)
		return c
	}
	for i, c := range r.Claims {
		out.Claims[i] = redactClaim(c)
	}
	out.Assessments = make([]Assessment, len(r.Assessments))
	for i, a := range r.Assessments {
		a.Source.ID, a.Source.Text, a.Source.URL = text(a.Source.ID), text(a.Source.Text), text(a.Source.URL)
		a.Source.Author, a.Source.SHA, a.Source.UpdatedAt = text(a.Source.Author), text(a.Source.SHA), text(a.Source.UpdatedAt)
		a.Issue = text(a.Issue)
		a.Claims = make([]Claim, len(r.Assessments[i].Claims))
		for j, c := range r.Assessments[i].Claims {
			a.Claims[j] = redactClaim(c)
		}
		out.Assessments[i] = a
	}
	return &out
}
