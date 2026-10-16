# Proposal from the advisor — one input among the fleet's, not the answer

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
- **anything that touches internal mechanisms:** none. No Saphan internals are needed.

## How this relates to Saphan (the owner's question, 2026-10-16)

The owner asked whether the hackathon project should be *"our own code review with different models"*.
Yes, **this proposal is that.** The reviewer is open source and stands on its own: it runs on a laptop
or in a CI job, with any OpenAI-compatible endpoint and a GitHub token.

Review with evidence was already on the Saphan Studio backlog. We hold it back so it is built here, in
public, during the window. Saphan will import the same library for its own pull-request review. The
rest of Saphan stays closed.

**Two rules this creates:**

1. **Nothing in Saphan that overlaps the reviewer is built before 2026-10-22.** All hackathon code must
   be new in the window. The hackathon builds the open-source core (22–27.10); Saphan imports it after.
2. **Licence: Apache-2.0**, so anyone, Saphan included, can use the library.

## The question I could not answer this month

One of our own configuration files allowed more than anyone meant it to, and the question was simply
*why*. The answer lived in four places: the file, a note inside it that described a file that was not
there, one commit, and a failed run the day before. Assembling it by hand took most of an hour, and the
first explanation I wrote down was wrong. An agent that answers
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
