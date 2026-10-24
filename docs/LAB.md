# The judges' lab

One command builds Proven and starts a Gitea lab on this machine. The lab binds to `127.0.0.1`. Nothing in it is published on another address: `make up` refuses a `LAB_BIND` that is not a loopback address (`127.0.0.0/8` or `[::1]`), names the value, and starts no container. The lab does not serve SSH. The only published port is the web UI.

You need GNU make and Docker. Docker Compose v2 is the `docker compose` command. Windows is not a target. On macOS the `make` from Homebrew is GNU make; the make that ships with the system is not.

Every version, port and name is a setting in `lab/versions.env`. The environment overrides the file. The list is in [CONFIGURATION.md](CONFIGURATION.md). The image tags were chosen as pins and were not pulled when this page was written. The first `make up` is the check that each tag exists. A missing tag is changed by setting the variable, not by editing the Makefile.

## Start

1. From a checkout of this repository, run `make up`.
2. The first run prints the admin user, a generated password, and the lab URL. The password is printed once. It is kept in the lab's volume until `make clean`. A later `make up` does not print it again. It does not stop because the volume is already there. It creates the admin user, the runner registration and the pull request when any of those is missing.
3. Open the URL in a browser. Sign in with that user and password.
4. Open the `discount` repository. It has one open pull request.
5. The pull request runs the workflow in `lab/workflow/proven.yml`. When the job finishes, the pull request has a comment whose body starts with `<!-- proven-review run=<run number> sha=<head commit> -->`. That comment is the sample review, and the marker names the run that posted it.

`make status` prints the containers, Gitea's health, whether the runner is registered, whether that comment is present, and the URL.

`make logs` follows every service. `make logs SERVICE=gitea` follows one of `gitea`, `dind`, `runner`.

`make down` stops the lab and removes its containers. The volume, the admin user and the password remain. `make clean` removes the volumes, the local images, `./bin` and `./dist`.

## The keys a live review needs

This page says which keys the lab needs for each way of running the model generator, and where to get each one
([ADR-0031](adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md) point 6). It covers keys only. Every setting
and its default is in [CONFIGURATION.md](CONFIGURATION.md). Sample mode needs no key and makes no call:
[SAMPLE_MODE.md](SAMPLE_MODE.md).

### With one NVIDIA key (no Nightshift needed)

A free NVIDIA API key is enough to see Proven's agents work on live models.

1. Go to <https://build.nvidia.com> and sign in.
2. Open any model and choose *Get API Key*. Copy the key; it is shown once.
3. Put the key in the environment, never in a file that is committed:

   ```
   PROVEN_GENERATOR_MODE=nvidia
   PROVEN_GENERATOR_KEY=<the key>
   PROVEN_GUARDRAIL_MODE=nvidia
   PROVEN_GUARDRAIL_KEY=<the same key, or another>
   ```

4. Check it before a review with `livecheck`. Its generator line names the mode, and a missing or rejected key.

The address and the models have defaults, so nothing else is needed. In this mode the address is `PROVEN_NVIDIA_BASE_URL`
(`https://` only; any other value is refused by name). `PROVEN_GENERATOR_BASE_URL` is not read here, so a Nightshift
address left in the environment never receives the key. `livecheck` names the host the answer came from. The hosted API signs no receipt, so the trace says
*no receipt* for each call, and the requirement that an agent be answered by an `nvidia/*` model is checked against the
`model` field of the answer. A free key is rate limited; a review that meets the limit names it as a failed service.

### With Meterless Nightshift (the default)

Set `METERLESS_RUNTIME_TOKEN` to the runtime token of the Nightshift you run, and `PROVEN_GENERATOR_BASE_URL` to its
address (this variable is read in `nightshift` mode only). `PROVEN_GENERATOR_MODE` is left empty or set to `nightshift`. Nightshift is the team's own service: a judge
without one uses the NVIDIA key above.

### Keys of the other services

| Service | Key | Where it comes from | If it is empty |
|---|---|---|---|
| Guardrail | `PROVEN_GUARDRAIL_KEY` | the same NVIDIA key page, when `PROVEN_GUARDRAIL_MODE=nvidia` | every model call is refused and the review names the guardrail |
| Zetaris evidence | see [CONFIGURATION.md](CONFIGURATION.md) | the Zetaris account | Zetaris is down and the review names it; the rest runs |

### A second reviewer: one more key

The lab runs with NVIDIA alone ([ADR-0037](adr/ADR-0037-bring-your-own-second-reviewer.md) point 6). Every change
gets one independent review, from the default reviewer, and nothing else is needed.

