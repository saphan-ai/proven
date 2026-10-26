# ORDER — world-model-bitemporal · round 7

- change: world-model-bitemporal
- round: 7
- base: fix/world-model-bitemporal @ e91cec0
- executor: claude-code
- written: 2026-10-25T23:19:01Z
- task: a target renewed after a bounded edge never leaves two edges valid at one instant, with 8/8 kept

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — world-model-bitemporal
track: proven-meterless-engines
- train: proven-meterless-engines
- wagon: W56
- round: 4
- review: named
- repo: proven
- base: fix/world-model-bitemporal @ e91cec0
- branch: fix/world-model-bitemporal
- lands-on: main
- backend: claude-code
- quote-usd: 10.00
- cap-usd: 20.00
- wall: 240m
- task: a target renewed after a bounded edge never leaves two edges valid at one instant, with 8/8 kept

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
FIRST `git rev-parse --short=7 HEAD` must print `e91cec0` and `git rev-parse --abbrev-ref HEAD` must
print `fix/world-model-bitemporal`; else `- state: blocked`, stop. Never detach, never check out another commit.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, e.g. `[world-model-bitemporal] world: <what>`. The area is ONE
lowercase token. No AI attribution anywhere. After the last commit paste
`git log --format=%B e91cec0..HEAD | grep -ci co-authored` — it must print `0`. Stage only the files you name; never
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

### WHY — THE NAMED REVIEW OF e91cec0 (codex, verdict reject), VERBATIM
> BLOCKING — original repeat-target overlap remains reachable. `internal/world/graph.go:89`: `Relate` calls `place` only when incoming `ValidTo` is empty. `internal/world/graph.go:46`: `sameChain` excludes an originally bounded edge (`ValidTo != ""`, `SupersededBy == ""`) even if the new edge is open. Write A from 2087-01-01 until 2089-01-01; write A again, open, from 2088-01-01 with the same from/type/context. The second has a different ID, skips the first in `place`, and is stored. At 2088-06-01, `Neighbors` (`internal/world/graph.go:214`) returns the same target twice. That violates renewal and gives duplicate point-in-time neighbors. Named remedy: test a bounded first edge followed by same-target renewal, expecting one interval and one neighbor; validate or refuse overlaps for bounded edges. `TestRepeatedTargetRenewsOneEdge` (`internal/world/graph_test.go:141`) covers only an open first edge. Finding 1: NOT ADDRESSED on this public bounded-edge path. Reproduced with a temporary Go caller under `████████████`: `go run` returned `distinct_ids=true neighbors_at_2088=2`.

The same review found findings 2 (reverse insertion) and 3 (instants) addressed; keep them as they are.

Master-measured at e91cec0: world-model 8/8 on Meterless's runner; whole suite 54 ok, 0 FAIL; round 3's four tests
fail on d7ffd71's graph code and pass at e91cec0. The fix keeps 8/8.

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
1. A same-target write over an edge that has an explicit `valid_to` never leaves two edges valid at one instant: renew
   (one interval) or refuse the overlap by name, by a rule you state in a doc comment. The same holds when the new edge
   is bounded too, and for overlap with any earlier edge of the same (from, type, context), not only the open one.
2. A Go test per case, each red before your fix (paste the red at e91cec0), green after, with the expected interval and
   neighbour set at an instant inside the overlap: bounded first edge, open renewal; bounded first edge, bounded renewal.
3. Round 3's four tests stay green. Meterless's suite still prints `world-model 8/8` where the reference is present.
   Docs unchanged unless a sentence becomes false.
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
| absolute or private path | 6 |
| how an order is bound to the laws | 1 |
| private workspace path | 5 |
