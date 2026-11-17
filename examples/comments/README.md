# Code comment reports

```sh
go run ./examples/comments /tmp/proven-comments-example
```

This uses the real Go AST adapter and deterministic comment rules on synthetic source, then renders HTML, JSON and Markdown for two cases:

- `comments-deterministic`: actual structural checks, with no model.
- `comments-linguistic`: the same structural observations plus recorded model answers passed through the real response/citation validator.

The second case is explicitly labelled as a replay. Neither case executes the example application's code, calls a live model or proves its behavior. See [code comment checks](../../docs/CODE_COMMENTS.md).
