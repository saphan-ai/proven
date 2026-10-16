# proven

A pull-request reviewer that has to prove what it says.

An entry for the Open Agent Hackathon 2026, track "The Agent That Can Explain Why".

**There is no code here yet, on purpose.** The build window opens on 22 October 2026, and no project code is written before that date. What is here already is the conversation about what to build: the question the owner put to the fleet, the proposals, and the plan for the submission. See `planning/`.

## How the topic was chosen (16 October 2026)

The topic was chosen in one day: in one conversation between the owner and the advisor, and in a question sent to the fleet. The owner's words are translated from Polish.

**Who is who.** The **owner** is the one person on the team. The **masters** are the agents that orchestrate the fleet's work. The **scribe** is the agent that keeps the fleet's record. The **advisor** is an agent that works with the owner on plans and documents. Everyone except the owner is an AI agent.

1. **Where we started.** On 12 October the owner decided that the fleet would enter, and set the rules: the fleet builds, he writes no code, Cursor joins as an executor, all project code is new from 22 October, and everything is recorded and published after his review. The organizers confirmed that our own tools may run the fleet as a pre-existing component.
2. **The idea came from our own work.** The owner and the advisor were discussing how Saphan Studio could use its integration with GitHub and Gitea. The owner asked: *"What if we built a feature like the one Claude Code has, but plugged into our own infrastructure and process: pull-request review on GitHub?"*
3. **The owner set the frame.** The topic should come from the fleet, through a survey of the masters and the scribe. The project should be open source on GitHub, with one part taken into the Saphan backend. Everything is written in English.
4. **The owner's candidate.** While the question to the fleet was being written, he added: *"Write your own code review with different models?"*
5. **The question to the fleet.** The advisor wrote `planning/QUESTION_TO_THE_FLEET.md`: six constraints, a request to critique the owner's candidate first, up to two proposals per author on one template, and one question for everyone: which question about your own work could you not answer this month? The owner sent it to the masters and the scribe. Answers are due on 19 October.
6. **The advisor's proposal.** The advisor answered on the same template in `planning/PROPOSAL_advisor.md`: *proof-carrying code review*. A finding counts only with evidence: a test that fails on the change and passes on the base, or a citation from the code history, an issue or the docs. Two reviewers from different model families. When they disagree, the evidence decides, not a vote.
7. **The submission form.** The owner read the form and decided what we send: a pitch deck with the fleet's reasoning, a video, and this repository, public from today, with the documents behind every change. The plan is in `planning/SUBMISSION_plan.md`.
8. **Why it waited for the hackathon.** Review with evidence was already a feature we wanted in Saphan Studio. The owner: *"We can write in the submission that it is already in the backlog and waits for the hackathon to be developed. We won't open the whole platform, but we will give part of it to the community."* The same day the advisor filed the feature in our backlog, held until 22 October, so that nothing overlapping this project is built before the window opens.
9. **Publish the conversation now.** The owner: *"There is no code yet, but there is already the conversation about what we will write."* That is why `planning/` is here before 22 October.
10. **The name, and with it the topic.** The advisor proposed two names and advised against using the name of the Saphan feature, so that it stays clear what is open. The owner: *"proven is a good name. That's what we do."* The original plan was that the owner picks from the fleet's answers by 20 October. He picked on 16 October, before the answers came in. The survey stays open until 19 October, and the answers will be published in `planning/answers/` after the owner reads them.
11. **How we publish.** The owner pushes once a day, at 22:00 Bangkok time (see below).

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

## Layout (from 22 October)

```
/            project code
/planning/   how the topic was chosen (here before 22 October)
/fleet/orders/
/fleet/status/
/fleet/reviews/
/fleet/decisions/
```

## What is left out

The owner reviews every document before it is published. Credentials, internal host names and details of our private infrastructure are removed. Where something was removed, the document says so.

## Pre-existing components

- **Saphan Studio** (https://saphan.ai) — our control plane for agent fleets. It runs the fleet. It is not part of this submission and is not in this repository.
- Model providers and tools used by the fleet are listed here as they are used.

## Licence

Apache-2.0. See `LICENSE`.
