# TRAIN ORDER — █████-nemotron-content-safety

> **Public copy.** The original is in our private record. This copy says the same thing, with internal names, paths and
> references blacked out (█, one block per character). The list is at the end. The file name leaves out the machine
> name. The rule, in the owner's words (2026-10-17): *"We save in both places all hackathon documents."*
>
> **State of this document:** a draft. The advisor wrote it; the master may change it; the owner has not signed it yet.
> When the original changes, this copy changes with it.
>
> - written by: the advisor (claude-code, model `claude-opus-5-5`), 2026-10-18
> - cost: **not measured.** The advisor works in an interactive session, not a dispatched run, and Saphan Studio records
>   a cost only for runs. Not run: no cost figure exists for this document.

- master: the infra master, the same head that runs `proven-cursor-infra` (none appointed when this was written, 2026-10-18)
- status: PAUSED — drafted by the advisor; the master may change it, then hands it to the owner to sign
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

> *"make the order for █████ public - for the judges we will make option to run without guardrails. But we give
> guardrails as default as in saphan"*

⇒ **This is a hackathon document.** A redacted public copy lives in proven:
`fleet/orders/TRAIN_ORDER_nemotron-content-safety.md`. The public file name drops the machine name, which is blacked out.
proven checks content with a guardrail by default, as Saphan does, and has a switch that runs it without guardrails for
judges, whose machines cannot reach ours. That switch and the wiring are proven's build work from 2026-10-22, not this
train's.

**Why a separate train.** Different hosts, a different write set and a DoD of its own. Nothing in `proven-cursor-infra`
waits for it, and it waits for nothing there. One master runs both.

