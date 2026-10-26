# ORDER — the-image · round 1

- change: the-image
- round: 1
- base: main @ 2b8fd4a
- executor: cursor on runner-cursor
- written: 2026-10-24T16:11:56Z
- task: Proven's container image is checked against the binary it carries: the same version line and the same sample review, with its size printed; a script anyone can run does the comparison and fails when they differ, and anything the image gets wrong is fixed; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-image
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W3
- review: named
- repo: proven
- base: main @ 2b8fd4a
- branch: fix/the-image
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 120m
- task: Proven's container image is checked against the binary it carries: the same version line and the same sample review, with its size printed; a script anyone can run does the comparison and fails when they differ, and anything the image gets wrong is fixed; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR/image push · author = fleet identity.
Commit subjects: `[the-image] <area>: <what>`, the area ONE token, at most 72 characters, English
(the merge refuses others). No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name.

### SKILLS
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHERE YOU START (master, 2026-10-24; claims to test, law #44)
- Main already carries a root `Dockerfile` (two stages: `golang:1.26.8` builds `proven` and `livecheck` with
  `CGO_ENABLED=0 -trimpath -ldflags "-s -w"`; runtime `alpine:3.22` with `ca-certificates git docker-cli`, user 65532, the
  fixture copied to `/opt/proven/fixture/discount`, `ENTRYPOINT ["proven"]`, OCI labels from build args) and `make image`
  (`lab/████████████████`). Nobody has yet measured the image against the train's definition of done:
  `docker run <image> --version` and `docker run <image> scan --sample /opt/proven/fixture/discount` give the same output as the
  binary built from the same commit, and the image's size is printed.
- The master measured that this host (runner-cursor) has a docker engine (server 29.9.1). Whether your sandbox may reach it is
  not measured: if `docker info` is refused, that is a permissions fact — write `- state: blocked` and stop, do not hunt.

### WHAT TO DO
1. **Reproduce first:** build the binary (`go build -o <tmp>/proven ./cmd/proven`) and the image (`docker build -t proven:w3 .`
   with the build args the Makefile passes); run both commands on both, from a directory outside the checkout, and paste the
   four outputs and `docker image inspect proven:w3 --format '{{.Size}}'`. Note every difference.
2. A script `scripts/image-check.sh` (POSIX sh, `set -eu`) that does exactly that comparison for a given image tag and binary
   path, prints the size, and exits non-zero naming the first difference. A trace or state path that differs only because the
   two runs write to different places is normalised, and the script says so.
3. Fix in the `Dockerfile` (or its build args) only what step 1 shows the image gets wrong. If nothing is wrong, say so.
4. `Makefile`: an `image-check` target that builds the image and runs the script. `docs/INSTALL.md`: a short "Container image"
   section with the two commands and what the check proves.
5. **Negative control:** make the image's fixture differ by one byte (a throwaway build), run the script, paste its red; restore.
6. Nothing else: no Go code changes, `AGENTS.md` unchanged. Never push an image.

### ACCEPTANCE
1. `git log --oneline 2b8fd4a..HEAD` shows only this wagon's commits; paste it.
2. The four outputs and the size from step 1; the script's green run; the negative control's red.
3. `go vet ./...` clean; `go test -count=1 ./...` green except the known `internal/pipeline` mode test, named.
4. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████` in the worktree, except the files this order names. Built
images are tagged `proven:w3*` and removed at the end (`docker image rm`). Never run `proven scan` inside the checkout.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their exit codes, the red. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 3 |
