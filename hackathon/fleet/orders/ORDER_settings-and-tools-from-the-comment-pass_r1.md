# ORDER — settings-and-tools-from-the-comment-pass · round 1

- change: settings-and-tools-from-the-comment-pass
- round: 1
- base: main @ 6b442ac
- executor: cursor on runner-cursor
- written: 2026-10-23T07:46:19Z
- task: two hard-coded values become named settings (ADR-0013), livecheck and settings-doc share the checks they duplicate, and three weak tests prove their property

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — settings-and-tools-from-the-comment-pass
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W25
- review: master-verify-only
- repo: proven
- base: main @ 6b442ac
- branch: fix/settings-and-tools-from-the-comment-pass
- lands-on: main
- worktree: ████████████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: two hard-coded values become named settings (ADR-0013), livecheck and settings-doc share the checks they duplicate, and three weak tests prove their property

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[settings-and-tools-from-the-comment-pass] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE DEFECTS (the master's helpers' reading, 2026-10-23 — CLAIMS for you to test, law #44)
- **F18:** `internal/limit` `Adapter = 5 * time.Second` (line 16) is hard-coded, not a named setting (ADR-0013).
- **F24:** `internal/osv/check.go` `KEVView = "px_evidence.px_kev_vulns"` (~33) is hard-coded, not a setting (ADR-0013);
  `PROVEN_EPSS_VIEW` (internal/epss) is the model to copy, view-name check included.
- **F9:** `cmd/livecheck` `validView` (~221) duplicates `epss.ValidView` with a different length limit.
- **F11:** `cmd/settings-doc` writes docs/CONFIGURATION.md without `envdoc.Check`; only a test catches a bad cell.
- **F7:** `internal/query/engine.go` `tokenize` (~557) refuses a number as a bad character, so the local engine cannot run
  `SELECT 1 … LIMIT 1`; one condition in it can never matter.
- **F12:** `cmd/livecheck` `TestNoSettingsNamesEveryServiceNotReady`: its no-secret assertion uses an empty environment,
  so it can never fail.
- **F19:** `internal/review` `TestSymlinkedRecordOutsideTheTargetIsRefused` and `TestSymlinkedRecordGlobIsRefused` pass
  when the review returns no error; they should require the refusal.
**Step 1, before any change:** for each item, show it at the base (a failing test, or the line and why). If one does not
hold, write it in `- master-was-wrong:` and leave it out.

### WHAT TO BUILD
1. F18, F24: named settings with defaults equal to today's values, refused when wrong (ADR-0013); registered so
   `go run ./cmd/settings-doc` lists them, and docs/CONFIGURATION.md regenerated.
2. F9: livecheck uses `epss.ValidView`. F11: settings-doc runs `envdoc.Check` before it writes, and refuses.
3. F7: the local engine reads an integer literal; `SELECT 1 … LIMIT 1` runs.
4. F12, F19: the tests can fail: F12 sets a secret and asserts it is not printed; F19 requires the refusal.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. W11's settings drift test passes with the two new settings.
3. **Negative controls (law #27):** F12 — print the secret → its test turns red; F19 — let the symlinked record be read →
   its test turns red; restore both.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
internal/pipeline (another wagon is changing it); any network.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `██████████████████████████████████████████████████████` in the worktree, except the
files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
