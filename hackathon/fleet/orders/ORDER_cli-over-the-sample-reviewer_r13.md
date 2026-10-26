# ORDER — cli-over-the-sample-reviewer · round 13

- change: cli-over-the-sample-reviewer
- round: 13
- base: main @ 5d8a8b8
- executor: cursor on runner-cursor
- written: 2026-10-22T16:13:36Z
- task: round 8 — carry W1b onto main @ 5d8a8b8 with main's new review-class output and tests; make the git fixture tests independent of the host's git configuration

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 8
- review: master-verify-only
- repo: proven
- base: main @ 5d8a8b8
- branch: fix/cli-over-the-sample-reviewer
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 60m
- task: round 8 — carry W1b onto main @ 5d8a8b8 with main's new review-class output and tests; make the git fixture tests independent of the host's git configuration

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

### ROUND 7 — NOT DONE (law #28)
run-20261022-150204-cb3b1d9e was killed after ~19 minutes (exit 137) with no commit: it set a refusing global git
config in its own shell and spent the run studying the engine's own `commit-msg` hook. Nothing of round 7 landed;
start from 6f30e14. Your worktree still carries round 7's UNCOMMITTED edit of `internal/cli/run_test.go`: look at it, then
discard it (`git checkout -- internal/cli/run_test.go`) before you begin.

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

### PART A — CARRY ONTO main @ 5d8a8b8 (master's measurement, 2026-10-22 ~16:10Z; law #44)
main moved again: the-reviewer-model-and-the-review-class landed (5d8a8b8) and `git merge-tree --write-tree main
6f30e14` reports CONFLICT in `cmd/proven/main.go` and `cmd/proven/main_test.go`. Main's old `printResult` now begins with:
```go
fmt.Fprintf(w, "review class: %s\n", res.Class)
fmt.Fprintf(w, "review rule: %s\n", res.ClassRule)
if res.SettingsStatement != "" { fmt.Fprintln(w, res.SettingsStatement) }
if res.AuthorModelStatement != "" { fmt.Fprintln(w, res.AuthorModelStatement) }
for _, line := range res.ReviewerStatements { fmt.Fprintln(w, line) }
```
and `cmd/proven/main_test.go` gained `TestScanOpenFamilyNameIsAccepted`, `TestScanMalformedFamilyExits7`,
`TestScanUnknownClassExits7`.
A1. **Reproduce first:** `git merge-tree --write-tree 5d8a8b8 HEAD` and paste the conflicted files.
A2. Rebase your commits onto main @ 5d8a8b8, same branch: `git rebase --onto 5d8a8b8 2a1ea87`; if refused,
    `git reset --hard 5d8a8b8` and cherry-pick `2a1ea87..6f30e14` in order; if both are refused, `- state: blocked`. No merge.
A3. Keep your thin `cmd/proven` main. Port the five lines above into `internal/cli`'s text formatter, first, in that
    order, byte for byte. Keep main's three new tests with their assertions unchanged, calling your entry point; if one
    fails, fix `internal/cli`, not the test.
A4. `internal/review` stays exactly as on main.

### PART B — HERMETIC GIT FIXTURES (the codex review of round 6)
1. **First, reproduce, inside the test process only:** add a test helper that runs the four CI fixture tests' git
   commands under a temporary `GIT_CONFIG_GLOBAL` whose `core.hooksPath` is a temp dir with a `commit-msg` that exits 1
   — set it with `t.Setenv` inside the Go test, never in your shell. Run `go test -count=1 -run 'TestCI' ./internal/cli`
   at 6f30e14 with that helper in place and paste the failures. Do NOT export `GIT_CONFIG_GLOBAL` in your own shell and
   do NOT run or study `saphan hook`: your own worktree's commits go through the engine's hook, and a refusing global
   config in your shell breaks your own commits (round 7 lost its whole run that way).
2. In `internal/cli/run_test.go`, make `runGit` run every fixture git command with `GIT_CONFIG_GLOBAL` set to an empty
   file in the test's temp dir, `GIT_CONFIG_NOSYSTEM=1` and `GIT_TERMINAL_PROMPT=0`, keeping the rest of the environment.
   Also give the fixture commits conventional subjects (`test: root`, `test: advance base`, `test: feature work`,
   `test: only`). Assertions stay exactly as they are.
3. The helper from step 1 stays as a test: with the refusing global hook set by `t.Setenv`, the four tests pass; paste it.
4. Nothing else changes; production code, `internal/review` and every other test stay as they are.

### ACCEPTANCE
1. `git log --oneline 5d8a8b8..HEAD` shows your commits re-created plus this round's, no merge commit; paste it.
2. `git diff --stat 5d8a8b8..HEAD -- internal/review` is empty; paste it. Main's five new tests above are green.
3. `go vet ./...` clean; `go test -count=1 ./...` green, which includes the refusing-hook test from step 1.
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
