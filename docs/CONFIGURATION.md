# Configuration

Every setting the reviewer reads, its default, and what it changes (ADR-0013: nothing is hardcoded; secrets come from the
environment only). The names and defaults below were read from the code on `main` on 22 October 2026
(`internal/generator/settings.go`, `internal/guardrail/settings.go`, `internal/query/settings.go`,
`internal/review/config.go`, `cmd/livecheck`). A wrong value stops Proven before any work, and the refusal names the
setting, the value and what is allowed.

The shell's own flags (`proven scan …`) are documented by the shell's track. Sample mode (`--sample`) reads none of the
settings below except `proven.yaml`'s `timeout`: see [SAMPLE_MODE.md](SAMPLE_MODE.md).

<!-- generated: settings -->
## The model generator: Meterless Nightshift

| Environment variable | Default | What it changes |
|---|---|---|
| `METERLESS_RUNTIME_TOKEN` | none (required live) | the runtime token sent to Nightshift. Secret: never printed |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` | Nightshift's address |
| `PROVEN_GENERATOR_APP` | `proven` | the application name sent with each call |
| `PROVEN_GENERATOR_MODEL` | `meterless/auto` | the model asked for; Nightshift's router chooses when `auto` |
| `PROVEN_GENERATOR_TIMEOUT` | `60s` | the limit for one model call |
| `PROVEN_GENERATOR_REQUIRE_NVIDIA` | `all` | which agents must be answered by an `nvidia/*` model (by the stream's receipt); `all`, `none`, or a list of agents. A non-NVIDIA answer to such an agent is refused by name |
| `PROVEN_SKEPTIC_MODEL` | `nvidia/nemotron-3.5-super-vl-120b-a12b` | the Skeptic's model (ADR-0006: another family than the finder's) |

## The guardrail: NVIDIA Nemotron content safety

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_GUARDRAIL_MODE` | `server` | `server` (your own model server), `nvidia` (NVIDIA's hosted API) or `off` (the trace says "no guardrail") |
| `PROVEN_GUARDRAIL_BASE_URL` | `http://127.0.0.1:8081` for `server`; `https://integrate.api.nvidia.com/v1` for `nvidia` | the guardrail's address |
| `PROVEN_GUARDRAIL_MODEL` | `nemotron-3.5-content-safety` for `server`; `nvidia/nemotron-3.5-content-safety` for `nvidia` | the safety model. The hosted id is not yet confirmed by a live call |
| `PROVEN_GUARDRAIL_KEY` | none | the guardrail's key. Secret: never printed |
| `PROVEN_GUARDRAIL_TIMEOUT` | `5s` | the limit for one check |

Every prompt and every answer is checked; an `unsafe` verdict, or a guardrail that does not answer, stops that step by name.

## Evidence: Zetaris over REST

| Environment variable | Default | What it changes |
|---|---|---|
| `ZETARIS_REST_URL` | none (required live) | Zetaris's REST address; an absolute `http` or `https` URL with no user info, query or fragment |
| `ZETARIS_API_KEY` | none (required live) | Proven's Zetaris key. Secret: never printed |
| `ZETARIS_ORG_ID` | none (required live) | the organisation id sent with each query |
| `ZETARIS_ROW_LIMIT` | `100` | the most rows one query returns |
| `ZETARIS_TIMEOUT` | `60s` | the limit for one query |
| `PROVEN_EVIDENCE_VIEW` | `px_evidence.px_kev_vulns` | the view `livecheck` queries to prove Zetaris answers (livecheck only) |

## The Prover's container

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_TEST_RUNTIME` | `docker` | the container command line, by name or absolute path |
| `PROVEN_TEST_IMAGE` | `golang:1.26` | the image each test runs in |
| `PROVEN_TEST_CPUS` | `1` | the CPU limit of one test run |
| `PROVEN_TEST_MEMORY` | `512m` | the memory limit of one test run; a positive size with suffix b, k, m or g |
| `PROVEN_TEST_TIMEOUT` | `120s` | the limit for one test run, base or change |
| `PROVEN_TEST_MAX_WRITTEN_BYTES` | `65536` | the largest test file the Prover may write |

Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these.
<!-- end generated -->

## `proven.yaml`

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

### Default path rules

Each of these is class `security`:

`**/auth/**`, `**/payment*/**`, `**/*secret*`, `**/migrations/**`, `go.mod`, `go.sum`, `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `Cargo.toml`, `Cargo.lock`, `requirements.txt`, `Pipfile.lock`, `poetry.lock`, `Gemfile.lock`, `composer.lock`, `pom.xml`, `build.gradle`, `build.gradle.kts`.

Each of these is class `docs`:

`**/*.md`, `docs/**`.

Anything else is the fallback class.

## Check the live settings

```bash
go run ./cmd/livecheck
```

With nothing set, it reports every service `not ready` with the reason and exits 1. Its output never contains a secret.
