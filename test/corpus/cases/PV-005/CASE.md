# PV-005 — A reader that fetches the latest document before the one it was asked for

- incident: I04 (one incident; this case is its first round)
- intended fixture language: TypeScript (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A web service shows stored documents. A request may name a version (`?version=N`); without one, it shows the latest.
The request handler used to read the query with a parser that dropped errors, so a malformed query was served as if it
named nothing. The change under review parses the raw query strictly: a malformed query is refused with a client error
before the document store is opened. Below the handler, a reader chooses between two store fetches, "latest" and "by
version". Each fetch is expensive: it lists the store and does costly work on each document. The reader has a test named
"naming a version reads that version and nothing else"; its store double counts the calls of each kind.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
The review order pointed both at the handler's query parsing. One accepted the change: the strict parse is real, and
removing it turns its test red. The other rejected it, on the reader one layer down: a request for a version fetches the
latest document first, and the test named "nothing else" does not notice.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one service, **base** and **change**, each offering one operation: given a raw query string, return
  the document shown, or a client error; a store double records every fetch, in order.
- **base** reads the query with a parser that drops errors.
- **change** parses the raw query strictly and refuses a malformed query before any fetch.
- in both versions, the reader calls "latest" unconditionally and then, when a version is named, calls "by version"
  and returns that result instead.
- **change** ships the reader test named "naming a version reads that version and nothing else". It checks the
  "by version" call and the version it was given, but not the number of "latest" calls.
- a **reference**: the fetch list the contract allows for each input — none for a refused query, exactly one "latest"
  when no version is named, exactly one "by version" when a version is named. Used only by the check.
- a **check** that runs the shipped tests, then runs every input below against base and change, and prints one line per
  input: input name, base (result and fetch list), change (result and fetch list), reference (result and fetch list).

Inputs the check must include:
- no version named;
- a version named that exists in the store;
- a version named that the store does not have;
- a malformed query.

Scope and assumptions: an in-memory store double; no real store is opened. The toolchain and any library are chosen
when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P04 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by a later round that confirmed the finding
  and repaired the reader.

## 4. Synthetic transformation

Replaced: the service, its routes, its store and its names. Kept, as the causal property of the disagreement: the
review order points at one layer (query parsing), which really is correct; one layer down, a reader computes a default
by an expensive fetch it always makes and then overwrites when a version is named; and the test whose name promises
"nothing else" counts calls but asserts only the wanted one.
Simplified: the case puts the same reader in base and change; the record does not settle whether the extra fetch was
older than the change.

## 5. Review target

The two versions of the service, the shipped tests, and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept):** the strict query parse is real; the three version cases give three different answers (no
  version, malformed, not stored); no other route reads a version; the store is never opened on a refusal. Evidence in
  the record: a mutation back to the error-dropping parser turned the named test red, and searches for other places
  that read the query. It did not examine the reader's call sequence, so it neither saw nor refuted the finding.
- **Reviewer B (reject):** a request that names a version fetches the latest document first; the test stays green
  because it never checks the "latest" count. Evidence in the record: the reader's source, and coverage showing both
  fetches are reached on the named-version path. It did not run a mutation, and it showed no wrong document reaching a
  user. Its statements about the handler layer were from source only: its sandbox could not build that layer.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- The harm shown is an unrequested, expensive fetch, not a wrong document. The record did not measure what a failing
  "latest" fetch does to a request that names a version.
- It shows one call-sequence error behind one control; it does not show that call-counting test doubles are unreliable.
