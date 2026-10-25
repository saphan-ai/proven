# Configuration

Every setting the reviewer reads, its default, and what it changes (ADR-0013: nothing is hardcoded; secrets come from the
environment only). The names and defaults below were read from the code on `main` on 22 October 2026
(`internal/generator/settings.go`, `internal/guardrail/settings.go`, `internal/query/settings.go`,
`internal/review/config.go`, `cmd/livecheck`). A wrong value stops Proven before any work, and the refusal names the
setting, the value and what is allowed. An empty secret does not: that service is down and the review names it, and an empty guardrail key refuses every model call unless `PROVEN_GUARDRAIL_MODE` is `off`.

The shell's own flags (`proven scan …`) are documented by the shell's track. Sample mode (`--sample`) reads none of the
settings below except `proven.yaml`'s `timeout`: see [SAMPLE_MODE.md](SAMPLE_MODE.md).

<!-- generated: settings -->
## The model generator: Meterless Nightshift or NVIDIA's hosted API

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_GENERATOR_MODE` | `nightshift` | `nightshift` (Meterless Nightshift) or `nvidia` (NVIDIA's hosted API, ADR-0031) |
| `METERLESS_RUNTIME_TOKEN` | none (required in `nightshift` mode) | the runtime token sent to Nightshift. Secret: never printed. Empty: Nightshift is down and the review names it. Not read in `nvidia` mode |
| `PROVEN_GENERATOR_KEY` | none (required in `nvidia` mode) | the NVIDIA API key, from build.nvidia.com. Secret: never printed. Empty: the hosted API is down and the review names it. Not read in `nightshift` mode |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` | Nightshift's address. Read in `nightshift` mode only; `nvidia` mode never reads it |
| `PROVEN_NVIDIA_BASE_URL` | `https://integrate.api.nvidia.com/v1` | the hosted API's address. Read in `nvidia` mode only. It must be `https://`: any other value is refused by name, so the key is never sent over plain http. The answer's `model` field is trusted as the answering model only from this address (ADR-0031) |
| `PROVEN_GENERATOR_APP` | `proven` | the application name sent with each call to Nightshift. Not sent in `nvidia` mode |
| `PROVEN_GENERATOR_MODEL` | `meterless/auto` for `nightshift`; `nvidia/llama-3.3-nemotron-super-49b-v1.5` for `nvidia` | the model asked for; Nightshift's router chooses when `auto`. The hosted default is from NVIDIA's published list and is not yet confirmed by a live call |
| `PROVEN_GENERATOR_TIMEOUT` | `60s` | the limit for one model call |
| `PROVEN_GENERATOR_REQUIRE_NVIDIA` | `all` | which agents must be answered by an `nvidia/*` model (by Nightshift's receipt, or in `nvidia` mode by the `model` field of the hosted API's answer); `all`, `none`, or a list of agents. A non-NVIDIA answer to such an agent is refused by name |
| `PROVEN_SKEPTIC_MODEL` | `nvidia/nemotron-3.5-super-vl-120b-a12b` | the Skeptic's model (ADR-0006: another family than the finder's) |

Which keys each mode needs (ADR-0031). `nightshift` (the default, ADR-0019): `METERLESS_RUNTIME_TOKEN`, issued by the operator of the Nightshift you run; nothing else is required. `nvidia`: `PROVEN_GENERATOR_MODE=nvidia` and `PROVEN_GENERATOR_KEY`, a free NVIDIA API key created at https://build.nvidia.com (sign in, open any model, choose *Get API Key*); the hosted API's address (`PROVEN_NVIDIA_BASE_URL`, `https://` only) and the models have defaults, so the key is the only value needed. Each mode reads its own address: `nightshift` reads `PROVEN_GENERATOR_BASE_URL`, `nvidia` reads `PROVEN_NVIDIA_BASE_URL` and ignores the other, so the key is never sent to a Nightshift address. The guardrail has its own key in either mode (below). In `nvidia` mode the hosted API gives no signed receipt, so the trace says *no receipt* and the `nvidia/*` requirement is checked against the `model` field of its answer. A free key is rate limited; a review that meets the limit names it as a failed service. Sample mode reads none of these.

