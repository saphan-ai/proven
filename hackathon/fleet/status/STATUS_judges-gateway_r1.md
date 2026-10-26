# STATUS — judges-gateway · round 1

- state: ready-for-review
- executor: claude-code
- model: `claude-sonnet-5-5`
- last commit: the tip of branch `fix/judges-gateway` (`[judges-gateway] docs: ADR-0051 and the owner's gateway runbook`)
- updated: 2026-10-26, UTC

## Summary
ADR-0051, a gateway (`cmd/judges-gateway`, `internal/gateway`, standard library only) and the owner's runbook. A judge's
lab reaches our model service and our evidence store with the judge's own expiring token. Evidence goes through Proven's
query templates only. Every call is limited and logged. The gateway tests pass against fake upstreams. The Cloudflare
steps are not run.

## Acceptance, item by item
1. **ADR-0051** — [`docs/adr/ADR-0051-…`](../../../docs/adr/ADR-0051-a-judge-reaches-our-services-through-a-gateway-with-their-own-token.md):
   the decision, the boundary, what each layer refuses, what is logged and not, the token's life, what a judge sees, and
   what is not protected. Added to the ADR index.
2. **The gateway** — `go test -race -count=1 ./internal/gateway ./cmd/judges-gateway` → ok, both packages.
   Two routes: `POST /v1/chat/completions` (model) and `POST /v1/evidence/query` (templates). Tokens: a settings file of
   id, SHA-256, expiry and limits; `judges-gateway token new|revoke|list`; the token is printed once. Listens on
   `127.0.0.1` only (a test asserts the address is loopback).
3. **The runbook** — [`docs/JUDGES_GATEWAY.md`](../../../docs/JUDGES_GATEWAY.md): a systemd user unit, the tunnel, the Access
   application, issue, send, revoke, and a table of every refusal with what to do.

## Controls (each names the field it mutates)
- Expired token refused by name: the clock moves past `expires`. Mutation check: with the expiry test disabled in the
  code, 3 tests failed; restored.
- Unknown, foreign and missing token refused: no upstream call.
- Revoking one token leaves another working: `revoked_at` of one record.
- Raw SQL, SQL as a template name, an extra field, an unknown template, a missing, extra or bad parameter, two JSON
  documents, an array, an empty object: each refused by name; the fake Zetaris counted 0 calls. Mutation check: with the
  strict decoder replaced by a lax one, the test failed; restored.
- The upstream receives our credential and never the judge's, in header and body, for both services; the judge's answer
  holds neither of ours. Mutation check: with the judge's header copied upstream, the test failed; restored.
- The log has one JSON line per call and none of: the judge token, our two credentials, the chat text, the SQL text, a
  parameter value. Zero hits.
- A minute limit of 5 refuses the 6th call and the 80% alarm line appears once; a day limit of 3 refuses the 4th.
- An unreadable settings file refuses every call by name until it is repaired.
- `go vet ./...` → exit 0.

## Not run
- **`go test ./...` is not green.** `internal/cli/sarifschema` fails: the review cannot create its log directory in the
  sandbox's home (`mkdir … permission denied`). The cause is the environment, not this change; the test was not run
  anywhere it can pass. Every other package passes.
- **`scripts/check-links.sh` exits 1.** Two links in the generated `docs-site/generated/settings.md` point at
  `OSV_TABLE.md` and `POLICY.md` and do not resolve from that directory. The file is untouched by this change and the
  two links come from the settings descriptions in `internal/osv` and the policy loader. Every link added here resolves.
- Nothing on Cloudflare: the `cloudflared` commands and the dashboard steps are written from Cloudflare's documentation
  and not run.
- No test against a real Nightshift or a real Zetaris.

## Blocked on / next
- **Lab side (next round):** Proven's generator sends no Cloudflare Access header, so a lab cannot reach an
  Access-protected host with a service token until a helper on the lab side adds them. Proven also has no query
  adapter that calls the evidence route; today it calls Zetaris's own address. Both belong to the lab's round.
- The two broken links and the sandbox failure are for their owners.
