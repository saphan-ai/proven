# ORDER — evidence-pack · round 2

- change: evidence-pack
- round: 2
- base: main @ fcbcb79
- executor: claude-code on macm5
- written: 2026-10-24T05:18:51Z
- task: round 2 — the evidence pack resolves commits without importing os/exec in internal/cli, so the banned-import rule holds

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — evidence-pack
track: proven-enterprise
- train: proven-enterprise
- wagon: W2
- review: named
- repo: proven
- base: main @ fcbcb79
- branch: fix/evidence-pack
- lands-on: main
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 60m
- task: round 2 — the evidence pack resolves commits without importing os/exec in internal/cli, so the banned-import rule holds

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[evidence-pack] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY THIS ROUND (the master's measurement, 2026-10-24)
Round 1 (tip 1e83a0a) is green in its own packages, but the whole module is not. The master ran `go test -count=1 ./...`
at base fcbcb79 (40 ok, 0 FAIL) and at tip 1e83a0a (one new red), pasted:
```
--- FAIL: TestInternalAndCmdDoNotImportBannedPackages (0.08s)
    imports_test.go:99: internal/cli/pack.go imports os/exec
FAIL	github.com/saphan-ai/proven/internal/pipeline
```
`internal/cli/pack.go:92` runs `git rev-parse` through `exec.Command` directly. Every git call in Proven goes through
`internal/change` (which gives git a minimal environment, `childenv.Minimal`).
**Step 1:** reproduce the red with `go test ./internal/pipeline/ -run TestInternalAndCmdDoNotImportBannedPackages`.

### WHAT TO BUILD
1. Resolve the commit through `internal/change` (an existing helper, or a small exported one there that uses its git
   runner); remove `os/exec` from `internal/cli/pack.go`. Do not add an exemption to the banned-import list.
2. Nothing else changes in behaviour.

### ACCEPTANCE
1. `go test -count=1 ./...` (whole module) green; `go vet ./...` clean.
2. `TestInternalAndCmdDoNotImportBannedPackages` passes; the evidence-pack tests still pass.

### WHAT IS NOT IN SCOPE
- Anything beyond the import. Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the whole-module
test counts. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
