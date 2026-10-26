# REVIEW — judges-lab-makefile · round 5

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T11:32:43Z

## Summary

The prior tree-mode failure reproduces but is unchanged from main; it is not a lab regression. BLOCKING: first make up prints the generated admin password to default output, contrary to the order's secret-output boundary. Docker is unavailable here; live lab checks are unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — judges-lab-makefile

### Ground

**RED — BLOCKING:** The default `make up` path prints the generated admin password (`lab/████████████████`), violating the order's secret-output requirement. The inherited Go test failure is not the reason for this verdict.

`git rev-parse --short=7 HEAD` printed `02981c8`. The reviewed merge has parents `9678e102bf1d37cca3424822a9af9a838640ff08` and `c1e4ba8e5c015941b63138c24b3c5c1192f91128`. The only lab file changed relative to the first parent is `docs/LAB.md`, adding the second-reviewer section from main at a lower heading level. `git diff c1e4ba8 02981c8 -- internal/` is empty. The lab implementation is present in `git diff c1e4ba8...02981c8`.

`make check-e2e-parse` exited 0 (`e2e-parse-check: ok`); `make check-links` exited 0; `go vet ./...` exited 0. `go test ./...` exited 1, with `TestMaterializedTreesAreReadableByOthers` failing in `internal/pipeline`, reporting base and change `go.mod` at mode 664. `make up` and `make e2e` both exited 2 via Make after `docker: not found`; neither reached the lab.

remark: 2026-10-24: parser fixtures and static inspection cannot prove that an actual Gitea runner posts a review; Docker is missing here.

### Differential

The previous review's tree-mode finding is **not attributable to this branch**. With `umask 0002`, `go test -count=1 -run '^TestMaterializedTreesAreReadableByOthers$' ./internal/pipeline/` exited 1 and reported base and change `go.mod` modes 664. The code and assertion at `internal/testrun/written.go:209-224` and `internal/pipeline/trees_readable_test.go:53-58` have identical Git blob IDs at main `c1e4ba8` and HEAD (`90ed8a1849528cae4b6392abce59ac2c6dd7bdeb`, `f69d79154acd8174ff7ebf530792c0ff9073056f`). The main-side run was not repeated: identical source and no `internal/` diff establish code provenance, not a second execution result. Send the main defect to its owner; do not count it as this lab's repair debt.

remark: 2026-10-24: equal blobs answer whether these files changed, not whether the full suite passes on every host or at main runtime.

### Unplug

`lab/████████████████████` refuses a missing or unregistered runner; its false-world answer is a nonzero exit naming the runner. `lab/████████████████████` requires the current head's highest run number, successful status and a review marker from the bot account for that run and commit; an old comment, wrong author or failed status should yield nonzero. `make check-e2e-parse` checks those parser false worlds offline; the live stopped-runner and failed-workflow controls were not run. A missing Docker binary is correctly named by `make up` and `make e2e`; that measurement does not distinguish a healthy Docker daemon from a broken one.

### Closure

**RED — BLOCKING (secrets, lab implementation):** `lab/███████████████████` generates the admin password, stores it in the lab volume and prints `lab admin password: $pass` to standard output on the first `make up`. `Makefile:33-34` invokes that script by default, and `docs/LAB.md:11-13` explicitly instructs the judge to read the printed value. Thus the fresh-install path violates the order's requirement that the generated admin password never be in a log line the Makefile prints by default. Redirecting/capturing `make up` output captures a reusable secret. This is an order-mandated design in the lab diff, not a defect imported from main. Remedy: keep the credential in the protected volume and provide a deliberately invoked, access-checked retrieval method instead of printing it in default command output; update the numbered start instructions and test the default output negative control. The live command was unavailable here, but the unconditional `echo` in the reachable fresh-install branch is direct code evidence. Verdict: **reject**; the inherited Go failure is not the reason for rejection.

### Remainder

The second-reviewer text now nested under `docs/LAB.md:65-108` came from main. Its claim that adding three variables makes the lab use two reviewers is at least ambiguous: `lab/workflow/proven.yml:17-20` runs `proven scan --sample`, and `lab/compose.yaml` does not inject these variables into that job. The paragraph calls them settings in the environment Proven runs in; this is not proof that the sample lab exercises live review. Owner of the main document should distinguish live setup from the sample workflow; this review does not assign the inherited documentation claim to this stream. No dependency delta in the lab branch against `c1e4ba8`; no performance claim benchmarked. Declined to judge third-party image vulnerabilities, supported architectures and live authentication while Docker is unavailable.

