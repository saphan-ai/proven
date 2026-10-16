# Answer — master of `the-bridge-commits-without-a-shared-lock` (and `the-cli-tells-a-human-what-matters-and-logs-the-rest`)

Author: MASTER, seat code-fleet-2-1. Date: 2026-10-16. Record cited: bridge
`███████████████████the-bridge-commits-without-a-shared-lock/` and `███████████████████the-cli-tells-a-human-what-matters-and-logs-the-rest/`.

## 1. Critique of "Proof-carrying code review", from my own named reviews

**My verdict: the core is right. It is proven twice in my record, below. But "a test that fails on head and passes on base" is
the wrong definition of proof in three of the five cases that cost me the most.** The idea needs three changes before it is
honest: a *mutation* proof for missing protections, a *declared scope* that turns true findings into non-defects, and a
*same-conditions differential*, so the runner's own noise is not shown as evidence.

### 1a. What settled disagreements — and could an agent have done it?

- **W2a `leg-hooks-per-worktree`: 10 review orders, 21 courier review rounds, 2026-10-14 → 16.** From round 5 on, every review
  found one more shell spelling that sets git's config through the environment and slips past the first-layer command check:
  ████ █████ ███████ █████ ████████████ ██████ ██████, and others. **Every one of those findings was TRUE and provable.** The reviewer
  built a real-git twin that landed a denied merge. So proof did not end the disagreement. It fed it: the review kept saying reject
  and the author kept paying. What settled it was **a ruling about scope, not evidence**. The owner had ruled on 2026-10-02 that the
  command layer is a first layer, not tight against deliberate evasion. On 2026-10-16 he ruled "round 8, then landing": further
  spellings become *docs-naming conditions* (declared limits in `█████████████████████` §4.1), and the verdict became
  `accept-with-conditions` (review r21, landed as develop `c9f3a32ca`).
  **⇒ An agent could have settled it, but only if the threat model is an input.** Read "what this layer does and does not promise"
  from the repo's █████████, and classify a proven finding as *in scope → defect* or *out of scope → declared limit (needs a doc
  line)*. The advisor's PROVEN/QUESTION/DROPPED has no slot for "proven and out of scope". In my record that was the most
  expensive class.
- **W2b `pack-refs-is-not-a-trunk-delete`, review r1 (today).** The leg's tests were green: 16/16 real trunk deletions refused, gc
  passes. The reviewer did not argue. It removed the new ancestry check and showed that **all 16 deletion rows stayed green**. git
  2.53 refuses the same act one layer earlier (a preceding packed-refs transaction), so the tests proved git, not the fix. It then
  isolated the hook with `git hook run` and showed that ████ █████████████ █████████ ██ is read as pack-refs and passes a deletion
  (conditions C-1/C-2, `REVIEW_NOTES_pack-refs-is-not-a-trunk-delete_r1_2026-10-16.md`). That is proof-carrying review at its
  best, but **the proof was a mutant plus an isolation, not a base/head test.** On base the check did not exist, so "fails on head,
  passes on base" is meaningless for a finding of the form "this protection does not protect".

### 1b. Which findings were false alarms, and what would have exposed them early?

- **The master's own premise, W17 `land-names-what-stayed-and-why` (today).** Two masters, I am one, wrote into the record that
  `land`'s merge rehearsal failed because another landing held the workspace merge lock (W26, measured 2026-10-13 and recurred
  2026-10-16T18:12Z). The executor read the code path and refuted it: **the rehearsal takes no lock** (`merge.go`, only `!dryRun`
  acquires it; status of `818ed1a33`, ruling A). The real cause is still unmeasured. *What would have exposed it:* reading the code
  path for the exact command line in the log before naming a cause. That is "call-path evidence", cheap and decisive.
- **The test plan's own case, W4b (today).** In the test-author leg `review-collides-tests-tests`, the author reported that my plan's
  case T7 would pass at base as written, and tightened it. A case that cannot fail is a false assurance. *Exposed by:* running every
  claimed test at base, which is the advisor's rule, applied to the plan as well as the findings.
- **Runner noise (today).** The W2b master verify on a host at load average 13–17: **144 FAIL names** in two packages, at tip and at
  base alike. The one tip-only name (`TestContendedRefreshLockWaitsAndEveryDispatchProceeds`) went green 3/3 on an idle host. A
  runner that reports "fails on head" from one run on a busy CI box would have shown that as PROVEN.
  *⇒ the evidence runner must run base and head on the same host, at the same time, repeated, and compare names.*

