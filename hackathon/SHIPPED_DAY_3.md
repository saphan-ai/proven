# What shipped on day 3, 24 October 2026: a walk through a terminal

Everything below was run by Marcin on 25 October 2026 at 01:40 UTC, on a fresh clone of the public `main` at `09a3a7b`,
on macOS, built with `make build`, and run with **no keys and no network** (`env -i`). The output is pasted as it came,
shortened where marked, and each step has a note for a person. To run it yourself, see
[docs/INSTALL.md](../docs/INSTALL.md); for the judges' lab, [docs/LAB.md](../docs/LAB.md). The day before:
[what shipped on day 2](SHIPPED_DAY_2.md).

## 1. A new language through one adapter: PHP

```text
$ proven scan --sample testdata/fixture/osv-php-composer
review class: security
review rule: composer.json
...
finding osv:monolog/monolog ERROR Dependency monolog/monolog 2.9.1 is in an OSV affected range
  EPSS 0.25000 (91st percentile, 2026-10-24)
  composer.lock:5-5
  monolog/monolog 2.9.1 matches GHSA-php-sample-0001: introduced 2.4.0 <= 2.9.1 < fixed 2.10.0. GHSA-php-sample-0002
  does not match: 2.9.1 < introduced 3.0.0. alias CVE-2026-9090 is on the KEV list Reachability: not measured.
  rule osv-affected-package
  evidence query local-sample snapshot=sample-2026-10-24 form=sql
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
       FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Packagist' AND package = 'monolog/monolog'
  row: GHSA-php-sample-0001 | CVE-2026-9090 | Packagist | monolog/monolog | 2.4.0 | 2.10.0 |  | HIGH | 2026-10-01
  row: GHSA-php-sample-0002 |  | Packagist | monolog/monolog | 3.0.0 | 3.1.0 |  | LOW | 2026-10-02
  ... (the KEV and EPSS queries and their rows)
  fix named from the advisory table, not run by a test: raise monolog/monolog 2.9.1 to 2.10.0
failed services: none
```

**What is new.** Proven reads a PHP project's `composer.json` and `composer.lock` through one small language adapter. The
changed package is checked against the OSV advisory table by version range, the known-exploited (KEV) list and its EPSS
score. Each step shows its evidence as the query that ran and the rows it returned. Both advisories are checked, and the
one that does not match says why. The fix is the smallest fixed version, and the line says plainly that it was named from
the table, not run by a test.

## 2. And Rust, through the same adapter

```text
$ proven scan --sample testdata/fixture/osv-rust-cargo
review class: security
review rule: Cargo.toml
...
finding osv:regex ERROR Dependency regex 1.9.1 is in an OSV affected range
  EPSS 0.33000 (91st percentile, 2026-10-24)
  Cargo.lock:15-15
  regex 1.9.1 matches GHSA-rs-sample-0001: introduced 1.7.0 <= 1.9.1 < fixed 1.10.0. ...
  ... (the same three queries, ecosystem crates.io)
  fix named from the advisory table, not run by a test: raise regex 1.9.1 to 1.10.0
failed services: none
```

**What is new.** Rust's `Cargo.toml` and `Cargo.lock`, with nothing changed in the reviewer: the adapter tells Proven where
the dependencies are and which ecosystem to ask about. That is the point of
[ADR-0038](../docs/adr/ADR-0038-a-language-is-one-small-adapter.md): a language is one small adapter.

## 3. Go modules, and the word "reachability"

```text
$ proven scan --sample testdata/fixture/osv-go-module
...
finding osv:example.com/widget ERROR Dependency example.com/widget v1.4.0 is in an OSV affected range
  EPSS 0.42000 (91st percentile, 2026-10-21)
  go.mod:5-5
  example.com/widget v1.4.0 matches GHSA-w19-sample-0001: introduced 1.3.0 <= v1.4.0 < fixed 1.5.0. ...
  alias CVE-2026-4242 is on the KEV list Reachability: not measured.
  ...
```

**What is new.** Each dependency finding now carries a word about reachability: whether your code actually calls the
vulnerable function. For Go it comes from `govulncheck` run as a tool with a local database
([ADR-0043](../docs/adr/ADR-0043-a-vulnerable-dependency-is-checked-for-reachability.md)). This sample has no such
database, so it says **not measured**, and the finding is still shown. Proven never turns "not measured" into "safe".

## 4. A fix, proved by the same test

```text
$ proven scan --sample testdata/fixture/fix
...
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test -json -run '^TestDiscounted$' ./billing
  output: base: PASS
    | change: FAIL got 60 want 80
  fix proved after 1 attempt: a suggestion for you to apply; Proven changes nothing
  fix changes: Subtract the discount once.
  fix edit billing/discount.go
  - return cents - cents*percent/100 - cents*percent/100
  + return cents - cents*percent/100
  fix run base: go test -json -run '^TestDiscounted$' ./billing exit=0 passed
  fix run change: go test -json -run '^TestDiscounted$' ./billing exit=1 failed
  fix run change+patch: go test -json -run '^TestDiscounted$' ./billing exit=0 passed
  fix run package on change+patch: go test -json ./billing exit=0 passed
finding F2 ERROR Shipping is always free
  shipping/shipping.go:5-5
  Fee returns 0 for every order, so an order of 1000 cents pays no shipping instead of 500.
  evidence test shipping/shipping_test.go TestFee fails_on_change=true passes_on_base=true
  ...
  fix not proved after 2 attempts: the patch breaks another test of the package
  fix run change+patch: go test -json -run '^TestFee$' ./shipping exit=0 passed
  fix run package on change+patch: go test -json ./shipping exit=1 failed
failed services: none
```

**What is new.** For a problem proved by a test, Proven now proposes a fix and proves it with the same test
([ADR-0042](../docs/adr/ADR-0042-a-proposed-fix-is-proved-by-the-same-test.md)). The test passes on the old code, fails on
the change, and passes on the change with the fix. Then every other test of the package runs too. F1 passes all four
runs. F2 is the interesting one: its fix makes `TestFee` pass but breaks another test in the package, so Proven does
**not** offer it, and says why. A fix is a suggestion; Proven never commits it.

## 5. JSON and SARIF, for pipelines

```text
$ proven scan --sample --json testdata/fixture/discount
{
  "class": "security",
  "class_rule": "go.mod",
  "doubled": true,
  ...
      "id": "F1",
      "title": "Discount is applied twice",
      "severity": "ERROR",
  ...
}

$ proven scan --sample --sarif testdata/fixture/discount
{
  "$schema": "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  ...
          "ruleId": "known-exploited-dependency",
          "level": "error",
  ...
          "ruleId": "discount-applied-twice",
          "level": "error",
  ...
}
```

**What is new.** The same review, written for machines. JSON carries every finding with its evidence. SARIF 2.1.0 is what
GitHub code scanning and most CI tools read, so Proven's findings can appear where developers already look. All three
outputs come from one review result. [docs/OUTPUT_FORMATS.md](../docs/OUTPUT_FORMATS.md) has the fields.

## What this walk did not show, and why

- **The level of proof per finding.** The reviewer computes it
  ([ADR-0044](../docs/adr/ADR-0044-every-finding-names-its-level-of-proof.md)), but this walk found that none of the
  three outputs prints it yet. The fix is in the command line's train.
- **The review comment on a pull request.** It is posted in the judges' lab (`make up`). Today it is the plain text above;
  a readable Markdown comment with a summary first is being built.
- **Everything live.** These runs used no keys and no network. A live review needs the keys in
  [docs/LAB.md](../docs/LAB.md).
