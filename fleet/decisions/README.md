# fleet/decisions/

A decision is the owner's. No agent can merge, and no agent's "yes" counts: a reviewer's accept is advice the owner reads, not a decision.

**Rules every decision follows**

- **One of four outcomes:** `accept`, `accept-with-conditions`, `reject` or `close-without-shipping`.
- **The reason is written by the person who decides.** The owner writes it himself, in his own words. A decision without a reason is not published.
- **Conditions are named.** If the owner accepts with conditions, each one says what closes it and in which later change.
- **A rejected change keeps its documents.** The order, the status report and the review of every round stay in this repository, including the ones that were sent back.

**Template**

```markdown
# DECISION — <slug> · round <n>

- outcome: <accept | accept-with-conditions | reject | close-without-shipping>
- decided by: the owner
- review: fleet/reviews/REVIEW_<slug>_r<n>.md
- merged as: <commit — or "not merged">
- decided: <date and time, UTC>

## Reason
<the owner's own words>

## Conditions
<each condition and the change that closes it — or "none">
```
