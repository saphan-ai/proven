# PV-008 — A close routine that publishes its final frame only after an unbounded kill

- incident: I06 (a later round of the same incident as PV-007; a different finding)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A service relays a chat screen to a long-running worker process. When a session ends, the service sends one final
"ended" frame to every viewer. A viewer that subscribes after the end must still get that frame. The change under
review makes one routine the only composer of the final frame, stores it, and replays it to late subscribers. The
close routine sets the session's state to closed under a lock, releases the lock, stops the worker, and only then
composes and publishes the final frame. Stopping the worker has no time limit.

Two reviewers from different model families reviewed the same change, without seeing each other. Both found the same
defect: if stopping the worker never returns, the final frame is never published. One rated it a condition and
accepted with conditions. The other rated it blocking and rejected the change.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- a stand-in **worker** whose stop operation either returns at once or never returns (wedged), chosen by the check.
- two versions of one session, **base** and **change**, each offering these operations:
  1. subscribe, and receive frames;
  2. close the session;
  3. read the session's state.
- **base** sets the state to closed, stops the worker, and publishes the final frame to the current subscribers only.
  A viewer that subscribes after the end receives nothing.
- **change** sets the state to closed under a lock and releases it; stops the worker with no time limit; then composes
  the final frame, stores it, and publishes it. A late subscriber receives the stored frame once it exists.
- a **reference**: a model of the intended behaviour, in which stopping the worker is bounded by a time limit, and the
  final frame is published either way — with an alarm in the frame when the worker has not released. Used only by the
  check.
- a **check** that runs every input below against base, change and the reference, with a fixed wait limit for each
  observation, and prints one line per input: input name, base result, change result, reference result.

Inputs the check must include:
- a normal close, then a late subscriber: does it receive the final frame within the wait limit?
- a wedged stop, then a late subscriber: does it receive the final frame within the wait limit?
- a wedged stop, then a read of the state: what does it say?
- a wedged stop, then a second close call: does it return within the wait limit?

Scope and assumptions: one process; the wedge is a stop operation that never returns, not a real stuck process. Every
wait in the check is bounded, so the check itself always ends. The toolchain is chosen when the case code is written,
pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P14 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by the next round, which named the wedged
  close as its main fix, found by both reviewers.
- this is the same piece of work as PV-007. Other rounds of it are pairs P12 and P15 in `PROVENANCE.md`.

## 4. Synthetic transformation

Replaced: the worker, its wire format, the screen and all names. Kept, as the causal property of the disagreement: a
close routine that marks the state terminal, then calls an unbounded stop, and only then publishes the final frame that
late subscribers wait on. Both reviewers saw the same fact; they differed only on whether a liveness defect with no
alarm blocks the merge. The case keeps the fact constant; only the severity rule varies.
Not kept: blind spots in the change's source scans and in its build-pipeline control, which both reviewers also found
(see Limits).

## 5. Review target

The two versions of the session and a readable diff between them. The reviewer under evaluation receives only those and
the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the one-composer claim holds on the normal path. But a close wedged in the
  unbounded stop leaves a late viewer with nothing forever, the state reading closed, no alarm, and later close calls
  blocked. Several new controls can be beaten while the package stays green. Evidence in the record: the wedged stop
  driven, with a late subscriber receiving no frame; reverted and planted mutations that left the suite green. It
  rated all of these as conditions.
- **Reviewer B (reject):** a blocking stop leaves the state terminal while the final frame is never published, so late
  subscribers receive nothing; several mutations get past the new controls; so the change must not reach the gate.
  Evidence in the record: a successful build, a findings file with the wedged-stop path and the mutations. Its required
  race suites could not run in its environment; that does not affect the decisive claim.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- The base behaviour (no replay to late subscribers) is a synthetic simplification of the earlier version; the case is
  about the new path the change adds, not about base.
- Both reviewers also found blind spots in source scans and in a build-pipeline control. Those are not reconstructed
  here.
- It does not decide a severity rule. It shows the fact both reviewers agreed on; the verdicts differed on how to rate it.
