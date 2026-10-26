# STATUS — corpus-pv-004 · round 23

- state: blocked
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 6 min
- last commit: a222a08ce76603d72f722f661cd06616e7c3b998 [corpus-pv-004] tests: assert case output and restore manifest status
- updated: 2026-10-23T09:00:38Z

## Summary

Codex GPT-6 (exact model ID unreported); PV-004 base printed command: while change printed command:echo hello for indent-four. Linux checks pass; macOS verification remains unrun.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-004
- state: blocked
- abstract: Codex GPT-6 (exact model ID unreported); PV-004 base printed command: while change printed command:echo hello for indent-four. Linux checks pass; macOS verification remains unrun.
- last commit: a222a08ce76603d72f722f661cd06616e7c3b998 [corpus-pv-004] tests: assert case output and restore manifest status
- done: Restored the manifest status object and made the 11-line check output an assertion.
  - `jq` README count command exited 0 with 21 cases and 21 contracts.
  - Linux `bash tests/corpus/cases/PV-004/check.sh` exited 0; `cmp` against committed output exited 0.
  - A one-line expected-output mutation made the check exit 1; the restored check exited 0.
  - `git diff --stat fe146e0..HEAD` lists only PV-004 files outside `expected/` and `tests/corpus/manifest.json`.
- next / blocked-on: Run the check on macOS and settle the cold-install network requirement.
- artifacts:
  - SHOUT_macos-verification.md
- evidence:
  - Linux run at 2026-10-23T08:55:06Z; node v22.23.2; npm 10.9.8; shipped-suite ok 8; 11 output lines.
  - `bash -n tests/corpus/cases/PV-004/check.sh` exit 0; `git diff --check` exit 0.
  - macOS check unrun because this seat is Linux (`uname -s` printed `Linux`).
- note: C-1 addressed. Output assertion addressed. C-2 addressed on Linux only; macOS remains unverified. Cold npm installation needs registry access. The published count command returns with_code 21 because it counts statuses other than the literal `not built`; that count does not measure built cases.

## REDACTED

none
