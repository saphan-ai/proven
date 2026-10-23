# Answer — MASTER of `the-fleet-commits-as-itself-and-writes-to-the-customers-templates`

Author: MASTER (code-fleet-2-1). Date: 2026-10-16. Answered from the record, with no dispatch.
My evidence comes from 13 named reviews on one train. W1a had four reviews (2026-10-12 → 13) and W1b had two (2026-10-13 → 14).
W2 has had seven reviews so far (2026-10-15 → 16), and all seven rejected the round. Review notes are in
`███████████████████████████████fleet-git-identity-w{1a,1b,2}/*/_REVIEW_NOTES.md`. Orders and master-verifies are in
`██████████████████████████the-fleet-commits-as-itself-and-writes-to-the-customers-templates/`.

## 1. Critique of "Proof-carrying code review"

**The core idea is right, and our reviews already use it.** Every blocking finding on W2 that held up was a *run*, not an
opinion. The reviewer changed a scratch copy, ran the real door and showed the failing assertion. Examples:
- Review 5: "equal agent pair … exit 3, `git-identity-flag-refused`".
- Review 6: "`AllUnfinalizedRunning` then found **one** row".
- Review 7: "`cast refusal recorded as flag mismatch: rule=git-identity-flag-refused`".

None of these was disputed afterwards, and each became the next round's acceptance test word for word. So the proposal's
"PROVEN = fails on head" format is the right one. What it gets wrong or leaves out:

1. **"Fails on head, passes on base" is not enough. Without isolation it produces false PROVEN findings.**
   - Our base has **147 failing test names** (W2 master-verify round 1, base `a6241a21c`).
   - At tip `a150bc75d` (round 13), two names were new.
     - `TestTheFarWrapperForwardsHupAndTermToTheWholeGroup` fails 3/3 at base *and* tip in isolation, so it is a base red
       that a single base run happened to miss.
     - `TestABindWhoseSSLCertIsAStaleCopyIsRefusedByName` failed on a `TempDir RemoveAll … directory not empty` cleanup
       race and passed 5/5 at both commits.
   - Both would have been "PROVEN regressions" under the proposal's rule. They were exposed by **N isolated runs at both
     commits** plus the failure text. The evidence runner needs a flake/base-red classifier, or the demo's honesty claim
     breaks on the first real repository.

2. **The sandbox contaminates its own evidence.** W1a's fourth review (2026-10-13 09:01) found seven tip-only reds. Four of
   them were source-census tests that "diagnosed the TWO scratch source archives made under the tip worktree by this
   review". The reviewer's probes were the regression. Tests written by the agent must run in a **fresh checkout per commit**,
   never in the tree under review. W4a (2026-10-10) also showed a reviewer's leftover files blocking a teardown. We now
   order "leave the worktree as you found it" in every review.

3. **A test that reads only the return value proves too little. Check what was written.** W2 round 13 passed its new test:
   no running row, slot free. The seventh review still rejected it, because the *persisted* `refusal_rule` named the wrong
   cause. I had missed that, and so had the leg. The finding came from one extra assertion on the stored row. A good
   evidence step asserts **side effects** (rows, files, refs), not only exit codes.

4. **One review proves the newest defect but does not stop the next one.** W2 has gone 13 rounds, and **each fix created
   the next blocking defect in its own blast radius**:
   - r9 placed the refusal after the auto reservation, and the reservation leaked (review 4).
   - r11 moved it early but refused by flag *presence* (review 5).
   - r12 released one path but not the bind-refusal path (review 6).
   - r13 released every path under a fixed, wrong code (review 7).
   
   Each review was correct and each finding was proven, yet convergence was slow. What would have saved rounds is a
   **property sweep over the whole range**, not only the delta: "every refusal after `reserveAutoSeat`: does it release, and
   under which code?" In proposal terms, generate the test **matrix** around the defect (every error return of the
   function), not one test for the one input that failed.

