# REVIEW — cli-over-the-sample-reviewer · round 8

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T11:53:31Z

## Summary

The CLI contract and round-13 guards pass the suite; the no-network sample control is incomplete and needs a negative control before the gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground

Reviewed HEAD `c57abe4` (measured with `git rev-parse --short=7 HEAD`), diff `90d6951...c57abe4`, and round-13 commits `19f292e..c57abe4`. The order's executor claims are hypotheses, not measurements. `go.mod` pins Go 1.26. `go test -count=1 ./...` passed for 31 tested packages, with 8 packages reporting no test files and 0 FAIL; `go vet ./internal/cli ./cmd/proven` returned 0; `git diff --check 90d6951...c57abe4` returned 0. `git diff --stat 90d6951..c57abe4 -- AGENTS.md internal/review` was empty. These instruments cannot prove a test goes red when implementation is removed.

### Differential

`internal/cli/config.go:186-223,290-394` implements ADR-0020 §1 flag/environment/default resolution; `internal/cli/config_test.go:66-562` exercises precedence, including variable and flag cases. `internal/cli/run.go:24-95,224-277` routes configuration failures to 7, fatal review failures to 2, and findings under error mode to 1; `internal/cli/run_test.go:50-411,791-923` covers exits and the CI event, merge base, and missing head. The remaining JSON and SARIF writers belong to W2 by the order; their absence is not a W1b finding. The tests and code name allowed values for invalid severity and boolean flags; unknown flags enumerate known flags.

### Unplug

Round-13 regression controls are `TestBadValuesAreRefused` (`internal/cli/config_test.go:622`), which should fail if the boolean-parse refusal in `config.go:303-307` is removed; `TestCIMissingContextExits2/explicit baseline without a head` (`internal/cli/run_test.go:838`), which should fail if the head guard in `run.go:260-266` is removed; and `TestTextCarriesOneRuleLinePerFinding` (`internal/cli/format_test.go:15`), which should fail if `format.go:159-163` is removed. These are test-predicate readings, not observed mutation runs. The no-network negative control is not discriminating: `internal/cli/run_test.go:690-730` replaces `http.DefaultTransport` but explicitly admits that a separate Transport is invisible. If sample code starts a request with its own transport, the counter still reads zero in healthy and broken worlds. A test must make outbound network attempts fail and count them across all sample adapter paths, then demonstrate red when a controlled network call is introduced.

### Closure

`internal/cli/run.go:260-266` addresses the previous empty-head finding even with an explicit baseline. `internal/cli/config.go:303-307` addresses the previous invalid-boolean exit code. `internal/cli/format_test.go:15-60` tests distinct and duplicated rule IDs. `cmd/proven/scenario_test.go:99-107` transfers W19 scenarios to `cli.Run`. The redundant workspace read was removed from `applyCI`, leaving its operative target choice in `internal/cli/config.go:330-336`. Code added in round 13 names its reasons in comments and uses property-named tests; no new readability blocker found.

### Remainder

OWED — `internal/cli/run_test.go:690-730`: replace the default-transport-only network observation with an instrument whose false-world answer differs, and show a red sample test with the network seam deliberately connected. Until then the absolute no-network claim is not verified by this test; the no-adapter-secret assertion is checked with decoy names and a counting `getenv`. The implementation's sample adapter wiring in `internal/review/open.go:47-52` is unchanged by this wagon. No blocking defect in the changed CLI was reproduced.

### Profile

Configuration and secret boundary. `internal/cli/run_test.go:700-730` checks adapter-secret lookups through the injected environment and default HTTP transport, but not every possible network transport. No dependency manifest changed in the branch diff. Refusal paths and the sample/live opener pair were read. No live external endpoint was contacted.

### Focus

ADR-0020 §§1-2, ADR-0001 sample mode, ADR-0027 readability, GitHub event fixture, and round-12 owed fixes. `--sample` still reads documented ordinary CLI environment settings; “no keys” is read here as no secret environment keys, since ADR-0020 explicitly defines `SAMPLE_MODE` and CLI variable precedence. Ruling: this interpretation preserves the stated configuration contract; if the owner meant no environment reads whatsoever, that is a different and currently unmet contract. Declined to judge: W2 JSON/SARIF writer and Action packaging, live adapter correctness, and the master's later merged tree.

### Conditions

Before the owner gate, make the sample no-network claim falsifiable using a network instrument covering alternate transports; show its negative-control red and passing result. Report the round-13 revert/red runs separately if the owner requires independently witnessed mutations rather than the existing regression test predicates.

### Unrun

Mutation runs reverting the round-13 implementation and alternate-transport network negative control were not run: this review did not alter the product code, and the existing network test cannot observe the proposed false world. Live service calls were not run: they are outside sample mode and the assigned diff.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The CLI contract and round-13 guards pass the suite; the no-network sample control is incomplete and needs a negative control before the gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T11:47:59Z

## REDACTED

none
