# STATUS — judges-lab-makefile · round 2

- state: ready-for-review
- executor: cursor executor
- model: grok-4.7-high
- pool: not shown to this leg
- cost: not shown to this leg
- duration: not measured as one figure; the work commit time is 2026-10-23T07:40:44Z
- last commit: 260003f0168254a5fd794c220138100ebb97aba1 "[judges-lab-makefile] lab: keep Gitea up and publish both platforms"
- updated: 2026-10-23T07:41:01Z

## Summary

Gitea SSH is off, so the container no longer starts its own SSH server on port 22. `make push` publishes Proven and the seed as one tag each, a manifest list of linux/amd64 and linux/arm64, and it does not run Docker Scout. `make image` still loads one platform. The seed Dockerfile gives `RUNTIME_IMAGE` the default from `lab/versions.env`. This seat cannot run `make` or `docker`, so the lab was not started here.

## Acceptance, item by item

1. UNRUN as `make`. `make --version` was tried once and refused; the allowed form was none. The help text was not printed. `sh -n` on `lab/scripts/image.sh`, `seed-image.sh`, `push.sh` and `scan.sh` exited 0. That is a syntax check, not a build.
2. UNRUN. `docker version` was tried once and refused; the allowed form was none. `make up`, `make e2e`, `make scan`, `make push` and `make down` were not run. No Gitea API answer. No `docker ps`. Whether Gitea becomes healthy is not measured on this seat.
3. UNRUN. The runner was never started, so the negative control was not executed. The message in `lab/scripts/e2e.sh` is `e2e: runner is not running, so the workflow cannot post a review`. That line was not produced by a run.
4. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0 in 36s: 29 packages `ok`, 8 packages with no test files, 0 failed. `git diff --name-only 859e4bccdae17f1091124324119b10f6195051f8 HEAD -- '*.go'` printed no paths.

## Evidence

- Toolchain: `go version go1.26.8 linux/amd64`. Module line `go 1.26`.
- `lab/versions.env` has 30 keys. Each appears in `docs/CONFIGURATION.md`. Missing keys: none.
- The help lines in the Makefile: 15 targets, each with a `##` comment. `image` says a local engine cannot load a manifest list.
- `sh -n` exit 0 on the four scripts this round edited. The other scripts were not re-checked.
- Image tags in `lab/versions.env` were not pulled. `docker` is denied on this seat. A tag that exists and a tag that does not exist look the same in that file.
- remark: image tags in lab/versions.env are UNMEASURED; the file cannot tell a tag that exists from a tag that does not.
- remark: the Docker Hub page was not opened. The runbook says the scan is read there because the owner said Scout runs on Docker Hub after the push. This seat did not see that page.
- Testing bar: no metric is declared for this order. UNRUN.
- Session recording: none. This leg was not given a recording file to publish.

## master-was-wrong

There is no local `main` ref (`git rev-parse --verify --quiet refs/heads/main` exit 1). The tree was read at `859e4bccdae17f1091124324119b10f6195051f8`. That tree has no Makefile, no Compose file, no `.github/` and no `.gitea/`, and the only Dockerfile is `fleet/infra/go-python/Dockerfile`. Claim (a) holds for that tree. Claim (b) holds outside test files: production `exec.Command` is `git` in `internal/change/between.go` (lines 57, 89, 127, 139, 204) and the container client in `internal/testrun/container.go:390`. Test files also exec `bash`, `go`, `git` and `sh`.

ADR-0026 point 5 still says `make push` runs Docker Scout first and refuses on a critical or high finding with a fix. Round 2 of this order says the opposite, in the owner's words: Scout is read on Docker Hub after the push, and `make push` must not need a local plugin. The Makefile follows round 2. The accepted ADR was not edited.

ADR-0027 is not in this checkout, and `fleet/orders/TRAIN_ORDER_proven-cursor-infra.md` has no section "CODE IS WRITTEN TO BE READ". Comments added in this round name the constraint a reader would otherwise miss: why SSH is off, why a local load is one platform, and why push does not call Scout.

## Rulings

- `GITEA_DISABLE_SSH=true` and `GITEA_START_SSH_SERVER=false` are settings. The SSH port is not published. `GITEA_SSH_PORT` stays in the file and is documented as not published. Cost if wrong: a judge who expects SSH on port 2222 finds no listener.
- `make seed-image` stays a one-platform `--load`, same as `make image`. `make up` has to start that image, and a local engine cannot load a manifest list. `make push` is where the seed becomes one tag of both platforms. Cost if wrong: the owner wanted the local seed target to refuse a one-platform load.
- `SEED_IMAGE` is now `saphan/proven-lab-seed`, the same owner prefix as `saphan/proven`, so the pushed name is a repository under that owner. Cost if wrong: the Hub repository name is not the one the owner will create. The setting overrides it.
- Proven's Dockerfile also has `ARG` defaults (`golang:1.26.8`, `alpine:3.22`). The warning the master pasted is line 5 of the seed file. Proven's file has the same empty-default shape. Cost if wrong: a bare `docker build` without `make` uses those pins even after `lab/versions.env` changes. `make` still passes the settings, and those values win.
- `make scan` exits 0 when `docker scout` is not installed, and prints that the scan is read on Docker Hub after `make push`. Cost if wrong: a caller that treated a missing plugin as a failed scan now sees success. The sentence names the gap.
- The third-party block in `docs/LAB.md` is left as round 1 wrote it. `make scan` still replaces it when the plugin is present. This seat could not run that replacement.

## Blocked on / next

The master runs acceptance lines 1, 2 and 3 on a host that can run Docker and bind a listener: `make`, `make build`, `make up`, `make e2e`, the negative control, `make scan`, and `make down`. This leg's edits and this status are committed.
