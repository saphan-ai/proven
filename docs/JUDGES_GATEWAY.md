# The judges' gateway: the owner's runbook

This page is for the owner. It says how to run the gateway that lets a judge's lab reach our model service
(Nightshift) and our evidence store (Zetaris) with the judge's own token, how to put a Cloudflare tunnel and Cloudflare
Access in front of it, how to issue and revoke a token, and what to do when a judge reports a refusal. The decision and
its reasons are in [ADR-0051](adr/ADR-0051-a-judge-reaches-our-services-through-a-gateway-with-their-own-token.md).

Where each step is run:

- **On our host:** steps 1 to 4, and every token command.
- **In the Cloudflare dashboard:** the Access application and the service tokens (step 5).

What was run when this page was written: the gateway's tests, against fake upstreams. **No step on this page that touches
Cloudflare has been run by the author**; the `cloudflared` commands and the dashboard names are written from
Cloudflare's own documentation and are marked UNVERIFIED until the owner has run them once.

## 1. Build the gateway

```
go build -o judges-gateway ./cmd/judges-gateway
```

Put the binary where the unit in step 3 expects it, for example `~/bin/judges-gateway`.

## 2. Its settings

The gateway reads its settings from the environment. Put them in one file that only you can read
(`chmod 600`), never in the repository.

| variable | default | meaning |
|---|---|---|
| `JUDGES_GATEWAY_TOKENS_FILE` | none; required | The token settings file. The `token` commands write it; the gateway reads it again whenever it changes |
| `JUDGES_GATEWAY_PORT` | `8089` | The port. The gateway always listens on `127.0.0.1` and on no other address |
| `JUDGES_GATEWAY_LOG` | none: the log goes to standard output | A file the call log is appended to (mode `0600`) |
| `JUDGES_GATEWAY_MODELS` | `meterless/auto` | The models a judge may name, comma-separated. Any other model is refused |
| `JUDGES_GATEWAY_UPSTREAM_TIMEOUT` | `120s` | The bound on one model call |
| `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` | `4194304` (4 MiB) | The most bytes the gateway relays from the model service in one call. An integer from `1024` to `67108864` |
| `METERLESS_RUNTIME_TOKEN` | none; required | Our Nightshift runtime token. The gateway puts it on each model call. Secret: it is never logged or sent to a judge |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` | Nightshift's address (an origin; a path is empty or `/v1`) |
| `PROVEN_GENERATOR_APP` | `proven` | The application name sent to Nightshift |
| `ZETARIS_REST_URL`, `ZETARIS_API_KEY`, `ZETARIS_ORG_ID` | none; required | The same three settings Proven reads for live evidence ([CONFIGURATION.md](CONFIGURATION.md)). The key is secret. `ZETARIS_ROW_LIMIT` and `ZETARIS_TIMEOUT` apply too |

A wrong value is refused by name when the gateway starts, and the refusal never prints a secret.

## 3. Run it

A systemd user unit keeps it running without root. Write `~/.config/systemd/user/judges-gateway.service`:

```
[Unit]
Description=Proven judges' gateway
After=network-online.target

[Service]
EnvironmentFile=%h/.config/judges-gateway/env
ExecStart=%h/bin/judges-gateway serve
Restart=on-failure
RestartSec=2

