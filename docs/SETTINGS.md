# Settings

`proven.yaml` is one flat mapping of scalar values. A missing file is the defaults below. An unknown key, a nested value, or a bad value is refused on load and proven exits 7. The refusal names the field and the allowed values. A key that is set replaces that setting. It does not merge with the default list.

Where the file is read:

- An explicit config path is that file. A missing path is refused.
- Otherwise, when the review names a baseline commit, the file is the baseline commit's `proven.yaml`. The change's copy is not read. If the change edits `proven.yaml`, the review prints `the change edits proven.yaml`.
- Otherwise there is no baseline (a sample fixture, or a review of the whole tree). The file in the target is used.

No environment variable sets these keys.

| key | default | what it does | allowed values |
|---|---|---|---|
| `timeout` | `0s` | When greater than zero, bounds a review whose context has no deadline. Zero adds no deadline. The context's own deadline wins when it has one. | A Go duration of zero or greater, for example `30s`, `5m`, or `0s`. |
| `default_reviewer` | `NVIDIA` | The reviewer used when a class is not doubled, and the first reviewer when it is. Also the name in "default reviewer" when the change names no model. | A non-empty model name. |
| `special_reviewer` | empty | The second reviewer for a doubled class. Empty prints `special-review model: not configured`. | A model name, or empty. |
| `families` | `claude-*=Anthropic`, `gpt-*=OpenAI`, `codex-*=OpenAI`, `nemotron-*=NVIDIA`, `qwen*=Alibaba`, `NVIDIA=NVIDIA` | Maps a model name to a family. A model that matches nothing is `family unknown`. The same family as the author prints `yes`; a different known family prints `no`. | Comma-separated `pattern=name`. Any non-empty name is a family. An empty or malformed entry is refused. A bad pattern is refused. |
| `path_rules` | the list under Default path rules | Maps a path pattern to a review class. The first matching rule of the highest class wins. A pattern with no slash also matches the base name, so `go.mod` matches a nested module. | Comma-separated `pattern=class`. A class is one of `security`, `backend`, `mobile`, `user interface`, `docs`. Patterns use `*` and `**`. |
| `class_order` | `security`, `backend`, `mobile`, `user interface`, `docs` | Highest risk first. The matching rule whose class is earliest in this list wins. | Every class exactly once. An order that leaves a class out is refused, and the refusal lists the classes. |
| `fallback_class` | `backend` | The class when no path rule matches. The rule name is `fallback`. | One of the five classes. |
| `doubled_classes` | `security` | Classes that use the default reviewer and the special reviewer. | `off`, or a comma-separated list of classes. |
| `author_commit_trailers` | `on` | Read author trailers from commits after the baseline. The baseline commit itself is not read. | `on` or `off`. |
| `author_description` | `on` | Read author trailers from the pull-request description. | `on` or `off`. |
| `author_record` | `off` | Read author trailers from a file inside the review target. | `off`, or a relative path or pattern inside the target. An absolute path, a `..` segment, a symlink, and a path that resolves outside the target are refused before the file is read. |
| `author_trailer_keys` | `Assisted-by` | Trailer keys that can name an author. The value must be `agent:model`. A recognised key with any other value is reported as `author statement unreadable: <line>`, not as "not stated". | A comma-separated list of keys, each once. |
| `raise_class` | `off` | Accepted and stored. It does not change the class. | `on` or `off`. |

## Default path rules

Each of these is class `security`:

`**/auth/**`, `**/payment*/**`, `**/*secret*`, `**/migrations/**`, `go.mod`, `go.sum`, `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `Cargo.toml`, `Cargo.lock`, `requirements.txt`, `Pipfile.lock`, `poetry.lock`, `Gemfile.lock`, `composer.lock`, `pom.xml`, `build.gradle`, `build.gradle.kts`.

Each of these is class `docs`:

`**/*.md`, `docs/**`.

Anything else is the fallback class.
