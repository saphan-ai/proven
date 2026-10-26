# ORDER — json-sarif-outputs · round 6

- change: json-sarif-outputs
- round: 6
- base: main @ 9303b2d
- executor: cursor on runner-cursor
- written: 2026-10-24T06:44:41Z
- task: W2 round 8 (stream round 6) — a path that walks through a missing directory never collides with another output; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — json-sarif-outputs
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- round: 6
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/json-sarif-outputs
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 120m
- task: W2 round 8 (stream round 6) — a path that walks through a missing directory never collides with another output; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[json-sarif-outputs] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit or a file.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHERE YOU START (master, 2026-10-24; claims to test, law #44)
Your tip \`e446790\` (round 7, subjects rewritten to the git rules by the master; tree as you left it). The master measured it
merged with main fa2a0ee: 41 packages ok, 0 FAIL.

### WHY THIS ROUND (codex review of e446790, verdict **reject**; carried verbatim, law #44)
> BLOCKING — `internal/cli/config.go:576-604` (`outputIdentity`), caller at `internal/cli/config.go:465-470`. Trigger: the two output names `report.json` and `missing/../report.json`, with `missing` absent. The resolver sets `missing = true` after `Lstat(missing)` fails, then collapses `..` lexically and reports both names as one file. The filesystem cannot traverse the absent directory; the second name is not writable. From the scratch directory, the built binary command `proven scan --sample --json-output report.json --sarif-output missing/../report.json` exited 7 and said the flags name the same file. This is a false collision refusal before review. The write-failure contract calls for exit 2. Control: retain the unresolved component until a write, or classify this path as unwritable instead of equating it with a reachable file; add a test that keeps the missing directory absent and asserts exit 2. An unplug that changes only that component to an existing directory should then assert exit 7. The current collision tests do not cover the absent-component-and-parent case.

### WHAT TO DO
1. **Reproduce first:** \`--json-output report.json --sarif-output missing/../report.json\` with \`missing\` absent; paste exit 7.
2. **The rule:** once a component is missing, the filesystem cannot traverse it, so a later \`..\` makes the path unreachable.
   \`outputIdentity\` returns, for such a path, an identity that can never equal another output's (for example the raw
   path marked unreachable), so the load check does not refuse it and the write fails later with exit 2 as for any unwritable
   path. A missing final component without a later \`..\` keeps today's behaviour (identity = resolved prefix + rest).
3. **Tests:** the reviewer's pair exits 2 (write failure), not 7; the alias table still refuses every row with 7; red with
   the rule removed — paste it.
4. Nothing else; new commits on top of e446790; commit subjects \`[json-sarif-outputs] <area>: <what>\`, one-token area, ≤ 72
   characters (the merge refuses others); \`internal/review\` and \`AGENTS.md\` unchanged.

### ACCEPTANCE
1. \`git log --oneline e446790..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat e446790..HEAD -- AGENTS.md internal/review\` is EMPTY.
3. \`go vet ./...\` clean; \`go test -count=1 ./...\` green, counts per package; \`gofmt -l\` empty.
4. The red of step 3; paste it.
5. A \`- master-was-wrong:\` line: \`—\`, or what you measured that contradicts this order.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████` in the worktree. Nothing outside it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their exit codes, the reds. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
