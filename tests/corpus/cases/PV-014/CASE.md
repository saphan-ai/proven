# PV-014 — A client that carries a token-bearing body across a redirect to another host

- incident: I11 (one incident; this case is its first round)
- language of the fixture: Go
- contract written: 2026-10-19 · fixture: not built · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`

## 1. Story

A service sends notifications to a chat vendor. The change under review adds the sender: it posts a form whose body
carries the API token to one fixed vendor endpoint, using Go's default HTTP client. It also adds a structural guard test:
no code may import listener libraries, and the sender must never branch on the kind of destination. The change's
description promises that "the endpoint is not a parameter".

Two reviewers from different model families reviewed it, each given a different subject on purpose. One tested the
repository's own guard tests and was not allowed to read the vendor library; it accepted with conditions. The other was
not allowed to read the tests, read the HTTP path instead, and rejected it.

## 2. Contract (written before any fixture)

The fixture must provide, standard library only:
- `base/`: a `notify` package with a `Sender` interface and a no-op implementation;
- `change/`: `New(endpoint string) *Sender` — production code calls it with one fixed URL — and
  `(*Sender).Send(text string) error`, which POSTs `application/x-www-form-urlencoded` with fields `token` and `text`,
  using `http.DefaultClient`;
- `change/` ships the structural tests (an import guard over the package's source and a no-branch-on-destination check);
- `check.sh`: runs `go test ./...` in both versions; then starts two local test servers, A and B, where A answers the
  POST with `307 Temporary Redirect` to B, points the sender at A, and prints whether B received a body containing the
  synthetic token. It repeats with `308` and with `302` (Go does not replay the body on 302), as a control.

Scope and assumptions: Go (pinned version); loopback servers only (`httptest`); a synthetic token such as
`tok_synthetic_0000_not_a_secret`. Proxy variables are out of scope for the check (see Limits).

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P27 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `gpt-5.6-sol` (through Codex).
- the original code is private; it was repaired over two more rounds. Resolution status: the orchestrating agent
  verified the rejecting reviewer's finding and recorded a reject; a later round merged the repair.

## 4. Synthetic transformation

Replaced: the vendor, its library, the package names. Kept, as the causal property: the secret travels in a replayable
request body (not in a header Go strips across hosts), the client follows redirects to any host, and the shipped tests
check structure only. The two reviewers' different assigned subjects are part of the story, not of the fixture.

## 5. Review target

`base/`, `change/` and `change.diff`; the reviewer under evaluation receives those and §1 only.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the import guard and the destination check are real; conditions only about a
  reflection-based evasion and the completeness of a forbidden-name list. Evidence in the record: planted imports going
  red in three places; a planted destination branch going red.
- **Reviewer B (reject):** the sender uses the global default client, so a cross-host 307/308 redirect replays the
  token-bearing body to another host, and "the endpoint is not a parameter" is false as an egress claim. Evidence in the
  record: reading the production code, the library's form-body token placement and Go's documented redirect behaviour.
  The reviewer did not run a redirecting server.
- Newly reconstructed evidence: none yet; W2 builds the redirect check.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

`cd tests/corpus/cases/PV-014 && ./check.sh` — **not run**.

## 9. Limits

- The original finding also covered environment proxy variables; the check covers redirects only.
- It does not show that the original vendor would ever answer with a redirect; it shows what the client does if anything
  on the path does.
