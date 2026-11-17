# Installation

Build Proven without sponsor accounts, a database server or a model key. Start with [a real model-free dependency review](MODEL_FREE.md), or use the recorded sample without network access. Models and project test runtimes are separate choices.

## Requirements by task

| Task | Required |
| --- | --- |
| Build CLI | Go version in `go.mod`, module dependencies, Git and a shell. |
| Recorded sample | Built CLI and repository fixture files. No model or container runtime. |
| Live Git comparison | Local clone containing base/head, external output/state paths. |
| Dependency evidence | `--refresh-data` with Internet access, or imported source tables. SQLite needs no server. |
| Model hypotheses | Compatible streaming API or text-only command adapter. |
| Generated tests / Go tools | Container runtime, prepared local runner image, dependencies and tools. |
| Documentation site | Pinned Node dependencies and Go for generated-reference checks. |

## Build from source

```sh
git clone https://github.com/saphan-ai/proven.git
cd proven
go build -trimpath -o bin/proven ./cmd/proven
./bin/proven --version
./bin/proven --help
```

The standard build uses Go's module download configuration. To build offline, populate the module cache first and select an installed compatible toolchain. A previously successful online build is not an air-gapped dependency bundle.

`make build` builds the CLI. `make build-all` produces macOS and Linux binaries for amd64 and arm64. The cross-build script is a build check, not a signing/notarization or package-manager installation service. Keep the executable at an explicit path or put it on your `PATH` according to your local conventions.

## Recorded example

```sh
make sample
```

The helper runs outside the checkout so state is not written into reviewed source. Follow [Quick start](QUICKSTART.md) for JSON/HTML outputs and exit codes. The [sample catalogue](SAMPLE_MODE.md) explains recorded cases.

Next, [review a real Git comparison with synthetic SQLite data](DATA_WALKTHROUGH.md). This exercises actual imports, parsing and queries without a model or Docker. Replace the synthetic data before operational use.

## Product container

```sh
docker build -t proven:local .
docker run --rm proven:local --version
```

Use [Containers](CONTAINERS.md) for architectures, patched runtime packages, equivalence checks, scans and publication. Choose a verified immutable digest for CI. The product image contains the CLI, not every target project's compiler/cache; [Runner](RUNNER.md) describes that separate image and host bind-path requirements.

For live generated tests, prefer the host CLI unless matching daemon-visible snapshot paths have been deliberately arranged. A mounted Docker socket alone does not supply them.

## Real changes

1. Use [automatic source refresh](SOURCE_REFRESH.md), or import your own [evidence data](DATABASES.md). Databases start without public feeds.
2. Configure an optional [model](MODELS.md) and run its one-request API probe.
3. Prepare the [restricted runner](RUNNER.md) if generated tests or tools are required.
4. Select a base commit present in the target clone and run:

```sh
PROVEN_GENERATOR_MODE=off proven scan --refresh-data --baseline-commit <base-commit> /path/to/repository
```

Omit `PROVEN_GENERATOR_MODE=off` after configuring a model. Omit `--refresh-data` when reviewing a previously imported snapshot. Refresh makes complete acquisition required by default; [completion policy](CONFIGURATION.md#completion-policy) remains separate from finding enforcement. Use [Operations](OPERATIONS.md) before enabling a required team check, and [Action](ACTION.md) for GitHub workflows.

## Verify a local installation

Check version/help, recorded sample and a controlled real comparison. Confirm expected data identities, enabled stages and completion. A model probe does not test Docker; a sample does not test live services. Source-checkout checks are described in [CONTRIBUTING.md](../CONTRIBUTING.md).

See [Troubleshooting](TROUBLESHOOTING.md) for unavailable services/images and [Air-gapped operation](AIR_GAPPED.md) before constraining network access.