## The second reviewer: any OpenAI-compatible model

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_SECOND_REVIEWER_KEY` | none (empty: no second reviewer) | the API key of the second reviewer. Secret: never printed. Empty: not connected, and the review has one reviewer |
| `PROVEN_SECOND_REVIEWER_BASE_URL` | none (required with the key) | the address of any OpenAI-compatible chat API, for example `https://api.openai.com/v1` or a local server |
| `PROVEN_SECOND_REVIEWER_MODEL` | none (required with the key) | the model asked for; the review names it as the second reviewer |

## The guardrail: NVIDIA Nemotron content safety

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_GUARDRAIL_MODE` | `server` | `server` (your own model server), `nvidia` (NVIDIA's hosted API) or `off` (the trace says "no guardrail") |
| `PROVEN_GUARDRAIL_BASE_URL` | `http://127.0.0.1:8081` for `server`; `https://integrate.api.nvidia.com/v1` for `nvidia` | the guardrail's address |
| `PROVEN_GUARDRAIL_MODEL` | `nemotron-3.5-content-safety` for `server`; `nvidia/nemotron-3.5-content-safety` for `nvidia` | the safety model. The hosted id is not yet confirmed by a live call |
| `PROVEN_GUARDRAIL_KEY` | none | the guardrail's key. Secret: never printed. Empty is not off: every model call is refused |
| `PROVEN_GUARDRAIL_TIMEOUT` | `5s` | the limit for one check |

Every prompt and every answer is checked; an `unsafe` verdict, or a guardrail that does not answer, stops that step by name. An empty key is not off: every model call is refused and the review names the guardrail.

