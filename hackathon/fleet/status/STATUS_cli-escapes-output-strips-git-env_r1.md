# STATUS — cli-escapes-output-strips-git-env · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 1e775983602e8edc9d26ffb513dd877be288507c [cli-escapes-output-strips-git-env] cli: text shows control bytes
- updated: 2026-10-23T20:30:39Z

## Summary

Text output now shows every control byte taken from the change (C0, DEL, C1; newline kept only inside multi-line output fields) as a visible `\xNN` escape, and the merge-base git child gets childenv.Minimal(). Both properties have a test that is red with the fix removed and green with it. internal/review and AGENTS.md are untouched. Six packages are red on the base and on the tip alike, all on a sandbox refusal to open a local listener.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-escapes-output-strips-git-env
- state: stop-2-ready
- abstract: Text output now shows every control byte taken from the change (C0, DEL, C1; newline kept only inside multi-line output fields) as a visible `\xNN` escape, and the merge-base git child gets childenv.Minimal(). Both properties have a test that is red with the fix removed and green with it. internal/review and AGENTS.md are untouched. Six packages are red on the base and on the tip alike, all on a sandbox refusal to open a local listener.
- last commit: 1e775983602e8edc9d26ffb513dd877be288507c [cli-escapes-output-strips-git-env] cli: text shows control bytes
- done: both findings fixed in internal/cli, one commit each
  - P4-F5 terminal side: format.go passes every change-derived field through visible()/visibleLines(); TestTextOutputCarriesNoControlByteFromTheChange, red with isControl returning false (raw U+001B in the output), green with it
  - P2-F3: run.go gitMergeBase sets cmd.Env = childenv.Minimal(); TestGitMergeBaseChildCarriesNoKeyFromProven, red with the line removed (child env dump held ████████████████████=marker-9f3a1c), green with it; TestCIMergeBaseIsTheForkPoint green
- evidence:
  - reproduction on the base before any change: both new tests failed; raw output began `finding Fx\x1b[2Jy\rz\x7f\u009bw ERROR title...` (ESC, CR, DEL, U+009B all raw); the fake git's environment held ████████████████████=marker-9f3a1c (██████████████████████████████████████████████████████)
  - reds with each fix removed: ███████████████████████████████████████████████████████████████████████
  - git log --oneline 0b954a9..HEAD: 1e77598 [cli-escapes-output-strips-git-env] cli: text shows control bytes · 025fb8c [cli-escapes-output-strips-git-env] cli: git child gets minimal env
  - git diff --stat 0b954a9..HEAD -- AGENTS.md internal/review: empty
  - go vet ./...: exit 0; gofmt -l .: empty
  - go test -count=1 ./...: ok for cmd/proven, cmd/osv-to-csv, cmd/settings-doc, demo, internal/agent/*, change, childenv, cli, envdoc, epss, fixture, hmem, limit, manifest, osv, pep440, pep503, querytpl, redact, semver, settingspage, testrun, trace, world, test/conformance/meterless
  - red on the base AND on the tip, not caused by this leg (httptest: failed to listen on a port: bind: operation not permitted, the sandbox): cmd/livecheck TestAllServicesReady, internal/generator TestLiveReadsTheProvisionedStream, internal/guardrail TestServerPostsTheMeasuredRequest, internal/pipeline TestUnsafePromptStopsTheStepAndNamesCategories, internal/query TestLiveSelectReadsRowsFromTheSQLEditor, internal/review TestLiveReviewCarriesTheRoutedModel; the internal/review package panics on the first of them, so its other tests are UNRUN here and the master must run that package outside the sandbox
  - per-test pass counts per package were not collected (only package verdicts via go test without -v)
- next / blocked-on: none
- artifacts:
  - internal/cli/format.go
  - internal/cli/run.go
  - internal/cli/terminal_escape_test.go
  - internal/cli/mergebase_env_test.go
- note: master-was-wrong: the finding named a package name, but no review.Finding field is a package name; the path, title, message, rule id and evidence fields carry the change and all are escaped. Escape format is `\x%02x` for every control rune, so CR shows as `\x0d`. A tab is kept unescaped (it only spaces), a deviation from "C0 except LF" the order did not name. The `-C` repo argument and refs reach git as argv, not environment, and are unchanged. The commit hook printed `packed-refs.lock: Operation not permitted` after both commits; `git log` shows both landed (E13).
- updated: 2026-10-24T00:00:00Z
CLOSE-OK · cli-escapes-output-strips-git-env · commit 1e775983602e8edc9d26ffb513dd877be288507c · branch fix/cli-escapes-output-strips-git-env · worktree █████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| private credential or variable name | 2 |
