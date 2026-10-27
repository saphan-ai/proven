# ADR-0024: Every coding session of a Cursor agent on Proven is recorded, and the recording is published in this repository

**Status:** accepted 2026-10-21 · amended 2026-10-28 (point 2 withdrawn for the build window) · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings); the owner's rule of 2026-10-21 that orders, executor reports and reviews are public, blacked out
**Trigger:** the preparation status check of 21 October ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../hackathon/planning/CONVERSATION_2026-10-21_advisor.md))

## Context

Proven is written by Cursor agents inside the build window, 22 to 27 October. The orders, the executors' reports and
the reviews are already public. What the agent actually did between the order and the report, turn by turn, is not.

Saphan Studio records the sessions of its other coding backends. For Cursor, that recording is being built now: wagon W6
of the Cursor backend train, *cursor-is-recorded-and-sealed*, sends every Cursor turn through Saphan Studio's recording
proxy, and refuses a turn the proxy cannot record. On 21 October the owner (translated from Polish):

> *"I would also like to have session recordings, as in the others. W6 will probably handle that for us. We will put
> the recordings of all the sessions of the Cursor coding agents into the proven repository."*

## Decision

1. **Every coding session of a Cursor agent that works on Proven is recorded.** The recording is the one Saphan Studio's
   recording proxy makes (W6). A coding session that was not recorded is said so in its report, with the reason.
2. **Recordings are not published for the hackathon build window.** The repository is frozen at the submission point
   until 6 November; after that, sample recordings of Cursor coding sessions are added here, next to the order and the
   report of their run, with the engine's run id and start time.
3. **A recording is blacked out before it is committed**, by the same rules as every other public document here. The
   credential of a Cursor seat must never be in a recording in the first place; the blackout is not what keeps it out.
4. **Where recordings are kept, their format and how they are blacked out are settings and tooling, not this ADR.** They
   are written down when the first recording is published.

## Consequences

- A reader can check what an agent did, not only what it reported, and can match each session to the commits it made.
- The run id and start time in each recording are part of the evidence that the code was written inside the window.
- **Not known yet:**
  - whether W6 lands before the first coding run on 22 October; until it does, a session is not recorded, and point 1
    makes its report say so;
  - the size of a recording, and whether the repository can hold all of them as plain files;
  - whether Cursor's terms allow publishing what Cursor itself sends in a session, such as its own system prompt; if not,
    that part is left out and the recording says so;
  - a blackout tool that works on recordings, and a check that refuses a recording with something left to black out.

## Amendment, 2026-10-28

The owner withdrew the promise to publish every recording during the build window: the recordings were not blacked out and published in time, and the repository is frozen at the submission point. Sample recordings will be published after 6 November, when the repository is unfrozen. Points 1, 3 and 4 stand.