## Evidence: the store, and Zetaris over REST

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_EVIDENCE_STORE` | `zetaris` | which evidence store a live review uses: `zetaris`, `local` or `couchdb`. Empty means `zetaris`. The review passes a template request, and the store runs the form of its own language (`sql` or `mango`). Sample mode uses the local store and does not read this variable |
| `PROVEN_EVIDENCE_FILE` | none (required when the store is `local`) | the one file of the local store, read only when the store is `local`. Empty is refused by name, and so is a path inside the repository under review. The file holds tables and no server is started |
| `PROVEN_COUCHDB_URL` | none (required when the store is `couchdb`) | CouchDB's address; an absolute `http` or `https` URL with no user info, query or fragment. Empty: CouchDB is down and the review names it |
| `PROVEN_COUCHDB_PREFIX` | `proven_` | the start of every database name. One database holds one registered table, named the prefix, then the table name in lower case with `.` written as `-`. A lower-case letter, then at most 31 lower-case letters, digits or `_` |
| `PROVEN_COUCHDB_USER` | none (required when the store is `couchdb`) | the CouchDB user Proven reads as. One line of printable ASCII with no `:`. Empty: CouchDB is down and the review names it |
| `PROVEN_COUCHDB_PASSWORD` | none (required when the store is `couchdb`) | that user's password. Secret: never printed. Empty: CouchDB is down and the review names it |
| `PROVEN_COUCHDB_ROW_LIMIT` | `100` | the most rows one query returns |
| `PROVEN_COUCHDB_TIMEOUT` | `60s` | the limit for one request to CouchDB |
| `PROVEN_COUCHDB_RECORDS` | none (records are not written) | the CouchDB database that holds a review's evidence records, on the server `PROVEN_COUCHDB_URL` names. Set: each finding of a review is written to it as one document, with the design document and the viewer page. Empty: nothing is written. A lower-case letter, then at most 63 lower-case letters, digits, `_` or `-` |
| `PROVEN_COUCHDB_RECORDS_USER` | none (required when `PROVEN_COUCHDB_RECORDS` is set) | the CouchDB user Proven writes the records as. It must be a different user from `PROVEN_COUCHDB_USER`, the read-only account the viewer is opened with. One line of printable ASCII with no `:` |
| `PROVEN_COUCHDB_RECORDS_PASSWORD` | none (required when `PROVEN_COUCHDB_RECORDS` is set) | that writing user's password. Secret: never printed |
| `PROVEN_EVIDENCE_LINK_BASE_URL` | none (no evidence-record link) | the address a reader reaches CouchDB at, as an absolute `http` or `https` URL with no user info, query or fragment. Set, with `PROVEN_COUCHDB_RECORDS`: each finding carries a link to its record in the viewer page. Empty: no link. It is never `/_utils`, the admin console |
| `ZETARIS_REST_URL` | none (required live) | Zetaris's REST address; an absolute `http` or `https` URL with no user info, query or fragment |
| `ZETARIS_API_KEY` | none (required live) | Proven's Zetaris key. Secret: never printed. Empty: Zetaris is down and the review names it |
| `ZETARIS_ORG_ID` | none (required live) | the organisation id sent with each query |
| `ZETARIS_ROW_LIMIT` | `100` | the most rows one query returns |
| `ZETARIS_TIMEOUT` | `60s` | the limit for one query |
| `PROVEN_OSV_TABLE` | `px_osvfile.px_osv_pkg` | the OSV package table the dependency check queries; see [OSV_TABLE.md](OSV_TABLE.md). Sample mode uses the default and does not read this variable |
| `PROVEN_KEV_VIEW` | `px_evidence.px_kev_vulns` | the KEV view an advisory's CVE alias is looked up in. Sample mode uses the default and does not read this variable |
| `PROVEN_EPSS_VIEW` | `px_evidence.px_epss_cve_rows` | the EPSS view a finding's CVE is scored from. Sample mode uses the default and does not read this variable |
| `PROVEN_EVIDENCE_VIEW` | `px_evidence.px_kev_vulns` | the view `livecheck` queries to prove Zetaris answers (livecheck only) |

The store is zetaris unless PROVEN_EVIDENCE_STORE says local or couchdb. The local store is one file and opens no network (ADR-0034, ADR-0047), and a review that chooses it does not read the Zetaris variables. The couchdb store reads the PROVEN_COUCHDB variables and not the Zetaris ones. A live review passes the template request, and each store runs the form of its own language: sql on Zetaris and the local store, mango on CouchDB (ADR-0047, docs/ADAPTERS.md). Sample mode uses the local store on the fixture file and reads none of these, except that it reads PROVEN_COUCHDB_RECORDS and PROVEN_EVIDENCE_LINK_BASE_URL, and, when PROVEN_COUCHDB_RECORDS is set, also the other PROVEN_COUCHDB_* variables (URL, USER, PASSWORD, PREFIX, ROW_LIMIT, TIMEOUT, RECORDS_USER, RECORDS_PASSWORD) and GITHUB_REF, and nothing else besides OTEL_*, so the lab's review can write its records (docs/ADAPTERS.md).

## The Prover's container

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_TEST_RUNTIME` | `docker` | the container command line, by name or absolute path |
| `PROVEN_TEST_IMAGE` | `golang:1.26` | the image each test runs in |
| `PROVEN_TEST_CPUS` | `1` | the CPU limit of one test run |
| `PROVEN_TEST_MEMORY` | `512m` | the memory limit of one test run; a positive size with suffix b, k, m or g |
| `PROVEN_TEST_TIMEOUT` | `120s` | the limit for one test run, base or change |
| `PROVEN_TEST_MAX_WRITTEN_BYTES` | `65536` | the largest test file the Prover may write |
| `PROVEN_TEST_USER` | `65534:65534` | the unprivileged user and group the test runs as, as numbers; root is refused |
| `PROVEN_TEST_PIDS` | `256` | the most processes one test run may start; at most 4096 |
| `PROVEN_TEST_TMP_SIZE` | `512m` | the size limit of the writable scratch mount; a positive size of at most 1g with suffix b, k, m or g |

Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these.

## Deterministic tools

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_TOOL_TIMEOUT` | `120s` | the limit for one side of a deterministic tool run, base or change. A run past it is named and is not evidence. Sample mode uses the default and does not read this variable |
| `PROVEN_GOVULNCHECK_DB` | `empty` | the file URL of the local govulncheck database, inside the checkout mount, for example file:///checkout/vulndb. Empty means govulncheck is not run and Go dependency findings are not measured. A network address is refused. Sample mode does not read this variable |

Each tool runs in the Prover's container with no network and no secret (ADR-0025, ADR-0040). A tool that cannot run is named. Sample mode reads none of these.

## One adapter call

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_ADAPTER_TIMEOUT` | `5s` | the limit on one adapter call when the caller set no shorter deadline. Sample mode uses the default and does not read this variable |

## Telemetry: OpenTelemetry (ADR-0045)

