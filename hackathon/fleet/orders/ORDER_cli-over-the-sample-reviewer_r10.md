# ORDER — cli-over-the-sample-reviewer · round 10

- change: cli-over-the-sample-reviewer
- round: 10
- base: main @ 6db4272
- executor: cursor on runner-cursor
- written: 2026-10-22T13:22:55Z
- task: round 5 — refuse a --timeout / PROVEN_TIMEOUT that overflows time.Duration, with tests at the boundary; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 5
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
- task: round 5 — refuse a --timeout / PROVEN_TIMEOUT that overflows time.Duration, with tests at the boundary; nothing else

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

### WHY THIS ROUND (codex review of round 4, 2026-10-22 12:58Z; claims to test, law #44)
Your round 4 tip is a4eae19 (on main @ 6db4272); the master re-verified it green. The codex review
(run-20261022-125403-ad2a62a1) returned `reject` with ONE blocking finding, quoted verbatim:
> internal/cli/config.go:462 and internal/cli/run.go:63. parseTimeout accepts any nonnegative int; multiplication by
> time.Second overflows above 9,223,372,036 seconds. Measured: go run ./cmd/proven scan --sample --timeout=9223372037
> testdata/fixture/discount returned exit 0, printed failed services: none and review partial: the deadline was
> reached, and printed no findings.
Its condition, verbatim: "Validate timeout seconds before conversion to time.Duration; test the largest valid second
and the next second for both --timeout and PROVEN_TIMEOUT. Repeat affected fixture-based checks with their commits
permitted."
It also confirmed the earlier repairs still hold (opener panic recovery run.go:130, precedence config.go:348, CI merge
base run.go:257).

### WHAT TO DO
1. **First, reproduce:** on a4eae19 run `go run ./cmd/proven scan --sample --timeout=9223372037 testdata/fixture/discount;
   echo $?` and paste it. If it does NOT show the defect, say so in `- master-was-wrong:` and stop with `- state: blocked`.
2. In `internal/cli`, refuse a timeout whose seconds do not fit `time.Duration` (largest valid: 9223372036) as a
   configuration error: exit 7, and the message names the largest value accepted. Same rule for `--timeout` and
   `PROVEN_TIMEOUT`; flag over variable over default stays as it is.
3. Tests: 9223372036 accepted and 9223372037 refused (exit 7), each for the flag and for the variable — red when the
   check is removed (paste one red).
4. Run the CI merge-base tests (`TestCIMergeBaseIsTheForkPoint` and any other test that commits into a fixture) and
   paste their result; if your environment refuses the fixture commits, paste the refusal verbatim.
5. Nothing else changes. Same branch `fix/cli-over-the-sample-reviewer`, new commits on top of a4eae19; no rebase,
   no merge. `internal/review` stays exactly as on main.

### ACCEPTANCE
1. `git log --oneline a4eae19..HEAD` shows only this round's commits; paste it.
2. `git diff --stat 6db4272..HEAD -- internal/review` is empty; paste it.
3. `go vet ./...` clean; `go test -count=1 ./...` green; counts per package.
4. `scan --sample testdata/fixture/discount` exit 0 with F1 and F2; `--error` exit 1; `--config=auto` exit 7;
   `--timeout=9223372037` exit 7; `--timeout=9223372036` exit 0.
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
| private workspace path | 4 |
