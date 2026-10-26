# REVIEW — org-policy · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 3 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T03:31:28Z

## Summary

Reject: repository path rules can lower organisation fallback; JSON acceptance is missing and required unplug tests were not run.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — org-policy

### Ground
RED — BLOCKING cross-field downgrade of organisation policy. Verdict: reject. Reviewed `16cf7b7...c7ebc39`; measured `git rev-parse --short=7 HEAD` = `c7ebc39`, and `git diff --stat 16cf7b7...c7ebc39` names 14 changed files. No product code edited.

BLOCKING · `internal/review/policy.go:285`, `internal/review/policy.go:265`, callers `internal/review/open.go:222` and `internal/review/decide.go:93`: policy `fallback_class: security` and local `path_rules: **=docs` pass `parseLocal` because `weaker("path_rules", ...)` checks only when the policy explicitly set that key and `weaker("fallback_class", ...)` checks only a changed fallback. A path that previously took the policy's security fallback now matches the repository's docs rule. Remedy: enforce effective-class monotonicity across the fallback and all local rule regions; test this exact cross-field case with a refusal naming the weaker rule. This conclusion is from code paths, not a successful CLI reproduction: the standalone sample run was refused by the existing state-directory guard before policy evaluation.

BLOCKING · `internal/cli/format.go:19`, `docs/POLICY.md` Exceptions: `--json` has no writer, so accepted findings cannot appear with reason, owner and expiry in the explicitly required JSON output. Remedy: implement and verify that output, or obtain an explicit change to acceptance. This is an order-mandated defect, not a speculative future format.

BLOCKING · `internal/review/policy_test.go:77` and `internal/review/policy_test.go:170`: no isolated weaker-check or expiry-comparison mutation and failing-test output are recorded. The named negative controls in the acceptance are missing. Remedy: unplug each guard separately, show a named test fails, and restore the guard; include cross-field fallback/path rule coverage.

OWED · `internal/review/policy_test.go:15`, `internal/review/policy.go:420`: the boundary test constructs UTC time and advances hours; no non-UTC location is tested. `now.UTC()` is correct by inspection, but the independent time-zone boundary lacks a test. Add two zone representations of one instant near local midnight and check the same UTC-date answer.

### Differential
Base `16cf7b7` has no organisation policy; tip adds parser, accepted/expired results, CLI text writer and tests. `git diff --check 16cf7b7...c7ebc39` returned 0; no module manifest changed. Fresh `go test ./internal/review ./internal/cli ./...` passed all reported packages at tip; `go vet ./internal/review ./internal/cli` returned 0. These instruments detect their own failures, not missing cross-field cases or mutant survival. Base suite not run. No bar-weakening moves in existing checks observed in diff; adding a test that expects JSON refusal establishes a conflicting new bar.

### Unplug
- Weaker-than guard: not mutated; no test turning red observed. False-world local wildcard overrides fallback without refusal; `TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne` varies one key at a time.
- Expiry guard: not mutated; no test turning red observed. `TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo` is a candidate, not mutation evidence.
- Owner/expiry: `TestAnExceptionWithoutAnOwnerOrAnExpiryIsRefused` removes one field at a time; validator not independently mutated.
- Policy confinement: `TestAPolicyInsideTheTargetIsRefused` exercises direct and linked paths; guard not independently mutated.
- JSON: `FormatterFor("json")` refuses, and `TestJSONFormatHasNoWriter` currently treats that refusal as success.

### Closure
REJECT. The cross-field floor and JSON acceptance are broken; the required negative-control measurements have not been run. The reviewer does not downgrade blocking conditions. Obtain a repaired change and independent review before any owner gate.

### Remainder
`TestNoPolicyChangesNothing` and the full tip suite cover the no-policy regression. Missing owner and expiry have field-specific refusal tests. The expiry is checked by UTC date string and expired findings remain in `Findings`; UTC day-of/day-after cases exist. In-target and linked policy paths have tests. Declined to judge future SARIF/comment writers, live service behavior, licences (outside order), and performance (no claimed metric). No new dependencies. Model cost and subscription-pool shares were not available as measurements and are not invented.

### Profile
Security/configuration and docs. The target controls local rules while the operator controls the policy file; the cross-field downgrade crosses that boundary. Direct and symlink containment is tested; policy reads are bounded. The diff adds no network, authorization, or destructive path. Overlapping patterns and alternate zones still need controls as above.

### Focus
Order conditions: 1 fails cross-field rule/fallback comparison; 2 fails JSON while text displays metadata; 3 UTC branch inspected, host-zone test absent; 4 owner and expiry refused by field tests; 5 policy file inside target and link refused by tests; 6 isolated unplug controls not performed. The CLI's `--error` reads `len(result.Findings)` after exception processing; `Accepted` remains separate. The pipeline processed the original findings before exceptions, but no JSON count consumer exists in this build.

### Conditions
Repair cross-field effective policy comparison and provide JSON output, then run both isolated negative-control mutations by name and add a timezone-boundary test. These are preconditions for a fresh verdict, not an accept-with-conditions recommendation.

### Unrun
Base-commit suite (review at immutable tip); isolated guard mutations and alternate-timezone runtime test (not implemented on this review leg); standalone CLI reproduction (attempted but state-directory confinement refused before review); JSON acceptance (writer absent); production integration and full security probes (outside local-only boundary). No claim that these ran is made.

Cleanup of reviewer-created scratch files under `█████████████████████████████████████` was denied by the command guard (`rm -f style commands are not permitted. Use a safer approach`). The denied act was not retried with another spelling; the scratch directory remains. The two manifest files above are the return artifacts; no source paths were changed.

## The review status as filed

## REVIEW — org-policy
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject: repository path rules can lower organisation fallback; JSON acceptance is missing and required unplug tests were not run.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T03:25:10Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
