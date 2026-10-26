# REVIEW — a-second-reviewer-is-any-model · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T03:50:39Z

## Summary

W36 @3267b82 reviewed on claude-code claude-opus-5-5: verdicts independent (M1 killed), key only to the base URL with no redirects, redacted in result and trace, one reviewer unchanged. Conditions: C1 a failed default Planner shows as a false disagreement (pipeline.go:552,1437); C2 no test checks the guardrail wraps the second reviewer (M3 survived); C3 LAB.md promises a disagree block the binary does not render. Listener tests UNRUN in the sandbox. The NVIDIA second review is owed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — a-second-reviewer-is-any-model

Reviewer: claude-code, `claude-opus-5-5`, run 2026-10-24, headless. Reviewed `8770336...3267b82` (f463a5f code and tests,
3267b82 docs). HEAD was checked first and was `3267b82`. Profiles: security (a third-party key, outbound calls) and
backend. Verdict: **accept-with-conditions**. The NVIDIA/Nightshift half of the owner's two-review rule is **owed**:
this review ran alone.

### Ground

What I measured on this seat (macOS sandbox, 2026-10-24):

- `go build ./...` exit 0. `go vet ./...` exit 0.
- `go test -count=1 ./...` exit 1. Every failure is the sandbox refusing a listener:
  `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. The panic stopped
  six packages part-way: `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`,
  `internal/query` and `internal/review`. All other packages printed `ok`.
- I re-ran `./internal/pipeline ./internal/review ./internal/generator` with `-skip` covering only the listener tests
  (the list is under Unrun): all three `ok`.
- I ran the 15 new tests by name with `-v`, and all 15 PASS: the 7 in `internal/generator/second_test.go` and the 8 in
  `internal/review/second_reviewer_test.go`.
- No dependency change: the diff touches no `go.mod` or `go.sum` (R4).

Answers to the order's five questions:

1. **Independence (holds for both reviewers' verdicts).** The second Planner gets only the fixture root:
   `internal/pipeline/pipeline.go:508`, `planner.Plan(ctx, fx.Root, special)`. The Skeptic prompt is only the ID and the
   title (`internal/agent/skeptic/skeptic.go:58`, `Prompt(c.ID, c.Title)`). For a risk both reviewers named, the
   second Skeptic now gets the second reviewer's own risk (`pipeline.go:532-540` keeps it, `pipeline.go:644-647`
   swaps it in). `cleared` (`pipeline.go:1492`) clears the first Skeptic's refutation. Neither the first reviewer's
   verdict nor its reason reaches any second-reviewer prompt. The independence leak the leg reported is fixed, and
   mutant M1 shows the fix has a test that catches it (see Unplug).
2. **The key.** It is sent only to `<PROVEN_SECOND_REVIEWER_BASE_URL>/chat/completions`, as `Authorization: Bearer`,
   without the Nightshift `meterless-*` headers (`internal/generator/live.go:262-271`, proven by
   `TestSecondSendsOnlyTheKeyAndThePrompt`). The client follows no redirects, because `NewSecond` passes a nil client
   to `NewLive`, which installs `CheckRedirect: ErrUseLastResponse` (`live.go:64-70`).
   - **Scheme:** `checkBaseFor` (`internal/generator/settings.go:235-253`) accepts `http` or `https` for **any
     host**, so the key can travel in cleartext to a remote address (owed item O1).
   - **Refusals:** a wrong value is refused by its variable name, including a key with a line break or over 4096
     characters (`second.go:62-64`). An empty key means "not connected" (`second.go:65`, `:93-95`).
   - **Redaction:** `redact.Names` now holds `EnvSecondKey` (`internal/redact/redact.go:26`). The pipeline redacts
     `Verdicts[].Model` and `Verdicts[].Reason` (`pipeline.go:290-297`). The trace recorder for the second run carries
     the redactor (`pipeline.go:468`). `TestTheSecondReviewersKeyNeverAppearsInAnyOutput` proves the result JSON and
     the trace are clean.
   - **SARIF and JSON:** neither exists yet (`internal/cli/format.go:16,26`: "JSON and SARIF ... refuse until they
     are built"). Nothing leaks there today. A future renderer reads only the already-redacted `Result`.
   - **livecheck:** `cmd/livecheck/main.go:78` builds `redact.FromEnv(env)` from the same `Names`, so it covers the
     new key automatically. livecheck does not call the second reviewer at all.
3. **One reviewer alone is unchanged.**
   - `withSecond` returns `cfg` unchanged when the model is empty (`internal/review/open.go:204-206`).
   - `Verdicts` is set only when `special != nil` (`pipeline.go:494`).
   - `LoadSecond` is read only in live mode (`open.go:101`).
   - Tests: `TestOneReviewerAloneIsUnchanged`, `TestNoKeyMeansNoSecondReviewer`, and the existing
     `TestNonDoubledClassRunsOnce`, `TestDoubledWithoutSpecialModelRunsOnce`, `TestDoubledSecurityRunsBothGenerators`
     and `TestSecurityIsDoubledAndBackendIsNot` (all PASS here).
4. **The guardrail (holds by reading, not guarded by any test).**
   - Every second-reviewer call goes through `p.guard(recording{inner: second ...}, rail)` (`pipeline.go:468`).
     `secondGenerator` (`pipeline.go:1383-1385`) is the only path to `p.special`.
   - With the guardrail key empty in live mode, `open.go:113-117` installs `guardrail.Absent`, and the same `rail`
     wraps the second run, so every call is refused (`guardrail/wrap.go:33-35` checks the prompt before
     `inner.Complete`).
   - **No test checks this for the second reviewer:** mutant M3, which removes the wrap, survives (condition C2).
5. **Suite:** see the first list above, and Unrun.

### Differential

Base `8770336` against tip `3267b82`:

- **Before:** a "special" reviewer existed only as the Nightshift endpoint with another model, only for doubled
  classes (security). A finding carried `Runs` and `Models` and nothing else.
- **After:**
  - Three environment variables connect any OpenAI-compatible endpoint as `p.special`.
  - Once connected, every class is doubled unless `proven.yaml` sets `doubled_classes`, and security is always
    doubled.
  - Every finding of a doubled run now carries `Verdicts` and `Disagree`. This includes the old ADR-0009 Nightshift
    doubled run.
  - `Result` gains `Reviewers` and `Disagreements`.
  - `checkBase` was split into `checkBaseFor(field, raw)`. The Nightshift behaviour is identical: same checks, same
    field name.
  - `Live.authorize` skips the routing headers when `plain`.
- **Behaviour the order did not ask for:** the old Nightshift doubled-security run now also emits `Verdicts`. This
  follows from ADR-0037 point 4 and is additive, not a regression.

### Unplug

All three mutations were made in the worktree, run, and reverted with `git checkout -- internal/pipeline/pipeline.go`.
`git status --porcelain` afterwards showed nothing of mine.

- **M1 · `pipeline.go:646`, `own.Risk = r` → `_ = r`** (the independence fix removed): **KILLED**.
  `TestSecondReviewerNeverSeesTheFirstsAnswer` FAIL: "the second reviewer was shown "FIRST-ID-31" from the first",
  with the Skeptic prompt quoted.
- **M2 · `pipeline.go:1441`, the `no answer` branch forced off:** **KILLED**.
  `TestAFailedSecondReviewerIsNotADisagreement` FAIL: `Verdict:not reported ... disagreements 1`.
- **M3 · `pipeline.go:468`, the second generator not wrapped by `p.guard`:** **SURVIVED**. `./internal/pipeline` and
  `./internal/review` are both `ok` with only listener tests skipped. No skipped test mentions `Special` or `DoubleRun`
  (grep of the five listener test files), so the survival is not an artifact of the sandbox. → Condition C2.
- **Master's control (not re-run by me):** forcing `f.Disagree` false in `internal/review/open.go` turns
  `TestTwoReviewersThatDisagreeShowBothVerdicts` red. The master reported this. By reading, that test asserts
  `!b1.Disagree` → fatal, so the control is credible, but it is the master's measurement, not mine.

### Closure

**F1 · `internal/pipeline/pipeline.go:476-488` with `:552` and `:1431-1445`.** The default reviewer's failure is shown
as a disagreement. Proof level: from reading, not reproduced. A reproduction needs a scratch test in the package, and
the order confines scratch to outside the worktree.

- **Trigger:** the default Planner's call fails as a service failure, so `st.soften` degrades it and
  `risks = nil` (`:476-488`), while the second Planner answers.
- **What happens:** `addAbsentVerdicts(req, cands, secondFailed)` (`:552`) only knows whether the *second* reviewer
  failed. Every candidate the second reviewer named gets a default verdict of `not reported` with the reason "this
  reviewer's own review did not name this risk" (`:1437-1440`). `disagree` then flags it, and `Disagreements` counts
  it.
- **Why the comment is wrong:** the comment at `:1429-1430` ("the default run is never the one that fails here,
  because a failed default Planner leaves no risks to be absent from") inverts the case. The default run's failure
  leaves the *second* run's risks with no default answer.
- **Reach:** the finding survives to the result whenever the default generator's later calls succeed, because the
  Investigator and Prover use it (`:566`, `:585`) to build evidence. Example: one timeout or 5xx on the default
  Planner call only.
- **Consequence:** the result says, falsely, that NVIDIA looked and did not name the risk. That breaks the stream's
  own rule, "a missing answer is not a different answer" (`candidate.go` VerdictNoAnswer doc, `docs/LAB.md` last
  section).
- **Remedy:** carry the default Planner's failure into `addAbsentVerdicts` the same way as `secondFailed`, and give
  that run `VerdictNoAnswer`. Add a test with a default generator whose Planner call returns `service.Down`, the
  second answering, and `Disagreements == 0`. → Condition C1.

**F2 · No test checks that the guardrail covers the second reviewer** (M3 survived). The property holds today by
reading `pipeline.go:468`. Nothing turns red if a refactor drops the wrap. That is the ADR-0005 rule with a third-party
endpoint behind it.

- **Remedy:** a test in `internal/review` using `secondAPI` (no port needed) with an `Absent` guardrail, asserting
  `len(api.prompts) == 0` and a guardrail refusal in the trace. Also one with a refusing sample guardrail. → Condition
  C2.

**F3 · `docs/LAB.md:21-27` and ADR-0037 point 4 promise a "The reviewers disagree" block, but the binary does not
render it.** The block is promised in text, Markdown, HTML, the PR comment, and JSON and SARIF. Today
`internal/cli/format.go` renders only `ReviewerStatements` (`:71`), and JSON and SARIF refuse. `docs/ARCHITECTURE.md`
correctly scopes the block to "the shell train", but LAB.md tells a user "What you see once it is connected" in the
present tense. → Condition C3.

### Remainder

Owed items. None of them blocks:

- **O1 · `settings.go:235-253`:** the second reviewer's key may be sent over plain `http` to any host. LAB.md's own
  example is `http://localhost:11434/v1`, which is legitimate. Suggested rule: refuse `http` unless the host is
  loopback, or print a named warning. The same applies to the pre-existing Nightshift base. The client also uses the
  default transport, so `HTTP_PROXY` and `HTTPS_PROXY` apply. Over `http`, the proxy sees the key. Owner's call.
