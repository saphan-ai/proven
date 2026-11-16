package prover

import (
	"path"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/testrun"
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
		if f.Path != c.Path && path.Dir(f.Path) == path.Dir(c.Path) && isContextTest(f.Path) {
			add(f)
		}
	}
	for _, f := range files {
		if (f.Path == "go.mod" || f.Path == "pyproject.toml" || f.Path == "package.json" || f.Path == "tsconfig.json") && f.Path != c.Path {
			add(f)
		}
	}
	spec := testSpec
	switch path.Ext(c.Path) {
	case ".py":
		spec.Example = `{"path":"tests/test_example.py","content":"def test_example():\n    assert 1 == 1\n","name":"test_example"}`
		spec.Task += " Use the pytest adapter (also supports unittest-style classes). Name a test_* function or TestClass::test_method."
	case ".ts", ".tsx", ".mts", ".cts":
		spec.Example = `{"path":"src/example.test.ts","content":"import { test, expect } from 'vitest';\ntest('example', () => { expect(1).toBe(1); });\n","name":"example"}`
		spec.Task += " Use Vitest with a .test.ts/.spec.ts (or corresponding TSX/MTS/CTS) file. Name the exact full test name, including describe blocks."
	case ".js", ".mjs", ".cjs":
		spec.Example = `{"path":"src/example.test.mjs","content":"import test from 'node:test';\nimport assert from 'node:assert/strict';\ntest('example', () => assert.equal(1, 1));\n","name":"example"}`
		spec.Task += " Use node:test in .test.js/.test.mjs/.test.cjs. For Vitest JavaScript use .spec.js/.spec.mjs/.spec.cjs."
	}
	spec.Task += " Use the supplied source and existing tests to locate the package and its API. Prefer naming an existing test when it proves the risk. A written test must compile on both sides. Do not guess paths or signatures missing from the supplied context. If the risk is unsupported or no differential test can demonstrate it, return empty path, content and name strings."
	return answer.Instruct(spec, answer.Encode(struct {
		Risk  candidate.Risk `json:"risk"`
		Files []contextFile  `json:"files"`
	}{c.Risk, context})), spec
}

func isContextTest(p string) bool {
	for _, a := range testrun.LanguageAdapters() {
		if a.Matches(p) {
			return true
		}
	}
	return false
}
