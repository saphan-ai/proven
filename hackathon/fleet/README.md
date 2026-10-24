# fleet/

The paperwork behind every change in this repository. From 22 October 2026, each change arrives together with the documents that explain it.

| folder | document | who writes it | question it answers |
|---|---|---|---|
| `orders/` | the order | a master | what was asked for, and how we will know it is done |
| `status/` | the status report | the executor (a Cursor agent) | what was built, what was run, what failed |
| `reviews/` | the review | a reviewer that did not write the change | does it hold, and what is still owed |
| `decisions/` | the decision | the owner | accepted or not, and why |
| `messages/` | a message between the fleet's heads, and its answer | a master, the scribe or the advisor | who asked whom for what, and what was done |

**One change, followed end to end.** A master writes the order. A Cursor agent builds the change in its own copy of the repository and writes the status report. A reviewer that did not write the change checks it, starting from a fresh context, and writes the review. The owner reads the review and decides. If the change goes back, the next round gets a new order, and the documents of every round stay here.

**Orders written before the build, run inside it.** The owner, 21 October (translated from Polish): the orders for the build are prepared in advance, as Markdown, and run when the build window opens. Writing an order is planning, not code. What shows that the work itself was done in the window is the record of each run: its run id and start time in the status report, written by Saphan Studio when the agent starts, together with the commits it produced. A commit's own date can be set by whoever makes it, so it is never the only evidence. The orders, the agents' reports and the reviews are public, with private data blacked out, because showing how work is ordered, done and checked is part of what this repository is for.

**How files are named.** Every document carries the change's short name (a slug) and its round: `ORDER_<slug>_r1.md`, `STATUS_<slug>_r1.md`, `REVIEW_<slug>_r1.md`, `DECISION_<slug>_r1.md`. The same slug appears in the commit messages of the change.

**How every change went.** Every status report and every review states the model that ran it. From 25 October they carry no cost; the documents published before then keep the costs they were written with. Every decision adds up the change: who built it on which model, how many repair rounds it took, and how many BLOCKING and OWED conditions the reviews found. Each decision adds one row to [`LEDGER.md`](LEDGER.md), and at the end of the build the ledger answers how good each model's code was.

**What is left out.** The owner reviews every document before it is published. Credentials, internal host names and details of our private infrastructure are removed. Where something was removed, the document says so.

Each folder has its own README with the template we use. The templates are simplified public versions of the ones our fleet uses every day.
