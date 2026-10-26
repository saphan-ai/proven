# ADR-0051: A judge's lab reaches our model service and our evidence store through one gateway, behind Cloudflare Access (the judge's own expiring token is the fallback)

**Status:** accepted · **Date:** 2026-10-26 · **Amended:** 2026-10-26, round 3 (Access first, token fallback, the runner class, the infra scripts)
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
2. **Cloudflare Access sits in front of the tunnel, and the gateway trusts only what Access asserts.** Access is the
   first method (`JUDGES_GATEWAY_AUTH=access`, the default). A judge signs in with an e-mail one-time PIN: Cloudflare
   mails a code to an address on the allowlist, so no secret of ours is ever sent to the judge. The GitHub workflow,
   which runs on a GitHub-hosted runner and has no browser, presents an Access **service token**. Cloudflare checks
   the PIN or the token, then adds the header `Cf-Access-Jwt-Assertion`, a JWT it signs. The gateway verifies that JWT
   itself and serves a call only if it is valid for a listed identity (decision 11). Nothing anonymous gets through.
   The Access application, its policies, the tunnel and the DNS record are made by scripts in the repository
   (decision 13), so the setup is code that is read, planned and removed, not a dashboard recipe.
3. **The fallback is each judge's own token** (`JUDGES_GATEWAY_AUTH=token`, the round-1 method). It is random (32
   bytes), starts with `pjt_`, and has an expiry that the owner must give: no token lasts for ever. The gateway stores
   only its SHA-256, in a settings file with the token's id, expiry and limits. The token is printed once, when it is
   made. Revoking one token marks that one; the others go on working. The gateway re-reads the file when it changes,
   so a revocation works on the next call with no restart. The method is a setting, not a code path that is always
   on: in `access` mode a token is not a way in, and in `token` mode an Access header is not.
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
8. **Every call is logged**, one JSON line: time, the caller (the token id in the token method; in the Access method
   the identity, which is an e-mail or a service token's client id, and its class), route, model or template, status,
   bytes, and the refusal code when there is one.
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
11. **The gateway verifies the Access assertion itself, with the standard library only.** The header
    `Cf-Access-Jwt-Assertion` must be an RS256 JWT. The gateway fetches the team's signing keys from
    `https://<team>.cloudflareaccess.com/cdn-cgi/access/certs` and matches the token's key id. It keeps the set for ten
    minutes; a key id it does not know triggers one new fetch, at most once a minute, so a caller who invents key ids
    cannot make the gateway call Cloudflare on every request. If a refresh fails, the old set stays in use; if there is
    no set at all the call gets `503 access_keys_unavailable`. Keys under 2048 bits are not accepted. The signature is
    checked before any claim is believed. Then: the issuer must be the team domain, the audience must contain the
    application's AUD tag (a setting), the token must not be expired (30 seconds of clock leeway) and must not be
    used before `nbf`. Any algorithm but RS256 (including `none`) is refused. The identity is the `email` claim (a
    judge) or, when there is none, the `common_name` claim (a service token's client id).
12. **Identities are rows in the settings file, and each is limited, logged and revoked alone.** A row is an e-mail, an
    e-mail domain (`*@jury.example`) or a service token's client id, with a class (`judge` for an address or a domain,
    `runner` for a client id; a row that mixes them is refused when the file is read), an expiry that is required, a
    per-minute and a per-day limit, and a revoked flag. An identity with a valid assertion and no row is refused by
    name (`identity_not_listed`): passing Access is not being listed. An exact row beats a domain row, so one address of
    a domain is cut off by an exact row that is revoked. A domain row admits only that domain (not a subdomain, not a
    longer name that begins with it), and each address under it is counted and limited on its own. The e-mail is
    compared in lower case; an e-mail claim never matches a client-id row and a client id never matches an address.
    The `judges-gateway identity add|revoke|list` commands write the file; the gateway re-reads it when it changes.
13. **The Access setup is code: `infra/cloudflare/`.** POSIX `sh`, `curl` and `jq`, against the Cloudflare API v4.
    `up.sh plan` reads and prints what would change and changes nothing. `up.sh apply` finds each object by name
    first and creates only what is missing: the named tunnel and its ingress to `127.0.0.1:<gateway port>`, the DNS
    CNAME for the one host (tagged in its comment), the e-mail one-time-PIN identity provider (only if the account has
    none), the Access application, the runner's service token (its secret goes once to a mode `0600` file and never to
    a screen or a log), and two policies: the judges' allowlist (Allow, by e-mail or domain) and the runner's
    service-auth policy. It writes the application's AUD tag and the team domain to the gateway's environment file and
    can add the matching identity rows to the gateway's settings. A second run changes nothing. `down.sh` removes, in
    reverse order, only what `up.sh` recorded as created (by id, in a state file) and only if the name matches; a
    record the script did not tag is never deleted; an identity provider that existed before is left. The Cloudflare
    API token comes from the environment or a file and reaches `curl` on its standard input, never on a command line,
    never in a log; every API error is printed as its Cloudflare code and message.
