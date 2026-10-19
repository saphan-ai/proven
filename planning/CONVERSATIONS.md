# CONVERSATIONS — the diary

> **A note on the owner's English (2026-10-19).** The owner asked us to correct the grammar of his own English in this
> repository. Where he wrote in English, his quoted words have been lightly corrected for grammar only; nothing was added
> to their meaning. Where he wrote in Polish, the quotes are translations, as each one says.


This page is an index of every conversation page in `planning/`, one day at a time and one line per hop. Each line links to the hop itself. Read it top to bottom to follow the project from the first idea, or jump to a day.

**Who is who.** The **owner** is the one person on the team. The **advisor** works with him on plans and documents. The **masters** run the fleet's work, and the **scribe** keeps its record. Everyone except the owner is an AI agent.

**How this page is kept.** Every new hop on a conversation page gets its line here in the same change. A new day gets its own page, `CONVERSATION_<date>.md`, and its own section here. The first page was called `CONVERSATION.md` until 18 October, when the owner had it renamed to match the others; git keeps its history.

| day | page | what the day was about |
|---|---|---|
| 16 October 2026 | [`CONVERSATION_2026-10-16.md`](CONVERSATION_2026-10-16.md) | choosing the topic: the question to the fleet, eleven answers, the name proven |
| 17 October 2026 | [`CONVERSATION_2026-10-17.md`](CONVERSATION_2026-10-17.md) | preparing the machines for the Cursor executors; then reshaping the design for the sponsors' technology |
| 18 October 2026 | [`CONVERSATION_2026-10-18.md`](CONVERSATION_2026-10-18.md) | a local NVIDIA guardrail model, made public and on by default in proven |
| 19 October 2026 | [`CONVERSATION_2026-10-19.md`](CONVERSATION_2026-10-19.md) | the front page rewritten for readers new to agent fleets |

---

## 16 October 2026 — choosing the topic

Page: [`CONVERSATION_2026-10-16.md`](CONVERSATION_2026-10-16.md)

