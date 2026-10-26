# STATUS — the-reviewer-model-and-the-review-class · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 40 min
- last commit: ccb0e16343bb9fbc0469b6973fbb5e1c1d6034ff [the-reviewer-model-and-the-review-class] review: author model and class
- updated: 2026-10-22T12:44:12Z

## Summary

The review states its class, the rule that chose it, and each reviewer model against the author's family. Result fields were added; nothing was renamed or removed. A bad family, class, or priority list is refused on load.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-reviewer-model-and-the-review-class
- state: stop-2-ready
- note: RED: git log --format=%B 7ad0803..HEAD | grep -ci co-authored prints 1. The message file had no Co-authored-by line. Amend stored Co-authored-by: Cursor <██████████████████████> again. --no-verify was not used. master-was-wrong: none. M1 holds (config.go's only key was timeout). M2 holds (git grep -n Class -- internal/review printed no line).
- abstract: The review states its class, the rule that chose it, and each reviewer model against the author's family. Result fields were added; nothing was renamed or removed. A bad family, class, or priority list is refused on load.
- last commit: ccb0e16343bb9fbc0469b6973fbb5e1c1d6034ff [the-reviewer-model-and-the-review-class] review: author model and class
- done: The review names its class, the rule, and the reviewer against the author's model, and a bad list is refused on load.
  - go test -count=1 ./... exited 0 after the highest-risk step was restored
  - go build ./... exited 0 and go vet ./... exited 0
  - TestHighestRiskWinsOverEarlierLowerRule failed while highest() ignored rank (got docs by **/*.md), then passed after restore
  - go run ./cmd/proven scan --sample testdata/fixture/discount exited 0, printed F1 and F2, did not print F3, and printed the class and the author line
  - an unknown family and an unknown class each printed exit status 7
- next / blocked-on: none
- artifacts:
  - cmd/proven/main.go
  - cmd/proven/main_test.go
  - internal/change/between.go
  - internal/change/between_test.go
  - internal/pipeline/imports_test.go
  - internal/review/config.go
  - internal/review/config_test.go
  - internal/review/decide.go
  - internal/review/decide_test.go
  - internal/review/decision_test.go
  - internal/review/open.go
  - internal/review/review.go
  - internal/review/selection_test.go
- evidence:
  - go version go1.26.8; go.mod says go 1.26. No Makefile and no workflow file. The acceptance command is the one that was run.
  - Baseline before the feature: go test -count=1 ./... exited 1. internal/change and internal/review failed because fixture subjects first, second, base, and head were refused by the fleet commit hook. Those subjects were changed so the hook accepts them. The tests do not assert the subject text.
  - go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./... exit 0. Every package with tests printed ok. Packages with no tests printed [no test files].
  - Unplug: highest() kept the first matching rule. go test -count=1 -run TestHighestRiskWinsOverEarlierLowerRule ./internal/review exited 1. decide_test.go:79 got docs by **/*.md, want security by **/*secret*. The comparison was restored. The same test then exited 0.
  - go run ./cmd/proven scan --sample testdata/fixture/discount exit 0. Printed review class security, review rule go.mod, author model not stated; default reviewer NVIDIA, reviewer NVIDIA; role default; family NVIDIA; author model not stated, special-review model not configured, finding F1, finding F2, failed services none. finding F3 was absent.
  - go run ./cmd/proven scan --sample testdata/fixture/discount --config=bad-family.yaml printed configuration refused: families="Google": unknown family; allowed: Anthropic, OpenAI, NVIDIA, Alibaba and exit status 7.
  - go run ./cmd/proven scan --sample testdata/fixture/discount --config=bad-class.yaml printed configuration refused: path_rules="frontend": unknown class; allowed: security, backend, mobile, user interface, docs and exit status 7.
  - remark: go run's own process exits 1 when proven exits 7. The program's code is the line exit status 7. Reading the go tool's 1 as proven's code is the false-world reading. TestScanUnknownFamilyExits7 and TestScanUnknownClassExits7 call run() and require 7, and cmd/proven tests passed.
  - A change with no trailer: TestNoTrailerSaysAuthorModelNotStated and the sample scan. The statement is author model not stated; default reviewer NVIDIA.
  - Session recording: none. ADR-0024's recorder is the Cursor recording proxy, and this worktree received no recording file.
  - testing-bar: no metric is declared in the order or the repository, so none was measured.
  - ruling: lists in proven.yaml stay flat scalars, comma-separated, because a nested value is refused. A key that is set replaces that list. Cost if wrong: a team cannot merge a short addition onto the defaults without copying them.
  - ruling: family names are the closed set Anthropic, OpenAI, NVIDIA, Alibaba. A model the table does not match is family unknown. Cost if wrong: a team cannot name another vendor's family until the set grows.
  - ruling: the default reviewer string is NVIDIA, and the default table also maps that exact name to family NVIDIA, because the ADR globs do not match the word. Cost if wrong: the shipped default would be reported as family unknown.
  - ruling: the special-review model defaults to empty, and a doubled class says it is not configured. Cost if wrong: the sample security review does not name a second model.
  - ruling: raise_class is on or off, default off, and this wagon does not ask a model to raise the class. Cost if wrong: turning the setting on does not raise anything.
  - ruling: these settings are read from the file only, as timeout already was. Environment variables do not override them. Cost if wrong: a team cannot override a file value from the environment.
  - ruling: the class is taken from the whole change, not from include or exclude. Cost if wrong: a filtered scan is still classed by files the filter hid.
  - ruling: the default Assisted-by trailer was not planted with git commit, because the fleet hook refuses that attribution. The scanner is tested on the message text. A real commit is tested with the key Model-of-record, which the hook accepts. Cost if wrong: the git path for the default key is proved by the same scanner, not by a commit that carries Assisted-by.
  - ruling: default security rules include the dependency manifests named in defaultPathRules, plus the ADR's auth, payment, secret, migrations, markdown, and docs patterns. Cost if wrong: a manifest that is not in that list falls through to the fallback class.
- updated: 2026-10-22T12:40:43Z

CLOSE-OK · the-reviewer-model-and-the-review-class · commit ccb0e16343bb9fbc0469b6973fbb5e1c1d6034ff · branch fix/the-reviewer-model-and-the-review-class · worktree ███████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
