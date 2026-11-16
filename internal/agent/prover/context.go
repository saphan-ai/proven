package prover

import (
	"path"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
)

const maxContextBytes = 64 << 10

type contextFile struct {
	Path      string `json:"path"`
	Base      string `json:"base"`
	Change    string `json:"change"`
	Truncated bool   `json:"truncated,omitempty"`
}

// promptFor reads no filesystem paths supplied by a model. Its source comes
// only from the snapshot already checked by the review's read policy.
func promptFor(c candidate.Candidate, files []change.File) (string, answer.Spec) {
	if files == nil {
		return Prompt(c.ID, c.Title), testSpec
	}
	remaining := maxContextBytes
	var context []contextFile
	add := func(f change.File) {
		if remaining == 0 || len(context) == 8 {
			return
		}
		clip := func(b []byte, limit int) string {
			if len(b) > limit {
				b = b[:limit]
			}
			s := strings.ToValidUTF8(string(b), "\uFFFD")
			if len(s) > limit {
				s = s[:limit]
			}
			for len(s) > 0 && !utf8.ValidString(s) {
				s = s[:len(s)-1]
			}
			remaining -= len(s)
			return s
		}
		baseLimit := remaining
		if len(f.Change) > 0 {
			baseLimit /= 2
		}
		base := clip(f.Base, baseLimit)
		head := clip(f.Change, remaining)
		context = append(context, contextFile{f.Path, base, head, len(base) != len(f.Base) || len(head) != len(f.Change)})
	}
	for _, f := range files {
		if f.Path == c.Path {
			add(f)
		}
	}
	if strings.HasSuffix(c.Path, "_test.go") {
		for _, f := range files {
			if path.Dir(f.Path) == path.Dir(c.Path) && strings.HasSuffix(f.Path, ".go") && !strings.HasSuffix(f.Path, "_test.go") {
				add(f)
			}
		}
	}
	for _, f := range files {
		if f.Path != c.Path && path.Dir(f.Path) == path.Dir(c.Path) && strings.HasSuffix(f.Path, "_test.go") {
			add(f)
		}
	}
	for _, f := range files {
		if f.Path == "go.mod" && f.Path != c.Path {
			add(f)
		}
	}
	spec := testSpec
	spec.Task += " Use the supplied source and existing tests to locate the package and its API. Prefer naming an existing test when it proves the risk. A written test must compile on both sides. Do not guess paths or signatures missing from the supplied context. If the risk is unsupported or no differential test can demonstrate it, return empty path, content and name strings."
	return answer.Instruct(spec, answer.Encode(struct {
		Risk  candidate.Risk `json:"risk"`
		Files []contextFile  `json:"files"`
	}{c.Risk, context})), spec
}
