# STATUS — corpus-pv-018 · round 4

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 00391174ee86b43da7ef6f7ef0a4e92ea96b705d test/corpus: name the corpus by its path after the ADR-0030 move (master)
- updated: 2026-10-24T02:55:04Z

## Summary

Codex model not reported to this leg; base=pass and change=fail for the planted citing row, while the mutant=pass. The five-line check and counted suites pass; commit 8d10cc2 awaits master verification.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-018
- state: stop-2-ready
- abstract: Codex model not reported to this leg; base=pass and change=fail for the planted citing row, while the mutant=pass. The five-line check and counted suites pass; commit 8d10cc2 awaits master verification.
- last commit: 00391174ee86b43da7ef6f7ef0a4e92ea96b705d test/corpus: name the corpus by its path after the ADR-0030 move (master)
- done: Paid both reviews' actionable PV-018 conditions and recorded the complete check output.
  - master, 2026-10-24: the branch was carried onto proven's rewritten main (no AI trailer, no seat author; case files identical) and the master replaced the retired tests/corpus path with test/corpus in the case (0039117); check.sh re-run on macm5 matches CASE.md §8; one codex review of this tip accepted (owner: one review where one settles it)
  - Baseline `bash tests/corpus/cases/PV-018/check.sh` on 10dfbe9: exit 0, five lines matched the prior CASE.md §8.
  - Final `bash tests/corpus/cases/PV-018/check.sh`: exit 0, three input lines and suites pass(4/4) each; CASE.md §8 carries the full output.
  - `go -C tests/corpus/cases/PV-018 test -count=1 -v ./...`: exit 0, four tests passed and zero skipped; `go -C tests/corpus/cases/PV-018 vet ./...`: exit 0.
  - Negative control: remove the test file from a temporary copy, run its check: exit 1 and both suites report fail(0/4) with no-test diagnostics.
  - `git diff --check fe146e0..HEAD`: exit 0; write-set is only PV-018 outside expected/ and the existing manifest line from round 1.
- next / blocked-on: Master verification and gate.
- artifacts:
  - tests/corpus/cases/PV-018/CASE.md
  - tests/corpus/cases/PV-018/check.sh
  - tests/corpus/cases/PV-018/data/data.go
  - tests/corpus/cases/PV-018/go.mod
- note: Master-was-wrong: the manifest status at 10dfbe9 was already an object with the four required fields, so no manifest repair was made. The reference repeats the stated rule and is not independent evidence, now stated in CASE.md. The scratch cleanup command was refused by the command guard; the negative-control copy remains under ███████████████████████████████████ and is untracked.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
