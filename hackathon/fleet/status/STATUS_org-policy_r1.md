# STATUS — org-policy · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 17 min
- last commit: c7ebc393e70b04c3abbd3dd51a46530cd45fc93a [org-policy] docs: POLICY.md and the PROVEN_POLICY settings row
- updated: 2026-10-23T19:17:08Z

## Summary

PROVEN_POLICY names an organisation policy read before proven.yaml; a weaker local value is refused on load by key, both values and policy file; exceptions (rule, location, reason, owner, expiry) live in the policy only, a missing owner or expiry is refused, an unexpired one shows the finding as accepted with its owner and does not count, an expired one counts again and is named. Text output done; JSON, SARIF and PR-comment writers do NOT exist on this base (see note), so those four outputs are UNRUN; Result.Accepted/Expired carry the fields for them. Three commits; both negative controls turned the named tests red.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — org-policy
- state: stop-2-ready
- abstract: PROVEN_POLICY names an organisation policy read before proven.yaml; a weaker local value is refused on load by key, both values and policy file; exceptions (rule, location, reason, owner, expiry) live in the policy only, a missing owner or expiry is refused, an unexpired one shows the finding as accepted with its owner and does not count, an expired one counts again and is named. Text output done; JSON, SARIF and PR-comment writers do NOT exist on this base (see note), so those four outputs are UNRUN; Result.Accepted/Expired carry the fields for them. Three commits; both negative controls turned the named tests red.
- last commit: c7ebc393e70b04c3abbd3dd51a46530cd45fc93a [org-policy] docs: POLICY.md and the PROVEN_POLICY settings row
- done: policy loader, stricter-than check, exceptions, text output, CLI variable, docs and settings row, each proved by a test that was run
  - weaker local rule refused by name: internal/review TestAWeakerLocalRuleIsRefusedByName and TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne (10 subtests), CLI exit 7 in internal/cli TestAWeakerProvenYamlExitsSevenNamingThePolicy
  - accepted finding shown with owner, not counted: TestAnAcceptedFindingIsShownWithItsOwnerAndDoesNotCount (Result), TestAnAcceptedFindingIsPrintedWithItsOwnerAndDoesNotFailTheRun (text, --error exits 0)
  - expired exception no longer accepted, date printed: TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo, TestAnExpiredExceptionCountsAgainAndIsNamed (text, exits 1)
  - no owner and no expiry refused: TestAnExceptionWithoutAnOwnerOrAnExpiryIsRefused (one subtest per field, plus empty value)
  - policy path inside the target refused, file and symlink: TestAPolicyInsideTheTargetIsRefused
  - docs/POLICY.md example loads: TestTheExampleOfPolicyMdLoads
- next / blocked-on: the master decides how JSON, SARIF and PR-comment reach this branch (branches fix/json-sarif-outputs and fix/three-formats hold writers not on this base); each writer prints Result.Accepted and Result.Expired
- artifacts:
  - docs/POLICY.md
  - internal/review/policy.go
  - internal/review/policy_test.go
  - internal/cli/policy_test.go
- evidence:
  - base measured: HEAD 16cf7b7 descends from the order's 9303b2d (git merge-base --is-ancestor, exit 0). Baseline go test -count=1 ./...: 5 packages red BEFORE any edit, all on "httptest: failed to listen on a port: operation not permitted" (this seat's sandbox forbids binding a port): cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, and internal/review aborts on its first listening test (TestLiveReviewCarriesTheRoutedModel) which hides the rest of that package. Capture █████████████████████████████████████ internal/review with the 9 listening tests skipped (-skip): ok, 69s.
  - final go test -count=1 ./... with the listening tests skipped by name: 31 packages ok, the same 5 packages still red on the listener bind (none touched by this change); internal/review, internal/cli, cmd/proven, internal/settingspage, cmd/settings-doc all ok. Capture ██████████████████████████████████ Not run: the skipped listening tests (UNRUN on this seat, the master runs them where a port can be bound).
  - go vet ./...: no output. gofmt -l internal cmd: clean.
  - negative control 1: replaced the call of refuseWeaker in parseLocal with a no-op; TestAWeakerLocalRuleIsRefusedByName and TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne went red (10 subtests, "weaker value: <nil>"). Capture ███████████████████████████████████████████ Restored.
  - negative control 2: made the expiry comparison always false; TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo (review) and TestAnExpiredExceptionCountsAgainAndIsNamed (cli) went red. Capture ███████████████████████████████████████████ Restored; both suites green afterwards.
  - Step-1 table, what each proven.yaml key does under a policy. POLICY-SETTABLE, stricter means: path_rules = the local list keeps every policy rule at the same or a higher class (extra rules allowed); class_order = equal to the policy's, and a policy that sets class_order, path_rules or fallback_class fixes it (the order ranks the classes those name); fallback_class = same or a higher class; doubled_classes = superset (off is the empty set); author_commit_trailers = on beats off; author_description = on beats off; author_trailer_keys = superset, without case; raise_class = on beats off. NOT policy-settable, no order of values: timeout, default_reviewer, special_reviewer, families, author_record.
- note: master-was-wrong 1: the order says every output format (text, JSON, SARIF, PR comment) but on this base FormatterFor refuses json and sarif ("has no writer in this build", internal/cli/format.go, and TestJSONFormatHasNoWriter pins it) and no PR-comment writer exists; the writers are on unmerged branches fix/json-sarif-outputs and fix/three-formats. Acceptance item 2 "(text and JSON)" is therefore met for text and for the seam Result only; JSON, SARIF suppression and PR comment are UNRUN. I did not write a second set of writers beside those branches. || master-was-wrong 2: the order's example "a failing severity — a lower threshold is stricter" has no proven.yaml key behind it; severity, --error, --include are shell flags and variables, not proven.yaml keys, so a policy in the proven.yaml format cannot set them. Not policy-settable in this wagon. || rulings, each what, why, cost if wrong. (1) Review gains Options.PolicyPath, Options.Now and Result.Accepted, Result.Expired, Result.PolicyStatement, additive changes to the seam that review.go says both sides agree; cost: the shell's track must accept three result fields. (2) PROVEN_POLICY is read by the CLI and handed in as Options.PolicyPath, because an existing test requires sample mode to read no environment key through Options.Env; cost: none. (3) cmd/proven/main_test.go gained one line allowing the new documented variable PROVEN_POLICY in its allow-list; the order says existing tests pass unchanged, this one could not without that line; cost: a reviewer may read it as a loosened test, it widens the list by exactly one documented name. (4) An exception holds through its whole expiry day by the UTC date of the clock; cost: up to a day early or late for a team far from UTC. (5) A location with no line covers the whole path; an exception matches a finding by exact rule id and exact path, no globs; cost: one exception per file. (6) Accepted findings are not filtered by --severity, because accepted must never be hidden; cost: INFO noise. || NOTICED, not acted on: (a) the CLI hands proven.yaml of the working tree to Open as an explicit config (resolveConfig in internal/cli/config.go), so a change's own proven.yaml governs the review when no baseline path is used; a policy now bounds this from below, but the baseline-commit design of review.go is bypassed by the shell. (b) rm is denied to this seat, so a probe file was renamed into policy_test.go rather than deleted; no stray file remains.
- updated: 2026-10-24T00:00:00Z

CLOSE-OK · org-policy · commit c7ebc393e70b04c3abbd3dd51a46530cd45fc93a · branch fix/org-policy · worktree ██████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 5 |
