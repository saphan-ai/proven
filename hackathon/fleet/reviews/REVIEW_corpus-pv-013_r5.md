# REVIEW — corpus-pv-013 · round 5

- verdict: accept-with-conditions
- reviewer: qwen-code, glm-5.3
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T06:48:59Z

## Summary

Second opinion on corpus case PV-013 at 6932236: accept-with-conditions. The case is faithful — base and change differ exactly where CASE.md §2 says, the expected outcome follows from the check's design with the false-world answer named, nothing private, no dangling path. The check itself could not run on this seat (permission denial, quoted in Unrun), so §8 rests on the executor's run plus a line-by-line code trace; the one condition is an independent re-run on a seat with a go grant.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-013

### Ground

- Order corpus-pv-013-review-r8, role review, second opinion from model family glm-5.3 on corpus case PV-013. The earlier codex review was not shown to me and does not bind me.
- HEAD, measured with `git rev-parse --short=7 HEAD`, output verbatim: `6932236`. Measured twice (at the role-lock and again before writing this return); identical both times. It matches `- reviewed-sha: 6932236` and `- base: review-tip/corpus-pv-013-review-r8 @ 6932236ac542abe11703dedebf1e7a8ed184e84a`.
- Scope, measured with `git log --oneline 9303b2d..6932236`: `6932236 test/corpus: name the corpus by its path after the ADR-0030 move` · `4c1d2cb [corpus-pv-013] tests: verify change-side refusal leak` · `d4ff8b7 [corpus-pv-013] tests: build the notice-config case and record the check`. `git diff --stat 9303b2d...6932236`: 7 files, 578 insertions, 5 deletions — the case's seven files plus the PV-013 entry in test/corpus/manifest.json.
- Files judged: test/corpus/cases/PV-013/CASE.md, check.sh, base/go.mod, base/main.go, change/go.mod, change/main.go, expected/EXPECTED.md, and test/corpus/manifest.json:228-243 (the PV-013 entry; status fields at 239-241).
- Toolchain: the case modules pin `go 1.26` (base/go.mod:3, change/go.mod:3); CASE.md §8 records `go version go1.26.5 linux/amd64` with GOTOOLCHAIN=local, GOPROXY=off, GOSUMDB=off. The local toolchain is unmeasured: `go version` was denied by the permission layer (Unrun).
- Worktree baseline before my acts, `git status --porcelain`: `?? ███████████`, `?? .saphan-kickoff-brief`, `?? _REVIEW_ORDER.md`, `?? _REVIEW_STATUS.md`. Final state after the return, same command: those four plus `?? _REVIEW_NOTES.md` — the channel file this order itself mandates; no tracked file touched, no scratch left. The only delta the review leaves is the order's own return.
- expected/EXPECTED.md was read last, after the code and §8, as the order instructs.

### Differential

- The two program directories are symmetric: each holds exactly go.mod and main.go (glob over the case directory).
- `git diff --no-index test/corpus/cases/PV-013/base test/corpus/cases/PV-013/change`: go.mod differs only by the module name (pv013base → pv013change; both pin `go 1.26`). main.go differs only by:
  - the package doc comment (validation and control described);
  - imports `net/url` and `regexp`;
  - `const syntheticToken` and the regexes `envName`, `channelName`;
  - the `control` subcommand (change/main.go:55-60) and its usage line (change/main.go:77);
  - `credentialsPath` (change/main.go:84-86);
  - `judge` (change/main.go:120-136): refusals for token_env, channel and endpoint, each message quoting the rejected value with `rejected value: %s` (change/main.go:124, 128, 132);
  - `validEndpoint` (change/main.go:138-147);
  - the `judge` call at the head of `setField` (change/main.go:150);
  - `runControl` (change/main.go:184-209).
