# Answer — MASTER of `the-gateway-is-the-only-door-to-the-store`

Author: MASTER (seat code-fleet-2-1). Date: 2026-10-16. Source: this train's named reviews and master verifies of
2026-10-15/16 (W2b6a, W2b7d, W2b7e, W4c, W4d, W6, W6b). Everything below is from the record; where I did not measure,
I say so.

## 1. Critique of "Proof-carrying code review"

The direction is right — on this train, every disagreement that was settled was settled by a **run**, never by a
second opinion. But the proposal's definition of proof ("a test that fails on the PR head and passes on the base") is
the instrument that misled me most this week. Five corrections, each from a stream.

**1.1 What settled disagreements: a probe, a mutation, or a counted instrument — never a vote.**
- *W2b6a review r1, F1:* the leg claimed the gateway's order-issue judge was safe; the reviewer **ran a probe** and got
  "three far-accepted $14 admissions against a $30 budget with $0 spent". Nobody argued after that. The fix round had
  to re-run the same probe (C1) and show zero.
- *W2b7d round 1:* the order (mine) said ███ ████████████████ ██████ █ █████. The leg **refuted me with a test**
  (`█████████████████████████████████████████████████████`: ███ ██████ █████████ ██ ███████ ██████████████████). The
  disagreement was master vs leg, not model vs model, and a test decided it.
- *W6 review r1:* the leg's status said "all five acceptance items pass by named tests"; the reviewer **ran the touched
  package** and found two reds the diff introduced (`TestConfigSetHelpListsEverySettableKey`, byte order in
  `configKeyClasses`).
- The strongest single kind of evidence in our reviews is not "fails on head, passes on base" but the **unplug**: remove
  the guard, show the control go red, restore it (W4d r2: "the `Nlink` mutant survives today"; W6b: "reverting it to the
  base's mode-only check (M5) reddens"; W2b7d: the `open`-filter mutant survives → owed control). A finding of the form
  *"this guard has no test that notices its removal"* is PROVEN by a mutation, and the proposal has no slot for it.

**1.2 False alarms — and what exposed them.** Almost all of mine were **red tests that were not this change's**:
- W2b7d, my verify after an absorb: **10 "new" reds at the tip**. Two were real (a census row); **eight passed when
  re-run alone** — load- and order-dependent on a busy host. W2b6a: `TestContendedRefreshLockWaitsAndEveryDispatchProceeds`
  "new at the tip", then 3/3 green alone at tip and base. W4d: `TestTheTerminalOpensOverAValidSessionOnThisConsole`
  "fixed at the tip" — untouched code, a flake.
- W2b7c's census went red on three different branches the same day; the cause was **develop's own committed table being
  stale**, not any branch. A base-vs-head comparison blames the head.
- What exposed each one: **re-run in isolation (`-count=3`), run the same name on the base, and read the trunk alone.**
  ⇒ A proof-carrying reviewer needs a *flake gate* before PROVEN: the failing test fails N of N on head, passes N of N on
  base, and the base itself is measured, not assumed green. Otherwise it produces confident false alarms faster than the
  opinion-reviewers it replaces.

**1.3 What reviewers missed that a run or the history would have shown.**
- *The merge, not the head.* My worst defect this week (W2b7d review r1 REJECT) was invisible on the PR head **and** on the
  base: develop moved a field (`Detail`→`Reason`) in one of two row constructors, and only the **head merged with today's
  trunk** was broken (5/26 harness tests red). My whole-package compare missed it because the package panics on a sandboxed
  host before reaching the wagon's tests — the reviewer found it by **running the wagon's harness by name**. Same class:
  W2b6a vs develop's new `██████████████████████████████████████████` (a name-based census matched a field called `Proof:`);
  W4d's census vs a new unredacted log write that another track added to develop. ⇒ Prove on **head ⊕ current base**, and
  re-prove when the base moves.
- *The history: "has this already been built?"* W6 (ssh config keys) went through three rounds and a review before I saw
  that another track had **landed the same keys first** (e2e21d272); our branch conflicted in 30+ files and was superseded.
  The same afternoon my own fix for the sealed-lane RED was duplicated by the track that owned the code (368370c78). One
  `git log` of the trunk over the touched files would have shown both. Neither the leg, the reviewer nor I asked.
- *Findings no sandbox can prove.* W4c review r1's four blocking findings (F1–F4) were stated **"by reading, not on a
  host"** — they need root, a second OS user, launchd/systemd. They were true (the next round reproduced each with a seam
  test), but in the proposal they would be "QUESTION". W6b's "second uid" and W4d's launchd `StandardOutPath` stayed
  **UNMEASURED** on the whole fleet (no second user exists). ⇒ The label set needs **UNMEASURED-with-the-reason** as a first-
  class outcome, distinct from QUESTION, and a reviewer that says *which world it could not build*.

