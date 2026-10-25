# STATUS — judges-lab-makefile · round 1

- state: ready-for-review
- executor: cursor executor
- model: grok-4.7-high
- pool: not shown to this leg
- cost: not shown to this leg
- duration: not measured as one figure; the work commit time is 2026-10-23T04:50:20Z
- last commit: ef04ad3cfb5accfbd812c8f1782ace21cf7f7a76 "[judges-lab-makefile] makefile: add the build and the judges' lab"
- updated: 2026-10-23T04:52:17Z

## Summary

One root Makefile builds Proven and runs the judges' lab. The lab workflow calls `proven scan --sample` and posts that text on the pull request, because the Action is not on main yet. `go vet ./...` and `go test -count=1 ./...` are green. No Go file changed. This seat cannot run `make` or `docker`, so the lab was not started here.

## Acceptance, item by item

1. UNRUN as `make`. `make --version` was tried once and refused; the allowed form was none. The script `make build` calls, `lab/scripts/build.sh`, was run with Go on PATH and exited 0. It wrote `bin/proven` (10444032 bytes) and `bin/livecheck` (9599940 bytes). The Docker fallback was not run. `make` with no target was not run, so the help text was not printed.
2. UNRUN. `docker version` was tried once and refused; the allowed form was none. `go run ./cmd/livecheck` also reported permission denied on the local Docker API. `make up`, `make e2e`, `make scan` and `make down` were not run. No Gitea API answer. No `docker ps`.
3. UNRUN. The runner was never started, so the negative control was not executed. The message in `lab/scripts/e2e.sh` is `e2e: runner is not running, so the workflow cannot post a review`. That line was not produced by a run.
4. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0 in 34s: 29 packages `ok`, 8 packages with no test files, 0 failed. `git diff 859e4bccdae17f1091124324119b10f6195051f8 HEAD -- '*.go'` printed no paths.

## Evidence

- Toolchain: `go version go1.26.8 linux/amd64`. Module line `go 1.26`.
- `sh lab/scripts/build.sh` exit 0.
- `go run ./cmd/livecheck` exit 1. Printed: `summary: 0 ready, 4 not ready`. Generator, guardrail and query refused empty configuration. The container line was permission denied on the local Docker API.
- `go run ./cmd/proven scan --sample testdata/fixture/discount` from the repository root exited 1. The program printed that its state directory is inside the repository and `exit status 2`. A directory outside the worktree could not be created (permission denied), so the outside sample was not run.
- `lab/versions.env` has 28 keys. Each appears in `docs/CONFIGURATION.md`. Missing keys: none.
- Image tags in `lab/versions.env` were not pulled. This seat has no DNS (`Name or service not known` on the registry hosts that were asked) and cannot run `docker`. A tag that exists and a tag that does not exist look the same in that file.
- remark: image tags in lab/versions.env are UNMEASURED; the file cannot tell a tag that exists from a tag that does not.
- Testing bar: no metric is declared for this order. UNRUN.
- Session recording: none. This leg was not given a recording file to publish.

## master-was-wrong

There is no local `main` ref, so `git ls-tree main` cannot be run. The tree was read at HEAD `859e4bccdae17f1091124324119b10f6195051f8`. That tree has no Makefile, no Compose file, no `.github/` and no `.gitea/`, and the only Dockerfile is `fleet/infra/go-python/Dockerfile`. Claim (a) holds for that tree. Claim (b) holds outside test files: production `exec.Command` is `git` in `internal/change/between.go` and the container client in `internal/testrun/container.go`. Test files also exec `sh`, `git`, `go` and `bash`.

## Rulings

- GNU make. Recipes are `sh` scripts under `lab/scripts/`.
- Host `make build` sets `CGO_ENABLED=0`.
- The runtime base is `alpine:3.22`, plus `ca-certificates`, `git` and `docker-cli`. The image contains `proven` and `livecheck` and runs as uid 65532.
- The workflow file repeats the default image name `saphan/proven:latest` so a GitHub user can copy it. The seed rewrites that name from `PROVEN_IMAGE`.
- The posted comment is sample text. reviewdog is out of scope.
- The seed starts as root only to change ownership of the volumes, then drops to uid 65532. The runner image and the DinD image are upstream and run as root. DinD is privileged and does not mount the host socket.
- Third-party tags were not pulled. The first `make up` is the existence check.
- Job containers use the DinD network namespace (`container.network: host` in the runner config) so a job can reach Gitea by the Compose name while the browser URL stays on 127.0.0.1. That path was not executed.
- No `.github/` directory was created in this repository. The workflow source is `lab/workflow/proven.yml`. The seed copies it into the demo repository.
- The admin password is stored in the lab volume until `make clean` and printed only on the first `make up`.

## Blocked on / next

The master runs acceptance lines 1, 2 and 3 on a host that can run Docker and bind a listener. This leg's build and the documents are committed.