- **O2 · `pipeline.go:566,585`:** for a risk only the second reviewer named, the *default model* runs the
  Investigator and Prover on the second reviewer's ID, title and path. Neither verdict is affected, because both are
  fixed before those steps and the default Skeptic skips special-only risks (`:614`). Still, "neither model sees the
  other's answer" is literally false at the evidence stage. The test docstring of
  `TestSecondReviewerNeverSeesTheFirstsAnswer` ("the first reviewer's prompts hold none of the second's") holds only
  because its one risk merges. The owner should rule whether ADR-0037 point 3 covers evidence gathering.
- **O3 · `pipeline.go:1509-1512` with `:1448`:** if the second run names two risks with the same `riskKey`, the merged
  candidate holds two `special` verdicts. The "default first" swap then only fires for `len == 2`. This is a cosmetic
  edge case.
- **O4 · `second.go:65`:** a whitespace-only key reads as "not connected" with no notice. Defensible, but undocumented.

### Profile

- **security:**
  - S1 trust boundaries: the second endpoint is a third party. The key is spoofable only through a user-set base URL.
    Redirects are blocked. Cleartext is possible over `http` (O1). The second endpoint's answers are untrusted text
    and get the same guardrail answer-side check and redaction as Nightshift's.
  - S3 control removal: the guardrail wrap was removed by M3 and no test failed (F2).
  - S4: no dependency delta. The new key is in the redactor list. I saw no invisible or bidirectional characters in
    the diff, from reading it; I did not run a byte scan.