**1.4 What made a review useful instead of something I triage.** The reviews I acted on without re-reading had: (a) a
`Conditions` list split **blocking / owed**, (b) an **Unrun** section that the gate refuses to accept over (our law #53),
(c) FILE·LINE·CONTROL per item, (d) the reviewer's own probes kept out of the tree (two reviewers left `zz_*` probe files
behind that I had to delete before teardown — the sandbox must be disposable). The proposal's report format should copy
(a)–(c) and enforce (d).

**1.5 Two models from different families.** In this train the decisive second opinion was never a second *model*; it was a
**second instrument** (a probe, a mutation, the trunk). Keep the two families for recall, but sell the *instrument*, not the
committee.

## 2. My proposals

## Proposal 1: Why did the merge break? A semantic-conflict detective

- track: The Agent That Can Explain Why.
- the problem, in one sentence a judge understands: two branches are each green, git merges them without a conflict, and
  the result is red — nobody can tell which change on the *other* side broke *this* branch's assumption.
- who has this problem outside our fleet: every team with a busy main branch and long-lived PRs; merge queues that reject a
  PR for a failure it did not cause; monorepos where a rename in one package breaks a census/golden test in another.
- the data sources the agent investigates (at least two, name them): (1) the PR's diff, (2) the commits that landed on main
  since the PR's merge-base, (3) test runs on base, head, main, and head⊕main, (4) git blame/history of the failing
  lines, (5) the PR conversation.
- what "explained with evidence" means here — what does one answer look like? (paste a mock answer)
  ```
  RED on head⊕main, GREEN on head, GREEN on main
  Broken by   main@d635d7d "the decider writes the reason" (bisected over 23 main commits on the merged tree)
  Mechanism   main moved the note Detail→Reason in gateBatchApply; this PR's batchRowOf still writes Detail;
              ███ █████████ ███ ███ ██████ █████ █████████ █████ → every touched row refused
  Evidence    1. TestW2b7dARowLandsThroughTheDoor… FAILS 3/3 on head⊕main, PASSES 3/3 on head and on main
              2. with only batchRowOf:62 Detail→Reason on head⊕main: 26/26 PASS
  Suggested   build both rows through one constructor (no copy)
  ```
- the 3-minute demo, step by step: a public demo repo; PR is green; main gets an innocent rename; the merge goes red →
  one command → the agent bisects main's commits on the merged tree, names the one commit and the line pair, shows the
  one-line fix that turns it green; then a second case where the red is a **flake** and the agent says so (fails 1/5).
- what goes to GitHub as open source: a CLI that builds head⊕base trees in scratch worktrees, a bisector over the other
  side's commits, a flake gate (N-of-N re-runs, base measured), a report; adapters for Go and Python test runners.
- the bounded piece Saphan later takes, and why we would build it anyway: the **merged-tree verifier with the flake gate** —
  "is this branch still green against today's trunk, and if not, which trunk commit broke it". Every landing on this train
  this week needed it by hand (absorb → regenerate → harness by name → re-attest); we would build it anyway.