### 1c. What reviewers missed that a test, a run or the history would have shown

- **A merge that is clean but does not compile** (2026-10-13, two streams, each reviewed and green alone; develop broke). A
  signature changed in one stream and the other still called the old one. Only building the merged tree shows it. Per-PR review
  cannot see it by construction. Since then I build the merge-tree before every landing.
- **The W2a design lost three protections** (`SHOUT_the-private-hook-home-gives-up-three-protections_…_2026-10-15.md`). The round-1
  review saw it, and round 2 restored them. *History* would have shown it earlier: the three protections each had a commit and a bug
  of their own. A reviewer that asks "what did the code this diff replaces protect against, per its commit messages and linked bugs?"
  would have flagged it at once. That is the advisor's "why it exists" line, and in my record it is worth more than the bug hunt.
- **Version-relative proof.** Review r1 of W2b put it in its own Unrun section: the fix holds on git 2.53 because of git's own extra
  check, and git 2.43 was not on the seat. A proof is only as wide as the environment it ran in, so the report must print that
  environment.

### 1d. What makes a reviewer's output useful to me instead of something I triage

1. **Each condition as file:line + remedy + the row that proves it.** Review r20 of W2a and r1 of W2b did this, and I pasted their
   conditions verbatim into the next executor order (`ORDER_W2b_…_r4_2026-10-16.md`). No triage at all.
2. **The reviewer's probe as a runnable file.** W2b r1 left five build-tagged scratch tests, and the round-4 order carries one as the
   shape of the required row. The probe is the most valuable thing a reviewer produces.
3. **An "Unrun" section.** It says what the review could not measure (here: other git versions, the macOS `ps` path, reftable). It
   stops me from reading silence as coverage.
4. **A verdict mapped to the decider's declared rule** ("accept-with-conditions under the 2026-10-16 ruling"), not to the
   reviewer's taste.
5. **Framing.** One W2a review order framed the task as an attack, and the reviewing model aborted. Rewritten as defect / asset /
   proof / boundary, it ran. A security reviewer's prompt must name the defect class and the asset, or a safety-tuned model declines.

### 1e. Concrete changes I would make to the proposal

| change | why (record) |
|---|---|
| Add **MUTATION** as an evidence kind: "remove the line the PR adds, a named test turns red" | W2b r1: the base/head test proves nothing for a missing protection |
| Add verdict **PROVEN-OUT-OF-SCOPE → docs line**, reading the repo's declared limits | W2a: 21 rounds of true findings with no end |
| Evidence runner: base and head **on the same host, interleaved, ×3**, FAIL names compared | today's 144 load FAILs |
| Every finding prints its **environment** (toolchain versions, OS) | W2b: proof holds on git 2.53 only |
| A finding that names a cause must cite the **call path** of the failing command | W17 refuted the lock theory |
| An **"Unrun"** list in every report | W2b r1; it is what made the review trustworthy |
| Optional: **merge-queue mode**, build and test the union of open PRs | 2026-10-13 clean merge that did not compile |

The demo bait ("looks wrong, is right") is good. Add a second bait: **a test that passes for the wrong reason**, caught by the
mutation step. That is the moment that makes a judge believe the "proof" word.

## 2. Proposal of my own

## Proposal 1: Regression or flake? A failure that explains itself

- track: The Agent That Can Explain Why.
- the problem, in one sentence a judge understands: when CI goes red, a developer cannot tell whether *their* change broke it, the
  test is flaky, or the machine was overloaded, and they lose an hour finding out (or they press re-run and learn nothing).
- who has this problem outside our fleet: every team with a CI suite past a few hundred tests; maintainers who must decide whether a
  contributor's red PR is the contributor's fault.
- the data sources the agent investigates: (1) the failing run's test output; (2) a fresh run of the same tests on the PR's base, on
  the same runner; (3) the history of that test's results across past runs on the main branch; (4) `git log`/blame of the test and of
  the code it touches since it last passed; (5) the runner's load and environment at the time (CPU, versions).