14. **What Access logs and what the gateway logs are two records.** Cloudflare's Access log says who signed in and from
    where, and is Cloudflare's. The gateway's log says what was asked of our two services under that identity, and is
    ours. Neither holds the other's data: the gateway never logs the assertion or a PIN, and it cannot see who tried and
    failed at the Access door.

## The trust boundary

```
judge (e-mail PIN) ─┐
                    ├─▶ Cloudflare Access ──▶ tunnel ──▶ gateway (our host, 127.0.0.1) ──▶ Nightshift
GitHub runner ──────┘   allowlist, service     cloudflared   verifies the assertion,      └──▶ Zetaris
(service token)         token; signs a JWT                    identity row, limits, log
```

| layer | what it refuses |
|---|---|
| Cloudflare Access | anyone not on the allowlist: no e-mail, no service token, no entry |
| tunnel and `127.0.0.1` | a connection that does not come through the tunnel |
| gateway: sinks | every call, while the call log or the alarm cannot be written (`log_unavailable`) |
| gateway: assertion (Access method) | no header, a malformed one, a key the team does not publish, a bad signature, an expired one, another audience, another issuer, no e-mail and no client id; the signing keys unreadable (`503`) |
| gateway: identity (Access method) | a valid assertion for an identity with no row, or a row that is revoked or expired |
| gateway: token (token method) | a missing, unknown, foreign, expired or revoked token; a settings file it cannot read (the gateway then refuses every call) |
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
`response_over_limit`. In the Access method the first names are these instead of the token ones: `missing_assertion`,
`assertion_malformed`, `assertion_unknown_key`, `assertion_bad_signature`, `assertion_expired`,
`assertion_wrong_audience`, `assertion_wrong_issuer`, `assertion_without_identity`, `access_keys_unavailable`,
`identity_not_listed`, `identity_expired`, `identity_revoked`. A cut stream has no refusal body: the judge's read ends
in an error after the bound.

## What is logged, and what never is

Logged: the time; the caller (the token id in the token method, `-` when the token is unknown; the identity and its
class in the Access method, `-` when the assertion was refused); the route; the model or the template name; the
status; the number of bytes sent back; the refusal code; for an upstream failure, a short reason such as `timeout`.

Never logged: the body of any request, the judge's code, the answer, any token or key, the `Authorization` header, the
Access assertion and its signature, a one-time PIN, a service token's secret, the parameter values of a template. A
model name or a template name is written only when it is on the gateway's own list, never as the caller typed it, so
a judge cannot write text into the log. The identity is the one thing taken from the caller, and it is taken from the
signed assertion, not from a header the caller writes.

## The token's life

Made by `judges-gateway token new --id <judge> --expires <time>` → sent to the judge privately → used until it expires
or is revoked → after the end of judging, the owner revokes what is left, or lets it expire. `token list` shows each
token's expiry and state, and never a token. The settings file holds hashes only and has mode `0600`. This is the
fallback method.

