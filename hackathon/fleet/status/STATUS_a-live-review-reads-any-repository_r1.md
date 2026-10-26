# STATUS — a-live-review-reads-any-repository · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 30 min
- last commit: f483a318dbcaf59e3fabe4f444e23c5a4d4c31ba [a-live-review-reads-any-repository] pipeline: read a plain git repo
- updated: 2026-10-23T06:02:39Z

## Summary

Live mode reviews a git repository with no fixture.json. Sample mode still loads the fixture.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-live-review-reads-any-repository
- state: stop-2-ready
- abstract: Live mode reviews a git repository with no fixture.json. Sample mode still loads the fixture.
- last commit: f483a318dbcaf59e3fabe4f444e23c5a4d4c31ba [a-live-review-reads-any-repository] pipeline: read a plain git repo
- done: A live review reads a plain git repository
  - two commits reach the planner and the prompt contains note.txt: go test -count=1 -run TestLiveGitReviewReachesThePlanner ./internal/review exit 0
  - an empty HeadCommit shows the uncommitted edit: go test -count=1 -run TestLiveWorktreeReviewSeesAnUncommittedEdit ./internal/review exit 0
  - temporary trees exist during the planner call and are gone after: go test -count=1 -run TestLiveGitTreesAreRemovedAfterTheReview ./internal/review exit 0
  - sample mode without fixture.json still fails: go test -count=1 -run TestSampleModeWithoutFixtureJSONStillFails ./internal/review exit 0
  - negative control put fixture.Load back on the live path: TestLiveGitReviewReachesThePlanner exit 1, missing fixture.json; restored; the same test exit 0
  - go build ./... exit 0; go vet ./... exit 0; go test -count=1 ./... exit 0
- next / blocked-on: none
- artifacts:
  - internal/change/archive.go
  - internal/change/between.go
  - internal/change/read.go
  - internal/pipeline/trees.go
  - internal/pipeline/pipeline.go
  - internal/pipeline/imports_test.go
  - internal/review/open.go
  - internal/review/review.go
  - internal/review/git_live_test.go
  - docs/ARCHITECTURE.md
  - docs/INSTALL.md
- evidence:
  - reproduction before the change: open ██████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████ no such file or directory
  - negative control after putting fixture.Load back on the live path: open ██████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████ no such file or directory
  - proven binary was not executed: permission class refused it, allowed form none
  - sample review from TestScanSamplePrintsF1AndF2NotF3, the same run function as proven scan --sample testdata/fixture/discount, exit 0
  - review class: security
  - review rule: go.mod
  - author model not stated; default reviewer NVIDIA
  - reviewer NVIDIA; role default; family NVIDIA; author model not stated
  - special-review model: not configured
  - finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
  - finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list
  - finding osv:evil.example/leftpad ERROR Dependency evil.example/leftpad v1.2.3 is in an OSV affected range
  - finding F1 ERROR Discount is applied twice
  - evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  - failed services: none
  - F3 was not printed
- note: master-was-wrong: no. The status key set has no master-was-wrong field, so it is recorded on this line. No session recording was visible to this leg; whether a proxy recorded the session is unverified (ADR-0024). Race detector UNRUN: the acceptance command has no -race and this change starts no goroutine. No testing-bar metric is declared. remark: the sample review above is the test's stdout, not a run of the proven binary, because that program is outside the executor allow list. ruling: the live snapshot is the literal live for a directory that already has base/ and change/, and baseline..head or baseline..worktree for git — a fixture snapshot is sample-only — cost if wrong is evidence cited under a name a reader did not expect. ruling: prepared base/ and change/ trees are used as they are, even inside a git repository, so the demo is not archived — cost if wrong is a demo review that also passes a baseline still mounting those trees while the file list comes from git. ruling: only tracked files are copied — cost if wrong is an untracked file that Between lists when HeadCommit is empty is absent from the planner prompt. ruling: the stub generator, query and runner are unexported fields on review.Options — cost if wrong is the shell cannot set them; a review leaves them unset.
- updated: 2026-10-23T05:56:52Z

CLOSE-OK · a-live-review-reads-any-repository · commit f483a318dbcaf59e3fabe4f444e23c5a4d4c31ba · branch fix/a-live-review-reads-any-repository · worktree ██████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
