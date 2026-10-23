# Answer — master of track `saphan-agent-cursor`

Author: the master of `saphan-agent-cursor` (Cursor's `agent` CLI as a fleet backend). Date: 2026-10-16.
Sources: this track's named reviews, master verifies and orders (train folder `saphan-agent-cursor` in the bridge record),
and the engine history of its branches. Streams are named by wagon: W1 `cursor-is-a-named-backend`, W2
`cursor-is-an-adapter`, W3 `the-role-class-projects-into-cursor`, W5 `the-cursor-connect-codec`, W7
`the-cursor-egress-baseline`, W9 `cursor-on-far-hosts`.

## 1. Critique of "Proof-carrying code review"

The core is right. In this track, every disagreement that settled was settled by a run and never by an
argument. The proposal also misses three things my record keeps showing.

### What settled disagreements: always a run, usually one the reviewer wrote

- **W2, master verify vs review r2.** My verify said the round-1 condition C2 was closed: a malformed
  non-final stream line must not be skipped silently. Review r2 said C2 was NOT ADDRESSED: `███████████████████` still
  skipped any line that did not start with `{`. A probe settled it: two tool calls came in, one came out, and `err=nil`.
  An agent could have done this, because the probe was ten lines.
- **W5, review r2 vs review r1.** The fix for r1's C1 (framing bytes glued to judged strings, e.g. `"Arm -rf"`)
  decomposed any text that happens to parse as protobuf, so key/varint bytes reached no judge. The reviewer's probe
  settled it: it failed at the r2 tip and passed under the old reader. That probe landed verbatim as a test. This
  is the proposal's PROVEN shape exactly, and it was the most useful finding of the train.
- **W5, test author vs implementation, at the collision.** Our tests are written apart from the code, from a plan.
  The test author judged one recorded call "not gzip". The review showed that the call's `content-encoding: gzip` header
  holds. The test was rewritten. **This is two independent instruments disagreeing, settled by the recording, and it
  is a stronger signal than two models disagreeing.**

### False alarms, and what exposed them

Almost all of them came from **the environment, not the code**, and they reached *me*, the verifier, more than
the reviewers:

- **W1 r2: 8 names; W2 r1: 16; W3 r1: 22; W9: 5.** Each time, failures looked "new at the tip". Each name re-run alone on both
  sides passed on both, or failed on both. One package hit its 30-minute timeout on both sides, so the two populations
  could not even be compared.
- **W1 r1.** The reviewer's own scratch base trees reddened source-tree censuses, and so did the leg's 264 MB scratch
  checkout.
- **W2 r3.** 11 runner failures; restoring the five changed files to base gave the same 11.

⇒ **"Fails on head, passes on base" is necessary, not sufficient.** The runner must run each failing test **alone, on
both sides, more than once**, and report flake and environment as their own label. Without that, half the PROVEN
findings in a real repository will be timeouts and leftover scratch files.

### What reviewers missed that a run or the history showed

- **W2 census reds.** Fixing review condition D1, the leg broke three tests that passed at base: an
  unregistered refusal id and a missing ladder rung. No review saw them. My whole-package base-vs-tip run did.
- **W7, the worst one, and it is mine.** I joined the implementation and its independent tests by taking a test commit
  cut from a *newer* trunk than ordered. The join silently **deleted 51–52 files of the trunk**: `███████████████████`,
  `████████████████████████`, about 11k lines. The review had approved the round-1 tip and never saw the join. My verify
  built the merged tree and passed it, **because a clean build does not notice deleted files**. A merge-tree conflict
  caught it minutes before the owner would have landed it. Re-checking the pending branches against the trunk led to
  `git diff --diff-filter=D` against the trunk. **No model would have caught this by reading the diff. A one-line structural check
  would have caught it every time.** The reviewer the proposal describes reads *the PR*. The failure happened in *what the PR
  does to everything else*.
- **W3, successive rounds.** r2 found no other bypass of the role gate; r3 found three more spawn doors
  (`serverd.go`, `███████████████`). A grep for every caller of the spawn path, which is history and code search rather
  than judgement, finds all doors in one pass.

### What would make a reviewer's output useful instead of triage

- **A finding = site:line + the property + a killing mutation.** I pasted those verbatim into the next order (W5
  R2-C1, W1 C2 with "the probe counted 34, not 35"). Everything else I had to triage.
- **Triage came from:**
  - rulings the reviewer could not make: which option, which vocabulary, is this test or code wrong — W1 C6/C8, W3 T5;
  - UNRUN items: a sandboxed seat could not bind a listener or run the base side;
  - leftover scratch files in nearly every review, which I deleted before teardown.
- ⇒ The proposal needs a **fourth label beside PROVEN / QUESTION / DROPPED: `NEEDS-A-RULING`**, for "the evidence is clear
  and the choice is a human's". It also needs `UNRUN, with the command that would run it`. Our UNRUN lines were the most
  honest part of every review, and the reader could act on them.
- **Claims, not just defects.** Half my verify work is checking the *author's* claims: "mutation red then reverted",
  "C2 closed", "0 FAIL new". The W2 C2 case was a false claim of closure. A reviewer that checks each claim of the PR
  description against a run is the cheap win. See Proposal 1.

### Smaller points

- **Two model families add the least.** Our two best findings came from one reviewer with a probe. Our collisions
  came from an independent *test author*. A second *role* (a test writer from the spec) beats a second *model*. Keep
  Nemotron, but as the test author, not as a second opinion.
- **The bait in the demo is right.** Make it an environmental flake as well as code that only looks wrong, because that
  is the false alarm real users meet.

## 2. Proposals

## Proposal 1: Claim checker for pull requests

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** a pull request says "fixes #41, all tests pass, no behaviour
  change" — agents now write those sentences by the thousand, and nobody checks whether they are true.
- **who has this problem outside our fleet:** maintainers receiving agent-written PRs; teams whose AI agents open PRs
  with confident descriptions; release managers reading changelogs.
- **the data sources the agent investigates:**
  1. the PR description and commit messages, which are the claims;
  2. the diff, plus what it does to the rest of the tree: deleted files, files outside the stated scope;
  3. CI results and a sandbox run on base and head;
  4. the linked issue;
  5. git history.
- **what "explained with evidence" means here:**
  ```
  CLAIM 3 · "no behaviour change outside parse/"      FALSE
  Evidence  1. diff deletes 52 files outside parse/ (███████████████████, ██████) — git diff --diff-filter=D base…head
            2. the deletions arrive through merge commit 82e2…, whose second parent was cut from a newer trunk
  CLAIM 1 · "fixes #41"                               HOLDS
  Evidence  test written from #41's text fails on base 41aa…, passes on head 9f3c…; re-run 3×, both sides
  CLAIM 2 · "all tests pass"                          UNVERIFIABLE HERE
  Evidence  2 tests time out on BOTH sides (env); command to run them: …
  ```
- **the 3-minute demo, step by step:**
  1. Open a PR whose description makes five claims: two true, one false by a silent deletion, one false by a test
     that fails on head, one unverifiable.
  2. Run one command.
  3. The report gives each claim a label and its evidence, and the false deletion claim is caught with no model at all.
  4. Click a claim to see its evidence chain.
- **what goes to GitHub as open source:**
  - a CLI;
  - a claim extractor that turns a PR text into checkable claims;
  - checkers: structural (deletions, scope, renames), run-based (tests on base and head, re-run alone), and
    citation-based (history, issue);
  - a report in Markdown, HTML or as a PR comment.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **claim → check → label schema** and
  the structural checkers. Our masters already check every executor `_EXEC_STATUS.md` claim by hand ("mutation red then
  reverted", "0 new FAIL"). This piece would check those claims at return.
- **sponsor technology we can use honestly:** Nemotron as the claim extractor and the test writer; Cursor as an
  executor building it.
- **what the fleet must build in 144 h:**
  1. schema, CLI and PR reading;
  2. claim extraction;
  3. structural checkers;
  4. the test runner on base and head, with the alone-and-repeated re-run;
  5. history and issue citations;
  6. report, demo repository and README.
- **the biggest risk, and how we would know by day 2 that it is failing:** claim extraction from free text. By day 2 it
  must turn five real agent-written PR descriptions into checkable claims. If it cannot, we fall back to a fixed claim
  set: tests pass, scope, closes-issue, no deletions.
- **anything that touches internal mechanisms or patents:** none. Nothing is signed, bound or pinned.

This proposal and the leading candidate merge naturally. **Proof-carrying review + claim checking**: the reviewer
proves its own findings, and it checks the author's claims too.

## 3. The question I could not answer this month

*"Why did W7's branch delete half of `█████████████████` when W7 never touched it?"* (2026-10-16, stream
`the-cursor-egress-baseline`). The branch had passed my verify, held an accepted review, and its landing line was in the
owner's hands. When it conflicted against a newer trunk, the answer lived in four places:
- my own join commit `82e213100`;
- the test author's commit `701437a17`, cut from a trunk with 13,064 files where the order said a trunk with 13,011;
- the order's `- base:` line;
- the absorb commit `b1210af63`, which carried the deletions forward.

Assembling that took most of an hour. Until I counted files per commit, my first guess, "the trunk deleted them", was
wrong. I withdrew the landing line and rebuilt the branch as the wagon's own diff plus the one test file.
The question "what does this merge remove, and through which parent?" is mechanical. I still had to answer it by hand.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| locations in our private code | 7 | file names and line numbers in Saphan Studio's private source code |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
