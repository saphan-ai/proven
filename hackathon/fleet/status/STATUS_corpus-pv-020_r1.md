# STATUS — corpus-pv-020 · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 16 min
- last commit: 0b2082cd30ce226e5c972981318ba83838935155 [corpus-pv-020] tests: build the size-limit case and record the check
- updated: 2026-10-21T17:39:48Z

## Summary

Seat model Grok 4.7. Oversize body: base accepted all six routes; change refused four and accepted the two exempt uploads. Absent description and a refused listener could not run; in-process failed with the limit removed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-020
- state: stop-2-ready
- abstract: Seat model Grok 4.7. Oversize body: base accepted all six routes; change refused four and accepted the two exempt uploads. Absent description and a refused listener could not run; in-process failed with the limit removed.
- last commit: 0b2082cd30ce226e5c972981318ba83838935155 [corpus-pv-020] tests: build the size-limit case and record the check
- done: PV-020 case code is built and check.sh output is recorded in CASE.md §8
  - bash tests/corpus/cases/PV-020/check.sh from the repository root, exit 0, full output in CASE.md §8
  - git diff --stat main...HEAD is 12 files, all under tests/corpus/cases/PV-020/ outside expected/ plus tests/corpus/manifest.json
  - go vet ./... exit 0 in base and in change
  - go -C tests/corpus/cases/PV-020/change run -race ./cmd/controls -group listener printed passed, exit 0
  - go -C tests/corpus/cases/PV-020/change run -race ./cmd/controls -group in-process printed passed, exit 0
- next / blocked-on: none
- artifacts:
  - tests/corpus/cases/PV-020/CASE.md
  - tests/corpus/cases/PV-020/check.sh
  - tests/corpus/cases/PV-020/description/api.json
  - tests/corpus/cases/PV-020/base/go.mod
  - tests/corpus/cases/PV-020/base/service/service.go
  - tests/corpus/cases/PV-020/base/██████████████████
  - tests/corpus/cases/PV-020/change/go.mod
  - tests/corpus/cases/PV-020/change/service/service.go
  - tests/corpus/cases/PV-020/change/service/controls.go
  - tests/corpus/cases/PV-020/change/██████████████████
  - tests/corpus/cases/PV-020/change/████████████████████
  - tests/corpus/manifest.json
- evidence:
  - check.sh exit 0 at 2026-10-21T17:33:57Z
  - go version go1.26.8 linux/amd64
  - go vet base exit 0
  - go vet change exit 0
  - race listener passed exit 0
  - race in-process passed exit 0
  - testing-bar undeclared, UNRUN, no bar weakened
  - repository has no go.mod and no CI-gated Go suite, so a project-suite baseline is UNRUN
- note: ruling — manifest fixture set to built and checks left not run, because the order changes this case's status only, to built. CASE.md records checks run (2026-10-21). The date program is denied on this seat; the clock was read with python3. The session did not print a separate model field; the seat model is Grok 4.7.
- updated: 2026-10-21T17:36:53Z

CLOSE-OK · corpus-pv-020 · commit 0b2082cd30ce226e5c972981318ba83838935155 · branch fix/corpus-pv-020 · worktree █████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 3 |
