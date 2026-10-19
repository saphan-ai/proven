# proven

A pull-request reviewer that has to prove what it says, and shows where every proof came from.

An entry for the Open Agent Hackathon 2026, track "The Agent That Can Explain Why".

**There is no code here yet, on purpose.** The build window opens on 22 October 2026, and no project code is written before that date. What is here already is the conversation about what to build: the question the owner put to the fleet, the proposals, the plan for the submission, and the design below. See `planning/`.

## What proven will do

*The design as of 17 October 2026. It was reshaped that day, because the hackathon requires every entry to integrate Zetaris and Meterless (see step 12 below). Nothing here is built yet. Every claim in this section is a plan, not a result.*

**The problem.** A code review says "this is wrong" and rarely says why anyone should believe it. The evidence a good reviewer would need is spread over many places: the change itself, the project's history, its issues, the advisories published for its dependencies. Nobody reads all of them for every pull request.

**What proven does.** A finding counts only with evidence, and the evidence is something a reader can check again:

- **a test** that fails on the change and passes on the base, or
- **a query and the rows it returned**: one SQL query over the evidence sources below, written into the review next to the finding, so anyone can run it again and get the same rows.

A finding without either is dropped, not softened.

**The agents.** Each one has its own job, and they hand work to each other more than once:

| agent | what it does |
|---|---|
| Planner | reads the change, says what it claims to do and where the risk is, and asks the others for evidence |
| Investigator | turns each risk into queries over the evidence sources and returns the rows |
| Prover | writes the test that fails on the change and passes on the base, and runs both |
| Skeptic | a model from a different family, whose only job is to refute each finding |
| Judge | settles every disagreement by the evidence, never by a vote, and sends weak findings back for a second pass |
| Writer | writes the review: each finding with its proof, its sources and what the review cost |

The run is bounded: a fixed number of rounds, a timeout on every outside call, and a partial result instead of a hang.

### Where the sponsors' technology does the work

- **Zetaris: the evidence layer.** Every evidence source is one table behind Zetaris, which runs a query where the data already lives, without copying it first. The Investigator asks one question across all of them in one query. For example: *"Which dependency this change adds or upgrades has a published advisory for that exact version, is on the list of vulnerabilities known to be used in attacks, and has a high chance of being used in one?"* The query and its rows go into the review as the proof.
- **Meterless: the reviewer's memory and state.** Meterless publishes open specifications of agent engines, each with a conformance suite. We build two of them in Go:
  - **H-MEM**, the reviewer's memory. A finding that was refuted in one pull request does not come back in the next one. A maintainer's ruling ("we do not validate here, on purpose") is remembered.
  - **World Model**, the state of one review as a stream of events: every claim, piece of evidence, refutation and verdict. The trace a judge reads in `logs/` is that stream, not text written next to it.
  - Our Go builds are checked with Meterless's own conformance suites. The suites load a TypeScript module, so a thin TypeScript adapter connects them to the Go code. That this works is not measured yet.
- **NVIDIA: the second family.** The Skeptic runs on an NVIDIA Nemotron model through NVIDIA's hosted API, so the finder and the refuter never come from the same model family. Hosted NVIDIA embeddings for H-MEM's search are a candidate, not a decision.
- **NVIDIA: the guardrail, on by default.** Before an agent's prompt or answer is used, proven checks it with a guardrail model, as Saphan Studio does: NVIDIA's Nemotron 3.5 Content Safety. The same model runs in two places, behind one adapter: on our own machine, and on NVIDIA's hosted API ([build.nvidia.com](https://build.nvidia.com/nvidia/nemotron-3.5-content-safety), the same endpoint as the Skeptic). So a judge with an NVIDIA API key keeps the guardrail on, and a setting turns it off for anyone without one. We call the hosted model with our own NVIDIA developer key.
- **Cursor: the executors.** The four Cursor agents of the team write the product code (see "How this project is built").

### The evidence sources

The first set, decided on 17 October 2026:

