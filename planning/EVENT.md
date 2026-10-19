# Open Agent Hackathon 2026 — what we enter, and the rules we keep

> **About this document.** The advisor's working README for the hackathon, as it stood on 16 October 2026 when the question went to the fleet. It is the `README.md` the question tells the fleet to read first. Published verbatim; the parts blacked out (█) are listed at the end. The whole exchange, step by step, is in [`CONVERSATION_2026-10-16.md`](CONVERSATION_2026-10-16.md).

Owner: Marcin. Written 2026-10-16. Language of everything in this folder: English.

## The event

- **Organizer:** GenAI.Works, platform HackOS — https://hackathon.genai.works/event/open-agent-hackathon-2026
- **Registration closes:** 2026-10-22 23:59 UTC (the owner registers). **Onboarding call:** 2026-10-21.
- **Build window:** 2026-10-22 00:00 → 2026-10-27 23:45 UTC (144 h). **Results:** 2026-11-06 16:00 UTC.
- **Teams:** 1–5 builders, online.

### Tracks (verbatim names from the submission form)

| track | what it asks for |
|---|---|
| **Solving Fragmented Intelligence** | find and work across scattered data sources |
| **The Agent That Can Explain Why** | agents that investigate complex questions across several data sources, connect the findings, and produce answers or recommendations **backed by evidence** |
| **Reasoning Architecture** | a new way to turn data, knowledge, memory and reasoning into dependable decisions and actions |
| **Wildcard [Tinkerer]: Bring Your Own Project** | extend an existing project; must integrate **both** Zetaris and Meterless; only the new work is scored |

Our natural fit is **The Agent That Can Explain Why** — evidence, explainability, decisions a person can check.

### Judging (100 points, weights not published)

Technical implementation · Originality · Demo clarity · Usability · Sponsor technology use.
Sponsor awards we can reach without bending the project: **Best use of NVIDIA Nemotron** (an open model
family — fits our "a second model family reviews the first"), **Best use of a Cursor-built product**.

## The owner's rules for our entry (decided 2026-10-12, unchanged)

1. **The fleet builds it.** The owner writes no line of project code. His role: specs, decisions,
   merges, the submission.
2. **Cursor joins the fleet as an executor** for this event (the organizers confirmed: code written by
   Cursor agents counts as Cursor use; describe how the agents produced it and how the owner reviewed
   and approved it).
3. **All project code is new, written 2026-10-22 → 27.** Saphan Studio may run the fleet, but it is a
   **pre-existing component** — named as such in the submission and not judged.
4. **Everything is recorded.** The process record (orders, returns, reviews, decisions) may be published
   after a review.
5. **Nothing leaves without that review:** no secrets, no internal paths, no machine names, and **no
   mechanism covered by our patent filings** █████████████ ████████ ████████ ██████ ██████████.
6. **Owner time:** about one hour of decisions per day during the window.

## What we want out of it (the owner's goal for this round)

- **An open-source project on GitHub** that stands on its own: useful to anyone, installable in minutes,
  with a demo a judge understands in three minutes.
- **A part we later take into the Saphan backend.** Design the project so that one clearly bounded
  piece (a schema, a library, an adapter) is what we would have built anyway — and keep the boundary
  explicit from day one, so the open-source part never needs Saphan and Saphan never needs the demo.
- Proof and content, not the prize: the story *"the fleet entered a hackathon and the owner wrote no
  code"* is itself the post.

## The owner's leading candidate (2026-10-16)

*"Our own code review with different models"* — the open-source cousin of the *Saphan PR Review*
feature. The advisor's version: `PROPOSAL_advisor.md` (*Proof-carrying code review*). The fleet is asked to
critique it first and to propose alternatives second.

## How we choose the topic

1. The fleet proposes — `QUESTION_TO_THE_FLEET.md` (this folder). Masters and the scribe answer into
   `answers/`.
2. The advisor's own proposal is in `PROPOSAL_advisor.md` — one input among the others, not the answer.
3. The owner picks one by **2026-10-20**, so the first order is ready before the onboarding call.
4. The chosen topic gets an abstract (draft in the advisor's proposal) and a first train order for the
   window.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| platform mechanisms | 1 | how our platform signs, binds or pins, which we have not published |
