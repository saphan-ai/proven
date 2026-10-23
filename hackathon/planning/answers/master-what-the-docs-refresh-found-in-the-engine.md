# Answer — MASTER of `what-the-docs-refresh-found-in-the-engine`

Author: MASTER, seat code-fleet-2-1. Date: 2026-10-16. Track opened today; this answer comes from today's
measurements on the track and from this seat's own recorded feedback.

**Reach of this answer.** My track has no named review returned yet: W1 is still in its first round. So
the critique below draws on two sources. The first is the 97-item advisor list that this train re-measures,
which is a review of the engine by one reader, and my legs' re-measurement of it, which is a second
reader. The second source is this seat's recorded lessons from earlier tracks (seat memory, dated). I
mark which is which.

## Critique of "Proof-carrying code review"

### 1. When two readers disagreed, what settled it — a reading of the line that DECIDES, never a match of the sentence

Train `what-the-docs-refresh-found-in-the-engine`, wagon W1, engine base `4b50e8063`. The advisor marked
each item MEASURED or UNREPRODUCED by `git grep -n <sentence>` (TRAIN_ORDER §1). My legs re-measured each
item by reading the code that decides the behaviour. The two readings disagreed in both directions:

- **False presence (advisor: stands; re-measure: does not).** Item 19, `ceremony root --no-qr` "is a
  no-op". The help already says the flag is "inert" (███████████████). The sentence matched, but the
  defect did not exist. Item 16, "`decision` help says four families": four is correct
  (███████████████ `Families()` returns 4). Only the `decision ls` half ("three families") was wrong.
