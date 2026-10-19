# CONVERSATION — the front page for a wider audience, 19 October 2026

This page continues [`CONVERSATION_2026-10-18.md`](CONVERSATION_2026-10-18.md). For every hop it says who spoke, to whom, what the advisor reasoned and did, and which document came out of it. The owner's words are in English; where he wrote in Polish, they are given in the advisor's translation.

## Hop by hop

### 1 · owner → advisor: make the front page understandable

The owner handed the advisor a written order, titled *"Make Proven's GitHub front page understandable to a wider audience"*, with one word of his own (translated): *"Do it."* Its first lines:

> *"The owner clarified that the current material was written technically. This revision must change the reader's point of view, not just shorten technical paragraphs. Assume a curious reader with no knowledge of agent orchestration: a journalist, founder, prospective user or general technology reader."*

The reason was the day itself: the owner's and Saphan's first public announcements, at 15:00 UTC.

**What the advisor read first.** This repository as it stood (README, SAPHAN.md, the diary, the fleet's answers, `fleet/`, the ledger, the decisions and both infrastructure orders), the owner's public biography at [wato.org](https://wato.org/), the product site [saphan.ai](https://saphan.ai/), and five pages of the technical documentation at [docs.saphan.ai](https://docs.saphan.ai/docs): egress, agent security, the sandbox matrix, session recording, and the documented limits. Both infrastructure orders were still drafts, not yet approved, so the new status table says so.

**The advisor's reasoning.** The old front page answered a judge's questions in the order a judge asks them. A first-time visitor has different questions: who is doing this, why, what is unusual, and what can I follow. So the page now starts with the person and the story: commercial telco work, then Saphan Studio, then Proven. The technical design is one click away, not gone.

Checking the product claims against the documentation changed several sentences. The old SAPHAN.md said that network access is closed by default, that our sandbox always wraps the vendor's, and that every order is signed. The documentation is more precise: the network is limited to a run's model and its registries unless a policy opens more, and some kinds of run are refused instead; confinement covers some agent and system combinations, not all; signature enforcement is off by default. SAPHAN.md now says what the documentation says, links to its limits, and keeps this fleet's own, stricter policy separate from the product's defaults.

**What the advisor did.**

- **[`../README.md`](../README.md)**, rewritten: who builds Proven and why, the introduction video, where things stand on 19 October, one example from the fleet's earlier work, what Proven is meant to do (marked as planned), who builds it (and how that differs from the agents inside Proven), how to follow the work, and contact.
- **[`../DESIGN.md`](../DESIGN.md)**, new: the technical design that used to be on the front page, plus the owner's intended reviewer independence, and the one question about it that is still open.
- **[`../SAPHAN.md`](../SAPHAN.md)**, reorganized: what Saphan Studio does today, with links and limits; how this fleet uses it; the roles and one change from start to finish; skills; the fleet's rules and orchestration in broad groups only; the vocabulary; the scale figures with their date and source.
- **[`README.md`](README.md)** of `planning/`: the story of how the topic was chosen, moved here from the front page unchanged.
- **[`../fleet/README.md`](../fleet/README.md)** and **[`../fleet/LEDGER.md`](../fleet/LEDGER.md)**: what a recorded cost is, what "not measured" means, and what the figures leave out.

The original conversations, the fleet's answers and the decisions are unchanged.

### 2 · owner → advisor: which source is right

The advisor had reported two places where public sources disagreed. The owner asked what they were, and gave the hackathon's event page.

- **Network access.** The product site said that limiting a run's network was opt-in per run, with no default yet. The technical documentation said that every run is limited by default. The engine's own documentation settled it: since 8 October 2026, a run that declares no policy gets its backend's baseline (its model and its package registries) and nothing else.
- **Cursor.** The documented release lists Cursor as not built yet, while the development version the fleet uses has it. [`../SAPHAN.md`](../SAPHAN.md) says exactly that.

> *"The docs are right. The site needs fixing. Fix it right away."* (translated from Polish)

**What the advisor did.** It rewrote the network passages of [saphan.ai](https://saphan.ai/) in its four languages (English, Polish, Thai and Arabic) on the home, controls and security pages, to say what the documentation says, including what is not shipped yet. The owner built, checked and deployed the site. The README's hackathon link now points to the [event page](https://hackathon.genai.works/event/open-agent-hackathon-2026).

### 3 · owner → advisor: how the reviewer is chosen

The rewrite had left one design question open: proven should review a change with a model from a different family than the one that wrote it, but how does it know which model wrote it?

> *"ADR-0008: the model is stated in the commit and PR notes. If not specified, the default model from config is taken. Default it is NVIDIA. A second model is configured also: for special types of reviews, like security, the review is doubled."*

**What the advisor did.** It wrote [ADR-0008](../fleet/decisions/ADR-0008-reviewer-model-chosen-against-the-authors-model.md). Proven reads the author's model from what the change states and does not guess it. The reviewer's model comes from configuration, NVIDIA by default. A second configured model doubles special reviews such as security. Evidence is still required for every finding. Three details are listed as not decided yet: the exact form in which a change names its model, the case where the author's family equals the default reviewer's, and which reviews besides security are doubled. [`../DESIGN.md`](../DESIGN.md) and the front page now say the same.

## Not run

- Links were checked inside this repository. Links to other websites were checked by reading the pages, not by a link checker.
- How the page looks on a phone was not checked on a device.
