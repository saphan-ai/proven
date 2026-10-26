# STATUS — corpus-pv-012 · round 2

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: c2ae9053c677c3a7e0f87c0a710cb5f5513349d7 [corpus-pv-012] tests: pay review conditions on case check
- updated: 2026-10-23T01:39:59Z

## Summary

Codex (GPT-6; exact route unreported): application served five foreign events in base and five in change; named-team control served five in base and zero in change. Review conditions C1–C4 addressed; master must settle the literal request-team wording in §2 before a new verdict.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-012
- state: stop-2-ready
- abstract: Codex (GPT-6; exact route unreported): application served five foreign events in base and five in change; named-team control served five in base and zero in change. Review conditions C1–C4 addressed; master must settle the literal request-team wording in §2 before a new verdict.
- last commit: c2ae9053c677c3a7e0f87c0a710cb5f5513349d7 [corpus-pv-012] tests: pay review conditions on case check
- done: Reproduced the seven-line baseline and added an ownership negative control.
  - `bash tests/corpus/cases/PV-012/check.sh` at 657ce27 printed the seven lines recorded in CASE.md §8; exit 0.
  - The final check printed eight lines, including `control-wrong-ownership` with base 5, change 4, control FAIL, reference 0; exit 0.
  - `go vet` on each version, `gofmt -l`, and `git diff --check` exited 0; the formatter printed no paths.
  - `git diff --stat fe146e0..HEAD` lists only PV-012 case files and the manifest.
- next / blocked-on: Master must resolve §2's literal request-team wording before seeking a new review verdict; §8 records this fixture's constructor-team interpretation.
- artifacts:
  - tests/corpus/cases/PV-012/CASE.md
  - tests/corpus/cases/PV-012/base/main.go
  - tests/corpus/cases/PV-012/change/main.go
  - tests/corpus/cases/PV-012/check.sh
  - tests/corpus/manifest.json
- evidence:
  - C1 ADDRESSED: baseline seven-line run matched the review table and the prior §8 record.
  - C2 ADDRESSED: CASE.md §5 states the target and withheld driver inputs.
  - C3 ADDRESSED: command failures now abort the check, and temporary binaries are removed.
  - C4 ADDRESSED: CASE.md §8 states constructor-team semantics and the local Go minimum.
  - Codex blocking finding ADDRESSED by an ownership negative control that makes the change control FAIL; its §2 wording ruling remains for the master.
  - Not run: whole repository suite, macOS execution, and a forced binary crash; this round's acceptance calls for the case check, and this seat is Linux.
- note: Master-was-wrong: the manifest status was already an object at 657ce27, so this round did not change it. The reviewer requested a master ruling on the §2 phrase "requesting team's events when the request carries a team"; the round order permits only §5 and §8 changes in CASE.md, so §2 remains unchanged. No expected/ files were read.
- updated: 2026-10-23T01:36:44Z

## REDACTED

none
