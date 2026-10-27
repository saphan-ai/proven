# STATUS — corpus-pv-010 · round 35

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 1 min
- last commit: 6b101a9050e314501b04a584a4017a8d546b1fcc test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T06:56:02Z

## Summary

Attested carried tip 6b101a9 with the path repair. On the signed and other-script inputs, base and change both passed while the reference marked them stale. The check matched CASE.md §8 on Linux; no code changed. Seat: codex (exact model not reported to this leg).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-010
- state: stop-2-ready
- abstract: Attested carried tip 6b101a9 with the path repair. On the signed and other-script inputs, base and change both passed while the reference marked them stale. The check matched CASE.md §8 on Linux; no code changed. Seat: codex (exact model not reported to this leg).
- last commit: 6b101a9050e314501b04a584a4017a8d546b1fcc test/corpus: name the corpus by its path after the ADR-0030 move
- done: Master-carried tip and path repair attested; check rerun.
  - `git rev-parse --short=7 HEAD` printed `6b101a9` (exit 0); `git rev-parse HEAD` printed `6b101a9050e314501b04a584a4017a8d546b1fcc` (exit 0).
  - The master carried the branch onto rewritten main; the measured HEAD is the ordered carried tip.
  - The master fixed the retired `tests/corpus` path to `test/corpus`; `bash test/corpus/cases/PV-010/check.sh` ran from the repository root (exit 0).
  - The check output matches CASE.md §8's recorded Linux run byte for byte; no master claim was refuted.
- next / blocked-on: Master verifies and handles the gate; no executor code change or commit in round 3.
- artifacts:
- evidence:
  - `bash test/corpus/cases/PV-010/check.sh` exit 0; full stdout follows.
  - `toolchain: go version go1.26.5 linux/amd64`
  - `pinned count: 42`
  - `current | base=pass | change=pass | reference=current`
  - `stale | base=fail | change=fail | reference=stale`
  - `twenty-digits | base=pass | change=fail | reference=stale`
  - `minus-999 | base=pass | change=pass | reference=stale`
  - `plus-41 | base=pass | change=pass | reference=stale`
  - `arabic-indic-41 | base=pass | change=pass | reference=stale`
  - `description | stated widths: base=none | change=0..48 | change-edited=0..256 | reference=0..48 (49 widths) | change tests on change=pass | change tests on change-edited=pass`
- note: Full repository suite and a fresh macOS run were not run in this attestation round; the order requires the corpus check only. No declared quantitative testing bar was found in the reviewed case.

## REDACTED

none