- sponsor technology we can use honestly: Nemotron for the mechanism sentence (the bisect is deterministic, the
  explanation is the model's); Cursor as an executor.
- what the fleet must build in 144 h: (1) scratch-tree builder for base/head/main/merge; (2) flake gate; (3) bisector over
  main's commits on the merged tree; (4) the explainer (diff-pair + blame → mechanism); (5) report + demo repo.
- the biggest risk, and how we would know by day 2 that it is failing: test-suite cost — bisecting reruns the suite
  log₂(n) times. Day 2 signal: on the demo repo the bisect names the right commit in under 2 minutes using only the
  failing test by name; if not, cut to "list candidate commits by touched-file overlap".
- anything that touches internal mechanisms or patents: our engine has a merge/union-verify step (`verify-union`) and a
  landing sequence; this proposal uses neither and needs no engine code — but the owner should confirm the idea itself is
  not covered by a filing.

## Proposal 2: Flake, pre-existing or mine? A red-test explainer

- track: The Agent That Can Explain Why.
- the problem, in one sentence a judge understands: a CI run is red and the author must find out whether the red is
  theirs, was already there, or is noise — today by re-running and reading logs by hand.
- who has this problem outside our fleet: every CI user; maintainers triaging "your PR broke the build" claims.
- the data sources the agent investigates: CI logs of this run and of recent main runs, a rerun of the named tests on base
  and head, git history of the failing test and of the code it covers, the environment of the runner (sandbox denials,
  disk, timeouts).
- what "explained with evidence" means here — what does one answer look like?
  ```
  14 FAIL on head · 13 of them also FAIL on base (pre-existing, listed) · 1 head-only:
  TestContendedRefreshLock…  FLAKE — 0/3 fail alone on head, 0/3 on base; failed under a parallel suite (load)
  verdict: this PR introduces no red
  ```
- the 3-minute demo, step by step: a PR with a red CI badge; one command; the agent splits the reds into mine / pre-
  existing / flake / environment ("listen tcp6 [::1]:0: bind: operation not permitted" = sandbox, not code) with the rerun
  counts; the badge story changes from "broken" to "one real failure, here".
- what goes to GitHub as open source: the comparator (FAIL names base vs head), the rerun-in-isolation step, the
  environment classifier, a GitHub Action.
- the bounded piece Saphan later takes, and why we would build it anyway: the **base-vs-tip FAIL-name comparator with
  reruns** — the master's verify on this train is exactly this, by hand, several times a day.
- sponsor technology we can use honestly: Nemotron for classifying log text; Cursor as an executor.
- what the fleet must build in 144 h: (1) test-name extraction for Go/Python/JS; (2) base/head runner; (3) rerun-in-
  isolation; (4) environment classifier; (5) report + Action + demo.
- the biggest risk, and how we would know by day 2 that it is failing: suites too slow to rerun; day-2 signal: the demo
  repo's classification done in under 3 minutes.
- anything that touches internal mechanisms or patents: none.

(Proposal 1 is the stronger fit for "explain why"; Proposal 2 is the safer build, and is also the flake gate Proposal 1 and
the advisor's proposal both need — it could be wagon 1 of either.)

## 3. The question I could not answer this month

*"Which landing made develop's `TestGatewayStartCensus` red?"* — 2026-10-16, measured on develop 4b50e8063: the census of
owner decision 8 ("nothing that starts a model or a shell runs as the gateway's user") fails with *"the gateway's
process-start set changed"*. It guards the property this train exists for, it is red on the trunk, and I found it only
because my W4d verify swept the gateway package. To answer it I would need the trunk's commits since the census last
passed, a rerun of that one test on each, and the diff of the start set — none of which the record holds together; the
census prints that the set changed, not who changed it or why. **Answered later the same evening, by hand:** a walk of
one test over 40 first-parent merges on a remote host named f0ec18d3a (`cursor-is-an-adapter`) as the first red, and the
census output at that commit showed two unclassified `git` starts reachable through the Cursor composer — a fixed program,
not a decision-8 violation. Finding the culprit took a scripted bisect and a full disk being cleared first; explaining it
("is this a real violation or bookkeeping?") took reading the call chain. That is Proposal 1's demo, done manually.
Second, smaller: *"has another track already built this?"* — W6 (`gw-ssh-config-keys`) was superseded by
`a-finished-far-leg-always-comes-home` (e2e21d272) after three rounds and a review; nothing on the way asked the trunk's
history.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| platform mechanisms | 5 | how our platform signs, verifies or licenses, which we have not published |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
