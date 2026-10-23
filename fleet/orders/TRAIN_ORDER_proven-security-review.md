# TRAIN ORDER — proven-security-review
<!-- Public copy. Original: the train's order in the fleet's private record (identical except this line). -->

- master: MASTER (claude-code, claude-opus-5-5), appointed by the owner on 2026-10-23, mode: economy
- status: DRAFT — written by the master 2026-10-23; waits for the owner's one signature (delegated train)
- track: proven-security-review
- repo: proven (`saphan-ai/proven`), the product code
- base: main at each piece's cut (main was d59e888 when this order was written)
- author: MASTER of proven-security-review, from SECURITY_REVIEW.md and ADR-0028

## §0 THE OWNER, VERBATIM

2026-10-23: *"Maybe start a systematic security review? We do it in pieces, but a whole one."* (ADR-0028, trigger)

## §1 WHAT THIS TRAIN IS

Every change to Proven gets two reviews of its diff. No review has read the code as a whole. This train reads it as a
whole, in nine pieces. Each piece is one boundary where something untrusted meets something Proven protects
([SECURITY_REVIEW.md](../../SECURITY_REVIEW.md), [ADR-0028](../decisions/ADR-0028-proven-is-security-reviewed-as-a-whole-in-pieces.md)).

When the train is done, each piece in SECURITY_REVIEW.md reads *holds*, and every weakness found is fixed, merged and
written down in SECURITY.md with the change that fixed it and the test that proves it.

This train writes no product code. It writes review orders, reads the reviews and sends each weakness to the train that
owns the code. The decisions it follows: [0005](../decisions/ADR-0005-guardrail-on-by-default.md) guardrail ·
[0006](../decisions/ADR-0006-a-skeptic-from-another-model-family.md) Skeptic ·
[0007](../decisions/ADR-0007-no-exploit-code.md) no exploit code ·
[0021](../decisions/ADR-0021-a-self-hosted-runner-demo-set-up-by-one-script.md) the runner demo ·
[0025](../decisions/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) the Prover's container ·
[0026](../decisions/ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md) the Makefile and the lab ·
[0028](../decisions/ADR-0028-proven-is-security-reviewed-as-a-whole-in-pieces.md) this review ·
[0029](../decisions/ADR-0029-a-piece-is-reviewed-at-one-commit-when-its-code-is-on-main.md) how a piece is cut.

## §2 HOW EVERY WAGON IS REVIEWED

These rules hold for every wagon below. A wagon adds only its property, its seed and its proof.

1. **A wagon is a REVIEW of `main` at one commit**, named in its order as `- reviewed-sha:`. It is not a diff review.
2. **Two independent reviews, in parallel, by two model families:** one on a claude-code seat, one on a codex seat,
   both `--role review`. Neither sees the other's notes. Never a qwen seat (security class). If the codex review
   returns no product, or its lane refuses the work as security, the second review goes to another claude-code seat
   with a different model.
3. **The order names the property that must hold**, never an attack (ADR-0007). It asks: does the property hold?
   Where it does not, the reviewer writes a test that fails on `main` and states what a fix must make it do. The test
   is returned in the review notes, not committed.
4. **The seed is a starting point, not the boundary.** Each wagon lists files where the boundary is known to be crossed.
   The reviewer finds every other crossing by search, and the notes list the searches run and what each one returned.
5. **A check the reviewer cannot run goes under `## Unrun`, and that review does not return "holds".** A review with a
   non-empty Unrun returns `accept-with-conditions` or `reject`, never `accept`.
6. **Verdict words.** `accept` = the property holds. `accept-with-conditions` = it holds with OWED or NOTE rows.
   `reject` = it does not hold; every BLOCKING row carries its failing test.
7. **The master's readings in an order are claims, dated, for the reviewer to test.** Each order has a field
   `master-was-wrong:` the reviewer fills with the measurement that contradicts the master.
8. **A weakness becomes a fix wagon in the train that owns the code.** The master sends the owning master the
   property, the finding and the failing test, by a message to its track. The master never fixes code.

   | code | owning track |
   |---|---|
   | `internal/` (the reviewer, the agents, testrun, osv, query, guardrail, and the rest) | proven-reviewer-core |
   | `cmd/proven`, `internal/cli`, the action, the workflow | proven-cli-and-action |
   | the lab, the Makefile, the images | proven-cursor-infra |
   | `cmd/` tools not named above (livecheck, osv-to-csv, settings-doc, meterless-conform) | proven-reviewer-core |