[Install]
WantedBy=default.target
```

Then:

```
systemctl --user daemon-reload
systemctl --user enable --now judges-gateway
```

The call log is in the journal (`journalctl --user -u judges-gateway`), or in the file `JUDGES_GATEWAY_LOG` names.
Alarms are on standard error and start with `JUDGES-GATEWAY-ALARM`:

```
journalctl --user -u judges-gateway -o cat | grep '^JUDGES-GATEWAY-ALARM'
```

### If the log or the alarm cannot be written

The gateway will not serve calls it cannot log. If the disk is full or the log file is gone, the first write that fails
puts one line on standard error that starts `JUDGES-GATEWAY-FAULT` and names the sink (`sink=log` or `sink=alarm`).
From then on every call gets `503 log_unavailable`, and the health probe says the same. Watch for it:

```
journalctl --user -u judges-gateway -o cat | grep '^JUDGES-GATEWAY-FAULT'
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8089/healthz
```

The probe prints `200` when the gateway is healthy and `503` when a sink is down. It needs no token and answers nothing
else. To mend it: free the disk or restore the log path (for the file, check `JUDGES_GATEWAY_LOG`). The gateway needs
no restart: on the next call, or the next probe, it writes the held lines, prints one `JUDGES-GATEWAY-RECOVERED` line
with the number of lines lost, and serves again. Standard error is the second channel: if it is not read, only the probe tells you.

### If an answer is cut

A model answer larger than `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` is never relayed whole. If the model service declares its
size, the judge gets `502 response_over_limit` and no model byte. If not, the judge receives exactly the bound and the
connection is then aborted, so the judge's tool reports a broken stream. The call-log line has `status` 200, `bytes` =
what was relayed, `"refused":"response_over_limit"` and a `reason`. Find them:

```
grep '"refused":"response_over_limit"' <the call log>
```

An ordinary review answer is far smaller than 4 MiB. If a legitimate one is cut, raise the setting; if answers keep
coming near the bound, read the model service's side, because something is wrong there.

Check that it is up, from the host: a call with no token must be refused by name.

```
curl -s -X POST http://127.0.0.1:8089/v1/chat/completions
```

The answer is `{"refused":"missing_token", ...}`. If the connection is refused, read the journal for the setting it
named.

## 4. Issue a token to each judge

One token per judge. The token's id is the judge's name in lower case.

```
export JUDGES_GATEWAY_TOKENS_FILE=$HOME/.config/judges-gateway/tokens.json
judges-gateway token new --id judge-ada --expires 2026-11-05T00:00:00Z --per-minute 20 --per-day 500
```

`--expires` is required: set it to the end of judging. `--per-minute` and `--per-day` default to 20 and 500; they are
stored in the file per token, and you may edit them there. The command prints the token **once**:

```
id: judge-ada
expires: 2026-11-05T00:00:00Z
token: pjt_...
```

Only its hash is stored; if you lose it, revoke the id and make a new one.

**Send the token to the judge privately** (a direct message or an email to that judge alone). The tunnel's host name is
public and may go in the lab's page; the token never goes in a public place, a commit, an issue or a chat room.

List what you have issued, with each token's expiry and state (it shows no token):

```
judges-gateway token list
```

## 5. The Cloudflare tunnel and Access

**UNVERIFIED: written from Cloudflare's documentation, not run by the author.** Where Cloudflare's current screens differ,
Cloudflare's wording wins; the order of the steps does not change.

### 5a. The tunnel (on our host)

```
cloudflared tunnel login
cloudflared tunnel create proven-judges
cloudflared tunnel route dns proven-judges <the public host name, for example judges.example.org>
```

Write `~/.cloudflared/config.yml`:

```
tunnel: proven-judges
credentials-file: <the path `tunnel create` printed>
ingress:
  - hostname: <the public host name>
    service: http://127.0.0.1:8089
  - service: http_status:404
