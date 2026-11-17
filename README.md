# Proven

**A code reviewer that shows its evidence.** Proven checks a change against its base and reports findings backed by a differential test, a controlled data query or a deterministic tool result.

Models propose risks, tests and fixes. Deterministic rules decide which findings have evidence. Reviews can be exported as JSON, SARIF, Markdown, text or standalone HTML, with signed evidence packs for offline inspection.

Proven started at the **Open Agent Hackathon 2026** and has since grown into an independent code review tool under active development. The [original hackathon version](https://github.com/saphan-ai/proven/tree/open-agent-hackathon-2026) is preserved on a separate branch.

**Development preview.** Read [readiness and limitations](docs/READINESS.md) before using Proven as a required check. A clean result means no defect was proved under the checks that ran; it does not establish exhaustive coverage. Final unresolved planning hypotheses are reported separately and make completion incomplete.

## Start here

**[Start with Proven](docs/START.md)** — choose your CLI, build your own image, log in and paste one GitHub workflow. Covers your own runner and a GitHub-hosted runner.

**[What Proven checks and proves](docs/WHAT_PROVEN_CHECKS.md)** — evidence, coverage and limits in a table.

## Supported languages in this version

| Language | Generated tests and fix validation | Dependency checks |
| --- | --- | --- |
| **Go** | Native `go test` | Go modules |
| **Python** | pytest, including unittest-style tests | PyPI / requirements and Poetry locks |
| **TypeScript** | Vitest, including TSX | npm manifests and lockfiles |
| **JavaScript / Node.js** | Native `node:test` or Vitest | npm manifests and lockfiles |
| **PHP** | Not yet available | Composer / Packagist |
| **Rust** | Not yet available | Cargo / crates.io |
| **Kotlin** | Not yet available | Gradle Kotlin DSL, version catalogs and lockfiles / Maven |

Generated tests run against base and head in restricted containers; the selected test framework and offline dependencies must be prepared. Dependency checks cover packages added or upgraded by the change and require prepared advisory data. See [language coverage](docs/LANGUAGES.md) for supported file formats, test naming and parser limits.

## Try the recorded demo

Install the Go version in [go.mod](go.mod), then build:

```sh
go build -o bin/proven ./cmd/proven
```

Run the recorded example without accounts, model keys or a container runtime:

```sh
make sample
```

Sample mode replays recorded model answers and test outcomes. Follow the [quick start](docs/QUICKSTART.md) to produce JSON/HTML, then [walk through the evidence](docs/WALKTHROUGH.md). The [Git/SQLite tutorial](docs/DATA_WALKTHROUGH.md) runs real imports and queries without a model or Docker.

For live code review, follow [Start with Proven](docs/START.md). The [documentation map](docs/README.md) covers operations, policy, reports, extensions and troubleshooting.

## Your models and data

[See GitHub Actions, comments and expandable evidence](docs/GITHUB_TOUR.md).

![GitHub rendering of the recorded finding example with differential test evidence](docs/images/github-recorded-findings.jpg)

*Recorded sample rendered by GitHub, with synthetic data and replayed test outcomes. The [tour](docs/GITHUB_TOUR.md) separately shows real workflow/comment screenshots.*

- **Models:** an OpenAI-compatible API or a command adapter using a locally authenticated CLI. Choose your endpoint and model; credentials are optional for endpoints that need none. Legacy Nightshift and NVIDIA configurations remain opt-in adapters.
- **Evidence:** SQLite by default, with PostgreSQL, CouchDB, local fixture files and Zetaris adapters. Source tables and query templates are independent of the backend.
- **Safety checks:** an optional model service, configured separately from the reviewer. Container isolation, controlled commands and secret filtering remain enforced independently.
- **Local engines:** bounded review of large changes (Markovian), review planning (Scout Intent), persistent memory (H-MEM) and world state remain available with explicit switches. Their effect on review quality needs further evaluation.

The same adapters serve the CLI and the [GitHub Action](docs/ACTION.md). [Adapter examples](docs/ADAPTERS.md) show how to extend them.

Databases start without public feeds. See [implemented and planned sources](docs/SOURCES.md), [bounded-planning behavior](docs/MARKOVIAN.md), [container security and publication](docs/CONTAINERS.md) and the [roadmap](ROADMAP.md).

## What a finding means

The Planner identifies risks, the Investigator gathers data, the Prover runs a test against base and head, the Skeptic attempts a refutation, the Judge applies fixed evidence rules, and the Writer assembles the result. Optional fixes are checked again before being offered.

- Generated tests and tools run in restricted containers, with no network or inherited secrets.
- Queries come from registered templates rather than model-written SQL.
- The result records evidence, failed services, incomplete work and the models that answered.
- `proven verify` checks a signed pack and re-derives its recorded decisions offline. It does not re-run the recorded tests or prove that every defect was found.

See [architecture](docs/ARCHITECTURE.md), [evidence packs](docs/EVIDENCE_PACK.md), [output formats](docs/OUTPUT_FORMATS.md) and [security](SECURITY.md).

## Guides

Bring an account you already use: [API key setup](docs/CONNECT_API.md), [Claude Code / GitHub Copilot](docs/CLI_MODELS.md), [Claude + Copilot together](docs/TWO_MODELS.md). [Build your own CLI image](docs/CLI_IMAGES.md) from the supplied recipes; our published image contains no vendor CLI or credentials.

| Task | Documentation |
| --- | --- |
| Run locally or on PRs | [Install](docs/INSTALL.md), [Action](docs/ACTION.md), [operations](docs/OPERATIONS.md). |
| Bring your own model and data | [Models](docs/MODELS.md), [databases](docs/DATABASES.md), [queries](docs/QUERIES.md). |
| Understand coverage and failures | [Languages/tools](docs/LANGUAGES.md), [bounded planning](docs/MARKOVIAN.md), [diagnostics](docs/TROUBLESHOOTING.md), [traces](docs/OBSERVABILITY.md). |
| Validate fixes and results | [Fix gates](docs/FIXES.md), [formats](docs/OUTPUT_FORMATS.md), [evaluation](docs/EVALUATION.md). |
| Extend or reuse code | [Architecture](docs/ARCHITECTURE.md), [design](docs/DESIGN.md), [adapters](docs/ADAPTERS.md), [examples](examples/README.md). |

## Limitations

Evidence requirements can leave real defects unreported. Dependency-manifest support covers more languages than the generated-test adapters, which cover Go, Python/pytest, TypeScript/Vitest and JavaScript/node:test. Sample results and engine conformance scores do not establish real-world recall or precision. A missing dataset or failed analyzer is incomplete analysis, not a clean measurement.

## Development

```sh
go test ./...
go vet ./...
make check-links
```

See [CONTRIBUTING.md](CONTRIBUTING.md). Report vulnerabilities through [GitHub private vulnerability reporting](https://github.com/saphan-ai/proven/security/advisories/new).

Code: [Apache-2.0](LICENSE). Documentation: [CC BY 4.0](LICENSE-docs).
