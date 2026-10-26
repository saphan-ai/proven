# The judges' gateway: the owner's runbook

This page is for the owner. It says how to run the gateway that lets a judge's lab, and the public GitHub workflow,
reach our model service (Nightshift) and our evidence store (Zetaris); how Cloudflare Access is put in front of it by
scripts; how a judge signs in; how the workflow authenticates; how to cut one identity off; and what to do when a
refusal is reported. The decision and its reasons are in
[ADR-0051](adr/ADR-0051-a-judge-reaches-our-services-through-a-gateway-with-their-own-token.md).

There are two ways in, chosen by `JUDGES_GATEWAY_AUTH`:

- **`access` (the default).** Cloudflare Access is in front of the tunnel. A judge signs in with an e-mail one-time
  PIN; the GitHub workflow presents a service token. Cloudflare signs a JWT and the gateway verifies it and looks the
  identity up in its settings file. No secret of ours is sent to a judge.
- **`token` (the fallback).** A judge sends their own expiring `pjt_` token. See step 4b.

Where each step is run:

- **On our host:** steps 1 to 4, the gateway, and the `identity` and `token` commands.
- **By an agent or the owner, with one Cloudflare API token:** step 5 (`infra/cloudflare/`). Nothing is done by hand in
  the Cloudflare dashboard.

What was run when this page was written: the gateway's tests against fake upstreams and signed test assertions, and
the scripts against a **stub** of the Cloudflare API. **No call to Cloudflare has been made by the author.** The
endpoint paths, the token's permission names and the `cloudflared` flags below are written from Cloudflare's documentation and are marked UNVERIFIED
until the scripts have been run once live (step 5c).

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
| `JUDGES_GATEWAY_AUTH` | `access` | `access` or `token`. In `access` mode the two Access settings below are required and a judge token is not a way in; in `token` mode an Access header is ignored |
| `JUDGES_GATEWAY_ACCESS_TEAM` | none; required in `access` mode | The team domain, `<team>.cloudflareaccess.com`. `up.sh` writes it |
| `JUDGES_GATEWAY_ACCESS_AUD` | none; required in `access` mode | The Access application's AUD tag, 64 hex digits. `up.sh` writes it |
| `JUDGES_GATEWAY_TOKENS_FILE` | none; required | The settings file. It holds the identity rows (`identity` commands) and the token rows (`token` commands); the gateway reads it again whenever it changes |
| `JUDGES_GATEWAY_PORT` | `8089` | The port. The gateway always listens on `127.0.0.1` and on no other address |
| `JUDGES_GATEWAY_LOG` | none: the log goes to standard output | A file the call log is appended to (mode `0600`) |
| `JUDGES_GATEWAY_MODELS` | `meterless/auto` | The models a caller may name, comma-separated. Any other model is refused |
| `JUDGES_GATEWAY_UPSTREAM_TIMEOUT` | `120s` | The bound on one model call |
| `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` | `4194304` (4 MiB) | The most bytes the gateway relays from the model service in one call. An integer from `1024` to `67108864` |
| `METERLESS_RUNTIME_TOKEN` | none; required | Our Nightshift runtime token. The gateway puts it on each model call. Secret: it is never logged or sent to a caller |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` | Nightshift's address (an origin; a path is empty or `/v1`) |
| `PROVEN_GENERATOR_APP` | `proven` | The application name sent to Nightshift |
| `ZETARIS_REST_URL`, `ZETARIS_API_KEY`, `ZETARIS_ORG_ID` | none; required | The same three settings Proven reads for live evidence ([CONFIGURATION.md](CONFIGURATION.md)). The key is secret. `ZETARIS_ROW_LIMIT` and `ZETARIS_TIMEOUT` apply too |

A wrong value is refused by name when the gateway starts, and the refusal never prints a secret. A gateway started in
`access` mode with no team or no AUD tag names the missing one and does not start.

## 3. Run it

A systemd user unit keeps it running without root. Write `~/.config/systemd/user/judges-gateway.service`:

```
[Unit]
Description=Proven judges' gateway
After=network-online.target

