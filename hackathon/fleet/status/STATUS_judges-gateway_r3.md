# STATUS — judges-gateway · round 3

- state: ready-for-review
- executor: claude-code
- model: `claude-sonnet-5-5`
- last commit: the tip of branch `fix/judges-gateway` (`[judges-gateway] docs: ADR-0051 and runbook for Cloudflare Access`)
- updated: 2026-10-26, UTC

## Summary
Round 3 puts Cloudflare Access in front of the gateway, as code. The gateway now trusts only a valid Access assertion
(a judge by e-mail one-time PIN, the GitHub runner by a service token). Each identity is limited, logged and revoked
alone. Idempotent scripts create and remove the tunnel, the DNS record, the Access application, both policies and the
runner's service token. The round-1 per-judge token stays as the fallback (`JUDGES_GATEWAY_AUTH=token`).

## What changed
1. **The gateway verifies the assertion itself** (`internal/gateway/access.go`, standard library only). RS256 JWT from
   `Cf-Access-Jwt-Assertion`; keys fetched from the team's certs endpoint, kept ten minutes, at most one extra fetch a
   minute for an unknown key id, keys under 2048 bits refused; signature first, then issuer, audience (the application's
   AUD tag), expiry and `nbf`. Identity is the `email` claim or the `common_name` claim. Each failure has its own name.
2. **Identity rows** in the settings file (`internal/gateway/identity.go`): an e-mail, `*@domain` or a client id; class
   `judge` or `runner`; required expiry; per-minute and per-day limits; revoked flag. An identity with a valid assertion
   and no row is refused (`identity_not_listed`). An exact row beats a domain row. A domain row admits only that domain,
   and counts each address on its own. New commands: `judges-gateway identity add|revoke|list`.
3. **`infra/cloudflare/`**: `up.sh plan|apply|rotate-runner`, `down.sh plan|apply|runner`, `verify.sh`, `lib.sh`,
   `settings.example`. POSIX `sh`, `curl`, `jq`. Objects are found by name first and only the missing are created.
   `down.sh` deletes only what `up.sh` recorded by id (and the DNS record only with its tag). The API token reaches
   `curl` on standard input; the runner's secret goes once to a mode `0600` file.
4. **ADR-0051 amended in place** (decisions 2, 3, 8, 11 to 14; boundary, refusal names, log, lifecycle, limits of
   protection, consequences) and **`docs/JUDGES_GATEWAY.md` rewritten** for the agent-run setup, a judge's sign-in, the
   workflow's headers and revocation per identity.

## Controls (each names the field it mutates)
- A JWT signed by an unknown key, a known key id with another key, `alg` none and HS256, an expired one, a wrong
  audience, a wrong issuer, no e-mail and no client id, tampered claims: each refused by its own name. Mutation check:
  audience check disabled → `TestAssertionIsVerifiedOrRefusedByName` (`wrong_audience`, `no_audience`) and
  `TestAccessRefusalsAreNamedAndNothingGoesUpstream` failed; restored.
- A valid assertion for an unlisted e-mail → 403 `identity_not_listed`, no upstream call; a listed one is served; a
  domain row admits only that domain (a subdomain and a longer name are refused); an exact revoked row cuts one address
  of a domain; a revoked runner is refused while a judge keeps working (mutated field: the runner row's `revoked_at`).
- A judge token is not a way in in `access` mode. The log names `identity` and `class` and holds no assertion, no
  signature, no credential, no body.
- The Cloudflare API token and the runner's secret appear in no script output, no file under the state directory, and
  no command line in 636 process-list samples of the stub run. The sampler saw `curl`, so its zero is not blind.
  Mutation check: `curl -K - -H "X-Mutant: $CF_TOKEN"` → `TestTheTokenAndTheSecretAreNeverShown` failed; restored.
- `up.sh apply` twice against the stub: the second run makes no write request and the stub's state is byte-identical.
- `up.sh plan` makes no write request. Mutation check: `act` returns "go" in plan mode → `TestPlanMakesNoWriteRequest`
  failed; restored.
- `down.sh` removes only what `up.sh` made: a pre-existing identity provider, tunnel, application and DNS record stay;
  an object with our name that is not in the state stays; without a state file it refuses. Mutation check: the
  ownership test always true → `TestDownLeavesAnObjectWithOurNameThatUpDidNotMake` failed; restored.
- Every API error prints as `Cloudflare error <code>: <message>` and the call; a wrong token is named by its code and
  never echoed.
- `go vet ./...` clean; `gofmt -l cmd internal infra` empty; `sh -n` on all four scripts; the gateway and script
  tests pass with `-race` (`-count=3` for the gateway).

## Not run
- Any call to Cloudflare. The endpoint paths, the policy decision names, the token's permission names, `cloudflared
  tunnel run --token-file` and `cloudflared access login` are written from documentation and are UNVERIFIED. The
  master runs `up.sh` live once.
- `go test ./...` fails in `internal/cli/sarifschema`: it opens a file under the home cache, which this sandbox
  denies. Not caused by this round.
- `scripts/check-links.sh` exits 1 for the same two links in `docs-site/generated/settings.md`; no link added here is
  broken.
- Key rotation against the real Cloudflare certs endpoint; a stub serves the key set.
- Lists longer than 100 objects in an account (the scripts ask for 100 and read no further page).
