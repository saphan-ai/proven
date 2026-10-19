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

### 4 · owner → advisor: when the reviewer is the author's own model

ADR-0008 had left open what happens when a configured reviewer model is the same as the model that wrote the change.

> *"Two second models can be configured. Then, in the scenario where a customer has model A and NVIDIA, the review is done by NVIDIA; this satisfies the independent-reviewing-model requirement. For special classes like security, the same model A is used. It is still valuable, because the review is done by a fresh agent who sees the code for the first time. The best setup will be: model A writes the code, NVIDIA reviews by default, model B reviews in special cases like security."*

**What the advisor did.** It wrote [ADR-0009](../fleet/decisions/ADR-0009-two-reviewer-models-and-the-same-model-fallback.md): two reviewer models, a default (NVIDIA) and one for special reviews. With only two models, a doubled review's second reviewer may be the author's own model, and it is still a fresh agent. Three models is the recommended setup. Each review says which model it used and whether that is the author's, so nobody mistakes a fresh agent for a different family. ADR-0008 now points to ADR-0009 for the case it settles. The one case still open is a change written by the default reviewer's own family. [`../DESIGN.md`](../DESIGN.md) and the front page follow.

### 5 · owner → advisor: NVIDIA is a default, not a requirement

> *"And we make NVIDIA also configurable. Maybe some customers will want to use coding A, reviewing A, without NVIDIA, or A and B. We don't block this. NVIDIA is just a model like the others, just the default first reviewing model. Got that?"*

**What the advisor did.** It wrote [ADR-0010](../fleet/decisions/ADR-0010-every-reviewer-model-is-configurable.md). Every reviewer model is configuration, NVIDIA's included, and NVIDIA is only the value proven ships with. No setup is refused: A writing and A reviewing, A and B without NVIDIA, or the recommended A, NVIDIA and B. Independence is recommended and reported in every review, never enforced. Evidence is the one thing no setting changes. This also closes the case ADR-0009 had left open, a change written by the default reviewer's own family: the review runs and says so. [`../DESIGN.md`](../DESIGN.md) and the front page follow.

### 6 · owner → advisor: the last open points, and a rule for every setting

The owner answered the three points the earlier decisions had left open:

> *"1. Propose something and make it configurable. Look at industry best practices.*
> *2. Take the same classes that we have in Saphan Studio today.*
> *3. The same as class A-A: it's NVIDIA-NVIDIA. Still valuable, because the review is done by an agent with a fresh context, with no contact with the code before. A limited scenario, yet still useful."*

While the advisor was writing, he added:

> *"By the rule: we NEVER hardcode any settings. All is configurable. Write this as an ADR in Proven."*

**What the advisor found.** Industry practice for naming the model behind a change is a Git commit trailer. The Linux kernel's guidance for AI coding assistants asks for `Assisted-by:` and forbids AI agents from adding `Signed-off-by:`, which only a person may give. Saphan Studio's review classes are security, backend, user interface, mobile and docs, one per change, chosen by what the change touches; security reviews are done twice.

**What the advisor did.**

- [ADR-0011](../fleet/decisions/ADR-0011-how-a-change-names-its-model.md): `Assisted-by: <agent>:<model>` by default. The sources, the trailer keys and the model-family table are configurable, including a record file for teams that keep AI attribution out of their history, as our own fleet does. Proven only reads trailers.
- [ADR-0012](../fleet/decisions/ADR-0012-review-classes-as-in-saphan-studio.md): Saphan Studio's five review classes; security doubled by default; which classes are doubled is configurable.
- [ADR-0010](../fleet/decisions/ADR-0010-every-reviewer-model-is-configurable.md) now quotes the owner on the NVIDIA-NVIDIA case.
- [ADR-0013](../fleet/decisions/ADR-0013-no-hardcoded-settings.md): no setting is hardcoded; every value has a documented default; the run records the configuration it used. It separates settings from rules: evidence for every finding, no exploit code, and a review that names its model are what proven is, not values to tune.

[`../DESIGN.md`](../DESIGN.md) has a new "Configuration" section, and the front page counts thirteen decisions.

## Not run

- Links were checked inside this repository. Links to other websites were checked by reading the pages, not by a link checker.
- How the page looks on a phone was not checked on a device.
