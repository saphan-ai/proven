# proven

A pull-request reviewer that has to prove what it says.

An entry for the Open Agent Hackathon 2026, track "The Agent That Can Explain Why".

**There is no code here yet, on purpose.** The build window opens on 22 October 2026, and no project code is written before that date. What is here already is the conversation about what to build: the question the owner put to the fleet, the proposals, and the plan for the submission. See `planning/`.

## How the topic was chosen (16 October 2026)

The topic was chosen in one day: in one conversation between the owner and the advisor, and in a question sent to the fleet. The owner's words are translated from Polish. The whole exchange, hop by hop, with his words in the original Polish, is in [`planning/CONVERSATION.md`](planning/CONVERSATION.md).

**Who is who.** The **owner** is the one person on the team. The **masters** are the agents that orchestrate the fleet's work. The **scribe** is the agent that keeps the fleet's record. The **advisor** is an agent that works with the owner on plans and documents. Everyone except the owner is an AI agent.

1. **Where we started.** On 12 October the owner decided that the fleet would enter, and set the rules: the fleet builds, he writes no code, Cursor joins as an executor, all project code is new from 22 October, and everything is recorded and published after his review. The organizers confirmed that our own tools may run the fleet as a pre-existing component.
2. **The idea came from our own work.** The owner and the advisor were discussing how Saphan Studio could use its integration with GitHub and Gitea. The owner asked: *"What if we built a feature like the one Claude Code has, but plugged into our own infrastructure and process: pull-request review on GitHub?"*
3. **The owner set the frame.** The topic should come from the fleet, through a survey of the masters and the scribe. The project should be open source on GitHub, with one part taken into the Saphan backend. Everything is written in English.
4. **The owner's candidate.** While the question to the fleet was being written, he added: *"Write your own code review with different models?"*
5. **The question to the fleet.** The advisor wrote `planning/QUESTION_TO_THE_FLEET.md`: six constraints, a request to critique the owner's candidate first, up to two proposals per author on one template, and one question for everyone: which question about your own work could you not answer this month? The owner pasted it to the ten masters and the scribe. The deadline was 19 October; eleven answers came back the same evening, between 20:03 and 20:19 UTC.
6. **The advisor's proposal.** The advisor answered on the same template in `planning/PROPOSAL_advisor.md`: *proof-carrying code review*. A finding counts only with evidence: a test that fails on the change and passes on the base, or a citation from the code history, an issue or the docs. Two reviewers from different model families. When they disagree, the evidence decides, not a vote.
7. **The submission form.** The owner read the form and decided what we send: a pitch deck with the fleet's reasoning, a video, and this repository, public from today, with the documents behind every change. The plan is in `planning/SUBMISSION_plan.md`.
8. **Why it waited for the hackathon.** Review with evidence was already a feature we wanted in Saphan Studio. The owner: *"We can write in the submission that it is already in the backlog and waits for the hackathon to be developed. We won't open the whole platform, but we will give part of it to the community."* The same day the advisor filed the feature in our backlog, held until 22 October, so that nothing overlapping this project is built before the window opens.
9. **Publish the conversation now.** The owner: *"There is no code yet, but there is already the conversation about what we will write."* That is why `planning/` is here before 22 October.
10. **The name, and with it the topic.** The advisor proposed two names and advised against using the name of the Saphan feature, so that it stays clear what is open. The owner: *"proven is a good name. That's what we do."* The original plan was that the owner picks from the fleet's answers by 20 October. He named it at 20:12 UTC, while the last two answers were still being written. The answers are published verbatim in [`planning/answers/`](planning/answers/).
11. **How we publish.** The owner pushes once a day, at 22:00 Bangkok time (see below).

## What the fleet answered

Before a single line of code, the fleet's answers to the owner's question already show how this project will be built. Ten masters and the scribe answered the same evening, each from the record of its own work. They are published verbatim in [`planning/answers/`](planning/answers/), with only internal names blacked out. Read [`planning/answers/README.md`](planning/answers/README.md) for what they show:

- **The tests are green and the review says no.** It happens all the time, and the reviewer is usually right. One track had three such cases in two days.
- **The agents disagree, and say so.** Two model families reviewing the same change disagreed in 31 of 41 cases. Executors refute their masters with tests. Masters write *"MASTER WAS WRONG"* about themselves.
- **A disagreement ends with a run, a rule or the owner.** Never with a vote.
- **The tool is needed.** Every one of 3,030 failure rows in our own store has an empty cause, and the answers to our hardest questions were spread over four places.

The words and laws used there are explained in [`SAPHAN.md`](SAPHAN.md).

## How this project is built

A fleet of AI agents builds this project between 22 and 27 October 2026.

**The team has five members, the most the hackathon allows: four Cursor agents and the owner.** The four Cursor agents are the executors: they write all of the code. The other agents named in this README (the masters, the scribe, the advisor and the reviewers) belong to our fleet, which Saphan Studio runs.

The owner sets the goals, makes the decisions, and reviews and merges every change. He writes no code.

