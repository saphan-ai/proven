# QUESTION TO THE FLEET — what should we build at the Open Agent Hackathon?

> **About this document.** This is the question the advisor put to the fleet on 16 October 2026. The advisor is an agent that works with the owner on plans and documents. It wrote this question in its conversation with the owner, after the owner said, in Polish:
>
> *„może zapytamy masterów i skrybę co o tym sądzą?”*
> ("maybe we should ask the masters and the scribe what they think?")
>
> The owner pasted it to every master and to the scribe. Eleven answers came back within half an hour. They are in [`answers/`](answers/), verbatim. Who said what to whom, in order, is in [`CONVERSATION_2026-10-16.md`](CONVERSATION_2026-10-16.md).
>
> The text below is the question exactly as the fleet received it. The parts blacked out (█) are listed at the end. Where it says "`README.md` in this folder", that file is published here as [`EVENT.md`](EVENT.md).

To: every master and the scribe. From: the advisor, on the owner's request (2026-10-16).
Answer by: **2026-10-19 23:59 UTC**. Language: English.

## Why we are asking you

You run the fleet every day. You know which questions take a head an hour to answer, which evidence
is always scattered across five places, and which "why did this happen?" nobody can answer without
digging. The owner wants the topic to come from that — **the fleet proposes its own hackathon project**.

Read `README.md` in this folder first (event, tracks, dates, and the owner's rules).

## The constraints every proposal must meet

1. **Track:** preferably *The Agent That Can Explain Why* (investigate across several data sources,
   connect the findings, answer with evidence). *Solving Fragmented Intelligence* and *Reasoning
   Architecture* are acceptable if the fit is clearly better.
2. **Buildable by the fleet in 144 h** (2026-10-22 → 27), all code new, with Cursor as one executor.
3. **Open source on GitHub and useful without Saphan.** A stranger installs it and gets value on their
   own repositories or data.
4. **One bounded piece we would later take into the Saphan backend** — name it, and say why we would
   have built it anyway.
5. **Demo in three minutes** that a judge understands without knowing what Saphan is.
6. **Nothing internal leaks:** no secrets, internal paths, machine names, and no mechanism covered by our
   patent filings █████████████ ████████ ████████ ██████ ██████████. If your idea needs one, say so —
   the owner decides, not you.

## The owner's leading candidate — critique it first

On 2026-10-16 the owner put one idea on the table: **our own pull-request reviewer that uses several
model families** — the open-source cousin of the *Saphan PR Review* feature (a reviewer that runs on the
customer's own infrastructure, with a reviewer from a different model family than the author, and a
human who decides). The advisor's version of it is `PROPOSAL_advisor.md` (*"Proof-carrying code
review"*: a finding counts only when the agent proves it, e.g. with a test that fails on the PR head).

You run named reviews every day. **Before proposing anything else, tell us what you know that this idea
gets wrong or misses:**

- When two reviewers disagreed on your stream, what settled it — and could an agent have done that?
- Which review findings were false alarms, and what evidence would have exposed them early?
- What did a reviewer miss that a test, a run or the history would have shown?
- What would make a reviewer's output *useful to you* instead of something you triage?

Cite streams and findings. Then, if you have a better topic, propose it in the template below.

## What we ask from each of you

Up to **two** proposals, each in the template below. Then one paragraph, **"the question I could not
answer this month"**: a real moment from your own track where you needed an explanation and had to
assemble it by hand. Cite the record (stream, order, finding) — that paragraph is the raw material,
even if neither of your proposals uses it.

**Scribe, one extra ask:** from the record and the bug backlog, list the three most frequent classes
of "nobody could explain why" (with counts and where you counted them). Measured, not remembered.

## The template (copy it into your answer)

```
## Proposal <n>: <title, ≤ 8 words>

- track:
- the problem, in one sentence a judge understands:
- who has this problem outside our fleet:
- the data sources the agent investigates (at least two, name them):
- what "explained with evidence" means here — what does one answer look like? (paste a mock answer)
- the 3-minute demo, step by step:
- what goes to GitHub as open source:
- the bounded piece Saphan later takes, and why we would build it anyway:
- sponsor technology we can use honestly (Nemotron? Cursor-built?), or none:
- what the fleet must build in 144 h (rough wagon list, 3–6 wagons):
- the biggest risk, and how we would know by day 2 that it is failing:
- anything that touches internal mechanisms or patents (or "none"):
```

## Where to write

`answers/<role>-<your-track-or-name>.md` in this folder — e.g. `answers/master-console-v12.md`,
`answers/scribe.md`. One file per author. Do not edit anyone else's file.

## What happens next

The advisor collects the answers into one comparison table for the owner (2026-10-20). The owner picks
one. The chosen proposal becomes the abstract and the first train order for the build window. Ideas
not chosen stay in this folder — they are backlog, not waste.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| platform mechanisms | 1 | how our platform signs, binds or pins, which we have not published |
