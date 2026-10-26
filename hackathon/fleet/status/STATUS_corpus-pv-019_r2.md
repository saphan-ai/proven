# STATUS — corpus-pv-019 · round 2

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: e3ffdafd336e45b79b98837b8ab8141d6989da48 test/corpus: name the corpus by its path after the ADR-0030 move (master)
- updated: 2026-10-24T02:58:31Z

## Summary

gpt-6-sol; preview-with-session-switch printed base fail and change pass. Reconciled PV-019's manifest and removed the stale CASE note.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-019
- state: stop-2-ready
- abstract: gpt-6-sol; preview-with-session-switch printed base fail and change pass. Reconciled PV-019's manifest and removed the stale CASE note.
- last commit: e3ffdafd336e45b79b98837b8ab8141d6989da48 test/corpus: name the corpus by its path after the ADR-0030 move (master)
- done: Paid the two review findings and reproduced the check output.
  - master, 2026-10-24: the branch was carried onto proven's rewritten main (no AI trailer, no seat author; case files identical) and the master replaced the retired tests/corpus path with test/corpus in the case (e3ffdaf); check.sh re-run on macm5 matches CASE.md §8; one codex review of this tip accepted (owner: one review where one settles it)
  - CASE.md §8 and the fresh check output matched byte for byte: 1,354 bytes; exit 0.
  - The PV-019 manifest status is an object with fixture date 2026-10-22 and checks date 2026-10-23, matching CASE.md.
  - CASE.md no longer names the missing commit or the internal order.
  - git diff --stat fe146e0..HEAD names only PV-019 files outside expected/ and tests/corpus/manifest.json.
- evidence:
  - bash tests/corpus/cases/PV-019/check.sh: exit 0, nine lines; full output is in CASE.md §8.
  - go -C tests/corpus/cases/PV-019 test -count=1 ./...: exit 0, one test package passed and three packages had no tests.
  - go -C tests/corpus/cases/PV-019 vet ./...: exit 0.
  - python3 -m json.tool tests/corpus/manifest.json: exit 0.
  - git diff --check: exit 0.
  - Linux run used Go go1.26.5 linux/amd64; macOS was not run in this round.
- next / blocked-on: Master independently runs the check at 5b65589 for review condition 1, then verifies this commit.
- artifacts:
- note: Master-was-wrong: round 1's instruction to replace the status object with a string would break the corpus status shape. The status object was retained and reconciled. check.sh already uses portable Bash and Go commands without GNU-only flags, so no script change was needed. The check's exit 0 proves execution; the printed lines carry the differential result.
- updated: 2026-10-23T01:43:36Z

## REDACTED

none