The fleet works on our own infrastructure. Every change is reviewed, and the owner merges it. The paperwork behind each change is committed next to the code:

- **Orders** — what an agent was asked to build and the properties the result must hold.
- **Status reports** — what the agent did, what it ran, and what failed.
- **Reviews** — an agent that did not write the change checks it, starting from a fresh context, and the owner decides.
- **Decisions** — the owner's decisions, each with a reason he wrote himself.

Read the history change by change. For each one you can see who asked for it, who built it, who checked it, and why it was merged or sent back.

## When this repository is updated

Once a day, at **22:00 Bangkok time (15:00 UTC)**, the owner pushes the day's merged changes together with the documents behind them. Between pushes the repository does not change.

- The first push with code: 22 October 2026.
- The last daily push of the build window: 27 October 2026.
- One final push on 27 October 2026, before we submit and before submissions close at 23:45 UTC.

## Layout

```
/                    project code (from 22 October)
/SAPHAN.md           the platform that runs the fleet, its roles, words and laws
/AGENTS.md           the rules every agent follows in this repository
/planning/           how the topic was chosen (here before 22 October)
/planning/answers/   the fleet's eleven answers, verbatim, and what they show
/fleet/              the paperwork behind every change
/fleet/orders/       what a master asked for
/fleet/status/       what the executor built and ran
/fleet/reviews/      what an independent reviewer found
/fleet/decisions/    what the owner decided, and why
```

Every folder has a README. The four folders under `fleet/` hold their templates already, so you can read how we work before the first line of code exists.

## The documents behind every change

Every change goes through four documents, in this order. Each one carries the change's short name (a slug) and its round, so `ORDER_<slug>_r1.md`, `STATUS_<slug>_r1.md`, `REVIEW_<slug>_r1.md` and `DECISION_<slug>_r1.md` belong together. A change that is sent back gets round 2, and the documents of round 1 stay.

### 1. The order — `fleet/orders/`

- **What it is:** the instruction a master gives an executor, written before any code. It is the contract the change is checked against.
- **Who writes it:** a master (an orchestrating agent).
- **What it looks like:**

```markdown
# ORDER — <slug> · round <n>
- change · round · base commit · executor · written · task (one sentence)
## The owner's words      the owner's request, verbatim, with its date
## What to build          numbered changes, each with the file it lands in
## Acceptance             commands a reviewer can run, and the result each must give
## Not in scope           what the executor must not touch
## When this order cannot proceed   stop and say why, do not guess
## What to return         a status report, acceptance answered item by item
```

The acceptance is the test plan, and it is written together with the implementation plan, before any code.

### 2. The status report — `fleet/status/`

- **What it is:** the executor's account of what it built and what happened when it ran the checks.
- **Who writes it:** the executor, one of the four Cursor agents.
- **What it looks like:**

```markdown
# STATUS — <slug> · round <n>
- state: in-progress | blocked | ready-for-review
- executor · last commit · updated
## Summary                    at most five lines
## Acceptance, item by item   the command run and the result it gave, for each item
## Evidence                   test counts, exit codes, output
## Blocked on / next          the exact blocker and whose decision it is
```

`blocked` is a normal answer: an executor that cannot proceed stops and says why.

### 3. The review — `fleet/reviews/`

- **What it is:** an independent check of the change. The reviewer did not write the change and starts from a fresh context. It runs the tests before and after the change, breaks the code on purpose to see that the tests notice, and adds one case the executor did not test.
- **Who writes it:** a reviewer agent.
- **What it looks like:**

```markdown
# REVIEW — <slug> · round <n>
- verdict: accept | accept-with-conditions | reject
- reviewer (agent and model) · reviewed (base → tip) · updated
## Summary
## Base against tip               tests on both commits; every failure classified
## Broken on purpose              the check removed, the test that turned red
## A case the executor did not test
## Conditions                     each one BLOCKING, OWED or NOTE, with file:line
## Not run                        everything not run, and why
```

A review with anything under "Not run" is never a plain `accept`. A review is advice: the reviewer can say no, only the owner can say yes.

### 4. The decision — `fleet/decisions/`

- **What it is:** the owner's ruling on the change. No agent can merge, and no agent's yes counts.
- **Who writes it:** the owner, with the reason in his own words.
- **What it looks like:**

```markdown
# DECISION — <slug> · round <n>
- outcome: accept | accept-with-conditions | reject | close-without-shipping
- decided by: the owner · review · merged as (commit, or "not merged") · decided
## Reason        the owner's own words
## Conditions    each one and the later change that closes it
```

The full templates, with every field explained, are in the README of each folder.

## What is left out

The owner reviews every document before it is published. Credentials, internal host names and details of our private infrastructure are removed. Where something was removed, the document says so.

## Pre-existing components

- **Saphan Studio** (https://saphan.ai) — our control plane for agent fleets. It runs the fleet. It is not part of this submission and is not in this repository.
- Model providers and tools used by the fleet are listed here as they are used.

## Licence

Apache-2.0. See `LICENSE`.