Proven can review every change with two models. The second one is any model behind an OpenAI-compatible chat API:
OpenAI, Anthropic through its compatible endpoint, or a server on your own machine. You switch it on with three
settings in the environment Proven runs in, and the key is the switch:

| Variable | Value |
|---|---|
| `PROVEN_SECOND_REVIEWER_KEY` | the API key. Secret: never printed, never written to the trace or the result |
| `PROVEN_SECOND_REVIEWER_BASE_URL` | the API address, for example `https://api.openai.com/v1` or `http://localhost:11434/v1` |
| `PROVEN_SECOND_REVIEWER_MODEL` | the model name the API knows, for example `gpt-5` |

Add the three variables and run the lab as before. Take the key away and the lab is back to one reviewer. Nothing
else changes.

#### What you see once it is connected

- **Both reviewers look at every change.** They work independently: neither is shown the other's answer. A change
  under review cannot narrow this for itself: `proven.yaml` is read from the baseline.
- **Where they differ, the review says so.** Each finding carries both reviewers' verdicts and reasons, and the
  review counts the disagreements; the two are kept side by side and never averaged. The block *The reviewers
  disagree*, which shows them in the text, Markdown, HTML and pull-request comment output, and the verdicts in the
  JSON and SARIF output, arrive with the shell train (see [ARCHITECTURE.md](ARCHITECTURE.md)); the command line does
  not render them yet. Evidence still decides what is reported: a risk neither reviewer can prove is not a finding,
  however many models name it.
- **Every review names its reviewers**, and whether either of them is the model that wrote the change.

#### What it costs

Two reviewers cost two model calls per review step. To spend less, name the classes the second reviewer covers in
`proven.yaml`, for example `doubled_classes: security,backend`. The security class is always doubled when a second
reviewer is connected. See [CONFIGURATION.md](CONFIGURATION.md) for the settings and their refusals: a key with no
address or no model, or an address that is not an `http` or `https` URL, stops Proven before any review starts and
names the variable.

#### If the second reviewer is down

A review whose second reviewer fails still completes with the first reviewer's findings. The failure is listed
under `second-reviewer` in the review's failed services, and the findings show the second reviewer's verdict as
`no answer`, which is not a disagreement.

## What the workflow runs

Proven's Action is not on main yet, so the workflow does not use it. It runs:

```text
proven scan --sample /opt/proven/fixture/discount
```

That is the recorded sample in `testdata/fixture/discount`, baked into Proven's image. It does not review the diff of the demo pull request. Sample mode uses no network, no keys and no Docker. The comment is the text of that review. reviewdog, which will post review comments from SARIF, is not part of this lab.

The same file is what a GitHub user commits. On GitHub the job's Docker runs `saphan/proven:latest` from Docker Hub, and the job posts with `PROVEN_TOKEN` or with `GITHUB_TOKEN`. The image has to have been pushed first. `make push` does that, and only after Docker Scout has passed on the image. This repository does not push it for you.

### Which account posts the review

The workflow posts as its own account, `proven-bot` (`LAB_BOT_USER`). It is not an admin. The lab admin is the account you sign in with, so a review the workflow wrote and a comment you or any other person wrote must not look alike. If the admin posted the review, an admin who types the marker by hand would pass for the workflow.

- The seed creates `proven-bot`, makes it a collaborator with write access on the demo repository, and mints that account's token. The token is stored only as the repository's Actions secret `PROVEN_TOKEN`, which the workflow reads. The admin's token is not given to the workflow.
- The account's generated password stays in the lab's state volume until `make clean`. Neither it nor the token is printed, written to a tracked file or baked into an image layer.
- `make e2e` accepts a review comment only from `proven-bot`, with the marker for the run and the head commit. The same marker from the admin or any other account is ignored, and e2e names the account it ignored.
- If the secret is missing, the job falls back to `GITHUB_TOKEN`; on Gitea that posts as another account, and e2e then fails closed.

The seeded copy replaces `saphan/proven` with `PROVEN_IMAGE` when that setting is not the default, and replaces `runs-on: ubuntu-latest` with `RUNNER_LABEL`.

## What `make` does besides the lab

`make` with no target prints every target, the version from `git describe`, whether `go` and `docker` were found, and the lab URL when the lab is up.

