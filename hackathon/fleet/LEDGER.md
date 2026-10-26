# fleet/LEDGER.md — how every change went

One row per change, added with its decision. Every number comes from the documents of that change in this folder.

## The ledger

| change | built by (model, or Auto → model) | rounds | repair rounds | BLOCKING | OWED | NOTE | outcome |
|---|---|---|---|---|---|---|---|
| world-model-bitemporal | claude-sonnet-5-5 | 4 | 3 | 4 | 2 | 0 | merged without a recorded gate |
| couchdb-evidence-viewer | not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| org-policy | claude-sonnet-5-5 | 4 | 3 | 10 | 2 | 0 | accept |
| kotlin-gradle-adapter | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| generator-adapter-example | nvidia/nemotron-3.5-super-vl-120b-a12b | 1 | 0 | 0 | 0 | 0 | accept |
| evidence-pack | claude-sonnet-5-5, grok-4.7-high | 6 | 5 | 2 | 0 | 3 | accept |
| corpus-pv-013 | grok-4.7-high, gpt-6-sol | 3 | 2 | 2 | 6 | 1 | accept |
| evidence-scope-per-role | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| owed-review-tests | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| corpus-pv-001 | grok-4.7-high, gpt-6-sol | 3 | 2 | 3 | 4 | 3 | accept |
| review-on-a-dashboard | claude-sonnet-5-5 | 3 | 2 | 1 | 1 | 1 | accept |
| proof-level-in-json-and-sarif | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| more-evidence-through-zetaris | claude-sonnet-5-5, grok-4.7-high | 2 | 1 | 0 | 0 | 0 | accept |
| the-query-seam-carries-the-template | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| held-out-set-1 | claude-sonnet-5-5 | 2 | 1 | 1 | 0 | 0 | accept |
| guardrail-trace-names-its-backend | claude-opus-5-5 | 1 | 0 | 0 | 0 | 0 | accept |
| the-image | grok-4.7-high, not recorded | 2 | 1 | 3 | 5 | 0 | accept |
| couchdb-evidence-store | claude-sonnet-5-5 | 2 | 1 | 0 | 0 | 0 | accept |
| the-source-adapter-interface | grok-4.7-high, not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| formatter-docs-and-example | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-evidence-store-interface | grok-4.7-high, not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| security-piece-8-fixes | not recorded | 2 | 1 | 1 | 3 | 0 | accept |
| judges-lab-makefile | grok-4.7-high, claude-sonnet-5-5 | 9 | 8 | 5 | 3 | 0 | accept |
| govulncheck-says-what-is-reached | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| an-agreed-opinion-is-shown-apart | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| json-sarif-outputs | claude-sonnet-5-5, grok-4.7-high | 6 | 5 | 4 | 1 | 2 | accept |
| corpus-pv-012 | gpt-6-sol | 3 | 2 | 2 | 0 | 0 | accept |
| corpus-pv-010 | claude-sonnet-5-5, gpt-6-sol | 3 | 2 | 2 | 5 | 4 | accept |
| corpus-pv-009 | gpt-6-sol | 3 | 2 | 2 | 2 | 4 | accept |
| rust-is-the-second-new-language | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| corpus-pv-008 | claude-sonnet-5-5, gpt-6-sol | 3 | 2 | 1 | 2 | 0 | accept |
| corpus-pv-006 | claude-sonnet-5-5, gpt-6-sol | 3 | 2 | 1 | 3 | 0 | accept |
| a-second-reviewer-is-any-model | claude-sonnet-5-5, not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| corpus-pv-004 | grok-4.7-high, gpt-6-sol | 3 | 2 | 2 | 6 | 0 | accept |
| corpus-pv-003 | claude-sonnet-5-5, gpt-6-sol | 3 | 2 | 1 | 1 | 3 | accept |
| a-language-is-one-adapter-php | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| security-piece-5-fixes | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| a-fix-is-proved-by-the-same-test | not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| cli-escapes-output-strips-git-env | claude-sonnet-5-5, not recorded | 4 | 3 | 2 | 1 | 0 | accept |
| the-generator-calls-nvidia-hosted | claude-sonnet-5-5 | 2 | 1 | 0 | 0 | 0 | accept |
| a-tool-run-is-evidence | grok-4.7-high | 2 | 1 | 0 | 0 | 0 | accept |
| corpus-pv-020 | grok-4.7-high, gpt-6-sol | 2 | 1 | 0 | 2 | 0 | accept |
| corpus-pv-019 | claude-sonnet-5-5, gpt-6-sol | 2 | 1 | 0 | 1 | 0 | accept |
| corpus-pv-018 | claude-sonnet-5-5, gpt-6-sol | 2 | 1 | 0 | 8 | 2 | accept |
| corpus-pv-016 | claude-sonnet-5-5, gpt-6-sol | 2 | 1 | 4 | 3 | 2 | accept |
| every-finding-names-its-proof-level | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-owed-security-items | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-test-container-is-locked-down | not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| query-evidence-comes-from-templates | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| a-git-range-wins-and-reads-are-bounded | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| no-key-reaches-a-trace-or-error | claude-sonnet-5-5, not recorded | 2 | 1 | 0 | 0 | 0 | accept |
| cli-reads-as-written | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| cli-over-the-sample-reviewer | grok-4.7-high, not recorded | 14 | 13 | 8 | 6 | 2 | accept |
| the-prover-runs-only-its-test | claude-sonnet-5-5 | 1 | 0 | 0 | 0 | 0 | accept-with-conditions |
| a-missing-key-is-a-named-service | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| a-symlink-is-skipped-and-named | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| settings-and-tools-from-the-comment-pass | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-engines-keep-memory-and-disk-equal | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| a-scenario-per-kind-of-evidence | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| a-live-review-reads-any-repository | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| poetry-locks-and-pypi-names-are-read | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| osv-git-ranges-are-not-versions | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| lock-files-are-checked | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| epss-scores-each-cve-finding | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| more-manifests-are-checked | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| dependencies-are-checked-in-osv | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-doubled-review-runs-twice | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-reviewer-model-and-the-review-class | grok-4.7-high, not recorded | 2 | 1 | 2 | 1 | 0 | accept-with-conditions |
| the-container-is-configured | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-settings-page-is-generated | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| a-live-demo-a-judge-can-rerun | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| livecheck-names-every-service | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-prover-writes-its-test | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| meterless-conformance-shim | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-skeptic-and-the-judge | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-reviewer-guards-hold | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| evidence-through-zetaris | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| world-model-and-h-mem-minimum | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| the-guardrail | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| agents-ask-a-real-model | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-generator-through-nightshift | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-prover-in-a-container-tests | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-prover-in-a-container | grok-4.7-high | 2 | 1 | 1 | 3 | 1 | accept |
| partial-reviews-report-judged-findings | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-reviewer-honours-the-request-tests | grok-4.7-high | 1 | 0 | 0 | 0 | 0 | accept |
| the-reviewer-honours-the-request | not recorded | 1 | 0 | 0 | 0 | 0 | accept |
| a-review-end-to-end-in-sample-mode | grok-4.7-high | 1 | 0 | 2 | 2 | 0 | accept |
| *(a7124ef: "Merge the master's no-keys live fix by hand" — a hand merge with no change slug; no order, status or review in the record)* | | | | | | | merged by hand |

