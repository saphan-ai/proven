# AGENTS.md — for every agent working on this repository

You are working on **proven**, a public, open-source repository (Apache-2.0). Everything you commit here will be read by strangers, including the judges of the Open Agent Hackathon 2026. Read this file before you do anything else.

## 1. Two places, one change

This repository is a sub-repository of our workspace. The fleet keeps working exactly as it works on every other track.

1. **Write and register every document in the bridge, as usual.** Orders, status reports, reviews, test plans and decisions are written in our internal format and registered in the record. The record stays the source.
2. **Then put a public copy here**, in the public template:

   | document | goes to | template |
   |---|---|---|
   | order | `fleet/orders/ORDER_<slug>_r<n>.md` | [`fleet/orders/README.md`](fleet/orders/README.md) |
   | status report | `fleet/status/STATUS_<slug>_r<n>.md` | [`fleet/status/README.md`](fleet/status/README.md) |
   | review | `fleet/reviews/REVIEW_<slug>_r<n>.md` | [`fleet/reviews/README.md`](fleet/reviews/README.md) |
   | decision | `fleet/decisions/DECISION_<slug>_r<n>.md` | [`fleet/decisions/README.md`](fleet/decisions/README.md) |

3. **The public copy says the same thing as the original.** Same findings, same verdict, same numbers. Translate our internal format into the public template; do not soften, shorten or improve it.
4. **Every document carries its model and its cost**, to the cent, as recorded by Saphan Studio. A decision adds up the whole change and adds its row to [`fleet/LEDGER.md`](fleet/LEDGER.md): models, rounds, repair rounds, BLOCKING / OWED / NOTE, cost of building and of reviewing.
5. **The copy lands in the same change as the work it describes.** A change without its documents is not finished.

## 2. What never goes into this repository

Black it out (█, one block per character it covers) and list it in a `REDACTED` section at the end of the document, with what kind of thing it was and how many times. Never delete it silently, and never reword around it.

- credentials, tokens, keys, and the names of the variables that hold them;
- names of our machines, host names, IP addresses, user names;
- absolute paths and paths into our private record or workspace;
- file names and line numbers in Saphan Studio's private code;
- how our platform signs, verifies, pins or licenses anything, and how orders are bound to policies;
- the exact input or condition that gets past one of our own guards;
- table and column names of our private store.

You may keep: slugs, commit hashes of this repository, test names, law numbers, and anything that is already public on https://saphan.ai.

## 3. The rules of this repository

- **English only.** Short sentences, plain words.
- **All project code is written between 22 and 27 October 2026.** Nothing written before 22 October may be copied in.
- **Executors are Cursor agents.** They write the product code. Tests are written by a test author from the test plan, apart from the code. Every other role works as described in [`SAPHAN.md`](SAPHAN.md).
- **Record the model that really ran**, including the model Auto chose, and the share of the monthly pool before and after the run.
- **The owner writes no code.** He decides and merges.
- **Never push.** The owner pushes once a day at 22:00 Bangkok time (15:00 UTC), and once more before the submission on 27 October.
- **Never change `.github/`, `LICENSE` or this file.** Those are the owner's.
- **No document claims more than was run.** If you did not run it, it goes under "Not run" with the reason.
- **Do not write about Saphan Studio's internals.** Saphan is a pre-existing component. What may be said about it is in [`SAPHAN.md`](SAPHAN.md).
