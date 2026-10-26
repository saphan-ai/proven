# ORDER — world-model-bitemporal · round 5

- change: world-model-bitemporal
- round: 5
- base: fix/world-model-bitemporal @ c9e12f2
- executor: claude-code
- written: 2026-10-25T15:15:18Z
- task: the World Model's committed scorecard and its figures in the README, the architecture page and the design page show the measured 8 of 8

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — world-model-bitemporal
track: proven-meterless-engines
- train: proven-meterless-engines
- wagon: W56
- round: 2
- review: named
- repo: proven
- base: fix/world-model-bitemporal @ c9e12f2
- branch: fix/world-model-bitemporal
- lands-on: main
- backend: claude-code
- quote-usd: 10.00
- cap-usd: 20.00
- wall: 60m
- task: the World Model's committed scorecard and its figures in the README, the architecture page and the design page show the measured 8 of 8

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
FIRST `git rev-parse --short=7 HEAD` must print `c9e12f2` and `git rev-parse --abbrev-ref HEAD` must print
`fix/world-model-bitemporal`; else `- state: blocked`, stop. Never detach.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, e.g. `[world-model-bitemporal] docs: World Model passes 8 of 8`.
No AI attribution anywhere. After the last commit paste `git log --format=%B c9e12f2..HEAD | grep -ci co-authored` — it must
print `0`. Stage only the four files named below.

### SKILLS
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY
Round 1 built the bitemporal edges but could not run Meterless's suite: this seat's permission class refuses every spelling
that sets `METERLESS_ENGINES`. The master ran it at your tip `c9e12f2` on 2026-10-25 (macm5, runner pinned at upstream commit
`1742dec5`): `world-model exit 0`, `8/8 checks passed`. Do not run the suite; write its printed result as given here.

### WHAT TO WRITE (exactly)
1. Replace the whole of `test/conformance/meterless/SCORECARD_world-model.txt` with these bytes (the runner's own output,
   machine paths already stripped by `run.sh`):
```

World Model conformance: implementation = test/conformance/meterless/world-model/index.ts

  PASS sid-001 stable ids: diacritics and NBSP normalize identically (NFKD + strip marks)
  PASS sid-002 stable ids: distinct names stay distinct (no over-normalization)
  PASS band-001 merge bands at the boundaries: 0.85 auto, 0.84 review, 0.82 review, 0.81 new
  PASS prov-001 writes without provenance are rejected
  PASS idem-001 ingest idempotency: same batch twice produces an empty diff
  PASS idem-002 the idempotency check can fail: a mutated batch changes the shape
  PASS bitemp-001 bitemporal edges: a superseding edge closes the old one; point-in-time queries respect validity
  PASS audit-001 every mutation writes an audit entry in order

  8/8 checks passed
```
2. `README.md`: in the sponsor table's Meterless row, `9 of 13 and 7 of 8` becomes `9 of 13 and 8 of 8`; in the earlier
   day's summary, `World Model 7 of 8` becomes `World Model 8 of 8`. Change no other figure (H-MEM's is another wagon's).
3. `docs/ARCHITECTURE.md`: after the sentence ending `` `test/conformance/meterless/run.sh`. `` add: `The World Model
   passes its suite 8 of 8, bitemporal edges included (runner pinned at upstream commit ``1742dec5``).`
4. `hackathon/DESIGN.md`: replace `That this works is not measured yet.` with `The World Model passes its suite 8 of 8,
   bitemporal edges included (runner pinned at upstream commit ``1742dec5``).`

### ACCEPTANCE
1. `git diff c9e12f2..HEAD --stat` touches only those four files.
2. `go build ./... && go test -count=1 ./test/conformance/... ./cmd/meterless-conform/...`: paste the summary lines.
3. Paste each changed sentence.

### WHAT IS NOT IN SCOPE
Code. H-MEM's figures. Running the suite.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████` in the worktree, except the four files named.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run
and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 3 |
