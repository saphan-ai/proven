# STATUS — proven-zetaris-local · master progress

- state: in-progress
- executor: none yet — this train runs owner-act scripts and the master's measurements; no Cursor agent has run in it
- model: claude-opus-5-5 (the master's session)
- pool: not used — no Cursor turn has run
- cost: not measured — an interactive master session; no priced act has run
- duration: not measured
- last commit: ab342ee "[proven-zetaris-local] public copy follows the amended train order"
- updated: 2026-10-21 (W1 done)

## Summary
On 21 October the owner handed this train to the infrastructure master: *"We install a local VM."* (translated from
Polish). The master re-measured the cluster node; the size of the machine became decision Z5, and the owner chose 12 GiB.
The owner signed the order. W1 is done: the Zetaris host runs, with 8 vCPU, 12 GiB of memory and a 160 GB disk, and it is
a tailnet member. Next is W2: Docker, the platform's bundle, the owner's sign-in to the registry, and the preflight.

## Done, 2026-10-21
1. The owner handed the train to the infrastructure master (train order, master line). His words: *"We install a local
   VM."* (translated from Polish).
2. The master re-measured the node, read-only (train order §1a, Z-A9). The node has **25 GiB available**, not 30, because
   runner-cursor now runs. With the file-system cache, which may shrink, the real headroom is about **27 GiB**. The
   advisor's 24 GiB machine did not fit beside the node's other guests with a safe margin.
3. The master added decision Z5, the size: 12 GiB (recommended) or 16 GiB. The owner chose **12 GiB**.
4. The owner signed the train order. His first signing was refused, because the file did not open with its title line,
   and the signing reads the stream's name from that line. The master moved the title line to the top; the second
   signing worked.
5. A first run of the creation script failed at the step that creates the VM. The script computed a memory size on the
   wrong side of an ssh call. This was the master's script error. Nothing was created. The owner's words: *"Fix
   yourself."* (translated from Polish). The master corrected the script, and the corrected run was clean.
6. **W1 done** (`zetaris-host-exists`). The Zetaris host runs with 8 vCPU, 12 GiB of memory, a 160 GB disk and the guest
   agent, and it is a tailnet member. On its LAN address only ssh listens.

## Wagon state
| wagon | state | next |
|---|---|---|
| W1 `zetaris-host-exists` | done 2026-10-21: 8 vCPU, 12 GiB, 160 GB, guest agent, tailnet member; only ssh listens on the LAN address | — |
| W2 `zetaris-platform-installed` | not started | Docker, the bundle, the owner's sign-in to the registry, the preflight |
| W3 `zetaris-platform-runs` | not started | in order, after W2 |
| W4 `zetaris-accounts-and-key` | not started | in order; the owner holds the platform's registration key |
| W5 `zetaris-answers-over-rest` | not started | in order |
| W6 `cursor-legs-reach-zetaris` | not started | in order, and after a Cursor agent can work (train `proven-cursor-infra`) |
| W7 `zetaris-data-survives` | not started | in order |
| W8 `zetaris-runbook-and-record` | the public train order is current | goes with each wagon |
| W9 `lab-rehearsal-host` | not started | in order |

## Blocked on / next
- **Next:** W2. The master composes and rehearses the script; the owner runs it and signs in to the platform's registry
  himself.
- **Owed by the owner:** Z2, Z3 and Z4 (train order §3), and his sign-in to the registry in W2.

## Not run
- W2–W9: not started. W2 is next; the rest follow in order, and W6 also waits for a working Cursor agent.

## REDACTED
| kind | count |
|---|---|
| the new machine's host name (written as "the Zetaris host") | 3 |
| the VM number, the cluster node and the machine's addresses (left out) | 4 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17.
