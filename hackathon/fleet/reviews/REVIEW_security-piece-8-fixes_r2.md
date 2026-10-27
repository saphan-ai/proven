# REVIEW — security-piece-8-fixes · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T12:40:30Z

## Summary

claude-opus-5-5 named review of W46 round 2 at 3e737d8: B1, B2, B3, O1, O2 all ADDRESSED with a control each (O1 unplugged by me: red).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — security-piece-8-fixes

Reviewer: claude-code, claude-opus-5-5, 2026-10-24, headless, worktree at `3e737d8` (`git rev-parse --short=7 HEAD` = `3e737d8`).
Scope per R14: the round-1 conditions and the round-2 fix diff (`f87ab7d..ce440e1`, four first-parent commits from
`37de26a`) plus the absorb `3e737d8`. Anything outside that is recorded as owed and opens no round.

### Ground

- Round-2 delta read whole: `git diff f87ab7d ce440e1` (second.go, pathguard.go, redact.go comments, ADR-0013 point 4a,
  five new or changed test files). The stream as a whole: `git diff --stat main...3e737d8` — 32 files, all stream files.
- `internal/pathguard/pathguard.go` read whole at the tip (Resolve, Inside, sameFileAbove), and its callers in
  `internal/pipeline/memory.go:begin` (logs, state, work vs target; state vs module).
- `internal/service/secret.go` (MinSecretBytes = 8, TooShortSecret trims, empty is not short) and
  `internal/redact/redact.go` read whole; `internal/change/archive.go` treeFileMode, writeTarFile, copyTracked.
- `internal/pipeline/trees.go:workArea` read: pure path arithmetic, writes nothing, so Build's "nothing is written here"
  holds.
- The author's self-assessment and the master's measurements are claims (law #44); every one below was re-measured
  or is marked as not re-measured.

### Differential

Round-1 conditions, each ADDRESSED or NOT ADDRESSED (R14):

1. **B1 · ADDRESSED.** `git diff --name-status 00d21f7 3e737d8` lists 32 entries, every one a stream file (M/A),
   **zero `D`**. The tip carries the Rust adapter (`internal/manifest/cargo.go`, `cargolock.go`) and the Composer /
   Packagist readers. The two controls in the order no longer read as written, and the reason is not a defect in the
   stream: `main` has moved from `00d21f7` to `891a7e9` (`govulncheck-says-what-is-reached`), so
   `git diff --name-status main 3e737d8 | grep '^D'` prints four files (`internal/tool/govulncheck.go`,
   `govulncheck_container_test.go`, `govulncheck_test.go`, `race_level_test.go`) that main added after the absorb.
   `git merge-base main 3e737d8` = `00d21f7`, so a three-dot merge does not delete them. Likewise
   `git diff --stat 656dbcb 3e737d8` now also lists other streams' merges (JSON/SARIF output, opinions, corpus) that
   landed on main after `656dbcb`; the absorb-base diff is the control that decides.
   The B1 test set ran green inside the full suite: `internal/manifest`, `internal/osv`, `internal/querytpl`,
   `cmd/proven` printed `ok`.
2. **B2 · ADDRESSED.** `internal/generator/second.go:66` adds `service.TooShortSecret(key)` to the refusal, by
   `EnvSecondKey`, value empty in the ConfigError. `TestLoadSecondRefusesAKeyTooShortToBeRedacted` (3 bytes, refused,
   key not printed) and the boundary control `TestLoadSecondAcceptsAKeyOfTheMinimumLength` (8 bytes, connected) PASS.
   An empty key still loads as disconnected (TooShortSecret is false on empty) — ADR-0013 4a preserved.
3. **B3 · ADDRESSED.** `internal/pathguard/pathguard.go:67-70` falls back to `sameFileAbove(rd, rr)` when the
   spelling test says outside; `sameFileAbove` (`:79-92`) compares device+inode of root with each ancestor of the
   resolved dir. `TestInsideDecidesByFileIdentityNotSpelling` PASSES on this darwin host and did NOT skip (the
   case-folded `REPO` stat succeeded), with a sibling `REPO2` control judged outside.
   Order question 2, answered:
   - *Ancestors that do not exist yet:* holds. `Resolve` resolves the longest existing prefix and appends the
     missing names, and `sameFileAbove` walks up past the missing names to the first existing ancestor; the test
     case `REPO/new/deep` covers it.
   - *A link on a parent segment:* holds. `Resolve` runs `EvalSymlinks` on the existing prefix, so every link on a
     parent segment is resolved before either comparison; `TestInsideComparesResolvedPaths` "through a link"
     (`other/link/state`, link → `root/in`) PASSES. The combination (a parent link whose target is a case-respelling
     of root) is not tested; by the code path it reduces to the two cases above (resolved spelling `…/REPO/x`, then
     identity). Recorded under Remainder, not a condition.
4. **O1 · ADDRESSED.** `internal/change/worktree_mode_test.go` sets umask 077 and pins 0644 / 0755. The master's
   change from 002 to 077 is correct: `copyTracked` already opens with `treeFileMode(...)` (0644/0755), so under
   umask 002 the file lands 0644 with or without the Chmod — a 002 test cannot redden. Under 077 the open alone
   gives 0600/0700, so only the Chmod produces the pinned modes. Unplugged by me (see Unplug): red.
5. **O2 · ADDRESSED.** `redact.go:25-30` (Names) says a short organisation id is not redacted and can appear in a
   trace; `redact.go:103-111` says "five credentials" and names them; ADR-0013 point 4a states the 8-byte floor,
   the organisation-id exception and that an id of 8 bytes or more is redacted. Matches `FromEnv` (`:117`) and
   `TestFromEnvKeepsAShortOrganisationIdOutOfTheRedactor` (PASS).
6. **OWED by the owner's rule — NOT ADDRESSED (not this stream's to address).** The independent NVIDIA review
   through Nightshift did not run; this review ran alone, as the order states.

