# STATUS — corpus-pv-013 · round 4

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 6932236ac542abe11703dedebf1e7a8ed184e84a test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T07:08:36Z

## Summary

At 6932236, PV-013 check exited 0 and matched CASE.md section 8; the pasted synthetic token appears in both base show output and the change refusal. Seat model was not reported to this leg.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-013
- state: stop-2-ready
- abstract: At 6932236, PV-013 check exited 0 and matched CASE.md section 8; the pasted synthetic token appears in both base show output and the change refusal. Seat model was not reported to this leg.
- last commit: 6932236ac542abe11703dedebf1e7a8ed184e84a test/corpus: name the corpus by its path after the ADR-0030 move
- done: Attested the master's carry and path fix, then reproduced the seven documented case lines.
  - git rev-parse --short=7 HEAD printed 6932236; git rev-parse HEAD printed 6932236ac542abe11703dedebf1e7a8ed184e84a.
  - git diff 4c1d2cb..6932236 showed four path replacements from tests/corpus to test/corpus in CASE.md and check.sh, with no other content changes.
  - bash test/corpus/cases/PV-013/check.sh exited 0; all seven lines matched CASE.md section 8.
- next / blocked-on: Master to resolve the review's OWED construction-date discrepancy before claiming compliance with the 22–27 October authorship rule.
- artifacts:
  - _EXEC_STATUS.md
- evidence:
  - 'Command: bash test/corpus/cases/PV-013/check.sh; exit code: 0.'
  - 'control | base: none | change: pass | base_token_in_output: no | change_token_in_output: no'
  - 'valid-name | base: accepted | change: accepted | base_token_in_output: no | change_token_in_output: no'
  - 'token-in-credentials-show | base: (empty) | change: (empty) | base_token_in_output: no | change_token_in_output: no'
  - 'bad-token-env | base: accepted | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: chat-token | base_token_in_output: no | change_token_in_output: no'
  - 'bad-channel | base: accepted | change: refused: channel must be a channel name matching #[a-z0-9][a-z0-9_-]*; rejected value: not a channel | base_token_in_output: no | change_token_in_output: no'
  - 'bad-endpoint | base: accepted | change: refused: endpoint must be an http or https URL with a host; rejected value: not-a-url | base_token_in_output: no | change_token_in_output: no'
  - 'token-in-name-then-show | base: accepted / token_env=tok_synthetic_0000_not_a_secret | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: tok_synthetic_0000_not_a_secret / (empty) | base_token_in_output: yes | change_token_in_output: yes'
- note: Review verdict is accept-with-conditions and records an OWED construction-date discrepancy at CASE.md line 5 and the manifest; the order forbids code or document changes in this round, so it remains for the master. No commit was made.
- updated: 2026-10-24T07:05:53Z

## REDACTED

none
