// Package delivery holds an advisory assessment of stated PR scope. It has no
// I/O or model dependency and is separate from proved findings.
package delivery

type Commit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
}
type Metadata struct {
	URL         string   `json:"url,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Commits     []Commit `json:"commits"`
}
type Citation struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Quote     string `json:"quote"`
}
type Claim struct {
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
	Complete         bool    `json:"complete"`
	Snapshot         string  `json:"snapshot"`
	MetadataSHA256   string  `json:"metadata_sha256"`
	URL              string  `json:"url,omitempty"`
	ContextTruncated bool    `json:"context_truncated"`
	Claims           []Claim `json:"claims"`
	Issue            string  `json:"issue,omitempty"`
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
	for i, c := range r.Claims {
		c.Requirement, c.SourceQuote, c.Reason = text(c.Requirement), text(c.SourceQuote), text(c.Reason)
		c.Citations = append([]Citation(nil), c.Citations...)
		for j := range c.Citations {
			c.Citations[j].Path = text(c.Citations[j].Path)
			c.Citations[j].Quote = text(c.Citations[j].Quote)
		}
		c.FindingIDs = append([]string(nil), c.FindingIDs...)
		out.Claims[i] = c
	}
	return &out
}
