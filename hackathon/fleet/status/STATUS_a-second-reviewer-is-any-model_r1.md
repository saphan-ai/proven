# STATUS — a-second-reviewer-is-any-model · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 29 min
- last commit: 3267b82 [a-second-reviewer-is-any-model] docs: second reviewer, lab, fields
- updated: 2026-10-24T03:13:00Z

## Summary

A second reviewer behind any OpenAI-compatible API (key, address, model) is built and wired as ADR-0009's special reviewer; with it every class is doubled, proven.yaml narrows, security always doubled.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-second-reviewer-is-any-model
- state: stop-2-ready
- abstract: A second reviewer behind any OpenAI-compatible API (key, address, model) is built and wired as ADR-0009's special reviewer; with it every class is doubled, proven.yaml narrows, security always doubled.
  Each finding carries both reviewers' verdicts and reasons plus a Disagree flag; the result names its reviewers and whether either is the author's model.
  Found and fixed an independence leak the order did not know: the second Skeptic was shown the first reviewer's wording of a shared risk.
  Build, vet and every test that can run here are green. Listener-binding tests could not run on this seat (UNRUN, listed in evidence).
- last commit: 3267b82 [a-second-reviewer-is-any-model] docs: second reviewer, lab, fields
- done: second reviewer, verdicts, reviewer names, docs and the negative control are committed in two commits (f463a5f code and tests, 3267b82 docs)
  - generator.NewSecond / LoadSecond (internal/generator/second.go, live.go): a wrong value is refused under its own name, empty key is not connected; proved by TestNoKeyMeansNoSecondReviewer, TestAWrongSecondReviewerValueIsRefusedByName, TestSecondSendsOnlyTheKeyAndThePrompt
  - the key is in redact.Names; proved by TestTheSecondReviewersKeyNeverAppearsInAnyOutput (key echoed by the second service; result JSON and trace file checked, redaction marker required present)
  - two reviewers that disagree give both verdicts on the result, three kinds of difference; proved by TestTwoReviewersThatDisagreeShowBothVerdicts; agreeing reviewers show none (TestReviewersThatAgreeShowNoDisagreement); a failed second reviewer is "no answer", not a disagreement (TestAFailedSecondReviewerIsNotADisagreement)
  - one reviewer alone is unchanged: TestOneReviewerAloneIsUnchanged; the existing TestDoubledWithoutSpecialModelRunsOnce and TestNonDoubledClassRunsOnce still pass
  - every class doubled when connected, narrowed by proven.yaml, security always: TestASecondReviewerWidensTheDoubledClasses and TestASecondReviewerReviewsAClassThatWasNotDoubledBefore
  - reviewers named with author-model answer: TestTheResultNamesBothReviewersAndTheAuthorsModel
  - independence: TestSecondReviewerNeverSeesTheFirstsAnswer
- next / blocked-on: none; the shell train renders the block from the fields in docs/ARCHITECTURE.md section "Two reviewers (ADR-0037)", and ADR-0041 (W37) is not started
- note: master-was-wrong: (1) the order says each reviewer sees only its own run, but at base the second Skeptic was built from the MERGED candidate whose title and id are the first reviewer's when both name a risk; with the fix removed TestSecondReviewerNeverSeesTheFirstsAnswer fails with "the second reviewer was shown FIRST-ID-31 from the first"; fixed in pipeline.go (specialOwn). (2) docs/LAB.md did not exist at the base, so it is created, not edited. (3) the order's base is 5a9f2fe but the branch stood on 8770336, a descendant, so review the diff 8770336..HEAD.
- artifacts:
  - internal/generator/second.go
  - internal/generator/second_test.go
  - internal/review/second_reviewer_test.go
  - docs/LAB.md
- evidence:
  - ruling: verdict vocabulary reported / not reported / refuted / no answer, where refuted means the reviewer's own Skeptic claimed a refutation — why: the order asks for verdict and reason and names no vocabulary — cost if wrong: the shell train renders other words, one constant block to change
  - ruling: a connected second reviewer overrides proven.yaml special_reviewer — why: ADR-0037 point 2 says special_reviewer points at it — cost if wrong: a user who set both loses the yaml model, documented in CONFIGURATION.md
  - ruling: NewSecond sits in live.go, not second.go — why: TestInternalAndCmdDoNotImportBannedPackages allows net/http only in live.go and an exemption would move that bar (C4) — cost if wrong: none, a move
  - ruling: a second reviewer that fails is no answer and not a disagreement — why: a failed call says nothing about the finding — cost if wrong: a real outage is not flagged in the block, but it is still in FailedServices under second-reviewer
  - ruling: the second reviewer is named by its configured model with FromReceipt false — why: an arbitrary API has no receipt (ADR-0019 is Nightshift's) — cost if wrong: none for the review, the trace says the name is not from a receipt
  - NEGATIVE CONTROL (law #27), mutated field: the line that gives the second Skeptic its own risk (own.Risk = r replaced by _ = r). `go test -count=1 -run '^TestSecondReviewerNeverSeesTheFirstsAnswer$' ./internal/review` exit 1, "--- FAIL ... the second reviewer was shown FIRST-ID-31 from the first". Restored; same command exit 0, "ok github.com/saphan-ai/proven/internal/review 0.732s".
  - gofmt -l internal cmd prints nothing; go build ./... and go vet ./... exit 0; go run ./cmd/settings-doc exit 0 and internal/settingspage ok (CONFIGURATION.md matches the registry)
  - FULL SUITE: `go test -count=1 ./...` cannot complete on this seat. The sandbox refuses to bind a listener (listen tcp6 [::1]:0: bind: operation not permitted) and httptest panics, which aborts a whole package; the same panic occurs in files this change does not touch (cmd/livecheck, internal/guardrail, internal/query, internal/generator live_test.go). I did not lift the sandbox. Each package was run with the tests that call httptest.NewServer skipped (script █████████████████████████████████████████████████████). Result: every package ok except the two below; internal/review ok 156s, internal/pipeline ok 26s, internal/generator ok, internal/settingspage ok, internal/cli ok, cmd/proven ok, demo ok.
  - STILL RED, not caused by this change (a listener bind inside a helper the skip pattern does not see): cmd/livecheck TestAllServicesReady and internal/guardrail TestSilentServerIsATimeout; both fail with bind: operation not permitted.
  - UNRUN (C3): every httptest.NewServer test in cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query and internal/review (the live_generator, secret_echo, missing_secret and zetaris families). They need a listener. The acceptance asks for httptest recorded-answer tests; the new tests use the real second-reviewer client over an in-process http.RoundTripper, which proves the same properties without a port. An Open-level test through env and a real listener is UNRUN here: the master should run `go test -count=1 ./...` on a seat that can bind.
  - Co-authored-by trailer: absent from both commits (git log -2 --format=%B)
  - after each git commit, packed-refs.lock: Operation not permitted was printed; git log -1 shows the commit landed (f463a5f, then 3267b82)
  - session recording (ADR-0024): none made; this leg is a one-shot headless run and no recorder was provided
  - commands and exit codes: gofmt 0 · go build 0 · go vet 0 · settings-doc 0 · runall.py: livecheck 1, guardrail 1, every other package 0 · negative control 1 then 0 · git commit 0 (lock warning) twice
- updated: 2026-10-24T03:30:00Z

CLOSE-OK · a-second-reviewer-is-any-model · commit 3267b8201035b747242e21bf7e9d166ee0d0b84c · branch fix/a-second-reviewer-is-any-model · worktree ██████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