### Unplug

- **O1, run by me:** `internal/change/archive.go` copyTracked `out.Chmod(treeFileMode(...))` replaced by a no-op →
  `go test -count=1 -run TestExportWorktreeSetsFileModesWhateverTheUmask ./internal/change` →
  `FAIL … plain.txt: mode 600, want 644 · run.sh: mode 700, want 755`. Restored; `git status --porcelain` shows no
  tracked change.
- **B2 and B3, NOT re-run by me:** the master's measurement (2026-10-24, macm5) is that removing `TooShortSecret` from
  second.go reddens `TestLoadSecondRefusesAKeyTooShortToBeRedacted` and that `return false, nil` in place of
  `sameFileAbove` reddens `TestInsideDecidesByFileIdentityNotSpelling`. I confirmed by reading that each test reaches
  exactly that line with an input only that line refuses or accepts (3-byte key, no other refusal clause matches;
  `REPO` respelling, Rel yields `../REPO`), and that the B3 test did not skip on this host. The mutation runs
  themselves are the master's, not mine.

### Closure

All five round-1 code conditions (B1, B2, B3, O1, O2) are ADDRESSED with a control each. The five ADR-0025 /
ADR-0013 properties, as far as the fix diff reaches them:
1. State, trace, work outside the target — `begin` refuses before any write; review tests
   `TestReviewRefusesLogsAndStateInsideTheReviewedTarget`, `TestReviewRefusesATraceInsideAnEmptyTarget` PASS.
2. Inside-the-repository check resolves links and decides by identity — B3 above; hmem/world placement tests PASS.
3. Wider state file narrowed on open; atomic write not carrying a wider temp mode — hmem/world
   `TestOpenNarrowsAnExistingWideStateFile`, `TestAtomicWriteDoesNotCarryAWiderTemporaryMode` PASS.
4. Every configured secret redacted whatever its length; short reviewer keys refused by name; short org id left out —
   B2, O2; redact/trace/query/generator/guardrail short-secret tests PASS.
5. Materialized trees not group-writable whatever the umask — O1 + `TestMaterializedTreesAreNotGroupWritableUnderAPermissiveUmask` PASS.

### Remainder

Owed, non-blocking (R6, R14 — none opens a round):
- **R-1 · re-absorb before merge.** `main` is at `891a7e9`, two merges past the absorb base `00d21f7`, and
  `git diff --name-only 00d21f7 main` includes `internal/pipeline/pipeline.go`, which this stream also changes.
  `git merge-tree --write-tree main 3e737d8` was refused by this seat's grant (see Unrun), so whether that overlap
  merges clean is unmeasured. Remedy: the master re-absorbs main (or runs merge-tree) and re-runs the suite before
  the gate.
- **R-2 · `redact.New` doc comment** (`internal/redact/redact.go:45-49`) still says no short secret "reaches here
  from the environment", while `FromEnv`'s corrected comment (`:105-107`) now says a credential set but unused by the
  active mode is handed over unchecked. Remedy: align New's comment with FromEnv's.