- `load`, `store`, `showText`, `parseArgs` and `configPath` are byte-identical between the two versions. `setField` in base stores any value in any field (base/main.go:100-102) with no judge.
- Against CASE.md §2: base stores any value — met; change refuses a bad value for several fields, the name field among them — met; the name-field refusal message quotes the rejected value in full — met (change/main.go:124); change ships its own no-leak control that writes the synthetic token into the credentials file and checks the set messages and the show output — met (change/main.go:184-209, the search at change/main.go:204); a check runs the shipped controls on both versions and every contract input — met by check.sh.
- The other scope commits: 4c1d2cb added `assert_change_refusal_leaks` and the negative-control files and modes to check.sh; 6932236 fixed the documented command path (tests/corpus → test/corpus) in CASE.md §8 and check.sh after the ADR-0030 move, so the documented command runs from the repository root.
- The manifest entry (test/corpus/manifest.json:239-241): fixture `built 2026-10-21`, checks `run 2026-10-23`, independent_check `not done` — matches the CASE.md header. P26 in test/corpus/PROVENANCE.md maps to incident I10 and case PV-013, reviewers glm-5.2 (C) and gpt-5.6-sol (R), established defect — consistent with CASE.md §6 (Reviewer A accept with conditions, Reviewer B reject).

### Unplug

