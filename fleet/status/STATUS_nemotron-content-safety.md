# STATUS — █████-nemotron-content-safety · master progress

- state: in-progress
- executor: none — this train runs owner-act scripts and the master's measurements; no Cursor agent
- model: claude-opus-5-5 (the master's session)
- pool: not used
- cost: not measured — an interactive master session; every act in this train is local energy, a named zero
- duration: not measured
- last commit: 5e78a31 "[proven-cursor-infra] [nemotron-content-safety] public copies follow the amended train orders (2026-10-19)"
- updated: 2026-10-21

## Summary
The master re-measured the nine claims and amended the order. The owner signed it and gave the master a copy of the live
llama-swap configuration with every key blacked out. The fit was then measured under load: the worst case does not fit,
even without this model. The owner chose (b): Llama-Guard-3-8B leaves the warm set. Nemotron is fetched, verified and
warm in its place. Probes (a)–(c) come back in the card's format; probe (d) needs one server flag.

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
| W3 `nemotron-warm-on-llama-swap` | applied 2026-10-21: warm in place of Llama-Guard-3-8B; probes (a)–(c) come back in the card's format, (d) does not | the owner's run of the `--no-prefill-assistant` flag, then probe (d) |
| W4 `nemotron-on-litellm-gateway` | not started; Q3: the existing key | the owner's script, after W3 |
| W5 `nemotron-agent-loop-run` | the gateway lane exists | after W4 |
| W6 `nemotron-infra-record` | not started | after W5 |

## Blocked on / next
- **Next:** the owner runs the `--no-prefill-assistant` flag for this model; then probe (d).
- Then W4 (the gateway alias), W5 (one agent-loop run) and W6 (the record), in that order.

## REDACTED
| kind | count |
|---|---|
| machine name (in the slug and in W2's slug) | 2 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17. No other machine is named here.
