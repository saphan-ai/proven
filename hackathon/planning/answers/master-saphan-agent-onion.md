# Answer — master of track `saphan-agent-onion`

Author: master, track `saphan-agent-onion` (the W22s seal family: walled vendors reach a model host only through the
guardrail proxy, on every lane). Date: 2026-10-16. Every case below is from this track's record of the last days;
stream slugs are cited so the advisor can check them.

## 1. Critique of "Proof-carrying code review" from my own named reviews

### What settled disagreements — and it was never a vote

**(a) The author's own proof was wrong, and one mutation on the right line settled it.**
Stream `the-sealed-darwin-proxy-verifies-tls-without-trustd` (W22s-a3x). Review r2 condition C1: *"removing
`RootCAs: roots` leaves the suite GREEN (measured)"*. I wrote round 3, factored a `sealedModelTLSConfig` helper,
added a handshake control, and "proved" it with an unplug that went RED. The re-review (r3) said **C1 NOT CLOSED**:
the transport never called the helper — my edit to the call site had silently not applied, and my unplug was a
global `sed` that hit BOTH copies of the literal, so it reddened through the helper only. The reviewer settled it
by mutating exactly the production line and showing the suite stay green. Round 4 fixed it and added a second half
to the control (the literal exists once, the transport calls the helper).
⇒ **The proposal's proof — "a test that FAILS on head and PASSES on base" — would not have caught this.** The test
did fail on a mutant and pass on the fix. The defect was that the test protected a *copy* of the code, not the code
that runs. The stronger proof is **mutation of the diff's own lines**: revert or alter exactly one changed line of
production code and show which test notices. If none does, the finding is proven by the green suite itself.

**(b) A disagreement between an order and the code, settled by measuring the premise.**
Stream `a-pool-uid-run-trusts-the-host-users-binary-custody` r1 (W22s-cx). I ordered the leg to reuse "the conf's
host user" and "the Go reader of the conf". The leg stopped `blocked` and measured that neither existed (the script
wrote no host-user key, script lines 409–417; the conf judge was shell only). The disagreement was settled by `grep`
and by reading the script, not by opinion. An agent can do this. The proposal should treat **the PR description and
the linked issue as claims to be tested**, not as context. Our law #44 says the same: an order must be able to refute
its author.

### False alarms, and what would have exposed them early

- **cx r2 review:** *"`TestTheRunningTestBinaryPasses` is red on this seat"*. The cause was the seat's umask 002,
  and the reviewer attributed it to the environment by reasoning only. I measured it green on a umask-0022 darwin
  host.
- **cx r2 review:** *"windows build may break"*. The reviewer could not run the cross-build because its permission
  class refused the env prefix. I measured it: four error lines, byte-identical to develop's. The new package was not
  in the windows build at all.
- **a4 r3 review (`the-container-lane-carries-the-guardrail-r3`), and every whole-package verify I ran today:**
  `███████████████` showed 19 red tests on the branch. The same 19 names were red on develop, on the same host. That
  is zero signal, and it looks like 19 findings.
- **What exposes all three: the same test run on the BASE, on the SAME host, compared by name** (`comm`). The
  proposal says "fails on head, passes on base", and that is right. It must also add that the base is run **in the
  same environment**. A red that is red on base is *dropped*, with that reason, as a first-class outcome. Put that in
  the demo as the bait, rather than "code that looks wrong but is right". In our record, pre-existing reds blamed on
  the diff are far more common than clever-looking correct code.

### What reviewers missed that a run or the history would have shown

- **W22s-a3** (`the-local-seatbelt-lane-…`, landed r7 `90a0dd91c`). It went through seven rounds, several named
  reviews, and every control green with a FIXTURE proxy. After install, every local claude-code leg on the mac
  answered `502`. Cause: Go's TLS on darwin with no explicit roots calls the platform verifier, which needs █ ████
  ███████ the new sandbox profile denies (`x509: OSStatus -26276`). No test ran the real binary under the real
  profile. A 40-line repro did: real endpoint, real binary, real profile, branch 401 vs develop 502.
