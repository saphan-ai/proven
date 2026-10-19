# STATUS — proven-cursor-infra · master progress

- state: in-progress
- executor: none yet — this is the master's own progress; no Cursor agent has run in this train
- model: claude-opus-5-5 (the master's session)
- pool: not used — no Cursor turn has run
- cost: not measured — an interactive master session; no priced act has run
- duration: not measured
- last commit: 5e78a31 "[proven-cursor-infra] [nemotron-content-safety] public copies follow the amended train orders (2026-10-19)"
- updated: 2026-10-19 16:40 UTC

## Summary
The master re-measured the sixteen claims of the train order and amended the order. The owner signed it. proven is now a
workspace repository. No machine exists yet and no seat was admitted: runner-cursor waits for the owner's creation script.

## Done, 2026-10-19
1. Re-measured A1–A16, read-only. Results are in §1a of the train order (`fleet/orders/TRAIN_ORDER_proven-cursor-infra.md`).
2. Amended the order: Go 1.26.8 (the image's Go), Python 3.13 and Node 22 in both lanes; VM id named explicitly; the tailnet
   reached by a route, as on the model machine; two new owner decisions, D5 (backups) and D6 (the visibility rule).
3. Landed the order in the orchestration store; the stored digest equals the file.
4. Sent three engine requests (R1) to the master of the Cursor backend train: a far Cursor key carrier, a far container lane,
   and egress to proven's data hosts. Reply owed.
5. The owner signed the train order and registered proven (technology Go) as a workspace repository.
6. Corrected W2: the model machine is a tailnet member, so runner-cursor joins the tailnet too. An older report had said
   the model machine was only routed; a fresh reading showed that this was wrong.
7. Composed the creation script for runner-cursor and rehearsed its checks: the VM number and the address are free, and
   memory and disk fit (30 GiB of memory and 616 GB of disk left after the VM).

## Wagon state
| wagon | state | next |
|---|---|---|
| W1 `proven-is-a-workspace-repo` | repository registered; claims re-measured | one zero-cost compose for a Cursor executor on macm5 |
| W2 `runner-cursor-exists` | creation script composed, checks rehearsed | the owner runs it; then the wiring script |
| W3 `runner-cursor-toolchain` | versions chosen | waits for W2 |
| W4 `proven-go-python-image` | layers chosen: Python 3.13, gcc, cgo on; Node already in the base | waits for W3 and D2 |
| W5 `cursor-seats-admitted` | macm5: reuse the existing seat, key needed (D3); runner-cursor: native, login, no key | waits for W3 |
| W6 `cursor-seats-smoke-turn` | not started | priced; waits for W5 and the owner's consent |
| W7 `proven-public-infra-record` | the public train order is current | goes with each wagon |
| W8 `cursor-legs-reach-data-hosts` | measured from the egress policy: no data host is admitted today | R1, item 3 |
| W9 `proven-judge-rehearsal` | not started | waits for W4 and D4 |

## Blocked on / next
- The owner's decisions D1–D6 (train order §3).
- The R1 answer from the Cursor backend train's master.
- Next act of the master: the rehearsed creation script for runner-cursor (W2).

## REDACTED
| kind | count |
|---|---|
| none in this file | 0 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17 (the Cursor executors work there).