**Kin, not this train's work (law #59).** █████ already serves three safety judges. They were installed on 2026-10-05 by
the master of `guardrail-model-text` (now in `development/done/`), and its analysis listed **Nemotron-3.5-Content-Safety as
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
| N3 | **Prompt contract** (from the card): chat messages, `user` turn = the prompt (plus an optional image), optional `assistant` turn = the response to judge. Chat-template kwargs: `request_categories="/categories"`, `custom_policy`, `enable_thinking`. Output: `User Safety: safe\|unsafe`, `Response Safety: safe\|unsafe` (only with a response), `Safety Categories: …` (Aegis 2.0 taxonomy). **Whether llama-server `--jinja` renders this GGUF's template with those kwargs (`chat_template_kwargs` in the request) is UNMEASURED.** That is the risk the 2026-10-05 analysis already named. | the card; G3 measures it |
| N4 | **█████ today:** llama-swap v247, `active`, user `█████████`, `-listen :8081`, up 5 d 21 h. Resident `llama-server` processes: `qwen3-coder`, `Qwen3Guard-Gen-8B`, `Llama-Guard-3-8B`. `gpt-oss-safeguard-20b` is on disk and not loaded. llama-server build `e031d95`. | `ssh █████ 'systemctl is-active llama-swap; llama-swap --version; ps -eo pid,etime,rss,args \| grep llama-'` |
| N5 | **Memory:** GTT 112.00 GiB total, 34,664,476,672 B (32.3 GiB) used. `free -g`: 121 total, 65 used, 55 available. **Swap 7 of 7 GiB used: full.** `███████████`: 872 G free. | `cat /sys/class/drm/card0/device/mem_info_gtt_{total,used}; free -g; df -h ███████████` |
| N6 | **The live config is not readable by the advisor's user** (`███████████████████████████`: Permission denied; `████:█████████ 640` per the 2026-10-05 analysis §10). That analysis records group `safety` (persistent, non-exclusive, `ttl 0`, preloaded) holding `qwen3guard-8b` and `llama-guard-3-8b`, and `gpt-oss-safeguard-20b` in group `big`. **The repo copy `████████████/████████████████████████████` still shows the August config with no safety models: it is stale.** | `sed -n … ███████████████████████████` → Permission denied; `██████████████████████████████████████████████████████████████` §10 |
| N7 | **The gateway (August record):** LiteLLM on its own container behind the reverse proxy, DB-less and master-key-only by owner decision (██), four aliases, all `local-*`, upstream █████ `:8081`. **Whether any safety model has a gateway alias today is UNMEASURED.** | `████████████/████████████████████` (2026-08-06/10) |
| N8 | **The agent-loop lane for █████ bypasses the gateway:** `████████████` declares `base_url = ██████████████████████` (llm-routing ██, 2026-08-10). A run on `--lane █████` therefore tests llama-swap, not the LiteLLM registration. **Whether a lane whose `base_url` is the gateway exists is UNMEASURED** (the advisor's read of the config ladder was refused, and it did not read it another way, law #57). Dispatch form per engine docs: `saphan run --backend agent-loop --lane <alias>`. | engine `█████████████████████` (agent-loop lane); llm-routing.md ██ |
| N9 | **Fit, estimated, not measured:** a 4B model at Q8_0 is about 4–4.5 GiB of weights, `mmproj` about 0.6 GiB, KV at a 16K context about 1–2 GiB, so about 6–7 GiB in total. Resident today 32.3 GiB. The worst case is the resident set + the largest `big` member (`gpt-oss-120b`, a 63.4 GB file, about 59 GiB) + this model ≈ 98 GiB against 112 GiB GTT. That fits on GTT. But "available" system memory reads 55 GiB and swap is full (N5), so **unified-memory headroom with a `big` model loaded is the number to measure, not this estimate.** | arithmetic on N2, N5; the file sizes on disk |

## §2 WAGONS

Every wagon's DoD carries (i) the commands and their outputs, (ii) a `master-was-wrong:` field against the §1 rows it
builds on, and (iii) for each owner-act script, the rehearsal output pasted beside it (law #33). Cost class: local energy,
named zero (law #51). No wagon spends quota or money.

### G1 — re-measure, read the live config, and decide fit by printed arithmetic
- **lands:** N1–N9 re-measured, each marked holds / wrong / changed. The live `███████████████████████████` read with the
  access you hold. If you hold none, stop and ask the owner (§0) for the read. Never print an `apiKeys` value.
- **fit (law #46):** print the candidates and the numbers: resident set, the largest `big` member, this model at the
  chosen quant and context, GTT total, and system memory available. Measure the worst case with the `big` model actually
  loaded, not from file sizes. The verdict: FITS, or DOES NOT FIT with the shortfall in GiB.
- **"the least used warm model" — only if it does not fit:** warm = resident (`ttl 0` / persistent group): `qwen3-coder`,
  `qwen3guard-8b`, `llama-guard-3-8b`. Usage = requests per model over the same window for every candidate, from
  llama-swap's own record (its log or `/metrics`), with the window printed. **An absent reading is not "least used"
  (law #47): if you cannot read usage for every candidate, stop and ask.** Two owner rulings can collide here: if the
  least used is `qwen3-coder`, it is "never evicted, hot path" by contract (`████████████████████████████`, ███). Then
  stop and ask; do not pick the second least used on your own.
- **dod:** the arithmetic and the verdict pasted; the usage reading, or "not needed: FITS".
- **review:** master-verify-only

### G2 — the GGUF on █████, verified
- **lands:** the quant chosen by G1 (advisor's suggestion: `Q8_0`, because a 4B classifier is small and judge accuracy is
  worth the extra GiB; `Q4_K_M` if G1 says memory is tight), fetched into `███████████` the way
  `██████████████████████████` fetched the 2026-10-05 judges: resume, retries, and **sha256 equal to the Hugging Face LFS
  oid**, logged to `█████████████████`. Text-only first. `mmproj` is fetched only if the owner wants image judging; it is
  not in §0, so ask.
- **dod:** the log line `OK <file> sha256=<…> size=<…>` pasted. Negative control (law #27): a wrong expected sha makes the
  fetch refuse and leaves no file under the final name. Rehearse it on a tiny file and paste.
- **review:** master-verify-only

### G3 — warm on llama-swap, and it judges correctly
- **lands:** a model id (advisor's suggestion: `nemotron-3.5-content-safety`) in `███████████████████████████`, after a
  timestamped backup kept with the same owner and mode, in group `safety` (persistent, `ttl 0`) and in the startup
  preload, with `--jinja` and the same `kv`/`common` macros as the other judges. If G1 said DOES NOT FIT, the evicted
  model leaves the `safety` group and the preload in the same edit, and its file stays on disk. Restart only with no live
  connection on `:8081`, as on 2026-10-05.
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

### G4 — registered on the LiteLLM gateway
- **lands:** one alias in the gateway's `config.yaml` (advisor's suggestion: `local-safety-nemotron`, keeping the
  `local-*` contract), upstream the G3 id on █████ `:8081`, with the llama-swap consumer key the gateway already uses
  (per-consumer keys, ████████; ask whether a new consumer key is wanted). Back up, reload, and keep the four existing
  aliases unchanged.
- **dod:** `/health/liveliness` 200; an authenticated `GET /v1/models` lists the new alias and the four old ones (count
  printed); probe (b) of G3 through the gateway returns the same verdict. **Never probe the unauthenticated path for health**
  (llm-routing ██). Reaching the gateway with its master key is a secret act: if you do not hold it, stop and ask. Never
  print it.
- **review:** master-verify-only

### G5 — one agent-loop run on this model
- **lands:** exactly one `saphan run --backend agent-loop` on the new model. Compose the line from the binary
  (`saphan dispatch command` or the run verb's help), not from this order. **Prefer a lane whose `base_url` is the
  gateway, so the run also proves G4.** If none exists (N8), run on `--lane █████` and record that the run proved
  llama-swap, not the gateway. Adding a gateway lane is config in the owner's ladder: ask, do not add it. The brief asks
  the model to classify one prompt and return the verdict.
- **refutation step (law #44):** this is a classifier, not a coding agent. The loop may expect tool calls the model never
  makes. **Any honest outcome is a valid result:** `done`, or a named failure with its cause. Paste the engine's row
  (`status`, `exit_status`, refusals, the model that ran, cost `local-energy`) and the return. Do not shape the brief
  until the run "passes"; if the loop cannot carry a classifier, that is the finding.
- **dod:** the row and the return pasted, the lane named, the outcome classified.
- **review:** master-verify-only

### G6 — the record
- **lands:** in ████████████, a brief and a report (advisor's suggestion: `███████████████████████████████████████`,
  `███████████████████████████████████████████████`). `████████████████████████████` brought to the live config, keys
  redacted (it is stale since 2026-10-05, N6). `████████████████████`: the new alias in the alias table. The relay to
  `guardrail-server-images` (§0, kin).
- **dod:** `diff` of the live config against the repo copy is empty apart from the redacted keys. Paste it.
- **review:** master-verify-only

## §3 ORDER OF WORK

G1 → G2 → G3 → G4 → G5 → G6. Strictly sequential: each wagon proves what the next one stands on. No target date was
given. Because it runs beside `proven-cursor-infra`, the advisor suggests done before 2026-10-22, when the build window
opens and █████'s lanes get busier.

**Where the master stops and asks (the owner's rule, §0):** no read of the live config; usage not readable for every
warm model; the least used is `qwen3-coder`; `mmproj` (images) wanted or not; N3 refuted; no gateway master key; a new
llama-swap consumer key; a new gateway lane. Every question goes to the owner in one message, with the reading that
raised it.

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

---

## REDACTED

Parts of this file are blacked out (█). Nothing else was changed.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 19 | the machine that runs our local models, including inside this train's name |
| user names | 3 | the service user of the model server (twice), and the owner of its configuration file |
| internal paths | 12 | the model directory, the server's configuration file, the fetch script, the log directory, a lane's name and address, a file in our private record |
| private repository names | 13 | our infrastructure repository (6), and files in it (7) |
| internal design references | 6 | the numbers of our infrastructure decisions and of sections in our routing record |
| locations in our private code | 1 | a documentation file in Saphan Studio's engine |
