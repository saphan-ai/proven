# ORDER — cli-over-the-sample-reviewer · round 16

- change: cli-over-the-sample-reviewer
- round: 16
- base: main @ 79083a7
- executor: cursor on runner-cursor
- written: 2026-10-23T01:24:58Z
- task: round 10 — carry W1b onto main @ 79083a7; keep main's new OSV expectations in cmd/proven/main_test.go; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 10
- review: master-verify-only
- repo: proven
- base: main @ 79083a7
- branch: fix/cli-over-the-sample-reviewer
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 60m
- task: round 10 — carry W1b onto main @ 79083a7; keep main's new OSV expectations in cmd/proven/main_test.go; nothing else

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

### WHY THIS ROUND (master's measurement, 2026-10-23 ~00:30 local; claims to test, law #44)
Your round 9 (9412b82, on fe146e0) passed the master's verify: vet clean, `go test -count=1 ./...` 0 FAIL, the
docs/SAMPLE_MODE.md run exit 0 with 26 guardrail events in the trace, `review class`/`review rule` printed first.
main has moved to 79083a7. `git merge-tree --write-tree 79083a7 9412b82` reports ONE conflict: `cmd/proven/main_test.go`.
Main's side (dependencies-are-checked-in-osv, 4534506 and 299e617) added, to the sample-output expectations:
`finding osv:evil.example/leftpad ERROR Dependency evil.example/leftpad v1.2.3 is in an OSV affected range`,
`introduced 1.0.0 <= v1.2.3 < fixed 1.2.4`, an `SQL: SELECT id, aliases, ecosystem, package, … FROM px_osvfile.px_osv_pkg …`
line, a `row: GHSA-px-sample-0001 | …` line, `alias CVE-2026-0001 is on the KEV list`, and a check that
`example.com/unused` is not queried.

### WHAT TO DO
1. **Reproduce first:** `git cat-file -t 79083a7` (must print `commit`), then `git merge-tree --write-tree 79083a7 HEAD`; paste
   the conflicted files. If it is not that one file, say so in `- master-was-wrong:` and resolve what you find.
2. `git rebase --onto 79083a7 fe146e0` on the same branch; if refused, `git reset --hard 79083a7` and cherry-pick
   `fe146e0..9412b82` in order; if both are refused, `- state: blocked`. No merge.
3. In `cmd/proven/main_test.go` keep BOTH sides: your entry point and every expectation main added, unchanged. If the
   sample output then lacks a line main expects, port it into `internal/cli`'s text formatter byte for byte (the OSV
   finding lines come from the reviewer; your formatter must print them as main's old `printResult` did).
4. Nothing else changes; `internal/review` stays exactly as on main.

### ACCEPTANCE
1. `git log --oneline 79083a7..HEAD` shows the 9 commits re-created plus anything step 3 needed; no merge commit; paste it.
2. `git merge-base --is-ancestor 79083a7 HEAD` true; `git diff --stat 79083a7..HEAD -- internal/review` empty; paste both.
3. `go vet ./...` clean; `go test -count=1 ./...` green; counts per package.
4. From a directory outside the repository: `proven scan --sample <repo>/testdata/fixture/discount` exit 0 and its
   output contains the OSV finding; paste the output.
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