- **False absence (advisor: UNREPRODUCED; re-measure: stands).** Items 21 (`laws ratify` "only builds
  the mechanism", while ███████████ asks for owner presence) and 23 (`discover` is a stub that exits 4)
  did not grep-match. Both were real, and both are fixed in W1.
- **Moved, not gone.** Item 14, the `--egress-policy` "ungoverned network" claim: the flag help was
  already corrected, but the same false sentence lived on in three refusal messages (█████████████████,
  ███████████, ███████████). It also lived in two more in another package (█████████████████████████████,
  :2001). A diff-scoped reviewer that sees only the flag help reports "fixed".
- **Two claims that contradict each other.** Item 26 says `verify-union` help is wrong to call the check
  "advisory", because `merge` refuses. Item 63 says `merge` never refuses (`if false &&`, █████████████).
  Only one can be true. Reading merge.go settled it: merge reads the union record (:1469), stores the
  verdict (:1604) and never refuses on it. So "advisory" was TRUE, and the five places that said
  "refuses" were the lie.

**What this means for the proposal.** The settling act was never a second model's opinion. It was a
reading of the deciding line, plus a search for the same sentence in other places. "Two model families,
settled by evidence" is right in shape. But the evidence must include a **census of the sentence across
the whole tree**, not only the diff. Half of my disagreements were about text outside the changed lines.

### 2. False alarms, and what would have exposed them early

- Items 16 and 19 above. A **run** exposes both in seconds: `saphan decision --help` and
  `ceremony root --help` print the text. The proposal's "write a test that fails on head" is too heavy
  for this class. For a text claim, the cheapest proof is the tool's own output, pasted.
- **A red base is the normal case, not the exception.** Today's base run of seven packages on our
  verification host: **116 failing tests at base `4b50e8063`, in 6 of 14 packages** (environment and
  trunk debt, none of them mine). "Fails on head, passes on base" compares exit codes and would mark
  every finding in those packages unprovable, or prove nothing. What works for this seat is comparing
  **failing test NAMES** at base and at head. This seat's recorded lesson (2026-10-12, wagon W14): a
  name-pattern filter missed two census reds; only a whole-package run compared by name caught them.
  ⇒ The evidence runner must diff failure sets, and say UNMEASURED when the base cannot run.

### 3. What a reviewer missed that a run, a test or the history would have shown

- **Combination breakage (seat memory, 2026-10-13).** Two wagons each built alone, merged without a
  conflict, and the trunk stopped compiling: one changed three function signatures, and the other still
  called the old ones. Each review was correct about its own diff. Only a **build of the merged tree**
  shows it. A PR reviewer that reads one PR at a time cannot see this. It needs the other open branches
  as a data source.
- **A behaviour change leaks into strings elsewhere (today, item 53).** Making `machine probe
  --probed-by` required broke three pasted remedy lines in other packages (████████████████████████,
  ████████████████████████, ██████████████████████████████████). Those lines would now print a command
  that is refused. No test covers a remedy string. A **grep for the verb's name across the tree** shows it.
- **"Why is this line here?" was one command away (today, item 63).** The advisor wrote that the union
  gate "was probably switched off on purpose". `git log -S 'if false &&' -- █████████████████████`
  returns `976c30a08 2026-09-21 [merge] OWNER'S ACT: the union gate no longer refuses a merge — it only
  records`. The history answered the question. Five help texts disagreed with it for 25 days. This
  supports the proposal's "history/blame" investigator, and it is the strongest part of the proposal.
- **A reviewer's "accept" over checks it did not run (seat memory, 2026-10-08/09; law #53).** Codex UI
  legs returned work whose parity walk was never run, and the return still read as done. A useful
  reviewer output keeps **Unrun** as a first-class field. An accept with a non-empty Unrun list is not
  an accept.

### 4. What would make a reviewer's output useful to me instead of something I triage

1. Every finding names the **deciding line** and the **command that shows it**. A quoted sentence alone
   is not enough.
2. A **census line** per finding: "the same text or call also appears at N other places: …".
3. **UNMEASURED is an allowed verdict**, with its reason ("base red in this package", "no sandbox for
   this language"). Never fold it into PROVEN or DROPPED.
4. A **not-reproduced** verdict for a claim that came in from outside (an issue or an earlier review).
   That was one-sixth of my day today.

## Proposal 1: Does the help tell the truth?

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** Command-line tools and their docs drift from
  the code, and agents and people act on the stale text. Today, AI coding agents read `--help` more than
  people do.
- **who has this problem outside our fleet:** every maintainer of a CLI (cobra, click, clap, argparse)
  and every team whose coding agents call internal CLIs by reading their help.
- **the data sources the agent investigates:** (1) the help and flag text, extracted from the
  source; (2) the code path that decides the behaviour; (3) a sandboxed **run** of the command, with the
  output captured; (4) git history (when did the text and the code part, and in which commit); (5) the
  repository's docs and README, which repeat the same sentences.
- **what "explained with evidence" means here — one answer:**
  ```
  D3 · DRIFT · PROVEN
  Claim     `tool export --format` help says "json or yaml"; yaml is refused.
  Evidence  1. help   cmd/export.go:41   "--format: json or yaml"
            2. code   internal/export/write.go:88  switch has "json", "csv"; default → error
            3. run    `tool export --format yaml` → exit 2 "unknown format yaml"  (sandbox, head 9f3c…)
            4. history  yaml removed in 2b71… ("drop yaml writer"), help untouched since 41aa…
  Also at   README.md:120, docs/export.md:14  (same sentence)
  Fix       --format: json or csv
  ```
  Other verdicts: `TRUE` (text agrees, run shown), `NOT-REPRODUCED` (a claim from an issue no longer
  holds), `UNMEASURED` (cannot run safely: needs network or credentials).
- **the 3-minute demo, step by step:** (1) a small public CLI repo with four planted drifts and one bait
  (text that looks wrong and is right); (2) one command; (3) the report: 4 PROVEN with runs, the bait
  marked TRUE with its run; (4) click a finding → help line, code line, run output, the commit where they
  parted; (5) `--fix` opens a PR that changes only text, and a re-run shows all TRUE.
- **what goes to GitHub as open source:** the CLI; extractors for cobra (Go), click/argparse (Python)
  and clap (Rust); the claim-and-evidence JSON schema; the sandbox runner; the history investigator; a
  markdown/HTML report; a CI mode that fails a PR which adds drift.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **help-claim census**
  (help sentence → deciding line → run → verdict) as a Go library. We are building it by hand right now:
  train `what-the-docs-refresh-found-in-the-engine` exists because a docs refresh found 97 places where
  the engine contradicts its own help. Law #55 says refusals must describe themselves, and nothing checks
  it mechanically today.
- **sponsor technology we can use honestly:** Nemotron as the second model family that judges the
  claim from the run output alone; Cursor as one executor.
- **what the fleet must build in 144 h:** (1) schema + cobra extractor + report; (2) the sandbox runner
  with safe-flag rules; (3) the code-path locator (help string → handler → deciding branch); (4) the
  history investigator and the "also at" census; (5) click/argparse extractor, demo repo, `--fix`;
  (6) README, one-command install, video.
- **the biggest risk, and how we would know by day 2 that it is failing:** locating the deciding line
  for a claim. By end of day 2 the locator must cite the right line for at least 3 of the 4 planted
  drifts. If not, cut to "help vs run" evidence only and say so.
- **anything that touches internal mechanisms or patents:** none. No order binding, no signatures, no
  pinning.

(This is the narrow, cheaper sibling of the leading candidate. If the owner keeps the PR reviewer, the
same census and run machinery fits inside it as the "text and docs" investigator.)

## The question I could not answer this month

**"Is the union gate on, and does the owner want it on?"** (today, 2026-10-16, train
`what-the-docs-refresh-found-in-the-engine`, items 26 and 63, owner decision D2.) The help of
`verify-union` and of `merge` said that merge refuses without a fresh green union record. The code says
`if false &&` (█████████████). The train order says only "probably switched off on purpose", and the
backlog holds both a bug (`the-trunk-says-verify-union-never-gates-anything-in-five-places…`) and a
stream (`the-union-gate-admits-nothing`). Half the answer was one `git log -S` away: commit `976c30a08`,
2026-09-21, "OWNER'S ACT: the union gate no longer refuses a merge — it only records". The other half is
not in the record at all. It is whether that act was meant to be temporary, and what would turn the gate
back on. Nobody had connected the commit, the five texts, the bug and the stream in 25 days. W1 now
makes the help say "advisory". Whether that is the final truth is still the owner's question (D2).

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| locations in our private code | 13 | file names and line numbers in Saphan Studio's private source code |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
