# ORDER — world-model-bitemporal · round 6

- change: world-model-bitemporal
- round: 6
- base: fix/world-model-bitemporal @ d7ffd71
- executor: claude-code
- written: 2026-10-25T15:52:39Z
- task: the World Model's bitemporal edges stay correct when a target is renewed, when writes arrive out of time order and when timestamps carry offsets, with 8/8 kept

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — world-model-bitemporal
track: proven-meterless-engines
- train: proven-meterless-engines
- wagon: W56
- round: 3
- review: named
- repo: proven
- base: fix/world-model-bitemporal @ d7ffd71
- branch: fix/world-model-bitemporal
- lands-on: main
- backend: claude-code
- quote-usd: 10.00
- cap-usd: 20.00
- wall: 240m
- task: the World Model's bitemporal edges stay correct when a target is renewed, when writes arrive out of time order and when timestamps carry offsets, with 8/8 kept

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
FIRST `git rev-parse --short=7 HEAD` must print `d7ffd71` and `git rev-parse --abbrev-ref HEAD` must
print `fix/world-model-bitemporal`; else `- state: blocked`, stop. Never detach, never check out another commit.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, e.g. `[world-model-bitemporal] world: <what>`. The area is ONE
lowercase token. No AI attribution anywhere. After the last commit paste
`git log --format=%B d7ffd71..HEAD | grep -ci co-authored` — it must print `0`. Stage only the files you name; never
`git add -A`, never stage `████████`.

### SKILLS
- ██████████████████████████████████
- ███████████████████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S ORDER, 2026-10-25 13:00 UTC (relayed by the advisor), VERBATIM IN SUBSTANCE
> Full Meterless conformance. Today Proven reports H-MEM 9/13 and World Model 7/8 because sections are not built;
> competitors publish 13/13 and 8/8. Build the missing sections of both engines so their own suites pass in full,
> scorecards published as they come, in their minimum form (ADR-0018 point 1 allowed 'after the minimum works': it works).

And the owner's standing rule of the same day: when a feature lands, the same change updates the documentation and
diagrams, not a later one. This order covers the World Model; another wagon covers the other engine in parallel.

### WHY — THE NAMED REVIEW OF d7ffd71 (codex, verdict reject), VERBATIM
> 1. **BLOCKING — repeat of the same target creates overlapping history.** `internal/world/graph.go:52-69`: the content address includes `ValidFrom`, but the closure loop excludes an open edge with the same `To`. Control: `internal/world/graph_test.go:52-91` changes A to B only and never repeats A at a later time. Reproduction: A→B at 2087, A→B again at 2088, A→C at 2089. `Neighbors(A, member, 2088-06-01)` returned `[B,B]`, though one relationship is valid. The 2087 edge remains valid through 2089 and overlaps the 2088 edge. Remedy: define whether repeating the target renews or preserves one edge, then close or reject the overlapping interval and test the temporal result.
> 2. **BLOCKING — out-of-order writes close an edge before it starts.** `internal/world/graph.go:57-69`: any open edge is closed at the new edge's start without checking order. Control: `internal/world/graph_test.go:73-91` inserts only increasing times. Reproduction: A→B at 2089, then A→C at 2088. The store accepts both; `Neighbors(A, late, 2089-06-01)` returned B, but the open/current neighbor is C. B has `ValidTo=2088` before its `ValidFrom=2089`; a historical query treats it as absent and current query treats C as current. Remedy: reject retroactive supersession or re-split intervals in temporal order; test reverse insertion and valid interval bounds.
> 3. **BLOCKING — timestamp strings are compared as bytes, not instants.** `internal/world/graph.go:102-104` compares caller strings without parsing or normalizing; `Relate` accepts timestamp offsets at lines 41-52. Control: `internal/world/graph_test.go:80-84` uses only one fixed UTC format. Reproduction: edge starts `2087-01-01T01:00:00+01:00` (00:00Z); query at `2087-01-01T00:30:00Z` returned zero neighbors, though the edge had started. Remedy: parse and compare instants, or enforce one canonical timestamp format at writes and queries; cover equivalent offsets.

Master-measured: world-model 8/8 on Meterless's runner at c9e12f2; whole suite at d7ffd71 54 ok, 0 FAIL. The check passes,
but the review found three temporal cases that return wrong results; the check does not cover them. The fix keeps 8/8.

### THE REFERENCE (placed by the master at a pinned commit, Apache-2.0, read-only)
Meterless upstream `github.com/Meterless/Meterless` at commit `1742dec56a66e58eef9dda2d2b865823758ca8cd`. This seat has no egress to github.com, so the
master placed a clone in your worktree at `█████████████████████████████████████████████`. Do NOT fetch anything. Verify it first:
`git -C █████████████████████████████████████████████ rev-parse HEAD` must print the commit above. If the directory is absent (this round may run on a host without it), do NOT
fetch anything and do not block: skip every step that runs Meterless's suite, say so in `- note:`, and the master runs the suite. Never commit or stage that directory. Below, `$M` means `█████████████████████████████████████████████████████`.
Read: `$M/world-model/AGENTS.md` (the entry point), `$M/world-model/docs/` (the specification of each failing section),
`$M/world-model/reference/` (the TypeScript reference), `$M/world-model/conformance/` (the runner and each check).
Run the suite with the project's own script: `METERLESS_ENGINES=<absolute $M> bash test/conformance/meterless/run.sh`. If your permission class refuses setting that variable (round 1 of a sibling wagon met that), do not work around it: say so, and the master runs the suite.

### CLAIMS FROM THE MASTER — TEST THEM (law #44)
- The master claims the World Model lives in `internal/world` and its shim in `test/conformance/meterless/world-model/index.ts` +
  `cmd/meterless-conform`. Measure it.
- The master has not read the failing checks' specification. Where a check needs a model (for example a "dream"), the
  engine still calls no model itself (ADR-0018 point 2: one generator interface, passed in); the check's own scripted
  generator drives it. If a check cannot pass without breaking that rule, stop and say so.

### WHAT TO BUILD
1. Renewing the same target never leaves two open edges for one (from, to): the open one is closed or extended, by a rule
   you state, and history has no overlap.
2. An out-of-order write never closes an edge before its own start: insertion orders by validity time, and a write that
   starts before an open edge is placed in history, not over it.
3. Timestamps are parsed to instants on input (offsets normalised) and compared as instants; an unparsable timestamp is
   refused by name.
4. A Go test per finding, each red before your fix (paste the red at d7ffd71), green after, with expected neighbour sets:
   same-target renewal; reverse-time insertion; equivalent timestamps with different offsets.
5. Meterless's suite still prints `world-model 8/8`; if the scorecard changes, commit it as printed. Docs unchanged unless a
   sentence becomes false.
### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: paste the summary lines; name any test that cannot bind a port.
2. Step 0's run and the final run of `run.sh`, pasted; `world-model` at `8/8`; the other engines' lines unchanged.
3. The negative control's red, pasted, and restored.

### COMMENTS FOR THE JUDGES (ADR-0027)
Every exported name has a doc comment beginning with the name; every guard says WHY; code that carries out a decision names
it (`ADR-0018: …`); every test says which property it proves; English, plain; no host names, internal paths or credential
names; no AI attribution.

### WHAT IS NOT IN SCOPE
- The other engine's code and figures (a parallel wagon). Markovian and Scout Intent. The generator interface.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████` in the worktree, except the files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, each master claim a
measurement contradicted (in `- note:`), the commands run and their exit codes.
Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 5 |
| how an order is bound to the laws | 1 |
| private workspace path | 5 |
