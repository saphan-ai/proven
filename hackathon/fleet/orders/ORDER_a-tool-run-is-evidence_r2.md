# ORDER — a-tool-run-is-evidence · round 2

- change: a-tool-run-is-evidence
- round: 2
- base: fix/a-tool-run-is-evidence @ 94a5b2c
- executor: cursor on runner-cursor
- written: 2026-10-24T03:21:28Z
- task: round 2 — a Docker-gated real test of the race tool, so the master can run it on a host with Docker

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-tool-run-is-evidence
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W40
- review: master-verify-only
- repo: proven
- base: fix/a-tool-run-is-evidence @ 94a5b2c
- branch: fix/a-tool-run-is-evidence
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 90m
- task: round 2 — a Docker-gated real test of the race tool, so the master can run it on a host with Docker

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-tool-run-is-evidence] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.


### WHY THIS ROUND
Round 1 (94a5b2c) was master-verified on 2026-10-24: build, vet and tests green; the negative control on
`Introduced`'s `presentOnBase` went red and back. Acceptance 3 of round 1 ("the master runs a real run") has nothing to
run: no test starts the container. Your seat has no Docker; the master runs the test you write on a host that has it.

### WHAT TO BUILD
1. One test in `internal/tool`, skipped unless `PROVEN_TEST_DOCKER=1` (the same gate as W28's
   `TestWrittenGoTestRunsInTheContainer` in `internal/testrun` — copy its skip and container setup), that runs the real
   race tool in the Prover's locked-down container on two tiny Go trees written by the test: a base with no race and a
   change that adds a data race. It asserts a finding proved by a tool on the change and none on the base, and that the
   recorded command and tool version are filled.
2. Its comment names the property and ADR-0040.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./internal/tool`: green; the new test reports SKIP without the gate.
2. Say in your status the exact command the master runs on the Docker host.
3. **Negative control (law #27):** none possible on your seat (no Docker); write UNRUN and why. The master runs it.

### COMMENTS FOR THE JUDGES (ADR-0027)
Every test says which property it proves; English, plain; no AI attribution, host names, internal paths or credential
names.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████` in the worktree, except the files this
order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you
end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
