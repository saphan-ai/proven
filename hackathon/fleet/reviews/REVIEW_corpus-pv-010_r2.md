# REVIEW — corpus-pv-010 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 3 OWED · 4 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:36:45Z

## Summary

The PV-010 fixture builds §2's condition. Base and change differ only in the cannot-convert branch and the stated width, and the false world can be told apart for -999, twenty-digits and the description.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-010

### Ground

- Reviewed: `fe146e0...4570315`, one commit, `[corpus-pv-010] tests: build PV-010 base, change and check, run it`.
- `git rev-parse --short=7 HEAD` printed `4570315`; `git rev-parse HEAD` printed `457031534a09d291251572cb9e4f87577070fcb0`. This matches the order's `reviewed-sha`.
- Seat toolchain: `go version go1.26.5 linux/amd64`.
- Read: `CASE.md` (§2 first), `check.sh`, `base/`, `change/`, `reference/`, the manifest diff. `expected/EXPECTED.md` was read only after the code and the check script.
- Run on this seat, each from the repository root:
  - `go -C tests/corpus/cases/PV-010/change test -count=1 -v ./...` → `TestCheck` (current, stale, overflow, no_arrow) and `TestDescriptionFollowsMatcher` PASS, `ok pv010/change`.
  - `go vet ./...` in `base/`, `change/` and `reference/` (via `go -C`) → no output, exit 0 for each.
- **`bash tests/corpus/cases/PV-010/check.sh` was NOT run.** See Unrun. The check output quoted in this review is the author's, from `CASE.md` §8 (run 2026-10-22T15:46Z, `go1.26.8 darwin/arm64`). I treat it as a claim. I did not re-measure it.
- Author self-assessment (CASE.md header "built inside the build window", §8 "exit code 0") is set aside under R1.

### Differential

Item 1 — does base differ from change exactly where §2 says?

`diff base/main.go change/main.go` → no difference. `base/guard.go` against `change/guard.go`, read line by line:
- Same matcher in both: `LANGUAGES[%s]{0,%d}-> ?([0-9]+)` with `gapClass = "0-9 "` and `matcherMaxGap = 48`. A sign or non-ASCII digits find no pair in both. A note with no pair passes in both (`continue` on `sub == nil`). This matches §2.
- `Check`: base `if err != nil || v == pinned { continue }` (base/guard.go). Change refuses on `err != nil` and on `v != pinned`. This is the three-way split §2 names. It is the only behavioural difference in `Check`.
- `Describe`: base derives the characters only and states no width. Change derives the characters from the matcher and the width from `statedMaxGap = 48`. Nothing else reads that constant (grep: it appears only at its definition and in `Describe`). This matches §2.
- `change/guard_test.go` `TestDescriptionFollowsMatcher` changes `gapClass` (the matcher), not the width. So it cannot detect an edit to `statedMaxGap`. This matches §2 and §4.
- change-edited: `check.sh` lines 23–27 copy `change/` and run an anchored `sed` on `^const statedMaxGap = 48$` only. `cmp -s` refuses if nothing changed. Only that one line differs.
- `reference/main.go`: `readNote` follows §2's token rule (text after `->`, up to a space, trailing `.,;:!?` removed, current only if it equals the pinned count). `readWidths` probes widths 0..256 with `LANGUAGES` + w spaces + `-> 41` against the change binary. It counts a width as accepted when the output starts with `FAIL`. The base matcher is the same pattern, so probing one binary covers both. CASE.md §8 says so.
- Inputs in `check.sh` lines 44–49 are §2's list exactly: current, `41 -> 40`, 20 digits, `-999`, `+41`, Arabic-Indic `41` (`\xd9\xa4\xd9\xa1` = U+0664 U+0661), plus the description line for base, change and change-edited.

According to the author's §8 output, base and change differ on the `twenty-digits` line (base=pass, change=fail) and on the description (base=none, change=0..48). That is where §2 puts them, and nowhere else. **Conclusion from reading the code: the fixture builds the condition §2 names. I did not re-measure the printed lines (Unrun).**

### Unplug

