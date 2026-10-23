# Proven — the design

*The design, written on 20 October 2026, before the build window (22–27 October 2026) opened. It is the plan, and it is kept as written so you can compare it with what was built. What is built so far, day by day, is in the README ([Where things stand](../README.md#where-things-stand-end-of-day-1-22-october-2026) and [Timeline](../README.md#timeline-what-each-push-brought)); how to run it and every setting are in [`docs/`](../docs/README.md). Each decision is recorded, with its reasons, as an ADR in [`fleet/decisions/`](../docs/adr/).*

This page is for technical readers and the hackathon's judges. The short version is in the [README](../README.md).

## The problem

A code review says "this is wrong" and rarely says why anyone should believe it. The evidence a good reviewer would need is spread over many places: the change itself, the project's history, its issues, the advisories published for its dependencies. Nobody reads all of them for every pull request.

## The rule: every finding carries evidence

A finding counts only with evidence that a reader can check again ([ADR-0002](../docs/adr/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)):

- **a test** that fails on the change and passes on the code before it, or
- **a query and the rows it returned**: one SQL query over the evidence sources below, written into the review next to the finding, so anyone can run it again and get the same rows.

A finding without either is dropped, not softened. This holds for every finding in every review: a single review, and two reviews that agree. Agreement is never a substitute for evidence.

## Independent reviewers

The owner's intended design, as of 19 October 2026:

- **The reviewer comes, where possible, from a different model family than the agent that wrote the code.** A reviewer from the same family tends to share the author's assumptions.
- **In selected cases, two independent reviews by different models.** Not every review; the cases are to be defined.
- **When two judgements conflict, the evidence behind both is examined**: what each test or query actually establishes, what it assumes, and whether it reproduces. Never a vote, never confidence. Evidence that does not settle it stays unsettled, and a person decides.

**How the reviewer is chosen** ([ADR-0008](../docs/adr/ADR-0008-reviewer-model-chosen-against-the-authors-model.md), 19 October 2026):

- Proven reads the author's model from what the change states in its commit messages or pull-request description. It does not try to detect it from the code.
- The reviewer's model comes from configuration. If the change names no author model, the configured default is used. The default is an NVIDIA model.
- A second reviewer model is configured too. For special kinds of review, security among them, the review is doubled: one review per configured model.

With only two models, model A for writing and NVIDIA, the doubled review's second reviewer may be model A itself ([ADR-0009](../docs/adr/ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)). That is still worth having: a fresh agent sees the code for the first time. The recommended setup has three models: A writes, NVIDIA reviews by default, B reviews the special cases. Each review says which model it used and whether that model is the author's.

Every reviewer model is configuration, NVIDIA's included ([ADR-0010](../docs/adr/ADR-0010-every-reviewer-model-is-configurable.md)). NVIDIA is only the default first reviewer. No setup is refused: A writing and A reviewing, or A and B without NVIDIA, both run, and each review says which kind of independence it had. Evidence is the one thing that is not configurable.

A change names its model in an `Assisted-by: <agent>:<model>` commit trailer by default, the Linux kernel's convention ([ADR-0011](../docs/adr/ADR-0011-how-a-change-names-its-model.md)). The trailer keys, the sources (commits, the pull-request description, or a record file for teams that keep AI attribution out of their history) and the table that maps a model to its family are all configurable. Proven only reads; it never writes a trailer.

Review classes are Saphan Studio's: security, backend, user interface, mobile and docs, one per change, chosen by what the change touches ([ADR-0012](../docs/adr/ADR-0012-review-classes-as-in-saphan-studio.md)). Security is doubled by default; which classes are doubled is configurable. For the hackathon the class is chosen by configured rules that map file paths to classes, never by a model; a change touching several classes gets the highest-risk one, and one that matches nothing gets a configurable fallback ([ADR-0014](../docs/adr/ADR-0014-review-class-chosen-by-configured-rules.md)). The classes are a closed list, checked when the configuration loads. An optional setting, off for the hackathon, lets a model raise a change's class, never lower it.

## Configuration

Proven hardcodes no settings ([ADR-0013](../docs/adr/ADR-0013-no-hardcoded-settings.md)). Every value it runs with, from models and endpoints to timeouts and review classes, comes from a configuration file or environment variables, each with a documented default. Every setting is validated when it loads; a wrong value stops proven with a message that says what is allowed, and never falls back silently. Rules are not settings: evidence for every finding, no exploit code, and a review that says which model reviewed it.

[ADR-0006](../docs/adr/ADR-0006-a-skeptic-from-another-model-family.md) adds a second, narrower independence inside one review: the Skeptic that tries to refute a finding comes from a different model family from the agent that made it.

## Sponsor technology, in two pictures

**How Proven is built** (22–27 October 2026):

```
  owner ──decides──▶ master ──one order per change──┐
    ▲                                               ▼
    │                         ┌───────────────────────────────────────┐
    │                         │ CURSOR                       (Build)  │
    │                         │ Cursor agents on Cursor's own models  │
    │                         │ write Proven's Go code and run tests  │
    │                         └───────────────────┬───────────────────┘
    │                                             ▼
    │   ┌──────────────────── Proven's repository ────────────────────┐
    │   │ METERLESS ENGINES, implemented in Go                        │
    │   │   H-MEM ............ memory store, retrieval                │
    │   │   World Model ...... world store, ingest, context builder   │
    │   │   checked by Meterless's own conformance suites             │
    │   │ ZETARIS adapter .... tested against our Zetaris platform    │
    │   │ NIGHTSHIFT adapter . tested against our Nightshift          │
    │   └─────────────────────────────┬───────────────────────────────┘
    │                                 ▼
    │                 the agent's report: model, run id, tests
    │                                 ▼
    └──merges── a review by another model, in a fresh context
```

**How Proven uses it** (one pull-request review):

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

## In a pipeline, and in the judges' lab

proven drops into a pipeline in place of a scanner such as Semgrep or Opengrep ([ADR-0016](../docs/adr/ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md)). It has `proven scan` for a local run and `proven ci` for a pipeline, which on a pull request reviews only the change. The flags keep Semgrep's names where they mean the same: `--text` (the default), `--json` and `--sarif`, each with a `--<format>-output=<file>` twin, `--error`, `--baseline-commit`. The exit codes keep Semgrep's numbers: 0 OK, 1 findings, 2 a fatal error. Every finding carries its evidence in every format, SARIF included, so the proof reaches a code-scanning dashboard with the finding. proven ships as a container image and as a GitHub Action.

The judges get a lab ([ADR-0017](../docs/adr/ADR-0017-a-gitea-lab-for-the-judges.md)): one Docker Compose file that starts Gitea, a Gitea Actions runner and a demo repository with an open pull request. One command, no keys, and proven's review appears on the pull request in Gitea's web page. The workflow there is the same file a GitHub user would use. A Jenkins example comes only if time is left.

## The agents inside Proven

These are proven's own runtime agents, not the fleet that builds proven:

| agent | what it does |
|---|---|
| Planner | reads the change, says what it claims to do and where the risk is, and asks the others for evidence |
| Investigator | turns each risk into queries over the evidence sources and returns the rows |
| Prover | writes the test that fails on the change and passes on the base, and runs both |
| Skeptic | a model from a different family, whose only job is to refute each finding |
| Judge | weighs the evidence behind each finding and each refutation, and sends weak findings back for a second pass |
| Writer | writes the review: each finding with its proof, its sources and what the review cost |

The run is bounded: a fixed number of rounds, a timeout on every outside call, and a partial result instead of a hang.

## Where the sponsors' technology does the work

- **Zetaris: the evidence layer** ([ADR-0002](../docs/adr/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)). Every evidence source is one table behind Zetaris, which runs a query where the data already lives, without copying it first. The Investigator asks one question across all of them in one query. For example: *"Which dependency this change adds or upgrades has a published advisory for that exact version, is on the list of vulnerabilities known to be used in attacks, and has a high chance of being used in one?"* The query and its rows go into the review as the proof.
- **Meterless: the reviewer's memory and state** ([ADR-0003](../docs/adr/ADR-0003-meterless-engines-in-go.md)). Meterless publishes open specifications of agent engines, each with a conformance suite. We build two of them in Go:
  - **H-MEM**, the reviewer's memory. A finding that was refuted in one pull request does not come back in the next one. A maintainer's ruling ("we do not validate here, on purpose") is remembered.
  - **World Model**, the state of one review as a stream of events: every claim, piece of evidence, refutation and verdict. The trace in `logs/` is that stream, not text written next to it.
  - Our Go builds are checked with Meterless's own conformance suites through a thin TypeScript adapter. That this works is not measured yet.
- **NVIDIA: a second model family** ([ADR-0006](../docs/adr/ADR-0006-a-skeptic-from-another-model-family.md)). The Skeptic runs on an NVIDIA Nemotron model through NVIDIA's hosted API. Hosted NVIDIA embeddings for H-MEM's search are a candidate, not a decision.
- **NVIDIA: the guardrail, on by default** ([ADR-0005](../docs/adr/ADR-0005-guardrail-on-by-default.md)). Every prompt and answer passes NVIDIA's Nemotron 3.5 Content Safety first. The same model runs in two places behind one adapter: on our own machine, and on NVIDIA's hosted API ([build.nvidia.com](https://build.nvidia.com/nvidia/nemotron-3.5-content-safety)). A judge with an NVIDIA API key keeps the guardrail on; a setting turns it off for anyone without one.
- **Cursor: the development environment.** The four Cursor agents of the team write the product code.

## The evidence sources

The first set ([ADR-0004](../docs/adr/ADR-0004-first-evidence-sources.md)):

| source | what it answers |
|---|---|
| [OSV](https://osv.dev) | which package versions an advisory covers, per ecosystem |
| [CISA KEV](https://www.cisa.gov/known-exploited-vulnerabilities-catalog) | whether a vulnerability is known to be used in real attacks |
| [FIRST EPSS](https://www.first.org/epss/) | how likely a vulnerability is to be used in an attack |
| the repository's own history | what was changed, reverted or fixed before, and in which commit |

Next, if the build leaves time: the NVD CVE records, the CWE list of weakness classes, and OpenSSF Scorecard.

All data is downloaded by scripts written in the build window, from 22 October. Small samples are kept in this repository, so `SAMPLE_MODE=true` runs the whole review with no network and no keys; the same queries then run on a local engine over the samples instead of on Zetaris. The licence of every source is listed here once its script exists.

Proven writes no exploit code ([ADR-0007](../docs/adr/ADR-0007-no-exploit-code.md)). A finding about a vulnerable dependency is proved by the query and its rows, never by an attack.

## Why Go

Proven is open source, and one part of it will also be used inside Saphan Studio. The owner chose Go for that reason ([ADR-0001](../docs/adr/ADR-0001-go-behind-adapters.md)). Every outside service (Zetaris, Meterless's suites, NVIDIA, the model providers) sits behind an adapter: an outage during judging degrades the review instead of stopping it, and a judge without our keys can still run it in sample mode.

## What proven does not claim

- It does not prove a whole program correct. It supports individual findings with a test or a query.
- It does not find every vulnerability. It reports what its evidence sources cover.
- A finding that cannot be supported is dropped, so an empty review means "nothing proven", not "nothing wrong".

## Not known yet

- **Zetaris, in practice.** Access arrived on 20 October: a hosted instance, and the whole platform as a Docker Compose stack, which we are setting up on a machine of our own ([order](fleet/orders/TRAIN_ORDER_proven-zetaris-local.md)). proven reaches it through its REST API, so proven carries no Java ([ADR-0015](../docs/adr/ADR-0015-zetaris-through-its-rest-api.md)). Neither instance has been called from our side yet.
- **Meterless, settled on 21 October.** The sponsor's guide says the engines are reference architectures that each team implements in its own language, with no hosted service and no key, which is what [ADR-0003](../docs/adr/ADR-0003-meterless-engines-in-go.md) does. Every engine calls models only through one injected generator function, and writes its audit, traces and provenance as JSONL in `logs/`.
- **The hosted guardrail, in detail.** NVIDIA's model page names the endpoint and the answer format, but its code sample leaves the model id blank, and it states no free tier and no rate limit. The first live call settles both.
- **The conformance adapter.** Whether a Go engine passes Meterless's suites through a TypeScript adapter is the first thing we measure.
- **The checks of each review class.** Saphan Studio's classes are adopted ([ADR-0012](../docs/adr/ADR-0012-review-classes-as-in-saphan-studio.md)); what each one checks in proven is built from 22 October.
