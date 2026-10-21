# TRAIN ORDER — █████-nemotron-content-safety

> **Public copy.** The original is in our private record. This copy says the same thing, with internal names, paths and
> references blacked out (█, one block per character). The list is at the end. The file name leaves out the machine
> name. The rule, in the owner's words (2026-10-17): *"We save all hackathon documents in both places."*
>
> **State of this document:** DONE. The advisor wrote it on 2026-10-18. The master re-measured the advisor's claims
> and amended the order on 2026-10-19 (§1a). The owner signed it on 2026-10-19. The master added the owner's answers and
> the W1 measurement on 2026-10-20, and closed the train on 2026-10-21 (CLOSING, below). This copy follows the original
> as of 2026-10-21.
>
> - written by: the advisor (claude-code, model `claude-opus-5-5`), 2026-10-18
> - amended by: the master (claude-code, model `claude-opus-5-5`), 2026-10-19
> - cost: **not measured.** The advisor and the master work in interactive sessions, not dispatched runs, and Saphan
>   Studio records a cost only for runs. Not run: no cost figure exists for this document.

- master: MASTER (claude-code, claude-opus-5-5, seat macm5/██████████████), the infra master of `proven-cursor-infra`, appointed by the owner on 2026-10-19, mode: economy
- status: DONE — 2026-10-21: W1–W6 done; the W5 stream closed without a merge (its closure closed)
- note: the wagons W1–W6 are the advisor's G1–G6, renamed by the master on 2026-10-19 because the train-order grammar counts W-numbers only
- track: █████-nemotron-content-safety
- runs: IN PARALLEL with `proven-cursor-infra`, as a separate train. The two share no host and no file, except new
  briefs and reports in ████████████, which are different files.
- repo: ████████████ (briefs, reports, `state/`) and the bridge registry. Hosts: █████ (llama-swap) and the LiteLLM gateway.
- base: main at the cut
- author: ADVISOR of the hackathon track "The Agent That Can Explain Why", by the owner's explicit hand-over (§0)
- public copy: proven `fleet/orders/TRAIN_ORDER_nemotron-content-safety.md` (the owner, 2026-10-18)

## §0 THE OWNER, VERBATIM

2026-10-18, to the advisor (written in English):

> *"We need to prepare our local LLM infrastructure for the NVIDIA model: Nemotron 3.5 Content Safety (4B), our local
> guardrail model.*
>
> *Write this as a parallel train order for the infra master.*
>
> *Target setup:*
> *1. The NVIDIA model is installed on █████ as a warm model. If it does not fit on that machine, take out the least used
>    warm model and replace it with the NVIDIA one.*
> *2. The model is registered on the LiteLLM proxy.*
> *3. The agent-loop backend is tested with one test run on this model.*
>
> *If the master needs anything from me, the master stops and asks.*
>
> *The master runs this at the same time as the other work. Make it a separate train, or attach it as wagons that run in
> parallel."*

2026-10-18, later, to the advisor (written in English):

> *"Make the order for █████ public. For the judges we will add an option to run without guardrails, but guardrails
> are on by default, as in Saphan."*

⇒ **This is a hackathon document.** A redacted public copy lives in proven:
`fleet/orders/TRAIN_ORDER_nemotron-content-safety.md`. The public file name drops the machine name, which is blacked out.
proven checks content with a guardrail by default, as Saphan does, and has a switch that runs it without guardrails for
judges, whose machines cannot reach ours. That switch and the wiring are proven's build work from 2026-10-22, not this
train's.

**Why a separate train.** Different hosts, a different write set and a DoD of its own. Nothing in `proven-cursor-infra`
waits for it, and it waits for nothing there. One master runs both.

