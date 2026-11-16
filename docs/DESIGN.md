# Design decisions and tradeoffs

These decisions explain the current public project. They are not a claim that the implementation is complete or formally verified. The code and tests define delivered behavior; [Readiness](READINESS.md) records limitations and [Roadmap](../ROADMAP.md) records planned work.

## 1. Models propose; evidence decides

**Decision.** Keep adjudication deterministic and separate from the model transport. A persuasive explanation or two agreeing models cannot replace counting evidence.

**Why.** Reviewers need to inspect what established a finding and reproduce the decision from recorded inputs. This also makes a model replacement less likely to redefine the meaning of “proved”.

**Cost.** Real defects can be missed when no runnable test, controlled query or tool observation can be obtained. Evidence requirements are not a recall guarantee. Query interpretation still has heuristic portions requiring further work.

## 2. Restrict evidence requests before execution

**Decision.** A model chooses registered query templates and parameters, not SQL. Proven constructs runner argv; a model supplies test source, not shell commands. Preparation, size limits and isolation are application responsibilities.

**Why.** A source comment or model answer is untrusted input. Adding a model safety service does not make unrestricted command/database execution safe.

**Cost.** New sources, tools and languages need real adapters and tests. Some useful requests will be refused until their controlled path exists. The runtime, image and host remain trusted infrastructure.

## 3. Treat models and databases as replaceable adapters

**Decision.** Default to no model when unconfigured and local SQLite for storage. Support a generic API, an operator-configured CLI and alternative stores without requiring a sponsor account. Preserve existing explicit legacy transports for migration.

**Why.** Endpoint, account, subscription, data location and runtime are operator choices. The core evidence rules should not depend on a routing vendor or a specific database product.

**Cost.** “OpenAI-compatible” is not one identical behavior. Thinking fields, token accounting, stream details and vendor CLI envelopes vary. Profiles must be explicit, probed and evaluated. A backend also does not automatically acquire upstream data.

## 4. Configure response behavior per model and agent

**Decision.** Separate header, idle and attempt clocks; retain the outer review deadline. Keep token fields, thinking controls, retry budgets and per-agent overrides configurable. Observe reasoning separately from final content.

**Why.** A thinking model can be active without answer text. A long Prover response and a short query selector have different budgets. A single hard-coded timeout or thinking mode would hide those differences.

**Cost.** Operators need a measured profile. Large output limits can consume the whole review budget. Retries are bounded and do not rescue a timed-out or truncated answer; observability must explain which limit failed.

## 5. Report coverage gaps separately from findings

**Decision.** Preserve final unresolved bounded-planning descriptions and continue checking collected risks. Mark completion incomplete, without promoting the pending descriptions to vulnerabilities. Keep mid-plan carried-state overflow as a refusal.

**Why.** Discarding already collected risks loses useful work; pretending pending claims disappeared hides coverage. The last planning step and an unsafe mid-plan state overflow are different situations.

**Cost.** Consumers must read completion as well as findings. Strict CI may still fail with a useful report. Advisory use does not remove the need to inspect gaps. See [Markovian](MARKOVIAN.md).

## 6. Require observable checks for proposed fixes

**Decision.** Validate source fixes against base, head, patched head and package tests. Whole-change autofix makes one proposal and rechecks each finding plus touched packages. Never commit to the author branch automatically.

**Why.** A patch that changes a symptom but breaks neighboring behavior is not a proved repair. Human review still needs the diff and evidence.

**Cost.** Offline dependencies, unsupported files, packages without tests and collateral failures can reject a plausible patch. Advisory fixed versions remain distinct from executed source-fix proof. See [Fixes](FIXES.md).

## 7. Preserve evidence integrity without overstating it

**Decision.** Record query rows, source identities, test/tool outcomes and model identity claims. Offer signed packs and offline outcome re-derivation.

**Why.** A later reader should detect modified bytes and assess how recorded evidence led to the decision without keeping the original services running.

**Cost.** A valid signature proves custody under a key, not that a test executed honestly or that an endpoint used the claimed model. Verification is not execution replay or completeness certification. See [Evidence packs](EVIDENCE_PACK.md).

## 8. Keep reusable code independent of application setup

**Decision.** Keep candidate/report/observation values and deterministic judgment in small packages with explicit inputs. Configure services at the application boundary. Preserve focused tests with selected code.

**Why.** Another reviewer can take useful logic without adopting Proven's CLI, persistent state, environment loading or deployment. Proven can also evolve independently.

**Cost.** Current `internal/` packages are not a stable importable SDK. Reuse involves selecting code/tests and adapting the receiving boundary. A public API promise would need separate versioning and compatibility work. See [Adapters](ADAPTERS.md#reusing-review-logic).

## 9. Gate publication on the actual image contents

**Decision.** Keep Alpine as the runtime base, install explicitly pinned patched security packages above it, test both Linux architectures and reject HIGH/CRITICAL scan results before pushing the product image.

**Why.** A small base image can still contain a vulnerable Git or libc. A scan of one architecture does not establish the contents of the other.

**Cost.** Security overlays need maintenance, exact packages can disappear and scanners change their databases. A clean scan is a point-in-time observation, not permanent absence of vulnerabilities. See [Containers](CONTAINERS.md).
