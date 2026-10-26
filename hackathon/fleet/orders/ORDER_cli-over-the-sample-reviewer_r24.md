# ORDER — cli-over-the-sample-reviewer · round 24

- change: cli-over-the-sample-reviewer
- round: 24
- base: main @ 90d6951
- executor: cursor on runner-cursor
- written: 2026-10-23T09:44:57Z
- task: round 13 — pay the two reviews of round 12: a bad boolean flag value is refused with exit 7 and its allowed values; `proven ci` with an explicit baseline requires a resolved head; the owed tests; carry W19's scenario test onto cli.Run; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 13
- review: master-verify-only
- repo: proven
- base: main @ 90d6951
- branch: fix/cli-over-the-sample-reviewer
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 60m
- task: round 13 — pay the two reviews of round 12: a bad boolean flag value is refused with exit 7 and its allowed values; `proven ci` with an explicit baseline requires a resolved head; the owed tests; carry W19's scenario test onto cli.Run; nothing else

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

### WHY THIS ROUND (the two reviews of round 12, both `accept-with-conditions`; carried verbatim, law #44)
Your round 12 tip is b11e583; both reviews judged 09a52cd (b11e583 + the master's public status file
`fleet/status/STATUS_cli-over-the-sample-reviewer_r12.md`). The branch tip is now **19f292e**: the master absorbed main
@ 90d6951 (W19 of the reviewer train, `cmd/proven/scenario_test.go`) into it with a clean merge. At 19f292e
`cmd/proven` tests do NOT compile: `scanScenario` still calls the `run(...)` that W1b removed. That is step 0 below.

claude-code review (run-20261023-075149-65121aae):
> C1 (F1): a bad value on a boolean flag is refused with exit 7 and the allowed values `true, false`, the same as its
> variable. Add the negative-control rows to `TestBadValuesAreRefused`.
> Owed: tests for the `rule` line (it belongs) and for sample mode reading no key; one dead GITHUB_WORKSPACE branch.

codex review (run-20261023-074949-92409de6):
> OWED — `internal/cli/run.go:263-265`, `internal/cli/run_test.go:769-773`: when `ci --baseline-commit=from-flag` has no
> event and no `GITHUB_SHA`, it proceeds with `HeadCommit == ""` and exit 0; `internal/review/review.go:61` defines an empty
> head as the working tree. A CI invocation can therefore review a different checkout state from the requested PR head
> without a refusal. Control: require a resolved head in `ci`, including explicit-baseline mode, or explicitly constrain
> and document that mode; test the empty-head case and a valid event head.
> NOTE — `internal/cli/config.go:321` resolves `SAMPLE_MODE` and other non-secret CLI settings before the sample reviewer
> opens. [...] the no-secret/no-network sample claim needs its own instrument, not the passing suite alone.

### WHAT THE MASTER READ (2026-10-23 ~08:15Z, at 09a52cd, without building; claims to test, law #44)
- C1: `boolPresence.Set` (`internal/cli/config.go:70-77`) returns `strconv.ParseBool`'s error as is, so the flag package
  refuses it and the exit is 2. The variable path, `parseBoolSetting` (`config.go:494`), already refuses with
  `refuse(field, v, "value must be true or false", []string{"true", "false"})`.
- codex OWED: in `applyCI` (`internal/cli/run.go`), `if explicitBaseline { return nil }` comes BEFORE the
  `if cfg.head == ""` check, so an explicit baseline skips the head requirement.
- Dead branch: `applyCI` sets `cfg.target` from `GITHUB_WORKSPACE` (`run.go:228`); `parseArgs` already does it for `ci`
  (`config.go:265-270`), so the `run.go` branch can never change anything.

### WHAT TO DO
0. **Carry W19 (first):** in `cmd/proven/scenario_test.go`, `scanScenario` calls `run(args, &stdout, &stderr, os.Getwd,
   getenv)` with a `getenv` that fails the test on ANY key ("sample mode read %s"). Make it call
   `cli.Run(args, getenv, &stdout, &stderr, nil)` (import `internal/cli`; the test already `t.Chdir`s, so no `getwd`).
   `cli.Run` resolves the documented CLI variables of ADR-0020 §1 (`PROVEN_*`, `SAMPLE_MODE`, …) at load, so keep W19's
   property honestly: the wrapped `getenv` returns "" for exactly those documented names and fails the test on every
   other key. Name the list in a comment with its source. The four W19 scenario tests (refuted, unsafe prompt, OSV Go
   module, unknown author) are green after the change; paste their names and results.
1. **Reproduce first, in tests, before any change:** `proven scan --sample=maybe` and `proven ci --baseline-commit=<sha>`
   with no `GITHUB_EVENT_PATH` and no `GITHUB_SHA`; paste what each returns today.
2. A bad value on every boolean flag exits 7 and names `true, false`, the same refusal as its variable. Rows in
   `TestBadValuesAreRefused` for each boolean flag; red with the fix reverted (paste one red).
3. `proven ci` requires a resolved head in every baseline mode: with an explicit baseline and no head it exits 2 and
   names what is missing; with an explicit baseline and a head from the event or `GITHUB_SHA` it proceeds. Tests for both;
   `TestCIMergeBaseIsTheForkPoint` and the round-12 missing-baseline tests stay green.
4. A test that the text output carries one `  rule <rule-id>` line per finding (it belongs; the reviewer said so).
5. A test in `internal/cli` that sample mode reads no secret: `cli.Run` with `--sample`, a recording `getenv` that
   answers the documented CLI variables (ADR-0020 §1) and decoy secret keys of the adapters (name the ones you chose);
   the decoys are never asked for, and the run touches no network (say how you proved it). It may share the documented
   list with step 0.
6. Remove the dead `GITHUB_WORKSPACE` branch in `applyCI`, with `TestCI…` that cover `GITHUB_WORKSPACE` still green.
7. Nothing else changes; same branch, new commits on top of 19f292e; `internal/review` and `AGENTS.md` stay exactly as on
   main @ 90d6951.

### CODE IS WRITTEN TO BE READ (ADR-0027)
Every line you add or change carries: a doc comment on every exported name (go doc form); a comment saying WHY at every
step a reader would stop at (a guard, an order that matters, a refusal), never one that restates the code; the decision
named where code carries one out (e.g. `ADR-0020 §2: a configuration refused at load exits 7`); every test says which
property it proves. English, plain. No AI attribution, no internal host names, paths or credential names, no
commented-out code.

### ACCEPTANCE
1. `git log --oneline 19f292e..HEAD` shows only this round's commits; paste it.
2. `git diff --stat 90d6951..HEAD -- AGENTS.md internal/review` is empty; paste it.
3. `go vet ./...` clean; `go test -count=1 ./...` green; counts per package.
4. The two reds of steps 2 and 3, each with its fix reverted; paste them.
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
