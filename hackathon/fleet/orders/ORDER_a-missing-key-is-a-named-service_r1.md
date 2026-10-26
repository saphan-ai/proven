# ORDER — a-missing-key-is-a-named-service · round 1

- change: a-missing-key-is-a-named-service
- round: 1
- base: main @ 21af174
- executor: cursor on runner-cursor
- written: 2026-10-23T11:13:15Z
- task: the two BLOCKING findings of the train tester: an empty secret makes its service down and named instead of refusing the live review, a down guardrail refuses every model call, and sample mode is proven to open no network connection

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-missing-key-is-a-named-service
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W27
- review: master-verify-only
- repo: proven
- base: main @ 21af174
- branch: fix/a-missing-key-is-a-named-service
- lands-on: main
- worktree: ████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 150m
- task: the two BLOCKING findings of the train tester: an empty secret makes its service down and named instead of refusing the live review, a down guardrail refuses every model call, and sample mode is proven to open no network connection

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-missing-key-is-a-named-service] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23 (the owner's ruling on the tester's finding, now ADR-0013 point 4a and ADR-0005 point 5 on main):
*"An empty secret makes its service down and named in failed services; a wrong non-empty value still refuses. BUT a
guardrail that is down must never let a model call through unchecked (ADR-0005): with the guardrail key empty in live
mode, every model call is refused and the trace says why; only PROVEN_GUARDRAIL_MODE=off turns it off on purpose. W27
carries a test for that: empty guardrail key, live mode, zero unchecked model calls."*

### THE DEFECTS (the train tester's findings, run on main f5d26c7 — CLAIMS for you to test, law #44)
1. **Live, no keys:** `go run ./demo/live/review <a scratch git repo>` with no secrets set stops in `review.Open`
   (`internal/review/open.go`, the `if !opts.Sample` block): `configuration refused: ███████████████████████="": the live
   generator needs a runtime token`, program status 7. No service is named; the review never reads the repository.
   The secrets: `███████████████████████` (generator, Nightshift), `███████████████` (query, Zetaris),
   `████████████████████` (guardrail). Each `Load` (`internal/generator`, `internal/query`, `internal/guardrail`) refuses
   an empty secret today.
2. **Sample mode, no network proof:** W1's acceptance item 6 asks for a test that fails on any dial in sample mode; none
   exists. A green suite does not prove it, because nothing fails when a dial is added.
**Step 1, before any change:** reproduce defect 1 in a test (live mode through `review.Open` + `Review`, no secrets, a
scratch git repo, the test seams of `review.Options`) and paste the red output. If it does not fail, write that in
`- master-was-wrong:` and stop with `- state: blocked`.

### WHAT TO BUILD
1. **ADR-0013 point 4a:** in live mode an EMPTY secret makes its service down (`service.Down` with the service's name and
   a reason that names the variable, never its value); the review goes on with the other services and lists the down one
   in `FailedServices`. A NON-EMPTY wrong value (malformed, too long, a line break) is still refused before the review,
   as today.
2. **ADR-0005 point 5:** with `████████████████████` empty in live mode (and the guardrail mode not `off`), the guardrail
   is down and EVERY model call is refused before it reaches the generator; the trace says why (one line per refusal or
   one line for the run, your choice, say which); the review names the guardrail in its failed services. Only
   `PROVEN_GUARDRAIL_MODE=off` turns the guardrail off (point 3, unchanged). An empty key is never read as off.
3. With the generator down, the Planner names Nightshift and the review is empty, as a Planner whose service is down
   already does; with Zetaris down, query evidence is unavailable and named.
4. **Sample mode opens no network:** a test that runs a full sample review with every dial made to fail the test (for
   example `http.DefaultTransport` and `net.DefaultResolver` replaced in the test, plus a check that no live client is
   built in sample mode). Say in its comment what it can and cannot catch.
5. docs/CONFIGURATION.md (regenerate if the settings page changes), docs/INSTALL.md and demo/live/README.md: one sentence
   each on what an empty key does.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. The step-1 test passes: no secrets, live mode, a scratch git repo with a tracked symlink → the review completes,
   names nightshift, zetaris and the guardrail in its failed services, and names the skipped symlink.
3. **The owner's test:** guardrail key empty, live mode, a stub generator that COUNTS its calls, every other secret set →
   the generator's call count is ZERO, every refusal is in the trace, and the guardrail is in failed services.
4. A test: `PROVEN_GUARDRAIL_MODE=off` with the key empty → model calls go through, and the trace says the run had no
   guardrail.
5. A test: each secret set to a non-empty wrong value (for example with a line break) → still refused before the review.
6. The sample no-dial test from item 4 passes.
7. **Negative controls (law #27):** (a) let an empty guardrail key behave as `off` → test 3 turns red; (b) add one real
   dial to the sample path (a scratch edit) → test 6 turns red. Restore both.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
The CLI (`cmd/proven`, `internal/cli` — another train: `proven scan` without `--sample`); any real network call; real keys.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `██████████████████████████████████████████████` in the worktree, except the
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
| private credential or variable name | 5 |
| private workspace path | 4 |
