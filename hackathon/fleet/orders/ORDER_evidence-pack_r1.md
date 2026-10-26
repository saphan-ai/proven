# ORDER — evidence-pack · round 1

- change: evidence-pack
- round: 1
- base: main @ 1f53874
- executor: claude-code on macm5
- written: 2026-10-24T04:34:55Z
- task: every review can leave one signed evidence pack, and proven verify checks its signature, every hash and every receipt (ADR-0035)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — evidence-pack
track: proven-enterprise
- train: proven-enterprise
- wagon: W2
- review: named
- repo: proven
- base: main @ 1f53874
- branch: fix/evidence-pack
- lands-on: main
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: every review can leave one signed evidence pack, and proven verify checks its signature, every hash and every receipt (ADR-0035)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[evidence-pack] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23: features closer to enterprise. The decision is ADR-0035
(`docs/adr/ADR-0035-every-review-leaves-a-signed-evidence-pack.md`). Its four points, verbatim:
1. `--evidence-pack <file>` writes one archive per review: the findings, the trace, every test command and its output,
   every query with its rows, the Nightshift receipts, the reviewed commits and the settings used (never a secret).
2. The pack is signed with a key the organisation configures, and holds a manifest with a hash of every file.
3. `proven verify <pack>` checks the signature, every hash and every receipt, and prints what was reviewed, when, by
   which models, and on what evidence.
4. The proof is a test: a pack verifies; a pack with one byte changed is refused by name.

### THE MASTER'S READING (2026-10-24, at 1f53874 — CLAIMS to re-measure, law #44)
- The command line lives in `internal/cli` (`config.go` parses subcommands; `case "scan"` near line 352) and
  `cmd/proven`. These are the proven-cli-and-action train's files: keep the change small there and say in your status
  which of its files you touched.
- Secrets are removed by one redactor at the writers (`internal/redact`, the trace writer). The pack must go through it.
- **Step 1, before any change:** list what the review result and the trace already hold of point 1 (findings, test
  commands and output, query rows, receipts, commits, settings), and what is missing. Write it in your status.

### WHAT TO BUILD
1. `--evidence-pack <file>` on `proven scan`: one archive (tar or zip; say which and why) with the files of point 1,
   settings with every secret removed (by the redactor; never the raw environment), and `MANIFEST.json`: each file's
   path and SHA-256, the reviewed commits, the time, the models used.
2. A signature over the manifest with an Ed25519 key the organisation configures: a setting naming a private key file
   for signing (e.g. `███████████████`) and, for verify, a public key (file or setting). A missing or unreadable key is
   named, never printed; the private key never enters the pack.
3. `proven verify <pack>`: checks the signature, every file's hash against the manifest, that no file is missing or
   extra, and every Nightshift receipt the pack holds (as far as the receipt format allows offline; say what is
   checked and what is not). It prints what was reviewed, when, by which models and on what evidence, and exits 0; any
   failure is refused by name (which file, which check) with a non-zero exit.
4. `docs/EVIDENCE_PACK.md`: what is in a pack, the key settings, how an auditor verifies, and that the pack holds source
   excerpts and query rows (where it is kept is the organisation's decision). Rows in `docs/CONFIGURATION.md` via
   `cmd/settings-doc`.
5. Comments in the style of the surrounding code (ADR-0027).

### ACCEPTANCE
1. `go test ./...` green; `go vet ./...` clean.
2. Tests: a sample-mode review writes a pack that verifies; one byte changed in any file is refused naming the file; a
   changed manifest is refused by the signature check; a missing and an extra file are refused; no configured secret
   (set to a marker value) appears anywhere in the pack; a missing key is named, not printed.
3. Negative controls, named: remove the hash check — the changed-byte test turns red; remove the signature check — the
   changed-manifest test turns red. Restore both.
4. Without `--evidence-pack`, every existing test passes unchanged.

### WHAT IS NOT IN SCOPE
- The air-gapped mode (W1) and the policy (W3). JSON or SARIF writers (the CLI train's). Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the step-1 list,
the commands run and their counts, the negative controls, the CLI-train files touched, and `- master-was-wrong:` if any.
Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private credential or variable name | 1 |
| private workspace path | 4 |
