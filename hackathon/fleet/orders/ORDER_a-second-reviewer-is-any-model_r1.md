# ORDER — a-second-reviewer-is-any-model · round 1

- change: a-second-reviewer-is-any-model
- round: 1
- base: main @ 5a9f2fe
- executor: claude-code on macm5
- written: 2026-10-24T02:36:45Z
- task: a second reviewer behind any OpenAI-compatible API; with one connected every change is reviewed by both, and their disagreement is shown (ADR-0037)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-second-reviewer-is-any-model
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W36
- review: named
- repo: proven
- base: main @ 5a9f2fe
- branch: fix/a-second-reviewer-is-any-model
- lands-on: main
- backend: claude-code
- model: claude-sonnet-5-5
- quote-usd: 5.00
- cap-usd: 10.00
- wall: 180m
- task: a second reviewer behind any OpenAI-compatible API; with one connected every change is reviewed by both, and their disagreement is shown (ADR-0037)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-second-reviewer-is-any-model] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0037, docs/adr/ADR-0037-*.md) — read it whole; its points 2 to 5 are this wagon:
the second reviewer is named by `PROVEN_SECOND_REVIEWER_BASE_URL`, `██████████████████████████` (secret, never printed)
and `PROVEN_SECOND_REVIEWER_MODEL`; ADR-0009's `special_reviewer` points at it; with it connected every change is
reviewed by both independently, `proven.yaml` may narrow it to chosen classes, security always doubled; disagreement
is shown, never averaged (a block "The reviewers disagree" with each verdict and reason); every review names its
reviewers and whether either is the author's model.

### THE MASTER'S READING (2026-10-24 — CLAIMS to re-measure, law #44)
The doubled review exists (W7b): `internal/review/open.go` ~250 sets `DoubleRun` and `SpecialModel` from
`cfg.SpecialReviewer`; the pipeline runs a second Planner and Skeptic on the special model and merges by
rule_id+path+lines. The second reviewer today reuses the main generator with another model id.

### WHAT TO BUILD
1. A second generator client from the three settings (OpenAI-compatible chat completions), behind the same
   `Generator` interface; every call through the guardrail; the key in the redactor; empty key = not connected (one
   reviewer, as today), a wrong value refused.
2. When connected: both reviewers run independently on every change (the class rule of proven.yaml narrows it; the
   security class is always doubled); their disagreement on a finding is recorded on the review result as both
   verdicts and reasons, never averaged. Evidence still decides what is reported.
3. The result names its reviewers and whether either is the author's model. The shell train renders the disagreement
   block; record the fields it needs in docs/ARCHITECTURE.md.
4. docs/CONFIGURATION.md (regenerate) and docs/LAB.md: how one more key adds the second reviewer.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Recorded-answer tests (httptest): two reviewers that disagree produce both verdicts on the result; one reviewer
   alone is unchanged; the key never appears in any output.
3. **Negative control (law #27):** let the second reviewer see the first's answer → its independence test red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Rendering in internal/cli; ADR-0041 (agreed opinions, W37); any real key.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████████████` in the worktree, except the
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
| private credential or variable name | 1 |
| private workspace path | 4 |
