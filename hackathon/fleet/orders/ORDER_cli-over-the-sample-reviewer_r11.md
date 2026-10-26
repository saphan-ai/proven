# ORDER — cli-over-the-sample-reviewer · round 11

- change: cli-over-the-sample-reviewer
- round: 11
- base: main @ 2a1ea87
- executor: cursor on runner-cursor
- written: 2026-10-22T13:57:41Z
- task: round 6 — carry W1b onto main @ 2a1ea87 and port main's new text output (unresolved line, same-family sentence, KEV evidence lines) into internal/cli

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 6
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
- task: round 6 — carry W1b onto main @ 2a1ea87 and port main's new text output (unresolved line, same-family sentence, KEV evidence lines) into internal/cli

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

### WHY THIS ROUND (master's measurement, 2026-10-22 ~13:40Z; claims to test, law #44)
Your round 5 tip 83859af (on main @ 6db4272) passed the master's verify: vet clean, `go test -count=1 ./...` 0 FAIL,
`--timeout=9223372037` exit 7, `--timeout=9223372036` exit 0, `PROVEN_TIMEOUT=9223372037` exit 7. But main moved to
2a1ea87 (38 commits) and `git merge-tree --write-tree main 83859af` reports CONFLICT (content) in `cmd/proven/main.go`
and `cmd/proven/main_test.go`. Main's commits there: 02aa89f (evidence-through-zetaris: a KEV query test), ddc00c4 and
b63ffc5 (the-skeptic-and-the-judge). Main's old `printResult` in `cmd/proven/main.go` gained:
- after the `finding <ID> <SEVERITY> <TITLE>` line, a line `  unresolved` when `f.Unresolved` is true;
- before the `trace:` line, the review's same-family sentence: the `text` of the first JSONL line in
  `res.TracePath` whose `kind` is `family` (function `familySentence`), printed when non-empty.
Main's `cmd/proven/main_test.go` gained `TestSampleOutputPrintsUnresolved`, `TestSampleOutputPrintsTheSameFamilySentence`
and KEV expectations (`finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list`, its `SQL:` and `row:` lines).

### WHAT TO DO
1. **First, reproduce:** `git merge-tree --write-tree 2a1ea87 HEAD` and paste the conflicted files. If they are not
   those two, say so in `- master-was-wrong:` and resolve what you actually find.
2. Put your commits on top of main @ 2a1ea87, same branch: `git rebase --onto 2a1ea87 6db4272`. If the git-ref guard
   refuses, `git reset --hard 2a1ea87` and cherry-pick your commits `6db4272..83859af` in order; if both are refused,
   write the refusal verbatim with `- state: blocked` and stop. Do NOT run `git merge`.
3. `cmd/proven/main.go`: keep your thin `main` that calls `cli.Run`; main's `printResult`/`familySentence` must not come
   back there. Port both behaviours into `internal/cli`'s text formatter, byte for byte as main prints them.
4. `cmd/proven/main_test.go`: keep main's two new tests and the KEV expectations with their assertions unchanged,
   calling your formatter or entry point the way your other tests do. If one fails, fix `internal/cli`, not the test.
5. Nothing else changes; `internal/review` stays exactly as on main.

### ACCEPTANCE
1. `git log --oneline 2a1ea87..HEAD` shows your commits (re-created) and no merge commit; paste it.
2. `git diff --stat 2a1ea87..HEAD -- internal/review` is empty; paste it.
3. `go vet ./...` clean; `go test -count=1 ./...` green; counts per package.
4. `scan --sample testdata/fixture/discount` exit 0; `--error` exit 1; `--config=auto` exit 7; `--timeout=9223372037`
   exit 7; the unresolved line and the family sentence appear as main printed them (paste the test names, green).
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
