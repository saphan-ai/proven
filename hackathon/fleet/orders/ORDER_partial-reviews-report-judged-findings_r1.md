# ORDER — partial-reviews-report-judged-findings · round 1

- change: partial-reviews-report-judged-findings
- round: 1
- base: main @ 5075397
- executor: cursor on runner-cursor
- written: 2026-10-22T04:45:32Z
- task: a review cut by its deadline reports only findings the Judge kept, and the trace carries the full evidence and refutation of every finding

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — partial-reviews-report-judged-findings
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W1d
- review: master-verify-only
- repo: proven
- base: main @ 5075397
- branch: fix/partial-reviews-report-judged-findings
- lands-on: main
- worktree: ██████████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 90m
- task: a review cut by its deadline reports only findings the Judge kept, and the trace carries the full evidence and refutation of every finding

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[partial-reviews-report-judged-findings] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link. After EVERY
commit, run `git log -1 --format=%B` and amend away any `Co-authored-by:` line the tool added; in your status paste
`git log --format=%B 5075397..HEAD | grep -ci co-authored` — it must print `0`.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### THE DECISIONS (pasted)
- ADR-0002 point 2: a finding with no evidence is DROPPED, not softened.
- ADR-0018 point 3: every engine writes append-only JSONL to the run's `logs/`; the review's trace IS that JSONL, so a reader
  follows each finding from claim to verdict.

### THE FINDINGS (retroactive codex review of W1, verdict reject, 2026-10-22, pasted; line numbers are at W1's tip and may have moved)
**BLOCKING — Partial reviews publish unevidenced findings.** `internal/pipeline/pipeline.go:79-88` builds candidates before judging them; deadline returns at `:91-93`, `:112-114`, `:135-137` and `:157-159` call `state.partial`, which calls `writer.Write` at `:245-249`. `internal/agent/writer/writer.go:18-24` accepts every candidate whose `Dropped` default remains false; only `internal/agent/judge/judge.go:8-17` enforces evidence, and it has not run. A scratch Go reproduction passing an unjudged F3 candidate into the actual writer printed `partial=true findings=1 first=F3 evidence=0`. A deadline after claims can therefore return F3 with an empty evidence list, violating `internal/review/review.go:76-94`. Remedy: apply the evidence predicate on every partial-return path (including mid-agent cancellation), trace a verdict for each returned candidate, and add a cancellation-after-planning test that asserts every emitted finding has qualifying evidence.
**BLOCKING — The append-only trace omits the proof and refutation it claims to preserve.** `internal/pipeline/pipeline.go:266-285` reduces F1 test evidence to path/name and F2 query evidence to SQL plus a row count, discarding the recorded test outcomes/output, query columns and rows. `:140-150` records only `Refutation.Reason`, discarding `Attempted` and `Refuted`. The sample scan's JSONL contains a test name and `rows=1`, not the actual result row or outcomes. `internal/trace/trace.go:27-56` does use O_APPEND and propagates Write/Sync errors; that does not recover missing event data. Remedy: persist structured evidence and refutation values with their finding id, and assert JSONL round-trip for the F1 outcome and F2 row. Never claim full replay from the current trace.

### WHAT TO BUILD
1. A review cut by its deadline (any of the pipeline's partial returns) reports only findings the Judge has kept. A
   candidate the Judge has not judged is never in the result; the trace records it as `not judged: deadline`.
2. The trace carries, per finding: the full test evidence (path, name, command, fails-on-change, passes-on-base, output),
   the full query evidence (SQL, engine, snapshot, columns, rows), and the full refutation (attempted, refuted, reason).
3. Tests: a deadline after planning (and one after each later stage) → no unjudged finding in the result; the trace
   events for F1, F2 and F3 carry the fields of item 2.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. **Negative control (law #27):** remove the partial-path filter; the deadline test turns red. Paste the red line, then restore.
3. Paste the trace lines of F2 from `go run ./cmd/proven scan --sample testdata/fixture/discount`.

### THIS ORDER CAN BE WRONG (law #44)
The master's claim, 2026-10-22, not built: (M1) the reviewer's line numbers above still point at the code on your base
(W1c changed the pipeline since). Reproduce finding 1 BEFORE you change anything (a deadline right after planning); if it
does not reproduce, write it in `- master-was-wrong:` with the measurement and stop that item.

### WHAT IS NOT IN SCOPE
- The seam types in `internal/review/review.go`; the shell; the container runner; selection and configuration (W1c).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `████████████████████████████████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict. Read every `saphan`
verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
