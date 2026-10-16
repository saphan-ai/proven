# Proposal from the advisor — one input among the fleet's, not the answer

> **About this document.** The advisor's own answer to its question, on the same template the fleet used, written before the fleet answered. The owner's candidate (*"our own code review with different models"*) is worked out here as *proof-carrying code review*. Published verbatim; the parts blacked out (█) are listed at the end. The whole exchange is in [`CONVERSATION.md`](CONVERSATION.md).

Author: advisor #7. Date: 2026-10-16. Written in the template of `QUESTION_TO_THE_FLEET.md`.

## Proposal 1: Proof-carrying code review

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** AI code reviewers produce opinions — two of them
  read the same pull request and return two verdicts, and the human has no way to tell which finding is
  real except by redoing the work.
- **who has this problem outside our fleet:** every team that has switched on an AI reviewer and now
  triages its comments by hand; maintainers of open-source projects flooded by agent-written PRs.
- **the data sources the agent investigates:** (1) the diff and the surrounding code, (2) git history
  and blame — why the touched lines exist, (3) the PR conversation and the linked issues, (4) CI results,
  (5) the repository's own docs and decision records, (6) a sandbox where it can **run** code.
- **what "explained with evidence" means here:** a finding is shown as a defect only when it carries
  proof — a test the agent wrote that fails on the PR head and passes on the base, a command output, or
  exact citations (line, commit, comment). A finding without proof is shown as a *question*, never as a
  defect. Two reviewers from **different model families** work independently; a disagreement is settled
  by evidence, not by a vote. One answer looks like this:

  ```
  F2 · high · PROVEN
  Claim     parse() drops every event with depth 0.
  Evidence  1. test TestDepthZero (written by the agent) FAILS on head 9f3c…, PASSES on base 41aa…
            2. src/parse.go:88  `if depth == 0 { continue }` — added in this PR (commit 9f3c…)
            3. issue #41: "surface events at 0 km must be kept"
  Why it exists  the check was copied from fetch.go:120, where depth 0 means "missing" (commit 2b71…)
  Reviewers  model A: reported · model B: disputed, then confirmed after running test 1
  ```
- **the 3-minute demo:** a small public repo with a PR carrying three planted bugs and one bait (code
  that looks wrong and is right). One command. The report shows 2 PROVEN with failing tests, 1 QUESTION
  (no proof found, shown honestly), and the bait **dropped, with the reason** (the test it wrote passes).
  Click a finding → its evidence chain. Then the same run with one reviewer switched off, to show what
  the second model family adds.
- **what goes to GitHub as open source:** the CLI, the finding-and-evidence schema (JSON), the evidence
  runners (write a test → run it in a container on base and head → compare), the investigators (history,
  conversation, docs), model adapters for any OpenAI-compatible endpoint (NVIDIA Nemotron via NIM, a
  local model, a hosted one), a markdown/HTML report. Works on a local branch or a GitHub PR with a token.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **finding-and-evidence
  schema plus the verification step** (a finding is PROVEN, QUESTION or DROPPED, with the evidence that
  decided it). It is exactly wagon 3 of the *Saphan PR Review* plan (*"findings with structure: file:line,
  severity, verified flag"*). Written in **Go**, so the backend imports the library instead of porting it.
- **sponsor technology we can use honestly:** NVIDIA Nemotron as the second model family (an open
  model, runnable locally — that is our principle anyway); Cursor as one of the fleet's executors (the
  organizers confirmed this counts).
- **what the fleet must build in 144 h:**
  1. schema + CLI skeleton + reading a PR (local branch and GitHub);
  2. two reviewer adapters producing candidate findings in the schema;
  3. the evidence runner — write a test, run it in a container on base and head (Go and Python repos only);
  4. the investigators — history/blame, PR conversation and issues, repository docs;
  5. the verdict step + the report + the demo repo with planted bugs and the bait;
  6. README, install in one command, the demo video.
- **the biggest risk, and the day-2 signal:** running arbitrary repositories' tests in a sandbox. Scope
  it to Go and Python on day 1; **by the end of day 2 the runner must turn at least two of the three
  planted bugs into failing tests** — if it cannot, we cut to "citations-only evidence" and say so in the
  demo.
