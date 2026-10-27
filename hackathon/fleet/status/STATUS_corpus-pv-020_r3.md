# STATUS — corpus-pv-020 · round 3

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 5 min
- last commit: ec1c97f4b696f56369db26e18d5c6e77ab1eb56a test/corpus: name the corpus by its path after the ADR-0030 move (master)
- updated: 2026-10-24T03:02:23Z

## Summary

GPT-6 (exact routed model not exposed); base accepted and change refused the 2049-byte non-exempt input. Review conditions C1-C4 addressed in c247ee8; macOS remains unrun.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-020
- state: stop-2-ready
- abstract: GPT-6 (exact routed model not exposed); base accepted and change refused the 2049-byte non-exempt input. Review conditions C1-C4 addressed in c247ee8; macOS remains unrun.
- last commit: ec1c97f4b696f56369db26e18d5c6e77ab1eb56a test/corpus: name the corpus by its path after the ADR-0030 move (master)
- done: Paid the round-2 review conditions and committed the PV-020 case changes.
  - master, 2026-10-24: the branch was carried onto proven's rewritten main (no AI trailer, no seat author; case files identical) and the master replaced the retired tests/corpus path with test/corpus in the case (ec1c97f); check.sh re-run on macm5 matches CASE.md §8; one codex review of this tip accepted (owner: one review where one settles it)
  - C1: At unchanged 8c55cee, bash tests/corpus/cases/PV-020/check.sh printed the 16 CASE.md lines byte for byte, exit 0; rerun on c247ee8 also printed the same 16 lines, exit 0.
  - C2: git diff fe146e0..HEAD records PV-020 manifest status as an object with prior-art disclosure and checks run 2026-10-23, agreeing with CASE.md.
  - C3: go -C tests/corpus/cases/PV-020/change test -count=1 ./service first failed with HTTP 400 for both exempt routes above 8 MiB; after the change it passed. The handler now streams body counting without a second ceiling.
  - C4: CASE.md names Go and Bash preconditions; check.sh derives the source root without git and cleans each new temporary build directory.
- next / blocked-on: none
- artifacts:
- evidence:
  - go version: go1.26.5 linux/amd64; exit 0.
  - go -C tests/corpus/cases/PV-020/change test -count=1 ./...: exit 0; one tested package passed and two packages had no test files.
  - go -C tests/corpus/cases/PV-020/base test -count=1 ./...: exit 0; two packages had no test files.
  - go -C tests/corpus/cases/PV-020/change vet ./... and go -C tests/corpus/cases/PV-020/base vet ./...: both exit 0.
  - bash -n tests/corpus/cases/PV-020/check.sh: exit 0.
  - bash tests/corpus/cases/PV-020/check.sh: exit 0; full 16-line output in CASE.md section 8.
  - git diff --stat fe146e0..HEAD: 13 files, 801 insertions, 5 deletions; only PV-020 outside expected and the manifest changed.
  - git diff --name-only fe146e0..HEAD -- tests/corpus/cases/PV-020/expected: no output.
  - macOS: not run; no macOS host was available.
- note: Master-was-wrong: the round-1 instruction to use a string manifest status broke the object shape; restored the object. Ruling: count HTTP request bodies as streams to avoid memory growth; this keeps the exempt route acceptance promised by the contract. Earlier local scratch binaries remain under ██████████████████████████ because a cleanup command was refused; the new check cleans its own run directory. Exact model beyond GPT-6 was not exposed to this leg.
- updated: 2026-10-23T02:26:53Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
