# STATUS — the-reviewer-honours-the-request · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 5cea821ff3f [the-reviewer-honours-the-request] change: keep untracked files when the target path is a symlink or a subdirectory
- updated: 2026-10-22T04:30:59Z

## Summary

The reviewer honours baseline, head, include and exclude, and refuses a bad proven.yaml by field name.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-reviewer-honours-the-request
- state: stop-2-ready
- abstract: The reviewer honours baseline, head, include and exclude, and refuses a bad proven.yaml by field name.
- last commit: 5cea821ff3f [the-reviewer-honours-the-request] change: keep untracked files when the target path is a symlink or a subdirectory
- done: Request fields reach the pipeline and proven.yaml is refused by name on load. After 8f1a195 the master added 5cea821: ls-files --full-name and symlink-resolved paths in underTarget, so untracked files survive a macOS /var symlink and a subdirectory target, with the test TestBetweenWorkingTreeInSubdirectoryKeepsUntracked (all packages green on macOS).
  - go build ./... exit 0, go vet ./... exit 0, go test -count=1 ./... exit 0 (nine packages ok, nine with no test files, no FAIL line)
  - go run ./cmd/proven scan --sample --exclude='*' testdata/fixture/discount exit 0; stdout has no finding line (failed services: none, plus the trace path)
  - go run ./cmd/proven scan --sample --baseline-commit=does-not-exist testdata/fixture/discount prints baseline commit "does-not-exist" does not exist and exit status 2; the go run process itself exited 1
  - review.Open on a proven.yaml whose only key is colour returns configuration refused: colour="red": unknown key; allowed: timeout
  - go test -count=1 -race ./internal/change ./internal/review ./internal/pipeline ./cmd/proven exit 0
  - with keepFindings restored to return every finding, TestExcludeAllDropsFindings failed showing F1 and F2; the filter was put back and the test passed
- next / blocked-on: none
- artifacts:
  - AGENTS.md
  - cmd/proven/main.go
  - cmd/proven/main_test.go
  - internal/change/between.go
  - internal/change/between_test.go
  - internal/change/select.go
  - internal/pipeline/imports_test.go
  - internal/pipeline/pipeline.go
  - internal/review/config.go
  - internal/review/config_test.go
  - internal/review/open.go
  - internal/review/selection_test.go
- evidence:
  - █████████████████████████████████████████████████████████
  - ███████████████████████████████████████████████████████████
  - ████████████████████████████████████████████████████████████████
- note: master-was-wrong: none. M1 held — findings are filtered after the six agents and no agent package changed. M2 held — testdata/fixture/discount has no .git of its own (git toplevel is this worktree), so the two-commit tests use a temporary repository. Before the fix, both acceptance scans on this tree exited 2 with unknown flag, not 0 with F1 and F2. The sample shell now forwards --exclude, --include, --baseline-commit and --config so those commands reach Review; the command-line train still owns the other flags. imports_test.go allows os/exec only in internal/change/between.go, which runs git read commands and does not run the change. No YAML module was added: go get gopkg.in/██████████████ failed because the sumdb cache directory was absent, and a GOPATH prefix was refused. The reader accepts one flat mapping (comments, quotes, one document); a nested value is refused. AGENTS.md gained the Configuration section because the order requires the keys and defaults. testing-bar: no metric is declared, so the bar is UNRUN. lang-go: go 1.26 in go.mod, local go1.26.8, no toolchain line, no Makefile, no CI workflow in this tree. Baseline before edits was go test -count=1 ./... exit 0 on the same packages. Session recording: none — no recording path was provisioned for this leg (ADR-0024). logs/ and cmd/proven/logs/ are untracked output from the scans; rm is denied, so they were not deleted and were not committed.
- updated: 2026-10-22T03:48:44Z

remark: go test ./... prints one ok line per package and does not print per-test pass or skip counts; a 0 from a counter that was not run would be the same number as a suite that skipped every test, so those counts are UNMEASURED.
remark: go run exits 1 when the program exits 2, and the program's code is the line exit status 2; reading only the go run code would report a failure of 1 for a fatal review.

CLOSE-OK · the-reviewer-honours-the-request · commit 8f1a195b86e8446257830843469e2b51b10b3e2c · branch fix/the-reviewer-honours-the-request · worktree ████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 4 |
| e-mail or user name | 1 |
