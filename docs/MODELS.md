# Connecting and tuning models

Proven needs a model that can return the JSON requested by each agent. It can call a streaming OpenAI-compatible chat API or a local executable chosen by the operator. There is no mandatory model vendor, subscription or routing service.

Start with [Quick start](QUICKSTART.md) to separate recorded behavior from a live connection. The full variable catalogue is in [Configuration](CONFIGURATION.md).

## Choose a transport

| Mode | Use it for | Authentication and limits |
| --- | --- | --- |
| `off` | Deterministic checks without generated hypotheses | No model connection. Source/tool requirements still apply. |
| `openai` | A hosted or local streaming chat-completions API | Your endpoint, model and optional API key. Provider-specific options are explicit configuration. |
| `command` | A locally authenticated text-generation CLI or wrapper | The executable manages its normal login. Its permissions and subscription behavior need separate validation. |
| `nvidia` | Existing hosted NVIDIA configurations | Legacy adapter with NVIDIA-specific identity checks and options. |
| `nightshift` | Existing installations using that routing service | Legacy adapter; its receipt and routing settings do not apply to a generic API. |

An OpenAI-compatible API is a protocol choice; it does not require an OpenAI account. Compatibility is narrower than the label: the server must stream chat-completion events and the model must produce usable agent answers.

## Connect an API

For a local server, substituting the address and model you actually serve:

```sh
export PROVEN_GENERATOR_MODE=openai
export PROVEN_GENERATOR_BASE_URL=http://127.0.0.1:11434/v1
export PROVEN_GENERATOR_MODEL=your-model
export PROVEN_GENERATOR_TIMEOUT=300s
export PROVEN_GENERATOR_HEADER_TIMEOUT=40s
export PROVEN_GENERATOR_IDLE_TIMEOUT=90s
proven model check
```

For a hosted server, use its HTTPS API root and supply `PROVEN_GENERATOR_KEY` through your shell's secret handling or CI secret store. Do not put credentials in a committed profile, URL or `EXTRA_BODY`. Check that the selected endpoint is the intended recipient before sending source code.

`model check` makes **one HTTP request**, using the Investigator's options. It has no HTTP retries, empty-answer recovery or schema repair. Exit 0 means the small probe returned valid, nonempty JSON; exit 3 means the probe failed; exit 7 means configuration was refused. It does not run a whole review, exercise the command adapter or measure detection quality.

Keep the probe JSON when comparing configurations. It records observed compatibility and timings. A successful probe today does not establish another model's behavior or the latency of a long generated test.

## Know which stages call the model

| Agent | Requested work | Tuning implication |
| --- | --- | --- |
| Planner | Risks in the change; bounded carried state when Markovian is on | Large changes may require many sequential steps. Budget total review time, not only one response. |
| Investigator | Registered query names and parameters | Short structured answers; no SQL or shell commands. |
| Prover | A differential test, or an explicit abstention | Source and test output may need more response tokens than query selection. Generated-test support currently targets Go. |
| Skeptic | A refutation attempt and supporting evidence | A second opinion alone does not erase an observed proof. |
| Fixer | A proposed edit for a proved finding | Patch text may need a larger answer budget. Validation runs separately. |
| Judge / Writer | Deterministic adjudication / result assembly | No model call. |

`PROVEN_SKEPTIC_MODEL` selects another model on the primary endpoint. A separately configured second reviewer runs another review path; it is distinct from merely selecting a Skeptic model. Report requested and responding identities separately. An API's `model` label is its statement, not a signature proving who computed the answer. The command adapter reports the answering model as unknown.

## Three clocks, plus the review deadline

```text
review deadline ----------------------------------------------------|
     one completion attempt --------------------------------|
       connect / wait for headers | streamed body / reasoning |
       <--- header timeout ------>  idle clock resets on bytes
```

`PROVEN_GENERATOR_TIMEOUT` bounds the entire attempt, including any eligible HTTP retries and backoff. `HEADER_TIMEOUT` limits waiting for headers. `IDLE_TIMEOUT` limits silence after headers; every body read resets it, including reasoning fragments and SSE heartbeats. A stream that remains active can still reach the total attempt limit.

The CLI's `--timeout` is the whole review budget in seconds. The GitHub job can impose another outer limit. Increasing one attempt from 180 to 300 seconds does not enlarge either outer limit. Token budgets, number of planning steps, recovery, runner time and fixes all consume review time.

