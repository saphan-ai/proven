# STATUS — corpus-pv-009 · round 27

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 0a2ae588bbe4a5e4f3cadc447bf3340ead10e1ee [corpus-pv-009] tests: record review condition evidence
- updated: 2026-10-23T09:20:00Z

## Summary

Codex GPT-6; on overflow-int64 base=pass and change=pass while reference=fail. Linux check exits 0; macOS remains unrun.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-009
- state: stop-2-ready
- abstract: Codex GPT-6; on overflow-int64 base=pass and change=pass while reference=fail. Linux check exits 0; macOS remains unrun.
- last commit: 0a2ae588bbe4a5e4f3cadc447bf3340ead10e1ee [corpus-pv-009] tests: record review condition evidence
- done: Recorded Linux wrapper output and the remaining review conditions in PV-009.
  - C1 Linux: `bash tests/corpus/cases/PV-009/check.sh` exited 0 and printed the nine lines in CASE.md §8.
  - C2: CASE.md §6 now points to the reconstructed evidence in §8.
  - C3: CASE.md §8 states the Go 1.26 requirement and possible `GOTOOLCHAIN=auto` download.
  - Scope: `git diff --name-only fe146e0..HEAD` lists only PV-009 files outside expected and tests/corpus/manifest.json.
- next / blocked-on: Run the literal check on macOS and compare its output with CASE.md §8; that host was unavailable in this leg.
- artifacts:
- evidence:
  - `go version` exited 0: go version go1.26.5 linux/amd64.
  - `bash tests/corpus/cases/PV-009/check.sh` exited 0 on Linux: nine lines, recorded in CASE.md §8.
  - `go test -count=1 ./tests/corpus/cases/PV-009/...` exited 0: change package passed; base and root have no test files.
  - `go vet ./tests/corpus/cases/PV-009/...` exited 0 with no output.
  - `git diff --check` exited 0 with no output.
- note: Master-was-wrong: PV-009 manifest status was already the required object at a0987ed, so no round-two manifest edit was needed. C1 macOS is NOT ADDRESSED; no macOS host was available, and its output is unverified. C2 and C3 are ADDRESSED. No declared numeric testing bar was found in this order.

## REDACTED

none
