# Install and first run

The judges' lab starts with `make up`. The runbook is [LAB.md](LAB.md). The steps below are the Go build and the sample review without the lab.

The quick start: build Proven and run a first review with no network and no keys. Every command below was run on
22 October 2026 on a clean clone of `main` (`16e269a`), with Go 1.26.8 on macOS.

## Prerequisites

| Needed | Version | For |
|---|---|---|
| Go | 1.26 or later (`go.mod` says `go 1.26`) | building and running Proven |
| git | any recent | cloning; a live review of a git range reads the repository with git |
| Docker (or another container runtime) | any recent | **live mode and the container image** (`make image`, `make image-check`); in live mode the Prover runs each test in a container with no network and no secret. Without a runtime the Prover is off, and the review names it as a failed service |

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

## The container image

This needs Docker (with buildx) as well as Go. `make image` builds Proven's image for this machine. `make image-check` builds
the binary and the image and checks that they
print the same version line and the same sample review, then prints the image's size; it fails and names the first difference.
The same check by hand, from the repository, after `make build` and `make image`:

```bash
cd "$(git rev-parse --show-toplevel)"
sh scripts/image-check.sh saphan/proven:latest ./bin/proven
```

In a git checkout the image must carry this checkout's commit as its `org.opencontainers.image.revision` label, or the check
fails and names both commits; the binary is compared by its output only.

Inside the image the sample fixture is at `/opt/proven/fixture/discount`:

```bash
docker run --rm saphan/proven:latest scan --sample /opt/proven/fixture/discount
```

## Before a live review: livecheck

A live review needs only a git repository.

From the repository:

```bash
go run ./cmd/livecheck
```

With nothing configured it names every service that is not ready, and why, and exits 1. An empty secret makes that service down and named, and the review continues; a wrong non-empty value is still refused, and an empty guardrail key refuses every model call unless the mode is off.

```text
generator: not ready: nightshift: METERLESS_RUNTIME_TOKEN is empty
guardrail: not ready: guardrail: PROVEN_GUARDRAIL_KEY is empty
query: not ready: zetaris: ZETARIS_REST_URL is empty
container: not ready: prover: Prover off: container runtime "docker" not found: …
summary: 0 ready, 4 not ready
```

The Prover needs the test image on the host before a review; it never pulls one during a review. When livecheck says
the image is not present, its line ends with the act to run, for example:

```bash
docker pull golang:1.26
```

Each setting it names, with its default: [CONFIGURATION.md](CONFIGURATION.md). A live review a judge can rerun, with one
finding proved by a test and one by Zetaris: [demo/live/README.md](../demo/live/README.md).

The command-line flags beyond `scan --sample` are documented by the shell's track.
