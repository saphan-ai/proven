# ORDER — judges-lab-makefile · round 13

- change: judges-lab-makefile
- round: 13
- base: main @ 9303b2d
- executor: claude-code
- written: 2026-10-23T20:30:12Z
- task: round 6 — make e2e passes only on the CURRENT workflow run's own review (the codex review's BLOCKING), a failed-run negative control, the guide's digest paragraph

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — judges-lab-makefile
track: proven-cursor-infra
- train: proven-cursor-infra
- wagon: W9
- round: 6
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/judges-lab-makefile
- lands-on: main
- backend: claude-code
- quote-usd: 11
- cap-usd: 22
- wall: 120m
- task: round 6 — make e2e passes only on the CURRENT workflow run's own review (the codex review's BLOCKING), a failed-run negative control, the guide's digest paragraph

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

### ROUND 6 — WHAT THE MASTER MEASURED ON YOUR ROUND 5 (2026-10-23 ~19:10Z, runner-cursor, fresh clone; law #44)
- Your round 5 fixed task creation: Gitea now creates a task for PR #1 and the runner takes it. GREEN.
- Every task then failed: `OCI runtime exec failed: exec: "bash": executable file not found in $PATH` — the job image
  `docker:${DOCKER_VERSION}-cli` (Alpine) has no bash, and act runs a step with bash unless the workflow names a shell.
- The master committed d8514fe on your branch: `defaults: run: shell: sh` in `lab/workflow/proven.yml`, and
  `TEST_IMAGE_DIGEST` = the golang:1.26 index digest (read from Docker Hub that day). Start from d8514fe.
- At d8514fe, fresh clone: `make clean` 0 · `make up` 0 · `make e2e` 0 ("proven review is on the pull request", comment on
  PR #1) · runner container stopped → `make e2e` exit 2 naming the runner · `make down` 0 · `make clean` 0 · `make check-links` 0.

### ROUND 6 — THE REVIEW YOU MUST PAY (codex review of d8514fe, verdict reject; carried verbatim, law #44)
> BLOCKING: reject stale or unrelated matching comments when the current workflow fails; exercise a failed-current-run
> negative control and a clean positive run. OWED: fix the guide's digest paragraph.

The defect, in the master's words (a claim — refute it if wrong): `lab/██████████████` passes when ANY comment on the first
open pull request contains `proven-review`. With a kept volume (`up.sh` says "the volume is kept; repairing …") a comment
from an earlier run stays, so e2e reads green even when the current workflow run failed.

### ROUND 6 — WHAT TO DO
1. **e2e passes only on the current run's own review.** It passes only when (a) the workflow run for the pull request's
   CURRENT head commit finished successfully, and (b) the `proven-review` comment it checks was posted by that run (for
   example: created at or after that run started). Gitea is 1.22.6 (`lab/versions.env`); read which Actions API it
   serves at your base before you rely on one, and name it in your status. A run that failed makes e2e exit non-zero and
   say so, naming the run and its status, even when an older matching comment exists.
2. **A failed-current-run negative control** you can run and the master can repeat: a documented way to make the current
   run fail (for example a setting the seed reads, or a script under `lab/scripts/` used only by the check) while an old
   `proven-review` comment is present; e2e must then exit non-zero. Write the exact commands into your status.
3. **docs/LAB.md, the digest paragraph:** it still says the digest is empty until someone reads it. It is now pinned in
   `lab/versions.env`; say so, and say how to refresh it.
4. Keep the existing negative control (runner stopped → exit naming the runner) and `make check-links` green.

### ROUND 6 — ACCEPTANCE
- On a seat with Docker: `make clean && make up && make e2e` exit 0; the failed-current-run control exits non-zero with an
  older `proven-review` comment present; runner stopped → non-zero naming the runner; `make down`, `make clean` exit 0.
  If your seat cannot run make or docker, say which lines you could not run — the master runs them.
- `go vet ./...`, `go test ./...`, `sh scripts/check-links.sh` exit 0; `sh -n` on every changed script.
- `git log main..HEAD` shows no merge commit and no `Co-authored-by`.
- **master-was-wrong:** if the defect above is not what the review means, or e2e already rejects stale comments, show the
  measurement and say so in your status.

### ROUND 5 — WHAT THE MASTER MEASURED ON YOUR ROUND 4 (2026-10-23 ~10:40Z, runner-cursor, fresh clone of b6c121e; law #44)
- `make`, `make build` exit 0. `make up` exit 0: Gitea healthy, runner registered (`runner: proven-lab, version v0.2.11,
  labels [ubuntu-latest], declare successfully`), PR #1 open. GREEN.
- `make e2e` exit 2: `e2e: no proven review on the pull request after 180s`. The master looked inside: the runner
  daemon is up and idle (its log shows only "Starting runner daemon" and the declare line), `[actions] ENABLED = true`
  in app.ini, the workflow is at `.github/workflows/proven.yml` on the demo repo, refs `main`, `change`, `refs/pull/1/head`
  exist — but NO task was created for PR #1. Find why (the order the seed creates the repo, the workflow, the runner and
  the PR in; the repo's Actions unit; the event the workflow listens to) and fix it.
- Negative control GREEN: with the runner stopped, `make e2e` exits 2 naming the runner. `make down` GREEN.
- The master pinned `GOLANG_IMAGE=golang:1.26.10` (b574b7a, 47ebaf5: the security review's toolchain fix). Keep it.

### ROUND 5 — WHAT ELSE CHANGED
1. **ADR-0030 moved the repository (main e109e53).** `fleet/` is now `hackathon/fleet/`, `fleet/decisions/` is
   `docs/adr/`, `tests/corpus` is `test/corpus`. The master already carried your branch onto the moved and rewritten
   main (@ 9303b2d); your status files are in `hackathon/fleet/status/`. Start from your branch tip; no rebase is needed. Every link you write points at the new paths.
2. **`make check-links`** runs `scripts/check-links.sh` (on main since e109e53), listed in the help like every target.
3. **The Prover's test image** (`PROVEN_TEST_IMAGE`, default `golang:1.26`, `internal/testrun/settings.go`): measured on
   the owner's machine, the lab's livecheck reports `No such image: golang:1.26`. `make up` pulls it into the
   Docker-in-Docker daemon before the first review, pinned by DIGEST, the digest a setting in `lab/versions.env`
   (ADR-0026 point 13) and listed in docs/CONFIGURATION.md.

### ROUND 5 — ACCEPTANCE ADDED
- `make check-links` exit 0 on your tip; `git log --oneline main..HEAD` shows no merge commit.
- Run `make up && make e2e` if your seat can; otherwise say which lines you could not run — the master runs them.

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

### ATTRIBUTION (owner, 2026-10-23)
No commit of yours may carry a `Co-authored-by:` trailer or a seat identity as its author. Check
`git log --format="%an <%ae>%n%B" main..HEAD` before you end and say so in your status.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| file and line in Saphan Studio's private code | 5 |
| how an order is bound to the laws | 1 |
| machine name | 3 |
| private workspace path | 4 |
