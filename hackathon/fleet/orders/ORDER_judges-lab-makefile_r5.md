# ORDER — judges-lab-makefile · round 5

- change: judges-lab-makefile
- round: 5
- base: main @ 859e4bc
- executor: cursor on runner-cursor
- written: 2026-10-23T10:13:15Z
- task: round 4 — run every `gitea admin` command as the git user, so make up finishes and e2e can run

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — judges-lab-makefile
track: proven-cursor-infra
- train: proven-cursor-infra
- wagon: W9
- round: 4
- review: named
- repo: proven
- base: main @ 859e4bc
- branch: fix/judges-lab-makefile
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 120m
- task: round 4 — run every `gitea admin` command as the git user, so make up finishes and e2e can run

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[judges-lab-makefile] <area>: <what>`.
No AI attribution anywhere. After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23 (translated): "3 start" — build the judges' lab and the standard Makefile now, in parallel with the guardrail
round. The decision is proven's ADR-0026 (`fleet/decisions/ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md`
at your base): read it first, it is the specification; it amends ADR-0017 point 6.

### ROUND 4 — WHAT THE MASTER MEASURED ON YOUR ROUND 3 (2026-10-23 ~09:02Z, runner-cursor, fresh clone of af657df; law #44)
- `git grep -n '\.saphan' -- Makefile lab docs README.md` prints nothing. GREEN. `make`, `make build` exit 0. GREEN.
- After `make clean` (0 lab volumes left), `make up` exit 2: it printed the admin user, password and URL, then
  `up: Gitea did not accept an admin command`.
- The cause, measured by the master inside the running container:
  `docker exec proven-lab-gitea-1 gitea admin user list` (as root) →
  `[F] Gitea is not supposed to be run as root. Sorry. …`;
  `docker exec -u git proven-lab-gitea-1 gitea admin user list` → `ID   Username Email IsActive IsAdmin 2FA` (works).
  `lab/████████████████` and `:86` run `compose exec -T gitea gitea admin …` with no user.
- Without the admin and the runner registration, `make e2e` stops at `e2e: runner is not running`; the negative control
  is therefore still not distinguishable. `make down` exit 0, nothing left.
- Note: `make down` keeps the lab volume, so a second `make up` reads "lab already installed"; that is by design only
  if `up` then still repairs what is missing (admin, runner, PR). Make `up` idempotent on a kept volume.

### ROUND 4 — WHAT TO DO
1. Every `gitea admin …` (and any other Gitea CLI call) runs as the image's `git` user (`compose exec -T -u git gitea …`,
   or the user named in `lab/versions.env`).
2. `make up` on a kept volume repairs what is missing and finishes; it never stops because the lab was installed before.
3. Run as much of `make up && make e2e` as your seat allows and write which lines you could not run.

### ROUND 3 — WHAT THE MASTER MEASURED ON YOUR ROUND 2 (2026-10-23 ~07:50Z, runner-cursor, fresh clone of 746b4d2; law #44)
- `make` exit 0; `make build` exit 0. GREEN.
- `make up` exit 2, verbatim:
  `lab/█████████████: 71: cannot create …/judges-lab-r2/.saphan/work/judges-lab-makefile/gitea-users.txt: Directory nonexistent`
  then `up: Gitea did not accept an admin command`.
  `lab/██████████████████████` write and read `$ROOT/.saphan/work/judges-lab-makefile/gitea-users.txt`. That directory is
  YOUR seat's artifact folder (this order's ARTIFACT CONFINEMENT), not part of proven: a judge's clean clone has no
  `████████` at all. The product must never read or write Saphan Studio's tree.
- `make e2e` exit 2: `e2e: runner is not running, so the workflow cannot post a review` (follows from `up`).
- `make status`: runner not running; pull request: `curl: (22) … 404`, `none open`.
- Negative control: not distinguishable — the runner never started, so "runner stopped" and "runner never up" read alike.
- `make down` exit 0, 0 lab containers left. GREEN.

### ROUND 3 — WHAT TO DO
1. No product file (Makefile, lab/, docs, scripts) may name `.saphan`. Temporary files go to a `mktemp` path removed on
   exit (trap), or under a lab state directory that `lab/versions.env` names and `make clean` removes. Prove it:
   `git grep -n '\.saphan' -- Makefile lab docs README.md` prints nothing (paste it).
2. Run the lab scripts as a judge would — from a clean clone with no `████████` — as far as your seat allows; anything
   your seat cannot run, write which line and why.
3. The ARTIFACT CONFINEMENT rule of this order is for YOUR evidence only; it never becomes a path inside proven's code.

### ROUND 2 — WHAT THE MASTER MEASURED ON YOUR ROUND 1 (2026-10-23 ~06:30Z, runner-cursor, outside your jail; law #44)
Round 1 (f2f27eb) passed what you could run (vet, tests, `make build` through `lab/████████████████`). The master ran the
acceptance lines your seat could not, from a fresh clone of `fix/judges-lab-makefile`:
- `make` (help): exit 0, every target listed, go 1.26.8 and docker 29.9.1 found. GREEN.
- `make build`: exit 0, ./bin/proven and ./bin/livecheck written. With Go off PATH: exit 0, built in `golang:1.26.8`. GREEN.
- `make up`: exit 2, `up: Gitea did not become healthy`. Gitea's own log, verbatim:
  `Unable to GetListener: listen tcp :22: bind: address already in use` →
  `Failed to start SSH server: listen tcp :22: bind: address already in use`, then the container exits (0).
  The image already runs OpenSSH on :22 and Gitea's built-in SSH server tries the same port.
- `make e2e` and the negative control: both stopped at the same `up` failure, so neither was measured.
- `make down`: exit 0, no lab container left.
- A build warning: `InvalidDefaultArgInFrom: Default value for ARG ${RUNTIME_IMAGE} results in empty or invalid base image name (line 5)`.
- No fleet host has the local `docker scout` plugin (runner-cursor, █████████████, █████, ████████████).

### ROUND 2 — WHAT TO DO
1. The lab needs no SSH: turn Gitea's SSH off (e.g. `GITEA__server__DISABLE_SSH=true`, or `START_SSH_SERVER=false` if
   OpenSSH must stay) as a setting in `lab/versions.env`/Compose; `make up` brings Gitea to healthy.
2. The owner, 2026-10-23: "obie platformy" — every image we build (proven's image, the lab's seed image) is built for
   `linux/amd64` and `linux/arm64` and pushed as ONE tag carrying both (a manifest list). Only `make image`'s local `--load`
   stays single-platform, because a local engine cannot load a manifest list; say so in its help line and in docs/LAB.md.
3. The owner, 2026-10-23: "Docker Scout uruchamiany na docker hub po wgraniu obrazów" — Scout runs on Docker Hub after
   the push. `make push` must not require a local `docker scout`; `make scan` runs it when the plugin is present and
   otherwise says, by name, that the scan is read on Docker Hub after `make push`. docs/LAB.md says where the result is.
4. Give the seed Dockerfile's `ARG RUNTIME_IMAGE` a valid default (from `lab/versions.env`), so the warning is gone.
5. Comments follow ADR-0027 (the train order's section "CODE IS WRITTEN TO BE READ").
If you cannot run Docker on your seat, run what you can and say which acceptance lines you could not; the master runs
`make up && make e2e`, the negative control and `make down` again.

### WHAT THE MASTER MEASURED (2026-10-23 ~02:20Z; claims to test, law #44)
- (a) main @ 859e4bc holds no `Makefile`, no Compose file, no `.gitea/` or `.github/` workflow and no image Dockerfile of
  proven: `git ls-tree -r --name-only main` finds only `fleet/infra/go-python/Dockerfile`.
- (b) proven execs only two programs: `git` (`internal/change/between.go:57,89,127,139,204`) and the container client
  (`internal/testrun/container.go:390`, `exec.CommandContext(ctx, argv[0], …)`).
- **First act after your status file:** re-measure (a) and (b) yourself and paste what you found. If either is wrong, say
  so in `- master-was-wrong:` and build on what is there.

### WHAT TO BUILD (ADR-0026's points, carried)
1. `Makefile` at proven's root. `make` with no target = help: every target from its `##` comment, the version, whether
   go and docker were found, the lab address when the lab is up. OCI labels (created, revision, version) from `date`,
   `git rev-parse`, `git describe`; images with `docker buildx`.
