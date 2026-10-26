# REVIEW — corpus-pv-016 · round 4

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T16:30:28Z

## Summary

The checked-in runner fails before any claim is measured: it still points at tests/corpus, but the fixture lives at test/corpus. Direct Go runs confirm the intended no-delivery defect, not the claimed 24-line check run. Correct the path and rerun the full check and its negative control before filing a passing reproduction.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-016

### Ground

Scope: `9303b2d...3a84be4`; `git rev-parse --short=7 HEAD` printed `3a84be4`. The diff adds the fixture and changes only `test/corpus/cases/PV-016/` and its own `test/corpus/manifest.json` status object. It does not touch `expected/`. `git diff --check 9303b2d...3a84be4` found no whitespace errors. The three code files in base and change compare equal (`cmp -s` on each). Go reports `go1.26.5 linux/amd64`. These are fresh readings, not the author's claimed check run.

### Differential

BLOCKING — `test/corpus/cases/PV-016/check.sh:7` hard-codes `tests/corpus/cases/PV-016`, but the case is at `test/corpus/cases/PV-016`. `test/corpus/cases/PV-016/CASE.md:102` and `test/corpus/cases/PV-016/check.sh:2` repeat the obsolete command. The first `diff` in the code-equality guard reads neither variant, and the runner exits 1 before building or printing any claim. The reported 24-line output at `test/corpus/cases/PV-016/CASE.md:105` is not reproducible from this tree. Remedy: make the case path and documented command match the checked-in layout; rerun the actual script and update the run record only from that output. This leaves the round-one quote-presence condition unverified in the shipped layout; the round-one temporary-binary cleanup is present in the script (`mktemp`, EXIT trap), but is not exercised past its initial guard.

Exact output of `bash test/corpus/cases/PV-016/check.sh` (stderr, then observed exit):

```
diff: tests/corpus/cases/PV-016/base/go.mod: No such file or directory
diff: tests/corpus/cases/PV-016/change/go.mod: No such file or directory
check: base/go.mod and change/go.mod differ
EXIT=1
```

The command's false-world answer is also exit 1 at this first guard: a real difference between base and change would produce the same last line. Thus its exit and this comparison measure an absent path, not fixture equality or the delivery claim. Independent `cmp -s` on `go.mod`, `alerts.go`, and `main.go` found equality. Direct `go -C test/corpus/cases/PV-016/<version> run . emit <kind>` for all five named kinds in each version printed `considered=1 delivered=0 reason="no transport yet"` in every run. The independent direct change-side setting run printed `refused: destination id "t-backup-ops" must begin with c-`; `chat:c-backup-ops` printed `accepted: c-backup-ops`. Change-side field runs printed `error,host,job,time` for `backup_failed`, `expected_at,host,job,time` for `backup_late`, `host,job,time` for `restore_done`, `percent_free,time,volume` for `storage_low`, and `destination,reason,time` for `delivery_failed`. These direct runs support the defect but do not replace the promised joined page/claim check.

### Unplug

The new quote guard (`check.sh:31-40`) is before each operation and is intended to refuse a missing change-page quote. Its negative control is described at `CASE.md:134-136`, but cannot be reproduced through the shipped command because the path guard fails first. The check prints observed delivered counts but does not assert zero delivery: a hypothetical delivered=1 output, with every page quote present, would still exit 0. The printed count, not the exit status, is the instrument for the expected outcome; an evaluator must read it. This is consistent with the contract's requirement to print rather than adjudicate claims, but the current instrument never prints it.

### Closure

`test/corpus/cases/PV-016/expected/EXPECTED.md:3-8` requires evidence of zero delivery on all five listed alerts and refused documented setup. Independent direct runs support both on Linux; the original checked-in check does not establish either. An opposite outcome would print nonzero delivered counts or accept the documented setting; the independent direct runs would distinguish it, whereas the current check failure would not. `test/corpus/cases/PV-016/base/alerts.go:56-65` always supplies the null sender, and `:84-89` counts one considered and no delivery. This is a fixture/reproduction rejection, not a claim that the documented defect is false.

### Remainder

The page-only pairs in `reference.md:17-25` match the supplied pages: no actor versus actor routing, missing credential host/machine, failure alert to its failed destination, and `c` versus `t` destination ids. The scope-limited diff adds no private paths, real addresses, private repository names, or real credentials; the token-shaped example explicitly says synthetic. Licences and historical claims about the private source were not judged. No new dependency was added beyond the fixture's standard-library Go module.

### Profile

Docs/test-corpus review: the stated command does not run on this Linux checkout, so it cannot be claimed portable on Linux or macOS. `bash`, `mktemp -d`, `diff -q`, `grep -qF`, `rm`, and `rmdir` are available on both; Bash `${claimed//,/ }` needs Bash (the command explicitly invokes it); `go -C` requires the pinned Go toolchain. No GNU-only option was identified in the runner. Circularity: direct executable output is stronger than `CASE.md`'s recorded output; the page-quote check is a suite-owned check and can be made to pass without changing the underlying behavior.

### Focus

File · line · control: `check.sh:7` · the runner's source-directory choice prevents every contract observation; `CASE.md:99-105` · the recorded command/output diverges from the actual tree. Direct executable probes isolate the no-delivery code behavior from the broken runner. The manifest's `checks: run 2026-10-23` records a past run, not a fresh valid check on this layout.

### Conditions

Blocking: update the obsolete `tests/corpus` references in the case runner and documented command, rerun `bash test/corpus/cases/PV-016/check.sh` on the actual tree to exit 0 and print all 24 claim lines, and repeat the change-side quote-removal negative control to show that the quote guard—not the path guard—exits nonzero. Then reconcile §8 and the manifest with measured results. Round-one conditions: isolated binary cleanup is coded but not reached; quote-presence guard is coded but not reached. Neither is verified as paid by the shipped runner.

### Unrun

The complete 24-line check and its quote-removal negative control cannot pass the initial stale-path guard. macOS execution with Go 1.26 was not run (this seat is Linux). No independent end-to-end receiver test was run; the direct CLI observations and source inspection do not claim one.

## The review status as filed

## REVIEW — corpus-pv-016
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The checked-in runner fails before any claim is measured: it still points at tests/corpus, but the fixture lives at test/corpus. Direct Go runs confirm the intended no-delivery defect, not the claimed 24-line check run. Correct the path and rerun the full check and its negative control before filing a passing reproduction.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T16:22:26Z

## REDACTED

none