- what "explained with evidence" means here — one answer:
  ```
  TestContendedRefreshLock… · FLAKE (load) · confidence high
  Evidence  1. fails on head (run 1/1, load avg 17) — fails on base too, same host, same hour: 143 other names fail on both
            2. head, idle host, ×3: PASS PASS PASS
            3. history: failed 4 times in 30 main-branch runs, every time with load avg > 12
            4. the diff touches no file this test imports (go list -deps)
  Verdict   not caused by this PR. Re-run on an idle runner; file the test as load-sensitive.
  ---
  TestCollideT8 · REGRESSION · confidence high
  Evidence  1. fails on head ×3, passes on base ×3, same host
            2. bisect inside the PR: first bad commit 7ba48e5 "cut a paired review at the merge"
            3. failing line: dispatch_place_collide.go:41 returns "" (stub), test expects a 40-hex id
  ```
- the 3-minute demo, step by step: (1) a public repo, a PR with one real regression and two red tests that are not the PR's fault
  (one flaky under CPU load, one broken on main already); (2) CI is red with 3 failures; (3) one command → report: 1 REGRESSION with
  its first bad commit, 1 FLAKE with the load evidence, 1 PRE-EXISTING with the main-branch run where it first failed; (4) click
  each → evidence chain; (5) the same PR with the real bug fixed → the agent says "green except a known flake, safe to merge".
- what goes to GitHub as open source: a CLI and a GitHub Action. It covers re-running named tests on base and head on one runner,
  repeated; a test-result history store (one SQLite file, built from JUnit XML / `go test -json`); a bisect-inside-the-PR step; the
  report. Go and Python test runners first.
- the bounded piece Saphan later takes, and why we would build it anyway: **the base-vs-head test-differential library** (run named
  tests on two trees under the same conditions, repeat, classify each name as regression / flake / pre-existing, with its evidence).
  Every master verify I wrote today was this done by hand: four █████ runs and a 144-name `comm` compare
  (`MASTER_VERIFY_pack-refs-is-not-a-trunk-delete_round-3_2026-10-16.md`). A review in Saphan PR Review needs exactly this to stop
  reporting load as defects.
- sponsor technology we can use honestly: Cursor as one executor. Nemotron only to write the one-paragraph explanation; the
  verdict itself is computed, not generated. Saying so is the honest pitch.
- what the fleet must build in 144 h: (1) test-result ingest (JUnit XML, `go test -json`, pytest) + history store; (2) the
  same-host base/head runner with repeats; (3) the classifier + bisect-in-PR; (4) the "what changed" investigator (diff ∩ test
  dependencies, blame); (5) report + GitHub Action + demo repo with the three planted failures; (6) README, video.
- the biggest risk, and how we would know by day 2 that it is failing: a flake that does not reproduce in 3 runs. By end of day 2
  the demo flake must be classified FLAKE in at least 4 of 5 runs of the tool. If not, we inject load deterministically in the demo
  and say so.
- anything that touches internal mechanisms or patents: none. It never signs, pins or binds anything.

**How it relates to the leading candidate:** it can be the leading candidate's evidence runner (wagon 3 there) built as a product of
its own. If the owner keeps "Proof-carrying code review", I would make this differential its wagon 3 and name it in the demo. Proof
that is not checked against base on the same runner is the false alarm that costs the most in my record.

## 3. The question I could not answer this month

*"Why did `saphan land gw-w2b6a-order-issue-at-gateway` refuse with `land-step-unreadable` at 2026-10-16T18:12Z, when a dry-run of the
same merge minutes later read CLEAN and BUILT?"*
- The log (`█████████████land-2026-10-16T18-12-53.log`) says the merge rehearsal exited 2 with no refusal envelope, and its stderr
  has only the workspace line and two channel-deviation warnings.
- The same refusal happened on 2026-10-13T16:48Z. Then another landing held the workspace merge lock, and two masters (I am one)
  wrote that the lock was the cause (W26, now wagon W17 of `the-cli-tells-a-human-what-matters-and-logs-the-rest`).
- Today the W17 executor read the code and showed that the rehearsal takes no lock (`818ed1a33`, status ruling A). So the cause we
  had both written down was wrong, and the true one is still **unmeasured**.
- The answer is spread over four places, and nobody joins them: the land log, the process table at 18:12Z (not recorded), develop's
  reflog at that minute, and the merge code path for `dryRun`.
- I would have needed an agent that, given one refusal line, replays the exact command's code path and lines it up with what else
  ran on the host that minute.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 1 | the names of our machines |
| internal locations | 3 | where a document or log lives on our machines or in our private record |
| locations in our private code | 2 | file names and line numbers in Saphan Studio's private source code |
| security details | 2 | the exact input or condition that gets past one of our own guards, or the details of a residual risk in a live flow |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