[Service]
EnvironmentFile=%h/.config/judges-gateway/env
EnvironmentFile=%h/.config/judges-gateway/access.env
ExecStart=%h/bin/judges-gateway serve
Restart=on-failure
RestartSec=2

[Install]
WantedBy=default.target
```

`env` holds the secrets and addresses of the table above. `access.env` is the file `up.sh` writes (step 5): copy it, or
point `gateway_env_file` in the script settings at this path.

Then:

```
systemctl --user daemon-reload
systemctl --user enable --now judges-gateway
```

The call log is in the journal (`journalctl --user -u judges-gateway`), or in the file `JUDGES_GATEWAY_LOG` names.
Alarms are on standard error and start with `JUDGES-GATEWAY-ALARM`; in the Access method they read
`... identity=<address or client id> window=minute used=N limit=M`:

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
size, the caller gets `502 response_over_limit` and no model byte. If not, the caller receives exactly the bound and the
connection is then aborted, so the caller's tool reports a broken stream. The call-log line has `status` 200, `bytes` =
what was relayed, `"refused":"response_over_limit"` and a `reason`. Find them:

```
grep '"refused":"response_over_limit"' <the call log>
```

An ordinary review answer is far smaller than 4 MiB. If a legitimate one is cut, raise the setting; if answers keep
coming near the bound, read the model service's side, because something is wrong there.

Check that it is up, from the host. In `access` mode a call with no assertion must be refused by name:

```
curl -s -X POST http://127.0.0.1:8089/v1/chat/completions
```

The answer is `{"refused":"missing_assertion", ...}` (`missing_token` in `token` mode). If the connection is refused,
read the journal for the setting it named.

## 4. Who may come in

### 4a. Identities (the Access method)

An identity is a row in the settings file. Access lets a person or a token through its door; the gateway then asks
whether it has a row. A valid assertion for an identity with no row is refused (`identity_not_listed`).

```
export JUDGES_GATEWAY_TOKENS_FILE=$HOME/.config/judges-gateway/tokens.json
judges-gateway identity add --match ada@jury.example --expires 2026-11-05T00:00:00Z --per-minute 20 --per-day 500
judges-gateway identity add --match '*@jury.example'  --expires 2026-11-05T00:00:00Z
judges-gateway identity add --match abc123.access      --expires 2026-11-05T00:00:00Z
judges-gateway identity list
```

`--match` is one of three shapes. An e-mail address (a judge). `*@domain` (every address of exactly that domain, each
limited on its own; not its subdomains). A service token's client id (the GitHub runner). The class follows from the
shape: an address or a domain is `judge`, a client id is `runner`. `--expires` is required; set it to the end of judging.
The limits default to 20 a minute and 500 a day and are stored per row.

`up.sh` adds these rows for you when its settings name `gateway_bin`, `gateway_settings_file` and `identities_expire`
(step 5), so the Cloudflare allowlist and the gateway's list are made from the one `judges` setting.

### 4b. Tokens (the fallback method)

Set `JUDGES_GATEWAY_AUTH=token`. One token per judge. The token's id is the judge's name in lower case.

```
judges-gateway token new --id judge-ada --expires 2026-11-05T00:00:00Z --per-minute 20 --per-day 500
```

`--expires` is required. The command prints the token **once**:

```
id: judge-ada
expires: 2026-11-05T00:00:00Z
token: pjt_...
```

Only its hash is stored; if you lose it, revoke the id and make a new one. **Send the token to the judge privately** (a
direct message or an email to that judge alone). The host name may be public; the token never goes in a public place, a
commit, an issue or a chat room. `judges-gateway token list` shows each token's expiry and state, never a token.

## 5. The Cloudflare setup, by script

Everything is in `infra/cloudflare/`: `up.sh`, `down.sh`, `verify.sh`, a shared `lib.sh` and `settings.example`. They
need `sh`, `curl` and `jq`. They create the named tunnel, the DNS record, the Access application, the two policies
(the judges' e-mail allowlist and the runner's service-auth policy), the e-mail one-time-PIN identity provider if the
account has none, and the runner's service token.

### 5a. The one thing the owner provides: one scoped Cloudflare API token

Make an API token in the dashboard (**My Profile → API Tokens → Create Token → Custom token**) with these permissions
(names UNVERIFIED; Cloudflare's current screen wins), limited to our account and our zone:

- Account · **Access: Apps and Policies** · Edit
- Account · **Access: Service Tokens** · Edit
- Account · **Access: Organizations, Identity Providers, and Groups** · Edit
- Account · **Cloudflare Tunnel** · Edit
- Zone · **DNS** · Edit, and Zone · **Zone** · Read

Put it where the scripts can read it and nobody else can, **outside the repository**:

```
umask 077 && printf '%s\n' '<the token>' > ~/.config/judges-gateway/cloudflare-api-token
```

The scripts read it from `CLOUDFLARE_API_TOKEN`, or from the file the `token_file` setting names. They never print it,
never put it on a command line (curl reads it from its standard input), and never write it to a file. The domain must
already be in the Cloudflare account; the scripts do not create zones. If the account has no Zero Trust organisation
yet, set `team_name` and the scripts create it.

### 5b. Settings

```
cp infra/cloudflare/settings.example ~/.config/judges-gateway/cloudflare.settings
```

Edit it: `account_id`, `zone`, `hostname` (inside the zone), `judges` (e-mails and `*@domain`, comma-separated) and
`token_file`. The file is `key=value` lines; it is read, never run. The comments in the example say what each key does.

### 5c. Plan, up, verify

```
sh infra/cloudflare/up.sh plan  ~/.config/judges-gateway/cloudflare.settings
sh infra/cloudflare/up.sh apply ~/.config/judges-gateway/cloudflare.settings
sh infra/cloudflare/verify.sh   ~/.config/judges-gateway/cloudflare.settings
```

`plan` reads and prints. Every line is `ok:` (already as wanted) or `plan: would ...` (what `apply` would do); it makes no
write request. `apply` finds every object by name first and creates only what is missing, and updates only what
differs. **Run it twice: the second run prints only `ok:` lines and makes no write.** If an API call fails, the
script stops and prints the Cloudflare error code and message, then which call.

`apply` leaves four files, all mode `0600` and under the state directory (`state_dir`, default
`~/.local/state/proven-judges-gateway/`):

| file | what it is | what you do with it |
|---|---|---|
| `gateway-access.env` | `JUDGES_GATEWAY_AUTH`, `..._ACCESS_TEAM`, `..._ACCESS_AUD` | the gateway's `EnvironmentFile` (step 3); restart the gateway |
| `runner-service-token.env` | the runner's `CF_ACCESS_CLIENT_ID` and `CF_ACCESS_CLIENT_SECRET`, **shown once, here only** | put them into the repository secrets (step 5e), then delete the file |
| `tunnel.token` | the connector's run token | the next step |
| `state.env` | the ids `up.sh` created, which `down.sh` is allowed to delete | leave it |

`verify.sh` then looks from outside: a call with no credential must be turned away by Access (a redirect, 401 or 403),
and a call with the runner's service token must reach the gateway, which refuses an empty request by name
(`not_a_template_request`). One line says it works, one line says what is wrong.

### 5d. Run the connector

The tunnel's ingress is set by the script (the host to `127.0.0.1:<gateway_port>`, everything else 404), so the
connector needs only its token. UNVERIFIED: that `cloudflared` accepts the token from a file:

```
cloudflared tunnel run --token-file ~/.local/state/proven-judges-gateway/tunnel.token
```

Install it as a service the same way as the gateway (a user unit with that command as `ExecStart`).

### 5e. The GitHub workflow's credentials

In the repository: **Settings → Secrets and variables → Actions → New repository secret**, twice:
`CF_ACCESS_CLIENT_ID` and `CF_ACCESS_CLIENT_SECRET`, with the two values from `runner-service-token.env`. Then delete that
file. If the secret is lost or shown by mistake, `sh infra/cloudflare/up.sh rotate-runner <settings>` issues a new
one to the same file.

### 5f. Take it all away

```
sh infra/cloudflare/down.sh plan  ~/.config/judges-gateway/cloudflare.settings
sh infra/cloudflare/down.sh apply ~/.config/judges-gateway/cloudflare.settings
```

`down.sh` deletes, in reverse order, only objects that `up.sh` recorded as created (by id, in `state.env`) and whose name
still matches: the policies, the service token, the application, the DNS record (also only if its comment carries
the tag), the tunnel (its connections are cleared first), and the identity provider only if `up.sh` made it. What was
there before is left and named (`keep:`). If the state file is lost, `--by-name` accepts the names alone (it still
requires the tag on the DNS record and never deletes an identity provider). The gateway's own identity rows are not
touched: revoke them (step 6).

## 6. How a judge signs in, how the workflow authenticates, how to revoke

### A judge

A judge whose address is on the list opens the public host name in a browser: Cloudflare asks for the address, mails a
six-digit PIN, and the judge types it in. From a terminal or a script the judge runs, once:

```
cloudflared access login https://<the public host name>
```

which opens the same sign-in in a browser and stores a signed session for that host. UNVERIFIED: the exact output and
how a lab passes the session to a tool. `cloudflared access token -app=https://<host>` prints the session JWT; a call
that carries it as the header `Cf-Access-Jwt-Assertion`, or as a cookie, is a call Access has already passed. Proven's
generator sends no Cloudflare header today, so the lab's live flow ("ours") opens this sign-in first; that lab helper is
W17's round and is not part of this wagon.

