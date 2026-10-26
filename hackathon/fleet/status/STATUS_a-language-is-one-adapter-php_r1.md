# STATUS — a-language-is-one-adapter-php · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 30 min
- last commit: a2cd62a21865b8b05abb1dbe8efc30f7192fa1e4 [a-language-is-one-adapter-php] fixture: Packagist in encoded prompts
- updated: 2026-10-24T07:22:47Z

## Summary

Pre-existing red: go test -count=1 ./... exits 1 on internal/pipeline TestMaterializedTreesAreReadableByOthers (mode 664). umask is denied on this seat; the same test failed before this change. Packagist is registered. The PHP scenario reports monolog/monolog 2.9.1 from composer.lock. Tip aed57813a2babacf9b150fba706eb8561ac794b2. Both commits carry a Co-authored-by trailer the tool added; the master removes it.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-language-is-one-adapter-php
- state: stop-2-ready
- abstract: Pre-existing red: go test -count=1 ./... exits 1 on internal/pipeline TestMaterializedTreesAreReadableByOthers (mode 664). umask is denied on this seat; the same test failed before this change. Packagist is registered. The PHP scenario reports monolog/monolog 2.9.1 from composer.lock. Tip aed57813a2babacf9b150fba706eb8561ac794b2. Both commits carry a Co-authored-by trailer the tool added; the master removes it.
- last commit: a2cd62a21865b8b05abb1dbe8efc30f7192fa1e4 [a-language-is-one-adapter-php] fixture: Packagist in encoded prompts
- done: Packagist adapter is in Readers and the PHP scenario reports monolog/monolog 2.9.1 from composer.lock
  - go test -count=1 ./cmd/proven/ -run TestOSVPHPComposerReportsTheDependency exit 0 after the hash was set
  - negative control mutated Readers by dropping Composer and ComposerLock: the same test exited 1 missing the finding; restored, exit 0
  - go build ./... exit 0 and go vet ./... exit 0
  - go test -count=1 ./... exit 1, only the pre-existing pipeline mode 664 failure; every other package ok, including cmd/proven and internal/review
- next / blocked-on: none
- artifacts:
  - cmd/proven/scenario_test.go
  - docs/ADAPTERS.md
  - docs/ARCHITECTURE.md
  - docs/CONFIGURATION.md
  - docs/OSV_TABLE.md
  - docs/QUERIES.md
  - docs/README.md
  - docs/SAMPLE_MODE.md
  - internal/composerver/composerver.go
  - internal/composerver/composerver_test.go
  - internal/manifest/adapter.go
  - internal/manifest/adapter_test.go
  - internal/manifest/composer.go
  - internal/manifest/composer_test.go
  - internal/manifest/composerlock.go
  - internal/manifest/composerlock_test.go
  - internal/manifest/gomod.go
  - internal/manifest/gosum.go
  - internal/manifest/npm.go
  - internal/manifest/npmlock.go
  - internal/manifest/poetry.go
  - internal/manifest/requirements.go
  - internal/osv/check.go
  - internal/osv/range.go
  - internal/osv/range_test.go
  - internal/pep440/pep440.go
  - internal/querytpl/querytpl.go
  - internal/querytpl/querytpl_test.go
  - internal/review/config.go
  - internal/review/decide_test.go
  - internal/semver/semver.go
  - testdata/fixture/discount/fixture.json
  - testdata/fixture/refuted/fixture.json
  - testdata/fixture/osv-php-composer/base/composer.json
  - testdata/fixture/osv-php-composer/base/composer.lock
  - testdata/fixture/osv-php-composer/change/composer.json
  - testdata/fixture/osv-php-composer/change/composer.lock
  - testdata/fixture/osv-php-composer/fixture.json
  - testdata/fixture/osv-php-composer/tables.sql
- evidence:
  - baseline before edits, go test -count=1 ./... exit 1, only TestMaterializedTreesAreReadableByOthers mode 664
  - go version exit 0: go1.26.8 linux/amd64; go.mod says go 1.26; no toolchain line; no Makefile; no CI file
  - start tip 87703364e86eaaa7839d890fd447748c6ad25160; git merge-base --is-ancestor 5a9f2fe HEAD exit 0
  - after the adapter wiring, go test -count=1 ./internal/manifest/ ./internal/osv/ ./internal/semver/ ./internal/pep440/ exit 0
  - go build ./... exit 0; go vet ./... exit 0; gofmt -l cmd internal printed nothing
  - negative control exit 1: missing finding osv:monolog/monolog; review class security and review rule composer.json still printed
  - restored registry, TestOSVPHPComposerReportsTheDependency exit 0
  - final go test -count=1 ./... exit 1, same single pipeline failure; cmd/proven ok; internal/review ok
  - commit 0a997fee82f8f5684d8035bef21bd415879946f2 adapters compare versions
  - commit aed57813a2babacf9b150fba706eb8561ac794b2 add the Packagist adapter
  - master-was-wrong: none. Reader fields, six constructors, and semver plus pep440 matched the 2026-10-24 reading. Comparison was a switch, which the reading did not say.
  - ruling: Adapter aliases Reader inside internal/manifest, because Collect and the tests already use Reader and a language is added by a pull request. Cost if wrong: an external module cannot import the type.
  - ruling: version order is the documented subset in internal/composerver, not every Composer form. Cost if wrong: a refused version is not checked, and the trace names it.
  - ruling: platform packages are notes, not queries. Cost if wrong: a Packagist name with a php- prefix would be skipped.
  - ruling: composer.json is a security path rule next to composer.lock. Cost if wrong: a composer.json-only change is a security review.
  - testing bar UNRUN: no metric is declared. No skip, no loosened assertion, no suppression.
  - race detector UNRUN: the change starts no goroutine and shares no state.
  - session recording: none. This spawn named no recording path and this leg wrote none.
  - both commit messages end with Co-authored-by: Cursor, added by the tool after commit. Not amended, because amend puts the trailer back.
- note: master-was-wrong: none. The 2026-10-24 reading of Reader, the six constructors, and semver plus pep440 matched the code. Comparison was a switch beside Reader, which that reading did not claim. Adapter is an alias of Reader in internal/manifest. Composer order is the documented subset. Platform packages are not queried. composer.json selects the security class. Recorded prompts now name Packagist and the recorded answers were not changed; internal/review and cmd/proven passed. umask is denied, so the pre-existing mode 664 failure was not re-run under another mask.
- updated: 2026-10-24T03:04:59Z

CLOSE-OK · a-language-is-one-adapter-php · commit a2cd62a · branch fix/a-language-is-one-adapter-php · worktree █████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