### Profile

Security and docs: command-output confidentiality; local-only publish declared by `LAB_BIND=127.0.0.1` and `lab/compose.yaml` Gitea port mapping; `git check-ignore -v lab/.env` confirms the `.gitignore:5:.env` rule. `lab/versions.env` names image tags and a nonempty test-image digest; actual pulls and their digests were not verified. `lab/██████████████` uses a bot-bound marker, and `lab/████████████████████` reads the admin password from the volume; neither observation excuses the default output. No source changes were made.

### Focus

1. `lab/████████████████`, default credential disclosure: blocking; replace with opt-in protected retrieval and demonstrate default output never includes it.
2. `internal/testrun/written.go:209-224`, inherited mode bug: reproduced at tip, unchanged from main; report to the main owner, not as this lab's regression.

### Conditions

No downward reclassification requested. The master should resolve the default-output leak before a gate and repeat the live `make up`, `make e2e`, stopped-runner and failed-run checks on a Docker-capable seat. The previous review finding belongs to main and requires a separate owner; it does not repair the lab leak.

### Unrun

Live `make up`, `make e2e`, `make e2e-failed-run`, stopped-runner negative control, image pull/scan and `make down`/`make clean`: Docker is not installed (`docker: not found`). No lab resources were created here, so there is nothing to tear down. The main-side test was not executed; blob equality and the current-tip reproduction were checked instead. Race and repeat suites were not run. Dollar cost and routed model were not supplied by the seat's metering interface, and no figure is invented.

## The review status as filed

## REVIEW — judges-lab-makefile
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The prior tree-mode failure reproduces but is unchanged from main; it is not a lab regression. BLOCKING: first make up prints the generated admin password to default output, contrary to the order's secret-output boundary. Docker is unavailable here; live lab checks are unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T11:28:00Z

<!-- ENGINE-EMITTED REVIEW SKELETON, rendered from ██████████████████████████████.
     Everything below this line contains WORKED FORMS to copy as applicable.
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy the applicable form OUT.

     Write-legal `- state:` tokens: spawned · recommendation-ready.
     Last act before done: set `- state:` above to recommendation-ready and add every
     return-time key shown below. Their names, value classes, and vocabularies
     are projected from the review family's rows, never copied here.

     `verdict`: exactly one write-legal verdict token, alone on the line — mandatory for `recommendation-ready`; no parenthetical and no conditions after it; the conditions belong in the review body the `- artifacts:` manifest names
     - verdict: <accept | accept-with-conditions | reject>

     `ref`: the review's returned artifact — mandatory for `recommendation-ready`; one of this document's own `- artifacts:` entries, VERBATIM (a basename that is not a declared entry is refused at delivery); never a git ref
     - ref: _REVIEW_NOTES.md

     `artifacts`: NESTED DASH LIST of paths the review returns; mandatory for `recommendation-ready`, including an explicit empty list when nothing is returned
     - artifacts:
       - _REVIEW_NOTES.md
       - _REVIEW_STATUS.md

     An entry names a FILE, or a DIRECTORY: a directory is delivered as ONE
     deterministic `.tar` archive (sorted entries, fixed mtimes — the same
     bytes, and so the same digest, every time) recorded under the entry's
     own name plus that suffix, which is the name the projected manifest and
     the teardown gate then read. A path that EXISTS and is neither is
     refused by what it IS, never as `missing`.

     Explicit empty manifest form when the review returns no files:
     - artifacts:

     🔴 THE ANALYSIS DOES NOT GO IN THIS FILE. It goes in _REVIEW_NOTES.md, whose
     mandatory members are H2 HEADINGS — never top-level dash bullets in
     _REVIEW_STATUS.md. A section heading written as `- ground:` here is an unknown
     key: it is now reported as a WARNING rather than refused, so your
     verdict still lands — but the analysis lands nowhere at all.

     The complete _REVIEW_NOTES.md form, projected from the same one section source
     this engine renders it from (██████████████████████████████):

     # REVIEW NOTES — judges-lab-makefile

     ## Ground
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Differential
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unplug
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Closure
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Remainder
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Profile
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Focus
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Conditions
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unrun
     <replace with measured review content; write `none` only when none is the measured answer>
-->

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 9 |