| hop | who → whom | what happened |
|---|---|---|
| [1](CONVERSATION_2026-10-16.md#1--owner--advisor-where-the-idea-came-from) | owner → advisor | The idea: pull-request review like Claude Code's, plugged into our own infrastructure. |
| [2](CONVERSATION_2026-10-16.md#2--owner--advisor-let-the-fleet-choose) | owner → advisor | The topic should come from a survey of the masters and the scribe; the owner adds his own candidate. |
| [3](CONVERSATION_2026-10-16.md#3--advisor--documents-the-question-and-its-own-answer) | advisor → documents | The question to the fleet, the advisor's proposal (proof-carrying code review), and the event notes. |
| [4](CONVERSATION_2026-10-16.md#4--owner--advisor--owner-a-message-for-the-fleet) | owner → advisor → owner | A short message to paste to the fleet. |
| [5](CONVERSATION_2026-10-16.md#5--owner--ten-masters-and-the-scribe-the-question-goes-out) | owner → masters and scribe | The owner carries the question by hand, while fleet messaging is being repaired. |
| [6](CONVERSATION_2026-10-16.md#6--masters-and-the-scribe--answers-eleven-answers) | masters and scribe → answers | Eleven answers between 20:03 and 20:19, each from its own track's record. |
| [7](CONVERSATION_2026-10-16.md#7--in-the-meantime-owner--advisor-what-this-is-for) | owner → advisor | Part of the platform goes to the community; the name: proven. |
| [8](CONVERSATION_2026-10-16.md#8--owner--advisor-publish-the-answers-and-say-what-they-show) | owner → advisor | Publish the answers verbatim, blacked out, and say what they show. |
| [9](CONVERSATION_2026-10-16.md#9--advisor--this-repository) | advisor → this repository | The answers and what they show land here; the owner pushes. |
| [10](CONVERSATION_2026-10-16.md#10--scribe--the-record-a-bug-report-the-same-evening) | scribe → the record | Fleet messaging stops again; the scribe files a bug report in ten minutes. |

## 17 October 2026 — the machines, then the design

Page: [`CONVERSATION_2026-10-17.md`](CONVERSATION_2026-10-17.md)

| hop | who → whom | what happened |
|---|---|---|
| [1](CONVERSATION_2026-10-17.md#1--owner--advisor-the-kickoff) | owner → advisor | Kickoff: four Cursor seats, a new Linux machine, Go and Python. |
| [2](CONVERSATION_2026-10-17.md#2--advisor--the-record-read-first-then-measure) | advisor → the record | Read first, then measure. |
| [3](CONVERSATION_2026-10-17.md#3--advisor--the-cursor-backend-train-what-has-landed) | advisor → Cursor train | Which Cursor seats the engine can run today: two of four. |
| [4](CONVERSATION_2026-10-17.md#4--advisor--the-machines-what-is-there) | advisor → the machines | What the machines carry; one Python version in both lanes. |
| [5](CONVERSATION_2026-10-17.md#5--two-refusals-and-what-the-advisor-did-with-them) | — | Two refusals, and why the advisor did not work around them. |
| [6](CONVERSATION_2026-10-17.md#6--advisor--the-master-the-train-order) | advisor → master | The train order `proven-cursor-infra`, seven wagons. |
| [7](CONVERSATION_2026-10-17.md#7--advisor--owner-four-decisions-only-he-can-make) | advisor → owner | Four decisions only the owner can make. |
| [8](CONVERSATION_2026-10-17.md#8--owner--advisor-the-advisor-was-wrong) | owner → advisor | The advisor was wrong about the node, twice. |
| [9](CONVERSATION_2026-10-17.md#9--owner--advisor-is-everything-in-this-repository) | owner → advisor | Is everything in this repository? Not yet. |
| [10](CONVERSATION_2026-10-17.md#10--owner--advisor-both-places-always) | owner → advisor | Every hackathon document is saved in both places. |
| [11](CONVERSATION_2026-10-17.md#11--owner--advisor-the-design-does-not-meet-the-rules) | owner → advisor | The design does not meet the rules: Zetaris and Meterless are required. |
| [12](CONVERSATION_2026-10-17.md#12--owner--advisor-go-the-first-sources-the-same-track) | owner → advisor | Go, the first evidence sources, the same track. |
| [13](CONVERSATION_2026-10-17.md#13--owner--advisor-write-it-in-commit-it-and-explain-me-to-the-readers) | owner → advisor | Infra points written in; the owner, explained to the readers. |

## 18 October 2026 — a local guardrail

Page: [`CONVERSATION_2026-10-18.md`](CONVERSATION_2026-10-18.md)

| hop | who → whom | what happened |
|---|---|---|
| [1](CONVERSATION_2026-10-18.md#1--owner--advisor-a-guardrail-model-for-the-fleet) | owner → advisor | NVIDIA's Nemotron 3.5 Content Safety as our local guardrail model. |
| [2](CONVERSATION_2026-10-18.md#2--advisor--the-record-the-machine-and-the-model-measure-first) | advisor → record, machine, model | Measure first: it most likely fits; three points need care. |
| [3](CONVERSATION_2026-10-18.md#3--advisor--the-infra-master-the-train-order) | advisor → infra master | A parallel train order, six wagons. |
| [4](CONVERSATION_2026-10-18.md#4--owner--advisor-english-in-this-repository) | owner → advisor | English in this repository. |
| [5](CONVERSATION_2026-10-18.md#5--the-owner-corrected-the-advisors-redaction) | — | The owner corrected the advisor's redaction. |
| [6](CONVERSATION_2026-10-18.md#6--owner--advisor-keep-the-conversation-pages-up-to-date) | owner → advisor | Keep the conversation pages up to date. |
| [7](CONVERSATION_2026-10-18.md#7--owner--advisor-english-a-black-out-and-the-guardrail-goes-public) | owner → advisor | A black-out; the guardrail order goes public, on by default in proven. |
| [8](CONVERSATION_2026-10-18.md#8--owner--advisor-a-diary-of-the-conversations) | owner → advisor | This diary. |
| [9](CONVERSATION_2026-10-18.md#9--owner--advisor-rename-the-first-page) | owner → advisor | The first page renamed to `CONVERSATION_2026-10-16.md`. |
| [10](CONVERSATION_2026-10-18.md#10--owner--advisor-the-hosted-guardrail-and-design-decisions-as-adrs) | owner → advisor | NVIDIA hosts the guardrail model too, so judges can keep it on; design decisions will be ADRs. |
| [11](CONVERSATION_2026-10-18.md#11--owner--advisor-write-them-as-adrs-now) | owner → advisor | The five decisions taken so far written as ADRs 0001 to 0005. |
| [12](CONVERSATION_2026-10-18.md#12--owner--advisor-two-more-decisions-of-their-own) | owner → advisor | The Skeptic and "no exploit code" split out as ADR-0006 and ADR-0007. |
| [13](CONVERSATION_2026-10-18.md#13--owner--advisor-a-cover-the-video-and-the-way-in) | owner → advisor | A cover image that plays the video, and a "Start here" link to this diary. |
| [14](CONVERSATION_2026-10-18.md#14--owner--advisor-one-language-for-the-readers) | owner → advisor | One language for the readers: only the first day keeps the owner's Polish. |
| [15](CONVERSATION_2026-10-18.md#15--owner--advisor-watch-it-being-built) | owner → advisor | "Watch it being built, slice by slice": a link to the orders, reports, reviews and decisions. |
| [16](CONVERSATION_2026-10-18.md#16--owner--advisor-the-video-playing) | owner → advisor | The video's opening slides as a silent loop at the top of the README; a click plays the video. |

## 19 October 2026 — the front page for a wider audience

Page: [`CONVERSATION_2026-10-19.md`](CONVERSATION_2026-10-19.md)

| hop | who → whom | what happened |
|---|---|---|
| [1](CONVERSATION_2026-10-19.md#1--owner--advisor-make-the-front-page-understandable) | owner → advisor | The README starts with the person and the story; the design moves to DESIGN.md; SAPHAN.md checked against the documentation. |
| [2](CONVERSATION_2026-10-19.md#2--owner--advisor-which-source-is-right) | owner → advisor | The documentation was right about network access; the product site was corrected in four languages. |
| [3](CONVERSATION_2026-10-19.md#3--owner--advisor-how-the-reviewer-is-chosen) | owner → advisor | ADR-0008: the reviewer's model is chosen against the author's model the change states; NVIDIA by default; security reviews doubled. |
| [4](CONVERSATION_2026-10-19.md#4--owner--advisor-when-the-reviewer-is-the-authors-own-model) | owner → advisor | ADR-0009: two reviewer models; with only two, the second review may be the author's model as a fresh agent; A writes, NVIDIA reviews, B for special cases is best. |
| [5](CONVERSATION_2026-10-19.md#5--owner--advisor-nvidia-is-a-default-not-a-requirement) | owner → advisor | ADR-0010: every reviewer model is configurable, NVIDIA's included; no setup is blocked; evidence is not configurable. |
| [6](CONVERSATION_2026-10-19.md#6--owner--advisor-the-last-open-points-and-a-rule-for-every-setting) | owner → advisor | ADR-0011 `Assisted-by:` trailer, ADR-0012 Saphan Studio's review classes, ADR-0013 no hardcoded settings. |
| [7](CONVERSATION_2026-10-19.md#7--owner--advisor-who-decides-a-changes-class) | owner → advisor | ADR-0014: for the hackathon, a change's class comes from configured path rules, not a model. |
| [8](CONVERSATION_2026-10-19.md#8--owner--advisor-the-same-principles-as-saphan-studio) | owner → advisor | Saphan Studio's principles for settings applied: a closed list, validated on load, no heuristics; a model may only raise a class. |