### The GitHub workflow

The workflow stays on GitHub-hosted runners. It sends two headers on every call, from the two repository secrets:

```
curl -s -X POST https://<the public host name>/v1/evidence/query \
  -H "CF-Access-Client-Id: ${{ secrets.CF_ACCESS_CLIENT_ID }}" \
  -H "CF-Access-Client-Secret: ${{ secrets.CF_ACCESS_CLIENT_SECRET }}" \
  -H 'Content-Type: application/json' \
  -d '{"template":"kev_lookup","params":{"cve_id":"CVE-2021-44228"}}'
```

Cloudflare checks the two values against the service token and adds the assertion; the gateway sees the token's client
id as the identity and applies the **runner** row's limits. The runner reads the same read-only templates and shows up
in the same call log, as `"identity":"<client id>","class":"runner"`.

### Revoke one identity

Each identity has two locks. Turn both.

```
judges-gateway identity revoke --match ada@jury.example     # the gateway: the next call is refused (identity_revoked)
```

then remove the address from `judges` in the script settings and run `up.sh apply` (Cloudflare stops sending the PIN).
The other identities keep working. To cut one address out of a domain row, add an exact row for it and revoke that
(`identity add --match eve@jury.example ...` then `identity revoke --match eve@jury.example`): an exact row beats the
domain row.