- The check's leak assertion `assert_change_refusal_leaks` (check.sh:60-63) requires both `refused: token_env` and the synthetic token in the change output. Unplug it and the pasted-value case would pass silently: the alarm that fires is the pair of inline negative controls (failure prints at check.sh:70 and check.sh:74), files crafted to miss the leaked value or to miss the refusal; if the assertion passed on either, the script prints `negative control failed: ...` and exits 1. The mechanism has an observable red.
- False-world answer (order property 2, law #52): if change's refusal did NOT print the rejected value, the `token-in-name-then-show` line would read `change_token_in_output: no`, stderr would carry `pasted-value case failed: change output lacks refusal or synthetic value` (check.sh:185), and the exit code would be 1. The recorded §8 answer — `change_token_in_output: yes` and exit 0 — cannot arise in that world. The same output does not fit the opposite outcome.
- Could `change_token_in_output: yes` arise without the defect? On change the set is refused before store, so `show` prints `(empty)`; the only source of the token in the captured output is the refusal message itself. The base side's `yes` is base storing and showing the pasted value, which §9 Limits declares openly (base is not a leak-free baseline). The two flags are printed per side, so the sides are not confused.
- The designed trap, and the wrong reading the `control` line invites: `control | base: none | change: pass | change_token_in_output: no` reads as "no leak". The shipped control sets only valid values (runControl pairs: ██████████████, #ops, https://chat.example/hook, change/main.go:189-193), so the refusal path — the only path that quotes a rejected value — never fires inside the control. The control's pass is real, and its reach ends at the paths it walks; the leak lives on the path it does not. expected/EXPECTED.md says exactly this under "what must NOT be claimed".
- One library dependency checked for robustness: `validEndpoint` uses `url.ParseRequestURI` (change/main.go:139). For the recorded input `not-a-url` the refusal holds even if the parse succeeded (the scheme would be neither http nor https); the other refusals are pure regular expressions. Per C7, the ParseRequestURI semantics are UNVERIFIED by run on this seat, and they are not load-bearing.

### Closure

BLOCKING: none. Nothing found that makes the case wrong or unrunnable.

1. The check reproduces §8 — UNRUN on this seat: the permission layer denied the run (refusal quoted in Unrun). Judged from the code instead, as the order allows. I traced check.sh against both programs line by line: every §8 line is derivable with no contradiction — the seven output lines with exit 0, and both negative-control modes with exit 1. The refusal messages go to stdout with exit 0 (`setField` returns `(msg, 0)` at change/main.go:150-152, and main prints the message on the stdout path at change/main.go:43-48), so `run_set` captures them; `oneline` joins multi-line output with ` / `; `report` greps each side for the token; the control prints `pass` (change/main.go:207) and the check greps the whole line (check.sh:97). Condition C1 carries the execution.
2. The expected outcome follows — met. The finding to prove (the name-field refusal prints the pasted token) is exactly what the `token-in-name-then-show` line records, and the false-world answer differs (Unplug). The control line proves the shipped control never walks the refusal path. The "must NOT be claimed" items hold in the recorded output: the refusals fire (bad-token-env, bad-channel, bad-endpoint lines), and the credentials-file token does not leak through show (`token-in-credentials-show`: `(empty)`, no token on either side).
3. Base and change differ exactly where CASE.md §2 says — met, by `git diff --no-index` (Differential): validation, the control, and the module name; nothing else.
4. Nothing private; no path to a place that does not exist — met. A grep over the case directory for absolute paths, user names, host addresses and private-store names (`████████████████████████████████████████████████████████████████████████████████████████`) found no match; the only URL in the case is `https://chat.example/hook` (change/main.go:192), an example domain used as a valid control input; the relative references `../../manifest.json`, `expected/EXPECTED.md` and `PROVENANCE.md` all exist; check.sh's work directory `██████████████████████████` is created by the check itself (`mkdir -p`) and emptied by its trap; the token is synthetic and follows no vendor format.

### Remainder

- Not judged: licences (out of scope per the order); the earlier codex review (not provided, by design, does not bind me); the other cases' entries in the manifest (only PV-013's entry is in scope); the root module's build and test suite (the scope adds only nested-module fixture files under test/corpus/cases/PV-013/ — the root module never compiles them — and `go` is denied on this seat in any case); the original review's second finding, deliberately not reconstructed (§9).
- NOTICED (proposals, never acted on — C5):
  - change/main.go:150-152 — a refusal returns exit code 0 with the message on stdout. CASE.md §2 asks for "accepted, or refused with a message" and no exit code, and check.sh captures stdout, so the case is self-consistent; a production CLI would exit non-zero. If the corpus ever wants the fixture to teach that too, the refusal needs an exit code and the check would capture stderr. Cost today: none.
  - check.sh:97 — the control-line guard greps the literal line `pass` (`grep -q -F -x 'pass'`); a control stubbed to print `pass` unconditionally would satisfy it. The control in change/main.go:184-209 is real by reading, so the case stands; the check itself cannot tell a real control from a stub. Cost if a future copy stubs the control: the `control` line would carry no evidence. Proposal only.

### Profile

- The order names no profile. The change is a test-corpus fixture, not production code; no production authentication, payment or data access is touched. The case's subject is a secret-leak property, so the security reading was applied in the synthetic frame: the boundary is one process, one temporary home, no network (§2); the secret is synthetic; the leak path is the refusal message (change/main.go:124); the control covers the set messages and the show output (change/main.go:194-205).
- Skills named by the order, both read before acting:
  - lang-go — the toolchain discovery is the modules' `go 1.26` pin and §8's recorded go1.26.5; the E1/E6/E10 instruments could not run (`go` denied, Unrun), so every run-based number is absent by measurement, not by omission.
  - testing-bar — no numeric metric is declared for the corpus, so there is no bar to ratchet (the skill's "do not use when" case); the case check is a binary text contract. Check ranking per the skill: the case's own check is a suite-class check; its inline negative controls are what make the assertion falsifiable; no external or project-class check applies to this fixture.

### Focus

- The order's four properties, each judged as a property that must hold: the check reproduces §8; the expected outcome follows, with the false-world answer named; base and change differ exactly where CASE.md §2 says and nowhere else; nothing private and no path to a place that does not exist. BLOCKING was reserved for what makes the case wrong or unrunnable; nothing qualified.
- Second opinion: judged without seeing the earlier codex review, as the order states.

### Conditions

- C1 (class: execution evidence owed). Re-run, on a seat whose grant admits the go toolchain: `bash test/corpus/cases/PV-013/check.sh` (expect the seven §8 lines, exit 0), `bash test/corpus/cases/PV-013/check.sh --negative-control` (expect the negative-base-only line, exit 1), `bash test/corpus/cases/PV-013/check.sh --negative-control=no-refusal` (expect the negative-no-refusal line, exit 1). Why: this seat's permission layer denied both `go version` and the check itself (Unrun), so the §8 record rests on the executor's run of 2026-10-23 plus this code trace, not on an execution by this reviewer. Cost if unmet: a transcription error in §8 could survive; the trace makes that risk small, since every line is derivable from the code.
- Rulings returned (C6 — every judgement call where the order was silent or ambiguous):
  - ruling: H1 of _REVIEW_STATUS.md — kept the engine-substituted `# REVIEW — corpus-pv-013` over the order's `# REVIEW — corpus-pv-013-review-qwen`. Why: the verdict row binds to the slug the H1 names, and a row filed under the reviewer's own slug lands where the merge never looks. Cost if wrong: none measurable.
  - ruling: artifact confinement — the order's `███████████████████████████████████████` path was not used; no scratch artifact was made (the check was denied, so there is no output to capture), and the only files written are the two channel files the order itself mandates at the worktree root. Cost if wrong: none.
  - ruling: no public copy — AGENTS.md asks for a public copy of a review under fleet/reviews/, but the order's artifact confinement and leave-the-worktree-as-found clauses forbid writes beyond the two channel files; the public copy is the master's act on his side of the return. Cost if wrong: a missing public copy, repairable by one file copy in the train.
  - ruling: verdict class — bare `accept` is refused while Unrun is not `none` (law #53), so the verdict is accept-with-conditions with condition C1. Cost if wrong: none; that is the law's shape, not a judgement.

### Unrun

- `bash test/corpus/cases/PV-013/check.sh` — denied by the permission layer before execution. Refusal, verbatim: `Qwen Code requires permission to use "run_shell_command", but that permission was declined (non-interactive mode cannot prompt for confirmation).` Tried once; not retried, not re-spelled (law #23, law #57). Judged from the code and §8 as the order instructs.
- `bash test/corpus/cases/PV-013/check.sh --negative-control` and `bash test/corpus/cases/PV-013/check.sh --negative-control=no-refusal` — same denial class; not retried.
- `go version` — same refusal, verbatim as above. The local toolchain is unmeasured; the pin is read from go.mod and §8.
- Root-module build, vet and tests (`go build ./...`, `go vet ./...`, `go test ./...`) — not run: `go` is denied on this seat, and the scope adds no file the root module compiles.
- `git check-ignore -v _REVIEW_NOTES.md _REVIEW_STATUS.md .saphan` — denied; the git facts were read from .gitignore and `git status --porcelain` instead.
- `saphan doc check _REVIEW_STATUS.md --family review-status` — the order's mandated validation of the envelope was denied on this seat. Refusal, verbatim: `Qwen Code requires permission to use "run_shell_command", but that permission was declined (non-interactive mode cannot prompt for confirmation).` Tried once; not retried (law #23, law #57). The envelope was instead written strictly to the forms the engine projects: the H1 kept byte-for-byte from the emitted skeleton, the closed key set of the order, one verdict token alone on its line, `ref` matching an artifacts entry verbatim, a nested dash-list manifest, a one-line abstract, an ISO timestamp measured with `date -u`. The master can run the doc check before the gate.

## The review status as filed

## REVIEW — corpus-pv-013
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Second opinion on corpus case PV-013 at 6932236: accept-with-conditions. The case is faithful — base and change differ exactly where CASE.md §2 says, the expected outcome follows from the check's design with the false-world answer named, nothing private, no dangling path. The check itself could not run on this seat (permission denial, quoted in Unrun), so §8 rests on the executor's run plus a line-by-line code trace; the one condition is an independent re-run on a seat with a go grant.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T23:44:31Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| private credential or variable name | 1 |
| private workspace path | 1 |
