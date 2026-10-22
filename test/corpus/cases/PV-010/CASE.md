# PV-010 — A stale-value check whose number pattern silently skips signed values

- incident: I07 (an incident of several review rounds; this case is the round after PV-009; a later round gives PV-011)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22 (declared: built inside the build window) · checks: run 2026-10-22 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A project publishes a help page built from a table of entries. The project pins one number in its code: the count of
supported languages. Some entries carry a history note that records that count changing, written like
`LANGUAGES 41 -> 42`. A guard checks every note: the closing number of each arrow pair must equal the pinned count, or
the note is stale and refused. The help page tells readers that every closing number is compared with the pinned count.
It also describes what text the guard accepts between the name and the arrow, including the widths it accepts.

The previous version of the guard treated a closing number it could not convert as current. The change under review
splits that outcome in three: a number that differs is refused, an equal number is accepted, a number that cannot be
converted is refused. The help page gains a sentence saying so. The change also adds the accepted widths to the page's
description, and a test that the description and the guard agree.

Two reviewers from different model families reviewed the same change. Both saw that a note closing on `-999` finds no
arrow pair at all. One read that as the digit pattern correctly filtering input, and accepted. The other read it as a
stale note escaping a comparison the page promises, and rejected.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one guard, **base** and **change**, each offering two operations:
  1. given a list of notes and the pinned count, answer pass or fail, and name the stale notes;
  2. produce the page's description of what the guard accepts, including the accepted gap widths.
- in both versions, an arrow pair is found only when the arrow is followed by a run of ASCII decimal digits. A sign, or
  digits of another script, means no pair is found, and a note with no pair passes.
- **base** converts the digits to a signed 64-bit integer and skips the pair when the conversion fails or the value
  equals the pinned count. Its description names the accepted characters only; it states no widths.
- **change** refuses a pair that differs, accepts one that is equal, and refuses one that cannot be converted.
- in **change**, the character part of the description is derived from the matcher, but the width part is written from a
  separate constant that nothing reads back. The matcher itself accepts gaps of up to 48 characters.
- **change** ships a test that alters the matcher and checks that the description still agrees with it.
- a third version, **change-edited**: change with only the description's width constant edited (it states widths up to
  256). Nothing else differs.
- a **reference** reading, independent of the guard and used only by the check. For notes: the closing token is the text
  after the arrow up to the next space, without trailing sentence punctuation; the note is current only if that token
  is exactly the pinned count in ASCII digits; anything else is stale. For descriptions: it probes the matcher with gaps
  of every width from 0 to 256 and states the widths it accepts.
- a **check** that runs every input below and prints one line per input: input name, base result, change result,
  reference result. For the description input it prints the stated widths for base, change and change-edited, the
  reference widths, and whether change's own tests pass on change-edited.

Inputs the check must include (pinned count 42):
- a current note; a stale note (`41 -> 40`); a note with a twenty-digit closing number;
- a note closing on `-999`; a note closing on `+41`; a note closing on 41 written in Arabic-Indic digits;
- the page description, for base, change and change-edited.

Scope and assumptions: the toolchain is chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P18 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `claude-opus-5` (through Claude Code). The
  rejecting review was a second run of its family on the same commit, started about twenty minutes later, after the
  first run ended without a review.
- the original code was Go; it is not published. Resolution status: settled by a later round that adopted all three
  findings of the rejecting review and repaired them.

## 4. Synthetic transformation

Replaced: the help page, the pinned constant, the note format, the description's wording and the names of the original
code. Kept, as the causal property of the disagreement: the digit pattern filters out signed and other-script numbers
before any comparison, while the page promises that every closing number is compared; and the change's agreement test
alters the matcher, which moves the derived part of the description with it, so it cannot catch an edit to the part
written from a constant.
Not kept: a third finding of the same review, about a check of processing sites that does not follow a value into the
function it is passed to (see Limits).

## 5. Review target

The two versions of the guard, the change's tests and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the reference, change-edited or
`expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept):** all three findings of the previous round are fixed. Signed values give "no pair found", so
  only an overflowing number reaches the cannot-convert branch, which is exactly what the page says. Stated and enforced
  widths cannot differ. Evidence in the record: each repair reverted in turn, with its tests going red; a table of
  closing values (current, stale, overflowing, `+999`, `-999`, separators, hexadecimal, empty), in which `-999` and
  `+999` gave no pair; a mutation of the matcher with the description following it. These measurements were accurate.
  Two conclusions went beyond them: it treated "no pair found" as safe without asking whether the page's promise covers
  such a note; and its width test altered the matcher only, so it could not fail.
- **Reviewer B (reject):** each repair's published sentence is wider than the repair. A note closing on `-999` (also
  `+41` and other-script digits) finds no pair and passes, though the page says the closing number is compared. The
  width part of the description comes from a constant nothing reads back, so one edit to it publishes a false width
  with every test green. Evidence in the record: `-999` planted in a real note, passing with all tests green, against
  `999` in the same note, turning five tests red; a search showing the width constant is read only where it is defined;
  an edit of the description's width with all tests green; a third finding shown by a planted mutation with a negative
  control. It showed that the page and the guard disagree; it did not by itself decide whether the guard should grow or
  the page should shrink.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base, change and change-edited and record each line of its output.

Run 2026-10-22T15:46Z (UTC). Toolchain: `go version go1.26.8 darwin/arm64`, standard library only; each of `base/`,
`change/` and `reference/` is its own module. change-edited is built by `check.sh` into
`.saphan/work/corpus-pv-010/change-edited/` as a copy of `change/` with only the line `const statedMaxGap = 48` replaced
by `const statedMaxGap = 256`.

Command, from the repository root: `bash tests/corpus/cases/PV-010/check.sh` — exit code 0. Full output:

```
toolchain: go version go1.26.8 darwin/arm64
pinned count: 42
current | base=pass | change=pass | reference=current
stale | base=fail | change=fail | reference=stale
twenty-digits | base=pass | change=fail | reference=stale
minus-999 | base=pass | change=pass | reference=stale
plus-41 | base=pass | change=pass | reference=stale
arabic-indic-41 | base=pass | change=pass | reference=stale
description | stated widths: base=none | change=0..48 | change-edited=0..256 | reference=0..48 (49 widths) | change tests on change=pass | change tests on change-edited=pass
```

The matcher is `LANGUAGES[0-9 ]{0,48}-> ?([0-9]+)`; the reference's width probe runs `change`'s matcher (base's is the
same pattern). The check's exit code reports that it ran, not a verdict on the guard.

## 9. Limits

- It shows that one guard and one published promise disagree; it does not show which of the two should change.
- The reference's rule for the closing token is the one the later round adopted; another reasonable rule could read an
  other-script `42` as current.
- The same review also found that a check of processing sites stopped at the call and never looked inside the called
  function. That finding is not reconstructed here.
