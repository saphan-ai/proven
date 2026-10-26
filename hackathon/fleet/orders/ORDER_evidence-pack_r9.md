# ORDER — evidence-pack · round 9

- change: evidence-pack
- round: 9
- base: main @ fcbcb79
- executor: claude-code
- written: 2026-10-24T10:11:15Z
- task: round 3 — close the four gaps both reviews of the evidence pack found: redact the manifest, refuse a readable key, escape pack text, verify the exact archive bytes

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
- backend: claude-code
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: round 3 — close the four gaps both reviews of the evidence pack found: redact the manifest, refuse a readable key, escape pack text, verify the exact archive bytes

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
2026-10-23: features closer to enterprise; the decision is ADR-0035 (a signed evidence pack and `proven verify`).
2026-10-24: Cursor seats first for coding (the first try on Cursor died: session-died; moved to claude-code).

### THE REVIEWS OF ROUND 2 (tip 24cfc28) — PASTED (law #44: carry, do not point)
Two independent reviews: codex **reject** (four BLOCKING), claude-code **accept-with-conditions** (the same four plus
mutant gaps). The conditions, as written by the claude-code review:
1. **C1:** redact the manifest's model list at `internal/evidence/pack.go:134`, for example by building `Models` from
   the same redacted bytes as `receipts.json` (`pack.go:192` redacts the payload copy, so `verify.go:214` rejects a
   secret-bearing model). Add a test that puts the secret in `Route` or `Model`, then asserts the secret is absent from
   every file and that the pack verifies.
2. **C2:** in `LoadPrivateKey`, refuse a private key file with any group or other permission bit, by name and with
   exit 7 (`internal/evidence/key.go:71-85`; `internal/cli/run.go:63-71` calls it before review). Add a test with a
   `0644` key that fails before the fix. Docs: the refusal and its remedy, `chmod 600`, in `docs/EVIDENCE_PACK.md`.
3. **C3:** escape control and bidirectional characters in every pack-derived string that reaches stdout or stderr:
   `Refusal.Error` / `verify.go:41-47`, and `renderReport` at `internal/cli/pack.go:186-214` (it prints model, route,
   receipt, commit fields and the pack path with `%s`). main now has a terminal-escaping helper in `internal/cli` (the
   CLI train added it: tests `TestTextOutputCarriesNoControlByteFromTheChange`) — use it. Add a test with an ESC byte in
   a model name.
4. **C4:** make the pack byte-canonical: `Verify` rebuilds `archive(files)` and requires it to equal the input bytes
   exactly (`verify.go:164-192` stops at the first tar EOF, so appended bytes, metadata and padding changes pass). The
   writer is deterministic. Pin it with a test that appends one byte and one that flips a header byte.
5. **C5:** add tests that kill the six surviving mutants listed in the claude-code review's Unplug section, plus one
   that sets both `--public-key` and `██████████████████████` and asserts the flag wins.

### WHAT TO BUILD
1. **Rebase first:** `git rebase --onto main fcbcb79` (main has moved; it now holds the terminal-escape helper). Do not
   merge (merge is refused to a leg).
2. C1–C5 above. Keep ADR-0035's four points.

### ACCEPTANCE
1. `go test -count=1 ./...` (whole module) green; `go vet ./...` clean.
2. Each of C1–C4 has a test that fails on 24cfc28's behaviour and passes after; C5's tests kill the named mutants.
3. Negative controls, named: remove the redaction of C1, the mode check of C2, the escape of C3, the byte-equality of C4
   — each turns its test red by name. Restore all.
4. Subjects match `^\[evidence-pack\] [a-z0-9/._-]+: .+$`, at most 72 characters; no `Co-authored-by` line.

### WHAT IS NOT IN SCOPE
- The air-gapped mode (W1) and the policy (W3). JSON or SARIF writers. Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, each condition
with its test, the whole-module counts, the negative controls. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private credential or variable name | 1 |
| private workspace path | 4 |
