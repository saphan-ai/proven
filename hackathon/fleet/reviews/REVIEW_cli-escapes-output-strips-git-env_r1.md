# REVIEW — cli-escapes-output-strips-git-env · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:30:20Z

## Summary

reject at 1e77598 — P1 (every text field escaped), P3 (git child gets childenv.Minimal; fork-point test green), P4 (JSON/SARIF untouched) and P6 hold; both tests go red when unplugged.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-escapes-output-strips-git-env

### Ground
- `git rev-parse --short=7 HEAD` printed `1e77598` (verbatim). Scope `git diff 0b954a9...1e77598`: 4 files,
  +163/−20 — `internal/cli/format.go`, `internal/cli/run.go`, two new tests. Commits 025fb8c, 1e77598; author
  Marcin Marzec; no `Co-Authored-By`/vendor trailer in either message.
- The executor's self-assessment was set aside (R1); every claim below is re-measured on this host 2026-10-24.
- `go vet ./...` clean. `go test -count=1 ./internal/cli/ -run 'TestTextOutputCarriesNoControlByteFromTheChange|TestGitMergeBaseChildCarriesNoKeyFromProven|TestCIMergeBaseIsTheForkPoint' -v` → all three PASS.
- `go test -count=1 ./...` → 31 `ok`, 3 FAIL: `cmd/livecheck`, `internal/generator`, `internal/guardrail`, each
  `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` — this leg's sandbox
  refuses a listener; none of the three is in the write-set. The master's "37 ok, 0 FAIL" is NOT reproduced here (see Unrun).
- Profile: security (output encoding to a terminal; child-process environment).

### Differential
Property by property (FILE·LINE·CONTROL):
1. **Fields escaped — HOLDS.** Every `%s` in `textFormatter.Write`, `writeFinding`, `writeEvidence` that prints a
   string now goes through `visible`/`visibleLines`: class/rule `format.go:55,58`, statements `:62,67,72`, failed
   service+reason `:87`, family sentence `:98`, trace `:103`, name/severity/title `:148`, EPSS lines `:156`,
   path `:166`, message `:169`, rule `:175`, test path/name `:197`, command `:200`, output `:203`, engine/snapshot
   `:211`, SQL `:214`, columns `:217`, rows `:221`. The remaining raw prints are constant strings (`:82,93,162`)
   and `%d`/`%t`. No text field is printed raw.
2. **Total escaping and no forged line — DOES NOT HOLD.**
   - Range: C0 (except tab, and newline in `visibleLines`), DEL, C1 U+0080–U+009F escaped as `\xNN` — `format.go:258-266`. Correct.
   - Invalid UTF-8: `range s` yields U+FFFD and `WriteRune` writes EF BF BD, so a raw 0x9b (8-bit CSI) never
     reaches the terminal — measured (`m\x9braw\xff` → `m�raw�`). It is REPLACED, not shown as `\xNN` as the executor
     claimed; the order's proof admits "escaped or removed", so this is safe, but the claim is inexact.
   - **FINDING F1 (blocking): `format.go:203` — a kept newline forges Proven's own lines.** `visibleLines` keeps
     `\n` and the continuation lines start at column 0, with no prefix. Measured: `test.Output =
     "ok\nfailed services: none\nreview class: none"` renders
     ```
       output: ok
     failed services: none
     review class: none
     failed services: none
     ```
     The first two `failed services:`/`review class:` lines are written by the test's output (the code under the
     change runs inside that test and chooses its output), are indistinguishable from Proven's, and can hide or
     contradict a real `failed service:` line or invent a `finding … ERROR …` line. The order's property 2 names
     exactly this. Remedy: prefix every continuation line of a multi-line field with a fixed indent/marker that no
     Proven top-level line begins with (e.g. replace `\n` with `\n    | ` in `visibleLines`, or print each line
     under `  output:` with a constant prefix), and add a test asserting that no line of the output field begins
     at column 0 / begins with a Proven line keyword.
   - Remainder-class: the backslash is not escaped, so a literal `\x1b` in the change prints the same as an escaped
     ESC (`visible("a\\x1bb")` → `a\x1bb`). It can mislead a reader about what bytes existed; it cannot move the
     cursor. Owed, non-blocking.
3. **Subprocess environment — HOLDS.** `run.go:333` `cmd.Env = childenv.Minimal()`; `childenv.Minimal` passes
   PATH and HOME only (`internal/childenv/childenv.go:12`). `TestCIMergeBaseIsTheForkPoint` PASS (git works with
   the minimal env). The two other subprocess sites already used it (`internal/change/between.go:21`,
   `internal/testrun/container.go:553`); `run.go:329` was the only `exec.Command` in non-test code lacking it.
4. **JSON/SARIF byte-identical — HOLDS (vacuously).** `FormatterFor` (`format.go:22-31`) has no json/sarif
   writer in this build (they refuse with exit 2), and the diff does not touch them. There is nothing to compare.
   Note the WHY comment at `format.go:233-234` ("their encoders already escape") asserts a property of code that
   does not exist; when it lands, Go's `encoding/json` does NOT escape DEL or U+0080–U+009F — see Remainder.
5. **Tests with differing false-world answers — PARTIAL.** Both tests turn red when their fix is unplugged (see
   Unplug). But the escape test's fixture leaves `Class`, `ClassRule`, `SettingsStatement`,
   `AuthorModelStatement`, `ReviewerStatements`, `FailedServices[].Service`, `Severity`, `TracePath`, the family
   sentence and EPSS phrase lines clean, so dropping `visible` from any of them survives (measured U3). And no test
   covers the forged-line half of property 2 (F1 above passes the shipped test).
