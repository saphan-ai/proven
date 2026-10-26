# STATUS — json-sarif-outputs · round 4

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: f4819cf [json-sarif-outputs] cli: resolve output identity as the filesystem does
- updated: 2026-10-24T03:46:28Z

## Summary

outputIdentity now walks the raw path component by component (links resolved when met, `..` applied to the resolved directory, hop bound 40, rest appended lexically once a component is missing); the Abs-then-resolve path is deleted. A 10-row alias table proves exit 7 at load, both flags named, no review, no file; a second test proves the load check itself (not the write guard) refuses each row. Write guard in run.go untouched.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — json-sarif-outputs
- state: stop-2-ready
- abstract: outputIdentity now walks the raw path component by component (links resolved when met, `..` applied to the resolved directory, hop bound 40, rest appended lexically once a component is missing); the Abs-then-resolve path is deleted. A 10-row alias table proves exit 7 at load, both flags named, no review, no file; a second test proves the load check itself (not the write guard) refuses each row. Write guard in run.go untouched.
- last commit: f4819cf [json-sarif-outputs] cli: resolve output identity as the filesystem does
- done: link-then-`..` alias closed at load
  - red before the fix (old Abs path, table row `symlinked directory then ..`): `exit 2, want 7; stderr -o and --sarif-output are the same file (link/../report); -o is left as it was written`, and `the load check let "elsewhere/report" and "link/../report" through` — `go test ./internal/cli/ -run TestAliasTable`
  - green after: `go test -count=1 ./internal/cli/` ok (32s); both table tests pass on all 10 rows
  - `gofmt -l .` empty; `go vet ./...` could not complete for `internal/cli/sarifschema` only (module download blocked by the sandbox TLS); the rest compiled
  - `git diff --stat 3d54504..HEAD -- AGENTS.md internal/review` empty; `git log --oneline 3d54504..HEAD` is one commit, f4819cf
- next / blocked-on: none
- artifacts:
  - internal/cli/config.go
  - internal/cli/outputs_alias_test.go
- evidence:
  - `go test -count=1 ./...` in this sandbox: ok for cmd/proven, cmd/osv-to-csv, cmd/settings-doc, demo, internal/agent/*, change, cli, envdoc, epss, fixture, hmem, limit, manifest, osv, pep440, pep503, querytpl, redact, semver, settingspage, testrun, trace, world, test/conformance/meterless
  - FAIL, all environmental and unrelated to this diff: internal/cli/sarifschema [setup failed] (cannot download jsonschema/v5, TLS), cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review (httptest `listen tcp6 [::1]:0: bind: operation not permitted`). The master must run the full suite and sarifschema outside this sandbox; the review's own run reported them green on 3d54504.
  - `git commit` printed `packed-refs.lock: Operation not permitted`; `git log -1` shows f4819cf, the commit landed
  - master-was-wrong: — (the master's reading held; the walk-the-raw-path design closed the class as ordered)
- updated: 2026-10-24

## REDACTED

none