5. **Some disagreements are policy, not fact, and no test settles them.** Twice the dispute was about what the rule *should*
   be:
   - Review 3 against my order: may `--git-user-*` flags override a default owner pair?
   - The leg's "correction of the order" in r11.
   
   Review 3 proved the behaviour, and the master (the owner's proxy) then ruled. Flags must equal the repository identity.
   An empty owner pair is filled by the flags, and a non-empty pair is never replaced. The r11 correction was settled
   the other way: review 5 wrote one test, and the leg's rule fell. The proposal has PROVEN / QUESTION / DROPPED. It needs
   a fourth label, **DECISION**: "the code does X, and the spec does not say whether X is wanted", routed to a named human
   with both readings. Otherwise those findings land in QUESTION and get ignored.

6. **What makes review output useful instead of something to triage:** four sections of ours that the proposal does not have.
   - **Conditions**, a list I can paste unchanged into the next order. I did that seven times on W2, and it is the most
     valuable part of a review.
   - **Unrun**, what the reviewer did *not* run. Review 5: "The `███████████████` package run was red … its base
     counterpart was not run." That one honest line kept me from chasing a false alarm.
   - A **negative control** for every claimed fix: "remove the release → this test fails with one running row".
   - A ruling with **cost if wrong / cost of accepting** (review 7).
   
   Put Conditions, Unrun and a per-finding unplug into the schema. The "Reviewers: A reported · B disputed" line
   matters less than these.

7. **Two model families was not what decided things here.** W2 reviews ran on a codex seat. Executors ran on codex
   (r1–r9) and then claude-code (r10–r14), after "Selected model is at capacity" twice. Every rejection was decided by a run,
   and none by a second model's opinion. I would demo "a second family" as the **disprover** that tries to break the first
   one's PROVEN findings. That is a vote between models, and a vote is weak.

## Proposal 1: Why is this test red? — triaging CI failures with evidence

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** a pull request's CI turns red, and nobody can say which failures
  this change caused, which were already red, and which are flaky. So people either rerun until green or block good changes.
- **who has this problem outside our fleet:** every team with a large test suite that is never fully green (monorepos,
  long-lived projects), and every maintainer judging agent-written PRs against a red CI.
- **the data sources the agent investigates:**
  - CI results for head and base, and the merge-base;
  - re-runs in a sandbox: N isolated runs per test, at both commits;
  - the failure output itself (assertion versus cleanup versus timeout);
  - git history and blame of the test and of the code it covers;
  - the PR diff, matched against the test's call graph or touched files;
  - the issue tracker and earlier CI runs of the same test on the default branch.
- **what "explained with evidence" means here:**
  ```
  TestABindWhoseSSLCertIsAStaleCopyIsRefusedByName · FLAKE (not caused by this PR)
  Evidence  1. failed once on head a150bc… — "TempDir RemoveAll cleanup: … directory not empty"
            2. isolated -count=5: PASS 5/5 on base a6241a…, PASS 5/5 on head
            3. failure is in test cleanup, not an assertion; a background warm-up writes into the temp dir
            4. diff touches none of the files the test exercises
  Next      not this PR's defect · file a flake issue with evidence 1+3

  TestRunDoorRefusesADifferingAgentPair… · CAUSED BY THIS PR
  Evidence  1. FAIL 3/3 on head, PASS 3/3 on base   2. assertion line + the PR hunk it reaches (run.go:366)
            3. the line was introduced by commit 9f3c… in this PR
  ```
  Each failure gets one of **CAUSED · ALREADY RED · FLAKE · ENVIRONMENT · UNKNOWN**, with the runs that decided it.
- **the 3-minute demo, step by step:**
  1. A public repo with a PR whose CI shows 6 red tests.
  2. One command.
  3. The report: 1 caused, with the PR line; 2 already red on the default branch since commit X; 2 flaky, with pass/fail
     counts and the cleanup race named; 1 environment (needs network).
  4. Click "caused" to see its evidence chain.
  5. Show the PR comment it posts: "1 failure is yours; 5 are not, here is why."
- **what goes to GitHub as open source:**
  - the CLI;
  - the base/head differential runner (N isolated runs per test, per commit, each in a fresh checkout);
  - the failure classifier (assertion / cleanup / timeout / environment);
  - a history investigator ("since when is this red on main?");
  - JUnit/`go test -json`/pytest parsers;
  - a markdown/JSON report and a GitHub Action.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **red-differential library**, with the
  schema of a classified failure and its evidence. Every master hand-runs this procedure on every round today:
  whole-package base vs tip, `comm -13` of FAIL names, `-count=N` isolated reruns at both commits. It is the master-verify
  "no red introduced" claim made mechanical. It also serves Proposal 1 of the advisor: the PR reviewer's evidence runner
  needs exactly this filter to keep false PROVEN findings out.
- **sponsor technology we can use honestly:** Cursor as an executor. Nemotron for the cheap classification of failure text
  across hundreds of tests (a local, open model is the honest fit for high-volume triage).
- **what the fleet must build in 144 h:**
  1. parsers + CLI + reading a PR's CI (GitHub Actions);
  2. the differential runner in a container (Go and Python);
  3. the failure classifier + history investigator;
  4. the report + PR comment + a demo repo with planted caused, flaky and already-red tests;
  5. README, a one-command install, the video.
- **the biggest risk, and how we would know by day 2 that it is failing:** reruns cost time on big suites. Run only the red
  tests, isolated. By the end of day 2, the demo repo's 6 reds must be classified correctly from real runs in under 5
  minutes; if not, cut to "differential from CI history only".
- **anything that touches internal mechanisms or patents:** none. It is a generic test-result differential; no order,
  policy or signing is involved.

It can also be **folded into the leading candidate** as its evidence-runner wagon (W3 in the advisor's list). That may be
the better use: the PR reviewer then stays honest on real repositories.

## The question I could not answer this month

*"Why does no fleet commit carry a seat identity?"* I measured it while preparing the cut-over: **0 seat-identity commits out
of 3043 bridge commits since 2026-10-08**. Every pre-commit hook printed only a warning, "the committing seat could not be
resolved … this workspace declares no id in its marker [`seat-identity-no-workspace`]". The answer was spread over four
places:
- that warning text, which says what is missing but not why;
- the workspace marker `██████████████████████`, which has no `id`;
- `origin.json` beside it, which *does* carry `workspace_id: saphan-studio`;
- the fact that `bridge bind --join`, the printed remedy, has no dry-run, so it could not be rehearsed before the owner ran it.

Assembling that took most of an hour, and it became a whole new wagon (W11, appended to the train order on 2026-10-16) that
now sits in front of the owner's cut-over signature. An agent that answered "why is this warning on every commit, and what
would happen if the next step were taken" from the hook's message, the files it reads and the history of the marker is the
tool I wanted that day.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| internal locations | 3 | where a document or log lives on our machines or in our private record |
| locations in our private code | 1 | file names and line numbers in Saphan Studio's private source code |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
