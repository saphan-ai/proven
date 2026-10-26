# REVIEW — corpus-pv-018 · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T16:52:21Z

## Summary

The check reproduces a passing name-list mutant and the planted-row differential; the expected test-gap finding follows. Fix the stale command and verify macOS.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-018

### Ground
Observed HEAD `b079dc8` by `git rev-parse --short=7 HEAD`. I read CASE.md §2 and the implementation before expected/EXPECTED.md. The three-dot diff `9303b2d...b079dc8` changes only PV-018 outside expected/ and PV-018 status fields in test/corpus/manifest.json:322-326. The base always uses 200 bytes (base/validator.go:11); change selects 120 when the row cites a measurement (change/validator.go:14-19); mutant selects 120 by three page names (check/mutant/validator.go:7-19). No external dependencies or install scripts; go.mod declares Go 1.20.

### Differential
Ran `bash test/corpus/cases/PV-018/check.sh` on Linux: exit 0, complete output:

```
input=shipped base=pass change=pass mutant=pass reference=pass
input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)
input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass
suite change=pass(4/4)
suite mutant=pass(4/4)
```

The planted row is 160 bytes with a marker on an unlisted page (check/main.go:44-46): only change and reference refuse at 120. A plain 160-byte row on install (check/main.go:47-54) is refused only by the name-list mutant. Shipped rows pass all variants; the same four tests run against change and copied mutant (check.sh:25-40). The expected outcome (expected/EXPECTED.md:3-15) follows: current behavior is right, but the tests do not pin the content rule. Both rules need not select identical budgets for rows that pass.

### Unplug
False-world answer: a mutant suite failure would mean the tests detect this replacement; absent planted differential would not prove these rules differ. Neither happened. check.sh:15-22 rejects changed input lines; check.sh:42-61 requires both suite exit 0, exactly four passes, and zero failures. These are actual test invocations, not only printed constants. The reference repeats the contract independently of validator code (check/main.go:15-31), but is not independent evidence the contract is right. Round 2 added logs and exact test counts; it did not loosen a check.

### Closure
Visible repairs addressed: exact suite counts and diagnostics, isolated temporary directory, explicit reference limitation, and preserved manifest status object. Earlier review conditions are not enumerated verbatim in the supplied order; no claim is made about unnamed conditions. Verdict: accept-with-conditions. No owner gate or merge was performed.

### Remainder
OWED — CASE.md:85 records `bash tests/corpus/cases/PV-018/check.sh`; the repository now uses `test/corpus/`. The recorded command does not execute from the root. Correct the path and rerun it. CASE.md:82 dates the reported run October 23, while CASE.md:4 and manifest.json:325 say checks ran October 22; an earlier run might support this, but the reported October 23 run cannot alone verify October 22. Identify that earlier measurement or correct the status. No private host/seat name, internal path, private repository name, actual credential or email was found in this diff. Historical provenance and licenses are declined, not independently verified.

### Profile
Test-corpus and documentation. The shipped suite is circular: a name-list mutant passes it. The mutation differential probes that suite and the behavior on two boundary inputs. No baseline metric or exception is declared. Script uses bash, mktemp -d, cp -R, awk, diff -u, and Go; no visible Linux-only call. Linux execution does not prove macOS execution; requires local Go >=1.20. Network module access is disabled by check.sh:9.

### Focus
FILE·LINE·CONTROL: change/validator.go:15-19 chooses budget from row content. change/validator_test.go:15-40 tests cited rows only on listed pages and plain between-budget rows only on unlisted pages. check.sh:26-61 confirms those four tests pass on the name-list mutant while check/main.go:44-64 shows both behavioral differences. This is the intended test gap; it is not an argument to change the deliberately weak fixture tests.

### Conditions
1. Correct the command in CASE.md:85 to `bash test/corpus/cases/PV-018/check.sh` and run that documented command.
2. Measure on macOS with local Go >=1.20, or label macOS explicitly unverified.
3. Reconcile the October 22 status with the October 23 run timestamp or cite an earlier October 22 check.

### Unrun
macOS: no macOS environment available. Original round-one condition list: not provided verbatim; only observable repairs could be checked. Whole-repository Go suite: outside the scoped case review; case script ran both relevant suites.

## The review status as filed

## REVIEW — corpus-pv-018
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The check reproduces a passing name-list mutant and the planted-row differential; the expected test-gap finding follows. Fix the stale command and verify macOS.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T16:50:00Z

## REDACTED

none
