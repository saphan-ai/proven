# PV-014 — A client that carries a token-bearing body across a redirect to another host

- incident: I11 (one incident; this case is its first round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A service sends notifications to a chat vendor. The change under review adds the sender. It posts a form to one fixed
vendor endpoint, and the form body carries the API token next to the message text. It sends through the HTTP client
that the runtime shares by default, with no redirect policy of its own. The change also adds two structural guard
tests: no code may reach the vendor library's listening (inbound) features, and the sender never branches on the kind
of destination. The change's description promises that "the endpoint is not a parameter".

Two reviewers from different model families reviewed the same change, each given a different subject on purpose. One
was told to test the repository's own guard tests and was not allowed to read the vendor library; it accepted with
conditions. The other was not allowed to read the tests, read the HTTP path instead, and rejected the change.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one sender, **base** and **change**, each offering one operation: given a message text, send it and
  answer success or error.
- **base** sends nothing; it answers success.
- **change** is built with one endpoint, fixed at construction (production uses one vendor address; the check points it
  at a local server). It sends one HTTP POST with a form-encoded body of two fields, `token` and `text`. Its client
  follows redirects to any host. On a 307 or a 308 it repeats the POST with the same body, as the HTTP semantics
  standard (RFC 9110) requires for those codes. On a 302 it follows with a GET and no body. It drops an authorization
  header when a redirect leaves the host, but the token is not in a header.
- **change** ships its two structural tests: one fails if any source of the sender reaches the vendor's listening
  features; one fails if the sender's code branches on the kind of destination.
- no separate reference: what leaves the sender is observed by the receiving servers, never by the sender itself.
- a **check** that starts two local servers, A and B. B is reached under a different host name than A (both on
  loopback). The check points the sender at A, runs every input below against base and change, and prints one line per
  input: input name, what B received under base, what B received under change. It also runs change's own structural
  tests and prints whether they pass.

Inputs the check must include:
- A answers success; no redirect (control: B receives nothing);
- A answers 307 with a location on B;
- A answers 308 with a location on B;
- A answers 302 with a location on B (control: B receives a request with no body).

The token is synthetic: `tok_synthetic_0000_not_a_secret`. Scope and assumptions: loopback servers only; redirects only
(proxy settings are out of scope, see Limits). The toolchain and any library are chosen when the case code is written,
pinned, and recorded in §8. If the chosen runtime's default client does not behave as described above, the change
implements that behaviour explicitly.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P27 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: the orchestrating agent checked the rejecting
  reviewer's finding and recorded a reject. A second round was rejected again on further findings; a third round was
  accepted and merged.

## 4. Synthetic transformation

Replaced: the vendor, its client library, the names and the layout of the original code. Kept, as the causal property
of the disagreement: the secret travels in a request body that a 307 or 308 replays, not in a header that is dropped
across hosts; the client follows redirects to any host; and the shipped tests check structure only. The reviewers'
different assigned subjects are part of the story, not of the case code.
Not kept: reviewer A's minor condition that a lookup of a feature by name at run time evades the structural guard.

## 5. Review target

The two versions of the sender, their structural tests and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the servers or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the structural guard and the destination check are real. Conditions only: a
  run-time lookup by name evades the guard, and the list of forbidden features should be confirmed complete. Evidence in
  the record: forbidden uses planted in three places, each turning the guard red; a planted destination branch turning
  the destination check red; the run-time lookup staying green. These measurements were accurate. Its assigned subject
  excluded the vendor library and the HTTP path, so it said nothing about where the token can travel.
- **Reviewer B (reject):** the sender uses the shared default client, so a 307 or 308 redirect to another host replays
  the token-bearing body to that host, and "the endpoint is not a parameter" is false as a claim about where data goes.
  Evidence in the record: a reading of the production code, of the vendor library's placement of the token in the form
  body, and of the documented redirect behaviour of the runtime's HTTP client. The reviewer did not run a redirecting
  server, and could not confirm a build in its own environment.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- The original finding also covered proxy settings taken from the environment; the check covers redirects only.
- It does not show that the original vendor would ever answer with a redirect. It shows what the client does if
  anything on the path does.
- The record does not say whether a live redirect was run when the finding was first checked; the later round required
  one as a negative control.
