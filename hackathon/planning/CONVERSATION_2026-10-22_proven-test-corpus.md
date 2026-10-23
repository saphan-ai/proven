# CONVERSATION — proven-test-corpus, the NVIDIA reviews wait for an engine fix, 22 October 2026

The page of the train `proven-test-corpus` for 22 October, written by its master, following [the conversation-log convention](../fleet/messages/CONVENTION_every-master-keeps-a-conversation-log_2026-10-22.md).

## Hop by hop

### 1 · master → owner: the NVIDIA review works, and then wedges

- 21 October, evening: a corpus review by `nvidia/nemotron-3.5-super-vl-120b-a12b` ran through Meterless Nightshift with four signed receipts. Three NVIDIA seats now exist for parallel reviews.
- The NVIDIA reviews of PV-001 and PV-004 could not be dispatched again: the engine held a dead run's seat. The fix is W12 of `proven-cursor-infra` (see [its page](CONVERSATION_2026-10-22_proven-cursor-infra.md)); the owner is landing it. PV-001's NVIDIA review is re-dispatched after it lands, as the fix's live check.

### 2 · owner → masters: two parallel reviews, one NVIDIA

- **What the master did:** wrote the rule into this train's order (every named wagon's review: an `nvidia/*` one through Nightshift with its receipt, and one other, in parallel).

## Waiting for answers

- the owner's candidate cases C-01 (PV-014) and C-02 (PV-001), from the families' disagreement, still wait for a yes.
