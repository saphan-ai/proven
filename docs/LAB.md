# The judges' lab

One command builds Proven and starts a Gitea lab on this machine. The lab binds to `127.0.0.1`. Nothing in it is published on another address. The lab does not serve SSH. The only published port is the web UI.

You need GNU make and Docker. Docker Compose v2 is the `docker compose` command. Windows is not a target. On macOS the `make` from Homebrew is GNU make; the make that ships with the system is not.

Every version, port and name is a setting in `lab/versions.env`. The environment overrides the file. The list is in [CONFIGURATION.md](CONFIGURATION.md). The image tags were chosen as pins and were not pulled when this page was written. The first `make up` is the check that each tag exists. A missing tag is changed by setting the variable, not by editing the Makefile.

## Start

1. From a checkout of this repository, run `make up`.
2. The first run prints the admin user, a generated password, and the lab URL. The password is printed once. It is kept in the lab's volume until `make clean`. A later `make up` does not print it again. It does not stop because the volume is already there. It creates the admin user, the runner registration and the pull request when any of those is missing.
3. Open the URL in a browser. Sign in with that user and password.
4. Open the `discount` repository. It has one open pull request.
5. The pull request runs the workflow in `lab/workflow/proven.yml`. When the job finishes, the pull request has a comment whose body starts with `<!-- proven-review -->`. That comment is the sample review.

`make status` prints the containers, Gitea's health, whether the runner is registered, whether that comment is present, and the URL.

`make logs` follows every service. `make logs SERVICE=gitea` follows one of `gitea`, `dind`, `runner`.

`make down` stops the lab and removes its containers. The volume, the admin user and the password remain. `make clean` removes the volumes, the local images, `./bin` and `./dist`.

## What the workflow runs

Proven's Action is not on main yet, so the workflow does not use it. It runs:

```text
proven scan --sample /opt/proven/fixture/discount
```

That is the recorded sample in `testdata/fixture/discount`, baked into Proven's image. It does not review the diff of the demo pull request. Sample mode uses no network, no keys and no Docker. The comment is the text of that review. reviewdog, which will post review comments from SARIF, is not part of this lab.

The same file is what a GitHub user commits. On GitHub the job's Docker runs `saphan/proven:latest` from Docker Hub, and the job posts with `PROVEN_TOKEN` or with `GITHUB_TOKEN`. The image has to have been pushed first. `make push` does that. Docker Scout for that image is read on Docker Hub after the push. This repository does not push it for you.

The seeded copy replaces `saphan/proven` with `PROVEN_IMAGE` when that setting is not the default, and replaces `runs-on: ubuntu-latest` with `RUNNER_LABEL`.

## What `make` does besides the lab

`make` with no target prints every target, the version from `git describe`, whether `go` and `docker` were found, and the lab URL when the lab is up.

`make build` writes `./bin/proven` and `./bin/livecheck`. It uses Go when `go` is on `PATH` and the version is at least the `go` line in `go.mod`. Otherwise it builds the same binaries in `GOLANG_IMAGE` and writes them to the host. If neither Go nor Docker is available, it prints how to install Go on macOS and on Linux and exits 1.

`make build-all` cross-compiles for darwin and linux, amd64 and arm64, with `CGO_ENABLED=0`, into `./dist/`, and writes `./dist/SHA256SUMS`.

`make sample` runs the sample review from a temporary directory outside this repository. It does not pass keys and it does not call Docker.

`make livecheck` runs the live-service check. If `lab/.env` exists, it loads the keys listed in [CONFIGURATION.md](CONFIGURATION.md). With nothing set, the check names each service that is not ready and exits 1.

`make image` builds Proven's image for this machine and loads it. The load is one platform, this machine's own. A local engine cannot load a manifest list, so this target does not build both platforms. `make seed-image`, which `make up` uses, is the same kind of load: one platform, because the lab has to start the image. The image is the static `proven` and `livecheck` binaries, `git`, and the `docker` client, on `RUNTIME_IMAGE`, and it runs as a user that is not root. Local tags: `latest` and the `git describe` value. OCI labels `created`, `revision` and `version` come from the date, `git rev-parse` and `git describe`.

`make push` builds Proven's image and the seed image for `PUSH_PLATFORMS` (`linux/amd64` and `linux/arm64`) and pushes each tag as one manifest list that carries both platforms. Proven's tags are `latest` and the `git describe` value, on `PROVEN_REGISTRY/PROVEN_IMAGE`. The seed's only tag is `SEED_IMAGE_TAG`, on `PROVEN_REGISTRY/SEED_IMAGE`. The build asks for provenance (`--provenance=mode=max`) and an SBOM (`--sbom=true`). `make push` does not run Docker Scout and does not need the `docker scout` plugin.

