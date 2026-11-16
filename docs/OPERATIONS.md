# Operating Proven for a team

Operate Proven as a reviewer with explicit inputs: code at base/head, a policy, model configuration, prepared evidence data and a prepared execution environment. Preserve enough of those inputs to explain a result later. Start in advisory use and evaluate known cases before requiring it on every pull request.

## Deployment choices

| Shape | Appropriate use | Prepare first |
| --- | --- | --- |
| Local CLI | A developer reviews a local Git comparison | Binary, available base commit, evidence snapshot and optional model/runner. |
| GitHub Action | Pull-request reports, comments and policy gates | Workflow permissions, secret handling, fetch depth, pinned tooling and report retention. |
| Local network | Source/data must remain within an allowed network | Local API, database, collector, images and dependencies; `PROVEN_NETWORK=local`. |
| Deterministic only | Dependency and selected tool evidence without model hypotheses | Source tables and required analyzers. Model mode `off`. |

See [Installation](INSTALL.md), [Action](ACTION.md), [Air-gapped operation](AIR_GAPPED.md) and [Readiness](READINESS.md). Proven has no hosted control plane that automatically provisions sources or manages these choices for you.

## Separate policy from operator configuration

`proven.yaml` describes repository policy: path rules, review plans, fixes, bounded planning and query scope. Environment settings select endpoints, credentials, database and runtime. CLI flags select the invocation, baseline and output destinations. Use [Configuration](CONFIGURATION.md) and [Policy](POLICY.md) as the reference for precedence and supported keys; do not invent arbitrary YAML keys for environment-only controls.

Treat changes to workflow files, policy and model profiles as reviewable operational changes. The pull request being reviewed is untrusted input; it must not choose privileged workflow code or expose secrets through its own configuration. Fork PRs need the Action's documented handling, not a blanket switch to privileged execution of their code.

## Keep state outside reviewed code

| Material | Location and handling |
| --- | --- |
| Evidence database | Default SQLite under the user cache, or an explicit external path/backend. Track source revision and freshness. |
| Trace and local engine state | Default `<user cache>/proven/logs`; use `PROVEN_LOG_DIR` outside the review target. Restrict access and set retention. |
| Reports and packs | A separate run/artifact directory. They can contain source excerpts, tests, queries and model-generated text. |
| Credentials | Environment or CI secret store; never repository files, command-line URLs or report text. |
| Runner images and caches | Prepared and scanned independently of the reviewed target. Record immutable image identities when comparing results. |

Proven rejects a log directory resolving inside the target, including symlink aliases. It also rejects unsafe links for files it writes there. H-MEM and World Model persist information across reviews when enabled. Decide whether reviewers of different trust domains should share a cache; independent directories make runs easier to isolate and reproduce.

## Establish evidence freshness

Selecting a database does not subscribe it to upstream feeds. Own the acquisition and import procedure. Keep the upstream origin, acquired-at time, revision, normalization version and intended update frequency. An empty registered table, missing table and failed read have different meanings. A content hash provides identity, not freshness or truth.

Multi-table imports publish one table at a time. If a review must use a coherent dataset release, prepare a separate database and switch consumers after all imports and validation succeed. Do not assume one import file makes all its tables visible atomically. See [Databases](DATABASES.md), [Source roadmap](SOURCES.md) and [Data walkthrough](DATA_WALKTHROUGH.md).

## Decide what blocks a merge

Use findings and completion as two separate decisions. `--error` blocks on kept findings. `PROVEN_REQUIRE_COMPLETE=true` blocks on incomplete analysis, writing the report before exit 3 when a result can be assembled. CI/Action require completion by default. A fatal refusal can exit before a complete result file exists.

Final unresolved planning hypotheses remain coverage gaps, not vulnerabilities. Their descriptions belong in the report even if all existing candidates were checked. Turning off completion enforcement does not make those hypotheses resolved. Set a team rule for who reads gaps, failed services, disagreements and unproved fixes. See [Output formats](OUTPUT_FORMATS.md) and [Markovian](MARKOVIAN.md).

## Roll out in stages

1. Run the recorded fixtures and the SQLite tutorial to learn the report contract.
2. Probe the chosen API once, then run a small real change with a prepared runner and datasets.
3. Evaluate a deliberate defect, its fixed/control version and representative changes without defects. Record misses as well as findings.
4. Use advisory PR comments while tuning data coverage, limits and model behavior.
5. Enable the required check only with an agreed policy for incomplete analysis and service outages.

Review duration grows with bounded planning steps, candidate count, second reviews, evidence runs and fixes. Size job limits from observed traces. Keep response and whole-review budgets separate. Do not hide quota or runner failures with a blanket success step.

## Upgrade and roll back

Record the old binary/image digest, policy, model profile, source snapshots and runner image before changing them. Change one dimension at a time when measuring regressions. Re-run your controlled cases, compare report structure/completion and inspect failures. For container publication, use the [build, test and scan gate](CONTAINERS.md); a new tag alone does not update a digest-pinned self-review.

Roll back to the recorded binary/image and compatible configuration if a new version fails the acceptance cases. Source timestamps and provider aliases can move independently of Proven. A rollback of Proven does not rewind an external model alias or vulnerability feed.

## Retain useful diagnostics

For a failed review, retain the exact base/head, Proven version, sanitized profile, dataset identities, runner image, exit code, report and trace. The trace is the place to distinguish rate limits, response truncation, call deadlines and unanswered planning claims. [Observability](OBSERVABILITY.md) explains local JSONL and optional OTLP export; [Troubleshooting](TROUBLESHOOTING.md) maps symptoms to checks.

Apply access controls to artifacts even when secret redaction is enabled. Redaction covers known configured secrets, not every confidential string in source or model output. Avoid reasoning-text capture unless needed for a bounded investigation. Follow your own retention and deletion policy for state and reports.
