# CONVERSATION — proven-zetaris-local, OSV from storage and the Wildcard data, 22 October 2026

The page of the train [`proven-zetaris-local`](../fleet/orders/TRAIN_ORDER_proven-zetaris-local.md) for 22 October, written by its master, following [the conversation-log convention](../fleet/messages/CONVENTION_every-master-keeps-a-conversation-log_2026-10-22.md). It continues hops 14–23 of [`CONVERSATION_2026-10-21.md`](CONVERSATION_2026-10-21.md).

## Hop by hop

### 1 · advisor → master: can Zetaris read a file from its own storage?

- **The question** (the owner had chosen a copy of OSV in storage, [ADR-0023](../fleet/decisions/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md)): a tiny JSON and a tiny CSV, registered by the administrator, read with proven's key; made-up rows; removed afterwards.
- **What the master did** (through a helper, 00:16–00:18 UTC): a **CSV** file and a **JSON array** are queryable with proven's key, 0.10–0.15 s per query; a JSON Lines file returns only its first record. Reads are live. Before the grant, proven's query is refused. Everything removed and shown gone.
- **Message sent:** the result to the advisor, about 20 minutes after the promised 00:00 UTC.

### 2 · advisor → master: the Wildcard track

> Proven also enters the Wildcard track; its entries "must integrate Zetaris and Meterless". (owner, relayed by the advisor)

- **What the master did:** added W10 `zetaris-holds-real-evidence` to this train: real KEV, EPSS and OSV data in Zetaris, readable with proven's key, by 24 October; the live review that shows a Zetaris query's SQL and rows in its trace by 27 October is the reviewer train's. Confirmed to the advisor.

## Waiting for answers

- the advisor: routing the live-review half of the hard point to the reviewer train.
