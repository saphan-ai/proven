# ORDER — cli-over-the-sample-reviewer · round 9

- change: cli-over-the-sample-reviewer
- round: 9
- base: main @ 6db4272
- executor: cursor on runner-cursor
- written: 2026-10-22T11:44:58Z
- task: round 4 — carry the three W1b commits onto main @ 6db4272 and resolve the three conflicts; no new behaviour

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 4
- review: master-verify-only
- repo: proven
- base: main @ 6db4272
- branch: fix/cli-over-the-sample-reviewer
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 60m
- task: round 4 — carry the three W1b commits onto main @ 6db4272 and resolve the three conflicts; no new behaviour

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

### WHY THIS ROUND (master's measurement, 2026-10-22 ~08:30Z; claims to test, law #44)
Your round 3 (2d600b4) passed the master's verify on macm5: `go vet ./...` clean, `go test -count=1 ./...` green,
`TestCIDefaultConfigIsTheWorkspaceFile` green, `internal/review` unchanged since 0ed349b. But it does not merge:
`git merge-tree --write-tree main 2d600b4` reports CONFLICT (content) in `AGENTS.md`, `cmd/proven/main.go`,
`cmd/proven/main_test.go`. main moved from 0ed349b to 6db4272; the commits that touch those files are 8f1a195
(the-reviewer-honours-the-request: `--exclude=`, `--include=`, `--baseline-commit=`, `--config=` added to
`cmd/proven/main.go`, a review-time `*review.ConfigError` → exit 7, and two tests `TestScanExcludeAllPrintsNoFinding`,
`TestScanMissingBaselineNamesTheCommit`), 701067e/c91f828/24b7e34 (the-prover-in-a-container: AGENTS.md testrun row
and a "Prover container defaults" section) and dc276c8 (the-generator-through-nightshift: AGENTS.md "Configuration"
and "Generator" sections).
Your `internal/cli` already reads all four of those flags (`internal/cli/config.go`, `run.go:68-71,140`), so the
conflicts are expected to be textual, not new behaviour.

### WHAT TO DO
1. Put your three commits (cca7e5c, ae34b74, 2d600b4) on top of main @ 6db4272, on the same branch
   `fix/cli-over-the-sample-reviewer`: `git rebase --onto 6db4272 0ed349b`. If the git-ref guard refuses the rebase,
   do `git reset --hard 6db4272` and `git cherry-pick cca7e5c ae34b74 2d600b4` instead; if both are refused, write the
   refusal verbatim into `_EXEC_STATUS.md` with `- state: blocked` and stop. Do NOT run `git merge`.
2. Resolve each conflict this way:
   - `cmd/proven/main.go`: keep your thin `main` that calls `cli.Run`. main's flag parsing must not come back.
   - `cmd/proven/main_test.go`: keep main's two tests `TestScanExcludeAllPrintsNoFinding` and
     `TestScanMissingBaselineNamesTheCommit` with their assertions unchanged (exit 0 and no `finding ` line for
     `--exclude=*`; exit 2 and the stderr names `does-not-exist`, without `unknown flag`), calling your entry point
     the way your other tests in that file do. If either fails, fix `internal/cli`, not the test.
   - `AGENTS.md`: keep BOTH sides — main's testrun row and its "Configuration", "Generator" and "Prover container
     defaults" sections, and your `internal/cli` / `cmd/proven` rows and "4. The command line" section.
3. A review-time `*review.ConfigError` (returned by `Review`, not by `Open`) exits 7 in the shell, as it did on main.
   If your shell does not do that yet, add it with a test that is red without it.
4. Nothing else changes; `internal/review` stays exactly as on main.

### ACCEPTANCE
1. `git log --oneline 6db4272..HEAD` shows exactly your commits (re-created) and no merge commit; paste it.
2. `git diff --stat 6db4272..HEAD -- internal/review` is empty; paste it.
3. `go vet ./...` clean; `go test -count=1 ./...` green; counts per package.
4. Rounds 1–3 behaviour unchanged: `scan --sample testdata/fixture/discount` exit 0 with F1 and F2; `--error` exit 1;
   `--config=auto` exit 7; a valid flag beats an invalid variable; `TestCIDefaultConfigIsTheWorkspaceFile` green.

### THIS ORDER CAN BE WRONG (law #44)
Before you resolve anything, run `git merge-tree --write-tree 6db4272 2d600b4` and paste the conflicted files. If the
list is not those three, say so in `- evidence:` and resolve what you actually find.

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
| private workspace path | 4 |