| source | what it answers | how it gets in |
|---|---|---|
| [OSV](https://osv.dev) | which package versions an advisory covers, per ecosystem | public data, downloaded by a script |
| [CISA KEV](https://www.cisa.gov/known-exploited-vulnerabilities-catalog) | whether a vulnerability is known to be used in real attacks | public data, downloaded by a script |
| [FIRST EPSS](https://www.first.org/epss/) | how likely a vulnerability is to be used in an attack | public data, downloaded by a script |
| the repository's own history | what was changed, reverted or fixed before, and in which commit | read from the repository under review |

Next, if the build leaves time: the NVD CVE records, the CWE list of weakness classes, and OpenSSF Scorecard for the health of each dependency.

All data is downloaded by scripts written in the build window, from 22 October. Small samples are kept in this repository, so `SAMPLE_MODE=true` runs the whole review with no network and no keys. The same queries then run on a local engine over the samples instead of on Zetaris. The licence of every source is listed here once its script exists.

proven writes no exploit code. A finding about a vulnerable dependency is proved by the query and its rows, never by an attack.

### Why Go

proven is open source, and one part of it will also be used inside Saphan Studio (see step 3 below). The owner chose Go for that reason. Every outside service (Zetaris, Meterless's suites, NVIDIA, the model providers) sits behind an adapter: an outage during judging degrades the review instead of stopping it, and a judge without our keys can still run it in sample mode.

### Not known yet

- **Zetaris access.** The instance and its documentation are promised to participants before 22 October. Until they arrive, we do not know which interface we will use (SQL driver, REST or MCP).
- **Meterless keys.** The organizers' example settings name a Meterless API key. The public Meterless repository describes no hosted service and needs no key. We will ask on the sponsor's help channel.
- **The hosted guardrail, in detail.** NVIDIA's model page names the endpoint and the answer format, but its code sample leaves the model id blank, and it states no free tier and no rate limit. The first live call settles the model id and the limits.
- **The conformance adapter.** Whether a Go engine passes Meterless's suites through a TypeScript adapter is the first thing we measure.

## How the topic was chosen (16 October 2026)

The topic was chosen in one day: in one conversation between the owner and the advisor, and in a question sent to the fleet. The owner's words are translated from Polish. The whole exchange, hop by hop, with his words in the original Polish, is in [`planning/CONVERSATION_2026-10-16.md`](planning/CONVERSATION_2026-10-16.md). Every conversation since, day by day and hop by hop, is indexed in [`planning/CONVERSATIONS.md`](planning/CONVERSATIONS.md).

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
12. **The sponsors reshape the design (17 October).** The owner read the hackathon's technical guide and found that our design did not meet it: *"it does not meet the requirements, because we have to use all the sponsors' technologies."* Every entry must integrate Zetaris and Meterless, and NVIDIA wherever it helps. Zetaris became the layer every piece of evidence is queried through, and Meterless the reviewer's memory and state. The owner also widened what the reviewer reads: *"Once we get the platform we can go wild. For example, pulling CVE databases as data, other data sources for the evaluating agent."* He chose Go, behind adapters: *"Later we will use it in Saphan. That is the point of this game."* The track stays the same: *"There will be a better explanation."* The hop-by-hop record is in [`planning/CONVERSATION_2026-10-17.md`](planning/CONVERSATION_2026-10-17.md), hops 11 and 12.

## What the fleet answered

Before a single line of code, the fleet's answers to the owner's question already show how this project will be built. Ten masters and the scribe answered the same evening, each from the record of its own work. They are published verbatim in [`planning/answers/`](planning/answers/), with only internal names blacked out. Read [`planning/answers/README.md`](planning/answers/README.md) for what they show:

- **The tests are green and the review says no.** It happens all the time, and the reviewer is usually right. One track had three such cases in two days.
- **The agents disagree, and say so.** Two model families reviewing the same change disagreed in 31 of 41 cases. Executors refute their masters with tests. Masters write *"MASTER WAS WRONG"* about themselves.
- **A disagreement ends with a run, a rule or the owner.** Never with a vote.
- **The tool is needed.** Every one of 3,030 failure rows in our own store has an empty cause, and the answers to our hardest questions were spread over four places.

The words and laws used there are explained in [`SAPHAN.md`](SAPHAN.md).

## How this project is built

A fleet of AI agents builds this project between 22 and 27 October 2026.

**The team has five members, the most the hackathon allows: four Cursor agents and the owner.** The Cursor agents are the executors: they write the product code. We start with two of them in parallel and set the number, from one to four, once a day. Planning, tests, reviews and landings are done by Claude agents of our fleet, which Saphan Studio runs: a master measures the code first and then gives the executor the exact places to change; a test author writes the tests from the test plan, apart from the code; a reviewer that did not write the change checks it. The other agents named in this README (the masters, the scribe, the advisor and the reviewers) belong to the same fleet.

The owner sets the goals, makes the decisions, and reviews and merges every change. He writes no code.

The fleet works on our own infrastructure. Every change is reviewed, and the owner merges it. The paperwork behind each change is committed next to the code:

- **Orders** — what an agent was asked to build and the properties the result must hold.
- **Status reports** — what the agent did, what it ran, and what failed.
- **Reviews** — an agent that did not write the change checks it, starting from a fresh context, and the owner decides.
- **Decisions** — the owner's decisions, each with a reason he wrote himself.

Read the history change by change. For each one you can see who asked for it, who built it, who checked it, and why it was merged or sent back.

## What every change costs, and how good it was

Every change in this repository has a price, to the cent.

- **Every review that proven writes states what it cost**, to the cent, next to its findings.
- **Every executor run and every review run of the fleet states its model and its cost.** The cost is the one Saphan Studio recorded for the run.
- **Every merge is priced:** the model that built it, the number of repair rounds, the BLOCKING and OWED conditions the reviews found, and the total cost of building and reviewing it.
- **[`fleet/LEDGER.md`](fleet/LEDGER.md)** collects one row per change. At the end of the build it answers, per model: how many changes were accepted in their first round, how many repair rounds a change needed, how many BLOCKING and OWED conditions it drew, and what a merged change cost.

- **Which model really ran.** Cursor can pick the model itself (Auto). Every run records the model it actually ran on, including the one Auto chose.
- **How much of the month it used.** Every change records the share of the account's monthly model pool before and after it. Spending beyond the included pool happens only within an amount the owner sets.
- **Steering, in the open.** Once a day the owner decides which models the executors use and how many run in parallel, from the ledger, and writes the reason down. The ledger shows each day's setting next to what it produced: cost, rounds and conditions per change.

The orders are public too, so a reader can judge whether a model wrote bad code or an order asked for the wrong thing.

## When this repository is updated

Once a day, at **22:00 Bangkok time (15:00 UTC)**, the owner pushes the day's merged changes together with the documents behind them. Between pushes the repository does not change.

- The first push with code: 22 October 2026.
- The last daily push of the build window: 27 October 2026.
- One final push on 27 October 2026, before we submit and before submissions close at 23:45 UTC.

## Layout

```
/                    project code, in Go (from 22 October)
/SAPHAN.md           the platform that runs the fleet, its roles, words and laws
/AGENTS.md           the rules every agent follows in this repository
/planning/           how the topic was chosen (here before 22 October)
/planning/CONVERSATIONS.md   the diary: every conversation, hop by hop, with links
/planning/answers/   the fleet's eleven answers, verbatim, and what they show
/fleet/              the paperwork behind every change
/fleet/orders/       what a master asked for
/fleet/status/       what the executor built and ran
/fleet/reviews/      what an independent reviewer found
/fleet/decisions/    what the owner decided, and why: design decisions (ADRs) and rulings on each change
/fleet/LEDGER.md     what every change cost, and how good it was
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
- **Who writes it:** the executor, a Cursor agent.
- **What it looks like:**

```markdown
# STATUS — <slug> · round <n>
- state: in-progress | blocked | ready-for-review
- executor · model · cost to the cent · last commit · updated
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
- reviewer (agent and model) · cost to the cent · conditions: n BLOCKING · n OWED · n NOTE
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
- decided by: the owner · merged as (commit, or "not merged") · decided
- rounds and repair rounds · model per round · BLOCKING / OWED / NOTE over all rounds
- cost of the change: building + reviewing, to the cent
## Reason        the owner's own words
## Conditions    each one and the later change that closes it
```

The full templates, with every field explained, are in the README of each folder.

## What is left out

The owner reviews every document before it is published. Credentials, internal host names and details of our private infrastructure are removed. Where something was removed, the document says so.

## Pre-existing components

- **Saphan Studio** (https://saphan.ai) — our control plane for agent fleets. It runs the fleet. It is not part of this submission and is not in this repository.
- **The Meterless engine specifications and conformance suites** (https://github.com/Meterless/Meterless, Apache-2.0) — open source, published by the sponsor. We build to them; we do not copy their reference code.
- Model providers, tools and data sources are listed here as they are used.

## Licence

Apache-2.0. See `LICENSE`.
