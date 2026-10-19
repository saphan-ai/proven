# Proven — the design

*The design as of 19 October 2026. Nothing here is built yet: the build window is 22–27 October 2026. Every statement on this page is a plan, not a result. Each decision is recorded, with its reasons, as an ADR in [`fleet/decisions/`](fleet/decisions/).*

This page is for technical readers and the hackathon's judges. The short version is in the [README](README.md).

## The problem

A code review says "this is wrong" and rarely says why anyone should believe it. The evidence a good reviewer would need is spread over many places: the change itself, the project's history, its issues, the advisories published for its dependencies. Nobody reads all of them for every pull request.

## The rule: every finding carries evidence

A finding counts only with evidence that a reader can check again ([ADR-0002](fleet/decisions/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)):

- **a test** that fails on the change and passes on the code before it, or
- **a query and the rows it returned**: one SQL query over the evidence sources below, written into the review next to the finding, so anyone can run it again and get the same rows.

A finding without either is dropped, not softened. This holds for every finding in every review: a single review, and two reviews that agree. Agreement is never a substitute for evidence.

## Independent reviewers

The owner's intended design, as of 19 October 2026:

- **The reviewer comes, where possible, from a different model family than the agent that wrote the code.** A reviewer from the same family tends to share the author's assumptions.
- **In selected cases, two independent reviews by different models.** Not every review; the cases are to be defined.
- **When two judgements conflict, the evidence behind both is examined**: what each test or query actually establishes, what it assumes, and whether it reproduces. Never a vote, never confidence. Evidence that does not settle it stays unsettled, and a person decides.

**How the reviewer is chosen** ([ADR-0008](fleet/decisions/ADR-0008-reviewer-model-chosen-against-the-authors-model.md), 19 October 2026):

- Proven reads the author's model from what the change states in its commit messages or pull-request description. It does not try to detect it from the code.
- The reviewer's model comes from configuration. If the change names no author model, the configured default is used. The default is an NVIDIA model.
- A second reviewer model is configured too. For special kinds of review, security among them, the review is doubled: one review per configured model.

Still open: the exact form in which a change names its model; what happens when the stated author model is from the same family as the default reviewer; and which kinds of review besides security are doubled.

[ADR-0006](fleet/decisions/ADR-0006-a-skeptic-from-another-model-family.md) adds a second, narrower independence inside one review: the Skeptic that tries to refute a finding comes from a different model family from the agent that made it.

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

- **Zetaris: the evidence layer** ([ADR-0002](fleet/decisions/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)). Every evidence source is one table behind Zetaris, which runs a query where the data already lives, without copying it first. The Investigator asks one question across all of them in one query. For example: *"Which dependency this change adds or upgrades has a published advisory for that exact version, is on the list of vulnerabilities known to be used in attacks, and has a high chance of being used in one?"* The query and its rows go into the review as the proof.
- **Meterless: the reviewer's memory and state** ([ADR-0003](fleet/decisions/ADR-0003-meterless-engines-in-go.md)). Meterless publishes open specifications of agent engines, each with a conformance suite. We build two of them in Go:
  - **H-MEM**, the reviewer's memory. A finding that was refuted in one pull request does not come back in the next one. A maintainer's ruling ("we do not validate here, on purpose") is remembered.
  - **World Model**, the state of one review as a stream of events: every claim, piece of evidence, refutation and verdict. The trace in `logs/` is that stream, not text written next to it.
  - Our Go builds are checked with Meterless's own conformance suites through a thin TypeScript adapter. That this works is not measured yet.
- **NVIDIA: a second model family** ([ADR-0006](fleet/decisions/ADR-0006-a-skeptic-from-another-model-family.md)). The Skeptic runs on an NVIDIA Nemotron model through NVIDIA's hosted API. Hosted NVIDIA embeddings for H-MEM's search are a candidate, not a decision.
- **NVIDIA: the guardrail, on by default** ([ADR-0005](fleet/decisions/ADR-0005-guardrail-on-by-default.md)). Every prompt and answer passes NVIDIA's Nemotron 3.5 Content Safety first. The same model runs in two places behind one adapter: on our own machine, and on NVIDIA's hosted API ([build.nvidia.com](https://build.nvidia.com/nvidia/nemotron-3.5-content-safety)). A judge with an NVIDIA API key keeps the guardrail on; a setting turns it off for anyone without one.
- **Cursor: the development environment.** The four Cursor agents of the team write the product code.

## The evidence sources

The first set ([ADR-0004](fleet/decisions/ADR-0004-first-evidence-sources.md)):

| source | what it answers |
|---|---|
| [OSV](https://osv.dev) | which package versions an advisory covers, per ecosystem |
| [CISA KEV](https://www.cisa.gov/known-exploited-vulnerabilities-catalog) | whether a vulnerability is known to be used in real attacks |
| [FIRST EPSS](https://www.first.org/epss/) | how likely a vulnerability is to be used in an attack |
| the repository's own history | what was changed, reverted or fixed before, and in which commit |

Next, if the build leaves time: the NVD CVE records, the CWE list of weakness classes, and OpenSSF Scorecard.

All data is downloaded by scripts written in the build window, from 22 October. Small samples are kept in this repository, so `SAMPLE_MODE=true` runs the whole review with no network and no keys; the same queries then run on a local engine over the samples instead of on Zetaris. The licence of every source is listed here once its script exists.

Proven writes no exploit code ([ADR-0007](fleet/decisions/ADR-0007-no-exploit-code.md)). A finding about a vulnerable dependency is proved by the query and its rows, never by an attack.

## Why Go

Proven is open source, and one part of it will also be used inside Saphan Studio. The owner chose Go for that reason ([ADR-0001](fleet/decisions/ADR-0001-go-behind-adapters.md)). Every outside service (Zetaris, Meterless's suites, NVIDIA, the model providers) sits behind an adapter: an outage during judging degrades the review instead of stopping it, and a judge without our keys can still run it in sample mode.

## What proven does not claim

- It does not prove a whole program correct. It supports individual findings with a test or a query.
- It does not find every vulnerability. It reports what its evidence sources cover.
- A finding that cannot be supported is dropped, so an empty review means "nothing proven", not "nothing wrong".

## Not known yet

- **Zetaris access.** The instance and its documentation are promised to participants before 22 October. Until they arrive, we do not know which interface we will use (SQL driver, REST or MCP).
- **Meterless keys.** The organizers' example settings name a Meterless API key. The public Meterless repository describes no hosted service and needs no key. We will ask on the sponsor's help channel.
- **The hosted guardrail, in detail.** NVIDIA's model page names the endpoint and the answer format, but its code sample leaves the model id blank, and it states no free tier and no rate limit. The first live call settles both.
- **The conformance adapter.** Whether a Go engine passes Meterless's suites through a TypeScript adapter is the first thing we measure.
- **Reviewer independence, in detail.** How a change names its model, the same-family case, and which reviews are doubled. See "Independent reviewers" above.
