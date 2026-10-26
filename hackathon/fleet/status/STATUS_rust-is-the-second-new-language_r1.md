# STATUS — rust-is-the-second-new-language · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 27 min
- last commit: 263a68b335af06382a13f412545aaadcf94928ee [rust-is-the-second-new-language] merge: absorb main
- updated: 2026-10-24T08:06:41Z

## Summary

Cargo.toml and Cargo.lock are a crates.io adapter. The sample reports regex 1.9.1 from Cargo.lock and names 1.10.0 as the fix. The suite's one red test was already red on the base.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — rust-is-the-second-new-language
- state: stop-2-ready
- abstract: Cargo.toml and Cargo.lock are a crates.io adapter. The sample reports regex 1.9.1 from Cargo.lock and names 1.10.0 as the fix. The suite's one red test was already red on the base.
- last commit: 263a68b335af06382a13f412545aaadcf94928ee [rust-is-the-second-new-language] merge: absorb main
- done: Cargo readers, the crates.io query and the Rust sample scenario are committed
  - cargover equal, lower, higher and a refused string: go test -count=1 ./internal/cargover exit 0
  - readers and ranges: go test -count=1 ./internal/manifest ./internal/osv ./internal/querytpl exit 0
  - negative control, mutated field the two Cargo lines in Readers: TestOSVRustCargoReportsTheDependency exit 1, missing the regex finding; restored, then the same test passed inside cmd/proven
  - go build ./... exit 0; go vet ./... exit 0
  - go test -count=1 ./... exit 1 on the base and on the tip, only TestMaterializedTreesAreReadableByOthers, mode 664
- next / blocked-on: none
- artifacts:
  - internal/cargover/cargover.go
  - internal/cargover/cargover_test.go
  - internal/manifest/cargo.go
  - internal/manifest/cargo_test.go
  - internal/manifest/cargolock.go
  - internal/manifest/cargolock_test.go
  - internal/manifest/adapter.go
  - internal/manifest/adapter_test.go
  - internal/manifest/gomod.go
  - internal/osv/fix.go
  - internal/osv/fix_test.go
  - internal/osv/range_test.go
  - internal/querytpl/querytpl.go
  - internal/querytpl/querytpl_test.go
  - cmd/proven/scenario_test.go
  - docs/ADAPTERS.md
  - docs/OSV_TABLE.md
  - docs/QUERIES.md
  - docs/SAMPLE_MODE.md
  - testdata/fixture/discount/fixture.json
  - testdata/fixture/fix/fixture.json
  - testdata/fixture/refuted/fixture.json
  - testdata/fixture/osv-rust-cargo/fixture.json
  - testdata/fixture/osv-rust-cargo/tables.sql
  - testdata/fixture/osv-rust-cargo/base/Cargo.toml
  - testdata/fixture/osv-rust-cargo/base/Cargo.lock
  - testdata/fixture/osv-rust-cargo/change/Cargo.toml
  - testdata/fixture/osv-rust-cargo/change/Cargo.lock
- evidence:
  - base go test -count=1 ./... exit 1, TestMaterializedTreesAreReadableByOthers, go.mod mode 664
  - tip go build ./... exit 0
  - tip go vet ./... exit 0
  - tip go test -count=1 ./... exit 1, same test, go.mod mode 664; cmd/proven exit 0
  - local Go go1.26.10; go.mod says go 1.26 and has no toolchain line; no Makefile and no .github directory in this worktree
  - recorded prompts updated: discount 8, fix 4, refuted 2; those planner hashes were not stale because the package tests passed
  - new fixture planner hash printed by the loader: 6b165d7d8b60f4d8f554c530c0dad75d18b2fef6a20f2c59b989570c5cb0e64f
  - five commits, no Co-authored-by trailer: f6da625 cargover, a707350 manifest, 5cc69e1 querytpl, f1d3244 osv, 47a2ff4 sample
  - testing-bar UNRUN: no metric is declared
  - race detector UNRUN: the change does not touch shared state or goroutines
- note: master-was-wrong: the acceptance suite was already red on this seat. TestMaterializedTreesAreReadableByOthers failed before any edit and after, because the materialized go.mod mode was 664. umask was not measured; that command is not in the grant. remark: a run whose materialized go.mod is not group-writable would make that test pass, so this exit 1 does not by itself say the Rust adapter is wrong. Rulings: a leading v is refused (cost: an OSV bound written with v does not match); crate names are kept as written, with no case fold and no underscore-to-hyphen fold (cost: a spelling difference misses the row); every Cargo.lock [[package]] table is a pin, including the package being built (cost: that package's own version bump is reported); build-dependencies, workspace.dependencies and target-specific tables are not read, and an inline table must be one line (cost: a raise only in those shapes is missed). SmallestFix used its own switch and omitted Packagist and crates.io, so a fixed bound that leaves the ranges was reported as no fix; it now uses the adapter Compare, and the PHP sample transcript was recaptured. Noticed: docs/ARCHITECTURE.md still omits Cargo.toml. Noticed: untracked logs/trace.jsonl is from a scan whose working directory was the repository and is not part of the change. Noticed: the order named base 2afad66, which is an ancestor; the checkout already had b130a7d and these five commits sit on it.
- updated: 2026-10-24T08:00:09Z

CLOSE-OK · rust-is-the-second-new-language · commit 263a68b · branch fix/rust-is-the-second-new-language · worktree ███████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
