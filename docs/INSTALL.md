# Install and first run

The quick start: build Proven and run a first review with no network and no keys. Every command below was run on
22 October 2026 on a clean clone of `main` (`16e269a`), with Go 1.26.8 on macOS.

## Prerequisites

| Needed | Version | For |
|---|---|---|
| Go | 1.26 or later (`go.mod` says `go 1.26`) | building and running Proven |
| git | any recent | cloning; a live review of a git range reads the repository with git |
| Docker (or another container runtime) | any recent | **live mode only**: the Prover runs each test in a container with no network and no secret. Without a runtime the Prover is off, and the review names it as a failed service |

Sample mode needs only Go.

## Build

```bash
git clone https://github.com/saphan-ai/proven.git
cd proven
go build -o /tmp/proven ./cmd/proven
```

## First review: sample mode

Run Proven from a directory outside the repository: it keeps its memory in `./state` and its trace in `./logs`, and it
refuses to keep its memory inside the repository it reviews.

```bash
mkdir -p /tmp/proven-run && cd /tmp/proven-run
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/discount"
```

It prints three findings (F1 proved by a test, F2 and F4 by a query and its row), drops F3 for lack of evidence, and
exits 0. The full output and what each line means: [SAMPLE_MODE.md](SAMPLE_MODE.md).

## Before a live review: livecheck

From the repository:

```bash
go run ./cmd/livecheck
```

With nothing configured it names every service that is not ready, and why, and exits 1:

```text
generator: not ready: configuration refused: METERLESS_RUNTIME_TOKEN="": the live generator needs a runtime token; …
guardrail: not ready: configuration refused: PROVEN_GUARDRAIL_KEY="": the guardrail needs a key; …
query: not ready: configuration refused: ZETARIS_REST_URL="": the live query adapter needs a base URL; …
container: not ready: prover: Prover off: container runtime "docker" not found: …
summary: 0 ready, 4 not ready
```

Each setting it names, with its default: [CONFIGURATION.md](CONFIGURATION.md). A live review a judge can rerun, with one
finding proved by a test and one by Zetaris: [demo/live/README.md](../demo/live/README.md).

The command-line flags beyond `scan --sample` are documented by the shell's track.