| Environment variable | Default | What it changes |
|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | none (telemetry is off) | the base URL of an OTLP/HTTP collector, for example `http://127.0.0.1:4318`. Setting it turns telemetry on; the trace goes to `<url>/v1/traces` and the metrics to `<url>/v1/metrics`. Sample mode reads this variable too |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | none | the full URL the trace is sent to. It wins over `OTEL_EXPORTER_OTLP_ENDPOINT` for the trace and, set alone, turns the trace on |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | none | the full URL the metrics are sent to. It wins over `OTEL_EXPORTER_OTLP_ENDPOINT` for the metrics and, set alone, turns the metrics on |
| `OTEL_SDK_DISABLED` | none | `true` turns telemetry off even when an endpoint is set |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf` | the OTLP transport. Only `http/protobuf` is built in; any other value is refused by name. `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` and `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` are read the same way |
| `OTEL_SERVICE_NAME` | `proven` | the service name a backend shows |
| `OTEL_METRIC_EXPORT_INTERVAL` | `60000` | how often the metrics are pushed, in milliseconds. The lab sets it low so a dashboard moves while a review runs |

Telemetry is off unless an endpoint is set. It only observes: once a review has started it never changes what the review decides, and a collector that cannot be reached never fails a review (the export is bounded at 5 seconds, and the failure goes to stderr). The one refusal is at the start: a malformed `OTEL_*` setting stops `Open` by name, like any other setting. No secret is put in an attribute, an event or a span status: the same redactor that cleans the trace file cleans the span events and the text of every failed review and failed agent, so a status carries the redactor's marker where a configured secret stood. The headers, timeout and certificates of the exporter are the SDK's own variables (`OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_TIMEOUT`, `OTEL_EXPORTER_OTLP_CERTIFICATE`), read from the process environment. Sample mode reads the `OTEL_*` variables and no other.

## Reads of the repository under review

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_MAX_ARCHIVE_BYTES` | `134217728` | the most one export reads from the repository under review: a git archive, or the tracked files of a working tree |
| `PROVEN_MAX_FILE_BYTES` | `8388608` | the largest single file, git blob or commit log read from the repository under review |
| `PROVEN_MAX_CONFIG_BYTES` | `1048576` | the largest `proven.yaml` Proven reads |
| `PROVEN_MAX_RECORD_BYTES` | `1048576` | the most the author record may hold, summed over the files `author_record` names |

A read over its bound is refused by name before the review goes on; the refusal names the file and the bound, never its content. Sample mode reads none of these.

## The organisation policy

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_POLICY` | none | a path to the organisation policy file, read before each repository's `proven.yaml`; see [POLICY.md](POLICY.md). A path inside the repository under review is refused, and so is a missing or unreadable file. Sample mode reads it too |

A repository's `proven.yaml` may make a policy setting stricter, never weaker, and an exception without an owner or an expiry is refused on load (ADR-0036).

## The evidence pack

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_PACK_KEY` | none (required with `--evidence-pack`) | path to the PEM file of the organisation's Ed25519 private key (PKCS #8) that signs a pack. A missing or unreadable file, or one that group or others can reach (fix: `chmod 600`), is named and the review does not start |
| `PROVEN_PACK_PUBLIC_KEY` | none (required by `proven verify` unless `--public-key` is given) | path to the PEM file of the Ed25519 public key (PKIX) that verifies a pack |

The pack holds source excerpts and query rows, so keep it where the review itself may be read. The private key is never written into it (ADR-0035). See `docs/EVIDENCE_PACK.md`.
<!-- end generated -->

## Counting guardrail checks

The guardrail's settings are in the generated table above; this section is what its checks leave on the trace.

Each check is one `kind: "guardrail"` line in `logs/trace.jsonl`. Its `guardrail` object carries:

