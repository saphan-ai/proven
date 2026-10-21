# fleet/decisions/

> **A note on the owner's English (2026-10-19).** The owner asked us to correct the grammar of his own English in this
> repository. Where he wrote in English, his quoted words have been lightly corrected for grammar only; nothing was added
> to their meaning. Where he wrote in Polish, the quotes are translations, as each one says.


This folder holds two kinds of decision, and both are the owner's:

- **Design decisions**, `ADR-<number>-<slug>.md`: how proven is built and why. One file per decision, written when the decision is taken.
- **Decisions on a change**, `DECISION_<slug>_r<n>.md`: the owner's ruling on one change, after its review.

## Design decisions (ADRs)

| ADR | decision | status |
|---|---|---|
| [0001](ADR-0001-go-behind-adapters.md) | proven is written in Go, and every outside service sits behind an adapter | accepted, 2026-10-17 |
| [0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) | a finding counts only with evidence: a test, or a query through Zetaris with the rows it returned | accepted, 2026-10-17 |
| [0003](ADR-0003-meterless-engines-in-go.md) | the reviewer's memory and state are Meterless's H-MEM and World Model, built in Go | accepted, 2026-10-17 |
| [0004](ADR-0004-first-evidence-sources.md) | the first evidence sources: OSV, CISA KEV, FIRST EPSS, the repository's own history | accepted, 2026-10-17 |
| [0005](ADR-0005-guardrail-on-by-default.md) | a guardrail on every prompt and answer by default, local or hosted, with a switch to turn it off | accepted, 2026-10-18 |
| [0006](ADR-0006-a-skeptic-from-another-model-family.md) | every finding faces a Skeptic from another model family, NVIDIA Nemotron; the evidence settles it, never a vote | accepted, 2026-10-18 |
| [0007](ADR-0007-no-exploit-code.md) | no exploit code; a vulnerability finding is proved by a query and its rows | accepted, 2026-10-18 |
| [0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) | the reviewer's model is chosen against the author's model stated in the change; NVIDIA by default; some reviews doubled | accepted, 2026-10-19 |
| [0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md) | two reviewer models; a doubled review may use the author's own model as a fresh second reviewer; A writes, NVIDIA reviews, B for special cases is best | accepted, 2026-10-19 |
| [0010](ADR-0010-every-reviewer-model-is-configurable.md) | every reviewer model is configurable, NVIDIA's included; NVIDIA is only the default; no setup is blocked | accepted, 2026-10-19 |
| [0011](ADR-0011-how-a-change-names-its-model.md) | a change names its model in an `Assisted-by:` trailer by default; sources, keys and the family table are configurable | accepted, 2026-10-19 |
| [0012](ADR-0012-review-classes-as-in-saphan-studio.md) | Saphan Studio's review classes: security, backend, user interface, mobile, docs; security doubled by default | accepted, 2026-10-19 |
| [0013](ADR-0013-no-hardcoded-settings.md) | no setting is hardcoded; every value has a documented default; rules are not settings | accepted, 2026-10-19 |
| [0014](ADR-0014-review-class-chosen-by-configured-rules.md) | a change's review class is chosen by configured path rules, not a model; the highest-risk class wins | accepted, 2026-10-19 |
| [0015](ADR-0015-zetaris-through-its-rest-api.md) | the Zetaris adapter speaks REST, not JDBC; proven carries no Java | accepted, 2026-10-20 |
| [0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md) | a drop-in replacement for Semgrep and Opengrep in a pipeline: their flags, exit codes and text, JSON and SARIF outputs | accepted, 2026-10-20 |
| [0017](ADR-0017-a-gitea-lab-for-the-judges.md) | a Docker Compose lab with Gitea and Gitea Actions, configured, for the judges; Jenkins only if time is left | accepted, 2026-10-20 |
| [0018](ADR-0018-meterless-engines-minimum-and-their-contract.md) | H-MEM and World Model at their minimum, one injected generator, JSONL in `logs/`; Markovian only if time is left | accepted, 2026-10-21 |
| [0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) | Meterless Nightshift is the primary AI provider; NVIDIA's models through it by default; other routes stay settings | accepted, 2026-10-21 |

**Rules for an ADR**

- **The decision is the owner's**, and the ADR quotes his words. An agent may write it down; it cannot decide it.
- **One decision per file**, numbered in order. A number is never reused.
- **An ADR is not rewritten when the decision changes.** A new ADR supersedes it, and the old one's status says `superseded by ADR-<number>`.
- **What is not known yet is said so**, under Consequences.

**Template**

```markdown
# ADR-<number>: <the decision, in one sentence>

**Status:** proposed | accepted | superseded by ADR-<number> · **Date:** <the day it was decided>
**Decided by:** the owner · **Written by:** <agent and model, date>
**Supersedes / Extends / Follows:** <other ADRs, where they apply>
**Trigger:** <what made the decision necessary, with a link to the conversation>

## Context
<the situation and the forces, with the owner's words>

## Decision
<numbered, each one a rule someone can check>

## Consequences
<what follows, good and bad, and what is not known yet>
```

## Decisions on a change

A decision is the owner's. No agent can merge, and no agent's "yes" counts: a reviewer's accept is advice the owner reads, not a decision.

**Rules every decision follows**

- **One of four outcomes:** `accept`, `accept-with-conditions`, `reject` or `close-without-shipping`.
- **The reason is written by the person who decides.** The owner writes it himself, in his own words. A decision without a reason is not published.
- **Conditions are named.** If the owner accepts with conditions, each one says what closes it and in which later change.
- **Every decision states what the change cost**, to the cent, over all its rounds: who built it on which model, how many repair rounds it took, and how many BLOCKING and OWED conditions the reviews found. The same line goes into [`../LEDGER.md`](../LEDGER.md).
- **A rejected change keeps its documents.** The order, the status report and the review of every round stay in this repository, including the ones that were sent back.

**Template**

```markdown
# DECISION — <slug> · round <n>

- outcome: <accept | accept-with-conditions | reject | close-without-shipping>
- decided by: the owner
- review: fleet/reviews/REVIEW_<slug>_r<n>.md
- merged as: <commit — or "not merged">
- rounds: <n> (<n-1> repair rounds)
- built by: <model per round, e.g. r1 model-a · r2 model-a>
- reviewed by: <model per round>
- conditions over all rounds: <n> BLOCKING · <n> OWED · <n> NOTE
- cost of the change: <$0.00> = <$0.00> building + <$0.00> reviewing
- decided: <date and time, UTC>

## Reason
<the owner's own words>

## Conditions
<each condition and the change that closes it — or "none">
```
