# ADR-0051: A judge's lab reaches our model service and our evidence store through one gateway, with the judge's own expiring token

**Status:** accepted · **Date:** 2026-10-26
**Decided by:** the owner · **Written by:** the executor of stream `judges-gateway` (claude-code, `claude-sonnet-5-5`), 2026-10-26
**Follows:** [ADR-0031](ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md) (a live review with one's own NVIDIA key), [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) (Nightshift is the primary provider), [ADR-0032](ADR-0032-proven-builds-every-query-from-a-template.md) (Proven builds every query from a template), [ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) (evidence through Zetaris), [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) (the judges' lab)
**Trigger:** the owner, 26 October: *"judges reach OUR Nightshift and OUR Zetaris through a Cloudflare tunnel, so a judge's lab can run a fully live review (NVIDIA via Nightshift, evidence via Zetaris) without their own accounts. Zero trust is zero trust."*

## Context

A judge who starts the lab gets a sample review. A live review needs two things the judge does not have: a Nightshift
runtime token (the model, NVIDIA's Nemotron by default) and a Zetaris key (the evidence). Both are ours. Handing them
to a stranger's machine is not an option: a copied key cannot be taken back without breaking every other judge.

## Decision

1. **One gateway on our host stands between the judge and both services.** It is `cmd/judges-gateway`, written with the
   standard library only. It listens on `127.0.0.1` only. A Cloudflare tunnel connects to it there. Nothing else can.
2. **Cloudflare Access sits in front of the tunnel.** The owner makes an Access application with a named allowlist of
   judges (an email each, or one service token each). Nothing anonymous gets through. This layer is Cloudflare's, set up
   by the owner; the runbook ([`docs/JUDGES_GATEWAY.md`](../JUDGES_GATEWAY.md)) says how.
3. **Each judge has their own token.** It is random (32 bytes), starts with `pjt_`, and has an expiry that the owner must
   give: no token lasts for ever. The gateway stores only its SHA-256, in a settings file with the token's id, expiry and
   limits. The token is printed once, when it is made. Revoking one token marks that one; the others go on working. The
   gateway re-reads the file when it changes, so a revocation works on the next call with no restart.
4. **Our two credentials never leave our side.** The gateway holds the Nightshift token and the Zetaris key in its
   environment. It builds each upstream request itself and puts our credential in it. The judge's `Authorization`
   header is dropped, never copied. The judge's answer is built from the upstream's body, never its headers; an upstream
   error is not passed on, because an error body could echo what was sent.
5. **The model route** is `POST /v1/chat/completions`, the path Proven's generator already calls, so a lab sets its
   generator base address to the tunnel, the judge's token as its runtime token, and changes nothing else. The gateway
   forwards only a request that names a model on its own list (default `meterless/auto`) and streams the answer back.
6. **The evidence route** is `POST /v1/evidence/query`. It accepts one thing: `{"template": <name>, "params": {...}}`.
   The gateway validates it with `internal/querytpl`, the same code Proven validates with, builds the SQL itself, and
   sends it to Zetaris with our key through Proven's own query client (one `SELECT`, a row limit, no redirect). Raw
   SQL, an extra field, an unknown template, a missing, extra or malformed parameter: each is refused by name, and
   nothing is sent to Zetaris. This is read-only by construction: the key can only run the templates.
7. **Limits per token**, from the settings file: calls per minute and calls per day. A call past a limit is refused by
   name (`rate_limit_minute`, `allowance_day`) and is not counted. When a window reaches 80% of its limit, one alarm
   line is written to stderr, once per window, starting `JUDGES-GATEWAY-ALARM` so the fleet's watchers can grep it.
8. **Every call is logged**, one JSON line: time, token id, route, model or template, status, bytes, and the refusal
   code when there is one.
9. **A call the gateway cannot log is not served for long.** The log line carries the status and the bytes sent, so it
   can only be written after the answer; it cannot be written before the call. We chose that the gateway **serves the
   call, makes the loss observable on a second channel, and refuses every later call until the sink works again**
   (fail closed after the first loss, not before it). The first write that fails (the call log or the alarm: a full
   disk, a broken pipe) does three things. It keeps the line in memory (at most 1024 lines). It writes one line on
   standard error that starts `JUDGES-GATEWAY-FAULT` and names the sink (`sink=log` or `sink=alarm`) and the error. It
   turns the health probe, `GET /healthz`, from 200 to 503. From then on each call is refused with `503
   log_unavailable` before the token is looked at and before any upstream is called. Each call, and `/healthz`, first
   tries to write the held lines; when they are written, one `JUDGES-GATEWAY-RECOVERED` line states how many lines were
   lost and how many refusals were not logged, the probe returns to 200, and calls are served again. So no request ends
   with a 200 whose log line is lost without the fault line and the probe saying so, and at most one call (the one that
   finds the sink broken, plus calls already in flight) is served unlogged for a moment: its line is held and written
   on recovery. The allowance alarm has the same treatment, because a lost alarm is a lost warning. We rejected
   writing the line before the call: it could not hold the status or the bytes, and a second line per call would
   break the one-line-per-call rule. If standard error fails too, nothing can name the fault, and only the probe says it.
10. **The model answer is bounded in bytes.** The gateway relays at most `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` bytes of a
    model answer per call (default 4 MiB, the size of the request bound; allowed 1 KiB to 64 MiB). A time limit is not
    a byte limit. If the model service declares a larger `Content-Length`, the judge gets `502 response_over_limit`
    before any byte. If it does not declare one and the stream passes the bound, the gateway has already sent `200`,
    so it relays exactly the bound and then aborts the connection: the judge's read ends in an error, never in a clean
    end. The log line has status 200, `bytes` equal to what was relayed, `refused: response_over_limit` and a `reason`
    that gives the bytes and the limit.

## The trust boundary

```
judge's lab ──▶ Cloudflare Access ──▶ tunnel ──▶ gateway (our host, 127.0.0.1) ──▶ Nightshift
                 allowlist of judges   cloudflared   judge token, limits, log   └──▶ Zetaris
```

| layer | what it refuses |
|---|---|
| Cloudflare Access | anyone not on the allowlist: no email, no service token, no entry |
| tunnel and `127.0.0.1` | a connection that does not come through the tunnel |
| gateway: sinks | every call, while the call log or the alarm cannot be written (`log_unavailable`) |
| gateway: token | a missing, unknown, foreign, expired or revoked token; a settings file it cannot read (the gateway then refuses every call) |
| gateway: route | any path other than the two; any method other than `POST` |
| gateway: limits | a call past the token's minute or day limit |
| gateway: model route | a model not on the list; a body over 4 MiB; an answer over the byte bound (declared: refused; undeclared: cut) |
| gateway: evidence route | anything but a valid template request; a body over 16 KiB |
| Nightshift, Zetaris | whatever their own services refuse; the gateway reports it only as `upstream_unavailable` or `upstream_error` |

## What a judge sees when refused

A JSON body with two fields: `refused`, a stable name, and `detail`, a fixed sentence. The detail never repeats what
the judge sent. The names: `missing_token`, `unknown_token`, `token_expired`, `token_revoked`, `rate_limit_minute`,
`allowance_day`, `no_such_route`, `method_not_allowed`, `body_too_large`, `bad_request`, `model_not_allowed`,
`not_a_template_request`, `unknown_template`, `missing_parameter`, `unknown_parameter`, `bad_parameter`,
`template_refused`, `upstream_unavailable`, `upstream_error`, `gateway_settings_unreadable`, `log_unavailable`,
`response_over_limit`. A cut stream has no refusal body: the judge's read ends in an error after the bound.

## What is logged, and what never is

Logged: the time; the token id (`-` when the token is unknown); the route; the model or the template name; the status;
the number of bytes sent back; the refusal code; for an upstream failure, a short reason such as `timeout`.

Never logged: the body of any request, the judge's code, the answer, any token or key, the `Authorization` header, the
parameter values of a template. A model name or a template name is written only when it is on the gateway's own list,
never as the caller typed it, so a judge cannot write text into the log.

## The token's life

Made by `judges-gateway token new --id <judge> --expires <time>` → sent to the judge privately → used until it expires
or is revoked → after the end of judging, the owner revokes what is left, or lets it expire. `token list` shows each
token's expiry and state, and never a token. The settings file holds hashes only and has mode `0600`.

## What is NOT protected

- **The judge's own code is sent to the model, by design.** A live review sends the change to NVIDIA's model through
  Nightshift. The gateway does not read it, log it or hide it. A judge who does not want that stays in sample mode.
- **The allowance counts calls, not tokens.** One model call can be large (up to 4 MiB of request). The model list and
  the call limit bound the spend; they do not meter it.
- **One call can be served with its log line held back.** The call that finds the sink broken is not stopped (decision
  9). If the process dies before the sink recovers, the held line is lost; the fault line on standard error is then
  the only trace.
- **Counts live in memory.** If the gateway restarts, each token's minute and day counts start again. A judge cannot
  cause a restart; a crash can.
- **The gateway does not check Cloudflare's own identity header.** It cannot verify one without fetching Cloudflare's
  keys, and a header that only the tunnel is supposed to add proves nothing if the tunnel is set up without Access. If
  Access is misconfigured, a judge's token is the only layer left, which is why the token is 256 bits and expires.
- **A stolen token works until it expires or is revoked**, from wherever Access lets the thief in. Revoke it alone.
- **Cloudflare sees the traffic.** The tunnel ends at Cloudflare's edge before it reaches us.
- **The host is ours.** The gateway's environment holds both credentials; whoever can read that process can read them.
- **The evidence an answer returns is whatever the templates select.** The templates read public advisory data; the
  gateway does not filter the rows further.

## Consequences

- A judge can run a fully live review with nothing of their own but the token the owner sends.
- The owner can end one judge's access in one command, and all of them by letting the date pass.
- Settings: `JUDGES_GATEWAY_TOKENS_FILE`, `JUDGES_GATEWAY_PORT`, `JUDGES_GATEWAY_LOG`, `JUDGES_GATEWAY_MODELS`,
  `JUDGES_GATEWAY_UPSTREAM_TIMEOUT`, `JUDGES_GATEWAY_MAX_RESPONSE_BYTES`, and the credential and address variables Proven already reads. None is hardcoded
  (ADR-0013); the defaults are in the runbook.
- The lab side (the lab's `.env` and the Gitea Actions secrets) and the lab's "Go live with our services" section are the
  next round and are not part of this decision.
