# ORDER — cli-over-the-sample-reviewer · round 12

- change: cli-over-the-sample-reviewer
- round: 12
- base: main @ 2a1ea87
- executor: cursor on runner-cursor
- written: 2026-10-22T15:02:30Z
- task: round 7 — make the git fixture tests independent of the host's git configuration and hooks; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 7
- review: master-verify-only
- repo: proven
- base: main @ 2a1ea87
- branch: fix/cli-over-the-sample-reviewer
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 60m
- task: round 7 — make the git fixture tests independent of the host's git configuration and hooks; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[cli-over-the-sample-reviewer] <area>: <what>`.
No AI attribution anywhere. After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-22 (translated): "Proven's coding runs on Cursor seats." "Every review of a diff is two independent reviews, run in parallel."

### WHY THIS ROUND (codex review of round 6, 2026-10-22 ~15:00Z; claims to test, law #44)
Your round 6 tip is 6f30e14; the master re-verified it green on macm5. The codex review (run-20261022-145118-84c5d632,
█████) returned `reject`, quoted verbatim:
> `internal/cli/run_test.go:742-746`, `:755`, `:765` create fixture commits named `root`, `advance base`, `feature work`,
> and `only`. The inherited commit-message hook refuses these subjects. Measured `go test -count=1 ./...`: `internal/cli`
> FAIL; four tests (`TestCIBaseRefWhenTheEventHasNoSHA`, `TestCIWorkspaceIsTheTarget`, …) never reach their assertions.
Its condition: "Reject until four Git fixture tests use accepted subjects and `go test -count=1 ./...` passes with the
active hook. Keep assertions contrasting fork point, base tip and head; rerun `TestCIMergeBaseIsTheForkPoint` as the
negative control. Do not suppress the hook."
The master's reading (measured on █████): the refusing hook is the HOST's global `core.hooksPath`, which reaches every
`git init` in `t.TempDir()`. A test that a user's own global hook or config can break is not hermetic; renaming subjects
only fits one host's hook. So this round makes the fixture git commands independent of the host's git configuration —
the reviewed repository's hooks are untouched, and the next review judges whether its condition is met.

### WHAT TO DO
1. **First, reproduce:** run the four tests with a global hook that refuses every commit (a temporary
   `GIT_CONFIG_GLOBAL` file whose `core.hooksPath` points at a `commit-msg` that exits 1) and paste the failures. If they
   pass, say so in `- master-was-wrong:` and stop with `- state: blocked`.
2. In `internal/cli/run_test.go`, make `runGit` run every fixture git command with `GIT_CONFIG_GLOBAL` set to an empty
   file in the test's temp dir, `GIT_CONFIG_NOSYSTEM=1` and `GIT_TERMINAL_PROMPT=0`, keeping the rest of the environment.
   Also give the fixture commits conventional subjects (`test: root`, `test: advance base`, `test: feature work`,
   `test: only`). Assertions stay exactly as they are.
3. A test proving it: with the refusing global hook from step 1 in the environment, the four tests pass; paste it.
4. Nothing else changes; production code, `internal/review` and every other test stay as they are.

### ACCEPTANCE
1. `git log --oneline 6f30e14..HEAD` shows only this round's commits; paste it.
2. `git diff --stat 6f30e14..HEAD` lists only `internal/cli/run_test.go` (and a new test helper file, if you add one).
3. `go vet ./...` clean; `go test -count=1 ./...` green; and green again with the refusing global hook set.
4. `TestCIMergeBaseIsTheForkPoint` red with the merge-base computation reverted (paste the red), green with it.
5. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse HEAD` output
verbatim; never restate HEAD from memory.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `██████████████████████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their counts. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| machine name | 2 |
| private workspace path | 4 |
