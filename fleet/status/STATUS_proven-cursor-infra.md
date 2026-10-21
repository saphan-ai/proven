# STATUS — proven-cursor-infra · master progress

- state: in-progress
- executor: none yet — this is the master's own progress; no Cursor agent has run in this train
- model: claude-opus-5-5 (the master's session)
- pool: not used — no Cursor turn has run
- cost: not measured — an interactive master session; no priced act has run
- duration: not measured
- last commit: 5e78a31 "[proven-cursor-infra] [nemotron-content-safety] public copies follow the amended train orders (2026-10-19)"
- updated: 2026-10-21 (W6 on the reinstalled engine refused twice; the Cursor backend train's fix W22 landed)

## Summary
runner-cursor exists, carries the toolchain and is admitted to the fleet with one logged-in Cursor seat. One Cursor turn
ran. **No Cursor agent can work unattended yet:** the engine runs Cursor inside no sandbox on any machine, so every run
needs the owner's recorded act to open its network, and the agent's shell tool was refused. Both fixes belong to the
Cursor backend train. Both landed on 21 October. On the reinstalled engine W6 was refused again, for two new reasons;
the Cursor backend train fixed the second one (its wagon W22). W6 runs again after the fleet's reinstall.

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
14. 2026-10-20/21: the owner accepted the recommendations for D1, D2, D4, D5 and D6; the DNS name for runner-cursor
    resolves; the leg user is in the docker group on runner-cursor.
15. 2026-10-21: W4 built on runner-cursor: `fleet/infra/go-python/Dockerfile` on our Go image. Inside the image:
    `go1.26.8`, Python 3.13.5 (venv works), Node v22.23.2, npm 10.9.8, gcc 14.2.0, `CGO_ENABLED=1`, user 65534.
    Image ID `sha256:d282b9ee1390b029a8567adf49117fa0c7448b09edaab1cbd874b4f0ecf9135b`; base image ID
    `sha256:36ac52c9ec887c2223b1dcda83cc857a7b64aea10e0560cf9fafa56c4ff67ac7`. The native lane has the same Go, the same
    Python minor and the same Node major; only gcc differs (13.3 natively).
16. 2026-10-21: both engine fixes for Cursor (a sandbox for Cursor runs, and the agent's shell) landed. They wait for the
    fleet's reinstall. The owner re-signed the W6 order.
17. 2026-10-21: W6 on the reinstalled engine, which carries both fixes. It did not run. One dispatch path refused the
    Cursor seat, because the seat's usage is unmeasured: the engine has no reader for Cursor usage. The other path
    composed the run, and the run was refused twice, because the Cursor class files (the agent's hooks and its command
    rules) did not reach an existing far work directory. The master stopped at the second refusal and reported both to
    the Cursor backend train's master, with high salience. That train fixed the second gap in its wagon W22
    (`cursor-far-reuse-packs`). W6 runs again after the fleet's reinstall.

**2026-10-21, 07:02 UTC — W6: the first Cursor agent turn on runner-cursor works.** It ran inside the engine's sandbox
(Landlock), its tool hook answered, and its shell, read and write tools ran: Go 1.26.8 with cgo on, and `go run` printed
`hello from proven`. The model was Composer. The network was opened by the owner's recorded act for this one run, because
Cursor's agent stream uses a host the Cursor network baseline does not admit yet. One refusal is left: the executor's
permission class does not allow `python3`, `node`, `npm`, `gcc` or `curl`, so the Python, Node and compiler checks and the
network reading for W8 did not run. The Cursor backend train has the request.

**2026-10-21, 11:00 UTC — W6 on Grok, under the polyglot class.** The Cursor leg on runner-cursor ran with Grok 4.7 and the new
`polyglot` executor class: Go 1.26.8 with cgo, Python 3.13.16, Node v24.5.0, npm 10.9.9, gcc 13.3.0, and `go run` printed
`hello from proven`. `curl` stays refused by the class (egress is measured with a Go program instead). The leg's Node is the
Cursor agent's own bundled Node 24, so the owner chose Node 24 everywhere: the build image now carries Node 24.5.0
(image `sha256:fb798be5c7e48bb6a74f85e19aebc999617908794f9206a2073a5ea45cdacb52`), and the host follows.

## Wagon state
| wagon | state | next |
|---|---|---|
| W1 `proven-is-a-workspace-repo` | done 2026-10-20: repository registered, claims re-measured, zero-cost compose passed | — |
| W2 `runner-cursor-exists` | the machine runs and is a tailnet member | the wiring: DNS name, monitor, ssh alias |
| W3 `runner-cursor-toolchain` | done 2026-10-20 | — |
| W4 `proven-go-python-image` | done 2026-10-21, image built and checked | — |
| W5 `cursor-seats-admitted` | runner-cursor: one seat admitted and logged in; macm5: key needed (D3) | the macm5 seat after D3 |
| W6 `cursor-seats-smoke-turn` | **2026-10-21: the Cursor turn ran behind the guardrail proxy with egress closed and was recorded live.** The agent reached the model only through the proxy, answered, and measured the toolchain (Go 1.26.8 with cgo, Python 3.13, Node 24.5, npm, gcc 13). The owner watched the session being recorded live in the console. `curl` is refused by the build class by design, so egress is W8's | — |
| W7 `proven-public-infra-record` | the public train order is current | goes with each wagon |
| W8 `cursor-legs-reach-data-hosts` | **measured 2026-10-21 behind the proxy, egress closed:** a Go program made one request per host. `registry.npmjs.org`, `proxy.golang.org` and `api2.cursor.sh` answer 200; OSV, CISA KEV, EPSS, GitHub's API, NVIDIA's API and the control `example.com` are refused at name lookup. So a Cursor leg reaches none of proven's data sources today | — (the owner: no data host for build legs) |
| W9 `proven-judge-rehearsal` | not started | waits for W4 and D4 |

## Blocked on / next
- **2026-10-21, the milestone:** Cursor legs no longer need the owner's per-run egress approval. The Cursor backend
  train put every Cursor leg behind the guardrail proxy, so the sandbox keeps egress closed, the model is reached through
  one recorded path, and every turn is recorded (ADR-0024). Measured on runner-cursor with W6; the first attempt failed on
  an HTTP/2-versus-HTTP/1.1 mismatch, fixed in the seat's settings.
- **Decided by the owner, 2026-10-21:** a build leg reaches no data host directly. Evidence is loaded through Zetaris
  (ADR-0023) and tests use recorded fixtures; the Zetaris host is the one address a leg may come to need.
- **Owed by the owner:** the wiring choice for runner-cursor (DNS name and monitor); decisions D1–D6 (train order §3).
- **Found on the way:** re-signing an unchanged order with the train's delegate key does not free a new placement; the
  owner re-signs as a person until it is fixed.

## REDACTED
| kind | count |
|---|---|
| engine refusal codes and run identifiers (rewritten in neutral words, item 17) | 4 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17 (the Cursor executors work there).
