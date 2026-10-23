# Answer — master of `saphan-message-is-one-verb-for-every-role-and-rides-the-bridge`

Author: MASTER code-fleet-2-1 (claude-code, claude-opus-5-5). Date: 2026-10-16.
Sources are this track's ledger, orders and review returns
(`██████████████████████████saphan-message-is-one-verb-for-every-role-and-rides-the-bridge/`).
Every claim here is from the record; where I remember rather than measured, I say so.

## 1. Critique of "Proof-carrying code review"

**Short verdict: the core is right.** What settled every real dispute on my track was a run, never an
argument or a vote. But the proposal proves the wrong half, against the wrong base, in an environment it
assumes works. Six points, each from a stream.

### 1.1 The proof that mattered most was a reviewer's failing test. That part is right; keep it.
- **Stream:** W31 `secret-scan-judges-path-segments`, review r6 (codex). Verdict `reject`, finding at
  `███████████████████████████████████`.
- **What the reviewer did:** wrote `TestReviewTwoInteriorSlashes` with █ ████████████ ██████ ██████████
  ███████ ███ ████████ ███. It **passes on base and fails on the tip**. The fix had classified the
  credential as a path and waved it through.
- **Why the executor's own tests missed it:** its control had a single interior slash
  (`███████████████████████`).
- **Lesson:** this is exactly the advisor's F2 mock-up, and it was the only finding on that stream nobody
  could argue with.

### 1.2 It proves defects but never proves that the PR's own tests mean anything. That half is missing.
- The most useful control in my reviews was the **unplug**: revert or disable the fix, and the PR's own
  tests must go red.
  - W31 r6: forcing `pathShaped` to return false reddened all four path subtests. That proved the split
    really controls the false positive.
  - W19 `class-reader-refuses-unreadable-yaml` r6: removing the new merge-key arm reddened the closure and
    push-door controls.
- A PR whose tests stay green with the fix removed is a vacuous PR. No comment-style reviewer catches
  that, and it is cheap to show: one mutation, one run.
- **Proposal:** every PR gets one more label, **TESTS PROVEN / TESTS VACUOUS**, next to
  PROVEN / QUESTION / DROPPED.

### 1.3 "Fails on head, passes on base" is only as true as the base, and the base is the hard part.
- **Wrong base.** I verified W30 `receipts-only-on-opens` and W31 against an order base (`e5201a0fc`)
  that was not the real fork point (`8aec337db`). Ledger entry: "MASTER WAS WRONG".
  - A differential against a stale base attributes trunk changes to the PR.
  - The fix is mechanical: use `git merge-base`, never "the branch's base field".
- **The PR alone is not the whole picture.** W24 and W13 were each reviewed and accepted alone. Their
  merge-tree was textually clean, and together they broke `develop` (2026-10-13).
  - Since then every landing on my track builds the **union with the current trunk** before it is handed
    over.
  - A PR reviewer that only compares base with head will repeat W24+W13. It needs a third cut: the head
    merged onto today's trunk.

### 1.4 False alarms on my track came from the environment, not from the models.
| stream | the false red | what exposed it |
|---|---|---|
| W25 `two-bridge-pulls-on-one-host-serialise` r3 | tests depended on the host's git identity: green on one host, red on another | a hermetic re-run, then round 4 made the identity hermetic |
| W25 r5 | a census failure naming `████████████████████`, a file absent at **both** base and tip | checking that the file exists at either cut, then a re-run with `-count=N`, which passed |
| W19 r3 and W31 | `TestRoleMemoryWriterReaderCensusIsDerivedFromTheTree` red | it fails identically at base, so it is pre-existing and not the PR's |

- **What would have caught them early:**
  - every red re-run at least twice;
  - every red checked at base;
  - every file a finding names checked for existence at both cuts.
- These three checks are cheaper than any model call. They belong in the verdict step **before** a
  finding becomes PROVEN.
- **What a test-writing agent will do:** write tests that depend on its own sandbox (clock, `$HOME`, git
  config). The proposal needs an explicit "is this proof hermetic?" check.

### 1.5 The output I actually use is the reviewer's list of what it could NOT run.
- My reviews must end with a section `## Unrun` that lists every check the reviewer could not run.
  `accept` is forbidden while it is non-empty.