Item 2 — would the same output also fit the opposite outcome (law #52)?

- `minus-999` / `plus-41`: if change had no defect (it compared signed values), change would print `fail`. The printed `change=pass` beside `reference=stale` separates the two worlds.
- `twenty-digits`: if the earlier repair were not real, change would print `pass`. `change=fail` separates the two worlds.
- description: if the width were derived from the matcher, change-edited would state `0..48`, or its tests would go red. `change-edited=0..256`, `reference=0..48`, `change tests on change-edited=pass` separate the two worlds.
- `verdict()` (check.sh line 32) lower-cases the first output line. A crashed binary prints an empty field, not `pass`. So a crash cannot read as a pass.
- **One weak line, `arabic-indic-41`:** the guard passes any closing token that is not ASCII digits, and the reference reads any token other than `42` as stale. If `printf` at line 49 wrote the escape text literally instead of the two Arabic-Indic code points, the line would still print `base=pass | change=pass | reference=stale`. The line cannot show that the note really held other-script digits. In bash, `printf '\xHH'` does produce the bytes, so this is very likely fine in practice. But the output does not prove it. NOTE.
- `-999` is the line EXPECTED.md leans on most, and the line above does not weaken it.

Item 2 against `expected/EXPECTED.md`: every piece of evidence it asks for appears in the §8 output: the signed and other-script lines with change=pass and reference=stale; twenty-digits with base=pass and change=fail; the description with change-edited=0..256, reference=0..48 and change tests on change-edited=pass. §8 does not claim the things EXPECTED.md forbids: it does not say `-999` forms a pair, does not say the three-way split is wrong, and does not say a published note carried such a number. So the expected outcome follows from the evidence, **on the author's run**.

### Closure

Item 3 — was a check bent? None found.
- No skip, no loosened assertion, no edited threshold. `TestCheck` leaves out `-999`/`+41`. That is the defect the case preserves (§6 reviewer A), not a bend.
- The check's exit code is 0 whatever the guard does. §8 says this openly ("reports that it ran, not a verdict").
- Mutation of change's guard was reasoned, not run (Unrun). Reverting the `err != nil` refusal turns `TestCheck/overflow` red. Inverting `v != pinned` turns `TestCheck/current` and `TestCheck/stale` red.

Item 4 — nothing private. `grep -rnE "/Users|/home|@|macm|███████|token|secret|bridge|studio|<IPv4>" tests/corpus/cases/PV-010 --exclude-dir=expected` found only the domain word "closing token" (CASE.md 41, 42, 133; reference/main.go 27, 28). There are no host or seat names, private paths, e-mails or credentials. `████████████████` is a path relative to the repository.

Item 5 — lane. `git diff --name-status fe146e0...4570315`: `M CASE.md`, `A` files under `base/`, `change/`, `reference/` and `check.sh`, `M tests/corpus/manifest.json`. Nothing under `expected/`. The manifest hunk is one line inside the `"id": "PV-010"` entry (id at line 176, change at line 187). **The master's claim (lane only, `expected/` untouched) holds. The claim "check.sh exit 0" was not re-measured.**

Item 6 — portability, by reading (not run on Linux or macOS):
- `bash` builtin `printf '\xHH'`, `cd … && pwd`, `cp`, `sed -e` (no `-i`), `cmp -s`, `tr 'A-Z' 'a-z'`, `head -1`, `rm -rf`, `mkdir -p`: all work on GNU and BSD/macOS. `grep -o` is not POSIX, but GNU grep and BSD grep (macOS) both have it.
- **Non-portable dependency: `go` ≥ 1.26 on the PATH.** All three `go.mod` files say `go 1.26`. A judge with an older Go and the default `GOTOOLCHAIN=auto` triggers a toolchain download. That is network access, and §2 says "no network". §8 records the author's toolchain but does not state this minimum. NOTE.
- `check.sh` line 10 writes scratch files into `<repo>/.saphan/work/corpus-pv-010/`. The repository `.gitignore` does not list `████████`. `git check-ignore` needed approval on this seat and was not run. So after one run, a judge's clone shows an untracked directory. Line 13's `rm -rf` targets fixed subpaths under the repository root, so its reach is bounded. NOTE.

### Remainder

- OWED-1 · `tests/corpus/manifest.json:188` keeps `"checks": "not run"` for PV-010. But `CASE.md:5` says `checks: run 2026-10-22`, §8 records a full run, and the commit subject says "run it". The record contradicts itself in the same commit. Remedy: set the manifest's `checks` to the run date, or explain in CASE.md why it stays `not run`. Whoever owns the corpus vocabulary rules on this.
- NOTE-1 · `check.sh:49`: the `arabic-indic-41` line cannot show what bytes it fed (Unplug). Remedy: print the input's hex, or have the reference name the token's script.
- NOTE-2 · the `go 1.26` minimum is not stated in §8, and an older Go means a network fetch. Remedy: one sentence in §8, or `GOTOOLCHAIN=local` in `check.sh` so the failure is named instead of a download.
- NOTE-3 · `check.sh:10` scratch directory is inside the repository and not ignored. Remedy: `mktemp -d` with a `trap` cleanup, or a `.gitignore` line. Note that `.gitignore` belongs to the owner's repo; it is not under `.github/`, so it is not off-limits by AGENTS.md.
- Declined to judge (R11): whether the Arabic-Indic `42` should count as current (CASE.md §9 already names this). Licences (out of scope by order).

### Profile

Test fixture / corpus case. Generic reviewer acts R1–R15 plus testing-bar: no declared metric for this case. The bar is made of suite checks only (`change/guard_test.go`), and the case is designed so that suite cannot catch its defect. Check ranking: suite. No security, backend, ui or mobile surface. Docs D4 (re-run every quoted command): not met by this review, because `check.sh` is Unrun.

### Focus

Order items 1–6, in the order given: the fixture matches §2 (Differential); the expected outcome follows and the false world can be told apart (Unplug); no bend, nothing private, lane kept, portability (Closure). The weight is on `minus-999` and the description line, which carry the EXPECTED finding.

### Conditions

1. OWED · `tests/corpus/manifest.json:188`: bring `checks` in line with `CASE.md:5` and §8, or record why not. Not blocking: the case's evidence is unaffected.
2. OWED (gate-side, master) · run `bash tests/corpus/cases/PV-010/check.sh` on Linux at `4570315`, paste the output, and compare it line by line with `CASE.md` §8. This review could not run it (Unrun). The case's claim rests on that output.
3. NOTE-1, NOTE-2 and NOTE-3 from Remainder: judgement, not blocking (R6).

### Unrun

- `bash tests/corpus/cases/PV-010/check.sh`: `bash` is not in this role's command grant. The hook's refusal on this leg listed the grant as `[saphan doc check go git ls cat grep find wc head tail diff pwd cut sort]`, under permission class `██████████████████████████`. Separately, the script writes `███████████████████████████` at the repository root, which is outside this leg's confinement (`█████████████████████████████████████████`). This grant cannot remove that directory, so the worktree could not be left as found. I did not rebuild the check from other commands (R15, law #57). So the check output pasted under item 1 is the author's §8 output, not mine.
- The check on macOS: no macOS seat.
- `git check-ignore -v ███████████████████████████████████`: needed an approval this headless leg cannot give.
- Mutation of change's guard and test (R3): needs a copy or an edit of the code under review; reasoned only (Closure).

## The review status as filed

## REVIEW — corpus-pv-010
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The PV-010 fixture builds §2's condition. Base and change differ only in the cannot-convert branch and the stated width, and the false world can be told apart for -999, twenty-digits and the description.
  Lane kept: expected/ untouched, one manifest line. Nothing private found. change tests and go vet pass on go1.26.5 linux/amd64.
  check.sh is UNRUN: bash is not in this grant, and the script writes outside this leg's confinement. The pasted output is the author's §8, not re-measured.
  OWED: manifest.json:188 still says checks "not run" against CASE.md:5 and §8, and the master runs check.sh on Linux before the gate.
  NOTEs: the arabic-indic line cannot show its input bytes, the go>=1.26 minimum (else a network fetch) is unstated, and the scratch dir in the repo is not ignored.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T18:55:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 5 |
| e-mail or user name | 1 |
| policy binding name | 1 |
