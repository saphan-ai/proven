# Answer — master of `a-finished-leg-is-returned-by-the-engine-not-by-the-master`

Author: the master of the track (seat code-fleet-2-1, session 528a7ba1). Written 2026-10-16.
Sources: this track's own named reviews and master verifications, 2026-10-15 → 16. Every number
below was measured on those streams. The ones I did not measure are marked **unmeasured**.

---

## Part 1 — critique of "Proof-carrying code review", from my own reviews

### What actually settled disagreements on my streams

My reviewers never disagreed with **each other**: the reviews on my track run one after another, round by
round. Disagreement happened **between the master's verification and the reviewer**. It happened three
times in two days, and each time the tests were green and the reviewer said reject:

| stream / round | master verify | named review | who was right |
|---|---|---|---|
| `finished-leg-return-w2`, round 4 (tip cd0567d66) | GREEN: `████████████` 97→99 FAIL names, both tip-only names pass alone at both SHAs | **reject**, 3 blocking + 2 owed: two concurrent gates can both start one return (read-then-append not atomic); `partial` is classified as owed and retried forever; teardown still prints a hand `saphan return` | the reviewer. Round 5 fixed all five. |
| `finished-leg-return-w27`, round 1 (tip 8d8d51e37) | GREEN: 96→97, the one tip-only name also fails 5/6 alone at base | **reject**: far branch delivery has no deadline while the claim is held; the capacity marker misses in-place row changes | the reviewer, confirmed by the leg's own step-0 reproduction in round 2 (the leg also refuted one narrow point) |
| `finished-leg-return-w27`, round 2 (tip 4a28bca97) | GREEN: 152→153, the 3 new names pass alone 3/3 at both SHAs | **reject**: far unpin unbounded; the marker cuts off live rows older than the newest 5,000; pack-deadline arithmetic overflows | the reviewer (round 3 in flight) |

**What settled each one was not a vote and not a second model. It was a reproduction ordered before
the fix.** Every continuation order on this track carries a step 0: *"reproduce each finding at the base
before you change anything; if you cannot, write that down in the refutation field."* The leg either
turns the finding red at the base, or it refutes it. Law #44 on our side makes the refutation a
deliverable. An agent can do this, and the proposal half-describes it. **The missing half is that the
reproduction must be allowed to come back "refuted", and a refuted finding is a result, not a failure
of the reviewer.**

⇒ **Critique 1.** The proposal frames disagreement as "model A vs model B". In my record the common and
expensive case is "**the test suite is green and the reviewer is right**". The demo should show that
case: a PR whose tests all pass, and a finding that is PROVEN anyway, with a test the agent wrote.
It is more convincing than two models arguing.

### Which findings were false alarms, and what exposed them

Every false alarm on my streams was a **failing test caused by the environment**, never a model's
opinion:

- W27 round-2 re-review (`finished-leg-return-w27-review-r3`, run-20261016-163120-9774991e). 15 tip-only
  `████████████` FAIL names, all of which failed again when re-run alone:
  - 14 were the reviewer seat's own confinement (`listen unix … socket: permission denied`);
  - 1 was a census test that walked a copy of the source the reviewer had placed under an ignored work
    directory.

  `repomirror` failed 14 names at **both** SHAs (`Invalid cross-device link`). The reviewer listed all of
  this honestly under its Unrun heading, but a naive "fails on head" runner would have reported 15 PROVEN
  defects.
- W27 round-1 verify. `TestReviewASecondDirectoryJoinsTheFirstsID` was tip-only, but it failed 5 of 6
  times alone at the **base**, so it is a flake and not a regression. The host's disk was near full at the
  time; whether that caused it is **unmeasured**.
- W2 round-4 verify. Two tip-only names (`TestBridgeInitRefusesAWorkspaceAlreadyBoundElsewhereAndSaysWhere`,
  `TestReviewJoinRefusesAnIdTheGatewayNeverRegistered`) passed alone at both SHAs, so they are load
  noise.

What exposed them was always the same three-step instrument:
1. run the **whole** package at base and at head, in the **same** environment;
2. diff the FAIL **name sets** (never counts, never a name list chosen in advance);
3. re-run every head-only name **alone at both SHAs**.

A finding survives only if it is red alone at head and green alone at base.