- **backend:**
  - B4: a second-reviewer failure is renamed to the `second-reviewer` service (`second.go:113-120`) and lands in
    `FailedServices`. Proven by `TestSecondFailureNamesTheSecondReviewerAndNotTheKey` and
    `TestAFailedSecondReviewerIsNotADisagreement`.
  - Every call is bounded by `PROVEN_GENERATOR_TIMEOUT`.

### Focus

- Independence of the second Planner and Skeptic prompts.
- The key's destination, scheme, headers and redaction.
- One reviewer alone.
- The guardrail on the second run.
- How verdicts are built: `reportedBy`, `addAbsentVerdicts`, `noteRefuted`, `disagree`.
- `withSecond` doubling.

Declined to judge (R11):

- Docs prose style.
- `settingspage/catalog.go` placement.
- `reviewerOf` family logic beyond the tests: it mirrors the existing `reviewerLine`.
- Licences, which the order excludes.

### Conditions

- **C1 (F1):** a default reviewer whose Planner failed must show `no answer`, not `not reported`, on the second
  reviewer's risks, and must not count as a disagreement. Fix `addAbsentVerdicts` and its caller, and correct the
  comment at `pipeline.go:1426-1430`. Add the test named under Closure; its negative control is the default Planner's
  `service.Down`.
