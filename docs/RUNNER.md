# Restricted runner

The runner turns a proposed test into an observed base/head comparison. It executes commands built by Proven inside restricted containers. A model supplies test source, not a shell command. The selected runtime, daemon, image and toolchain are trusted operator dependencies.

Generated-test proving currently targets Go. Reading another ecosystem's dependency file does not add that ecosystem's test runner. See [language support](LANGUAGES.md).

## Product image versus runner image

The **product image** contains the Proven CLI. The **runner image** contains the target project's compiler, tools and offline dependencies. They have separate versions and security responsibilities. Publishing a Proven image does not prepare every customer's test environment.

The default runner reference is `golang:1.26`. It must already exist locally: every run uses `--pull never`. The default is a starting toolchain, not a reproducible image pin or a populated dependency cache. Select an immutable prepared image for operational use.

## Isolation that always applies

| Boundary | Runner behavior |
| --- | --- |
| Network | `--network none`; no fetching modules, feeds or credentials during execution. |
| Source | One checkout mounted read-only at `/checkout`; generated/modified source is prepared in scratch copies before the run. |
| Root filesystem | Read-only. |
| Writable scratch | Bounded `/tmp` tmpfs with `nosuid,nodev`; executable because Go builds and runs test binaries there. |
| Identity | Numeric non-root UID/GID; default `65534:65534`. |
| Privileges | All capabilities dropped; `no-new-privileges`. |
| Resources | CPU, memory, process count, scratch and timeout limits. |
| Environment | Fixed `HOME=/tmp` and `GOCACHE=/tmp/go-build`; no inherited model/database secrets. The trusted image has its own environment. |
| Command | Explicit argv constructed by the adapter; no model-provided shell. |

There is no setting that turns off network isolation or mounts a secret into the test container. Host runtime access is privileged infrastructure; a container boundary is not a formal guarantee against a vulnerable kernel, daemon or toolchain. See [Security](../SECURITY.md).

## Defaults and settings

| Variable | Default | Purpose |
| --- | --- | --- |
| `PROVEN_TEST_RUNTIME` | `docker` | Installed container-runtime executable. |
| `PROVEN_TEST_IMAGE` | `golang:1.26` | Prepared local image. |
| `PROVEN_TEST_CPUS` | `1` | CPU limit per container. |
| `PROVEN_TEST_MEMORY` | `512m` | Memory limit per container. |
| `PROVEN_TEST_USER` | `65534:65534` | Non-root numeric identity. |
| `PROVEN_TEST_PIDS` | `256` | Maximum process count. |
| `PROVEN_TEST_TMP_SIZE` | `512m` | `/tmp` size bound; maximum 1 GiB. |
| `PROVEN_TEST_TIMEOUT` | `2m` | Time for one test side, not the entire review. |
| `PROVEN_TEST_MAX_WRITTEN_BYTES` | `65536` | Largest generated test file, in bytes. |

Invalid settings are refused by name before a container starts. Cleanup has a separate limit of 2 seconds within an overall margin of 3 seconds. Failed cleanup names a container that may remain, so an operator can investigate it. See the [generated settings reference](CONFIGURATION.md#the-provers-container) for accepted values.

## Prepare the target project's dependencies

For a Go module, create an image using that module's `go.mod` and `go.sum`. Resolve dependencies during image build in a trusted preparation environment, not during proof. An illustrative Dockerfile is:

```dockerfile
FROM golang:1.26.10
WORKDIR /deps
COPY go.mod go.sum ./
RUN go mod download
ENV GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0
```

Save it as `Dockerfile.proven-runner`, then build from the target module's directory:

```sh
docker build -f Dockerfile.proven-runner -t project-proven-runner:local .
export PROVEN_TEST_IMAGE=project-proven-runner:local
```

Pin the base digest and resolve private modules according to your build policy; do not bake authentication into image layers. Ensure the unprivileged runner can read the prepared cache. For modules with no external dependencies, a pre-pulled compatible toolchain image may suffice.

The repository's `Dockerfile.runner` prepares **Proven's own module**. Building it from the Proven checkout does not populate dependencies for an unrelated target. A dependency change may require a cache containing both base and head requirements. Keep preparation separate from the untrusted generated test.

Vendored dependencies can provide another offline path. Verify the selected method under the runner's actual non-root, read-only, network-free constraints; a successful ordinary host build is not equivalent.

## Differential proof

The same generated test is run against isolated base and head snapshots. A finding needs the required differential outcome: pass on base, fail on head. Compilation failure, timeout, cancellation and a test that did not run do not establish that outcome. A failure on both sides is not proof that the change introduced the defect.

The report retains the test identity, command and observed outcome. The Skeptic may try to refute it, but an unsupported model opinion cannot erase an independent observed proof. Fix validation adds patched-head and package checks; see [Fixes](FIXES.md).

## Deterministic tools

Tools use controlled execution too. The race adapter needs a toolchain/environment that can run the race detector, including compatible CGO/compiler support. The illustrative image above sets `CGO_ENABLED=0`, so it is not a ready-made race image.

The govulncheck adapter requires a prepared executable and a local vulnerability database. `PROVEN_GOVULNCHECK_DB` accepts a file URL under the checkout mount, for example `file:///checkout/vulndb`; it does not download a public database at proof time. Empty means that measurement is not run. `PROVEN_TOOL_TIMEOUT` independently bounds one side of a deterministic tool run, default 2 minutes. See [tool adapters](ADAPTERS.md#adding-a-tool).

## Running Proven in a container

If Proven talks to the host Docker daemon, bind paths are resolved on the daemon's host. A source or temporary directory visible only inside the Proven container is not automatically visible to that daemon. Installing a Docker CLI or mounting its socket does not solve the path mismatch.

Prefer a host CLI for generated-test proving unless the deployment deliberately provides matching host/container paths for all snapshots. This repository's self-review extracts the CLI from a pinned product image and runs it on the GitHub host for that reason. [Action](ACTION.md#proven-reviewing-itself) describes that arrangement. Avoid exposing the host socket to untrusted target code.

## Validate and diagnose

First run a small known differential defect with your actual image, base/head and limits. Confirm the proof ran rather than reading only an exit code. Then test a corrected head and an unavailable-image case. Sample mode records runner outcomes and does not test this environment.

For no-run outcomes, inspect missing images, dependency cache, toolchain version, UID permissions, bind paths, memory/scratch limits and the recorded command. Increase limits based on observed resource needs while keeping bounds. See [Troubleshooting](TROUBLESHOOTING.md#generated-tests-did-not-run) and [container security](CONTAINERS.md).
