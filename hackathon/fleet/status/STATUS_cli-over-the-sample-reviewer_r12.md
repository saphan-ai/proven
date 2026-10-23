# STATUS — cli-over-the-sample-reviewer · round 12

- state: ready-for-review
- executor: cursor (seat cursor-fleet-2-1)
- model: grok-4.7-high, on every round
- pool: unverified (no pool reading reaches a Cursor seat yet)
- cost: $0.00 recorded by Saphan Studio for each of the 12 rounds (a Cursor subscription; the run's own token counts are in its record, no dollar cost is billed per run)
- duration: round 12, 2026-10-23 04:02Z → 04:19Z
- last commit: b11e5831a3b32cbdd11f02c2d765a3902cdebf23 "[cli-over-the-sample-reviewer] agents: restore the owner file"
- updated: 2026-10-23T04:30:00Z

## Summary
proven's command line (`proven scan`, `proven ci`) now lives in `internal/cli` and reaches the reviewer only through
`internal/review`; `cmd/proven` is a thin `main`. Every flag and variable of ADR-0020 §1, flag over variable over default,
the exit codes of §2, and the text output with each finding's evidence, the review class, OSV and EPSS lines.
`proven ci` with no pull-request baseline exits 2 and says what is missing.

## Rounds
| round | why | result |
|---|---|---|
| 1–3 | the command line over the reviewer seam; the first codex review's three shell findings | merge base, precedence, panic exit 2 |
| 4, 6, 8–11 | main moved; carry onto it and move main's new output lines into `internal/cli` | each verified green |
| 5 | codex: an overflowing `--timeout` returned exit 0 | refused, exit 7 |
| 7–9 | codex: git fixture tests broken by a host's global hook | fixtures independent of host config |
| 12 | codex: `ci` without a baseline reviewed the whole checkout; an owner-owned file was edited | exit 2; the file restored |

## Acceptance, item by item (round 12)
1. `git log --oneline d488a0f..HEAD`: `b11e583 agents: restore the owner file`, `561eb8b cli: exit 2 when ci has no baseline`.
2. `git diff --stat 26343c2..HEAD -- AGENTS.md internal/review`: empty.
3. `go vet ./...` exit 0; `go test -count=1 ./...` exit 0, 38 packages, 0 failures.
4. Negative control: with the new check removed, `TestCIMissingContextExits2/no_event_no_base_ref_no_baseline` fails
   (exit 0, the whole checkout opened); restored, green.

## Master's verification (macm5, go1.26.8)
`go vet` clean, `go test ./...` 0 FAIL; `proven ci` with no context: exit 2,
`ci: no pull-request baseline: set --baseline-commit, or GITHUB_EVENT_PATH with a base SHA, or GITHUB_BASE_REF`;
the sample run of docs/SAMPLE_MODE.md from outside the repository: exit 0, 4 findings, 26 guardrail checks in the trace.
