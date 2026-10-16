# fleet/

The paperwork behind every change in this repository. From 22 October 2026, each change arrives together with the documents that explain it.

| folder | document | who writes it | question it answers |
|---|---|---|---|
| `orders/` | the order | a master | what was asked for, and how we will know it is done |
| `status/` | the status report | the executor (a Cursor agent) | what was built, what was run, what failed |
| `reviews/` | the review | a reviewer that did not write the change | does it hold, and what is still owed |
| `decisions/` | the decision | the owner | accepted or not, and why |

**One change, followed end to end.** A master writes the order. A Cursor agent builds the change in its own copy of the repository and writes the status report. A reviewer that did not write the change checks it, starting from a fresh context, and writes the review. The owner reads the review and decides. If the change goes back, the next round gets a new order, and the documents of every round stay here.

**How files are named.** Every document carries the change's short name (a slug) and its round: `ORDER_<slug>_r1.md`, `STATUS_<slug>_r1.md`, `REVIEW_<slug>_r1.md`, `DECISION_<slug>_r1.md`. The same slug appears in the commit messages of the change.

**What is left out.** The owner reviews every document before it is published. Credentials, internal host names and details of our private infrastructure are removed. Where something was removed, the document says so.

Each folder has its own README with the template we use. The templates are simplified public versions of the ones our fleet uses every day.
