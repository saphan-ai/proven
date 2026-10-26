# ORDER — a-review-end-to-end-in-sample-mode · round 3

- change: a-review-end-to-end-in-sample-mode
- round: 3
- base: main @ 58f653f
- executor: cursor on runner-cursor
- written: 2026-10-22T01:03:52Z
- task: the six reviewer agents as the simplest code that works, behind the shell's review seam, so a change in gives a review out in sample mode with no network and no key

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-review-end-to-end-in-sample-mode
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W1
- review: master-verify-only
- repo: proven
- base: main @ 58f653f
- branch: fix/a-review-end-to-end-in-sample-mode
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: the six reviewer agents as the simplest code that works, behind the shell's review seam, so a change in gives a review out in sample mode with no network and no key

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[a-review-end-to-end-in-sample-mode] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████
- ████████████████████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* · *"Explore today, build tomorrow."* On ADR-0025
(where the Prover runs code): *"OK."*

### THE DECISIONS YOU BUILD (pasted from proven `fleet/decisions/`, accepted by the owner)
- **ADR-0001:** every outside service sits behind an adapter: one Go interface per service, a live and a sample
  implementation. No agent calls a service directly. Every adapter call has a timeout and a fallback; a failing service
  degrades the review and the review names it. Keys come only from environment variables. `SAMPLE_MODE=true` is a set
  of sample adapters, not a separate code path: sample mode runs the same agent logic as a live run.
- **ADR-0002:** a finding counts only with evidence: a test that fails on the change and passes on the base, or a query
  and the rows it returned (the SQL is written next to the finding). A finding with neither is DROPPED, not softened.
  In sample mode the same queries run on a local engine over small samples.
- **ADR-0018 points 2 and 3:** one generator interface in Go, passed to every engine and agent; no engine or agent calls
  a model itself. Every engine writes append-only JSONL (audit, traces, provenance) to the run's `logs/`; the review's
  trace IS that JSONL, so a reader follows each finding from claim to verdict.
- **ADR-0025:** the change's code and its tests run only inside a container Proven starts, with no network and no
  secret. Where no container runtime is available the Prover runs no code, and the review says the Prover was off and
  why. **It never falls back to running the code outside a container.**

### THE SEAM (the contract, agreed by both masters 2026-10-21)
The Go interface between the shell (`proven-cli-and-action`) and the reviewer is package `internal/review` of module
`github.com/saphan-ai/proven`, in your base commit. Implement `review.Open` and the `Reviewer` it returns. Unresolved
findings stay in `Result.Findings` with `Unresolved` set. The shell keeps its own fixture reviewer under
`internal/review/fixture`; do not touch it. You implement it; you do not change it. If it cannot carry what this order asks, write
`SHOUT_the-seam.md` naming the field and why, and stop.

### WHAT TO BUILD
1. **The generator interface** (ADR-0018 point 2) and its sample implementation: it answers from recorded responses
   keyed by agent and prompt, deterministic, no network.
2. **The six agents**, each the simplest code that works, all calling models only through the generator:
   Planner (reads the change, names the risks), Investigator (turns a risk into a query, runs it through the query
   adapter), Prover (asks the test-run adapter for a test that fails on the change and passes on the base), Skeptic
   (one refutation attempt per finding through the generator), Judge (keeps a finding only with evidence; drops one with
   none), Writer (fills the seam's result).
3. **Two more adapters, sample implementations only in this wagon:**
   - the query adapter: runs SQL on a local engine over small sample tables committed in the repository;
   - the test-run adapter: in sample mode returns the test outcomes recorded in the fixture and executes NO code. Its
     live implementation in this wagon returns "Prover off: no container runner yet" (ADR-0025 point 6), and the review
     names it as a failed service. No code path runs change code on the host.
4. **The World Model trace:** an append-only JSONL file in the run's `logs/`, one event per claim, evidence, refutation
   and verdict, each event carrying the finding id.
5. **A fixture change** (base and change trees, sample tables, recorded generator answers, recorded test outcomes) with
   three candidate findings: F1 proved by a recorded test, F2 proved by a sample query and its rows, F3 with no evidence.
6. Wire the agents behind the seam so `proven scan --sample <fixture>` runs them. Write the package layout in one
   short section of the repository's `AGENTS.md`.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test ./...` from the repository root: green. Paste the summary lines.
2. `proven scan --sample <fixture>` prints a review with F1 and F2, each with its evidence, and **without F3**. Paste it.
3. The trace in `logs/` follows F1, F2 and F3 from claim to verdict; F3's verdict says "dropped: no evidence". Paste the
   events for F3.
4. **Negative controls (law #27), each a test:** (a) the fixture gives F3 a recorded query result → F3 appears;
   (b) the fixture's recorded test for F1 is removed → F1 disappears. Name the mutated fixture field in each test.
5. **Unplug (law #41):** a test runs the review with the test-run adapter's live implementation; the review completes,
   names the Prover as off, and F1 is not reported as proved.
6. A test proves no network is opened in sample mode (for example, a transport that fails the test on any dial).

### THIS ORDER CAN BE WRONG (law #44)
The master's claims, made 2026-10-21 without building: (M1) the seam can carry evidence of both kinds and the failed
services; (M2) a local SQL engine in pure Go is enough for the sample queries; (M3) six agents fit one package each
without import cycles. Before you build on M1, read the seam and check it. In your status, fill
`- master-was-wrong:` with each claim that did not hold and the measurement that shows it, or `none`.

### WHAT IS NOT IN SCOPE
- The shell: flags, output formats, exit codes. Do not edit the seam's package.
- Any live adapter beyond the Prover-off stub: Nightshift (W2), the guardrail (W3), H-MEM (W4), Zetaris (W5).
- A container runner. No `os/exec` of change code anywhere.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `████████████████████████████████████████████████`
in the worktree. Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict. Read every `saphan`
verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 5 |