2. `make build` → `./bin/proven` and `./bin/livecheck`: Go at or above `go.mod`'s version → `go build`; else Docker → the
   same build in a pinned golang image, binaries written to the host; else print how to install Go on macOS and Linux and
   exit 1.
3. `make build-all` → darwin and linux × amd64 and arm64, `CGO_ENABLED=0`, into `./dist/` with checksums.
4. `make image` (local platform, `--load`); `make push` (linux/amd64 and linux/arm64, `--provenance=mode=max --sbom=true`)
   to Docker Hub `saphan/proven`. Registry and image name are settings.
5. Images are minimal: proven's image = the static binary + `git` + the container client, the smallest base that holds
   them, a non-root user. `make scan` = Docker Scout over every image we build; `make push` runs it first and refuses on a
   critical or high finding with a fix available (severities are a setting). Third-party pinned images (Gitea, the
   runner, docker:dind) are scanned and their results written into `docs/LAB.md`.
6. Proven's image tags: `latest` and the `git describe` tag. Every other image pinned to a version.
7. `make up` / `make down`: one Compose file — Gitea, a Gitea Actions runner, a one-time seed (an admin with a generated
   password printed once, a demo repository with an open pull request and proven's workflow, the runner's
   registration). Everything binds to 127.0.0.1 only.