`make build` writes `./bin/proven` and `./bin/livecheck`. It uses Go when `go` is on `PATH` and the version is at least the `go` line in `go.mod`. Otherwise it builds the same binaries in `GOLANG_IMAGE` and writes them to the host. If neither Go nor Docker is available, it prints how to install Go on macOS and on Linux and exits 1.

`make build-all` cross-compiles for darwin and linux, amd64 and arm64, with `CGO_ENABLED=0`, into `./dist/`, and writes `./dist/SHA256SUMS`.

`make sample` runs the sample review from a temporary directory outside this repository. It does not pass keys and it does not call Docker.

`make livecheck` runs the live-service check. If `lab/.env` exists, it loads the keys listed in [CONFIGURATION.md](CONFIGURATION.md). With nothing set, the check names each service that is not ready and exits 1.

`make image` builds Proven's image for this machine and loads it. The load is one platform, this machine's own. A local engine cannot load a manifest list, so this target does not build both platforms. `make seed-image`, which `make up` uses, is the same kind of load: one platform, because the lab has to start the image. The image is the static `proven` and `livecheck` binaries, `git`, and the `docker` client, on `RUNTIME_IMAGE`, and it runs as a user that is not root. Local tags: `latest` and the `git describe` value. OCI labels `created`, `revision` and `version` come from the date, `git rev-parse` and `git describe`.

`make push` builds Proven's image and the seed image for `PUSH_PLATFORMS` (`linux/amd64` and `linux/arm64`) and pushes each tag as one manifest list that carries both platforms. Proven's tags are `latest` and the `git describe` value, on `PROVEN_REGISTRY/PROVEN_IMAGE`. The seed's only tag is `SEED_IMAGE_TAG`, on `PROVEN_REGISTRY/SEED_IMAGE`. The build asks for provenance (`--provenance=mode=max`) and an SBOM (`--sbom=true`). `make push` scans before it publishes: it builds each image for this machine, runs Docker Scout on it (the same scan as `make scan`, below), and only then builds and pushes the manifest lists. It refuses, exits non-zero and names the image when the `docker scout` plugin is missing, when a scan cannot run, or when Scout reports a critical or high finding that has a fix. Nothing is pushed in any of those cases.

`make scan` runs Docker Scout on the images this repository builds, Proven (`saphan/proven`, tag `latest`) and the seed (`saphan/proven-lab-seed`, tag `1`, unless `SEED_IMAGE` or `SEED_IMAGE_TAG` was overridden). It fails closed: it exits non-zero and names the image when the `docker scout` plugin is missing, when a scan cannot run, or when Scout reports a critical or high finding that has a fix. The severities are `SCOUT_REFUSE_SEVERITIES`. When the images pass, it also scans the pinned third-party images (`gitea/gitea`, `gitea/act_runner`, `docker`) and writes that output between the markers at the bottom of this page.

`make check-lab-gates` checks both gates offline, with a stub `docker` on `PATH`: `make up` refuses a `LAB_BIND` that is not a loopback address before any container starts, and `make scan` and `make push` refuse without a green scan, with `make push` never reaching the push.

## The runner and Docker

The lab has four services: Gitea, a Docker daemon that lives only inside the lab (`dind`), the Gitea Actions runner, and a one-shot seed.

Gitea does not serve SSH (`GITEA_DISABLE_SSH`, and `GITEA_START_SSH_SERVER` is false). The image already runs OpenSSH on port 22, and Gitea's own SSH server on that same port exited the container. The lab does not need SSH. The web UI is the only published port. `GITEA_SSH_PORT` is kept in the settings and is not published.

Every `gitea` command `make up` runs inside the container uses `GITEA_RUN_USER`. That account is `git` in this image. The same command as root is refused, and `compose exec` is root when no user is named. `make down` keeps the volume, so a second `make up` finds the password already stored. It still repairs a missing admin user, a missing runner registration and a missing pull request, then starts the runner.

The runner's Docker host is `tcp://dind:2375`. No service mounts the host's Docker socket. Job containers are started on that inner daemon. The inner daemon is privileged, because Docker-in-Docker needs it. That privilege stays inside the lab. The daemon's port is not published on the host.

`make up` also pulls the Prover's test image (`PROVEN_TEST_IMAGE`, `golang:1.26` by default) into that inner daemon, by digest when `TEST_IMAGE_DIGEST` is set. Without it the lab's `make livecheck` reports `No such image: golang:1.26`. The digest is pinned in `lab/versions.env` as `TEST_IMAGE_DIGEST`: it is the digest of the `golang:1.26` index (both platforms), read from Docker Hub on 2026-10-23. `make up` pulls `golang@<digest>` and tags it `golang:1.26`, so a judge gets the bytes we measured, not whatever the tag names later. To refresh it, run `docker buildx imagetools inspect golang:1.26`, copy the `Digest:` line of the index into `TEST_IMAGE_DIGEST`, and run `make up`. If you empty the setting, `make up` pulls the tag instead and prints the digest it got.