9. **A fix gets its own two reviews** in its owning train, like any change. This train then re-reads only the failing
   test and the property on the merged commit, with one review, and moves the piece forward.
10. **Publication.** The status column of SECURITY_REVIEW.md moves per piece: *not started* → *in review* → *fixes in
    progress* → *holds*. A weakness goes into SECURITY.md only after its fix is merged. Until then no finding appears
    anywhere in the proven repository: not in a status line, a log, a commit message or a public copy of an order.
    The review orders and notes of a piece are published in `fleet/reviews/` when the piece is closed.
11. **Quote per review leg:** 7 USD, cap 14 USD (the review class p90), wall 90 minutes. Eighteen first reviews:
    quote 126 USD, cap 252 USD. One fix re-read per piece at most: nine reviews, cap 126 USD. The delegated budget is the sum of the caps: 378 USD.

## §3 WAGONS

### W0 — the tools' baseline is recorded
- **property:** `govulncheck`, `gosec` and `staticcheck` have run over the whole repository at the cut, and each result
  is classed: a real weakness (with its piece), a false positive (with the reason), or out of scope.
- **seed:** the whole module (`go.mod`, every package under `cmd/` and `internal/`, `demo/`, `conformance/`).
- **proof:** each tool's version, its exact command line and its full output, saved in the notes; a table of every
  result with its class. A tool that cannot run (no vulnerability database, no install) goes under Unrun.
- **Docker Scout:** not applicable at W0, because `main` holds no image of Proven yet. It runs in W7 (ADR-0029).
- **owner of fixes:** by the class's piece.

### W1 — the Prover's container holds the change's code
- **property:** code from the change runs only in a container with no network, no secret and bounded resources.
  Nothing from the change can shape the container's arguments. The code cannot reach beyond its read-only checkout,
  gain privileges inside the container, or exhaust the host.
