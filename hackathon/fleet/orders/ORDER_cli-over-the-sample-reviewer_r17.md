# ORDER — cli-over-the-sample-reviewer · round 17

- change: cli-over-the-sample-reviewer
- round: 17
- base: main @ 26343c2
- executor: cursor on runner-cursor
- written: 2026-10-23T02:07:41Z
- task: round 11 — carry W1b onto main @ 26343c2; port main's EPSS lines into internal/cli; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-over-the-sample-reviewer
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1b
- round: 11
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
- task: round 11 — carry W1b onto main @ 26343c2; port main's EPSS lines into internal/cli; nothing else

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

### ROUND 10 — DONE BUT OVERTAKEN
Your round 10 put the nine commits on 79083a7 at db5a6f5, suite green; acceptance 4 was unrun only because the proven
binary is not an allowed command on your seat — the master runs it. Main moved again before it could land.

### WHY THIS ROUND (master's measurement, 2026-10-23 ~01:40Z; claims to test, law #44)
main is 26343c2. `git merge-tree --write-tree 26343c2 db5a6f5` reports CONFLICT in `cmd/proven/main.go` and
`cmd/proven/main_test.go`. Main's side is epss-scores-each-cve-finding (7fa08bc): the old `printResult`, inside each
finding, after its evidence, now prints for every query evidence:
```go
for _, e := range f.Evidence {
    if e.Kind != review.EvidenceQuery || e.Query == nil { continue }
    for _, line := range epss.Phrases(e.Query.Columns, e.Query.Rows) {
        fmt.Fprintf(w, "  %s\n", line)
    }
}
```
(`github.com/saphan-ai/proven/internal/epss`), and `cmd/proven/main_test.go` gained an EPSS test that expects
`EPSS 0.50000 (90th percentile, 2026-10-20)` inside the F2 and F4 sections and none in F1, with a helper `findingSection`.

### WHAT TO DO
1. **Reproduce first:** `git cat-file -t 26343c2` (must print `commit`), then `git merge-tree --write-tree 26343c2 HEAD`.
2. `git rebase --onto 26343c2 79083a7` on the same branch; if refused, `git reset --hard 26343c2` and cherry-pick
   `79083a7..db5a6f5` in order; if both are refused, `- state: blocked`. No merge.
3. Keep your thin `cmd/proven` main; port the EPSS block into `internal/cli`'s text formatter at the same place in each
   finding, byte for byte. Keep main's EPSS test and `findingSection` with assertions unchanged, calling your entry point.
4. Nothing else changes; `internal/review` stays exactly as on main. You may not be able to run the proven binary: the
   tests are the acceptance; the master runs the binary.

### ACCEPTANCE
1. `git log --oneline 26343c2..HEAD` shows the commits re-created, no merge commit; paste it.
2. `git merge-base --is-ancestor 26343c2 HEAD` true; `git diff --stat 26343c2..HEAD -- internal/review` empty; paste both.
3. `go vet ./...` clean; `go test -count=1 ./...` green, including main's EPSS test; counts per package.
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
