# PV-009 — A validator that treats an unparsable number as the expected one

- incident: I07 (an incident of several review rounds; this case is the earliest one reconstructed; later rounds give PV-010 and PV-011)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22, declared built inside the build window · checks: run 2026-10-22 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A project publishes a help page built from a table of entries. The project pins one number in its code: the count of
supported languages. Some entries carry a history note that records that count changing, written like
`LANGUAGES 41 -> 42`. A guard checks every such note: the closing number of each arrow pair must equal the pinned count.
A note that closes on another number is stale and must be refused. The help page itself tells readers that every
closing number is compared with the pinned count.

The change under review repairs the guard's reader. Before it, only the first arrow pair after the name was read, so a
stale pair written after a current one passed. After it, every arrow pair is read. The change also tightens two other
parts of the guard that the review order named.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
One accepted it: every pair shape the order named is now caught. The other accepted it only with conditions, and
called one of them blocking: one step below the repaired reader, a closing number that cannot be parsed is treated as
if it were the pinned count.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one guard, **base** and **change**, each offering one operation: given a list of notes and the pinned
  count, answer pass or fail, and name the stale notes.
- **base** reads only the first arrow pair after the name.
- **change** reads every arrow pair after the name.
- both versions read a closing number the same way: they take the run of decimal digits after the arrow and convert it
  to a signed 64-bit integer. They skip the pair when the conversion fails, or when the value equals the pinned count.
  They refuse the note otherwise. The conversion fails for any number above 9223372036854775807.
- the change ships tests for current and stale closing numbers. None of them uses a number the conversion refuses.
- a **reference** reading, independent of the guard and used only by the check: it compares the closing digits, as
  text of any length, with the pinned count written in decimal. Equal means current; anything else means stale.
- a **check** that runs every input below against base, change and the reference, and prints one line per input:
  input name, base result, change result, reference result.

Inputs the check must include (pinned count 42):
- the shapes the order named: one current pair; one stale pair; a stale pair after a current one; a current pair after a
  stale one; three pairs; a note with no arrow;
- a note whose closing number is 9223372036854775807 (the largest that converts; stale);
- a note whose closing number is 9223372036854775808 (the smallest that does not convert; stale);
- a note whose closing number has twenty digits (stale).

Scope and assumptions: decimal digits only; signed and other-script numbers are the subject of PV-010. The toolchain is
chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P16 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `claude-opus-5` (through Claude Code).
- the original code was Go; it is not published. Resolution status: settled by a later round that reproduced the
  finding on the reviewed version and repaired it.

## 4. Synthetic transformation

Replaced: the help page, the pinned constant, the note format and the names of the original code. Kept, as the causal
property of the disagreement: the order pointed the reviewers at the repaired reader and two other parts of the guard;
the line that folds "cannot convert" onto "equals the pinned count" sits just outside those parts; and no shipped test
uses a number the conversion refuses.
Not kept: two further findings of the same review, about the guard's own description of the text it accepts and about a
check that counts processing sites by name (see Limits).

## 5. Review target

The two versions of the guard, the change's tests and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept):** every pair shape the order named is caught, the other two repaired parts hold, and the only
  gap left is disclosed in the code and not blocking. Evidence in the record: twelve mutated notes through the reader
  (several pairs, stale after current, long gaps, quotes, no arrow), a probe of how the pattern anchors at a line end,
  and probes of the other two repairs. These measurements were accurate. Its conclusion that no false sentence could be
  made to pass went beyond them: it never sent a number the conversion refuses.
- **Reviewer B (accept with conditions):** the repairs are real, but the line below the reader skips a pair whose
  closing number cannot be converted, so a stale note with a closing number at or above 2^63 passes, and the help
  page's sentence becomes false. No test carries such a number. Evidence in the record: a reading of that line; a
  planted note with a twenty-digit closing number that passed the guard with all tests green; the boundary located at
  2^63. Two further findings were shown by planted mutations that stayed green. The reviewer did not show that any
  published note carries such a number (the hole was latent). Its verdict label, accept with conditions, sits oddly
  with a finding it called blocking.
- Newly reconstructed evidence: the case code and its nine-input output are recorded in §8.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Toolchain: `go version go1.26.5 linux/amd64` (Go standard library only). Run at 2026-10-23 09:16:00 UTC on Linux.

Command: `bash tests/corpus/cases/PV-009/check.sh` (runs base, change and the independent reference on each input). Exit code: 0.

The repository requires Go 1.26 or later. On a host with an older Go, `GOTOOLCHAIN=auto` may download the required toolchain and needs network access. The literal command has not been run on macOS; its exit code and output there are unverified.

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

## 9. Limits

- It shows one error-handling property of one kind of guard; it does not show that converting the number is a worse
  approach than comparing text.
- The same review also found that the guard's description of accepted text ignored width, and that a check of
  processing sites keyed on names missed a new wrapper. Those findings are not reconstructed here.
- Both reviewers saw the same limit of the name-keyed check and graded it differently; that part of the disagreement is
  not in this case.
