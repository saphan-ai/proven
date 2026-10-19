# Answer from the scribe

Author: the scribe. Date: 2026-10-16. Language: English.

Every count below was measured on 2026-10-16. The commands, SQL and raw outputs are in two working files
outside this folder. They live there because this folder holds one file per author:

- `███████████████████████████████unexplained-classes.md`
- `███████████████████████████████review-disagreements.md`

The method was read-only:
- the bug backlog (`████████████████████████████████`, █████ files);
- the record (`█████████████████████` and `██████████████████████`, about 8,500 documents);
- the engine store, opened read-only (92,390 run rows, 1,928 review-gate rows);
- the engine's own logs: 741 placement logs and 150 land logs.

⚠ **The bug counts are lower bounds.** Classes were assigned by hand after a title screen. A file whose title
has none of the screen words was never read. Record counts come from regexes and are noisy; they
corroborate and do not rank.

---

## Part 1: the three most frequent classes of "nobody could explain why"

| rank | class | bug files | store / logs | record docs |
|---|---|---|---|---|
| 1 | **A failure closed with no cause on its row** | ██ | **3,030 of 3,030** failure rows carry an empty rule and reason (1,176 in the last 30 days); **52 of 741** placements end `UNEXPLAINED` | 42 |
| 2 | **A status that contradicts reality** | ██ | **421** rows say "failed" while their own `█████████` says the work finished (176 in 30 days); **6 of 25** "Not landed" land logs quote the engine's own "stream is already merged" | 51 |
| 3 | **A check that passes or fails for a reason other than what it claims to measure** | ██ | no column holds it | 116 |

Runners-up, counted but not ranked:
- a refusal that misnames or hides its cause: ██ bug files;
- a silent wait, hang or blind watcher: ██ bug files.

### 1 · A failure closed with no cause (ranked first on event volume)

**What it is.** The run closes as `child-failed`, `vendor-error`, `session-died` or `auth-failed`, often at
$0 or with a 0-byte transcript. Nobody can say whether the code, the vendor, the network or the sandbox
killed it.

**SQL:**

```sql
SELECT … COUNT(*), SUM(████████████='' AND ██████████████='')
FROM ██████████
WHERE status IN ('child-failed', 'vendor-error', 'session-died', …)
```

It returns 1,176 / 1,176 for the last 30 days and 1,854 / 1,854 before that.

**The answer was on disk.** All 52 `UNEXPLAINED` runs have a non-empty local `stderr.log`, and 36 of the 52
contain a readable cause:

| cause found in `stderr.log` | runs |
|---|---|
| far binary "unknown command" | 13 |
| proxy `API Error: 5xx` | 12 |
| ssh broken pipe | 7 |
| OAuth 401 | 4 |

Nothing lifts that line onto the row.

**Examples:**
- `a-far-review-closes-session-died-with-no-reason-and-leaves-no-log_2026-10-16`: *"the engine HAD the reason
  on disk and still recorded the row reasonless"*.
- `OWED_confine-exec-refuses-a-codex-spawn-when-a-sqlite-shm-file-…`: *"dies at $0 in 14 s with
  `child-failed` … an EMPTY transcript; nothing in `saphan run`'s own summary names the cause"*.

### 2 · A status that contradicts reality (ranked first on bug files)

**What it is.** "Not landed" after a successful merge. A destroyed worktree shown as live. A finished leg
recorded as dead. A phantom `running` row that holds a seat.

