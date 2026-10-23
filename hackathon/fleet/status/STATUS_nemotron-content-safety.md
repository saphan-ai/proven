# STATUS — █████-nemotron-content-safety · master progress

- state: done — the train is closed (W1–W6), 2026-10-21
- executor: none — this train runs owner-act scripts and the master's measurements; no Cursor agent
- model: claude-opus-5-5 (the master's session)
- pool: not used
- cost: not measured — an interactive master session; every act in this train is local energy, a named zero
- duration: not measured
- last commit: 5e78a31 "[proven-cursor-infra] [nemotron-content-safety] public copies follow the amended train orders (2026-10-19)"
- updated: 2026-10-21 (DONE)

## Summary
The train is DONE: W1–W6. Nemotron is fetched, verified and warm on llama-swap in place of Llama-Guard-3-8B, and it judges
prompts and responses in the card's format, directly and through the LiteLLM gateway. One agent-loop run (W5) showed that
the agent loop carries the judge's call but not a classification task: callers call the judge directly. The W5 stream
closed without a merge. The record (W6) is written, and the relay to the guardrail-images train is sent.

## Done, 2026-10-19
1. Re-measured N1–N9, read-only (§1a of `fleet/orders/TRAIN_ORDER_nemotron-content-safety.md`).
2. Measured the GGUF files: `Q8_0` is 4,130,408,096 bytes, sha256 `217c703e74074fa57116fee18eb4531e4c7cbbf1e2ed7d626076004900cb7884`.
3. Found that the model's chat template reads a fourth option, `custom_taxonomy`, beside `request_categories`,
   `custom_policy` and `enable_thinking`.
4. Found that an agent-loop lane through the LiteLLM gateway already exists (N8 was wrong), so the test run (W5) also proves
   the gateway alias.
5. The owner signed the train order and ran the configuration copy.
6. Read the live configuration. The largest model in the swapped `big` group is `glm-4.5-air` Q4_K_M, 67.96 GiB.

## Done, 2026-10-20 and 2026-10-21
7. 2026-10-20: W1 measured (Q2). One load of `glm-4.5-air` ran the guardrail host out of memory 52 seconds later. The host
   stopped answering until a reboot, about an hour later. Every other master was told. **The worst case does not fit, even
   without this model.**
8. 2026-10-20/21: the owner accepted Q1 (text only) and Q3 (the gateway's existing key), and chose (b) for the fit: one
   judge leaves the `safety` group and the preload. On 2026-10-21 he named it: Llama-Guard-3-8B. Its file stays on disk.
9. 2026-10-21: W2 done. The fetch script first refused a wrong checksum on purpose and kept no file (the negative
   control). Then it fetched `Q8_0`: 4,130,408,096 bytes, sha256
   `217c703e74074fa57116fee18eb4531e4c7cbbf1e2ed7d626076004900cb7884`, as measured on 2026-10-19.
10. 2026-10-21: W3 applied by the owner's script. Nemotron is warm on llama-swap in place of Llama-Guard-3-8B. Probes
    through llama-swap, at temperature 0:

    | probe | verdict | latency |
    |---|---|---|
    | (a) benign | `User Safety: safe` | 156 ms |
    | (b) malicious | `User Safety: unsafe` | 378 ms |
    | (c) as (b), `request_categories="/categories"` | `User Safety: unsafe`, `Safety Categories: Criminal Planning/Confessions, Other` | 634 ms |
    | (d) prompt plus an `assistant` response | no verdict: the server continued the response | — |

    Qwen3Guard, for comparison: 199 ms and 388 ms. Probe (d) fails because the llama.cpp server, by default, treats a
    final assistant message as a prefix and continues it. The fix is one server flag, `--no-prefill-assistant`, for this
    model only. It waits for the owner's run.
11. 2026-10-21: the owner ran the `--no-prefill-assistant` fix, for this model only. Probe (d): a bad response →
    `User Safety: safe`, `Response Safety: unsafe`, `Safety Categories: Malware, PII/Privacy`, 706 ms; a good response →
    `Response Safety: safe`, 501 ms. **W3 done.**
12. 2026-10-21: the owner ran W4. The gateway's liveliness check answers 200. Its authenticated model list grew from 4 to
    5: the four old aliases are unchanged, and `local-safety-nemotron` is new. Probe (b) through the gateway →
    `User Safety: unsafe`, 1623 ms. **W4 done.**
13. 2026-10-21: W5, one agent-loop run through the gateway lane, model `local-safety-nemotron`, cost local energy. The
    order was re-signed with a quote of $0.01, because the order grammar refuses 0. The engine refused to place the run
    on an agent-loop seat and refused to run it without the owner's act, because the agent loop runs inside no sandbox on
    macOS. The owner ran it with his recorded act that opens the run's network. Outcome: the child failed, the loop made
    no progress. Three iterations went through the gateway lane (the alias works). The model answered
    `User Safety: safe`, `Response Safety: safe` about the loop's own composed prompt, and made no tool call. Tokens: 4367
    in, 25 out; 2.1 s. **Finding: the agent loop carries the judge's call, not a classification task. Call the judge
    directly.** W5 done, the outcome classified honestly.
14. 2026-10-21: the W5 stream closed by the owner's gate, without a merge; its closure is closed.
15. 2026-10-21: W6 done. The record of the hosts is written in our infrastructure repository, and the relay to the
    `guardrail-server-images` train is sent: the model, its two endpoints, the request contract and the probes. No tester:
    this train changed hosts, not a product. Open, outside this train: the gateway's configuration pin was not
    re-measured after W4.
16. 2026-10-21: the train order carries its CLOSING section and the closing line of law #38. **The train is DONE.**

## Fit, by file sizes (law #46)
| item | GiB |
|---|---|
| warm set now (coder model and two judges) | 32.28 |
| largest `big` member, `glm-4.5-air` Q4_K_M | 67.96 |
| this model, Q8_0 at a 16K context, with its cache and buffers | about 6.5 |
| total | about 106.7 of 112.00 GPU-addressable memory |

System memory shows 51 GiB available and swap is full. The margin is about 5 GiB before the large model's own cache.
The measurement of 2026-10-20 (item 7) replaced this estimate: the worst case does not fit.

## Wagon state
| wagon | state | next |
|---|---|---|
| W1 `nemotron-fit-measured` | done 2026-10-20: measured under load, the worst case does not fit; the owner chose (b) | — |
| W2 `nemotron-gguf-on-█████` | done 2026-10-21: fetched, sha256 verified, negative control passed | — |
| W3 `nemotron-warm-on-llama-swap` | done 2026-10-21: warm in place of Llama-Guard-3-8B; probes (a)–(d) in the card's format, (d) after `--no-prefill-assistant` | — |
| W4 `nemotron-on-litellm-gateway` | done 2026-10-21: 5 aliases, the four old ones unchanged; probe (b) through the gateway, 1623 ms | — |
| W5 `nemotron-agent-loop-run` | done 2026-10-21: one run, classified; the loop carries the judge's call, not a classification task; the stream closed without a merge | — |
| W6 `nemotron-infra-record` | done 2026-10-21: the record written, the relay sent | — |

## Blocked on / next
- Nothing. The train is DONE: nothing in flight, nothing owed to the owner.
- Open, outside this train: the gateway's configuration pin was not re-measured after W4.
- Wiring the model into proven is built in proven from 2026-10-22, not in this train.

## REDACTED
| kind | count |
|---|---|
| machine name (in the slug and in W2's slug) | 2 |
| seat identity (rewritten: "an agent-loop seat") | 1 |
| engine refusal codes and the owner's per-run flag (rewritten in neutral words, item 13) | 3 |
| the name of the gateway's configuration pin (rewritten) | 1 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17. No other machine is named here.
