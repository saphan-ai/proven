# fleet/LEDGER.md — how every change went

One row per change, added with its decision. Every number comes from the documents of that change in this folder.

## The ledger

| change | built by (model, or Auto → model) | rounds | repair rounds | BLOCKING | OWED | NOTE | outcome |
|---|---|---|---|---|---|---|---|
| *(rows for the merged changes are owed: their figures are not yet collected)* | | | | | | | |

- **rounds**: how many times the change was built. **repair rounds**: rounds after the first, each caused by a review or a decision that sent it back.
- **BLOCKING / OWED / NOTE**: conditions found by all the reviews of the change, counted by weight.
- **built by**: the model each round ran on. With Auto, the model Auto chose, as `Auto → model`.
- **outcome**: accept, accept-with-conditions, reject or close-without-shipping.

## Steering, day by day

Once a day the owner sets which models the executors use and how many run in parallel. The setting, its reason and what it produced the next day:

| day | executors in parallel | models | the owner's reason | next day: merged · repair rounds · BLOCKING |
|---|---|---|---|---|
| 22 Oct | 2 | | | |

## Totals

| | value |
|---|---|
| changes decided | 0 |
| merged | 0 |
| BLOCKING conditions found | 0 |
| OWED conditions found | 0 |

## How good was the agents' code?

At the end of the build, on 27 October, this section answers from the ledger alone, per model:

- how many changes were accepted in their first round;
- how many repair rounds a change needed on average;
- how many BLOCKING and OWED conditions each model's changes drew per change;
- which orders produced the most repair rounds, and what those orders had in common.

The orders are public in [`orders/`](orders/), so anyone can check whether a model did badly or the order asked badly.