- **C2 (F2):** add a test that fails when the second reviewer's generator is not wrapped by the guardrail, for both an
  `Absent` guardrail (live, key empty: zero calls reach the second API) and a refusing one. Re-run M3 and show it
  killed.
- **C3 (F3):** make `docs/LAB.md` truthful for the binary that ships. Either say that the "The reviewers disagree"
  block and the JSON/SARIF verdicts arrive with the shell wagon, or land that wagon before this one is released to
  users.
- **Process:** the second independent review the owner requires (NVIDIA through Nightshift) is owed. This verdict is
  one of the two.

### Unrun

These tests ran nowhere on this seat. The sandbox refused listeners, verbatim:
`httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`.

- **`internal/generator`:** every `TestLive*` in `live_test.go`, plus `TestClosedPortIsUnreachable`.
- **`internal/pipeline`:**
  - `TestUnsafePromptStopsTheStepAndNamesCategories`, `TestUnsafeAnswerIsNotUsed`,
    `TestGuardrailOffCompletesAndSaysNoGuardrail`, `TestClosedGuardrailRefusesTheStepByName` and
    `TestSafeGuardrailKeepsTheSampleFindings`.
  - `TestLivePipelineWithoutAGuardrailIsRefused`, which is not a listener test but was caught by my `TestLive.*` skip.
- **`internal/review`:**
  - `TestRuntimeTokenEchoedInTheReceiptIsNotWritten`, `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten` and
    `TestOrgIDEchoedInTheReceiptIsNotWritten`.
  - `TestZetarisUnauthorizedReviewCompletes` and `TestZetarisClosedEndpointDegradesTheReview`.
  - `TestLiveReviewCarriesTheRoutedModel`, `TestLiveReviewRefusesANonNVIDIAModelByName`,
    `TestUnauthorizedReviewDegrades`, `TestClosedEndpointDegradesTheReview` and
    `TestLiveModeWithoutRuntimeNamesProverOff`.
  - `TestLiveSettingsAreRefusedOnOpen`, which is not a listener test but was skipped by the same pattern.
- **`internal/guardrail`, `internal/query`, `cmd/livecheck`:** these packages stopped at their first listener test,
  so the rest of each was not run here either.

For these, I rely on the master's measurement (2026-10-24, macm5, outside any sandbox: whole suite green, listener
tests included). That run is the master's, not mine.

Also not run by me:

- The master's `f.Disagree` negative control (cited under Unplug).
- A reproduction of F1, which is read from code only.
- The NVIDIA/Nightshift second review (route down; owed).

## The review status as filed

## REVIEW — a-second-reviewer-is-any-model
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: W36 @3267b82 reviewed on claude-code claude-opus-5-5: verdicts independent (M1 killed), key only to the base URL with no redirects, redacted in result and trace, one reviewer unchanged. Conditions: C1 a failed default Planner shows as a false disagreement (pipeline.go:552,1437); C2 no test checks the guardrail wraps the second reviewer (M3 survived); C3 LAB.md promises a disagree block the binary does not render. Listener tests UNRUN in the sandbox. The NVIDIA second review is owed.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T11:30:00Z

## REDACTED

none