**Kin, not this train's work (law #59).** █████ already serves three safety judges. They were installed on 2026-10-05 by
the master of `guardrail-model-text` (now in `█████████████████`), and its analysis listed **Nemotron-3.5-Content-Safety as
an alternative kept** (§9: "NATIVE NeMo parser, custom policies; OpenMDW + Gemma Terms; llama.cpp template rendering
UNVERIFIED"). The train `guardrail-server-images` (backlog, not cut) will point NeMo's LLM-judge rails at █████'s models.
This train installs and proves the model. It does NOT wire it into NeMo, the images or the engine's guardrail adapter.
When this train lands, the master writes one relay for whoever cuts `guardrail-server-images`: the model id, the endpoint,
the measured verdict format, and the latency.

## §1 WHAT THE ADVISOR MEASURED — CLAIMS FOR YOU TO TEST, not ground (law #44)

Re-measure each one before you build on it. Where the advisor is wrong, write it in the wagon's `master-was-wrong:` field.

| # | claim (advisor, MEASURED 2026-10-18 unless marked) | instrument |
|---|---|---|
| N1 | **The model.** `nvidia/Nemotron-3.5-Content-Safety`: Gemma-3-4B-it fine-tune (LoRA merged), 128K context, text and image input, 12 languages. Licence: OpenMDW-1.1 plus the Gemma Terms of Use and the Gemma Prohibited Use Policy. The official repository has BF16 safetensors only (about 8.6 GB), no GGUF. Its card names transformers, vLLM and SGLang, not llama.cpp. | huggingface.co/nvidia/Nemotron-3.5-Content-Safety (read 2026-10-18) |
| N2 | **A community GGUF exists:** `mradermacher/Nemotron-3.5-Content-Safety-GGUF`, created 2026-06-05, architecture `gemma3`. It has `Q4_K_M`, `Q8_0`, `f16` and the rest, plus `mmproj-Q8_0` and `mmproj-f16` for images. **File sizes and sha256 are UNMEASURED.** | huggingface.co/api/models/mradermacher/Nemotron-3.5-Content-Safety-GGUF |
| N3 | **Prompt contract** (from the card): chat messages, `user` turn = the prompt (plus an optional image), optional `assistant` turn = the response to judge. Chat-template kwargs: `request_categories="/categories"`, `custom_policy`, `enable_thinking`. Output: `User Safety: safe\|unsafe`, `Response Safety: safe\|unsafe` (only with a response), `Safety Categories: …` (Aegis 2.0 taxonomy). **Whether llama-server `--jinja` renders this GGUF's template with those kwargs (`chat_template_kwargs` in the request) is UNMEASURED.** That is the risk the 2026-10-05 analysis already named. | the card; W3 measures it |
| N4 | **█████ today:** llama-swap v247, `active`, user `█████████`, `-listen :8081`, up 5 d 21 h. Resident `llama-server` processes: `qwen3-coder`, `Qwen3Guard-Gen-8B`, `Llama-Guard-3-8B`. `gpt-oss-safeguard-20b` is on disk and not loaded. llama-server build `e031d95`. | `ssh █████ 'systemctl is-active llama-swap; llama-swap --version; ps -eo pid,etime,rss,args \| grep llama-'` |
| N5 | **Memory:** GTT 112.00 GiB total, 34,664,476,672 B (32.3 GiB) used. `free -g`: 121 total, 65 used, 55 available. **Swap 7 of 7 GiB used: full.** `███████████`: 872 G free. | `cat /sys/class/drm/card0/device/mem_info_gtt_{total,used}; free -g; df -h ███████████` |
| N6 | **The live config is not readable by the advisor's user** (`███████████████████████████`: Permission denied; `████:█████████ 640` per the 2026-10-05 analysis §10). That analysis records group `safety` (persistent, non-exclusive, `ttl 0`, preloaded) holding `qwen3guard-8b` and `llama-guard-3-8b`, and `gpt-oss-safeguard-20b` in group `big`. **The repo copy `████████████/████████████████████████████` still shows the August config with no safety models: it is stale.** | `sed -n … ███████████████████████████` → Permission denied; `██████████████████████████████████████████████████████████████` §10 |
| N7 | **The gateway (August record):** LiteLLM on its own container behind the reverse proxy, DB-less and master-key-only by owner decision (██), four aliases, all `local-*`, upstream █████ `:8081`. **Whether any safety model has a gateway alias today is UNMEASURED.** | `████████████/████████████████████` (2026-08-06/10) |
| N8 | **The agent-loop lane for █████ bypasses the gateway:** `████████████` declares `base_url = ██████████████████████` (llm-routing ██, 2026-08-10). A run on `--lane █████` therefore tests llama-swap, not the LiteLLM registration. **Whether a lane whose `base_url` is the gateway exists is UNMEASURED** (the advisor's read of the config ladder was refused, and it did not read it another way, law #57). Dispatch form per engine docs: `saphan run --backend agent-loop --lane <alias>`. | engine `█████████████████████` (agent-loop lane); llm-routing.md ██ |
| N9 | **Fit, estimated, not measured:** a 4B model at Q8_0 is about 4–4.5 GiB of weights, `mmproj` about 0.6 GiB, KV at a 16K context about 1–2 GiB, so about 6–7 GiB in total. Resident today 32.3 GiB. The worst case is the resident set + the largest `big` member (`gpt-oss-120b`, a 63.4 GB file, about 59 GiB) + this model ≈ 98 GiB against 112 GiB GTT. That fits on GTT. But "available" system memory reads 55 GiB and swap is full (N5), so **unified-memory headroom with a `big` model loaded is the number to measure, not this estimate.** | arithmetic on N2, N5; the file sizes on disk |

## §1a THE MASTER'S RE-MEASUREMENT (MASTER, 2026-10-19 about 15:10 UTC, read-only)

These are the master's readings. They are claims too: a leg or a later reading that finds one wrong says so.

| # | verdict | what the master measured |
|---|---|---|
| N1 | HOLDS | Not gated. OpenMDW-1.1 + Gemma Terms + Gemma Prohibited Use Policy. 5 safetensors shards, 8,600,276,752 B, no GGUF. Runtime engines named: Transformers, vLLM, SGLang. |
| N2 | HOLDS, now MEASURED | `Q8_0.gguf` 4,130,408,096 B, oid `217c703e74074fa57116fee18eb4531e4c7cbbf1e2ed7d626076004900cb7884`; `Q4_K_M.gguf` 2,489,899,936 B, oid `c8a65e9ef2f8d7a20b0ffe9df6d5cd73767ae0bc9a6953acc8239dd0c929e52d`; `mmproj-Q8_0.gguf` 591,378,080 B, oid `4fa856d65fc512f086b1f4b7f1ffbc36e27422367c1c0652f694d66dbcc96769`. No public imatrix repo. |
| N3 | HOLDS, wider | The GGUF's embedded template is the official `chat_template.jinja` (6,604 B). It reads `request_categories`, `custom_policy`, `enable_thinking`, and a fourth kwarg, **`custom_taxonomy`** (it refuses `custom_policy` and `custom_taxonomy` together). Thinking mode emits a `</think>` block before the verdict. Rendering by llama-server is still UNMEASURED (W3). |
| N4 | HOLDS | llama-swap v247 (40027d6), user `█████████`, `-listen :8081`, up since 2026-10-13. Resident: qwen3-coder (ctx 65536), Qwen3Guard-Gen-8B and Llama-Guard-3-8B (ctx 32768), all `--flash-attn on`, KV `q8_0`, `--jinja`. llama-server `e031d95`. |
| N5 | CHANGED (small) | GTT 112.00 GiB, 32.28 GiB used (unchanged). `free -g`: 69 used, **51 available**. Swap 8 of 8 G: full. `███████████` 872 G free. |
| N6 | HOLDS | `config.yaml` is `████:█████████ 640`; the master's user is not in `█████████`. Read refused; no other path was tried (law #57). The repo copy is the August file (2958 B), stale. |
| N7 | HOLDS (record) | Four `local-*` aliases, no safety alias; the live gateway config (`████████████████████████` on its container) is UNREAD. Liveliness answers 200. |
| N8 | **WRONG** | A lane on the gateway exists: `lane.litellm.base_url` = the gateway's `/v1` through the reverse proxy, credential name `███████████████████`. W5 runs on it. |
| N9 | estimate HOLDS, worst case WRONG | This model: Q8_0 3.85 GiB + KV at 16K about 1.1 GiB (q8_0) + buffers ≈ **5.5–6.5 GiB**. The largest `big` member on disk is **glm-4.5-air Q4_K_M, 67.96 GiB**, not gpt-oss-120b (59.03 GiB). Worst case: 32.28 + 67.96 + 6.5 ≈ **106.7 of 112.00 GiB GTT**, before GLM's own KV. System memory (51 GiB available, swap full) is the tighter limit. Whether the live config still points `big` at that file is unread (N6). |
| usage | UNMEASURED per model | The journal is readable, but request lines carry no model name. 27 chat requests in 7 days, all on 2026-10-15, all from one client; the model per request is unknown. `/metrics` needs a key. |

## §2 WAGONS

Every wagon's DoD carries (i) the commands and their outputs, (ii) a `master-was-wrong:` field against the §1 rows it
builds on, and (iii) for each owner-act script, the rehearsal output pasted beside it (law #33). Cost class: local energy,
named zero (law #51). No wagon spends quota or money.

### W1 · `nemotron-fit-measured` — re-measure, read the live config, and decide fit by printed arithmetic
- **lands:** N1–N9 re-measured, each marked holds / wrong / changed: done by the master on 2026-10-19 (§1a). The live
  `███████████████████████████` read: the master holds no read (N6), so the owner runs `█████-read-llama-swap-config.sh`. It
  writes a copy with every `apiKeys` value replaced by █ into the master's home on █████ (mode 600), and prints only the
  line count and the number of redacted values. Never print an `apiKeys` value.
- **fit (law #46):** print the candidates and the numbers: resident set, the largest `big` member named by the LIVE config,
  this model at the chosen quant and context, GTT total, and system memory available. The verdict: FITS, or DOES NOT FIT
  with the shortfall in GiB. Measuring the worst case with the `big` model actually loaded means loading about 68 GiB on a
  host whose swap is full (§1a N5, N9): the master asks the owner before that load (§3), and without his word the verdict
  rests on file sizes and says so.
- **"the least used warm model" — only if it does not fit:** warm = resident (`ttl 0` / persistent group): `qwen3-coder`,
  `qwen3guard-8b`, `llama-guard-3-8b`. Usage = requests per model over the same window for every candidate, from
  llama-swap's own record (its log or `/metrics`), with the window printed. **An absent reading is not "least used"
  (law #47): if you cannot read usage for every candidate, stop and ask.** Two owner rulings can collide here: if the
  least used is `qwen3-coder`, it is "never evicted, hot path" by contract (`████████████████████████████`, ███). Then
  stop and ask; do not pick the second least used on your own.
- **dod:** the arithmetic and the verdict pasted; the usage reading, or "not needed: FITS".
- **review:** master-verify-only

### W2 · `nemotron-gguf-on-█████` — the GGUF on █████, verified
- **lands:** the quant chosen by W1 (advisor's suggestion: `Q8_0`, because a 4B classifier is small and judge accuracy is
  worth the extra GiB; `Q4_K_M` if W1 says memory is tight), fetched into `███████████` the way
  `██████████████████████████` fetched the 2026-10-05 judges: resume, retries, and **sha256 equal to the Hugging Face LFS
  oid**, logged to `█████████████████`. Text-only first. `mmproj` is fetched only if the owner wants image judging; it is
  not in §0, so ask.
- **dod:** the log line `OK <file> sha256=<…> size=<…>` pasted. Negative control (law #27): a wrong expected sha makes the
  fetch refuse and leaves no file under the final name. Rehearse it on a tiny file and paste.
- **review:** master-verify-only

### W3 · `nemotron-warm-on-llama-swap` — warm on llama-swap, and it judges correctly
- **lands:** a model id (advisor's suggestion: `nemotron-3.5-content-safety`) in `███████████████████████████`, after a
  timestamped backup kept with the same owner and mode, in group `safety` (persistent, `ttl 0`) and in the startup
  preload, with `--jinja` and the same `kv`/`common` macros as the other judges. If W1 said DOES NOT FIT, the evicted
  model leaves the `safety` group and the preload in the same edit, and its file stays on disk. Restart only with no live
  connection on `:8081`, as on 2026-10-05. The edit and the restart need root, so they are the owner's
  `█████-nemotron-warm.sh`: backup, apply the master's diff (pasted beside the script), validate, wait for no connection,
  restart, print `/running`. The master composes the diff from the copy W1 produced.
- **template kwargs:** the template also reads `custom_taxonomy`, and refuses it together with `custom_policy` (§1a N3).
  Probes send `enable_thinking: false` unless a probe tests thinking.
- **dod:** after the restart, `/running` (with a consumer key) lists the new id as ready, and `qwen3-coder` plus every
  judge still resident is ready again. Then, through llama-swap, at temperature 0:
  - **(a)** a benign prompt (the 2026-10-05 probe: "sort a list in Python") → `User Safety: safe`;
  - **(b)** the 2026-10-05 malicious probe → `User Safety: unsafe` with its categories;
  - **(c)** the same as (b) with `chat_template_kwargs` `request_categories="/categories"` → categories present;
  - **(d)** a prompt plus an `assistant` response → a `Response Safety:` line.

  Latency for each, beside the other judges' (355–577 ms and about 300 ms on 2026-10-05). **If (a)–(d) do not come back
  in the card's format, N3 is refuted: write it down, stop, and ask.** Do not hand-write a template to force the format;
  that is a decision for the owner.
- **review:** master-verify-only

### W4 · `nemotron-on-litellm-gateway` — registered on the LiteLLM gateway
- **lands:** one alias in the gateway's `config.yaml` (advisor's suggestion: `local-safety-nemotron`, keeping the
  `local-*` contract), upstream the W3 id on █████ `:8081`, with the llama-swap consumer key the gateway already uses
  (per-consumer keys, ████████; ask whether a new consumer key is wanted). Back up, reload, and keep the four existing
  aliases unchanged.
- **dod:** `/health/liveliness` 200; an authenticated `GET /v1/models` lists the new alias and the four old ones (count
  printed); probe (b) of W3 through the gateway returns the same verdict. **Never probe the unauthenticated path for health**
  (llm-routing ██). Reaching the gateway with its master key is a secret act: if you do not hold it, stop and ask. Never
  print it. The config edit and the reload need root on the gateway's container, so they are the owner's
  `gateway-add-nemotron-alias.sh`, with the master's diff pasted beside it. The authenticated checks use the
  `███████████████████` the `litellm` lane already names (§1a N8), read on the master's host and never printed.
- **review:** master-verify-only

### W5 · `nemotron-agent-loop-run` — one agent-loop run on this model
- **lands:** exactly one `saphan run --backend agent-loop` on the new model. Compose the line from the binary
  (`saphan dispatch command` or the run verb's help), not from this order. **The run uses `--lane litellm`**, whose
  `base_url` is the gateway (§1a N8), so it also proves W4. The model is the W4 alias. The brief asks
  the model to classify one prompt and return the verdict.
- **refutation step (law #44):** this is a classifier, not a coding agent. The loop may expect tool calls the model never
  makes. **Any honest outcome is a valid result:** `done`, or a named failure with its cause. Paste the engine's row
  (`status`, `exit_status`, refusals, the model that ran, cost `local-energy`) and the return. Do not shape the brief
  until the run "passes"; if the loop cannot carry a classifier, that is the finding.
- **dod:** the row and the return pasted, the lane named, the outcome classified.
- **review:** master-verify-only

### W6 · `nemotron-infra-record` — the record
- **lands:** in ████████████, a brief and a report (advisor's suggestion: `███████████████████████████████████████`,
  `███████████████████████████████████████████████`). `████████████████████████████` brought to the live config, keys
  redacted (it is stale since 2026-10-05, N6). `████████████████████`: the new alias in the alias table. The relay to
  `guardrail-server-images` (§0, kin).
- **dod:** `diff` of the live config against the repo copy is empty apart from the redacted keys. Paste it.
- **review:** master-verify-only

## §3 ORDER OF WORK

W1 → W2 → W3 → W4 → W5 → W6. Strictly sequential: each wagon proves what the next one stands on. No target date was
given. Because it runs beside `proven-cursor-infra`, the advisor suggests done before 2026-10-22, when the build window
opens and █████'s lanes get busier.

**Where the master stops and asks (the owner's rule, §0):** no read of the live config; usage not readable for every
warm model; the least used is `qwen3-coder`; `mmproj` (images) wanted or not; N3 refuted; no gateway master key; a new
llama-swap consumer key; a new gateway lane; loading the largest `big` model to measure the worst case (added 2026-10-19).
Every question goes to the owner in one message, with the reading that raised it.

**Questions the master asks now (2026-10-19), with the readings in §1a:**
- **Q1 — images:** text only (recommended: §0 asks for a guardrail, not an image judge; +0.55 GiB if yes), or also `mmproj`?
- **Q2 — worst-case load:** may the master load the largest `big` model once at a quiet hour to measure (recommended: yes,
  one request, the model unloads after its ttl), or does the verdict rest on file sizes?
- **Q3 — consumer key:** the gateway's existing llama-swap consumer key (recommended), or a new one for this alias?

**The owner's answers (2026-10-20):** Q1 — text only, no `mmproj`. Q2 — yes, load the largest `big` model once.
Q3 — the existing llama-swap consumer key.

**W1 measured (MASTER, 2026-10-20):** at 15:12:52Z one request loaded `glm-4.5-air` Q4_K_M (67.96 GiB) next to the warm
set (GTT 32.3 GiB used, 41 GiB available, swap 8/8 G full). The kernel's global OOM fired at 15:13:44Z; █████ stopped
answering (ssh banner and llama-swap timed out) until a reboot at 16:19Z. **Verdict: the worst case DOES NOT FIT today,
even WITHOUT this model**; the shortfall is not measurable, because the host stops before it can be read. Adding this
model as warm adds about 6.5 GiB to the set that is already over budget when a large `big` model loads. Per W1, the
master stops and asks the owner (§3, Q4).
- **Q4 — the warm set (asked 2026-10-20):** (a) add this model as warm, and the owner changes the `big` group's budget
  separately (out of this train); (b) evict one judge (`qwen3guard-8b` or `llama-guard-3-8b`) — per-model usage is
  unreadable (§1a), so the owner names it; (c) run this model on demand (`ttl`), not warm.

**Owner acts in this train** (root on █████ and on the gateway's container): `█████-read-llama-swap-config.sh` (W1),
`█████-nemotron-warm.sh` (W3), `gateway-add-nemotron-alias.sh` (W4). Each is rehearsed by the master, and the rehearsal
output is pasted beside it (law #33). The fetch (W2) runs as the master's own user, into `███████████`, with the
2026-10-05 fetch script's checks.

## §4 DEFINITION OF DONE

- The model on █████, sha-verified, warm in group `safety`, preloaded, and judging in the card's format, with latency.
- Fit decided by printed arithmetic; any eviction backed by a usage reading for every warm model.
- One gateway alias; the four old aliases unchanged; the same verdict through the gateway.
- One agent-loop run, its row and return pasted, its lane named, its outcome honestly classified.
- ████████████'s `state/` matches the live hosts. The relay to `guardrail-server-images` is written.
- The closing line of law #38 in this file.

## §5 OUT OF SCOPE

- Wiring the model into NeMo, the guardrail images or the engine's guardrail adapter (`guardrail-server-images`,
  `guardrail-model-text`).
- proven's code. proven uses this model as its default guardrail (the owner, 2026-10-18, §0), and a judge runs proven
  without it. Calling the model from proven and the switch that turns it off are built in proven from 2026-10-22. This
  train only makes the model ready on █████ and on the gateway.
- **The public copy (law #43 / proven `AGENTS.md` §1):** the master updates proven's
  `fleet/orders/TRAIN_ORDER_nemotron-content-safety.md` whenever this file changes, and each wagon's public status goes
  under proven's `fleet/` templates, with █████, internal paths, user names, private repository names and internal design
  references blacked out.
- Deleting any model file. Changing `qwen3-coder`'s hot-path contract. Any vendor key, any priced act.

## CLOSING (MASTER, 2026-10-21)

- W1 measured (the worst case does not fit; owner Q4 (b): llama-guard-3-8b left the warm set). W2 fetched, sha256 equal.
  W3 warm, probes a–d in the card's format (d after `--no-prefill-assistant`). W4 gateway alias, 5 aliases, probe b via the
  gateway. W5 one agent-loop run: the loop carries the judge's call, not a classification task. W6 record: ████████████
  ███████; relay to `guardrail-server-images` (bridge █████████). No tester: the train changed hosts, not a product.
- Open, outside this train: the gateway's configuration pin was not re-measured after W4 (████████████ report).

DONE · █████-nemotron-content-safety · nothing in flight · nothing owed to the owner · THANK YOU

---

## REDACTED

Parts of this file are blacked out (█). Nothing else was changed, except in the CLOSING section (2026-10-21): the
closure's identifier is left out of the status line, and the name of the gateway's configuration pin is written as
"the gateway's configuration pin".

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17 (the Cursor executors work there).

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 27 | the machine that runs our local models, including inside this train's name, one wagon's name and two script names |
| user names | 7 | the service user and group of the model server, and the owner of its configuration file |
| internal paths | 16 | the model directory, the server's configuration file, the gateway's configuration file, the fetch script, the log directory, a lane's name and address, a folder and a file in our private record |
| private repository names | 15 | our infrastructure repository (8), and files in it (7) |
| private commits | 2 | a commit of our infrastructure repository, and a commit of our private record |
| internal design references | 6 | the numbers of our infrastructure decisions and of sections in our routing record |
| locations in our private code | 1 | a documentation file in Saphan Studio's engine |
| credentials | 2 | the name of the variable that holds the gateway's key |
| seat identities | 1 | the master's seat |
