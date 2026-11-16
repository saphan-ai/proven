# Troubleshooting

Start with the report's completion status and failed services, then inspect the trace. A zero findings count is not enough to identify success. Preserve the exact base/head and profile before changing settings, so a rerun answers a specific question.

## Interpret the outcome

| Symptom | Inspect | Next action |
| --- | --- | --- |
| Exit 0 with no findings | `completion_status`, `partial`, failed services, unresolved hypotheses; which stages were enabled | Establish whether analysis completed and what actually ran. Advisory mode may exit 0 with gaps. |
| Exit 1 | Kept findings and their evidence | Review the proof and policy; `--error` makes findings blocking. |
| Exit 3 | Coverage gaps and failed services in the written result | Resolve the named dependency, source, model or planning issue. Required completion takes precedence over finding enforcement. |
| Exit 2 and no complete report | Stderr and any trace | A fatal execution/refusal occurred; do not treat a missing JSON file as an empty review. |
| Exit 7 before review | Named configuration field and allowed values | Correct the setting. A refused network destination or unsafe state path is configuration, not a model verdict. |

[Readiness](READINESS.md#reading-completion-correctly) gives the completion contract. JSON and SARIF layouts are in [Output formats](OUTPUT_FORMATS.md).

## Model connection and generation

| Symptom | Likely distinction | Check or adjustment |
| --- | --- | --- |
| Authentication failure | Missing/expired key, wrong endpoint or account | Update the CI secret or local environment. Check the configured recipient; never paste keys into logs. Retries do not fix authentication. |
| Headers never arrive | Connection, queue or endpoint delay | Inspect time-to-headers; tune `HEADER_TIMEOUT` only after checking address and service availability. |
| Model appears silent while thinking | Reasoning is arriving, answer content has not | Compare reasoning bytes, body activity and total duration. Idle clocks reset on reasoning; total attempt time still applies. |
| Idle timeout | No body bytes after headers for the configured interval | Inspect provider behavior and load. Increasing the total timeout alone does not change the idle clock. |
| Total timeout | One attempt exhausted its window | Compare all three clocks, output budget and remaining review time. Timeouts are not retried. |
| `finish_reason=length` / response token limit reached | Provider truncated the completion, possibly after spending tokens on reasoning | Raise the relevant agent token budget or use documented thinking controls; measure again. Do not assume partial JSON is a valid answer. |
| Empty or invalid JSON | No usable answer, vendor envelope, incompatible schema behavior or truncation | Inspect the named error and bounded answer. Use final text output for CLI wrappers. Empty/schema recovery shares one configured budget. |
| HTTP 429 | Rate limit, burst or quota | Inspect `Retry-After`; use bounded pre-stream retries/backoff for transient bursts. A persistent quota needs a provider/account change or later run. |
| More HTTP requests than logical model calls | Recovery, HTTP retries or failed attempts | Count `model-attempt` events separately from successful `model` identity events. |
| Second provider rejects an unknown field | Primary provider options were inherited | Set the second reviewer's extra-body maps to `{}` or provider-specific replacements. |

Use `proven model check` for one request to the API. It deliberately has no retry or repair and uses the Investigator's profile; a success does not exercise the Prover's longer output. See [Models](MODELS.md).

## Review stopped late or has open hypotheses

`partial: true` records interruption. An assembled result may contain candidates already checked and name work left unfinished. A final `unresolved_hypotheses` entry means bounded planning ended with a claim still awaiting examination; execution can finish with `partial: false` and incomplete coverage. Existing risks continue through evidence checks.

A mid-plan `CarryFullError` is different: carried state cannot safely fit without forgetting claims, so planning refuses to continue. It can exit 2 before a complete report. Inspect the trace and bounded-planning settings rather than discarding state. Increasing a buffer may permit a run, but it does not establish review quality. See [Markovian](MARKOVIAN.md).

For deadline exhaustion, inspect the number of planning steps, candidates, recovery requests, test runs and second-reviewer work. Keep time for finalization. Switching Markovian off is an explicit comparison configuration, not evidence that bounded review is unnecessary or broken for every model.

## Source and database problems

| Symptom | Check |
| --- | --- |
| OSV/KEV/EPSS source unavailable | The selected store, exact registered relation name and imported snapshot. A new SQLite file has no public feeds. |
| Imported data but still no finding | Ecosystem/package spelling, added-or-raised dependency, lock precedence, supported version syntax and affected-range boundary. |
| Empty query rows | Whether the table exists, its scope, revision and filters. No row is weaker than a universal claim of safety. |
| Different source identity after an update | Row/column content and supplied revision. Identity is expected to change when inputs change. |
| One table updated, others did not | Imports are transactional per table, not for the whole multi-table file. Prepare coherent releases externally when needed. |
| PostgreSQL/CouchDB connection failure | Credentials, TLS, endpoint availability and network policy. Do not print connection URLs containing secrets. |

Use the [SQLite exercise](DATA_WALKTHROUGH.md) to test parsing/import/query independently of models and containers. See [Databases](DATABASES.md), [Sources](SOURCES.md) and [version rules](OSV_TABLE.md).

## Generated tests did not run

Check Docker/runtime availability, local runner image, its toolchain and offline dependencies, bind-mount paths, resource limits and the test's supported language. The product image and runner image are different. Runner images are not pulled during proof; network is disabled. A default Go image cannot fetch your private module while running a proof.

An inaccessible path is common when Proven itself runs in a container but uses the host's Docker daemon. The daemon resolves bind paths on its host. A Docker client and socket do not make container-only paths visible there. Use a host binary with correctly prepared paths or deliberately arrange matching mounts. See [Runner](RUNNER.md).

A test must pass on base and fail on head under the proving rules. A build failure, a test that never ran or an identical failure on both sides is not the requested differential proof. Race/govulncheck adapters have their own requirements; a manifest adapter alone does not add a proving toolchain.

## Fix rejected or absent

Only a proved finding is eligible for a proved edit. Inspect base/head/patched/package runs; no-run and collateral failures reject the patch. A dependency's suggested fixed version comes from advisory data and is not equivalent to a tested source patch. Whole-change autofix is opt-in, one attempt and never commits to the author branch. [Fix validation](FIXES.md) explains the gates.

## Workflow failed but a report exists

Read the report before rerunning. A required-completion failure can be the intended gate exposing missing coverage. Artifact upload and PR comment steps should retain that result even when review exits 3. Verify the selected image digest contains the source change you expect; editing the repository does not replace a pinned reviewer binary.

For this repository's self-review, distinguish the GitHub job limit, CLI whole-review timeout and per-attempt model clocks. See [Action](ACTION.md#proven-reviewing-itself). Do not repeatedly rerun a known provider quota failure without a changed condition.

## What to include in an issue

Include Proven revision/version, OS/architecture, sanitized relevant configuration, base/head identities if shareable, expected versus actual behavior, exit code and a minimal reproduction. Attach a sanitized report and the relevant trace events. Mention model and runner image identities and source revisions. Do not attach credentials, private source, raw connection URLs or unreviewed reasoning logs. Use [private vulnerability reporting](../SECURITY.md) for a security issue.
