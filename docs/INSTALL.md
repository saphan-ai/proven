# Installation

Build from a clone using the Go version in `go.mod`:

```sh
go build -o bin/proven ./cmd/proven
```

`make build-all` builds macOS and Linux binaries for amd64 and arm64. A container build is available through `docker build -t proven .`.

## Recorded example

```sh
make sample
```

This runs outside the checkout, uses recorded answers and outcomes, and needs no keys or container runtime. It is a demonstration rather than a live analysis.

## Real changes

Import your advisory snapshots into [an evidence database](DATABASES.md). Then review a change:

```sh
proven scan --baseline-commit <base-commit> /path/to/repository
```

Without model configuration, Proven performs the deterministic checks in its plan. Missing datasets and tools are reported explicitly. Add [your model endpoint or CLI](CONFIGURATION.md) to enable hypothesis generation and proposed fixes.

Generated tests require a container runtime and a prepared image containing the project's toolchain and dependencies. Tests run with no network or inherited secrets. See [security](../SECURITY.md) and [air-gapped operation](AIR_GAPPED.md).
