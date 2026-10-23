# ADR-0026: One Makefile builds Proven, builds and pushes its image, and starts, checks and stops the judges' lab

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) (the Gitea lab), [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings), [ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) (the Prover's container)
**Amends:** [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md), point 6 (Proven's own image is tagged `latest`) and the runner choice left open in its Consequences (Docker in Docker)
**Trigger:** the owner, on 23 October: one standard Makefile, the same one used in every project, for the whole judges' environment end to end

## Context

On 23 October `main` held no Compose file, no lab and no image of Proven: the lab of ADR-0017 was decided and not
built. A judge also needs a way to build Proven on a machine that may have Go, only Docker, or neither.

The owner uses one Makefile shape in every project. Run with no target, it prints its help. It builds the software,
builds images for several platforms, pushes them to a registry, starts and stops the environment, and shows its status
and its logs.

## Decision

1. **A `Makefile` at the repository root is the one entry point.** `make` with no target prints the help: every target
   with one line, the version, whether `go` and `docker` were found, and the lab's address when it runs.
2. **`make build` builds the native binaries** (`proven`, `livecheck`) into `./bin/`, choosing the first path that works:
   - Go is installed at the version `go.mod` asks for or later: `go build`;
   - no Go, but Docker: the same build inside a pinned `golang` image, with the binaries written to the host;
   - neither: it prints how to install Go on macOS and on Linux, and exits 1.
3. **`make build-all` cross-compiles** for darwin and linux, amd64 and arm64, into `./dist/` with checksums.
4. **`make image` builds Proven's image for the local platform; `make push` builds it for several platforms** (by
   default linux/amd64 and linux/arm64) with provenance and an SBOM, and pushes `saphan/proven` to Docker Hub. The
   registry and the image name are settings.
5. **Proven's own image is tagged `latest`** (the owner, 23 October). Each push also carries the tag `git describe`
   gives, so the image a judge ran on a given day can still be pulled. Every other image in the lab (Gitea, its
   runner, `golang`, Docker) stays pinned to a version, as ADR-0017 point 6 says.
6. **`make up` starts the lab of ADR-0017; `make down` stops it.** One Compose file: Gitea, a Gitea Actions runner, and a
   one-time seed that creates the admin (password generated and printed once), the demo repository with an open pull
   request and Proven's workflow, and registers the runner. Everything binds to localhost.
7. **The runner starts the Prover's containers in Docker in Docker** inside the lab, never through the host's Docker
   socket (the owner, 23 October). The lab's jobs get no control of the judge's Docker.
8. **`make status`** shows the containers, Gitea's health, whether the runner is registered, the last workflow result on
   the pull request and the lab's address. **`make logs`** follows the logs (`SERVICE=` for one service).
9. **`make sample`** runs the sample-mode review from a directory outside the repository, with no network, no keys and
   no Docker. **`make livecheck`** names which live services are ready, reading an optional `lab/.env`.
10. **`make e2e` is the judge's rehearsal:** it starts the lab, waits for the workflow, asks Gitea's API whether
    Proven's review is on the pull request, and exits 0 or 1. It is the daily rehearsal run before each push.
11. **`make clean`** removes the lab's containers, volumes and local images, and `./bin` and `./dist`.
12. **Every version, port and name is a setting** with its default in one file, `lab/versions.env`, which the
    environment can override, and each one is listed in `docs/CONFIGURATION.md`. Keys live only in `lab/.env`, which
    is never committed; the repository holds `lab/.env.example`.
13. **The judge's runbook is `docs/LAB.md`**, and the README's quick start begins with `make up`.

## Consequences

- A judge with only Docker gets the binaries, the lab and the review on a pull request, without installing Go.
- `make e2e` turns ADR-0017 point 7 (the lab is rehearsed the way a judge runs it) into one command that fails loudly.
- Docker in Docker needs a privileged container inside the lab. That is a cost on the judge's machine, paid instead of
  handing the lab's jobs the host's Docker.
- `latest` moves with every push: a judge who pulls on a later day gets a later Proven. The dated tag is how a given
  day's image is pulled again.
- Windows is not a target: `make` and the scripts assume macOS or Linux.