`make scan` runs Docker Scout when that plugin is installed. It refuses when Scout reports a critical or high finding that has a fix, on the images this repository builds. The severities are `SCOUT_REFUSE_SEVERITIES`. It also scans the pinned third-party images and writes that output between the markers at the bottom of this page. When the plugin is not installed, `make scan` exits 0 and names where to read the result: on Docker Hub, after `make push`.

Read that result on the image's page on Docker Hub, for the tag `make push` wrote. Proven is `saphan/proven`, tags `latest` and the `git describe` value. The seed is `saphan/proven-lab-seed`, tag `1`, unless `SEED_IMAGE` or `SEED_IMAGE_TAG` was overridden. Docker Hub shows Docker Scout for a tag after the push. No host measured on 2026-10-23 had the local plugin. The third-party pins (`gitea/gitea`, `gitea/act_runner`, `docker`) are on their own Docker Hub pages the same way. The block at the bottom of this page is filled only when `make scan` has the plugin.

## The runner and Docker

The lab has four services: Gitea, a Docker daemon that lives only inside the lab (`dind`), the Gitea Actions runner, and a one-shot seed.

Gitea does not serve SSH (`GITEA_DISABLE_SSH`, and `GITEA_START_SSH_SERVER` is false). The image already runs OpenSSH on port 22, and Gitea's own SSH server on that same port exited the container. The lab does not need SSH. The web UI is the only published port. `GITEA_SSH_PORT` is kept in the settings and is not published.

Every `gitea` command `make up` runs inside the container uses `GITEA_RUN_USER`. That account is `git` in this image. The same command as root is refused, and `compose exec` is root when no user is named. `make down` keeps the volume, so a second `make up` finds the password already stored. It still repairs a missing admin user, a missing runner registration and a missing pull request, then starts the runner.

The runner's Docker host is `tcp://dind:2375`. No service mounts the host's Docker socket. Job containers are started on that inner daemon. The inner daemon is privileged, because Docker-in-Docker needs it. That privilege stays inside the lab. The daemon's port is not published on the host.

`make up` also pulls the Prover's test image (`PROVEN_TEST_IMAGE`, `golang:1.26` by default) into that inner daemon, by digest when `TEST_IMAGE_DIGEST` is set. Without it the lab's `make livecheck` reports `No such image: golang:1.26`. The digest in `lab/versions.env` is empty until someone with network reads it: the first `make up` pulls by tag and prints the digest to put there.

Gitea starts a workflow for a pull request only when the repository's Actions unit is on. The seed turns the unit on by name, and the Compose file adds `repo.actions` to Gitea's default units. If the unit was off when the pull request opened, the seed closes and reopens the pull request so the workflow starts. The seeded copy of the workflow posts to Gitea at `http://gitea:3000/api/v1` instead of `github.api_url`, because the job runs on the inner daemon, where `127.0.0.1` is not Gitea.

`make check-links` checks that every relative Markdown link and every path in `metadata.json` resolves.

The job image is `docker:<DOCKER_VERSION>-cli`, pinned. The runner does not pull an image it already has, so the Proven image loaded by `make up` is the one the workflow runs.

Sample mode does not start a Prover container. A later live run on this runner would start those containers on the lab's daemon, not on the host's.

## Rehearse

`make e2e` starts the lab when Gitea is not running, waits up to `E2E_WAIT_SECONDS`, and asks Gitea's API for the review comment. It exits 0 when the comment is there and prints the API's answer. It exits 1 when the wait ends without the comment.

If Gitea is already running and the runner container is not, `make e2e` does not start the runner. It exits 1 and prints:

```text
e2e: runner is not running, so the workflow cannot post a review
```

That is the check that a stopped runner cannot look like a successful rehearsal. From the repository root, stop the runner and run the rehearsal again:

```bash
docker compose --env-file lab/versions.env -f lab/compose.yaml --project-directory lab stop runner
make e2e
```

## Third-party images

These pins are not built here. `make scan` replaces the block below with Docker Scout's output. A finding in a third-party image does not refuse the scan: we cannot patch it, and a judge should see it.

<!-- scout-third-party:start -->
Not scanned in the change that added the lab. The seat that wrote this page could not run Docker Scout. `make scan` fills this block.

Images: `gitea/gitea:1.22.6`, `gitea/act_runner:0.2.11`, `docker:29.9.1-dind`, `docker:29.9.1-cli`.
<!-- scout-third-party:end -->
