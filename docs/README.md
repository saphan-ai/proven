# Proven documentation

Proven reviews a Git change and shows the evidence behind its findings. Start with a recorded review, then choose your own model, data store and execution environment. This documentation describes delivered behavior and separates it from planned work.

## Start here

**[Start without a model](MODEL_FREE.md)**: run the generic image with public-data and description checks. **[Connect your CLI](START.md)**: build your image, log in and paste one workflow. **[What it checks and proves](WHAT_PROVEN_CHECKS.md)**: evidence and limits in a table.

[CLI supply-chain case study](CLI_SUPPLY_CHAIN_AUDIT.md): measured Qwen/Cursor dependency findings, explicit fixes and retained before/after scan outputs.

[Declared scope, source by source](PR_DELIVERY.md): keep PR descriptions, commit messages and each comment separate; check descriptions without a model and cited implementation with one.

[Source refresh](SOURCE_REFRESH.md): acquire public data before a run, reuse unchanged records and report provider/observation dates.

[See it on GitHub](GITHUB_TOUR.md): actual Actions and PR-comment screenshots, plus a labelled recorded finding with expandable proof.

| You want to… | Read | Result |
| --- | --- | --- |
| Run without accounts | [Quick start](QUICKSTART.md) | Build the CLI, produce JSON/HTML and understand exit codes. |
| Understand the evidence | [Review walkthrough](WALKTHROUGH.md) | Follow a claim through evidence, refutation and Judge. |
| Try real Git and SQLite paths | [Data walkthrough](DATA_WALKTHROUGH.md) | Import teaching data, test a finding, fixed boundary and missing-source failure. |
| Prepare a live installation | [Installation](INSTALL.md) | Identify which components each capability needs. |
| Assess suitability | [Readiness and limitations](READINESS.md) | Know what is available, measured and still unestablished. |

[Code comments](CODE_COMMENTS.md): Go AST structure checks without a model, optional language/code-consistency assessment, separate results and adapter coverage.

[For language models](../llms.txt): product facts and guide links. The documentation build generates `/llms.txt` and `/llms-full.txt` from every rendered page.

## Configure and operate

Connection recipes: [own API key](CONNECT_API.md), [Claude Code or Copilot](CLI_MODELS.md), [two reviewers](TWO_MODELS.md), [build your own CLI image](CLI_IMAGES.md). Vendor CLIs and account state are customer-provisioned.

| Guide | Covers |
| --- | --- |
| [Optional guardrails](GUARDRAILS.md) | Prompt/answer safety checks, NVIDIA or self-hosted service, and runtime security boundaries. |
| [Models](MODELS.md) | APIs, CLI/subscription wrappers, roles, reasoning, clocks, tokens, recovery and second reviewers. |
| [Configuration reference](CONFIGURATION.md) | Generated environment catalogue, repository keys and defaults. |
| [Policy](POLICY.md) | Organization policy, class floors, doubling and scoped exceptions. |
| [GitHub Action](ACTION.md) | PR integration, permissions, artifacts, comments and self-review. |
| [Restricted runner](RUNNER.md) | Product versus runner images, offline dependencies, isolation and host paths. |
| [Operations](OPERATIONS.md) | Rollout, state, secrets, freshness, enforcement, upgrades and retention. |
| [Troubleshooting](TROUBLESHOOTING.md) | Exit codes, model limits, missing data, no-run outcomes and rejected fixes. |
| [Observability](OBSERVABILITY.md) | JSONL, model-attempt measurements, guardrail accounting and OTLP. |
| [Local and offline operation](AIR_GAPPED.md) | Network policy and preparation before disconnecting. |
| [Containers](CONTAINERS.md) | Patched Alpine, architectures, security scans and publication gate. |

## Data and coverage

- [Sources and planned connectors](SOURCES.md): implemented adapters, acquisition boundaries and future connections.
- [Evidence databases](DATABASES.md): import contract, identity and backend choices.
- [Query catalogue](QUERIES.md): registered names, parameters, tables and evidence rules.
- [OSV schema and version rules](OSV_TABLE.md): normalization and precise ecosystem subsets.
- [Languages and tools](LANGUAGES.md): dependency readers versus actual generated-test/tool support.
- [Bounded planning](MARKOVIAN.md): carried state, pending hypotheses, overflow and completion.

## Read and validate results

- [Output formats](OUTPUT_FORMATS.md): text, JSON, SARIF, Markdown, standalone HTML and completion.
- [Evidence packs](EVIDENCE_PACK.md): signing, verification, custody and limits of recorded evidence.
- [Fix validation](FIXES.md): suggestions, proof gates and whole-change autofix.
- [Sample catalogue](SAMPLE_MODE.md): recorded findings, refutations, fixes, languages and fixture format.
- [Evaluation](EVALUATION.md): controlled cases, comparisons, misses and incomplete runs.

## Understand and extend

- [Architecture](ARCHITECTURE.md): stages, trust boundaries, plans and reusable packages.
- [Design decisions](DESIGN.md): reasons and tradeoffs behind evidence, adapters and completion.
- [Adapter guide](ADAPTERS.md): worked examples for models, languages, tools, sources, formats, guardrails, reviewers and runners.
- [Runnable examples](../examples/README.md): small implementations with tests and synthetic import data.
- [Documentation maintenance](DOCUMENTATION.md): source pages, generated references and site checks.
- [Contributing](../CONTRIBUTING.md), [security](../SECURITY.md) and [roadmap](../ROADMAP.md).

Code is [Apache-2.0](../LICENSE); documentation is [CC BY 4.0](../LICENSE-docs). The CLI and repository pages remain useful without the optional documentation site.
