# STATUS — poetry-locks-and-pypi-names-are-read · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 23 min
- last commit: cbda174fe9860db94a67ce10b934057a0424c988 [poetry-locks-and-pypi-names-are-read] manifest: poetry.lock and PEP 503
- updated: 2026-10-23T05:57:00Z

## Summary

A real poetry.lock with multi-line arrays is read, and PyPI names are compared and queried after PEP 503 normalisation.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — poetry-locks-and-pypi-names-are-read
- state: stop-2-ready
- abstract: A real poetry.lock with multi-line arrays is read, and PyPI names are compared and queried after PEP 503 normalisation.
- last commit: cbda174fe9860db94a67ce10b934057a0424c988 [poetry-locks-and-pypi-names-are-read] manifest: poetry.lock and PEP 503
- done: A real poetry.lock is read, and PyPI names match after PEP 503 normalisation.
  - TestPoetryRealisticLockFindsARaisedPackage finds numpy raised from 1.22.0 to 1.26.4. Negative control: the sub-table skip was removed, and that test failed with "base: line 18: not a package table". The skip was restored. go test -count=1 ./... then exited 0.
  - TestRequirementsPEP503NamesAreTheSamePackage and TestPoetryPEP503NamesAreTheSamePackage: Requests and requests are one package, and zope_interface and zope.interface are one package. The higher pin is a raise, not an add.
  - TestPyPILockPairsNamesAfterPEP503: poetry.lock wins for those two names, and onlyreq from requirements.txt is kept.
  - TestCheckPyPIQueriesTheNormalisedName: a pin written Requests is queried as package = 'requests' and matches the row.
  - go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./... exit 0. Every package printed ok or [no test files].
- next / blocked-on: none
- artifacts:
  - docs/OSV_TABLE.md
  - internal/manifest/gomod.go
  - internal/manifest/lock_test.go
  - internal/manifest/poetry.go
  - internal/manifest/poetry_test.go
  - internal/manifest/requirements.go
  - internal/manifest/requirements_test.go
  - internal/osv/check.go
  - internal/osv/check_test.go
  - internal/pep503/pep503.go
  - internal/pep503/pep503_test.go
- evidence:
  - Baseline, before any edit, on 2d3126bcef329c0b435468125da88c6285579176: go test -count=1 ./... exit 0. Every package ok or [no test files]. No package failed. Per-test pass, fail and skip counts were not printed. UNRUN at that grain.
  - Go is go1.26.8. go.mod says go 1.26 and has no toolchain line. The local toolchain matches.
  - RED before the fix. go test -count=1 -v ./internal/manifest/ ./internal/osv/ -run 'TestPoetryRealisticLockFindsARaisedPackage|TestPoetryWithoutAPackageTableIsRefused|TestPoetryPEP503NamesAreTheSamePackage|TestRequirementsPEP503NamesAreTheSamePackage|TestPyPILockPairsNamesAfterPEP503|TestCheckPyPIQueriesTheNormalisedName' exit 1.
  - TestPoetryRealisticLockFindsARaisedPackage failed: base: line 18: not a package table.
  - TestPoetryWithoutAPackageTableIsRefused failed: change: line 4: not a package table. It did not say "no [[package]] table".
  - TestPoetryPEP503NamesAreTheSamePackage failed with requests Added true and zope.interface Added true.
  - TestRequirementsPEP503NamesAreTheSamePackage failed with requests Added true and zope.interface Added true.
  - TestPyPILockPairsNamesAfterPEP503 failed: zope_interface stayed beside zope.interface. Requests had already paired by case fold.
  - TestCheckPyPIQueriesTheNormalisedName failed: the query was package = 'Requests' and returned no rows. The table row was requests.
  - After the fix, that same test command exited 0.
  - Negative control mutated the sub-table skip in parsePoetry, which is what keeps a line with no '=' inside another table off the old refusal. TestPoetryRealisticLockFindsARaisedPackage exit 1: base: line 18: not a package table. TestPoetryMalformed still passed. The skip was restored. go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./... exit 0.
  - go test -race UNRUN. This change adds no goroutine and no shared mutable state.
  - Testing bar: no metric is declared in the order. UNRUN. No bar was moved.
- note: master-was-wrong: none. Both defects failed on the base for the stated reason. The key is in this note because a status key outside the closed set is refused. ruling: the order said the two failing tests were below the defects and they were not pasted — the tests were written from the acceptance text and the defect's example lines — cost if wrong is a fixture shape the master had wanted to paste, and the red lines matched the claimed errors. ruling: a PyPI Change.Package is the PEP 503 form, so tests that expected Flask now expect flask — cost if wrong is a finding that shows flask where the file says Flask. ruling: docs/OSV_TABLE.md gained the normalisation sentence and the poetry paragraph now says lines in any other table are not pins — cost if wrong is one sentence the order did not ask to reword. Session recording (ADR-0024): none. This seat was not given a recording path, and none is published in this repository for this run. The program date is denied for this role (allowed form none); updated below is the commit time from git log -1 --format=%cI. A write of the commit message outside the worktree was refused (tool-before-hook-payload, no run id in the refusal). The message file was written under the work artifact directory and was not committed. No Co-authored-by trailer is on the commit.
- updated: 2026-10-23T05:49:12Z

CLOSE-OK · poetry-locks-and-pypi-names-are-read · commit cbda174fe9860db94a67ce10b934057a0424c988 · branch fix/poetry-locks-and-pypi-names-are-read · worktree ████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
