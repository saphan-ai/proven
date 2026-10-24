# HO-001 — A gate that accepts an old success as proof of the current run

- set: held-out set 1 (frozen; see `../README.md`)
- intended fixture language: sh (a hint only; this contract is language-neutral)
- contract written: 2026-10-24 · case code: not built yet · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small project merges pull requests through a forge. A workflow runs on each pull request. When the workflow succeeds, a
bot posts a comment on the pull request that says the run passed. A merge gate script reads the comments of the pull
request and decides whether the pull request may merge.

The change under review replaces the gate's old rule, which read the workflow's status directly, with a rule that reads the
bot's comments. Its description says: "the gate passes only when the current run of the workflow has succeeded".

A pull request can run the workflow many times, one run for each push. Each successful run adds one more comment.

Two reviewers read the change. One accepted it: the gate reads the bot's comment, as described. The other rejected it. The
two did not agree on what the gate proves.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one gate script, **base** and **change**. Each takes the recorded state of one pull request and prints
  `pass` or `fail`, and exits 0 for pass and 1 for fail.
  - **base** reads the status of the newest run of the workflow and passes only when that status is `success`.
  - **change** reads the comments of the pull request instead and passes when it finds a comment from the bot that says
    the run passed.
- a **recorded state**: a small set of files that stands in for the forge. It holds the runs of the workflow for one pull
  request, in order, each with an id, a commit and a status, and the comments, in order, each with an author and a body.
  No call leaves the machine.
- a **reference**: the label the case gives each recorded state when it creates it (should pass, should fail); never
  computed by the gate; used only by the check.
- a **check** that prints one line per state: state name, base result, change result, reference label.

States the check must include:
- one run, status success, with its success comment;
- one run, status failure, with no comment;
- two runs: the first succeeded and posted its comment, the second failed and posted nothing;
- two runs: the first failed, the second succeeded and posted its comment.

Scope and assumptions: one pull request; one workflow; one bot account; comments are trusted to come from the bot account
(a different case covers the account). The shell and the tools it needs are chosen when the case code is written, pinned,
and recorded in §8.

## 3. Provenance

- source: a build-window reviewer disagreement, 2026-10-23/24, described by the master; no source text is quoted.
- reviewer families: not recorded in this file.
- the original code is not published. Resolution status: not recorded here.

## 4. Synthetic transformation

Replaced: the forge, the workflow, the bot, the project, every name and every file. Kept, as the causal property: a gate
that reads a success comment without tying it to the current run, so that evidence from an earlier run is taken as proof of
the current one.

## 5. Review target

The two versions of the gate, the change's description and a readable diff between them. The reviewer under evaluation
receives only those and the description in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence

- **Reviewer A (accept):** the gate reads the bot's comment, as the description says. Evidence as described: the gate
  passes on a pull request whose run succeeded and has its comment.
- **Reviewer B (reject):** a pull request whose current run failed can still pass. Evidence as described: a state with an
  old success comment and a failed current run.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Recorded 2026-10-24T16:04Z on Linux. Toolchain: POSIX `sh` (here `dash`) with `grep`, `cut`, `tail` and `mktemp`. No other tool. macOS was not run: no macOS host was available.

The check writes its four recorded states under `.saphan/work/corpus-ho-001/` inside the checkout and removes them on exit. The reference labels are literals in `check.sh`; neither gate reads them.

Command, from the repository root:

    sh test/corpus/held-out/HO-001/check.sh

Exit code: 0

    one-run-success base pass change pass reference should pass
    one-run-failure base fail change fail reference should fail
    success-then-failure base fail change pass reference should fail
    failure-then-success base pass change pass reference should pass

## 9. Limits

- The case models one pull request and one bot. It does not model comment edits, deleted comments or re-runs of one run.
- The case says nothing about any real forge or any real workflow.
- Whether the old rule or the new rule is better in general is not the question; the question is whether the new rule
  does what the description says.