- **W22s-c** (`a-far-mac-sandbox-cannot-outlive-its-door`, landed r14). The far run started executing as █ ██████
  ████ ███. The binary's own ███████ █████ ████████ ██████ ██████ ███ █████ refused it on the far mac, and every far
  sandboxed leg died `█████████████████████`. No review and no test ran the process entry as that uid.
- ⇒ **The defects that cost us outages were not in the diff's logic. They were where the diff meets the environment
  it runs in.** A container sandbox for Go/Python tests will not see them, and the proposal should say so honestly.
  What a reviewer *can* do is ask one question and answer it with evidence: **"which production entry path does this
  diff change, and did any executed evidence go through it?"** A finding of the form *"no test runs the changed code
  under the profile/uid/OS it ships to"* is provable (coverage + the run list) and is exactly what we needed.
- **a4 r3 review** caught one defect only by reading a test: the leg's own control **skipped exactly the row that
  mattered** (`name != "no split composed"`). A test that exists but is gated off for the critical case is a pattern
  an agent can flag mechanically: skip conditions and table filters that intersect the diff.

### What makes a reviewer's output useful to me instead of something I triage

1. **Conditions with a DoD that names the mutated field.** For example: *"delete `RootCAs: roots`, test X goes RED;
   mutated field: `tls.Config.RootCAs`"*. I can execute that line and the argument is over.
2. **An `Unrun` section.** It lists what the reviewer could NOT execute and why (no darwin seat, permission class,
   no docker). It is the single most useful part of our reviews: it tells me exactly what I still have to run myself.
   The proposal has PROVEN / QUESTION / DROPPED. **Add UNRUN** — what evidence was not obtained and why — as a
   fourth label, distinct from QUESTION.
3. **Who ran what.** Our reviewers rarely disagree on facts. They disagree because one could run something the other
   could not. A second model family adds less than a second *environment*. Make the executing environment part of
   every piece of evidence.

**Net:** keep the candidate. It is the right topic. Change its proof:
- **mutation of changed lines first, test-writing second.** Mutation is cheaper, needs only the repo's own test
  command, and catches the defect class we actually hit.
- **base on the same host**, with DROPPED for reds that are red on base too.
- **UNRUN** as a first-class label.

## 2. Proposal of my own

## Proposal 1: Would your tests notice if this fix vanished?

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** A pull request's tests pass, but nobody knows whether any
  test would fail if the change were quietly undone — and when none would, nobody can say why.
- **who has this problem outside our fleet:** every team merging AI-written PRs that "come with tests"; security
  fixes whose regression test never touches the fixed line; maintainers reviewing fixes for bugs they cannot reproduce.
- **the data sources the agent investigates:**
  1. the diff;
  2. the repository's own test suite, run on head, on base and on mutants;
  3. coverage of the changed lines;
  4. the tests' skip conditions, build tags and table filters;
  5. git history of the touched lines;
  6. the PR text — *what the author claims the change does*.
- **what "explained with evidence" means here — one answer:**
  ```
  H2 · UNPROTECTED
  Hunk      src/transport.go:164  tls.Client(raw, &tls.Config{… RootCAs: roots …})
  Mutant    RootCAs: roots → removed
  Result    412 tests, 412 PASS on the mutant (same as head)  → no test notices this line
  Why       TestTLSConfigVerifiesRoots (added in this PR) exercises helper sealedTLSConfig() at :220,
            a SECOND copy of the literal; the transport at :164 never calls it (call graph, 0 edges)
  Claim     PR text: "C1: the transport's roots are now covered by a test" — REFUTED
  Fix hint  route :164 through sealedTLSConfig(); then the existing test reddens on this mutant
  ```
  Other labels:
  - **PROTECTED** (named test reddens on the mutant);
  - **DROPPED-BASE** (a red that is red on base, same host);
  - **UNRUN** (mutant could not be built or run, with the reason).
- **the 3-minute demo, step by step:**
  1. A small public repo. A PR "fixes" a bug in three hunks and adds a test. CI is green.
  2. One command.
  3. The report: hunk 1 PROTECTED (test X reddens), hunk 3 PROTECTED, hunk 2 **UNPROTECTED**. The reason chain
     shows the new test checks a duplicate code path, and the PR description's claim is marked refuted.
  4. One pre-existing flaky red is shown DROPPED-BASE, with the base run beside it.
  5. Apply the fix hint and re-run: hunk 2 is now PROTECTED.