Gitea starts a workflow for a pull request only when the repository's Actions unit is on. The seed turns the unit on by name, and the Compose file adds `repo.actions` to Gitea's default units. If the unit was off when the pull request opened, the seed closes and reopens the pull request so the workflow starts. The seeded copy of the workflow posts to Gitea at `http://gitea:3000/api/v1` instead of `github.api_url`, because the job runs on the inner daemon, where `127.0.0.1` is not Gitea.

`make check-links` checks that every relative Markdown link and every path in `metadata.json` resolves.

The job image is `docker:<DOCKER_VERSION>-cli`, pinned. The runner does not pull an image it already has, so the Proven image loaded by `make up` is the one the workflow runs.

Sample mode does not start a Prover container. A later live run on this runner would start those containers on the lab's daemon, not on the host's.

## Rehearse

`make e2e` starts the lab when Gitea is not running and waits up to `E2E_WAIT_SECONDS`. Gitea 1.22.6 has no Actions runs, tasks, jobs or workflows route (HTTP 404), so e2e reads the statuses of the pull request's current head commit (`GET /repos/{owner}/{repo}/commits/{sha}/statuses`). It keeps the statuses of the workflow's job (`proven / review (pull_request)`, read from `lab/workflow/proven.yml`) and takes the highest run number in their `target_url`, because the seed may start more than one run on one head. It exits 0 only when that run's newest status is `success` and a comment by the workflow's own account (`LAB_BOT_USER`, `proven-bot` by default; see "Which account posts the review") starts with `<!-- proven-review run=N sha=<head commit> -->`, where N is that run's number and the commit is the pull request's current head. The workflow fills both from `github.run_number` and `github.event.pull_request.head.sha`. No time is compared, because two events in one second cannot be told apart by time. A run still `pending`, or no status yet, is waited for. `make check-e2e-parse` checks the parsers offline against comments and statuses shaped as Gitea serves them; it covers the lab admin posting the exact marker after a successful run that posted nothing, an unrelated account's marker, a successful run that posted nothing, an older run's review from the same second, and a review of another head commit. A review comment of an older run, of another commit, or by another account does not count and e2e says which it ignored: a lab volume kept by `make down` keeps its comments. If a kept volume still holds an older copy of the workflow on the pull request's branch, the run posts the old marker and e2e fails closed; run `make clean && make up`.

It exits 1 when that run finished as `failure` or `error`, and names the run and its status:

```text
e2e: workflow run #2 for pull request #1 (head commit <sha>) finished with status failure
```

It also exits 1 when the wait ends with no run, a run that is still going, or a successful run with no new review.

If Gitea is already running and the runner container is not, `make e2e` does not start the runner. It exits 1 and prints:

```text
e2e: runner is not running, so the workflow cannot post a review
```

That is the check that a stopped runner cannot look like a successful rehearsal. From the repository root, stop the runner and run the rehearsal again:

```bash
docker compose --env-file lab/versions.env -f lab/compose.yaml --project-directory lab stop runner
make e2e
```

A second check proves e2e cannot be fooled by an old comment. With the runner alive and an old review comment on the pull request, make the current run fail and require `make e2e` to refuse:

```bash
make e2e
make e2e-failed-run
```

`make e2e-failed-run` commits a workflow that fails on purpose to the pull request's head branch, runs `make e2e`, and puts the original workflow back. It exits 0 when e2e exited non-zero and named the failed run, and 1 when e2e accepted it. It needs the first `make e2e` to have passed, so an old comment exists. Restoring the workflow starts one more run; wait for it to finish before you run `make e2e` again.

## Third-party images

These pins are not built here. `make scan` replaces the block below with Docker Scout's output. A finding in a third-party image does not refuse the scan: we cannot patch it, and a judge should see it.

<!-- scout-third-party:start -->
Not scanned in the change that added the lab. The seat that wrote this page could not run Docker Scout. `make scan` fills this block.

Images: `gitea/gitea:1.22.6`, `gitea/act_runner:0.2.11`, `docker:29.9.1-dind`, `docker:29.9.1-cli`.
<!-- scout-third-party:end -->