```

The `service` line is the gateway's own address. If you changed `JUDGES_GATEWAY_PORT`, change the port here.
Run it (or install it as a service with `cloudflared service install`):

```
cloudflared tunnel run proven-judges
```

### 5b. The Access application (in the dashboard)

Do this **before** you tell anyone the host name. Without it the tunnel is open to the world and only the token stands.

1. Cloudflare dashboard → **Zero Trust** → **Access** → **Applications** → **Add an application** → **Self-hosted**.
2. Application name: `proven-judges`. Public hostname: the host name from 5a.
3. Add a policy: name `judges`, action **Allow**. For a judge who will use the lab from a script, use a service
   token (below). For a judge who will open the address in a browser, add **Include → Emails** with that judge's email.
   Do not add **Everyone**.
4. For a service token: **Zero Trust → Access → Service Auth → Create Service Token**, one per judge, named for the judge,
   with a duration that ends with judging. Copy the Client ID and Client Secret; the secret is shown once. In the policy,
   add a second rule: action **Service Auth**, **Include → Service Token** → that token.
5. Save the application.

Test from outside (your phone's network, not our host): a call with no Access credential must be turned away by
Cloudflare, not answered by the gateway:

```
curl -s -o /dev/null -w '%{http_code}\n' -X POST https://<the public host name>/v1/chat/completions
```

Expect a redirect or a 403 from Cloudflare. If the answer is the gateway's JSON, Access is not in front of it: stop
and fix that first.

A judge with a service token sends two headers, `CF-Access-Client-Id` and `CF-Access-Client-Secret`, on every call, and
the Proven token as `Authorization: Bearer`. **Proven's generator sends no Cloudflare header today.** A lab's
connection to an Access-protected host therefore needs a small helper on the lab side that adds them (Cloudflare's
`cloudflared access` client does this); that helper and the lab's `.env` are the next round
([ADR-0051](adr/ADR-0051-a-judge-reaches-our-services-through-a-gateway-with-their-own-token.md), Consequences).

## 6. Revoke a token

One judge:

```
judges-gateway token revoke --id judge-ada
```

It takes effect on the judge's next call. The other tokens keep working. Revoke the judge's Cloudflare service token
or remove their email from the Access policy as well: the gateway token and the Access entry are two locks.

When judging ends, run `judges-gateway token list`. Every token past its expiry already refuses by name. Revoke any
that is still active, then delete the Access application or its policy, then stop the tunnel:

```
systemctl --user stop judges-gateway
```

## 7. When a judge reports a refusal

Ask the judge for the `refused` name in the answer. It says which layer spoke.

| `refused` | meaning | what you do |
|---|---|---|
| `missing_token` | no `Authorization: Bearer` header | The judge's lab is not sending the token; point them to the lab page |
| `unknown_token` | the token is not one this gateway issued | Check for a typo or a token from another gateway; if lost, revoke the id and issue a new one |
| `token_expired` | the token's expiry has passed | Issue a new token only if the owner agrees to extend that judge |
| `token_revoked` | you revoked it | Nothing, unless it was a mistake: issue a new token |
| `rate_limit_minute` | too many calls in one minute | Tell the judge to wait for the next minute; raise `per_minute` in the file only if the judge's review is legitimate |
| `allowance_day` | the day's allowance is used | It renews at 00:00 UTC; raise `per_day` in the file if you agree |
| `model_not_allowed` | the judge asked for a model not on the list | Tell them the model the answer names; add it to `JUDGES_GATEWAY_MODELS` only if you want to serve it |
| `not_a_template_request`, `unknown_template`, `missing_parameter`, `unknown_parameter`, `bad_parameter` | the evidence call is not a valid query template request | The lab is not Proven, or it is a version that builds queries another way; Proven's own calls never get this |
| `body_too_large` | the request is over the size bound | The change is too large for one call; ask the judge to review a smaller one |
| `upstream_unavailable`, `upstream_error` | Nightshift or Zetaris did not answer or refused | Not the judge's fault. Read the log line's `reason`, then the service's own state |
| `gateway_settings_unreadable` | the token file cannot be read or fails its rules | Run `judges-gateway token list`: it says what is wrong. Until the file is right, every call is refused |
| `log_unavailable` | the gateway cannot write its call log or its alarm, so it serves nothing | Not the judge's fault. Follow "If the log or the alarm cannot be written"; the judge may retry once you see `200` from `/healthz` |
| `response_over_limit` | the model answer was larger than `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` (on a stream that was cut, the judge sees a broken stream, not this name) | Look for it in the log; raise the setting only if the answer was legitimate |
| `no_such_route`, `method_not_allowed` | wrong path or method | The gateway serves `POST /v1/chat/completions` and `POST /v1/evidence/query` only |

To follow one judge's calls, search the log for their id: `grep '"token":"judge-ada"'`. The log holds no request body and
no credential, so it is safe to paste into a message to the judge.
