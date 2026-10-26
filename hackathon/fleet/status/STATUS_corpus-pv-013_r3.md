# STATUS — corpus-pv-013 · round 3

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 6 min
- last commit: 92bcc0649e1deb396783b42957f37a0f685049d8 [corpus-pv-013] tests: verify change-side refusal leak
- updated: 2026-10-23T02:26:16Z

## Summary

Codex GPT-6; base accepted and showed the pasted synthetic value, while change refused it and printed that value in the refusal.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-013
- state: stop-2-ready
- abstract: Codex GPT-6; base accepted and showed the pasted synthetic value, while change refused it and printed that value in the refusal.
- last commit: 92bcc0649e1deb396783b42957f37a0f685049d8 [corpus-pv-013] tests: verify change-side refusal leak
- done: Repaired both review conditions and committed the case check.
  - `bash tests/corpus/cases/PV-013/check.sh` exited 0 and printed all seven contract inputs; CASE.md section 8 has its full output.
  - Both negative controls exited 1: base-only value and missing refusal; CASE.md section 8 records the output.
  - `jq -e '.cases[] | select(.id == "PV-013") | .status | type == "object"' tests/corpus/manifest.json` exited 0 and printed true.
  - `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...` exited 0 for the root module.
  - `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go vet ./...` exited 0 in each case module.
  - `bash -n tests/corpus/cases/PV-013/check.sh` and `git diff --check fe146e0..HEAD` exited 0.
- next / blocked-on: none
- artifacts:
  - tests/corpus/cases/PV-013/CASE.md
  - tests/corpus/cases/PV-013/check.sh
  - tests/corpus/manifest.json
- note: Master was wrong to require a string manifest status; the other cases and the corpus count use an object. PV-013 now has the original object shape with measured dates. Linux Go 1.26.5 was run; macOS execution is unrun because macOS was unavailable. No metric baseline was declared for this case.
- updated: 2026-10-23T02:21:53Z

## REDACTED

none
