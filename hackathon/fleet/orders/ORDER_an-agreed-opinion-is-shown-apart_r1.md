# ORDER — an-agreed-opinion-is-shown-apart · round 1

- change: an-agreed-opinion-is-shown-apart
- round: 1
- base: main @ 656dbcb
- executor: cursor on runner-cursor
- written: 2026-10-24T08:35:31Z
- task: a finding both reviewers name at the same place with no evidence is shown apart as an opinion, never counted or failing (ADR-0041)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — an-agreed-opinion-is-shown-apart
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W37
- review: master-verify-only
- repo: proven
- base: main @ 656dbcb
- branch: fix/an-agreed-opinion-is-shown-apart
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: a finding both reviewers name at the same place with no evidence is shown apart as an opinion, never counted or failing (ADR-0041)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[an-agreed-opinion-is-shown-apart] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.


### THE OWNER'S WORDS
2026-10-23: "we fit the big-tech trend, we must show more"; the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0041, docs/adr/ADR-0041-agreed-by-two-models-is-shown-apart.md) — read it whole:
1. A finding that both reviewers name independently, at the same place, and that no evidence proves, is shown in a
   separate section: "Agreed by two models, not proven".
2. It is never mixed with proven findings: it does not fail a pipeline, does not count in the totals, and carries both
   models' reasons word for word.
3. One model alone is not enough: a claim only one reviewer makes stays dropped, as before.
4. JSON and SARIF mark these with their own level (`note`), so a machine never reads them as proven.

### MASTER'S READINGS (claims, measured by the master 2026-10-24 — re-measure, you may refute them)
- The second reviewer (W36, ADR-0037) is on main: each finding carries `Verdicts` from both reviewers and `Disagree`;
  `review.Result` has `Reviewers` and `Disagreements` (`internal/review`). The text shell renders `ReviewerStatements`
  only; the JSON/SARIF writers may not exist at your base — if they do not, put the opinion in the result model and the
  text output and say so.
- "Same place" = same path and an overlapping line range; decide the rule, write it in a comment naming ADR-0041.

### WHAT TO BUILD
1. In the result: a separate list of agreed, unproven opinions (both reviewers' reasons verbatim), never in Findings,
   never in the proof-level totals, never affecting the exit code.
2. Text output: the section "Agreed by two models, not proven"; JSON/SARIF level `note` where those writers exist.
3. A recorded sample scenario with one agreed opinion and one single-reviewer claim (dropped); tests for both and for the
   exit code being unchanged.

### ACCEPTANCE
1. Reproduce first: run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green.
3. **Negative control (law #27):** count the opinion in the totals → its test red; restore. Name the mutated field.

### WHAT IS NOT IN SCOPE
New reviewers, new proof kinds.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it;
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code. Commit subjects `[an-agreed-opinion-is-shown-apart] <area>: <what>`, the area ONE token, AT MOST 72
CHARACTERS IN TOTAL (the merge refuses anything else).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `██████████████████████████████████████████████` in the worktree, except the files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you
end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
