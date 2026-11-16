# Proven

**A code reviewer that shows its evidence.** Proven checks a change against its base and reports findings backed by a differential test, a controlled data query or a deterministic tool result.

Models propose risks, tests and fixes. Deterministic rules decide which findings have evidence. Reviews can be exported as JSON, SARIF, Markdown, text or standalone HTML, with signed evidence packs for offline inspection.

**Development preview.** Read [readiness and limitations](docs/READINESS.md) before using Proven as a required check. A clean result means no defect was proved under the checks that ran; it does not establish exhaustive coverage. Final unresolved planning hypotheses are reported separately and make completion incomplete.

## Try it

Install the Go version in [go.mod](go.mod), then build:

```sh
go build -o bin/proven ./cmd/proven
```

Run the recorded example without accounts, model keys or a container runtime:

```sh
make sample
```

Sample mode replays recorded model answers and test outcomes. For a real change, configure an evidence database and any model or analyzer you want to use. See [installation](docs/INSTALL.md), [configuration](docs/CONFIGURATION.md) and [sample mode](docs/SAMPLE_MODE.md).

## Your models and data

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

## Limitations

Evidence requirements can leave real defects unreported. Dependency-manifest support covers more languages than the generated-test runner, which currently targets Go. Sample results and engine conformance scores do not establish real-world recall or precision. A missing dataset or failed analyzer is incomplete analysis, not a clean measurement.

## Development

```sh
go test ./...
go vet ./...
make check-links
```

See [CONTRIBUTING.md](CONTRIBUTING.md). Report vulnerabilities through [GitHub private vulnerability reporting](https://github.com/saphan-ai/proven/security/advisories/new).

Code: [Apache-2.0](LICENSE). Documentation: [CC BY 4.0](LICENSE-docs).
