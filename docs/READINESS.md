# Readiness and limitations

Proven is a development preview. Build and fixture tests establish implemented behavior; they do not establish production readiness or real-world defect-detection accuracy. Use reports as review evidence and inspect gaps before making an enforcement decision.

## What is available

| Area | Available behavior | Practical boundary |
| --- | --- | --- |
| Model transport | Streaming OpenAI-compatible API, optional local authentication, operator-selected CLI, recorded samples. Configurable limits, clocks, provider options and recovery. | Compatibility depends on endpoint/model options. A passing `model check` is one short request, not a review-quality benchmark. Vendor subscription CLI flows need their own configuration and validation. |
| Evidence stores | SQLite, PostgreSQL, CouchDB, local fixtures and legacy Zetaris adapters. | A backend does not fetch sources. SQLite/PostgreSQL accept table imports; live deployment coverage varies by backend. |
| Data sources | Seven adapters, public acquisition for selected dependency changes, advisory delta/conditional caches, source-check observations and table imports. | Provider coverage varies; refresh can be incomplete and observation dates can be older than retrieval. Automatic imports use SQLite/PostgreSQL. See [source status](SOURCES.md). |
| Proving | Differential base/head tests and controlled tool results. | Generated-test adapters cover Go, Python/pytest, TypeScript/Vitest and JavaScript/node:test. Wider dependency-manifest support is not equivalent to generated-test support for every language. A real runner needs a configured container runtime. |
| Markovian | Bounded planning with explicit final unresolved hypotheses; identified risks continue to verification. | Open hypotheses make coverage incomplete. Mid-plan carry overflow remains a refusal. See [bounded planning](MARKOVIAN.md). |
| Other local engines | Scout Intent, H-MEM and World Model remain configurable. | Conformance and fixture results do not show improved review quality. Persistent state must be kept outside the reviewed repository. |
| Reports and evidence packs | Text, JSON, SARIF, Markdown and HTML; signed packs and offline decision re-derivation. | A signature identifies recorded bytes and decisions under implemented rules. It does not rerun tests or establish that every defect was found. |

## Reading completion correctly

Findings and completion are separate. No findings can mean no defect was proved, not that no defect exists. A completed run can still miss real defects. Unavailable required sources, failed services, remaining planning hypotheses or a deadline can make analysis incomplete.

`completion_status` is `complete` or `incomplete`. `partial` specifically records interrupted execution, so an unresolved final hypothesis can produce `incomplete` with `partial: false`. `unresolved_hypotheses` are separate from findings, proof counts and the Action's existing unresolved-finding count.

| Exit | Meaning |
| --- | --- |
| `0` | Completed invocation without a blocking policy outcome; an advisory scan can still describe incomplete analysis. Read the report. |
| `1` | Findings under `--error` / strict mode, when required-completion failure does not take precedence. |
| `2` | Fatal execution/refusal, such as a carried-state overflow; a complete result file may not exist. Inspect stderr and any trace. |
| `3` | Required completion was not achieved; the result reports the gaps. |
| `7` | Configuration refused before review. |

CI and the GitHub Action require completion by default. Local `scan` is advisory unless configured otherwise. Turning off required completion changes enforcement, not the report's assessment. See [configuration](CONFIGURATION.md#completion-policy) and [output formats](OUTPUT_FORMATS.md).

## What has been measured

On 16 November 2026, one compatibility request per option was made on Nemotron 3.5 Lightning and Super. Three provider-body variants returned HTTP 200 and valid nonempty probe JSON on both models. This establishes compatibility for those calls, not support by other models or unchanged future behavior.

An earlier whole-change self-review completed in 63 seconds for the review step: 17 successful HTTP requests, five hypotheses dropped for lack of evidence, no failed services and no findings. It used thinking disabled for Planner/Investigator/Skeptic, nonempty-content options for Prover/Fixer, and Markovian off. [Run and retained evidence](https://github.com/saphan-ai/proven/actions/runs/43817450372). That run predates the final-hypothesis handling change and does not validate the current bounded mode. It also does not establish detection accuracy; the five Prover responses abstained from supplying a differential test.

The repository workflow profile is an operator choice, not a universal model preset. Published binaries are pinned by image digest in self-review, so changing source without rebuilding and selecting an image does not update that reviewer. See [Action operations](ACTION.md#proven-reviewing-itself) and [container builds](CONTAINERS.md).

The later [bounded self-review run](https://github.com/saphan-ai/proven/actions/runs/43826723846) used the image containing final-hypothesis handling, Markovian on, a 32,768-token Prover allowance and longer rate-limit backoff. It finished 67 planning steps and checked 11 candidates. The trace contains 106 HTTP attempts, including six 429 responses recovered by retry. One Prover response still ended with `finish_reason=length`. The written report preserved a final pending hypothesis, reported incomplete coverage with `partial: false`, and the required check exited 3. This verifies preservation of useful work and visible gaps; it is not a successful complete review or an accuracy benchmark. Raising the token limit alone did not eliminate that model's truncation behavior.

## Evaluation still needed

The [next bounded run](https://github.com/saphan-ai/proven/actions/runs/43832549285), with thinking disabled for all roles, completed 87 planning steps with no failed services. Its trace has 155 HTTP attempts: 142 successful responses and 13 rate-limit responses recovered by retry. It wrote reports with zero proved findings and one final pending hypothesis, so required completion still exited 3. The observed response truncation was absent in this run; review quality and resolution of every hypothesis were not established.

Track detection and misses against controlled defects, changes without defects, supported languages and source coverage. Compare whole-change and bounded planning with the same cases and document model/options, dataset revisions, runtime image and limits. Stronger recall claims, production backend guarantees and broad CLI subscription compatibility remain unestablished. See [ROADMAP.md](../ROADMAP.md).