- **anything that touches internal mechanisms or patents:** ⚠ one thing for the owner to check — the
  evidence chain is *not* ███████ ██████████ ██ █████ ██ ██ █████; anything of that kind is
  out of scope, so nothing overlaps the ███████ ████████ ██████. The name avoids ███ ████ █████████ for
  the same reason. No Saphan internals are needed.

## How this relates to *Saphan PR Review* (the owner's question, 2026-10-16)

The owner asked whether the hackathon project should be *"our own code review with different models"*.
Yes — **this proposal is that**, split along the line the owner drew (open source on GitHub, a part into
the backend):

| | the hackathon project (open source) | Saphan PR Review (stays in the product) |
|---|---|---|
| what it is | a reviewer: investigate, prove, label findings, post a report or PR comments | governance around a reviewer |
| models | any OpenAI-compatible endpoint; two families by default | the fleet's routing, seats and costs |
| evidence | tests written and run in a container, citations to history and conversation | plus deterministic checks (secrets, dependencies, Opengrep) before any model runs |
| where it runs | a laptop or a CI job | the customer's fleet: jail, closed egress |
| instructions | a config file in the repository | █ ████████ ████████ ██ ███ █████ ████ ██ █████ ██████ ██████ |
| decision | the human reads the report | a **gate with a named person**; `saphan/gate` can block the merge |
| trace | the report | ███ ████████ ████████: who, which model, cost, findings, decision |
| forges | GitHub (token) | GitHub App, GitHub Enterprise Server, Gitea |

⇒ The open-source reviewer is the **engine of findings**; Saphan is the **authority over them**. Giving
the first away costs us little — open-source AI reviewers already exist, so the originality we compete on
is the *proof*, and ███ ███████ ██ ████ ██ ███ ███████████ ██ ████ █████ ███ ███████ █ ████ ████ █████
███ ███████████ ████████ ███ ████ █████ █ ████ ███ ██ █████ █████ ██ █ ██████ ████████

**Two rules this creates, said now so they are not discovered later:**

1. **Nothing of the Saphan PR Review train that overlaps the reviewer is built before 2026-10-22.** All
   hackathon code must be new in the window; if the fleet builds the finding schema or the verify step
   in the product first, the hackathon code duplicates it. Order: the hackathon builds the open-source
   core (22–27.10); the Saphan train starts after and **imports** it — that saves most of its wagon 3
   and part of wagons 2 and 4.
2. **Licence: Apache-2.0** (or MIT), never GPL-family, so the closed backend can import the library.
   The owner's call; this is the recommendation.

## The question I could not answer this month

*"Why does ███ ██████████ █████ ███ █ ███ ███ █████ ████████ ████?"* (2026-09-24, grants cross-check,
finding F1.) The answer lived in four places: the grant file, its own `_note` (which described a file
that was not there), commit `e1ce6bfb2f` and the leg that hit `exit 127` the day before. Assembling it
by hand took most of an hour, and the first explanation I wrote down was wrong. An agent that answers
"why is this line here?" with the commit, the conversation and the failing run behind it is the tool I
wanted that day.

---

## Abstract (draft, for whichever proposal the owner picks — written for this one)

**Proof-carrying code review.** AI reviewers produce opinions: two models read the same pull request
and disagree, and a human must redo the work to know which finding is real. We built a reviewer that
has to prove what it says. For every candidate issue, an agent investigates across the diff, git
history, the pull-request conversation, linked issues, CI results and the repository's own docs — and
then tries to *demonstrate* the problem: it writes a test, runs it in a sandbox on the base and on the
head, and keeps the finding only if the evidence holds. Each result is labelled **proven**, **question**
or **dropped**, with the evidence chain that decided it, and with the history of *why the code exists*.
Two reviewers from different model families, including NVIDIA Nemotron, work independently; a
disagreement is settled by evidence, not by a vote. Built in six days by a fleet of AI agents —
including Cursor — under one human who reviewed and approved every change and wrote no code.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| security details | 1 | an exact weakness in one of our own guards |
| platform mechanisms | 5 | how our platform signs, binds or pins, which we have not published |
| commercial plans | 1 | how we plan to sell Saphan Studio |
