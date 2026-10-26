# STATUS — org-policy · round 13

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 12 min
- last commit: 1199ee2 [org-policy] docs: POLICY.md, the doubled-path rule and the inside check
- updated: 2026-10-24T15:36:41Z

## Summary

Round 3 on main (origin/main, rebased from 16cf7b7; five commits kept, three conflicts resolved by hand, no merge). B1: a doubled path stays doubled under every pairing of policy and repository, second reviewer included (swept). B2: a policy file anywhere in the git working tree that holds the target is refused. B3 and O1: the two default reads and the no-file case are tested. Whole module green: 47 packages ok, 0 failed, 10 without tests; vet clean.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — org-policy
- state: stop-2-ready
- abstract: Round 3 on main (origin/main, rebased from 16cf7b7; five commits kept, three conflicts resolved by hand, no merge). B1: a doubled path stays doubled under every pairing of policy and repository, second reviewer included (swept). B2: a policy file anywhere in the git working tree that holds the target is refused. B3 and O1: the two default reads and the no-file case are tested. Whole module green: 47 packages ok, 0 failed, 10 without tests; vet clean.
- last commit: 1199ee2 [org-policy] docs: POLICY.md, the doubled-path rule and the inside check
- done: B1, B2, B3, O1 fixed and guarded; two negative controls run and restored
  - B1 (doubled path): `TestDoubledPathsStayDoubled` (the reviewer's probe, both cases), `TestADoubledPathIsRefusedByName`, `TestAStricterClassStaysFreeWhenItIsDoubled`, and the sweep `TestNoCombinationOfSettingsUndoublesAPath` (policy x repository grammar, every path of a list, with and without a second reviewer; counts accepted and refused so it cannot pass vacuously). Red before the fix, green after (run in this leg before touching policy.go).
  - B1 rule: `doubled_classes` in the policy now fixes the classification as a floor; a rule or fallback the repository adds above a class the policy doubles must land in a class the repository doubles (`leavesDoubled`, `internal/review/policy.go`). Committed test and docs line that allowed the opposite were changed (`TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne`, `TestALocalValueOfAKeyThePolicyDidNotSetIsFree`, docs/POLICY.md).
  - B1 finding of my own, from the sweep: under a second reviewer security is doubled even with `doubled_classes: off`, so a policy that names `doubled_classes` (including `off`) now fixes the class floor. A policy that names none stays silent and the repository is free (ADR-0037 choice kept).
  - B2 (inside check): `change.TopLevel` plus `refuseInside` compares the policy file with the target and with the top level of its git working tree, links resolved. `TestAPolicyInTheRepositoryOfASubdirectoryTargetIsRefused` (probe, plus a link and an outside file) and `TestAPolicyInTheRepositoryIsRefusedByAReview` (through Review).
  - B3: `TestNoLocalFileRunsOnThePolicy` and `TestTheDefaultReadsRefuseAWeakerProvenYaml` (working tree and baseline), the reviewer's P3 and P4; they pass.
  - O1: `TestARepositoryInheritsThePolicyValueItDoesNotSet` now asserts the inherited class and rule of a review, not only the statement.
  - Negative control 1 (doubled-path check removed): with `leavesDoubled` forced false, `TestDoubledPathsStayDoubled/a_raised_fallback_leaves_the_doubled_class`, two cases of `TestADoubledPathIsRefusedByName` and `TestNoCombinationOfSettingsUndoublesAPath` fail; with `doubled_classes` removed from `fixesClasses`, `TestDoubledPathsStayDoubled/no_class_key_in_the_policy`, two cases of `TestADoubledPathIsRefusedByName` and the sweep fail. Captures in `█████████████████████████████████████` and `control1b.txt`. Restored with git checkout.
  - Negative control 2 (repository-wide check removed): with the top-level branch of `refuseInside` disabled, `TestAPolicyInTheRepositoryOfASubdirectoryTargetIsRefused` and `TestAPolicyInTheRepositoryIsRefusedByAReview` fail. Capture `control2.txt`. Restored.
  - Whole module: `go test -count=1 ./...` 47 ok, 0 FAIL, 10 no test files (`████████████████████████████████`); `go vet ./...` clean.
- next / blocked-on: none
- note: Refutation of the order: none needed on B1 to B3, the probes were red where stated (the two B1 cases and the B2 probe failed before the fix; P3 and P4 passed). Judgement calls: (1) a repository that raises a class above a doubled one must double the raised class, even where no doubled path could move there, because a pattern cannot be shown disjoint (cost if wrong: an extra refusal, named with the line to change). (2) The sweep covers a small grammar and seven paths, not every path. (3) Conflicts in `open.go`, `review.go`, `format.go` and `catalog.go` were resolved by keeping both sides. Not run: live reviews, a real second reviewer, JSON writers (out of scope). Model and cost: Sonnet 5.5, recorded by the engine, not measured here.
- artifacts:
  - internal/review/policy.go
  - internal/review/policy_doubled_test.go
  - internal/review/policy_test.go
  - internal/change/between.go
  - docs/POLICY.md
- updated: 2026-10-24T09:00:00Z

CLOSE-OK · org-policy · commit 1199ee2 · branch fix/org-policy · worktree ██████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
