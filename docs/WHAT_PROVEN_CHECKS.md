# What Proven checks and proves

Proven reviews a **change between two Git revisions**. A model suggests what might be wrong. Tests, prepared data and deterministic tools supply evidence; fixed rules decide whether to report a finding.

The word **proved** below means that a finding met those rules with the recorded inputs and environment. It does not mean mathematical correctness of the application or a complete security audit.

| Check | What Proven actually does | What the evidence establishes | What it does not establish |
| --- | --- | --- | --- |
| A suspected code regression | Proposes a test through a supported language adapter, then runs the same test against base and head in restricted containers. | The recorded test passed on base and failed on head, under the runner's differential-evidence rules. The report names the test and observed outcomes. | That every regression was found, that the test covers all inputs, or that an unsupported framework has generated-test support. A timeout, build failure or missing dependency is not automatically a proved regression. |
| A newly added or upgraded vulnerable dependency | Reads supported manifests/locks and compares affected versions with imported advisory rows. | The changed dependency matches an affected range in the named snapshot. | An executed exploit, runtime reachability, or current global advisory coverage. Proven does not download advisory feeds automatically. |
| Known exploitation and likelihood context | Queries prepared KEV and EPSS tables when the review requests them. | A matching record and its recorded values were present in that snapshot. | That this application is exploitable, or that the snapshot has refreshed itself. |
| Additional data-backed claims | Runs registered query templates over prepared source tables. | The recorded rows support a claim under that template's rules. | Arbitrary model-written SQL, a live web search, or a complete copy of every source. Missing required tables make coverage incomplete. |
| Go data races | Uses the optional race-tool adapter in a prepared runtime and compares base/head results. | The recorded tool result at the evidence level shown in the report. | Race freedom. The adapter needs a compatible Go/CGO toolchain and cannot prove anything when it did not run. |
| Go vulnerability-tool results | Uses the optional govulncheck adapter with a prepared executable and local database. | The structured result and differential evidence accepted by that adapter. | An automatic online vulnerability scan. Neither the executable nor its advisory database is provisioned by merely selecting a model. |
| A proposed fix | Rechecks an optional generated patch through the configured fix gates. | The patch passed the checks actually recorded for it, including the relevant proof recheck. | That the patch is correct for every behavior. Fixes are proposals, not automatically merged changes. |
| A second opinion | Optionally asks another configured model/CLI to challenge a candidate. | Another model's objections or agreement were recorded. | Independent executable proof. Two models agreeing is not enough for a finding. |
| Declared scope, per source | Compares the PR title/description, each commit message and each collected comment separately with head code; validates excerpts and citations. | The cited text exists; an already proved finding can be linked as a counterexample. | A model's code indication does not prove feature delivery. The check can miss requirements and does not execute acceptance tests. See [delivery assessment](PR_DELIVERY.md). |
| Description completeness, without a model | Checks each source for missing/template-only, placeholder or very short text. | The declared mechanical word/letter rule passed or failed. | Meaning, sensible requirements, correctness or feature delivery. |
| Public source refresh | Acquires relevant OSV, KEV, EPSS and deps.dev data and records delta/conditional reuse. | Which provider responded, loaded content identity and available observation dates. | Exhaustive coverage, a provider observation from today or freedom from undisclosed vulnerabilities. |
| Model-free data rules | Applies registered rules to package/CVE-bound source rows. | A source rule matched those recorded rows. | Automatically a proved code defect; observations remain separate. |
| Report integrity | Optionally signs an evidence pack; `proven verify` checks its bytes and re-derives the recorded decisions. | The pack's signature and decisions agree with its recorded evidence and rules. | A fresh test execution or proof that the original machine, input data or model was trustworthy. |

## Read the outcome

| Result | Read it as | Next step |
| --- | --- | --- |
| A finding with evidence | A specific claim met Proven's evidence rules. | Inspect the test, query or tool result attached to it. |
| No findings; completion is `complete` | No defect was proved by the checks that completed. | Use the report alongside ordinary code review and tests. |
| Completion is `incomplete` | Some required work could not be settled. This can happen even with zero findings. | Read failed services, missing sources and remaining hypotheses; repair the setup or investigate the gap. |
| An unresolved planning hypothesis | A possible issue remained open at the end of planning. It is listed separately from findings. | Investigate it; it has not earned the status of a proved defect. |
| An unavailable test/tool/source | That check produced no usable evidence. | Prepare the missing runtime or data. Increasing model confidence does not replace it. |

The Action fails on findings and incomplete reviews by default. Advisory settings change whether CI fails, not what the report proves.

## Coverage today

Generated-test adapters cover **Go, Python/pytest, TypeScript/Vitest and JavaScript/node:test**. Dependency readers also cover supported npm, PyPI, Packagist, Cargo and Maven/Kotlin files; their exact syntax limits are in [language coverage](LANGUAGES.md). Data stores start without public feeds. Optional tools need their own prepared runtime.

The recorded demo replays model answers, test outcomes and synthetic source rows. It is useful for learning the report, but it is not a live measurement of your project. See [the walkthrough](WALKTHROUGH.md), [source status](SOURCES.md) and [readiness](READINESS.md).
