# ORDER — evidence-pack · round 48

- change: evidence-pack
- round: 48
- base: main @ d9083c0
- executor: claude-code
- written: 2026-10-25T01:09:31Z
- task: round 5 — a signed evidence pack can never be replaced by another output of the same run

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — evidence-pack
track: proven-enterprise
- train: proven-enterprise
- wagon: W2
- review: named
- repo: proven
- base: main @ d9083c0
- branch: fix/evidence-pack
- lands-on: main
- backend: claude-code
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: round 5 — a signed evidence pack can never be replaced by another output of the same run

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

### THE OWNER'S WORDS
2026-10-23: ADR-0035 — a signed evidence pack per review and `proven verify`; *"the settings used (never a secret)"*.

### THE REVIEW OF ROUND 4 (tip 5856357) — PASTED (law #44)
codex: **reject**, one BLOCKING. (The second review of round 4 was not run; it reviews round 5.)
- **The pack is left out of the output-collision check.** `internal/cli/config.go:453` lists the output paths but omits
  `--evidence-pack`; `config.go:394` runs the collision check before `cfg.evidencePack` is assigned at `config.go:433`;
  `internal/cli/run.go:128` writes the signed pack and `writeOutputs` (`run.go:134`, `os.WriteFile` at `run.go:299`) then
  writes `-o` to the same path. `--evidence-pack X -o X` replaces the signed pack with text and exits 0.
  **Fix:** refuse a collision of the pack with `-o`, `--text-output`, `--json-output`, `--sarif-output` and their aliases
  (the same file reached by two spellings, an existing link) before the review starts; keep the written pack intact if
  the filesystem changed after that check.

The reviewer's test, verbatim:
```go
package cli

import (
 "bytes"
 "os"
 "path/filepath"
 "testing"

 "github.com/saphan-ai/proven/internal/evidence"
)

func TestEvidencePackOutputCannotReplacePack(t *testing.T) {
 target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
 t.Chdir(t.TempDir())
 privateKey, publicKey := packKeys(t)
 packPath := filepath.Join(t.TempDir(), "review.pack")
 var stdout, stderr bytes.Buffer
 exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, "-o", packPath, target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
 if exit == exitOK {
  t.Fatalf("colliding output was accepted: %q", stderr.String())
 }
 if body, err := os.ReadFile(packPath); err == nil {
  public, err := evidence.LoadPublicKey("test", publicKey)
  if err != nil { t.Fatal(err) }
  if _, err := evidence.Verify(body, public); err != nil { t.Fatalf("a written pack must remain verifiable: %v", err) }
 }
}
```

### WHAT TO BUILD
1. Stay on the branch as it is (main merges into it cleanly); do not rebase or merge.
2. The fix above, with the reviewer's test, and a test for an aliased path (a link to the pack path given as `-o`).
3. Owed from round 3 where cheap: assert the detail in the `duplicate` subtest; verify `FileRef.Size` or drop it;
   bound `VerifyFile`'s read; test `ResolveCommit` including `--end-of-options`.

### THE MASTER CAN BE WRONG (law #44)
The master read the finding from the review notes on 2026-10-25 and did not run it. **Run the reviewer's test at 5856357
BEFORE you change anything** and record its output. If it passes there, say so under `- evidence:` as `master-was-wrong:`
and stop with `- state: blocked`.

### ACCEPTANCE
1. `go test -count=1 ./...` (whole module) green; `go vet ./...` clean.
2. `TestEvidencePackOutputCannotReplacePack` and the alias test pass, and were red at 5856357.
3. Negative control, named: remove the pack from the collision check; both tests turn red by name. Restore it.
4. Subjects match `^\[evidence-pack\] [a-z0-9/._-]+: .+$`, at most 72 characters; no `Co-authored-by` line.

### WHAT IS NOT IN SCOPE
- W1, W3. JSON or SARIF writers. The general short-secret refusal at load (proven-reviewer-core). Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, each finding with
its test, the negative controls, the whole-module counts. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