**The answer was there too.** It was in git (`git log develop`, the worktree's own status file) and in
the store's own columns: `status` and `█████████` contradict each other on 421 rows.

**Examples:**
- `HIGH_land-reports-not-landed-after-a-merge-that-succeeded_2026-10-14`: the owner asked whether the tool
  *"shows a status different from reality"*. Today the same land printed **"ALREADY LANDED"** for a stream
  that had not merged.
- `a-leg-that-finished-its-work-is-recorded-as-child-failed-and-only-the-tree-says-otherwise`: *"The work is
  whole. The row says it died."* It happened four times in one shift.

### 3 · A check that passes or fails for the wrong reason

**What it is.** Exit 0 and green when nothing ran, or a red caused by the host and blamed on the branch.

**Examples:**
- `a-filtered-go-test-that-matches-nothing-exits-zero-and-prints-ok`: `tests run: 0 -> ok`. The red survived
  four rounds.
- `a-red-count-that-greps-FAIL-lines-reads-a-package-that-compiled-nothing-as-green`.
- `the-go-build-cache-grew-to-317G-on-█████-filled-the-disk-and-a-union-reported-88-false-reds`.

**The answer was in the raw tool output the summary threw away:**
- RUN/PASS/FAIL/SKIP counts;
- `[build failed]` and `no tests to run`;
- the host's own disk and load;
- for a red blamed on a branch, a same-host baseline run on the merge base.

**One line for all three classes.** In each one, the explanation existed in a second data source (stderr,
git, raw test output, host state), and the tool that reported the outcome never read it. That is
exactly the shape of the *Explain Why* track.

---

## Part 2: reviewers who disagreed on the same stream

**Volume.** 1,087 unique review returns in the record (1,050 with a readable verdict). 207 streams had two
or more reviews. 104 changed verdict class across rounds, most of it normal progress: reject→conditions
32, conditions→accept 26.

| question | count | source |
|---|---|---|
| designed blind pairs (two model families, one order, same SHA) | **14** | record, read by hand |
| … that disagreed | **11 of 14**; 10 split reject against non-reject | same |
| concurrent cross-family review runs (same target, ≤ 30 min apart, both verdicts readable) | **41**, of which **31 differ (76 %)** | store `██████████` + each run's return |
| … reject-vs-accept splits where the codex family was the rejecting side | **22 of 25** | same |
| reject rate by family (2,075 readable review runs) | codex **80 %** · claude **27 %** · qwen **10 %** | same |
| sequential reviews of the **same SHA** that disagree | **7**, of which **4 were environment artefacts** (missing ref, no DNS, unwritable `/tmp`, an order naming a commit that does not exist) | record + store |
| record documents saying an earlier review finding was refuted | **18** | phrase census over the 1,087 returns |
| explicit "N reviews missed it" statements | **5** (e.g. *"six reviews missed"*, *"four rigorous reviews missed"*) | record |

**How the disagreements were settled:**

1. **By rule, not by weighing.** The owner's rule *"a reject from either is a reject"* decided every designed
   pair. It kept a master's taste out of the gate.
2. **By measurement, twice.** Two pairs contradicted each other **on fact**. The master settled both with a
   probe of the code (`█████████████████████silent-class-guard-r3…/ARBITRATION.md`,
   `…/silent-class-pty-r5…/ARBITRATION.md`). Both times the **rejecting** reviewer was right. In one of them
   the accepting reviewer *"inherited the author's frame from the author's test names"*.
3. **By an owner or advisor act, three times.** A reject was recorded as conditions. The law lets a master
   raise a class, never lower it, so each downward move is recorded as the owner's or the advisor's.
4. **Never, once.** A real same-SHA conflict closed on `conditions` with no arbitration document
   (`the-trunk-does-not-build-and-no-test-run-can-say-so-r3`).

**The pattern that matters most for the candidate.** A disagreement almost never comes from two reviewers
reading the same evidence differently. The accepting reviewer checked what the order pointed at. The
rejecting reviewer found something **outside that pointing**, on *"different ground"*
(`SILENT_CLASS_TALLY_2026-09-01`).

---

## Part 3: critique of "Proof-carrying code review"

The record supports the core idea: two families, and a finding that must prove itself. Two families
disagree three times out of four, and twice a measurement showed the rejecter was right. Here is what
the record says the proposal gets wrong or misses.

1. **The evidence runner will be the main source of false findings, not the models.**
   - 4 of the 7 same-SHA disagreements were the environment: a missing ref, no DNS, an unwritable `/tmp`.
   - Class 3 above is a whole family of checks that lie: `tests run: 0 -> ok`, a build failure read as a test
     verdict, a disk-full host producing 88 false reds.

   A **PROVEN** label therefore needs controls, not only "fails on head, passes on base":
   - the test **ran** (its RUN count is greater than 0);
   - it fails **with the asserted message**, not a compile or setup error;
   - the base run is on **the same host, in the same container**;
   - an environment failure is a fourth label, **COULD NOT RUN**, never a finding and never a pass.

2. **Prove the misses, not only the hits.** Of the defects the record says reviews missed:
   - one was missed by six reviews because each reviewer's test run was cut short by its sandbox;
   - another was missed because every review ran only one package.

   The report must say **what was not run and why**: an `UNMEASURED` section. Without it, a clean report
   reads as "no defect" when it means "not looked at".

3. **Model family predicts the verdict more than the code does.** Reject rates are codex 80 %,
   claude 27 %, qwen 10 %. Show each reviewer's base rate beside its verdict.

   The proposal's rule of downgrading an unproven claim to a **question** is the right fix for that
   prior. Make it the headline of the demo, not a detail.

4. **The demo bug should sit outside the diff's pointing.** The record's blocking findings were found
   outside what the author pointed at, every time. Plant one bug in a **caller in another file** that the
   PR breaks. That is the case where a second family and a run earn their place, and a single model
   reading the diff will miss it.

5. **Two independent verdicts must survive side by side.** Our own engine could not keep two blind
   reviews: a second return overwrote the first, every designed pair was preserved by hand-copying files,
   and the fix stream was parked after six rounds. Build the schema to hold N independent verdicts per
   finding from day one. It is also the piece the backend needs.

6. **"Settled by evidence, not by a vote" needs a fallback.** Only 2 of 11 disagreeing pairs were settled
   by evidence. The rest went by a rule. When neither side can produce proof, the tool needs a stated
   rule, and the record's rule is "a reject from either stands as a QUESTION for the human". Say so,
   rather than implying evidence always exists.

**Verdict.** Keep it as the lead topic. Points 1 and 2 are the difference between a demo that holds up and
one a judge can break with one flaky test.

---

## Proposal 1: Is this red mine? A failure explainer

This proposal can stand alone, or become the reviewer's second mode, since it shares the evidence runner.

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** when a CI job or an AI agent's run fails, the log
  says *that* it failed, and a person spends an hour finding out whether the code, the machine, the network
  or a flaky test did it. The reverse case is worse: a green that ran nothing.
- **who has this problem outside our fleet:**
  - every team with CI;
  - every team running coding agents, whose runs die with "error" and a 0-byte transcript;
  - maintainers asked "is this red caused by my PR?".
- **the data sources the agent investigates:**
  1. the job's raw logs and stderr (not the summary line);
  2. the test runner's own counts (RUN/PASS/FAIL/SKIP, build failed, no tests to run);
  3. git history (what changed between the last green and this red);
  4. the same job on the merge base, re-run on the same runner;
  5. the runner's state (disk, memory, network reachability);
  6. the history of this test (flaky before?).
- **what "explained with evidence" means here:**
  ```
  RED on PR #128 · job test-linux · VERDICT: NOT THIS PR (environment)
  Cause     the runner's disk was full: "no space left on device" (log line 4,812)
  Evidence  1. same job on base 41aa… on the same runner: RED, same line
            2. runner disk at start: 99.8 % (df, step "setup")
            3. this PR changes 3 files, none read by the failing package
  Not run   integration suite (skipped: needs secret, not available to forks)

  GREEN on PR #131 · job test · VERDICT: HOLLOW
  Cause     the filter "-run TestParse" matched 0 tests in ./parser; exit 0
  Evidence  go test -v: "testing: warning: no tests to run" · RUN count 0
  ```
- **the 3-minute demo, step by step:**
  1. A public repo with four CI runs: a real regression, a flaky test, a red caused by the runner, and a
     green that ran nothing.
  2. One command per run.
  3. Each run gets one of four verdicts (**THIS PR · NOT THIS PR · FLAKY · HOLLOW**), each with its evidence
     chain.
  4. The punchline: the hollow green, which every dashboard shows as a pass.
- **what goes to GitHub as open source:**
  - the CLI and a GitHub Action;
  - the cause schema (verdict, evidence lines, what was not run);
  - log and test-output parsers (Go, pytest, jest to start);
  - the base-rerun step;
  - model adapters for any OpenAI-compatible endpoint.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **cause classifier over a
  run's stderr and test output**. Our store has 3,030 failure rows with no cause, and 36 of 52 measured ones
  had it in `stderr.log`. This is class 1 of Part 1. We owe it regardless.
- **sponsor technology we can use honestly:**
  - Nemotron as the reader of long logs (open model, runnable locally);
  - Cursor as one of the fleet's executors.
- **what the fleet must build in 144 h:**
  1. the schema, the CLI, and a reader for GitHub Actions logs;
  2. test-output parsers and the HOLLOW check;
  3. the base re-run and the runner-state probe;
  4. the model step that writes the cause from the evidence and must cite log lines;
  5. the demo repo with four planted runs, and the report;
  6. the README, the Action and the video.
- **the biggest risk, and how we would know by day 2:** re-running the base job reproducibly. By the end
  of day 2, the demo's environment red must reproduce on base and the regression must not. If it fails,
  we fall back to log-and-history evidence only and say so in the report.
- **anything that touches internal mechanisms or patents:** none. No order or policy binding, no pinning,
  no signed record. It reads logs and git.

---

## The question I could not answer this month

**"Did the merge step fail because another landing held the lock, or because develop moved under it?"**
(2026-10-16, stream `gw-w2b6a-order-issue-at-gateway`.)

The owner's `saphan land` acted the two gates and then refused the merge: *"the merge rehearsal exited 2 and
its stdout is not a refusal envelope"*, with the next step reading *"there is no by-hand paste for this"*.
The same screen printed **"ALREADY LANDED"**, and the owner read it as "it landed". It had not.

I had five sources:
- the land log, whose stderr was trimmed at 4,217 bytes, so the end of the error was cut off;
- the lock file, whose content named a process that was long dead;
- the logs directory, which showed another stream's landing finishing two minutes later;
- `git reflog develop`, which showed develop had not moved at 18:12Z;
- a dry-run of the same merge minutes later, which came back clean and built.

The record has a twin of this from 2026-10-13, where a held lock produced the same unreadable refusal. I
still cannot say which cause it was this time: **the holder at 18:12Z was never recorded**. The owner
re-ran the same command and the stream landed.

An agent that answers *"why did this step fail?"* from the full stderr, the lock's history, the reflog and
the concurrent acts would have answered in a minute. It could also have said what I had to say by hand:
the earlier gates stay recorded, nothing merged, and the same line can simply be run again.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 1 | the names of our machines |
| internal locations | 6 | where a document or log lives on our machines or in our private record |
| store schema | 6 | names of tables and columns in our private store |
| bug counts | 6 | how many bugs our backlog holds |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.

## Correction (2026-10-19, by the scribe)

This answer is published verbatim, so its numbers stay as written above. Two of them were wrong. On 2026-10-19 the scribe
re-read every pair by hand: the extractor had read a template line in one model's output as its verdict.
- Concurrent cross-family pairs with two readable verdicts: **40, of which 28 differ (70 %)** — not 41 and 31 (76 %).
  "Three times out of four" should read "seven times out of ten".
- Reject rate of qwen-code: **19 %**, not 10 %. The rates of codex (80 %) and claude-code (27 %) stand.
The full account of all 31 pairs is in [`tests/corpus/PROVENANCE.md`](../../tests/corpus/PROVENANCE.md).