- That section did more work than any finding:
  - **W19 r5** returned `accept-with-conditions` with Unrun = "live memory corpus not provisioned". It
    refused to infer real-tree compatibility from generated fixtures. Round 6 got the real corpus:
    1268 entries, identical at base and tip. That run settled it.
  - **W31 r6** said the importer suites timed out at 10 minutes. It also warned that "68 < 75 failing
    names" must **not** be read as an improvement, because the timed-out tests were simply not counted.
  - **W14 `bind-restore-guards-its-four-arms`**, rounds 5 and 7: the pristine-base run was unrun, because
    the review seat could not create a second worktree. Round 8 used a clone and passed.
- The proposal's QUESTION label is close to this, but it is about findings. **Unrun is about coverage.**
  It should be a first-class section of the report.

### 1.6 Two model families: I cannot attribute anything to the family.
- On my track, W19 rounds 4, 5 and 6 found different things: OC1, then the missing corpus, then a clean
  accept. The reason was that each round got a **different list of claims to test**, not a different
  model.
- Every review order pastes the executor's order and return as "every line is a CLAIM to test". The
  reviewer then refutes claims instead of reading freely. That structure is what made reviews useful.
- **Suggestion:** extract the claims from the PR description and commits, and test each one. Measure
  model diversity in the demo (one reviewer switched off) rather than assume it.

### 1.7 The reviewer must tell "found nothing" from "never ran".
- Today the W31 review died three times with no verdict:
  - on one host, `child-failed` and `session-died`, with nothing in the artifacts;
  - on another host, a model-proxy `502 upstream unreachable` after 180 s.
- An open-source tool that prints an empty report in that case would be worse than no tool.
- Bug filed: `█████████████████████████a-far-review-closes-session-died-with-no-reason-and-leaves-no-log_2026-10-16.md`.

**What would make the reviewer useful to me rather than something I triage:**
- (a) claims extracted from the PR and each one marked refuted / held / unrun;
- (b) one failing test per defect, plus one unplug per fix;
- (c) the true merge-base and a union with today's trunk;
- (d) every red re-run and checked at base before it is shown;
- (e) a mandatory Unrun list;
- (f) "the reviewer did not finish" printed as loudly as a defect.

## 2. Proposal

## Proposal 1: Why is this test red?

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** when CI turns red on a pull request, nobody knows
  in the first hour whether the PR broke it, it was already broken, it is flaky, or it only fails on that
  machine. People guess, re-run blindly, or merge anyway.
- **who has this problem outside our fleet:**
  - every team with a large test suite and a shared CI;
  - open-source maintainers who must decide whether a contributor's red is the contributor's fault.
- **the data sources the agent investigates:**
  - (1) the CI log and the failing test names;
  - (2) git: the true merge-base, the PR's commits, blame of the failing assertion;
  - (3) a sandbox where it re-runs the failing tests at base, at head and at head-merged-on-trunk, with
    repeats;
  - (4) the CI history of the same test on the default branch (its flake rate);
  - (5) the test source: what it reads from the environment (clock, `$HOME`, network, git config).
- **what "explained with evidence" means here — one answer:**
  ```
  TestBridgeLockReleasesOnExit · NOT CAUSED BY THIS PR · ENVIRONMENT
  Base 41aa… (merge-base)   FAIL 3/3      Head 9f3c…   FAIL 3/3      → pre-existing at this host
  Clean container, base     PASS 3/3      Clean container, head PASS 3/3
  Why    the test reads `git config user.email` (lock_test.go:57); this runner has none set.
         Introduced by commit 2b71… ("use committer identity in lock owner"), 2026-09-30.
  History on main: 0 failures in 40 runs on runners with identity set; 6/6 failures on the new runner pool.
  Next   set a hermetic identity in the test (GIT_AUTHOR_* env) — not this PR's job.
  ```
  The PR's own reds get a different label: `CAUSED BY THIS PR · PROVEN`, base PASS 3/3 and head FAIL 3/3,
  plus the commit that flips it found by bisecting the PR's commits.
