# ORDER — world-model-bitemporal · round 2

- change: world-model-bitemporal
- round: 2
- base: main @ 13349e5
- executor: claude-code on macm5
- written: 2026-10-25T13:59:27Z
- task: the World Model's bitemporal edges are built to Meterless's specification so its own conformance suite passes 8/8, with the scorecard and the README figures updated in the same change

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — world-model-bitemporal
track: proven-meterless-engines
- train: proven-meterless-engines
- wagon: W56
- review: named
- repo: proven
- base: main @ 13349e5
- branch: fix/world-model-bitemporal
- lands-on: main
- backend: claude-code
- quote-usd: 10.00
- cap-usd: 20.00
- wall: 240m
- task: the World Model's bitemporal edges are built to Meterless's specification so its own conformance suite passes 8/8, with the scorecard and the README figures updated in the same change

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
FIRST `git merge-base --is-ancestor 13349e5 HEAD && echo ok` must print `ok` and `git rev-parse --abbrev-ref HEAD` must
print `fix/world-model-bitemporal`; else `- state: blocked`, stop. Never detach, never check out another commit.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, e.g. `[world-model-bitemporal] world: <what>`. The area is ONE
lowercase token. No AI attribution anywhere. After the last commit paste
`git log --format=%B 13349e5..HEAD | grep -ci co-authored` — it must print `0`. Stage only the files you name; never
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

### WHAT FAILS TODAY (proven main @ 13349e5, `test/conformance/meterless/SCORECARD_world-model.txt`, pasted)
```
    FAIL bitemp-001 bitemporal edges: a superseding edge closes the old one; point-in-time queries respect validity
  7/8 checks passed
```

### THE REFERENCE (placed by the master at a pinned commit, Apache-2.0, read-only)
Meterless upstream `github.com/Meterless/Meterless` at commit `1742dec56a66e58eef9dda2d2b865823758ca8cd`. This seat has no egress to github.com, so the
master placed a clone in your worktree at `█████████████████████████████████████████████`. Do NOT fetch anything. Verify it first:
`git -C █████████████████████████████████████████████ rev-parse HEAD` must print the commit above; if it does not, write `SHOUT_reference-missing.md`, set
`- state: blocked`, and stop. Never commit or stage that directory. Below, `$M` means `█████████████████████████████████████████████████████`.
Read: `$M/world-model/AGENTS.md` (the entry point), `$M/world-model/docs/` (the specification of each failing section),
`$M/world-model/reference/` (the TypeScript reference), `$M/world-model/conformance/` (the runner and each check).
Run the suite with the project's own script: `METERLESS_ENGINES=<absolute $M> bash test/conformance/meterless/run.sh`.

### CLAIMS FROM THE MASTER — TEST THEM (law #44)
- The master claims the World Model lives in `internal/world` and its shim in `test/conformance/meterless/world-model/index.ts` +
  `cmd/meterless-conform`. Measure it.
- The master has not read the failing checks' specification. Where a check needs a model (for example a "dream"), the
  engine still calls no model itself (ADR-0018 point 2: one generator interface, passed in); the check's own scripted
  generator drives it. If a check cannot pass without breaking that rule, stop and say so.

### WHAT TO BUILD
0. Step 0: run the suite at your base and paste the failing check names; they must be the ones above. If they differ, say so.
1. Build each failing section in `internal/world` to the specification, in its minimum form, behind the existing interfaces. No
   engine calls a model or a network itself. Every write is JSONL to the run's `logs/` as the other sections do (ADR-0018 point 3).
2. Extend the shim only where a check needs a new operation; never change what the passing checks see.
3. Go tests for each new section (one per property the check names), each saying which property it proves, plus one
   negative control: revert one section's core rule → its Go test and its conformance check turn red; paste; restore.
4. Run the suite: `world-model` must print `8/8 checks passed`; commit `test/conformance/meterless/SCORECARD_world-model.txt` as printed.
5. Documentation in the same change (the owner's rule): README.md (the Meterless row of the sponsor table, and any
   sentence giving the World Model's score), docs/ARCHITECTURE.md (the conformance sentence) and hackathon/DESIGN.md (the Meterless
   section's scorecard line) state the new score as the committed scorecard prints it. Change only the World Model's figures.

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
