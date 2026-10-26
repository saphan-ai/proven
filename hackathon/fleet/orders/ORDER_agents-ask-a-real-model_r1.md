# ORDER — agents-ask-a-real-model · round 1

- change: agents-ask-a-real-model
- round: 1
- base: main @ 0e2e0c1
- executor: cursor on runner-cursor
- written: 2026-10-22T06:59:11Z
- task: the six agents ask a real model in words it can follow and read its answer tolerantly, and a bad answer degrades the review by name instead of ending it

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — agents-ask-a-real-model
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W2b
- review: master-verify-only
- repo: proven
- base: main @ 0e2e0c1
- branch: fix/agents-ask-a-real-model
- lands-on: main
- worktree: ███████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: the six agents ask a real model in words it can follow and read its answer tolerantly, and a bad answer degrades the review by name instead of ending it

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[agents-ask-a-real-model] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link. After EVERY
commit, run `git log -1 --format=%B` and amend away any `Co-authored-by:` line the tool added; in your status paste
`git log --format=%B 0e2e0c1..HEAD | grep -ci co-authored` — it must print `0`.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### WHAT THE MASTER MEASURED (2026-10-22, the owner's live run through Nightshift; claims to test, law #44)
The live generator reached Nightshift and got an answer; the review then stopped with
`review: planner: response is not a risk list: invalid character ' ' after decimal point in numeric literal` (exit 2).
Master's reading of the code: each agent's prompt is a sample key (`internal/agent/planner/planner.go`: `"name the risks\n"`
plus the files) with no instruction on the answer's form, and the planner, investigator, prover and skeptic parse the answer
with strict `json.Unmarshal`; any parse error ends the review.

### THE DECISIONS (pasted)
- ADR-0001: a failing service degrades the review, and the review says which service failed. It never hangs.
- ADR-0002 point 2: a finding with no evidence is dropped, not softened.
- ADR-0018 point 2: every agent calls models only through the generator.

### WHAT TO BUILD
1. Each agent's live prompt states its task and the exact JSON its answer must be (fields and types, one example), and
   asks for that JSON only. The sample generator keeps answering the sample fixture as today (keep its keys working; if a
   prompt's text changes, its recorded sample answer follows it).
2. A tolerant reader shared by the agents: take the JSON out of a fenced code block or surrounding prose; on a parse
   failure, ask the model ONCE to repair its answer to the schema; a second failure is a named, degraded outcome
   (`<agent>: the model's answer did not follow the schema`) in `Result.FailedServices` — the review completes, never exit 2.
3. Tests: answers in a code fence, with prose around them, and with a trailing comma → read; a non-JSON answer twice →
   degraded by name, review completes, F3-style findings without evidence still dropped.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. `proven scan --sample testdata/fixture/discount` prints exactly what it printed before (F1, F2, not F3). Paste it.
3. **Negative control (law #27):** a generator that answers prose twice → the test proves the review completes with the
   agent named as failed; with the degrade path removed, the same test turns red.

### THIS ORDER CAN BE WRONG (law #44)
If the parse error does not reproduce with a prose answer at your base, write it in `- master-was-wrong:` and stop that item.

### WHAT IS NOT IN SCOPE
- Any call to the real Nightshift (the master runs the live review after the round).
- The seam types; the shell; the generator transport (W2); Zetaris (W5); the Skeptic's model family (W6).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `█████████████████████████████████████` in the
worktree. Nothing outside the worktree; nothing committed unless this order names it.

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
