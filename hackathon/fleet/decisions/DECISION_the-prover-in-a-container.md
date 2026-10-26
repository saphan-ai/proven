# DECISION — the-prover-in-a-container

- change: the-prover-in-a-container
- merged: bcc51cd on main, 2026-10-22T05:35:53Z

## The owner's gates

| when (UTC) | gate | decision | by | reason |
|---|---|---|---|---|
| 2026-10-22T05:28:37 | stop2 | accept | Marcin Marzec | W1b: the Prover runs the change's tests only in a container with no network and no secret (ADR-0025); codex review r5 accept at bc41431, NVIDIA review accept-with-conditions whose conditions (tests, real-Docker T10) the master measured green; absorb 24b7e34 builds and tests green |
| 2026-10-22T05:30:02 | review | accept | Marcin Marzec | W1b: the Prover runs the change's tests only in a container with no network and no secret (ADR-0025); codex review r5 accept at bc41431, NVIDIA review accept-with-conditions whose conditions (tests, real-Docker T10) the master measured green; absorb 24b7e34 builds and tests green |
| 2026-10-22T05:35:55 | merge | ready-to-merge | Marcin Marzec |  |

## REDACTED

none
