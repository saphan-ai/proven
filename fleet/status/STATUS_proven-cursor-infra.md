# STATUS — proven-cursor-infra · master progress

- state: in-progress
- executor: none yet — this is the master's own progress; no Cursor agent has run in this train
- model: claude-opus-5-5 (the master's session)
- pool: not used — no Cursor turn has run
- cost: not measured — an interactive master session; no priced act has run
- duration: not measured
- last commit: 5e78a31 "[proven-cursor-infra] [nemotron-content-safety] public copies follow the amended train orders (2026-10-19)"
- updated: 2026-10-20 (W6 ran: blocked)

## Summary
runner-cursor exists, carries the toolchain and is admitted to the fleet with one logged-in Cursor seat. One Cursor turn
ran. **No Cursor agent can work unattended yet:** the engine runs Cursor inside no sandbox on any machine, so every run
needs the owner's recorded act to open its network, and the agent's shell tool was refused. Both fixes belong to the
Cursor backend train. Its master owes, by 20 October 18:00 UTC, whether they land before 21 October 12:00 UTC.

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

8. 2026-10-20: the owner created runner-cursor and joined it to the tailnet. It has 4 vCPU, 7 GiB of usable memory and a
   116 GB root disk; the guest agent answers.
9. 2026-10-20: composed the toolchain script (W3) and rehearsed its checks. A wrong checksum in the Cursor pin makes the
   installer refuse by name, and nothing is installed.
10. 2026-10-20: W3 done on runner-cursor: Go 1.26.8 with cgo on, Python 3.13.16, Node v22.23.3, gcc 13.3.0, Docker 29.9.1
    and the pinned Cursor agent. The leg user is not in the docker group (D4 is open).
11. 2026-10-20: W5: runner-cursor is admitted to the fleet with one Cursor executor seat. Its login under the seat's own
    home directory works on Linux: this was the first measurement of that question (A6).
12. 2026-10-20: the W6 smoke-turn order is written; it waits for the owner's signature. It runs on `composer-2.5`: the
    engine refuses `auto`, because Auto may switch to another vendor's model.
13. 2026-10-20: W6 found two engine gaps. First, the engine runs no Cursor agent inside a sandbox on any machine, so every
    Cursor run needs the owner's recorded act to open its network, and the owner must be present for each run. Second,
    when the owner ran the smoke turn that way, the agent's shell tool was refused on every call. The turn ran
    `composer-2.5` (35,892 input and 6,591 output tokens); its cost is not measured yet, because Saphan Studio has no
    price for that model. Both gaps went to the Cursor backend train's master, with a deadline of 21 October 12:00 UTC.

## Wagon state
| wagon | state | next |
|---|---|---|
| W1 `proven-is-a-workspace-repo` | done 2026-10-20: repository registered, claims re-measured, zero-cost compose passed | — |
| W2 `runner-cursor-exists` | the machine runs and is a tailnet member | the wiring: DNS name, monitor, ssh alias |
| W3 `runner-cursor-toolchain` | done 2026-10-20 | — |
| W4 `proven-go-python-image` | layers chosen: Python 3.13, gcc, cgo on; Node already in the base | waits for W3 and D2 |
| W5 `cursor-seats-admitted` | runner-cursor: one seat admitted and logged in; macm5: key needed (D3) | the macm5 seat after D3 |
| W6 `cursor-seats-smoke-turn` | ran once: the agent answers, its shell is refused | the Cursor backend train's fix |
| W7 `proven-public-infra-record` | the public train order is current | goes with each wagon |
| W8 `cursor-legs-reach-data-hosts` | measured from the egress policy: no data host is admitted today; the live reading failed, because the shell was refused | the shell fix; R1, item 3 |
| W9 `proven-judge-rehearsal` | not started | waits for W4 and D4 |

## Blocked on / next
- **Blocked:** the Cursor backend train's two fixes (a sandbox for Cursor runs, and the agent's shell). Its master owes an
  estimate by 20 October 18:00 UTC against the deadline of 21 October 12:00 UTC.
- **Owed by the owner:** the wiring choice for runner-cursor (DNS name and monitor); the Cursor usage reading before and
  after the W6 turn, because the engine has no price for its model; decisions D1–D6 (train order §3).
- Nothing is running now.

## REDACTED
| kind | count |
|---|---|
| none in this file | 0 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17 (the Cursor executors work there).
