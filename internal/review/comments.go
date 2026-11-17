package review

import "github.com/saphan-ai/proven/internal/envdoc"

// EnvCodeComments selects optional semantic review or deterministic checks.
const EnvCodeComments = "PROVEN_CODE_COMMENTS"

// CommentsEnvironment describes the code-comment setting for generated docs.
func CommentsEnvironment() []envdoc.Setting {
	return []envdoc.Setting{{Name: EnvCodeComments, Default: "deterministic", Description: "Go AST comment checks: off, deterministic, or model (also assesses language/code consistency; requires a connected model). Separate from proved code findings."}}
}