- **rounds**: how many times the change was built. **repair rounds**: rounds after the first, each caused by a review or a decision that sent it back.
- **BLOCKING / OWED / NOTE**: conditions found by all the reviews of the change, counted by weight.
- **built by**: the model each round ran on. With Auto, the model Auto chose, as `Auto → model`.
- **outcome**: accept, accept-with-conditions, reject or close-without-shipping.

## Steering, day by day

Once a day the owner sets which models the executors use and how many run in parallel. The setting, its reason and what it produced the next day:

| day | executors in parallel | models | the owner's reason | next day: merged · repair rounds · BLOCKING |
|---|---|---|---|---|
| 22 Oct | 2 | | | |

## Totals

Counted on main since the tag `prior-art`. BLOCKING, OWED and NOTE are counted from each review's notes: one finding per line that opens with its weight; a line that recalls a previous finding is not counted again. OWED found: 73; NOTE found: 28.

| | value |
|---|---|
| changes decided | 88 |
| merged | 88 |
| BLOCKING conditions found | 66 |
| OWED conditions found | 73 |

## How good was the agents' code?

At the end of the build, on 27 October, this section answers from the ledger alone, per model:

- how many changes were accepted in their first round;
- how many repair rounds a change needed on average;
- how many BLOCKING and OWED conditions each model's changes drew per change;
- which orders produced the most repair rounds, and what those orders had in common.

The orders are public in [`orders/`](orders/), so anyone can check whether a model did badly or the order asked badly.