| field | meaning |
|---|---|
| `side` | `prompt` or `answer` |
| `verdict` | `safe` or `unsafe`; empty when the check could not be made |
| `categories` | the categories the guardrail named |
| `latency_ms` | how long the check took |
| `backend` | the adapter that answered: `server`, `nvidia`, `sample` (sample mode, answers safe), `sample-recorded` (a fixture's recorded verdicts), `absent` (no key), `off`, or `unknown` for an adapter that does not name itself |
| `model` | the configured model id (`PROVEN_GUARDRAIL_MODEL` or its default). Empty for every backend that calls no model: `sample`, `sample-recorded`, `absent`, `off` |
| `down` | set only when the check could not be made: `not-configured`, `no-key`, `off`, `bad-verdict`, `timeout`, `unauthorized`, `empty-answer`, `unreachable`, `no-recorded-verdict`, `http-<status>`, or `other` for any error outside that set |

The line's `text` is unchanged: a check that could not be made still reads `guardrail: <reason>`, and an error outside the set keeps its text there only. The key is never written to the trace. A review with the guardrail off writes one `no guardrail` line and no checks.

`guardcount` reads one or more traces and counts the checks per UTC day, side, verdict and backend, with the median latency. A down check's verdict is `down:<reason>`. An older down line with no `guardrail` object counts as `down:unstructured`, and a field an older line lacks prints as `-`.

```bash
go run ./cmd/guardcount logs/trace.jsonl [more traces...]
```

Its output on the sample traces in `cmd/guardcount/testdata` (the golden file `counts.golden`, which the test compares byte for byte):

```text
day         side    verdict            backend          count  median_ms
2026-10-24  answer  down:http-503      nvidia           1      220
2026-10-24  answer  safe               server           1      80
2026-10-24  answer  unsafe             server           1      95
2026-10-24  prompt  down:timeout       server           1      5000
2026-10-24  prompt  safe               nvidia           2      355
2026-10-24  prompt  safe               server           3      100
2026-10-25  -       down:unstructured  -                1      0
2026-10-25  answer  safe               sample           1      0
2026-10-25  prompt  down:no-key        absent           1      0
2026-10-25  prompt  safe               sample           1      0
2026-10-25  prompt  unsafe             sample-recorded  1      0
```

The sample traces are written for the test, not taken from a live review. A trace of a review with the guardrail off prints the header line only, and exits 0.

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
| `fixes` | `on` | Proposes a fix for a finding a test proved, and a fixed version for a dependency finding (ADR-0042). `off` proposes none, runs none of the three proof runs and names no version. | `on` or `off`. |
| `fix_attempts` | `2` | The most attempts at one fix: the first try and, up to this number, a retry by the same model with the failure of the last try in hand. A fix that is not proved after the last attempt is reported as not proved, with no patch. | A whole number from `1` to `5`. With `autofix: on` this key is not read: the whole-change attempt has no retry (ADR-0050). |
| `autofix` | `off` | The one automated fix attempt for the whole change (ADR-0050). With `on`, after the review, one generator call gets every proved finding with its evidence and writes one patch. Proven applies it to copies only and runs each finding's own evidence again: a finding proved by a test closes when that test passes and the package's other tests still pass; one proved by data closes when the same query no longer returns the row; one proved by a tool closes when the same tool run no longer reports it. Then nothing else may have broken: Proven runs `go test` on the patched copy for every package a patched file belongs to (the whole module for an edit to a module file), and a failure, or a run that did not happen, makes the attempt `rejected` and names the run. A touched package with no test cannot show that nothing broke, so a patch to one is `rejected` too. No model judges the patch. If every proved finding closed and those runs passed, the review says `fixed automatically, waiting for a person`; otherwise it says `rejected` and lists each open finding with its evidence. There is no second attempt, nothing is committed to the author's branch, and the per-finding suggestions of `fixes` are not made while it is on. `fixes: off` together with `autofix: on` does not stop the attempt: `fixes` governs only the per-finding suggestions, so the attempt still runs its own proof runs in the container. The judges' lab turns it on. | `on` or `off`. |
| `autofix_generator` | `default` | Which generator writes the patch of the whole-change attempt. `default` is the review's own generator (NVIDIA's Nemotron unless the generator mode names another). `special` is the second reviewer's generator; with none connected the review makes no attempt and names the service as failed. | `default` or `special`. |
| `autofix_timeout` | `10m` | The time limit of the whole-change attempt: the generator's answer and every run of evidence again. An attempt that runs past it ends `rejected` and says the time limit ended it. | A Go duration greater than zero, for example `90s` or `10m`. |
| `investigator_templates` | every template | The query templates the Investigator may ask (ADR-0013, ADR-0032). Its prompt lists only these, and a request for any other is refused before a store is asked: the finding goes on with no query evidence and the trace says `refused: query template is outside the evidence scope of this role: <template> for the investigator`. Unset narrows nothing. | A comma-separated list of template names from `docs/QUERIES.md`, each once. A name that is not on the list, a name twice or an empty entry is refused on load. |
| `skeptic_templates` | every template | The query templates the Skeptic may ask. Its prompt lists only these, and a refuting query for any other is refused before a store is asked: the proposal is dropped, the finding is judged on the rest of the attempt, and the refutation says `refused: query template is outside the evidence scope of this role: <template> for the skeptic`. The two lists are independent, so the Skeptic can be held to fewer templates than the Investigator. Unset narrows nothing. | The same as `investigator_templates`. |