8. The runner starts the Prover's containers in Docker-in-Docker inside the lab, never through the host's Docker socket.
9. `make status` (containers, Gitea health, runner registered, last workflow result on the PR, lab URL); `make logs`
   (`SERVICE=` for one service).
10. `make sample` (a sample review run from a temp directory outside the repository; no network, keys or Docker);
    `make livecheck` (reads an optional `lab/.env`).
11. `make e2e`: up → wait for the workflow → ask Gitea's API whether proven's review is on the PR → exit 0 or 1.
12. `make clean`: the lab's containers and volumes, local images, `./bin`, `./dist`.
13. No hardcoded settings (ADR-0013): every version, port and name in `lab/versions.env`, overridable from the
    environment, each listed in `docs/CONFIGURATION.md`. Keys only in `lab/.env` (never committed); `lab/.env.example`
    is committed.
14. Runbook `docs/LAB.md` (law #39: the judge's steps are a deliverable co-equal with the code); the README quick start
    begins with `make up`.
The lab's workflow is the same file a GitHub user uses; proven's Action is not on main yet, so the lab's workflow may call
`proven scan --sample` and `docs/LAB.md` says so.

### ACCEPTANCE (the owner reads these)
1. From a clean checkout: `make` prints the help; `make build` works with Go; with Go off `PATH` it builds through Docker.
2. `make up && make e2e` exits 0 and the review is on the PR (paste Gitea's API answer); `make scan` runs; `make down`
   leaves nothing running (`docker ps` pasted).
3. **Negative control (law #27/#41):** with the runner stopped, `make e2e` exits 1 and names why; paste it.
4. `go vet ./...` clean and `go test -count=1 ./...` green; no Go file outside what the Makefile needs is changed.
If your seat cannot run Docker or bind a listener, build and commit everything, run what you can, write in your status
exactly which acceptance lines you could not run and why, and set `- state: stop-2-ready` — the master runs the lab on
another host (law #9a). Never claim a line you did not run.

### THIS ORDER CAN BE WRONG (law #44)
Write a `- master-was-wrong:` line: `—`, or what you measured that contradicts this order or ADR-0026.

### WHAT IS NOT IN SCOPE
- proven's Go code beyond what the Makefile needs; the GitHub Action (proven-cli-and-action's W4); pushing an image.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse HEAD` output
verbatim; never restate HEAD from memory.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `█████████████████████████████████` in the
worktree. Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run
and their exit codes, the acceptance lines run and not run. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| file and line in Saphan Studio's private code | 4 |
| how an order is bound to the laws | 1 |
| machine name | 3 |
| private workspace path | 4 |
