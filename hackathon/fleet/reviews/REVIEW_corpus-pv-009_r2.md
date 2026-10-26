# REVIEW — corpus-pv-009 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 2 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T19:39:15Z

## Summary

PV-009 at a0987ed reproduces §2. base and change differ only on stale-after-current and three-pairs. The overflow and twenty-digit lines show change=pass and reference=fail; max-int64 shows change=fail. The check prints all four lines EXPECTED.md requires, and each could print the other way. The re-run matches the §8 record verbatim (`grep -cvxF` printed 0).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-009

### Ground
- `git rev-parse --short=7 HEAD` printed `a0987ed`. Scope `fe146e0...a0987ed`, one commit: `a0987ed [corpus-pv-009] tests: build overflow validation corpus`.
- `go version` printed `go version go1.26.5 linux/amd64`. That is the toolchain §8 records.
- `bash tests/corpus/cases/PV-009/check.sh` was REFUSED by this leg's permission class (`bash` is not in the review role's program list). I did not run the wrapper. I ran its one operative line, `go run ./tests/corpus/cases/PV-009/check.go`, from the repository root, which is the directory the wrapper `cd`s to. `go` is in the review grant. The wrapper's own lines (`#!/usr/bin/env bash`, `set -euo pipefail`, `cd "$(dirname "$0")/../../../.."`) stay under Unrun.
- Output of `go run ./tests/corpus/cases/PV-009/check.go` (exit 0), verbatim:
```text
one-current base=pass change=pass reference=pass
one-stale base=fail:LANGUAGES 41 -> 43 change=fail:LANGUAGES 41 -> 43 reference=fail:LANGUAGES 41 -> 43
stale-after-current base=pass change=fail:LANGUAGES 41 -> 42; 42 -> 43 reference=fail:LANGUAGES 41 -> 42; 42 -> 43
current-after-stale base=fail:LANGUAGES 41 -> 43; 43 -> 42 change=fail:LANGUAGES 41 -> 43; 43 -> 42 reference=fail:LANGUAGES 41 -> 43; 43 -> 42
three-pairs base=pass change=fail:LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43 reference=fail:LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43
no-arrow base=pass change=pass reference=pass
max-int64 base=fail:LANGUAGES 41 -> 9223372036854775807 change=fail:LANGUAGES 41 -> 9223372036854775807 reference=fail:LANGUAGES 41 -> 9223372036854775807
overflow-int64 base=pass change=pass reference=fail:LANGUAGES 41 -> 9223372036854775808
twenty-digits base=pass change=pass reference=fail:LANGUAGES 41 -> 10000000000000000000
```
- `go run ./tests/corpus/cases/PV-009/check.go | grep -cvxFf tests/corpus/cases/PV-009/CASE.md` printed `0`: each of the 9 output lines appears verbatim as a whole line in CASE.md, and the §8 block holds exactly those 9 lines. The record in §8 matches the re-run.
- `go test ./tests/corpus/cases/PV-009/...` gave `ok .../PV-009/change`. `go vet ./tests/corpus/cases/PV-009/...` printed nothing.
- Lane: `git diff --stat fe146e0...a0987ed` lists 7 files, all under `tests/corpus/cases/PV-009/` except `tests/corpus/manifest.json`. `git diff --name-only fe146e0...a0987ed -- tests/corpus/cases/PV-009/expected` printed nothing. The manifest hunk (`-U12`) changes only the `fixture` and `checks` values inside the `"id": "PV-009"` object. The master's claim (own directory and manifest line only, never `expected/`, exit 0) holds, with the exit 0 measured on `go run`, not on `bash check.sh`.

### Differential
Property 1. The fixture reproduces §2.
- `base/guard.go:18` uses `arrow.FindStringSubmatch(after)`, which reads the first pair only. `change/guard.go:18` uses `FindAllStringSubmatch(after, -1)`, which reads every pair. That is §2's base/change split.
- Both versions use the same conversion. `base/guard.go:22-25` and `change/guard.go:19-21` call `strconv.ParseInt(pair[1], 10, 64)` and then `if err != nil || closing == pinned { continue }`. This is §2's "skip when the conversion fails, or when the value equals the pinned count", in the same words in both files.
- base and change differ on exactly two lines: `stale-after-current` and `three-pairs` (base=pass, change=fail). These are the inputs whose first pair is current and a later pair is stale. `current-after-stale` is fail in both, because its first pair is already stale. That is correct, not a missed difference. The overflow lines read the same in base and change, as §2 says (one shared conversion). There is no difference anywhere else.
- The reference in `check.go:13-30` compares the digit text with `strconv.FormatInt(pinned, 10)`. It never converts, so it is independent of the guard's conversion, as §2 asks.
- The change's tests (`change/guard_test.go:11-16`) cover current, stale, stale-after-current, current-after-stale, no arrow and 9223372036854775807. None of them uses a number above 2^63-1. That matches §2's "None of them uses a number the conversion refuses".
- All 9 inputs §2 requires are in `check.go:44-53`, with pinned count 42 (`check.go:56`).

Property 2. The expected outcome follows from the evidence. `expected/EXPECTED.md` names the outcome "established defect in change (also present in base)" and asks for four lines. All four are in the output:
- `overflow-int64` and `twenty-digits`: change=pass and reference=fail.
- `max-int64`: change=fail. This is the boundary.
- `stale-after-current`: base=pass and change=fail. This shows the repair is real.
- "also present in base" holds too: base=pass on both overflow lines.

