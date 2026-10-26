# ORDER — cli-over-the-sample-reviewer · round 18

- change: cli-over-the-sample-reviewer
- round: 18
- base: main @ 26343c2
- executor: cursor on runner-cursor
- written: 2026-10-23T04:03:41Z
- task: round 12 — `proven ci` refuses when it has no pull-request baseline; take W1b's edit of AGENTS.md out; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 12
- review: master-verify-only
- repo: proven
- base: main @ 26343c2
- branch: fix/cli-over-the-sample-reviewer
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 60m
- task: round 12 — `proven ci` refuses when it has no pull-request baseline; take W1b's edit of AGENTS.md out; nothing else

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

### WHY THIS ROUND (codex review of round 11, run-20261023-024441-d48a2233; carried verbatim, law #44)
Your round 11 tip is d488a0f; the master re-verified it green and it merges cleanly with main @ 859e4bc. The codex review
returned `reject` with three BLOCKING findings; two are yours, quoted verbatim:
> 1. BLOCKING — `internal/cli/run.go:243–257`. Trigger: `proven ci` with no `GITHUB_EVENT_PATH`, `GITHUB_BASE_REF`, or
> explicit baseline, or with an event lacking a base. `applyCI` returns nil with an empty baseline.
> `internal/review/review.go:60` defines that value as the whole target, and `internal/review/decide.go:64` only takes the
> change path when the baseline is nonempty. The command can report success for a review other than the pull-request diff
> promised by ADR-0020 §1. Remedy: require a u[sable baseline] …
> 2. BLOCKING — `AGENTS.md:67–82` changes an owner-owned file. `AGENTS.md` §3 says never change that file, and this diff
> adds a command-line chapter to it. Remedy: the owner decides and makes any change to this file; remove this edit from
> the executor's branch.
> Closure: "demonstrate `ci` exits 2 on absent pull-request context while preserving the fixture merge-base path."
(The third finding — public copies under `fleet/status/` and `fleet/orders/` — is the master's, not yours.)
Master-was-wrong, recorded: round 4's order told you to keep "both sides" of `AGENTS.md`; that was the master's error.

### WHAT TO DO
1. **Reproduce first:** run `proven ci` (through your entry point, in a test) with no `GITHUB_EVENT_PATH`, no
   `GITHUB_BASE_REF` and no `--baseline-commit`, and with an event that has no base; paste what it returns today.
2. `proven ci` with no usable baseline (none of: an explicit `--baseline-commit`, the event's base SHA, `GITHUB_BASE_REF`
   resolvable to a merge base) exits 2 and names what was missing; it never falls through to reviewing the whole
   checkout. `proven scan` is unchanged (a whole-target scan stays legal there). Tests: each missing-context case exits 2
   and names the cause — red with the check removed (paste one red); `TestCIMergeBaseIsTheForkPoint` still green.
3. Remove every W1b change to `AGENTS.md`: `git diff 26343c2 HEAD -- AGENTS.md` must be empty after this round.
4. Nothing else changes; same branch, new commits on top of d488a0f; `internal/review` stays exactly as on main.

### ACCEPTANCE
1. `git log --oneline d488a0f..HEAD` shows only this round's commits; paste it.
2. `git diff --stat 26343c2..HEAD -- AGENTS.md internal/review` is empty; paste it.
3. `go vet ./...` clean; `go test -count=1 ./...` green; counts per package.
4. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

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
