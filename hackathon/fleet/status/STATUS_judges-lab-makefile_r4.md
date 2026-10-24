# STATUS — judges-lab-makefile · round 4

- state: ready-for-review
- executor: cursor executor
- model: grok-4.7 (the order named grok-4.7-high; this session identifies as Grok 4.7)
- pool: not shown to this leg
- cost: not shown to this leg
- duration: not measured as one figure; the work commit time is 2026-10-23T10:28:20Z
- last commit: 04fffe2ac86b625593e718940cb5d405dd73606b "[judges-lab-makefile] lab: run gitea commands as git"
- updated: 2026-10-23T10:28:39Z

## Summary

`make up` runs every `gitea` command as `GITEA_RUN_USER` (`git`). A kept volume is not a stop: the same run creates a missing admin, writes a new runner registration token, runs the seed again, and starts the runner. This seat cannot run `make` or `docker`, so the lab was not started here.

## Acceptance, item by item

1. UNRUN as `make` and `make build`. `make --version` was tried once and refused; the allowed form was none. `sh -n lab/scripts/up.sh` exited 0. `sh -n lab/seed/seed.sh` exited 0. Those are syntax checks, not a build.
2. UNRUN. `docker version` was tried once and refused; the allowed form was none. `make up`, `make e2e`, `make scan` and `make down` were not run. No Gitea API answer. No `docker ps`.
3. UNRUN. The runner was never started, so the negative control was not executed. The message in `lab/scripts/e2e.sh` is `e2e: runner is not running, so the workflow cannot post a review`. That line was not produced by a run.
4. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0; the command returned in 51s: 29 packages `ok`, 8 packages with no test files, 0 failed. `git diff --name-only 859e4bccdae17f1091124324119b10f6195051f8 -- '*.go'` printed no paths.

## Round 4 proof

The only `compose exec` line in `lab/scripts/up.sh` is `compose exec -T -u "$GITEA_RUN_USER" gitea gitea "$@"`. Both admin calls go through that function. `lab/versions.env` sets `GITEA_RUN_USER=git`.

A shell probe loaded that function and called it. With `GITEA_RUN_USER=git` the recorded lines were `exec -T -u git gitea gitea admin user list` and `exec -T -u git gitea gitea admin user create --admin --username proven`. With `GITEA_RUN_USER=notroot` the line was `-u notroot` and did not contain `-u git`. With `GITEA_RUN_USER` set empty, the guard in `lab/scripts/up.sh` exited 1 and printed `up: GITEA_RUN_USER is not set`. The probe exited 0.

The kept-volume branch prints that it is repairing the admin user, the runner and the pull request. From that sentence to the function there is no `exit`. `lab/seed/seed.sh` no longer returns when a registration token file already exists, and it treats workflow HTTP 422 as already current.

`git grep -n -e '\.saphan' -- Makefile lab docs README.md` exited 1 and printed nothing.

## Evidence

- Toolchain: `go version go1.26.8 linux/amd64`. Module line `go 1.26`.
- `sh -n lab/scripts/up.sh` exit 0. `sh -n lab/seed/seed.sh` exit 0.
- The function probe: exit 0, `PROBE PASS`.
- `git grep` of `Makefile`, `lab`, `docs` and `README.md` for `\.saphan`: exit 1, empty output.
- Testing bar: no metric is declared for this order. UNRUN.
- Session recording: none. This leg was not given a recording file to publish.
- remark: the function probe runs `gitea_cli` with a shell `compose`, so it cannot show that the Gitea image accepts the git user.
- remark: a `compose exec` the script built in another way would not appear in that function's log.
- remark: the no-exit check reads the source between two lines; it is not a second `make up`.
- remark: `git grep` exit 1 with an empty page means no line matched the literal.
- remark: the `make` and `docker` refusals name the allow list; they do not say whether those programs are installed.
- remark: the 51s is the wall time of `go test -count=1 ./...`, not the sum of the per-package lines. A failed package would have been a `FAIL` line. There was none.

## master-was-wrong

There is no local `main` ref (`git rev-parse --verify refs/heads/main` exit 128, `fatal: Needed a single revision`). The tree was read at `859e4bccdae17f1091124324119b10f6195051f8`. That tree has no Makefile, no Compose file, no `.github/` and no `.gitea/`, and the only Dockerfile is `fleet/infra/go-python/Dockerfile`. Claim (a) holds for that tree. Claim (b) holds outside test files: production `exec.Command` is `git` in `internal/change/between.go` (lines 57, 89, 127, 139, 204) and the container client in `internal/testrun/container.go:390`. Test files also exec `bash`, `go`, `git` and `sh`. No Go file differs from that tree. Remeasured this round; the lines match the order.

## Rulings

- The account is the setting `GITEA_RUN_USER`, default `git`, because the order allows the user named in `lab/versions.env`. Cost if wrong: an image whose command-line account is not `git` fails the admin command. The captured stderr is printed when the list never succeeds.
- A kept volume re-runs the seed and replaces the registration token. Workflow HTTP 422 is treated as already current. ADR-0026 calls the seed one-time; this round's order says a second `make up` repairs what is missing and does not stop. The password is still printed once. Cost if wrong: a 422 that means the update was rejected is ignored, and the workflow stays as it was. A token written while a runner is waiting to register is the token that runner reads; a token written after the runner has registered is not read.
- `chmod` was tried once, to mark a stub executable, and refused; the allowed form was none. That stub was not run. The probe above does not need it. Cost if wrong: none for the product. The lab scripts were not executed as a whole.

## Blocked on / next

The master runs `make up` and `make e2e` on a host that can run Docker, then the negative control, `make scan` and `make down`. This leg's edit is committed. This status follows in the fleet commit.
