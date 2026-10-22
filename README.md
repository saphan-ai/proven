# Proven

**A code reviewer that has to show its evidence, built in public by one person and a fleet of AI coding agents.**

Proven will review pull requests: proposed changes to a codebase. Every problem it reports must come with evidence you can check yourself, such as a test you can run or a data query you can repeat. If it cannot show the evidence, it does not report the problem.

[Marcin "Wato" Marzec](https://wato.org/) used fleets of AI coding agents on commercial telco software, and built [Saphan Studio](https://saphan.ai/) to direct and check their work. Now he is taking that fleet to the [Open Agent Hackathon 2026](https://hackathon.genai.works/event/open-agent-hackathon-2026), organized by GenAI Works, from **22 to 27 October 2026**, in the track *The Agent That Can Explain Why*, to build Proven. The agents write the code. Marcin sets the direction, makes the decisions, reviews the outcomes and approves every merge.

**Coding started on 22 October 2026, at 00:00 UTC.** Proven's code is written only inside the build window. Everything before 22 October was preparation, and all of it is public.

<a href="https://vimeo.com/1236933110"><img src="assets/proven-loop.gif" alt="Proven: a pull-request reviewer that has to prove what it says. The opening slides of the introduction video, looping; click to watch it." width="100%"></a>

<p align="center"><a href="https://vimeo.com/1236933110"><b>▶ Watch the introduction</b></a>: the idea, the evidence, the team (6 min). An introduction, not a product demo.</p>

## Why this exists

Marcin has spent almost three decades in internet infrastructure and security, including products that run at mobile operators. When he began using fleets of AI coding agents on commercial telco software, the hard part was not getting code written. It was knowing what to trust: who asked for a change, who checked it, and on what evidence.

That work led to patent applications, filed in the U.S. and Europe and still pending, and to **[Saphan Studio](https://saphan.ai/)**: a product that runs fleets of coding agents under human authority and keeps the record of their work. Saphan Studio is itself built by one person working with an agent fleet. The agents develop the platform that governs their own work; Marcin decides what ships.

Now that fleet takes on something new. At the hackathon it will build **Proven**, an open-source tool for checking code of the kind it writes itself.

## Where things stand: end of day 1, 22 October 2026

Day one of six is behind us. **Sixteen changes were merged into Proven today**, each one written by an agent, checked again by its master on a clean checkout on another machine, and approved by Marcin. Every merge commit carries the request it answers and the reason it was accepted.

What Proven can do at the end of day one, in sample mode (recorded data, no network):
- **The six reviewer agents run end to end** behind the command-line shell's request, and honour every field of it: baseline, head, include and exclude paths, and a configuration file.
- **The Skeptic tries to refute every finding** with a model from another family, and **the Judge settles it by evidence**. A finding nobody could settle is reported as unresolved, never as true.
- **The Prover writes the test that proves a finding** and runs it only inside a container, once on the base and once on the change.
- **The Investigator's evidence comes through Zetaris**: KEV, EPSS and OSV queries over REST, and the same queries run locally over samples in sample mode.
- **Every prompt and every answer passes a guardrail**, on by default, with two backends: NVIDIA's safety model on our own machine, or NVIDIA's hosted API. An unsafe prompt or answer stops the step and is recorded with its categories.
- **A minimum of Meterless's H-MEM and World Model in Go**, so a finding refuted once does not come back in the next review. Meterless's own conformance suites run against it, and the scorecard is recorded as it is: H-MEM 9 of 13, World Model 7 of 8. The failures are the sections not built yet.
- **The live generator calls Meterless Nightshift**, reads which model answered, and refuses an answer that did not come from NVIDIA.
- **One command checks every live service a review needs** before a demo, and names each one that is not ready, without printing a secret.

| what | status | where |
|---|---|---|
| Choosing the topic, with eleven answers from the fleet | done, 16 October | [answers](planning/answers/) |
| The design, as recorded decisions | 25 decisions accepted, 17–21 October | [decisions](fleet/decisions/) |
| Building Proven | **16 changes merged on day one**, by two trains: the reviewer and the shell | [git log](https://github.com/saphan-ai/proven/commits/main) · [reviewer order](fleet/orders/TRAIN_ORDER_proven-reviewer-core.md) · [shell order](fleet/orders/TRAIN_ORDER_proven-cli-and-action.md) |
| How the agents code with Cursor | four Cursor accounts on one Linux machine, up to four agents coding at once on Grok 4.7, each inside a jail, its own network and an egress door, with every model request recorded | [how it works](fleet/infra/cursor/README.md) · [order](fleet/orders/TRAIN_ORDER_proven-cursor-infra.md) |
| Recording every Cursor coding session | every session is recorded by the proxy; the first published recordings will show the coding itself, as excerpts tied to the sealed original | [ADR-0024](fleet/decisions/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md) |
| Zetaris, the sponsor's data platform | running on our own machine; the Investigator reads its evidence through Zetaris over REST; live Zetaris evidence in a review is next | [order](fleet/orders/TRAIN_ORDER_proven-zetaris-local.md) · [ADR-0023](fleet/decisions/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) · [setup](fleet/infra/zetaris/README.md) |
| Meterless Nightshift and NVIDIA Nemotron | running in its own VM; 206 signed receipts checked, every signature verifies, and a receipt ties each answer to the model that gave it | [NVIDIA page](fleet/infra/nvidia/README.md) · [runbook](fleet/infra/nightshift-vm/README.md) · [ADR-0019](fleet/decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) |
| A local NVIDIA safety model | done: running on our own machine, used by the guardrail | [order](fleet/orders/TRAIN_ORDER_nemotron-content-safety.md) |
| A test corpus from 28 real review disagreements | 21 case contracts written; the first five cases coded; the other 16 are ordered, one agent per case; declared prior art | [tests](tests/README.md) · [order](fleet/orders/TRAIN_ORDER_proven-test-corpus.md) |
| Reviews of the fleet's own changes | every change is checked again by its master; independent reviews run in parallel by two model families. NVIDIA Nemotron reviews resume once they run on Saphan Studio's own agent loop (planned for 23 October); the changes waiting for one are listed in each train's log | [conversation logs](planning/CONVERSATIONS.md) |
| Public updates | daily, 15:00 UTC | this repository |

**Saphan Studio's Cursor backend, built in parallel.** While the agents coded Proven, Saphan Studio's own fleet kept improving the Cursor backend that those agents run on, and each fix reached the live coding sessions as soon as it landed. Today it gained confinement for Cursor on macOS, Grok models named with their effort (`grok-4.7-high`), every Cursor spawn speaking HTTP/1.1 so the recording proxy can read it, placement of orders signed by a train's master, and a reader of a Cursor account's usage, so the scheduler can place Cursor runs by itself. Still to do: a Cursor seat in a container, and publishing the first session recordings.

Before Proven exists, we also wrote down what it will be tested on: [`tests/`](tests/README.md) turns real review disagreements from our fleet into checkable cases, and is declared as prior art.

The preparation work sets up existing machines and tools for the agents. It is not product code, and it trains no models. Saphan Studio is a pre-existing component; the organizers confirmed that it may run the fleet.

## Why a tool like this is needed

One example from the fleet's earlier work, before Proven:

An agent wrote a protection against deleting the main branch of a repository, plus 16 tests to show it worked. All 16 passed. An independent reviewer then **removed the new protection** and ran the tests again. All 16 still passed: a different safeguard, in git itself, was stopping the deletions one step earlier. The tests proved git, not the fix. ([source](planning/answers/master-the-bridge-commits-without-a-shared-lock.md))

Passing tests do not always prove what they claim to prove. A reviewer that asks "what does this evidence actually show?" catches what a green checkmark hides.

## What Proven is meant to do

*Planned, not built. Each part is described as intended.*

- **Every finding carries evidence.** A test that fails on the change and passes without it, or a recorded query and its results, with sources. This applies to every finding, including when two reviewers agree. Agreement is never a substitute for proof.
- **The reviewer should come from a different AI model family than the one that wrote the code.** A reviewer from the same family tends to share the author's assumptions; another family can challenge them. The change's own commit or pull request says which model wrote it, and by default an NVIDIA model reviews it. Every model is configurable, and no setup is blocked; each review says which model reviewed it and whether that is the author's.
- **For special kinds of review, such as security, two independent reviews.** Best with two models other than the author's; with fewer, the second review is a fresh agent that sees the code for the first time. In the fleet's own record, reviewers from two model families looked at the same change within 30 minutes and left a readable verdict 40 times, and reached different verdicts 28 times ([source](planning/answers/README.md); first published as 31 of 41, corrected on 19 October). That is an observation from one fleet, not a benchmark, but it shows why a second opinion is worth having.
- **When reviews conflict, the evidence decides, not a vote.** Both sides' evidence is examined: what it actually establishes, what it assumes, and whether it reproduces. What stays unresolved is reported as unresolved, and a person decides.

Proven will not prove a whole program correct or find every vulnerability. It makes each claim it does make checkable.

How one review is meant to run, and where each sponsor's technology does the work:

```
  pull request ──▶ proven ci   (GitHub Action · Gitea Actions · command line)
                       │
  ┌────────────────────▼─────────── Proven's agents ─────────────────────┐
  │ Planner ─▶ Investigator ─▶ Prover ─▶ Skeptic ─▶ Judge ─▶ Writer      │
  │    ▲                                               │                 │
  │    └──────────── weak finding: back for evidence ──┘                 │
  └──────┬──────────────────────┬──────────────────────────┬─────────────┘
         │ SQL over REST        │ every model call         │ read · write
         ▼                      ▼                          ▼
  ┌──────────────┐  ┌──────────────────────────┐  ┌──────────────────────┐
  │ ZETARIS      │  │ METERLESS NIGHTSHIFT     │  │ METERLESS ENGINES    │
  │ (Data)       │  │ primary AI provider      │  │ H-MEM: refuted       │
  │  OSV         │  │ (Power)                  │  │  findings, rulings   │
  │  CISA KEV    │  │   └─▶ NVIDIA NEMOTRON    │  │ World Model: claims, │
  │  FIRST EPSS  │  │       reviewer, Skeptic  │  │  evidence, verdicts  │
  │  git history │  │                          │  │                      │
  └──────┬───────┘  └──────────────────────────┘  └──────────┬───────────┘
         │ the query and its rows                            │ JSONL
         ▼                                                   ▼
  every finding with its evidence ─▶ SARIF · JSON · text · PR comment, logs/

  NVIDIA Nemotron 3.5 Content Safety checks every prompt and every answer.
```

The technical design, including the sponsors' technology (Zetaris, Meterless, NVIDIA, Cursor), is in **[DESIGN.md](DESIGN.md)**.

## Who builds it

Two different groups of AI agents appear in this project. Do not mix them up:

| | who | what they do |
|---|---|---|
| **The fleet that builds Proven** | coding agents run by Saphan Studio, with Marcin | plan each change, write the code, test it independently, review it, and hand Marcin a decision |
| **The agents inside Proven** | Proven's own reviewers, once built | review someone else's pull request: plan, gather evidence, try to refute, judge, write the report |

In the fleet, four Cursor agents write the product code. Other agents plan the work, write tests separately from the code, and review each change without having written it. Marcin sets the goals, rules on scope, settles disputes the evidence cannot settle, and approves every merge. An agent can recommend; only he can approve. How the fleet is organized is explained in **[SAPHAN.md](SAPHAN.md)**.

## Follow the work

- **[The diary](planning/CONVERSATIONS.md)**: every conversation between Marcin and the fleet, day by day.
- **[Watch it being built, slice by slice](fleet/)**: from 22 October, each change arrives with the request that asked for it, the builder's report, the independent review and Marcin's decision.
- **[Design decisions](fleet/decisions/)**: what was decided, and why.
- **[What it costs](fleet/LEDGER.md)**: the recorded cost of every change, by model.
- **[How to code with Saphan Studio, step by step](https://github.com/saphan-ai/quake-watch-server)**: published alongside Proven, a sample app built as a series. First the server, an earthquake alert service on the USGS public feed; the mobile clients follow in their own repositories. Every step is recorded as it is built: the order, the agents' work, the review and the owner's decision.

Every agent run records its model and the cost Saphan Studio measured for it. Where a cost was not measured, such as interactive planning sessions, or a share of a subscription the tools cannot read yet, the document says so and why. These figures are what the AI models cost; machines and human time are not included.

**Updates:** once a day at **15:00 UTC** (22:00 in Bangkok; 17:00 in Warsaw during the build week). The last push comes before submissions close on 27 October, 23:45 UTC.

## Timeline: what each push brought

Every public push is tagged, so you can see exactly what existed before the build window and what each day added.

| push | tag | what came in |
|---|---|---|
| before 22 October, 00:00 UTC: **prior art** | [`prior-art`](https://github.com/saphan-ai/proven/tree/prior-art) (`bf05dd2`, 21 October, 19:53 UTC) | Preparation only, no product code: the topic and the fleet's eleven answers, 25 design decisions, the trains' orders, the machines and services for the agents (Cursor seats, Zetaris, Meterless Nightshift, NVIDIA's safety model), and the 21 test-corpus case contracts |
| 22 October, 00:03 UTC: **coding starts** | [`start`](https://github.com/saphan-ai/proven/compare/prior-art...start) (`3a1a9f0`) | The first line of product code: the agreed seam between the command-line shell and the reviewer, types only |
| 22 October, 15:00 UTC: **day 1** | [`day-1`](https://github.com/saphan-ai/proven/compare/start...day-1) | 16 changes merged: the six reviewer agents end to end in sample mode, the Skeptic and the Judge, the Prover writing and running its test in a container, evidence through Zetaris, the guardrail, the H-MEM and World Model minimum with Meterless's conformance scorecard, the live generator through Nightshift, and the live-service check. The first documentation and the Cursor and NVIDIA setup pages. |

**Contact:** Marcin Marzec, [marcin.marzec@wato.org](mailto:marcin.marzec@wato.org), [LinkedIn](https://www.linkedin.com/in/marcinmarzec).

## For technical readers

| | |
|---|---|
| [docs/SAMPLE_MODE.md](docs/SAMPLE_MODE.md) | try Proven with no network and no keys: the sample review, its findings and its trace |
| [DESIGN.md](DESIGN.md) | Proven's design, the sponsors' technology, evidence sources, open questions |
| [SAPHAN.md](SAPHAN.md) | the platform that runs the fleet, its roles and vocabulary |
| [AGENTS.md](AGENTS.md) | the rules every agent follows in this repository |
| [planning/](planning/) | how the topic was chosen, and the fleet's answers |
| [fleet/](fleet/) | orders, reports, reviews, decisions and the cost ledger, with their templates |

**Pre-existing components:** [Saphan Studio](https://saphan.ai/) runs the fleet; it is not part of this submission and not in this repository. The [Meterless engine specifications](https://github.com/Meterless/Meterless) (Apache-2.0) are open source; we build to them and do not copy their reference code. Model providers, tools and data sources are listed as they are used.

Before publication Marcin reviews every document. Credentials, internal host names and private infrastructure details are blacked out, and each document says where.

**Licence:** Apache-2.0. See [LICENSE](LICENSE).
