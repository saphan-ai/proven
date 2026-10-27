# STATUS — corpus-pv-003 · round 31

- state: blocked
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: 6535ea93db57c827b897d183dbda97c0ad2ad890 [corpus-pv-003] tests: record round two verification limit
- updated: 2026-10-23T08:50:18Z

## Summary

gpt-6-sol; move case printed base=refused(X), change=allowed. Linux recheck passed; native macOS verification remains open.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-003
- state: blocked
- abstract: gpt-6-sol; move case printed base=refused(X), change=allowed. Linux recheck passed; native macOS verification remains open.
- last commit: 6535ea93db57c827b897d183dbda97c0ad2ad890 [corpus-pv-003] tests: record round two verification limit
- done: Rechecked all five inputs and recorded the exact Linux output in CASE.md §8.
  - `bash tests/corpus/cases/PV-003/check.sh` exited 0 and printed five rows, matching the review.
  - `go vet ./tests/corpus/cases/PV-003/...`, `gofmt -l` on the four case Go files, and `git diff --check` exited 0 with no output.
  - `jq` read PV-003 status as an object with contract, fixture, checks and independent_check; the order's string-status claim was wrong at tip 93ce778.
  - `git diff --stat fe146e0..HEAD` listed only the PV-003 case outside expected/ and the PV-003 manifest entry.
- next / blocked-on: Run the five-row check on native macOS with Go 1.26 before asserting cross-platform behavior.
- artifacts:
  - tests/corpus/cases/PV-003/CASE.md
  - SHOUT_macOS-verification.md
- evidence:
  - `uname -s` printed Linux; no macOS runner was provisioned, so the review's native macOS condition was not run.
  - `jq '[.cases[]|select(.status.fixture!="not built")]|length' tests/corpus/manifest.json` printed 21 and exited 0.
- note: ADDRESSED — Linux recheck and portable Bash/Go source inspection. NOT ADDRESSED — native macOS execution. No manifest change: the object shape already exists. No cross-platform guarantee is claimed.
- updated: 2026-10-23T08:45:09Z

## REDACTED

none
