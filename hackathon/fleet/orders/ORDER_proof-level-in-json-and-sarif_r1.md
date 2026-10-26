# ORDER — proof-level-in-json-and-sarif · round 1

- change: proof-level-in-json-and-sarif
- round: 1
- base: main @ d4785e2
- executor: cursor on runner-cursor
- written: 2026-10-25T01:46:27Z
- task: Proven already computes how well each finding is proved, but its JSON and SARIF outputs leave it out; they now carry each finding's level of proof and the counts by level, under the names the reviewer core defines, and the docs say so; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — proof-level-in-json-and-sarif
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- review: named
- repo: proven
- base: main @ d4785e2
- branch: fix/proof-level-in-json-and-sarif
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 120m
- task: Proven already computes how well each finding is proved, but its JSON and SARIF outputs leave it out; they now carry each finding's level of proof and the counts by level, under the names the reviewer core defines, and the docs say so; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity.
Commit subjects: `[proof-level-in-json-and-sarif] <area>: <what>`, the area ONE token, at most 72 characters, English
(the merge refuses others). No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHERE YOU START (master, 2026-10-25; claims to test, law #44)
- Measured by the advisor on a fresh clone of main (2026-10-25 01:40Z): the reviewer core sets `review.Finding.ProofLevel` and
  `review.Result.ProofCounts` (`internal/review/review.go`, ADR-0044), yet `proven scan --sample testdata/fixture/discount --json` has
  no proof-level field and no counts, and the SARIF output has none either. The README tells readers every finding names its level.
- The level values are the core's, and are a released contract: `proved-by-test`, `proved-by-tool`, `proved-by-data`,
  `agreed-by-two-models-not-proven` (`review.go:169-186`; ProofCounts keeps the last at 0 — opinions are `Result.Opinions`).
- The writers: `internal/cli/format_json.go`, `internal/cli/format_sarif.go`; the contract table: `docs/OUTPUT_FORMATS.md`. The text
  writer and the Markdown comment are other wagons (W8, W9) — do not touch `format.go`'s text writer or any `format_text*`/
  `format_md*` file.

### WHAT TO DO
1. **Reproduce first:** a test that renders the sample result as JSON and as SARIF and asserts F1 carries `proved-by-test` and the
   dependency finding carries `proved-by-data`; run it on d4785e2 and paste the red.
2. JSON: `findings[].proof_level` and a top-level `proof_counts` object (every level a key, as the core keeps it). SARIF: the level in
   each result's `properties` (`proof_level`) and the counts in the run's `properties` (`proof_counts`). Same values in both.
3. `docs/OUTPUT_FORMATS.md`: the two JSON rows and the SARIF sentence; existing JSON/SARIF goldens updated by exactly these fields
   (paste the golden diff: nothing else may move).
4. Nothing else; `internal/review`, `AGENTS.md` unchanged.

### ACCEPTANCE
1. `git log --oneline d4785e2..HEAD` shows only this wagon's commits; paste it.
2. The red of step 1, pasted; the golden diff, pasted.
3. `go vet ./...` clean; `go test -count=1 ./...` green except the known `internal/pipeline` mode test, named; `gofmt -l` empty.
4. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

### CODE IS WRITTEN TO BE READ (ADR-0027)
Doc comments on new fields; every test says which property it proves. English, plain.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████` in the worktree, except the files this order
names. Never run `proven scan` inside the checkout.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their exit codes, the red. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