### Unplug
- False-world answer for the outcome. If the change refused a closing number it cannot convert, `overflow-int64` and `twenty-digits` would print `change=fail:...`. If the boundary were not at 2^63 (for example a 32-bit conversion), `max-int64` would print `change=pass`. If the reader repair were not real, `stale-after-current` would print `change=pass`. Each of the four lines EXPECTED.md names can come out the other way, so the printed output tells the two outcomes apart. It does not fit the opposite outcome.
- False-world answer for my comparison instrument. If any line of the re-run differed from §8, `grep -cvxF` would print a count above 0. It printed 0.
- Reach (law #52). The output shows WHAT the guard decides, not WHY. The cause, "cannot convert is read as equals the pinned count", comes from reading `change/guard.go:20` (`err != nil || closing == pinned`), not from the output alone. The boundary sitting exactly at 2^63 fits that cause and does not prove it. The wrong reading this case invites is that a published note carried such a number. EXPECTED.md already forbids that claim.
- I did not run a mutation (for example, change the `err != nil` branch to refuse the note and watch the two overflow lines turn to `change=fail`). This role cannot remove files, and the order requires a clean worktree. It is listed under Unrun. The false-world answer above is read from the code, not executed.

### Closure
- Property 3, is a check bent? No. `check.go` prints the raw guard results through one `result()` function (`check.go:32-37`). There is no per-input expectation, no filter and no special case. The reference arrow pattern `->\s*([0-9]+)` (`check.go:13`) is looser than the guard's `[0-9]+\s*->\s*([0-9]+)`: it does not need digits before the arrow. On these 9 inputs every arrow has digits before it, so the two patterns pick the same closing numbers. That looseness does not move any line.
- Property 4, nothing private. `grep -rnE 'macm|███████|████████████████████████████████████' tests/corpus/cases/PV-009`, outside `expected/`, found nothing. The module path `github.com/saphan-ai/proven` is the public module in `go.mod`. CASE.md §3's "the fleet's private record" was there before this commit, so it is not in this diff.
- Property 5, the diff stays in its lane. It holds. See Ground.

### Remainder
- NOTE. CASE.md:86, in §6, still says "Newly reconstructed evidence: none yet; it is built with the case code." This commit built the case code and filled §8, but did not update this sentence. It is now stale. This review does not block on it.
- NOTE, portability (property 6). `check.sh` needs `bash`, `dirname` and `go` on the path. `#!/usr/bin/env bash` and `set -euo pipefail` work in macOS's stock bash 3.2. `"$(dirname "$0")"` is quoted. The Go program uses only the standard library and RE2 `\s`. Its output does not depend on map order, so it is deterministic on any OS. The one dependency to name: `go.mod` declares `go 1.26`. A judge with an older Go either fails, or, with `GOTOOLCHAIN=auto`, downloads a toolchain, which needs the network. I found no Linux-only call. I did NOT run it on macOS.

### Profile
claude-code, model claude-opus-5-5. Review leg corpus-pv-009-review-r2, one of two independent reviews from different families. Read-only on the code under review. Programs used: `git`, `go`, `grep`, `cat`.

### Focus
The six properties of the order: reproduction of §2 (1), evidence to outcome with the false-world answer (2), no bent check (3), privacy (4), lane (5) and portability (6). No licences.

### Conditions
- C1 (OWED). Run the literal `bash tests/corpus/cases/PV-009/check.sh` once on Linux and once on macOS. Confirm exit 0 and output identical to §8. This leg could run only the wrapper's `go run` line (bash refused by permission class). Class OWED, not BLOCKING: the wrapper is four lines and does nothing to the output beyond `cd`.
- C2 (NOTE). Update CASE.md:86 (§6 "Newly reconstructed evidence: none yet") to point at §8.
- C3 (NOTE). State the Go ≥ 1.26 requirement, and the toolchain download that `GOTOOLCHAIN=auto` causes, next to the §8 command, for judges on an older Go.

### Unrun
- `bash tests/corpus/cases/PV-009/check.sh` as a literal command. Refused: `bash` is not in permission class ███████████████████████████ Its `go run` line was run in its place (see Ground). The shebang, `set -euo pipefail` and the `cd` line were not executed.
- The check on macOS. No macOS host is available to this leg.
- The executed unplug mutation of `change/guard.go:20`. Writing a mutated copy would leave files this role cannot remove, and the order requires a clean worktree. The false-world answer was derived by reading the code.

## The review status as filed

## REVIEW — corpus-pv-009
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-009 at a0987ed reproduces §2. base and change differ only on stale-after-current and three-pairs. The overflow and twenty-digit lines show change=pass and reference=fail; max-int64 shows change=fail. The check prints all four lines EXPECTED.md requires, and each could print the other way. The re-run matches the §8 record verbatim (`grep -cvxF` printed 0).
  The diff stays in its lane and nothing private is in it. Conditions: C1 OWED, run the literal `bash check.sh` on Linux and macOS (bash was refused here; its `go run` line was run in its place). C2 and C3 are NOTEs: a stale §6 sentence, and the Go 1.26 requirement.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T16:10:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| policy binding name | 1 |