- **the 3-minute demo, step by step:**
  1. A public repo with a PR whose CI shows four reds.
  2. One command.
  3. The report sorts the four reds:
     - one caused by the PR (bisected to the commit);
     - one pre-existing at the merge-base;
     - one flaky (2/5 at base and at head);
     - one host-dependent (green in a clean container, with the environment read that makes it red).
  4. Click any red to see its evidence table.
  5. Then the "bait": a red the PR's author swears is flaky, which the agent shows is the PR's.
- **what goes to GitHub as open source:**
  - the CLI;
  - the red-classification schema (JSON);
  - the re-run engine: base / head / head-on-trunk, N repeats, a clean-container variant;
  - the bisect over the PR's commits;
  - the environment-read detector for Go and Python tests;
  - CI-log readers (GitHub Actions first);
  - a markdown report.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **base / tip / union
  failing-name differential with repeat classification**. Every master on this fleet runs it by hand
  today: my `mv-prep.sh` and `recheck.sh` do exactly this, and verifying against the true fork point is a
  standing rule. It is what every named review and every master-verify needs. Written in Go, it is a
  library the backend imports.
- **sponsor technology we can use honestly:**
  - NVIDIA Nemotron for the "why" paragraph and for reading test sources to find environment reads, as
    an open model runnable next to the CI;
  - Cursor as one executor.
- **what the fleet must build in 144 h:**
  1. Schema, CLI, CI-log reader (GitHub Actions), merge-base resolution.
  2. Re-run engine: base / head / union, repeats, clean container (Go and Python).
  3. Bisect over the PR's commits for PR-caused reds, plus flake rate from default-branch history.
  4. Environment-read detector and the model-written "why" grounded in the table.
  5. Report, demo repo with the four planted reds and the bait.
  6. README, one-command install, video.
- **the biggest risk, and the day-2 signal:** re-running real suites is slow, and container parity is
  hard.
  - By the end of day 2 the engine must classify the demo repo's four reds correctly from a cold start in
    under five minutes.
  - If it cannot, we cut the clean-container leg and keep base / head / repeat only.
- **relation to the leading candidate:** this is the verdict step that the proof-carrying reviewer needs
  anyway (§1.3 and §1.4). If the owner keeps the leading candidate, take this as its wagon 3, not as a
  rival.
- **anything that touches internal mechanisms or patents:** none. No signing, no order binding, no
  pinning. Nothing of the engine is needed.

## 3. The question I could not answer this month

*"Why did the W31 review die three times today with no verdict?"* (2026-10-16, stream
`secret-scan-judges-path-segments`.)
- **The record:**
  - `dispatch place` printed `closed session-died … UNEXPLAINED [refusal-unexplained]` for two far runs,
    and `child-failed` for a third.
  - The far run directories held only empty `evidence/ provision/ state/ tmp/`.
- **Where the answer was:** I moved the review to a local seat as a control. That run died too, but there
  the answer was on disk: `transcript.txt` showed `API Error: 502 … the upstream could not be reached`,
  and the proxy's `█████████████████████████████` showed `guardrail-proxy-upstream-unreachable` 11 times
  out of 11.
- **Reading audits across worktrees showed:**
  - the same pattern on four other tracks' legs since about 12:08 local;
  - a direct request to the model API from the same host answering in 0.06 s;
  - codex legs unaffected.
- **Cost:** assembling that took most of an hour, across four files on two hosts and five worktrees.
  The engine had the reason on disk and still wrote the row as reasonless.
- **Recorded:** bug `a-far-review-closes-session-died-with-no-reason-and-leaves-no-log_2026-10-16.md`
  (commits dc1f3fcd9 and 21d496b5b), the scribe relay `a33d674d62d72160`, and the W31 review re-ordered
  on codex (66bf3005c).
- **What I wanted:** an agent that answers "why did this run die?" from the run row, the transcript, the
  proxy audit and the sibling runs that died the same way.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| internal locations | 3 | where a document or log lives on our machines or in our private record |
| locations in our private code | 3 | file names and line numbers in Saphan Studio's private source code |
| security details | 1 | the exact input or condition that gets past one of our own guards, or the details of a residual risk in a live flow |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
