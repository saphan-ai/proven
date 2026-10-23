# PV-007 — Acknowledge-by-equal-text and a check-then-act close race

- incident: I06 (one incident; this case is its first round in the corpus; PV-008 is a later round of it)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A service relays a chat screen to a long-running worker process. A message the user writes stays "pending" until the
worker echoes it back; the echo acknowledges it. Before the change, any echo acknowledged the pending write whenever
exactly one write was pending. The change under review removes that fallback. It acknowledges a pending write only on
an echo whose text equals the write. It also refuses events that arrive after the session is closed: it checks under
the session's lock that the session is open, releases the lock, and then applies the event.

Two reviewers from different model families reviewed the same change, without seeing each other. Their orders differed
on purpose. One was told to run the controls. It showed the old fallback is gone and every listed bad echo is refused,
and accepted with conditions. The other was forbidden to run the suite and told to read the code. It rejected the
change.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one session, **base** and **change**, each offering these operations:
  1. write a text, which becomes a pending write;
  2. deliver an event from the worker: an echo of a text, an echo naming another session, or a reply;
  3. close the session;
  4. list the pending writes, the acknowledged writes, and any event applied after close.
- **base** acknowledges the pending write on any echo when exactly one write is pending.
- **change** acknowledges the oldest pending write whose text equals the echo's text. To deliver an event, it takes
  the lock, checks that the session is open, releases the lock, and then takes the lock again to apply the event.
  Close takes the same lock.
- **change** exposes a pause point between its open-check and its apply step, so the check can land a close inside
  that window deterministically. It does nothing unless the check uses it.
- **change** ships its own controls: the listed bad echoes below, each run in sequence and each with distinguishable
  text, plus an honest echo that must still acknowledge.
- a **reference**: a model of the intended behaviour, in which each write carries its own id, each echo names the id it
  answers, and the open-check and the apply step happen under one lock hold. Used only by the check.
- a **check** that runs the shipped controls on both versions, then runs every input below against base, change and the
  reference, and prints one line per input: input name, base result, change result, reference result.

Inputs the check must include:
- the order's listed shapes, in sequence: an honest echo; a duplicate echo; an empty echo; an echo of a previous turn
  with different text; an echo naming another session; two replies; an echo after close;
- write "same", acknowledge it with its honest echo, write "same" again, then deliver a delayed duplicate of the first
  echo before the second honest echo arrives;
- an event whose delivery pauses after the open-check, a close that lands in the pause, and then the apply step.

Scope and assumptions: one process; the race is made deterministic through the pause point, not by timing. The
toolchain is chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P11 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by a later round. The orchestrating agent
  ruled that both reviews were right about different things, and the next round repaired the reader's findings.
- later rounds of the same piece of work (pairs P12 and P15 in `PROVENANCE.md`) found more of the same kind: other
  routes to the same bad states, and test-side blind spots. They are not reconstructed here. Pair P14 is PV-008.

## 4. Synthetic transformation

Replaced: the worker, its wire format, the screen and all names. Kept, as the causal property of the disagreement: an
acknowledgement keyed on equal text, where an honest echo and a delayed duplicate of an earlier identical write are the
same bytes; a state check whose lock is released before the mutation, while a concurrent close takes the same lock; and
shipped controls that are sequential and use distinguishable texts, so they can observe neither.
Not kept: two further findings of the reader, about a session fork and about a pending write losing its id (see
Limits).

## 5. Review target

The two versions of the session, with the change's own controls, and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the check, the reference or
`expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the old fallback is gone, and every listed bad echo is refused by name. No
  finding blocks the merge; there are four follow-ons. Evidence in the record: the old fallback restored verbatim, with
  five controls going red while the honest-echo control stayed green; seven echo shapes driven in sequence. It did not
  try an old echo whose bytes equal a new write, and it never put a close between the check and the mutation.
- **Reviewer B (reject):** the open-check and the mutation are separate lock holds, so a close can land between them;
  equal text cannot identify a turn, so a delayed duplicate of an old "same" echo acknowledges a new "same" write; and
  two more items block the merge (see Limits). Evidence in the record: reading the code, as ordered — the lock release before the
  mutation, the text comparison, and a shipped control whose pass case and attack case are byte-identical. No execution;
  on its own, the reading does not show that the race happens in practice.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- The race is shown through a pause point the case adds. It shows the window exists; it does not show how often a real
  close lands in it. The record does not say the race was ever seen in production.
- The reader also found that a session fork adopts any id the worker claims, and that a pending write loses its id on
  the way to the screen. Those are not reconstructed here. Reviewer A saw the fork issue too and rated it a follow-on.
- It shows one property of one acknowledgement scheme; it does not show that every text comparison is wrong.