6. **ADR-0027 WHY comments — HOLDS.** WHY at the escaping (`format.go:229-244`) and at the env line
   (`run.go:330-332`); tests name their property in their doc comments and names.

### Unplug
- U1 — `escapeControls` returns `s` unchanged → `TestTextOutputCarriesNoControlByteFromTheChange` FAIL
  (`raw control byte U+001B at offset 39`). Red. 
- U2 — `cmd.Env = childenv.Minimal()` removed at `run.go:333` → `TestGitMergeBaseChildCarriesNoKeyFromProven`
  FAIL (`the git child inherited Proven's key`). Red.
- U3 — `visible` removed from `ReviewerStatements`, `TracePath` and `failed.Service` together → the escape test
  PASSES. Survivor: names a missing fixture value (condition C2).
- F1 has no unplug at all: there is no test of the forged-line property.
- Each mutant was reverted with `git checkout -- internal/cli/format.go internal/cli/run.go`; `git diff --stat` empty afterwards.
- Raw evidence: `█████████████████████████████████████████████████████████████████████████████`.

### Closure
- Trust boundary map (S1): (a) change → Prover test/query evidence → text writer → reader's terminal. Tamper of the
  terminal by control bytes is closed for every field (P1 holds); spoofing of Proven's own lines via the kept
  newline in `test.Output` is OPEN (F1). (b) Proven's environment → git child: disclosure closed (P3 holds).
- S2: F1 reproduced with a probe test driving `textFormatter.Write` directly (output above).
- S4: no dependency/manifest change in the diff; no new secrets; the diff itself scanned by reading — no
  invisible or bidirectional characters in the added lines (the test constant uses `\u009b` as an escape in source).
- R13 pairs: `visible` vs `visibleLines` judged together — the pair differs only in the newline, and that is where F1 lives.

### Remainder
- Bidi and other format characters (U+202A–U+202E, U+2066–U+2069, U+200B, U+2028/U+2029) pass `visible` raw. They
  are outside the byte range the order claims (C0/DEL/C1/invalid UTF-8), so not a finding against this wagon, but
  a reordering override in a path or title can make a terminal display text that reads differently from its bytes. Owed, for the master to rule on.
- Backslash not escaped → `\xNN` output ambiguous with literal input (Differential §2). Owed.
- When JSON/SARIF writers land: `encoding/json` leaves DEL and C1 raw in UTF-8; the comment at `format.go:233-234`
  would then be false. Owed to that future wagon, not this one.
- Invalid UTF-8 is replaced by U+FFFD rather than shown as `\xNN` — executor claim inexact; safe.
- stderr (declined to judge, outside property 1's list): `run.go:320` prints git's stderr and the baseline/head
  verbatim; `run.go:92` prints reviewer errors verbatim; `run.go:99` prints target/baseline/head. Their bytes
  come from CI variables, the event file and git, not obviously from the change; I did not trace whether any
  reviewer error embeds change content.

### Profile
security — S1 trust map, S2 reproduction, S4 delta/character scan performed above; S3 (authorization checks) not applicable: the diff adds none.

### Focus
The two properties of the order (P4-F5 terminal output carries no control byte from the change; P2-F3 no key reaches
a subprocess) and the six items of WHAT TO JUDGE — nothing beyond the diff `0b954a9...1e77598`.

### Conditions
- **C1 (blocking, F1)** — `internal/cli/format.go:203` / `visibleLines`: a continuation line of a multi-line field
  must not be able to read as a line of Proven's own output. Prefix every kept newline with a constant marker no
  Proven line starts with, and add a test whose fixture `Output` holds `"\nfailed services: none"` (and a
  `finding … ERROR` line) and which asserts that no output line other than Proven's own starts at column 0 with
  those keywords; unplug: remove the prefix → red.
- **C2 (owed, test coverage)** — `internal/cli/terminal_escape_test.go:20-37`: give every printed field a value
  carrying `raw` (Class, ClassRule, the three statement kinds, Service, TracePath, an EPSS phrase), so that removing
  `visible` from any one call site turns the test red; unplug: U3 must go red.

### Unrun
- Full suite on this host: `cmd/livecheck`, `internal/generator`, `internal/guardrail` FAILED only on
  `listen tcp6 [::1]:0: bind: operation not permitted` (sandbox refuses listeners); they were not run green here,
  so the master's "37 packages ok, 0 FAIL" is unverified by this review.
- Property 4 byte-identity was not measured by comparison: no JSON or SARIF writer exists to produce bytes.
- Not run against a real terminal emulator; escaping was judged on the bytes written.
- Housekeeping: the probe file `████████████████████████████████████` (untracked, test-only, logs and always
  passes) could not be removed — this leg's permission class does not hold `rm`, and law #57 forbids composing
  the deletion from another verb. It must be deleted by the master before teardown; its content is mirrored in
  `█████████████████████████████████████████████████████████████████████████████`.

## The review status as filed

## REVIEW — cli-escapes-output-strips-git-env
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: reject at 1e77598 — P1 (every text field escaped), P3 (git child gets childenv.Minimal; fork-point test green), P4 (JSON/SARIF untouched) and P6 hold; both tests go red when unplugged.
  P2 fails: visibleLines keeps the newline at column 0, so test output "\nfailed services: none" prints a line indistinguishable from Proven's own (C1, format.go:203, blocking).
  C2 owed: the escape test leaves Class, statements, Service, TracePath unpopulated — removing visible() there survives.
  Unrun: 3 packages red only on a sandbox listen denial; probe file ████████████████████████████████████ left for the master to delete.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T00:00:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 2 |