The runner:

```
judges-gateway identity revoke --match <the client id>
sh infra/cloudflare/down.sh runner ~/.config/judges-gateway/cloudflare.settings
```

The second command deletes the service token and the runner's policy only; the judges are untouched.

When judging ends, run `judges-gateway identity list`. Every row past its expiry already refuses by name. Revoke any
that is still active, run `down.sh apply`, then stop the gateway:

```
systemctl --user stop judges-gateway
```

(Tokens of the fallback method: `judges-gateway token revoke --id judge-ada`, and `token list`.)

## 7. When a refusal is reported

Ask for the `refused` name in the answer. It says which layer spoke. A redirect to a Cloudflare sign-in page, or a
Cloudflare 403 page, is Access speaking before the gateway: the address is not on the `judges` list, or the service
token is wrong or expired. Look in Cloudflare's Access log.

| `refused` | meaning | what you do |
|---|---|---|
| `missing_assertion` | the call reached the gateway with no Access assertion | The call did not come through Access: the tunnel may be set up without the application, or the caller is on the host. Run `verify.sh` |
| `assertion_malformed`, `assertion_unknown_key`, `assertion_bad_signature` | the header is not a JWT signed with a key the team publishes | Not a call Cloudflare made. Do nothing for the caller; if it recurs, someone is guessing |
| `assertion_expired` | the session is older than its life | The judge signs in again (`cloudflared access login`) |
| `assertion_wrong_audience`, `assertion_wrong_issuer` | the JWT is for another application or team | Check `JUDGES_GATEWAY_ACCESS_AUD` and `_TEAM` against `gateway-access.env`; if you re-ran `up.sh` after deleting the application, the AUD tag changed |
| `assertion_without_identity` | the JWT names neither an e-mail nor a service token | Not a policy `up.sh` makes; look at the Access application's policies |
| `access_keys_unavailable` | the gateway has never been able to read Cloudflare's signing keys | Check that the host can reach `https://<team>.cloudflareaccess.com/cdn-cgi/access/certs`; the gateway tries again after a minute |
| `identity_not_listed` | Access let the caller in; the gateway has no row for them | Add a row (step 4a), or leave it if the caller should not be in |
| `identity_expired`, `identity_revoked` | the row's expiry has passed, or you revoked it | Extend only if the owner agrees: edit that row's `expires` in the settings file (and delete its `revoked_at` if the revocation was a mistake); the gateway reads the change on the next call |
| `missing_token`, `unknown_token`, `token_expired`, `token_revoked` | (token method) the judge token is missing, not ours, expired or revoked | Step 4b; issue a new token only if the owner agrees |
| `rate_limit_minute` | too many calls in one minute | Tell the caller to wait for the next minute; raise `per_minute` in the file only if the use is legitimate |
| `allowance_day` | the day's allowance is used | It renews at 00:00 UTC; raise `per_day` in the file if you agree |
| `model_not_allowed` | the caller asked for a model not on the list | Tell them the model the answer names; add it to `JUDGES_GATEWAY_MODELS` only if you want to serve it |
| `not_a_template_request`, `unknown_template`, `missing_parameter`, `unknown_parameter`, `bad_parameter` | the evidence call is not a valid query template request | The caller is not Proven, or is a version that builds queries another way; Proven's own calls never get this |
| `body_too_large` | the request is over the size bound | The change is too large for one call; ask for a smaller one |
| `upstream_unavailable`, `upstream_error` | Nightshift or Zetaris did not answer or refused | Not the caller's fault. Read the log line's `reason`, then the service's own state |
| `gateway_settings_unreadable` | the settings file cannot be read or fails its rules | Run `judges-gateway identity list`: it says what is wrong. Until the file is right, every call is refused |
| `log_unavailable` | the gateway cannot write its call log or its alarm, so it serves nothing | Not the caller's fault. Follow "If the log or the alarm cannot be written"; they may retry once you see `200` from `/healthz` |
| `response_over_limit` | the model answer was larger than `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` (on a stream that was cut, the caller sees a broken stream, not this name) | Look for it in the log; raise the setting only if the answer was legitimate |
| `no_such_route`, `method_not_allowed` | wrong path or method | The gateway serves `POST /v1/chat/completions` and `POST /v1/evidence/query` only |

To follow one caller's calls, search the log for their identity: `grep '"identity":"ada@jury.example"'` (token method:
`grep '"token":"judge-ada"'`). The log holds no request body, no assertion and no credential, so it is safe to paste into
a message to the judge. What Access logs (who signed in, from where) is in Cloudflare's own Access log, not here.