`PROVEN_GENERATOR_DEADLINE_RESERVE` leaves time to assemble the result. It is capped at one tenth of the original remaining review budget. With a reserve enabled, a new attempt needs a full call window before that reserve; otherwise the result is partial. See [Configuration](CONFIGURATION.md#call-limits-and-provider-options) for the exact admission rules.

## Separate reasoning from answer text

Some models stream reasoning before emitting `delta.content`. Proven observes both, so reasoning counts as activity. It does **not** pass reasoning to the agent decoder as if it were the answer. The trace records reasoning byte counts by default; optional bounded text capture is described in [Observability](OBSERVABILITY.md).

The response's `finish_reason` matters. `length` is a named failure even when some answer text arrived. If the provider counts reasoning against the response budget, a model may consume its allowance before producing usable JSON. Silence, reasoning-only output, truncation and invalid JSON are different failure modes; choose the control that matches the trace.

## Set a profile explicitly

This illustrates independent settings, rather than a preset guaranteed for every model:

```sh
export PROVEN_GENERATOR_MAX_TOKENS=8192
export PROVEN_GENERATOR_AGENT_MAX_TOKENS='{"planner":4096,"investigator":2048,"skeptic":2048,"prover":32768,"fixer":32768}'
export PROVEN_GENERATOR_EXTRA_BODY='{"temperature":0}'
export PROVEN_GENERATOR_RECOVERY_RETRIES=1
export PROVEN_GENERATOR_DEADLINE_RESERVE=60s
```

The common token limit is omitted by default. Agent overrides use the lower-case names in the table; zero omits the token field for that agent. If your provider requires it, select `PROVEN_GENERATOR_MAX_TOKENS_FIELD=max_completion_tokens` instead of `max_tokens`. Enable `INCLUDE_USAGE` only on endpoints accepting `stream_options.include_usage`.

`EXTRA_BODY` is a validated JSON object. `AGENT_EXTRA_BODY` maps agent names to option objects. Matching top-level fields from the agent replace common fields; nested objects are not recursively merged. Client-owned fields such as `messages`, `model`, `stream` and token controls cannot be overridden through these maps.

Thinking controls have no universal spelling or meaning. For a provider that documents a `chat_template_kwargs` object, an operator might compare thinking disabled, nonempty-content enforcement and a bounded reasoning allowance. Do not send these controls to another provider on assumption. Measure each chosen model once with `model check`, then run known defect/control cases before relying on the profile. The [readiness measurements](READINESS.md#what-has-been-measured) distinguish compatibility from quality.

## Recovery and HTTP retries

| Situation | Behavior |
| --- | --- |
| Empty answer or invalid agent JSON | Shares one configured additional-completion budget per agent step. Default 1 means at most two completions. |
| HTTP 429 before a successful stream | Optional rate-limit retries; honors `Retry-After` when it fits the configured cap and attempt deadline. |
| Connection failure or HTTP 500/502/503 before a successful stream | Optional transport retries, with their own counter. |
| Timeout, cancellation, HTTP 408/504, redirect, broken successful stream | No retry or schema recovery. |
| `finish_reason=length` | Explicit token-limit failure; inspect reasoning/output usage and tune the profile. |

Both HTTP retry counters default to zero. Backoff stays inside the attempt deadline. Rate-limit fallback is exponential; transport fallback is linear. A server delay larger than the cap skips the retry, rather than contacting the server sooner than requested. Automatic retries cannot repair a daily quota, wrong key or incompatible model.

JSON repair sends the requested schema and bounded previous answer. It omits the original source context by default. `PROVEN_GENERATOR_REPAIR_CONTEXT=true` restores that context if a measured model needs it. A recovery still needs a full configured call window in the remaining budget.

## Use a local CLI

```sh
export PROVEN_GENERATOR_MODE=command
export PROVEN_GENERATOR_MODEL=your-model
export PROVEN_GENERATOR_COMMAND='["/absolute/path/to/text-only-model-wrapper","--model","{model}","--role","{agent}"]'
```

Replace the example executable with your own installed wrapper. Proven starts argv directly, without a shell. It puts the prompt on stdin and expects only the final answer on stdout. The wrapper is trusted operator code: configure it for text generation, disable autonomous tools/edits, and validate it outside a sensitive repository first.

For Claude, use `--output-format text`. Its `json` and `stream-json` formats are vendor envelopes, not the requested agent object. Proven rejects an object with none of the expected fields. Configure Codex or Copilot wrappers to emit only final answer text too. Do not copy CLI flags between vendors or assume subscription terms and login locations are identical; inspect the installed executable's help.

The adapter inherits a small allowlist of ordinary process and login/configuration locations. It does not pass Proven's model/database keys as arbitrary environment variables. CLI stderr is not included in the report; a wrapper can retain its own access-controlled diagnostic log. Command mode is refused under `PROVEN_NETWORK=local`, because Proven cannot constrain an external program's networking with its HTTP policy.

## Add another reviewer

Set `PROVEN_SECOND_REVIEWER_BASE_URL` and `PROVEN_SECOND_REVIEWER_MODEL`, plus its optional key. Call controls have corresponding `PROVEN_SECOND_REVIEWER_` settings. Unset overrides inherit the primary controls; `{}` clears extra-body maps, `0` omits token limits and `false` disables inherited switches. Explicitly clear provider-specific options when the second endpoint is from a different vendor.

The report retains both reviewer roles, identities and disagreements. Shared family labels and unknown identity remain visible. Two configurations do not by themselves establish independent training or eliminate shared mistakes.

## Validate an operational profile

Record the model and endpoint identity, options, per-agent token limits, all timeouts, retry settings and Proven revision. Use a known defect, a matching fixed version and an unrelated clean change. Check generated tests actually ran, required data existed, unsupported claims were dropped and incomplete analysis remained visible. Compare time and failure rate as well as findings. See [Operations](OPERATIONS.md), [Troubleshooting](TROUBLESHOOTING.md) and [Markovian](MARKOVIAN.md).