The Planner, the Prover, the Judge, the Writer and the Fixer have no query path, so they have no scope key. A template name is the only thing a scope names: it does not narrow the tables or views a template reads.

For example, an Investigator that may ask every KEV, EPSS and OSV template and a Skeptic that may ask only KEV:

```yaml
investigator_templates: kev-by-cve, epss-by-cve, osv-by-package, osv-by-advisory
skeptic_templates: kev-by-cve
```

### Default path rules

Each of these is class `security`:

`**/auth/**`, `**/payment*/**`, `**/*secret*`, `**/migrations/**`, `go.mod`, `go.sum`, `package.json`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `Cargo.toml`, `Cargo.lock`, `requirements.txt`, `Pipfile.lock`, `poetry.lock`, `Gemfile.lock`, `composer.json`, `composer.lock`, `pom.xml`, `build.gradle`, `build.gradle.kts`, `libs.versions.toml`, `gradle.lockfile`.

Each of these is class `docs`:

`**/*.md`, `docs/**`.

Anything else is the fallback class.

## Check the live settings

```bash
go run ./cmd/livecheck
```

With nothing set, it reports every service `not ready` with the reason and exits 1. Its output never contains a secret.

## The judges' lab

`make` reads these from `lab/versions.env`. A variable already set in the environment wins. Secrets are not in that file: copy `lab/.env.example` to `lab/.env` for `make livecheck`. `lab/.env` is not committed.

