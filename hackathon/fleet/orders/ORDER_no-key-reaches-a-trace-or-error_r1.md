# ORDER — no-key-reaches-a-trace-or-error · round 1

- change: no-key-reaches-a-trace-or-error
- round: 1
- base: main @ 21af174
- executor: claude-code on macm5
- written: 2026-10-23T11:20:54Z
- task: no configured secret reaches the trace, the result, a refusal reason or an error message, even when a service echoes it back

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — no-key-reaches-a-trace-or-error
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W30
- review: named
- repo: proven
- base: main @ 21af174
- branch: fix/no-key-reaches-a-trace-or-error
- lands-on: main
- backend: claude-code
- model: claude-sonnet-5-5
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 150m
- task: no configured secret reaches the trace, the result, a refusal reason or an error message, even when a service echoes it back

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[no-key-reaches-a-trace-or-error] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### PRIVATE UNTIL MERGED (read first)
These are security findings of the proven-security-review train, PRIVATE until each fix is merged. Proven is a public
repository: **no finding text in a commit message, a code comment, a test name or a doc**. Commit subjects describe the
behaviour plainly (for example `testrun: run the container as an unprivileged user`), never as a defect, a weakness or
an attack. Comments say what the code guarantees and why (ADR-0025, ADR-0005, ADR-0013), not what was wrong.
The reviewers' notes with their failing tests are provisioned at `████████████████████████` — read the sections named
below; copy a test only after renaming it to the property it proves.

### THE PROPERTY THAT MUST HOLD (SECURITY_REVIEW.md, piece 2)
No key reaches the trace, the logs, the reports, a PR comment, an error message or a container; a missing key is named,
never printed.

### WHAT THE REVIEWS MEASURED (at main 90d6951 — CLAIMS for you to re-measure, law #44)
- `internal/pipeline/pipeline.go` ~1016-1052 and `internal/generator/live.go` ~169: a runtime token echoed back inside a
  model receipt is written to the trace and the result. Notes: `REVIEW_NOTES_p2-claude.md`
  (`TestRuntimeTokenEchoedInTheReceiptIsNotWritten`), `REVIEW_NOTES_p2-codex.md`
  (`TestReviewConfiguredSecretNeverReachesTraceFromReceipt`).
- `internal/pipeline/pipeline.go` ~989-997 and `internal/guardrail/guardrail.go` ~53: a guardrail key echoed in the
  safety categories reaches the trace and the refusal reason. Note: `REVIEW_NOTES_p2-claude.md`
  (`TestGuardrailKeyEchoedInTheCategoriesIsNotWritten`).
- `internal/query/settings.go` ~120-127 and ~152: a refused `ZETARIS_ORG_ID` value is printed. Notes:
  `REVIEW_NOTES_p2-codex.md` (`TestReviewOrgIDNeverPrintedOnConfigurationError`,
  `TestReviewOrgIDNeverPrintedWhenServiceEchoesIt`), `REVIEW_NOTES_p2-claude.md` (`TestARefusedSecretIsNamedNotPrinted`).
**Step 1, before any change:** add the tests (renamed to the property) and paste their red output at the base.

### WHAT TO BUILD
1. One redactor built from the configured secrets' VALUES (every secret Proven loads), applied at the writers: the
   trace writer, the result's model and refusal fields, and every error or refusal string that carries text from a
   service. A value is replaced by a fixed marker naming the variable, never a prefix of it.
2. A refused secret-like setting (the org id included) names the variable and the rule, never the value.
3. Short or empty values are not redacted as substrings of everything: say the rule you chose and why in a comment.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. The step-1 tests pass: an echoed token in a receipt, an echoed guardrail key in categories, a refused org id — none
   appears in the trace, the result or any error text.
3. **Negative control (law #27):** bypass the redactor at the trace writer → the receipt test turns red; restore.
### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
The container (W28); the prover command (W29); the CLI; livecheck output beyond the shared redactor.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `█████████████████████████████████████████████` in the worktree, except the
files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 5 |
