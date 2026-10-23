# CONVENTION — every master keeps a public conversation log

- from: the advisor, relaying the owner, 22 October 2026
- to: every master of a Proven train
- the owner's words (translated from Polish): *"Remember to record the correspondence. Not only with me, but the messages and answers from the masters."* · *"Send the masters how they are to keep the conversation log."*

## What

Each master keeps **one page per day** in this repository:
`planning/CONVERSATION_<YYYY-MM-DD>_<track-slug>.md`, and adds one row for it to [`planning/CONVERSATIONS.md`](../../planning/CONVERSATIONS.md) (date · link · one line on what the day was about).

The model to follow: [`planning/CONVERSATION_2026-10-22_advisor.md`](../../planning/CONVERSATION_2026-10-22_advisor.md).

## The shape of the page

1. A title: `# CONVERSATION — <track>, <what the day was>, <date>`, and one paragraph saying what the page continues and how it is written.
2. `## Hop by hop`. One `### N · <who> → <who>: <subject>` per exchange. In each hop:
   - **the owner's words**, quoted, in English (translate Polish and say so);
   - **what the master read** before acting (orders, record, measurements);
   - **what the master did**, with commit hashes and stream slugs;
   - **messages sent**: each message id, its addressee track, and its substance;
   - **answers received**: the id, who answered, and what the answer said;
   - **decisions**: what was decided and by whom (the owner decides; a master proposes).
3. `## Waiting for answers`: every message id still unanswered, one line each.
4. Where something was not run or not measured, say so (a `## Not run` section).

Record every exchange: with the owner, with other masters, with the advisor and the scribe, and with the agents you dispatch: each order, its run id, its return, every verify and review verdict, and each gate.

## Black out before you commit

One █ per character for: login names, machine names other than macm5, macmini and runner-cursor, IP addresses, internal paths, private repository names, seat identities (for example cursor seat names), credential names, and where the machines are. Commit hashes and message ids stay.

## How to commit

- Write the page as you go, not at the end of the day. Commit often.
- **No AI attribution** in the commit: no Co-Authored-By, no "Generated with", no tool name.
- **Never push.** The owner pushes once a day, at 15:00 UTC (22:00 Bangkok), with the daily update.
