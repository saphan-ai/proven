# Proven

**A code reviewer that has to show its evidence, built in public by one person and a fleet of AI coding agents.**

## Quick start

```bash
make up
```

`make up` builds Proven and starts the judges' lab on this machine. It prints the lab address and a generated admin password, once. Open that address, then the open pull request: the sample review is a comment on it. The steps are in [docs/LAB.md](docs/LAB.md). `make down` stops the lab.

Proven will review pull requests: proposed changes to a codebase. Every problem it reports must come with evidence you can check yourself, such as a test you can run or a data query you can repeat. If it cannot show the evidence, it does not report the problem.

[Marcin "Wato" Marzec](https://wato.org/) used fleets of AI coding agents on commercial telco software, and built [Saphan Studio](https://saphan.ai/) to direct and check their work. Now he is taking that fleet to the [Open Agent Hackathon 2026](https://hackathon.genai.works/event/open-agent-hackathon-2026), organized by GenAI Works, from **22 to 27 October 2026**, in the track *The Agent That Can Explain Why*, to build Proven. The agents write the code. Marcin sets the direction, makes the decisions, reviews the outcomes and approves every merge.

**Coding started on 22 October 2026, at 00:00 UTC.** Proven's code is written only inside the build window. Everything before 22 October was preparation, and all of it is public.

<a href="https://vimeo.com/1236933110"><img src="assets/proven-loop.gif" alt="Proven: a pull-request reviewer that has to prove what it says. The opening slides of the introduction video, looping; click to watch it." width="100%"></a>

<p align="center"><a href="https://vimeo.com/1236933110"><b>▶ Watch the introduction</b></a>: the idea, the evidence, the team (6 min). An introduction, not a product demo.</p>

## Why this exists

Marcin has spent almost three decades in internet infrastructure and security, including products that run at mobile operators. When he began using fleets of AI coding agents on commercial telco software, the hard part was not getting code written. It was knowing what to trust: who asked for a change, who checked it, and on what evidence.

That work led to patent applications, filed in the U.S. and Europe and still pending, and to **[Saphan Studio](https://saphan.ai/)**: a product that runs fleets of coding agents under human authority and keeps the record of their work. Saphan Studio is itself built by one person working with an agent fleet. The agents develop the platform that governs their own work; Marcin decides what ships.

Now that fleet takes on something new. At the hackathon it will build **Proven**, an open-source tool for checking code of the kind it writes itself.

## Where things stand

Updated every evening. The newest day comes first.

### Day two, 23 October

Day two of six is behind us. **Eighteen more changes were merged** (35 in two days), each written by an agent, checked again by its master on a clean checkout, and approved by Marcin; the larger ones were also reviewed by two models other than the one that wrote them.

**The day's biggest step: Proven's code was put through a security review of its own.** Not one diff at a time, but
the whole code, cut into nine pieces, each one place where something untrusted meets something Proven protects: the
sandbox that runs the change's code, the keys, the reviewed repository, the evidence queries, the models, the output.
Each piece was reviewed by **two model families on the same commit**, Anthropic's Claude and OpenAI's, neither seeing
the other's verdict. They did not always agree: on two of the first three pieces one reviewer accepted with conditions
and the other rejected the same code and proved it with failing tests. With one reviewer, those pieces would have
passed. **The first weakness was found, fixed, proved by six tests and merged the same day**: the command run inside
the sandbox could come from a model's answer, and now Proven builds it itself. Tomorrow NVIDIA's Nemotron reviews the
pieces again as a third family. The plan, the status of every piece and the verdicts side by side:
[docs/security/SECURITY_REVIEW.md](docs/security/SECURITY_REVIEW.md); what was fixed: [SECURITY.md](SECURITY.md).

What else came in today:
- **Proven's own command line**: `proven scan` for a developer and `proven ci` for a pipeline, with the flags and exit codes a CI pipeline already expects. It took 14 rounds: coded by Grok on Cursor, reviewed each round by Claude and OpenAI.
- **More dependency evidence**: `go.sum`, npm, `requirements.txt` and `poetry.lock` are read; each changed package is checked against OSV by version range, the known-exploited list, and its EPSS score.
- **One sample scenario per kind of evidence and refusal**, all runnable with no network and no keys.
- **A missing key no longer stops a review**: the service is named as down, and a guardrail that is down refuses every model call instead of letting it through unchecked.
- **Code written to be read**: the masters went over the code written so far and added the comments a judge needs.
- **We reshaped the repository, because Proven is no longer a toy**: the product at the top, laid out the way a Go project is, and the build's record in [`hackathon/`](hackathon/), so that after the hackathon what is left is a product. Eleven design decisions were written today ([ADR-0026 to ADR-0036](docs/adr/)).
- **Two more agent masters joined**, one for the command line and one for the security review, because one master was carrying four streams of work.

**See it run:** [what shipped on day two, step by step in a terminal](hackathon/SHIPPED_DAY_2.md).

### What works today with the sponsors' technology

| Sponsor | What works, measured | In a review |
|---|---|---|
| **Zetaris** | The platform runs on our own machine and its checks pass 9 of 9. Proven's own key reads over REST and is refused every write. OSV by package answers from a file in Zetaris storage, 0.10 to 0.15 s per query. | The Investigator's evidence (KEV, EPSS, OSV) is shaped as Zetaris queries; sample mode runs the same SQL over local tables. [setup](hackathon/fleet/infra/zetaris/README.md) |
| **Meterless** | Nightshift answers every model call with a signed receipt: 206 receipts, all 206 signatures verify. H-MEM and the World Model run in Go and pass Meterless's own conformance suites 9 of 13 and 7 of 8; the rest are sections not built yet. | H-MEM and the World Model run in every review, sample mode included; the fleet's own reviews also go through Nightshift. [NVIDIA and Nightshift](hackathon/fleet/infra/nvidia/README.md) |
| **NVIDIA** | Nemotron 3.5 Super answers through Nightshift, named in each signed receipt. Nemotron 3.5 Content Safety runs on our own machine as the guardrail. | Every prompt and answer passes the guardrail; the Skeptic that tries to refute a finding defaults to Nemotron. |
| **Cursor** | The command line was written by Cursor agents on Grok 4.7 and landed today; up to four agents code at once on one Linux machine, each inside its own jail, and every run's model traffic is recorded. | Cursor is how Proven is built. [how it works](hackathon/fleet/infra/cursor/README.md) |

### Next: everything still in the plan

Each day, what lands moves from this list to the day's summary.

- **Connect your own second reviewer**, for example Grok, OpenAI or Claude, and see where the two disagree; what both agree on without proof is shown apart, as an opinion ([ADR-0037](docs/adr/ADR-0037-bring-your-own-second-reviewer.md), [ADR-0041](docs/adr/ADR-0041-agreed-by-two-models-is-shown-apart.md)).
- **Run Proven live with a single free NVIDIA key** ([ADR-0031](docs/adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md)).
- **Reports people can read**, in HTML and Markdown, next to JSON and SARIF for machines ([ADR-0033](docs/adr/ADR-0033-the-text-review-is-written-for-a-person.md)).
- **Comments on the pull request**, in GitHub and in Gitea, and a GitHub Action to add Proven to any pipeline.
- **A ready-to-run Docker environment with a local Gitea**, where you can watch Proven review a pull request ([ADR-0026](docs/adr/ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md)).
- **More proof, not only tests**: a data race or a crash found by running real tools on the old and the new code ([ADR-0040](docs/adr/ADR-0040-a-tool-run-is-evidence.md)).
- **More sources**: packages published on purpose to do harm, project health, severity scores and licences ([ADR-0039](docs/adr/ADR-0039-more-evidence-sources.md)), and **your own data**: plug in an internal list of banned packages, how critical each service is, or who owns the code, and Proven judges a change against what your organisation knows ([ADR-0046](docs/adr/ADR-0046-a-data-source-is-one-adapter.md)).
- **More languages**: PHP, Rust, Java, C# and Ruby, next to Go, JavaScript and Python, and a simple adapter so anyone can add theirs ([ADR-0038](docs/adr/ADR-0038-a-language-is-one-small-adapter.md)).
- **An air-gapped installation**, where no code leaves your network ([ADR-0034](docs/adr/ADR-0034-proven-runs-air-gapped.md)).
- **Features for compliance audits**: a signed record of every review that an auditor can check ([ADR-0035](docs/adr/ADR-0035-every-review-leaves-a-signed-evidence-pack.md)).
- **One security policy for a whole organisation**, and accepted risks that expire ([ADR-0036](docs/adr/ADR-0036-an-organisation-policy-and-exceptions-that-expire.md)).
- **A fix you can trust**: for a problem proved by a test, Proven proposes a fix, and the same test proves the fix works ([ADR-0042](docs/adr/ADR-0042-a-proposed-fix-is-proved-by-the-same-test.md)).
- **Is the vulnerable code actually called?** For Go, each vulnerable dependency says whether your code reaches the vulnerable function ([ADR-0043](docs/adr/ADR-0043-a-vulnerable-dependency-is-checked-for-reachability.md)).
- **How far to trust each finding, in one word**: proved by a test, by a tool, by data, or only agreed by two models ([ADR-0044](docs/adr/ADR-0044-every-finding-names-its-level-of-proof.md)).
- **Watch a review live**: a dashboard of the agents and the two reviewers working side by side, their verdicts and where they disagree; the numbers go to any monitoring you already use, through OpenTelemetry ([ADR-0045](docs/adr/ADR-0045-a-review-is-observable-through-opentelemetry.md)).
- **Large pull requests at a flat cost**: the review goes file by file and carries only what it has established, using Meterless's Markovian engine. Saphan Studio has its own, patent pending; for the hackathon we use Meterless's. We keep it only if it finds the same problems for less ([ADR-0048](docs/adr/ADR-0048-a-large-change-is-reviewed-in-bounded-steps.md)).
- **NVIDIA's model joins our security review** as a third reviewer.
- **A measurement on 21 real review disputes**: how often Proven is right compared with a single model.

Four days to go. It will be tough, but we will make it.

### Day one, 22 October

**Seventeen changes were merged into Proven on day one**, each one written by an agent, checked again by its master on a clean checkout on another machine, and approved by Marcin. Every merge commit carries the request it answers and the reason it was accepted.

**See it run:** [what shipped on day one, step by step in a terminal](hackathon/SHIPPED_DAY_1.md).

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
| Choosing the topic, with eleven answers from the fleet | done, 16 October | [answers](hackathon/planning/answers/) |
| The design, as recorded decisions | 25 decisions accepted, 17–21 October | [decisions](docs/adr/) |
| Building Proven | **17 changes merged on day one**, all from the reviewer train; the shell's first change is still in progress | [git log](https://github.com/saphan-ai/proven/commits/main) · [reviewer order](hackathon/fleet/orders/TRAIN_ORDER_proven-reviewer-core.md) · [shell order](hackathon/fleet/orders/TRAIN_ORDER_proven-cli-and-action.md) |
| How the agents code with Cursor | four Cursor accounts on one Linux machine, up to four agents coding at once on Grok 4.7, each inside a jail, its own network and an egress door, with every model request recorded | [how it works](hackathon/fleet/infra/cursor/README.md) · [order](hackathon/fleet/orders/TRAIN_ORDER_proven-cursor-infra.md) |
| Recording every Cursor coding session | every session is recorded by the proxy; the first published recordings will show the coding itself, as excerpts tied to the sealed original | [ADR-0024](docs/adr/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md) |
| Zetaris, the sponsor's data platform | running on our own machine; the Investigator reads its evidence through Zetaris over REST; live Zetaris evidence in a review is next | [order](hackathon/fleet/orders/TRAIN_ORDER_proven-zetaris-local.md) · [ADR-0023](docs/adr/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) · [setup](hackathon/fleet/infra/zetaris/README.md) |
| Meterless Nightshift and NVIDIA Nemotron | running in its own VM; 206 signed receipts checked, every signature verifies, and a receipt ties each answer to the model that gave it | [NVIDIA page](hackathon/fleet/infra/nvidia/README.md) · [runbook](hackathon/fleet/infra/nightshift-vm/README.md) · [ADR-0019](docs/adr/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) |
| A local NVIDIA safety model | done: running on our own machine, used by the guardrail | [order](hackathon/fleet/orders/TRAIN_ORDER_nemotron-content-safety.md) |
| A test corpus from 28 real review disagreements | the 21 case contracts are prior art, written before 22 October; their code is written inside the build window: the first five cases are coded on branches, not merged yet, and the other 16 are ordered, one agent per case | [tests](test/README.md) · [order](hackathon/fleet/orders/TRAIN_ORDER_proven-test-corpus.md) |
| A held-out set of six fresh cases | frozen in commit `0fabc62` (text only, no run yet): contracts and answer keys, written before Proven is run on any of them; one run follows, misses published | [held-out set 1](test/corpus/held-out/README.md) |
| Reviews of the fleet's own changes | every change is checked again by its master; independent reviews run in parallel by two model families. NVIDIA Nemotron reviews resume once they run on Saphan Studio's own agent loop (planned for 23 October); the trains keep a list of the changes waiting for one | [conversation logs](hackathon/planning/CONVERSATIONS.md) |
| Public updates | daily, 15:00 UTC | this repository |

**Saphan Studio's Cursor backend, built in parallel.** While the agents coded Proven, Saphan Studio's own fleet kept improving the Cursor backend that those agents run on, and each fix reached the live coding sessions as soon as it landed. Today it gained confinement for Cursor on macOS, Grok models named with their effort (`grok-4.7-high`), every Cursor spawn speaking HTTP/1.1 so the recording proxy can read it, placement of orders signed by a train's master, and a reader of a Cursor account's usage, so the scheduler can place Cursor runs by itself. Still to do: a Cursor seat in a container, and publishing the first session recordings.

Before Proven exists, we also wrote down what it will be tested on: [`tests/`](test/README.md) turns real review disagreements from our fleet into checkable cases, and is declared as prior art.

The preparation work sets up existing machines and tools for the agents. It is not product code, and it trains no models. Saphan Studio is a pre-existing component; the organizers confirmed that it may run the fleet.

## Why a tool like this is needed

One example from the fleet's earlier work, before Proven:

An agent wrote a protection against deleting the main branch of a repository, plus 16 tests to show it worked. All 16 passed. An independent reviewer then **removed the new protection** and ran the tests again. All 16 still passed: a different safeguard, in git itself, was stopping the deletions one step earlier. The tests proved git, not the fix. ([source](hackathon/planning/answers/master-the-bridge-commits-without-a-shared-lock.md))

Passing tests do not always prove what they claim to prove. A reviewer that asks "what does this evidence actually show?" catches what a green checkmark hides.

## What Proven is meant to do

*The intent, as planned before the build window. What already works is listed under [Where things stand](#where-things-stand-end-of-day-1-22-october-2026).*

- **Every finding carries evidence.** A test that fails on the change and passes without it, or a recorded query and its results, with sources. This applies to every finding, including when two reviewers agree. Agreement is never a substitute for proof.
- **The reviewer should come from a different AI model family than the one that wrote the code.** A reviewer from the same family tends to share the author's assumptions; another family can challenge them. The change's own commit or pull request says which model wrote it, and by default an NVIDIA model reviews it. Every model is configurable, and no setup is blocked; each review says which model reviewed it and whether that is the author's.
- **For special kinds of review, such as security, two independent reviews.** Best with two models other than the author's; with fewer, the second review is a fresh agent that sees the code for the first time. In the fleet's own record, reviewers from two model families looked at the same change within 30 minutes and left a readable verdict 40 times, and reached different verdicts 28 times ([source](hackathon/planning/answers/README.md); first published as 31 of 41, corrected on 19 October). That is an observation from one fleet, not a benchmark, but it shows why a second opinion is worth having.
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

The technical design, including the sponsors' technology (Zetaris, Meterless, NVIDIA, Cursor), is in **[DESIGN.md](hackathon/DESIGN.md)**.

## How Proven differs

Large vendors are moving the same way: a security finding should be proved, not only reported. Proven takes that idea
somewhere they do not:

- **Open source, and it runs where your code is**: on GitHub, on your own Gitea, on your own servers, or fully
  air-gapped. It reviews every pull request, not a nightly scan.
- **More than one model, in the open.** Proven names the models that reviewed each change. Connect a second reviewer
  and you see both verdicts, and exactly where they disagree.
- **Proof anyone can rerun**: a test that passes on the old code and fails on the new, or a query and the rows it
  returned. No exploit code is written or kept.
- **A fix is proved too**: the same test that showed the problem shows the fix works.

## Who builds it

Two different groups of AI agents appear in this project. Do not mix them up:

| | who | what they do |
|---|---|---|
| **The fleet that builds Proven** | coding agents run by Saphan Studio, with Marcin | plan each change, write the code, test it independently, review it, and hand Marcin a decision |
| **The agents inside Proven** | Proven's own reviewers, once built | review someone else's pull request: plan, gather evidence, try to refute, judge, write the report |

In the fleet, four Cursor agents write the product code. Other agents plan the work, write tests separately from the code, and review each change without having written it. Marcin sets the goals, rules on scope, settles disputes the evidence cannot settle, and approves every merge. An agent can recommend; only he can approve. How the fleet is organized is explained in **[SAPHAN.md](hackathon/SAPHAN.md)**.

## Follow the work

- **[Documentation](docs/README.md)**: how to run Proven, starting with a review on samples (no network, no keys), its architecture and every setting.
- **[The diary](hackathon/planning/CONVERSATIONS.md)**: every conversation between Marcin and the fleet, day by day.
- **[Watch it being built, slice by slice](hackathon/fleet/)**: from 22 October, each change arrives with the request that asked for it, the builder's report, the independent review and Marcin's decision.
- **[Design decisions](docs/adr/)**: what was decided, and why.
- **[What it costs](hackathon/fleet/LEDGER.md)**: the recorded cost of every change, by model.
- **[Security](SECURITY.md)**: what Proven treats as untrusted, what it protects and how, and **[how its code is reviewed for security as a whole](docs/security/SECURITY_REVIEW.md)**, piece by piece.
- **[How to code with Saphan Studio, step by step](https://github.com/saphan-ai/quake-watch-server)**: published alongside Proven, a sample app built as a series. First the server, an earthquake alert service on the USGS public feed; the mobile clients follow in their own repositories. Every step is recorded as it is built: the order, the agents' work, the review and the owner's decision.

Every agent run records its model and the cost Saphan Studio measured for it. Where a cost was not measured, such as interactive planning sessions, or a share of a subscription the tools cannot read yet, the document says so and why. These figures are what the AI models cost; machines and human time are not included.

**Updates:** once a day at **15:00 UTC** (22:00 in Bangkok; 17:00 in Warsaw until 25 October, 16:00 after). The last push comes before submissions close on 27 October, 23:45 UTC.

## Timeline: what each push brought

Every public push is tagged, so you can see exactly what existed before the build window and what each day added.

| push | tag | what came in |
|---|---|---|
| before 22 October, 00:00 UTC: **prior art** | [`prior-art`](https://github.com/saphan-ai/proven/tree/prior-art) (`bf05dd2`, 21 October, 19:53 UTC) | Preparation only, no product code: the topic and the fleet's eleven answers, 25 design decisions, the trains' orders, the machines and services for the agents (Cursor seats, Zetaris, Meterless Nightshift, NVIDIA's safety model), and the 21 test-corpus case contracts |
| 22 October, 00:03 UTC: **coding starts** | [`start`](https://github.com/saphan-ai/proven/compare/prior-art...start) (`3a1a9f0`) | The first line of product code: the agreed seam between the command-line shell and the reviewer, types only |
| 22 October, 15:17 UTC: **day 1** | [`day-1`](https://github.com/saphan-ai/proven/compare/start...day-1) | 17 changes merged: the six reviewer agents end to end in sample mode, the Skeptic and the Judge, the Prover writing and running its test in a container, evidence through Zetaris, the guardrail, the H-MEM and World Model minimum with Meterless's conformance scorecard, the live generator through Nightshift, the live-service check, and a small live demo a judge can rerun. The first documentation and the Cursor and NVIDIA setup pages. [What shipped, in a terminal](hackathon/SHIPPED_DAY_1.md). |
| 23 October, 15:00 UTC: **day 2** | [`day-2`](https://github.com/saphan-ai/proven/compare/day-1...day-2) | 18 changes merged: Proven's command line (`scan`, `ci`), dependency evidence for go.sum, npm, requirements.txt and poetry.lock with OSV ranges, KEV and EPSS, one sample scenario per kind of evidence, a missing key named as a down service. A security review of the whole code in nine pieces, with its first fix. The repository split into the product and `hackathon/`; ADR-0026 to ADR-0036. [What shipped, in a terminal](hackathon/SHIPPED_DAY_2.md). |

**Contact:** Marcin Marzec, [marcin.marzec@wato.org](mailto:marcin.marzec@wato.org), [LinkedIn](https://www.linkedin.com/in/marcinmarzec).

**Found a vulnerability?** Please report it privately: [Report a vulnerability](https://github.com/saphan-ai/proven/security/advisories/new) on this repository's Security tab (GitHub's private vulnerability reporting is on). Not in a public issue. See [SECURITY.md](SECURITY.md).

## Repository layout

The repository holds two things: **the product**, and, in one folder, **the record of how it was built in the hackathon
week** ([ADR-0030](docs/adr/ADR-0030-the-product-and-the-hackathon-live-apart.md)).

| Path | What it is |
|---|---|
| `cmd/`, `internal/` | Proven's code: the command line and the reviewer |
| `docs/` | how to install, run and configure Proven; `docs/adr/` holds the design decisions, `docs/security/` the security review |
| `SECURITY.md` | what Proven protects, and how to report a vulnerability |
| `test/` | the test corpus (`test/corpus/`) and the Meterless conformance runner (`test/conformance/`) |
| `testdata/`, `demo/`, `assets/` | sample scenarios, the live demo, images |
| `hackathon/` | the build week's record: `fleet/` (every change's order, report, review and cost), `planning/` (the topic, the daily conversations), `SAPHAN.md`, `DESIGN.md` |
| `metadata.json`, `metadata.schema.json` | the hackathon's submission metadata, at the root until the hackathon ends |

## For technical readers

| | |
|---|---|
| [docs/SAMPLE_MODE.md](docs/SAMPLE_MODE.md) | try Proven with no network and no keys: the sample review, its findings and its trace |
| [DESIGN.md](hackathon/DESIGN.md) | Proven's design, the sponsors' technology, evidence sources, open questions |
| [SAPHAN.md](hackathon/SAPHAN.md) | the platform that runs the fleet, its roles and vocabulary |
| [AGENTS.md](AGENTS.md) | the rules every agent follows in this repository |
| [planning/](hackathon/planning/) | how the topic was chosen, and the fleet's answers |
| [fleet/](hackathon/fleet/) | orders, reports, reviews, decisions and the cost ledger, with their templates |

**Pre-existing components:** [Saphan Studio](https://saphan.ai/) runs the fleet; it is not part of this submission and not in this repository. The [Meterless engine specifications](https://github.com/Meterless/Meterless) (Apache-2.0) are open source; we build to them and do not copy their reference code. Model providers, tools and data sources are listed as they are used.

Before publication Marcin reviews every document. Credentials, internal host names and private infrastructure details are blacked out, and each document says where.

**Licence:** the code is Apache-2.0 ([LICENSE](LICENSE)); the documentation and the build record (the Markdown pages,
`docs/` and `hackathon/`) are CC BY 4.0 ([LICENSE-docs](LICENSE-docs)).
