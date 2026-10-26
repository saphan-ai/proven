package trace

import "github.com/saphan-ai/proven/internal/envdoc"

// EnvLogDir is the trace directory. An empty value means the user cache.
// Sample mode reads it: the sample reviewer writes a trace too.
const EnvLogDir = "PROVEN_LOG_DIR"

// Environment is the trace directory row on the configuration page. It is
// declared here, not in internal/cli, which reads it: internal/cli imports
// the configuration page, so the page cannot import internal/cli.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{{
		Name:        EnvLogDir,
		Default:     envdoc.Tick("<user cache>/proven/logs"),
		Description: "where the trace is written. The directory is created with mode 0700. An existing configured directory is narrowed to 0700. A path that resolves inside the review target, including through a symlink, is refused. A letter-case alias of the target is refused on a volume that ignores case. That check is exercised only on darwin and Windows, and the skip on other volumes is measured, not assumed. Each file written in that directory, trace.jsonl, hmem.jsonl and world.jsonl, is refused when it is a symlink (wherever it points) or has more than one name. The link count is read on every OS. Sample mode reads this",
	}}
}
