# STATUS — judges-lab-makefile · round 3

- state: ready-for-review
- executor: cursor executor
- model: grok-4.7 (the order named grok-4.7-high; this session identifies as Grok 4.7)
- pool: not shown to this leg
- cost: not shown to this leg
- duration: not measured as one figure; the work commit time is 2026-10-23T08:52:00Z
- last commit: fc468c72d561378bbe26ce9b8c3d91f5fa490cf2 "[judges-lab-makefile] lab: keep the user list in a temp file"
- updated: 2026-10-23T08:52:00Z

## Summary

`make up` no longer writes the Gitea user list under a private tree. The list is a temporary file, created with `mktemp` and removed when `lab/scripts/up.sh` exits. A clean clone has no such tree, and the old redirect failed before Gitea was asked. This seat cannot run `make` or `docker`, so the lab was not started here.

## Acceptance, item by item

1. UNRUN as `make`. `make --version` was tried once and refused; the allowed form was none. The help text was not printed. `sh -n lab/scripts/up.sh` exited 0. That is a syntax check, not a build.
2. UNRUN. `docker version` was tried once and refused; the allowed form was none. `make up`, `make e2e`, `make scan` and `make down` were not run. No Gitea API answer. No `docker ps`.
3. UNRUN. The runner was never started, so the negative control was not executed. The message in `lab/scripts/e2e.sh` is `e2e: runner is not running, so the workflow cannot post a review`. That line was not produced by a run.
4. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0 in 34s: 29 packages `ok`, 8 packages with no test files, 0 failed. `git diff --name-only 859e4bccdae17f1091124324119b10f6195051f8 HEAD -- '*.go'` printed no paths.

## Round 3 proof

Before the edit, a redirect into a missing private-tree child exited 2. The shell said `Directory nonexistent`. That is the same failure the master pasted for `lab/scripts/up.sh:71`.

After the edit, `git grep -n '\.saphan' -- Makefile lab docs README.md` exited 1 and printed nothing. Exit 1 from `git grep` means no match.

A separate check created a temporary file in a directory that has no private-tree child, wrote a line, found it, and removed the file. It printed `TEMP-OK`. That check does not start the lab.

## Evidence

- Toolchain: `go version go1.26.8 linux/amd64`. Module line `go 1.26`.
- `sh -n lab/scripts/up.sh` exit 0.
- The reproduction of the old redirect: exit 2, `Directory nonexistent`.
- The temporary-file check: exit 0, `TEMP-OK`.
- `git grep` of `Makefile`, `lab`, `docs` and `README.md` for `\.saphan`: exit 1, empty output.
- Testing bar: no metric is declared for this order. UNRUN.
- Session recording: none. This leg was not given a recording file to publish.
- remark: git grep exit 1 with an empty page means no line matched the literal; a path built from pieces would also print nothing, and up.sh was read after the edit.
- remark: TEMP-OK shows a temporary file can be written where a missing directory cannot; it does not show Gitea accepting an admin command.
- remark: the make and docker refusals name the allow list; they do not say whether those programs are installed.

## master-was-wrong

There is no local `main` ref (`git rev-parse --verify --quiet refs/heads/main` exit 1). The tree was read at `859e4bccdae17f1091124324119b10f6195051f8`. That tree has no Makefile, no Compose file, no `.github/` and no `.gitea/`, and the only Dockerfile is `fleet/infra/go-python/Dockerfile`. Claim (a) holds for that tree. Claim (b) holds outside test files: production `exec.Command` is `git` in `internal/change/between.go` (lines 57, 89, 127, 139, 204) and the container client in `internal/testrun/container.go:390`. Test files also exec `bash`, `go`, `git` and `sh`. Remeasured this round; the lines match the order.

## Noticed

`conformance/meterless/run.sh` on the base writes under a private tree. It is outside the proof paths (`Makefile`, `lab`, `docs`, `README.md`) and it was not the failure the master measured. It was not changed.

## Rulings

- The user list is a `mktemp` file with an EXIT trap, not a directory named in `lab/versions.env`. The script reads the list once and deletes it, which is what the old script did after the check. A state directory would be a new setting and a `make clean` path for a file that must not outlive `up`. Cost if wrong: a judge who wanted the list kept after `up` returns will not find it.
- `conformance/meterless/run.sh` stays as it is on the base. Cost if wrong: that script still writes a private tree on a clean clone. The measured failure, and the proof command, are the lab.

## Blocked on / next

The master runs acceptance lines 1, 2 and 3 on a host that can run Docker and bind a listener: `make`, `make build`, `make up` from a clean clone, `make e2e`, the negative control, `make scan`, and `make down`. This leg's edit and this status are committed.