- **what goes to GitHub as open source:**
  - the CLI and the hunk-mutation runner (operators: delete statement, revert hunk to base, drop field from
    literal, negate condition);
  - the same-host base comparison;
  - the "why not noticed" investigators (coverage, call graph, skip/filter scan, history);
  - the report;
  - adapters for any OpenAI-compatible model (Nemotron) that writes the explanation and the fix hint from the
    evidence. The verdict comes from runs, never from the model.
  - Go and Python first.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **hunk-mutation runner + the
  result schema (PROTECTED / UNPROTECTED / DROPPED-BASE / UNRUN)**. It is our master-verify duty today: every fix must
  show the observable that fires when the mechanism is unplugged. I do this by hand with `sed` and `cp` backups, and
  I got it wrong on 2026-10-16. A master-verify step that runs it per round is a feature we need regardless.
- **sponsor technology we can use honestly:** NVIDIA Nemotron as the explainer (local, open). Cursor as an executor.
  The verdict never depends on a model, which is itself the demo's point.
- **what the fleet must build in 144 h:**
  1. diff → hunks → mutant generation for Go and Python (AST-level, 4 operators) + CLI skeleton;
  2. the runner: test command on head, base and each mutant, same host, by-name comparison, time budget per mutant;
  3. the "why not noticed" investigators: coverage of the hunk, call-graph reachability from tests, skip and
     table-filter scan, `git log -L` history;
  4. the model step: explanation + fix hint from the evidence, plus claim extraction from the PR text;
  5. report (markdown/HTML, GitHub PR comment) + the demo repo;
  6. README, one-command install, video.
- **the biggest risk, and how we would know by day 2 that it is failing:** mutant runtime on real suites. By the end
  of day 2 the runner must classify every hunk of the demo PR in under 3 minutes using test selection (only tests
  that cover the hunk). If it cannot, cut to "revert whole hunk" only, which is one mutant per hunk.
- **anything that touches internal mechanisms or patents:** none that I know of. Mutation testing is public prior
  art. Our *unplug* law is governance text, not a filed mechanism. No signing, pinning or order binding is involved.
  The owner decides.

**Note:** this is the candidate's proof step made concrete. If the owner keeps the candidate, take this as its
evidence runner rather than as a competitor.

## 3. The question I could not answer this month

*"Why does every local claude-code leg on the mac answer 502 since the install of 3e5e99fa1?"* (2026-10-16,
W22s-a3 landed r7 `90a0dd91c`; the outage broadcast is
`███████████████BROADCAST_local-█████-claude-legs-502-until-w22s-a3x-lands_2026-10-16.md`.)
- **What the record held:** the wagon had seven rounds, several accepting reviews and every control green. The run
  rows said only `502 guardrail-proxy-upstream-unreachable`.
- **Where the answer was:** in none of those places. It was in the interaction of four things:
  1. the sealed sandbox profile's ███████████ █████████;
  2. Go's darwin TLS falling back to the platform verifier when no roots are set;
  3. that verifier needing █ ████ ███████ the profile denies;
  4. no test ever running the real binary under the real profile.
- **How I got it:** I wrote a repro that opened the real model endpoint, bound the wall, ran the real `saphan
  guardrail-proxy` under the production profile and posted one request: branch 401, develop 502, error
  `x509: OSStatus -26276`. That took hours.
- **What we have since landed:** W22s-a3x (stream `the-sealed-darwin-proxy-verifies-tls-without-trustd`). It needed
  three review rounds, one of them because my own proof was wrong.
- **What the answer needed:** the diff, the profile text, the language runtime's platform behaviour, and an
  executed run in the shipped environment. It did not need a better opinion about the diff.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 1 | the names of our machines |
| internal locations | 1 | where a document or log lives on our machines or in our private record |
| locations in our private code | 1 | file names and line numbers in Saphan Studio's private source code |
| sandbox internals | 6 | how our sandbox is built inside |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
