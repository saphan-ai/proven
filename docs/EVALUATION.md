# Evaluating a Proven deployment

Separate implementation conformance, operational compatibility and review quality. The repository's tests establish specific behavior. A successful API probe establishes one connection. Neither is a measurement of real-world defect detection.

## Three questions

| Question | Useful evidence | Does not establish |
| --- | --- | --- |
| Did this implementation obey its rules? | Unit/integration fixtures, failure controls, pack decision checks | Whether the underlying source/model was correct or every defect was found. |
| Can this deployment complete a review? | Real model calls, database queries, runner execution, retained report and trace | Accuracy across unrelated projects or languages. |
| Does it find the defects we care about? | A labelled corpus with known defects, fixed controls and misses counted | Universal precision/recall outside that corpus. |

The [sample catalogue](SAMPLE_MODE.md) is useful for learning and regression checks. Its external responses are recorded. Do not publish those replayed findings as a live-model benchmark.

## Establish controlled cases

Begin with a deliberately introduced Go regression, its corrected version and an unrelated clean change. Add a dependency raised into an affected range, its exclusive fixed boundary and an unsupported requirement form. Exercise an unavailable required source and runner no-run. For bounded planning, include both a resolved plan and final pending hypotheses.

The [SQLite tutorial](DATA_WALKTHROUGH.md) supplies a small real Git/query control using explicitly synthetic data. The test suite additionally checks that final planning gaps do not erase an independently proved finding. Those are targeted acceptance cases, not a representative security benchmark.

## Keep the experimental inputs

For each run retain:

- Target repository and exact base/head identities, including the known expected outcome.
- Proven revision/binary or image digest, policy and enabled stages/engines.
- Requested/responding model, endpoint class, extra-body controls, tokens, clocks and recovery/retry budgets.
- Source revisions/content identities, acquisition dates and normalization procedure.
- Runtime/image/toolchain identities, prepared dependency method and resource limits.
- Result, trace, exit code, total duration and any provider/runner failures.

Use sanitized profiles without keys. Provider aliases and public feeds may move independently of a Proven commit. Preserve the relevant inputs if a comparison needs to be reproducible.

## Count misses and incomplete runs

Classify expected defects that received a kept evidenced finding, expected defects with no finding, and findings that the labelled case does not support. Separately record incomplete/no-run/service-failure outcomes. Do not silently drop unsuccessful runs from the denominator or count a missing source as a clean true negative.

Review whether the evidence actually supports the named defect. A test that fails for an unrelated reason is not a useful true positive merely because a finding was kept. Report evidence kind and language/source coverage, not only a single aggregate score.

## Compare profiles fairly

Change one dimension at a time where practical: model, thinking control, token budget, bounded versus whole-change planning, or persistent engine state. Use the same cases and data. Isolate state when comparing H-MEM/World Model behavior, or document what prior reviews were available. Record provider load/rate limits and rerun policy.

Latency includes planning steps, recovery, HTTP retries, evidence execution and optional fixes. Report call failures and completion alongside speed. A faster profile that abstains from all differential tests has not demonstrated a better reviewer.

## Measurements in this repository

[Readiness](READINESS.md#what-has-been-measured) links measured operational runs and states their limits. Engine conformance results, successful sample replays and a small number of self-reviews do not establish general quality. Broader labelled evaluation and source/backend integration coverage remain [planned work](../ROADMAP.md).