- **seed:** `internal/testrun/` (the container's command line, the written test), `internal/agent/prover/`, the
  settings `PROVEN_TEST_*` (docs/CONFIGURATION.md).
- **questions the review answers:** every argument of the container's command line and where its value comes from;
  where the command run inside the container comes from, and whether any part of it is text from the change or a model;
  which user the code runs as; which limits exist on processes, capabilities, privileges and the root filesystem;
  where the written test file is placed and whether its path can leave the checkout.
- **proof:** the full command line as built, with the origin of each argument; for each missing protection, a test that
  fails on `main` and names the flag or check a fix must add.
- **owner of fixes:** proven-reviewer-core.

### W2 — no secret leaves Proven's process
- **property:** no key reaches the trace, the logs, the reports, a pull-request comment, an error message or a
  container. A missing key is named, never printed.
- **seed:** the settings readers (`internal/generator/settings.go`, `internal/guardrail/settings.go`,
  `internal/query/settings.go`, `internal/review/config.go`), `cmd/livecheck`, `cmd/proven`, `internal/trace/`,
  every HTTP client that sends a key, every error that wraps a request or a response.
- **questions the review answers:** every place each secret is read, held and sent; every path by which it could be
  formatted into text (errors, `%v` of a struct, request dumps, trace events); whether livecheck's no-secret test can
  fail.
- **proof:** for each secret, a test that sets it to a marker value, runs the path, and asserts the marker is absent
  from every output; a leak is a test that fails on `main`.
- **owner of fixes:** proven-reviewer-core; proven-cli-and-action for `cmd/proven` and `internal/cli`.

### W3 — the reviewed repository is input, never instructions
- **property:** a git range, a path, a manifest, a lock file or `proven.yaml` from the change cannot inject arguments
  into a command, and cannot lead Proven to read or write outside the places it should.
- **seed:** `internal/change/` (the git calls), `internal/manifest/`, `internal/review/` (proven.yaml, author record,
  symlink checks), `internal/fixture/`, `internal/pipeline/`.
- **questions the review answers:** every git argument and whether a value can begin with `-`; every file path built
  from repository content and whether `..`, an absolute path or a symlink can move it; size limits on files read.
- **proof:** a test per crossing with a hostile-shaped but harmless value (a ref that starts with `-`, a `..` path, a
  symlink out of the target) that asserts the refusal.
- **owner of fixes:** proven-reviewer-core.

### W4 — a value from the change enters a query only as a literal
- **property:** a package name, a version, a CVE id or any other value from the change or a model enters a query only
  as a literal and cannot change the query's meaning.
- **seed:** `internal/osv/`, `internal/epss/`, `internal/agent/investigator/`, `internal/query/` (the local engine and
  the REST client), `cmd/livecheck`.
- **questions the review answers:** every query built in the code and how each value enters it; whether one quoting
  function serves every site; whether that function is correct for the query dialect, including backslashes and
  control bytes; whether a query an agent writes is limited to reading.
- **proof:** a table of every query site and its quoting; for each site that does not hold, a test with a value
  carrying a quote, a backslash or a control byte that asserts the value comes back as one literal.
- **owner of fixes:** proven-reviewer-core.

### W5 — a model is advised by the change, never ruled by it
- **property:** text in the change cannot make an agent drop its rules. The guardrail checks every prompt and every
  answer. A model's answer without evidence never becomes a finding.
- **seed:** `internal/agent/` (every agent), `internal/generator/`, `internal/guardrail/`, `internal/pipeline/`, the
  Judge.
- **questions the review answers:** every model call and whether it passes the guardrail before and after; whether
  change text is separated from the instructions in each prompt; every path from a model's answer to a finding and
  whether each requires a test result or query rows; what an `unsafe` verdict or a silent guardrail does.
- **proof:** a test per model call path, in sample mode, showing that an unchecked call or an evidence-free finding is
  refused.
- **owner of fixes:** proven-reviewer-core.

### W6 — output and CI carry no active content and the least permission
- **applies when:** the action and the workflow are on `main` (ADR-0029). Until then, W6 reviews the output formats
  only.
- **property:** a pull-request comment carries no active content taken from the change. The workflow asks for the
  least token permission, never runs on `pull_request_target`, and pins every action it uses.
- **seed:** `cmd/proven`, `internal/cli`, the output formats (markdown, SARIF, JSON), the action, the workflow files.
- **proof:** a test that renders a finding with markdown, HTML and link syntax from the change and asserts it is
  escaped; a check of the workflow's `permissions:`, triggers and pins.
- **owner of fixes:** proven-cli-and-action.

### W7 — the lab and the images are local, minimal and scanned
- **applies when:** the Makefile, the lab and Proven's image are on `main` (ADR-0029).
- **property:** the lab binds to localhost, generates its password and keeps the Prover's containers inside Docker in
  Docker. The images are minimal, do not run as root, and are scanned by Docker Scout.
- **seed:** `Makefile`, `lab/`, every Dockerfile, the Compose file, `lab/versions.env`.
- **proof:** each port binding, the password's origin, the Docker socket's absence from every container, each image's
  user, and the Docker Scout report of every image we build.
- **owner of fixes:** proven-cursor-infra.

### W8 — data on disk is outside the target, private and secret-free
- **property:** the state and the trace are kept outside the reviewed repository, readable only by their user, and hold
  no secret.
- **seed:** `internal/trace/`, `internal/hmem/`, `internal/world/`, `internal/testrun/written.go`, every file write
  in `cmd/`.
- **proof:** each write site with its path origin and file mode; a test that asserts the mode and the refusal to start
  inside the target.
- **owner of fixes:** proven-reviewer-core.

## §4 ORDER OF WORK

- **Day 1 (2026-10-23 or the day the order is signed):** W0, W1 and W2, six reviews in parallel.
- **Then two or three pieces a day:** W3 and W4, then W5 and W8, then W6 and W7 once their code is on main.
- A finding never holds the next piece. Its fix runs in the owning train while this train reviews on.
- Each piece is a stream named `security-piece-<n>`; its two reviews are dispatched against it. How the engine records
  a review of `main` with no diff is measured by rehearsal before the W0 dispatch and written into §6.

## §5 DEFINITION OF DONE

1. Every piece reads *holds* in SECURITY_REVIEW.md, on `main`.
2. Every weakness found has a merged fix, a test that failed before it and passes after it, and a line in SECURITY.md.
3. Each piece's orders and notes are in `fleet/reviews/`, published after the piece closed.
4. No finding was published before its fix was merged.
5. The record carries the closing line:
   `DONE · proven-security-review · nothing in flight · nothing owed to the owner · THANK YOU`.

## §6 LEDGER

| date | act |
|---|---|
| 2026-10-23 | order written by the master |

## §7 OUT OF SCOPE

- Fixing code. Every fix is a wagon in the owning train (§2 rule 8).
- Licences. The owner decides them once; no review checks them.
- The diff reviews of single changes; they continue in each train as before.
- Saphan, the fleet that builds Proven. This train reviews Proven's code only.
