# STATUS — org-policy · round 6

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 7 min
- last commit: f56ff13 [org-policy] docs: POLICY.md, the effective-class rule and its example
- updated: 2026-10-24T10:21:15Z

## Summary

Round 2. A repository can no longer lower the organisation policy across fields: when the policy sets path_rules, class_order or fallback_class, the local rules and fallback are compared with the policy's effective classification and a weaker rule is refused by name.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — org-policy
- state: stop-2-ready
- abstract: Round 2. A repository can no longer lower the organisation policy across fields: when the policy sets path_rules, class_order or fallback_class, the local rules and fallback are compared with the policy's effective classification and a weaker rule is refused by name.
  The reviewer's case (policy fallback_class: security, local path_rules: **=docs) and three more cross-field cases are tests; all four are red on c7ebc39 and green now.
  Both negative controls were run in isolation and went red; the expiry day is tested in two non-UTC zones on both sides of midnight.
  Whole module: go test -count=1 ./... green (36 ok, 5 no test files, 0 fail); go vet ./... clean. JSON writer not built (out of scope, owner's ruling pending).
- last commit: f56ff13 [org-policy] docs: POLICY.md, the effective-class rule and its example
- done: findings of round 1 addressed except the JSON writer
  - cross-field rule (BLOCKING, ADDRESSED): policy.go fixesClasses and weakerRules; instrument go test -run TestNoPathLoosensTheClassAcrossFields, red on c7ebc39 (4 of 4 cases) in ██████████████████████████████████████████, green on d180cc6
  - negative controls (BLOCKING, ADDRESSED): outputs below, files in ████████████████████████
  - time zone (OWED, ADDRESSED): TestTheExpiryDayIsTheUTCDateInEveryZone, Asia/Bangkok and America/New_York at 23:59 and 00:00 UTC
  - JSON (BLOCKING, NOT ADDRESSED): out of scope by the order; FormatterFor("json") still refuses
  - docs: POLICY.md section The effective class, with the reviewer's example
- next: independent review of the stream, then the owner's gate
- evidence:
  - THE CROSS-FIELD RULE. When the policy sets path_rules, class_order or fallback_class, no path may get a lower class than under the policy. A path takes the highest class of the rules that match, else the fallback. The repository keeps every rule of the policy (or of the defaults) at the same class or higher; a repository rule that is not the policy's must sit at or above the policy's fallback; the repository's fallback must sit at or above the policy's. A rule below the fallback is refused even if it might only match paths a policy rule covers, because pattern overlap cannot be decided; that is conservative and is written in docs/POLICY.md. A policy that sets none of the three keys leaves rules to the repository, as before. Refusal names the key, the rule, both classes and the policy file.
  - NEGATIVE CONTROL 1 (weaker-than check removed alone: refuseWeaker returned nil at its first line; restored after). Command go test -count=1 -run 'TestAWeaker|TestEachPolicyKey|TestNoPathLoosens|TestALocalValue|TestARepositoryInherits|TestTheExampleOf|TestAnExceptionInProvenYaml' ./internal/review. Failing tests, FAIL ok line is `FAIL github.com/saphan-ai/proven/internal/review 1.316s`: TestAWeakerLocalRuleIsRefusedByName; TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne (all 9 subtests: path_rules_dropped, path_rules_lowered, fallback_class_lowered, doubled_classes_dropped, author_commit_trailers_off, author_description_off, raise_class_off, author_trailer_keys_dropped, class_order_changed, class_order_under_path_rules); TestNoPathLoosensTheClassAcrossFields (4 subtests); TestTheExampleOfPolicyMdLoads. Full text in ████████████████████████████████████████████████████████████
  - NEGATIVE CONTROL 2 (expiry comparison removed alone: `if e.Expires < today` became `if false && e.Expires < today`; restored after). Command go test -count=1 -run 'TestAnExpired|TestTheExpiryDay|TestAnAccepted|TestAnExceptionCovers' ./internal/review. Output:
    --- FAIL: TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo (0.64s)  policy_test.go:240: findings 3 accepted 1
    --- FAIL: TestTheExpiryDayIsTheUTCDateInEveryZone/Bangkok_07:00_on_1_Jan,_UTC_00:00_on_1_Jan  now 2027-01-01 07:00:00 +0700 Asia/Bangkok (UTC 2027-01-01): accepted 1 findings 0, want accepted false
    --- FAIL: TestTheExpiryDayIsTheUTCDateInEveryZone/New_York_19:00_on_31_Dec,_UTC_00:00_on_1_Jan  now 2026-12-31 19:00:00 -0500 America/New_York (UTC 2027-01-01): accepted 1 findings 0, want accepted false
    FAIL github.com/saphan-ai/proven/internal/review 3.970s. Full text in █████████████████████████████████████████████████████████████████
  - After both restores: grep -c "NEGATIVE CONTROL" internal/review/policy.go printed 0; go vet ./... clean; go test -count=1 ./... green.
  - WHOLE-MODULE COUNTS: go test -count=1 ./... 36 packages ok, 5 no test files, 0 FAIL (output ██████████████████████████████████████). Run after d180cc6; f56ff13 changes docs/POLICY.md only, and TestTheExampleOfPolicyMdLoads was re-run green after it.
  - REFUTATION OF THE ORDER (law 44): the order says rebase with git rebase --onto main 16cf7b7. There is no local ref named main in this worktree, and the branch already had 16cf7b7 as an ancestor (git merge-base --is-ancestor 16cf7b7 HEAD printed ancestor), so the rebase was a no-op and was not run. origin/main is now c1e4ba8, ahead of 16cf7b7; the branch was not moved to it.
  - One slip of mine, reported: a denied cp left my first copy of policy.go unsaved while a following git show overwrote the file with c7ebc39's version. This was the planned red-first step; I re-applied the same edits by hand and re-ran the tests, so d180cc6 is what the green runs measured.
  - NOT RUN: base-commit suite (no baseline run before changes; the tip c7ebc39 suite was green per the reviewer, and red-first was run on c7ebc39's policy.go for the new tests only); no standalone CLI reproduction; model and cost shares not recorded here (not available to this leg).
  - Judgement calls: (1) the trigger is any of path_rules, class_order, fallback_class set by the policy, so a policy that sets only doubled_classes still lets a repository lower the fallback (TestALocalValueOfAKeyThePolicyDidNotSetIsFree unchanged). Cost if wrong: a policy author who wants the default classes protected must set a class key. (2) A policy that sets only fallback_class now also pins the default rules, so a repository that drops a default rule is refused. Cost if wrong: a repository relying on that is refused with a named rule.
- artifacts:
  - ██████████████████████████████████████████
  - ███████████████████████████████████████████████████████████
  - ████████████████████████████████████████████████████████████████
  - ██████████████████████████████████████
- updated: 2026-10-24T11:20:00Z

CLOSE-OK · org-policy · commit f56ff13 · branch fix/org-policy · worktree ██████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 10 |