⇒ **Critique 2.** "A test that fails on head and passes on base" is necessary, but by itself it
**generates** false PROVENs. The evidence runner must:
- run base and head in one environment;
- re-run in isolation before it calls anything proven;
- classify the sandbox's own refusals (permissions, sockets, cross-device moves, full disk) as
  **UNMEASURED**, never as a defect.

Without that, the tool's headline feature, proof, produces the false alarms it claims to remove.

### What a reviewer missed that a test, a run or the history would have shown

- **A rule the reviewer could not have seen.** W2 round 3 added a repair sentence to two messages. Two
  existing tests that guard those messages went red (`TestASelectorOverAnEmptyRecordSaysTheRecordIsEmpty`,
  `TestAStaleVerdictPairingIsNotHandedOverAsDelivered`). The whole-package run caught it. A review of the
  diff alone would not have, because the tests sit in files the diff does not touch. Our memory records
  the same class on another stream: a list of test names chosen in advance missed two census reds (W14,
  2026-10-12). **The history data source has to include "which tests elsewhere pin this string".**
- **A weak test that looks like coverage.** The strongest finding of the W2 round-5 review
  (`finished-leg-return-w2-review-r3`, run-20261016-051255-e6d87530, verdict accept-with-conditions)
  was a **mutation result**. The reviewer disabled five guards by hand. Three tests went red. Two
  mutations survived:
  - M4: the fail-closed partial-record read had no cell;
  - M5: the lost-claim branch could be dropped while the loser printed a false sentence.

  Those became conditions C1 and C2. Round 6 added cells that go red under M4 and M5, and the master's
  verify confirmed that.

⇒ **Critique 3.** The proposal's **bait** ("looks wrong, is right, dropped because the test it wrote
passes") has the same hole. A passing test proves nothing until it has been shown able to fail. A
DROPPED needs a test that goes **red under a mutation of the suspected line**; otherwise it is a
QUESTION. Make **mutation survival** a first-class evidence type: *"guard X can be deleted and every
test still passes"*. On my track it found more real gaps than any other technique, and it is cheap to
demo.

### What made a reviewer's output useful to me instead of something I triage

1. **file:line + the remedy + the proof the fix needs, per finding.** The W2 round-5 order carried the
   reviewer's five findings **verbatim**, and the leg fixed all five in one round. Output I can paste
   into the next order is output I do not triage.
2. **An honest Unrun section.** Every review I accept says what it did **not** run. On my track a
   non-empty Unrun beside an accept is refused by law #53. On 2026-10-16 that law made the owner choose
   round 6 over a landing (W2), and the choice cost about 3 hours instead of a landing on unrun ground.
   ⇒ **The report schema needs `unrun[]` beside `findings[]`, and a clean verdict must be impossible
   while `unrun` is non-empty.** Few AI reviewers have this, and it is the most transferable idea here.
3. **The verdict arrives where I look.** Twice the review finished but its verdict did not come home on
   its own: the engine's return refused once on a store schema gap, and once the notes stayed on the far
   host. Unmeasured for the hackathon tool, but the same lesson applies: the report must be one artifact
   in one place, never "see the notes on the runner".

### Net position on the candidate

**I support it, with three changes:**
- **(a)** the headline case is "green tests, real defect";
- **(b)** evidence = differential run + isolated re-run + **mutation survival**, with the runner's own
  environment failures labelled UNMEASURED;
- **(c)** `unrun[]` in the schema, blocking a clean verdict.

These are what my record shows separating a useful review from noise.

---

## Part 2 — my proposal

## Proposal 1: Why did this test fail — regression, flake, or the box?

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** when CI goes red on a pull request, someone
  spends an hour finding out whether the change broke it, the test is flaky, or the machine is
  misconfigured — and the usual answer is "re-run and hope".
- **who has this problem outside our fleet:** every team with a large test suite on shared CI runners;
  maintainers who must decide whether a contributor's red build is the contributor's fault.
- **the data sources the agent investigates (at least two, name them):** (1) the test logs of the
  failing run; (2) the same tests at the base commit, re-run by the agent; (3) each failing test re-run
  **alone**, at base and head; (4) the test's own failure history across past CI runs; (5) the diff and
  blame — does the change touch the test or the code it covers; (6) the runner environment (disk,
  permissions, sockets, OS errors in the log).