| Setting | Default | What it changes |
|---|---|---|
| `COMPOSE_PROJECT_NAME` | `proven-lab` | Compose project name |
| `LAB_BIND` | `127.0.0.1` | Address the published ports bind to |
| `GOLANG_IMAGE` | `golang:1.26.10` | Image used when `make build` has no usable Go, and for Proven's image build. Proven's Dockerfile repeats this as the `ARG` default; `make` still passes this setting, and that value wins |
| `RUNTIME_IMAGE` | `alpine:3.22` | Base of Proven's image and of the seed image. Both Dockerfiles repeat this as the `ARG` default so a build with no `--build-arg` has a real base; `make` still passes this setting, and that value wins |
| `PROVEN_REGISTRY` | `docker.io` | Registry `make push` pushes to |
| `PROVEN_IMAGE` | `saphan/proven` | Image name, without the registry. Tagged `latest` and the `git describe` value. Each tag is one manifest list of `PUSH_PLATFORMS` |
| `PUSH_PLATFORMS` | `linux/amd64,linux/arm64` | Platforms `make push` puts in each tag. `make image` does not use this: a local engine cannot load a manifest list |
| `SCOUT_REFUSE_SEVERITIES` | `critical,high` | Severities that make `make scan` refuse when the `docker scout` plugin is installed and a fix is available. `make push` does not run Scout |
| `GITEA_VERSION` | `1.22.6` | Gitea image tag |
| `GITEA_HTTP_PORT` | `3000` | Host port for the Gitea web UI. The lab publishes three ports in all, each on `LAB_BIND`: this one, `LGTM_GRAFANA_PORT` and `LGTM_OTLP_PORT` |
| `GITEA_SSH_PORT` | `2222` | Kept. Not published. The lab does not serve SSH |
| `GITEA_DISABLE_SSH` | `true` | `GITEA__server__DISABLE_SSH`. Turns Gitea SSH off |
| `GITEA_START_SSH_SERVER` | `false` | `GITEA__server__START_SSH_SERVER`. Gitea's own SSH server stays off, so it does not bind port 22 |
| `GITEA_CONTAINER_HTTP_PORT` | `3000` | Port Gitea listens on inside the lab network |
| `GITEA_INTERNAL_HOST` | `gitea` | Name of the Gitea service on the lab network |
| `GITEA_RUN_USER` | `git` | Account `make up` uses for every `gitea` command in the container. The image refuses that command as root, and `compose exec` is root when no user is named |
| `DIND_HOST` | `dind` | Name of the Docker-in-Docker service |
| `DIND_PORT` | `2375` | Port that daemon listens on inside the lab network. Not published on the host |
| `LGTM_IMAGE` | `grafana/otel-lgtm` | The dashboard's image: collector, Prometheus, Tempo, Loki and Grafana in one container (ADR-0045). It is not part of Proven's image |
| `LGTM_VERSION` | `0.11.0` | That image's tag. Not pulled when it was written; the first `make up` is the check |
| `LGTM_INTERNAL_HOST` | `lgtm` | Name of that service on the lab network. The workflow's `OTEL_EXPORTER_OTLP_ENDPOINT` is built from it |
| `LGTM_GRAFANA_PORT` | `3001` | Host port for Grafana, and so for the dashboard |
| `LGTM_GRAFANA_CONTAINER_PORT` | `3000` | Port Grafana listens on inside the lab network |
| `LGTM_OTLP_PORT` | `4318` | Host port for the collector's OTLP/HTTP endpoint, so a review run on the host can be exported to the lab too: `OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318` |
| `LGTM_OTLP_CONTAINER_PORT` | `4318` | Port the collector listens on inside the lab network |
| `LGTM_DASHBOARD_UID` | `proven-review` | The dashboard's uid, from `lab/grafana/dashboards/proven-review.json`. `make up` prints the address built from it |
| `LAB_OTEL_SERVICE_NAME` | `proven` | `OTEL_SERVICE_NAME` the lab's pull-request review is given |
| `LAB_OTEL_METRIC_EXPORT_INTERVAL` | `1000` | `OTEL_METRIC_EXPORT_INTERVAL` it is given, in milliseconds. Short, so the dashboard moves while the review runs |
| `LAB_SPECIAL_REVIEWER` | `qwen` | `special_reviewer` in the `proven.yaml` the lab's review runs with, so the dashboard has two reviewer models to show side by side (ADR-0009). Empty: one reviewer. The sample generator answers for both, so the findings do not change |
| `LGTM_WAIT_SECONDS` | `120` | How long `make up` waits for Grafana before it goes on and names the delay |
| `TRACE_WAIT_SECONDS` | `90` | How long `make trace` waits for the review's trace to reach the store |
| `ACT_RUNNER_VERSION` | `0.2.11` | Gitea Actions runner image tag |
| `DOCKER_VERSION` | `29.9.1` | Tag of the `docker` dind and cli images |
| `PROVEN_TEST_IMAGE` | `golang:1.26` | The image the Prover runs each test in (also in the table above). `make up` pulls it into the lab's inner Docker daemon before the first review, so the Prover does not meet `No such image` there |
| `TEST_IMAGE_DIGEST` | empty | `sha256:<64 hex>` of `PROVEN_TEST_IMAGE`. When set, `make up` pulls `<image>@<digest>` and tags it with the `PROVEN_TEST_IMAGE` name. When empty, `make up` pulls by tag and prints the digest it got, so the pin can be written here. The image name must have no registry port |
| `RUNNER_NAME` | `proven-lab` | Name the runner registers with |
| `RUNNER_LABEL` | `ubuntu-latest` | Label the lab's jobs use. The workflow file's `runs-on` matches this default |
| `LAB_ADMIN_USER` | `proven` | Admin user. `make up` creates it when the account is missing, including on a volume that already holds the password |
| `LAB_ADMIN_EMAIL` | `proven@localhost` | That user's email |
| `LAB_BOT_USER` | `proven-bot` | The account the workflow posts its review as. Not an admin; a collaborator on the demo repository. The seed makes it, keeps its generated password in the lab's state volume and stores its token only as the repository's `PROVEN_TOKEN` secret. `make e2e` accepts a review only from this account |
| `LAB_BOT_EMAIL` | `proven-bot@lab.local` | That account's email |
| `LAB_DEMO_REPO` | `discount` | Demo repository name |
| `LAB_DEMO_BASE` | `main` | Base branch |
| `LAB_DEMO_HEAD` | `change` | Branch the pull request comes from |
| `SEED_IMAGE` | `saphan/proven-lab-seed` | Name of the seed image we build. `make push` publishes `PROVEN_REGISTRY/SEED_IMAGE:SEED_IMAGE_TAG` as one manifest list |
| `SEED_IMAGE_TAG` | `1` | Pin for that image. It is not tagged `latest` |
| `E2E_WAIT_SECONDS` | `600` | How long `make e2e` waits for the review comment |
| `E2E_POLL_SECONDS` | `5` | How often it asks Gitea |

`make livecheck` reads these keys from `lab/.env` when the file exists, and ignores an empty value: `METERLESS_RUNTIME_TOKEN`, `PROVEN_GUARDRAIL_KEY`, `ZETARIS_API_KEY`, `ZETARIS_REST_URL`, `ZETARIS_ORG_ID`.