## The identity's life (Access method)

A judge's address goes in the `judges` setting of `infra/cloudflare` and in an identity row (`up.sh` adds both) →
Cloudflare mails the judge a PIN when they sign in → the row expires at the end of judging or is revoked with
`judges-gateway identity revoke --match <address>`. The runner's service token is made by `up.sh`, its secret goes to a
file the owner moves into the repository secrets, and it expires on its own date (`runner_token_duration`). Each
identity is cut off alone in two places, which are two locks: the gateway's row (takes effect on the next call) and
Cloudflare's policy or service token (takes effect when `up.sh apply` or `down.sh runner` runs). One lock is enough to
stop a call; both should be turned.

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
- **The gateway depends on Cloudflare's signing keys.** It fetches them from the team's public endpoint. While no key
  set has ever been read it refuses every call (`503 access_keys_unavailable`); once it has one, it keeps using it for
  as long as refreshes fail. A key Cloudflare withdraws is trusted here until the next successful refresh (at most ten
  minutes, or a minute after an unknown key id).
- **A one-time PIN proves control of a mailbox, nothing more.** A judge whose mail account is taken over is taken over.
  A domain row (`*@jury.example`) admits every address of that domain, including ones created later; use exact
  addresses when that is too wide.
- **The runner's secret sits in a repository secret.** Anyone who can write a workflow in that repository can make it
  send the secret to the gateway, and a workflow can print it to a log it controls. The runner's limits, expiry and
  its own revocation bound the damage; they do not remove it. The secret is never in the gateway's log.
- **If Access is misconfigured, the assertion is the only layer.** A request that did not pass Access has no valid
  assertion, because only Cloudflare holds the signing key, so the gateway refuses it. A tunnel set up without an
  Access application sends no assertion at all and every call is refused (`missing_assertion`).
- **A stolen token works until it expires or is revoked**, from wherever Access lets the thief in. Revoke it alone.
  The same holds for a stolen service token, until its expiry or `down.sh runner`.
- **Cloudflare sees the traffic.** The tunnel ends at Cloudflare's edge before it reaches us.
- **The host is ours.** The gateway's environment holds both credentials; whoever can read that process can read them.
- **The evidence an answer returns is whatever the templates select.** The templates read public advisory data; the
  gateway does not filter the rows further.

## Consequences

- A judge can run a fully live review with nothing of their own but the token the owner sends.
- The owner can end one judge's access in one command, and all of them by letting the date pass.
- Settings: `JUDGES_GATEWAY_AUTH` (`access`, the default, or `token`), `JUDGES_GATEWAY_ACCESS_TEAM`,
  `JUDGES_GATEWAY_ACCESS_AUD`, `JUDGES_GATEWAY_TOKENS_FILE` (the file also holds the identity rows),
  `JUDGES_GATEWAY_PORT`, `JUDGES_GATEWAY_LOG`, `JUDGES_GATEWAY_MODELS`, `JUDGES_GATEWAY_UPSTREAM_TIMEOUT`,
  `JUDGES_GATEWAY_MAX_RESPONSE_BYTES`, and the credential and address variables Proven already reads. None is
  hardcoded (ADR-0013); the defaults are in the runbook. The default is now `access`: a gateway started with no
  Access setting refuses to start and names the one that is missing.
- The public GitHub workflow stays on GitHub-hosted runners. It reaches the gateway with the runner's service token
  (headers `CF-Access-Client-Id` and `CF-Access-Client-Secret` from repository secrets), under the runner class's
  own limits, the same read-only templates and the same call log. No self-hosted runner is involved.
- What has been run: the verifier and the identity rows against signed test assertions, and the scripts against a stub of
  the Cloudflare API. Not run: any call to Cloudflare. The master runs `up.sh` live once.
- The lab side (the lab's `.env` and the Gitea Actions secrets) and the lab's "Go live with our services" section are the
  next round and are not part of this decision.