- **what "explained with evidence" means here — one answer:**
  ```
  TestBridgeInit… · FLAKE (not caused by this PR)
  Evidence  1. fails in the full run on head; PASSES alone on head (3/3) and alone on base (3/3)
            2. failed in 4 of the last 30 CI runs on main, none touching this package
            3. the PR does not touch bridge_init.go or anything it imports
  Verdict   safe to merge on this test; open a flake ticket (draft attached)

  TestUnreadableStore… · ENVIRONMENT (unmeasured on this runner)
  Evidence  1. error text: "listen unix …: socket: permission denied" — the runner's sandbox, not the code
            2. same error at base → not introduced by this PR
  Verdict   cannot judge here; run it on a runner that allows unix sockets

  TestAStaleVerdictPairing… · REGRESSION (caused by this PR)
  Evidence  1. fails alone on head 3/3, passes alone on base 3/3
            2. the PR changed the message string this test pins (gate_engine_return.go, +4 −2)
  Verdict   blocking; the fix is in the message or in the test, not in a re-run
  ```
- **the 3-minute demo, step by step:** a public repo with a red PR and three failing tests (one planted
  regression, one flaky test, one test that needs a unix socket the runner denies). One command
  `whyred <pr>`. The agent re-runs base and head, isolates, reads history, and prints three different
  verdicts with evidence. Then the "merge on red?" line: 1 blocking, 2 not this PR's fault.
- **what goes to GitHub as open source:** the CLI; the FAIL-set differ (base vs head, by name); the
  isolate-and-rerun runner (Go and Python first); the environment-error classifier (a pattern table
  people can extend); the history reader (GitHub Actions run logs); the report.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **FAIL-set differ
  plus the isolate-and-classify step**. It is exactly what our masters do by hand for every master
  verification today (base vs tip FAIL names, re-run alone at both SHAs). We would build it anyway
  because every verify on my track repeated it by hand. It also answers critique 2 of the leading
  candidate, so the two can share it.
- **sponsor technology we can use honestly:** Nemotron for reading logs and writing the explanation
  (the verdicts themselves are computed, not generated); Cursor as an executor.
- **what the fleet must build in 144 h:**
  1. the differ and the base/head runner (Go, Python);
  2. the isolate-and-rerun loop with repeat counts;
  3. the environment classifier and its pattern table;
  4. the CI history reader (GitHub Actions);
  5. the report, the demo repo, and the README.
- **the biggest risk, and how we would know by day 2 that it is failing:** the history source (CI
  logs) varies across providers. By the end of day 2, the runner must classify the three planted tests
  correctly on base and head **without** the history. If it cannot, the history is cut to "optional"
  and the demo stands on re-runs alone.
- **anything that touches internal mechanisms or patents:** none. No order, policy, pinning or engine
  mechanism is used; it reads public test logs and runs tests.

It can also be a **component of the leading candidate** rather than a rival: the PR reviewer's evidence
runner needs exactly this step to avoid false PROVENs.

---

## The question I could not answer this month

On 2026-10-15, far legs across the whole fleet died at spawn with one stderr line,
`saphan: unknown command "saphan" for "saphan"`. Each run left a 0-byte transcript, and the engine then
refused its return because the run had written nothing.

I counted 27 such runs that day, from 06:54 to 14:43 local, on many tracks. Two of them were my own:
- `finished-leg-return-w2` round 5, run-20261015-214209-4bc0c610;
- `finished-leg-return-w27` round 2, run-20261015-214749-934a5c58.

At the same time other far runs started fine. I could measure **that** it happened and **when**. I could
not measure **which component** composed `saphan saphan …`. My notes to the scribe say so:
RED 72383fef22458351, "not measured: which component composes it".

Then it stopped, and nobody wrote down why. The next morning a different spawn failure took its place:
the egress wall could not cover `/run/user/1000`, five runs out of five on my streams. My first
explanation of that one, "Linger is off on the hosts", was wrong. My own next two runs refuted it, and
I withdrew it in 1fd66a39686f043b. That failure also went away after a fix I never saw named.

The evidence lived in four places: the run's stderr, the placement log, the host's state and an engine
install I did not watch. Assembling it took hours, and twice I reached a confident wrong answer.
**"Why did this run die, and is the cause still there?"** is the question I could not answer. It is the
same question as Proposal 1, asked about agent runs instead of tests.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| locations in our private code | 2 | file names and line numbers in Saphan Studio's private source code |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