- **R-3 · a short UNUSED credential is redacted as it is.** Per `FromEnv` (`:105-107`), e.g. a one-byte
  `EnvKey` set while the mode does not use it is not floor-checked, and the redactor will then replace that byte in
  every trace line (commit ids, line numbers) — the very damage the organisation-id exception exists to avoid. It
  never leaks a secret (it over-redacts), so it is a trace-integrity defect, not a disclosure. Remedy: refuse a short
  value in every Names credential at load whether or not the mode uses it, or leave it out of FromEnv like the org
  id; a test with a 1-byte unused key and a trace line containing that byte.
- **R-4 · combined case untested:** a parent-segment link whose target is a case-respelling of root. Correct by the
  code path (Differential B3); a one-row addition to `TestInsideDecidesByFileIdentityNotSpelling` would pin it.

Declined to judge (R11): the ADR-0013 wording beyond point 4a; the umask-002 pipeline test's discrimination
(round-1 scope); licences (out of scope per order).

### Profile

security (R9: configuration secrets, filesystem placement, file modes). S1 boundaries: reviewed target ↔ proven's
state/trace/work (tamper/disclosure via placement inside the target) — guarded by `begin` + pathguard identity;
environment ↔ trace (disclosure of a secret) — guarded by the length floor + redactor; materialized trees ↔ other
local users (tamper via group/other-writable files) — guarded by treeFileMode + Chmod. S3: the O1 guard removed once
→ red (mine); B2 and B3 removed once → red (master's measurement, not re-run). S4: dependency delta of the stream is
empty (`go.mod`/`go.sum` not in `main...3e737d8`); no destructive operation on a derived path in the fix diff; no
invisible or bidirectional characters observed in the fix diff as read.

### Focus

Order items: (1) each round-1 condition with its control — Differential 1–6; (2) B3 for not-yet-existing ancestors and
a parent-segment link — Differential 3; (3) `go build ./... && go vet ./... && go test -count=1 ./...` — build clean;
vet clean on every package but `internal/cli/sarifschema` (dependency fetch refused, see Unrun); test results below and
in Unrun.
Full-suite outcome on this host: every package `ok` except `internal/cli/sarifschema` (setup failed: module download
refused) and six packages aborted by the first `httptest` listener panic (`cmd/livecheck`, `internal/generator`,
`internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`). In those six I ran the stream's own
tests by name; all PASS (listed in Differential and Closure).

### Conditions

1. **C1 · OWED (owner's rule).** The second, independent NVIDIA review through Nightshift is still owed before merge.
2. **C2 · OWED.** Re-absorb `main` (now `891a7e9`; overlap on `internal/pipeline/pipeline.go`) and re-run
   `go build ./... && go vet ./... && go test -count=1 ./...` outside the sandbox, including the six
   listener-bound packages and `internal/cli/sarifschema` listed under Unrun.
3. **C3 · OWED, non-blocking.** R-2 (New's doc comment) and R-3 (short unused credential over-redacts) — a follow-up,
   not another round.

### Unrun

- `go test` of `internal/cli/sarifschema`, quoted: `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip": tls: failed to verify certificate: x509: OSStatus -26276` → `[setup failed]`. Same cause stops `go vet ./...` at load; vet was run on every other package instead (clean).
- Every `httptest`-bound test in `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`,
  `internal/query`, `internal/review`, quoted: `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. The panic aborts the package, so the remaining tests there ran only by name (the stream's own, all PASS); the listener-bound ones did not run.
- `git merge-tree --write-tree main 3e737d8` — refused by this seat's permission layer ("This command requires approval"); the merge against current main is unmeasured.
- The B2 and B3 mutation runs — taken from the master's measurement, not re-run by me.

## The review status as filed

## REVIEW — security-piece-8-fixes
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: claude-opus-5-5 named review of W46 round 2 at 3e737d8: B1, B2, B3, O1, O2 all ADDRESSED with a control each (O1 unplugged by me: red).
  Owed: the NVIDIA second review (owner's rule), a re-absorb of main (now 891a7e9, overlaps pipeline.go) with a full suite outside the sandbox,
  and two non-blocking redact follow-ups. Unrun: sarifschema (fetch refused), httptest-bound tests (bind refused), merge-tree (not granted).
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T13:30:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
