# REVIEW — the-generator-calls-nvidia-hosted · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T03:00:32Z

## Summary

Reviewed 1a45a80 on claude-opus-5-5: nvidia mode keeps the nvidia/* check, "no receipt", the guard and Nightshift's receipt rule (14 mutants, transport probes).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-generator-calls-nvidia-hosted

Reviewer: claude-code, model `claude-opus-5-5`, 2026-10-24. Reviewed `0b954a9...1a45a80` (one commit,
`[the-generator-calls-nvidia-hosted] generator: add nvidia mode`). `git rev-parse --short=7 HEAD` printed `1a45a80`.
This is the single review; the owner's second, independent NVIDIA review through Nightshift is OWED (the order says
that route is down today).

Verdict: **accept-with-conditions** — two conditions (C1, C2), six owed items (O1–O6).

### Ground

The property (ADR-0031 point 3, ADR-0019, ADR-0005), part by part, each read in the code and tied to the instrument that
proves it:

1. **Every agent that must be answered by an `nvidia/*` model is, in nvidia mode.**
   `internal/generator/live.go:166` sends nvidia mode to `finishHosted` (`live.go:201-219`): the requirement is checked
   against the response's `model` field; a missing field is refused (`live.go:212`), a non-`nvidia/` one is refused
   (`live.go:215`). Tests: `TestNVIDIARefusesANonNVIDIAAnswerWhereRequired`,
   `TestNVIDIARefusesAnAnswerThatNamesNoModelWhereRequired`, `TestNVIDIAKeepsANonNVIDIAAnswerWhereNotRequired`.
   Mutant m1 (the hosted check forced off) turns three tests red — measured.
2. **The model named is the one that answered; "no receipt" where none exists.** `finishHosted` records
   `Use{Model: chunkModel or "unknown", ReceiptID: NoReceipt, FromReceipt: false}` (`live.go:202-208`) before the check,
   so a refused answer still names its model. `pipeline.recording.Complete` (`internal/pipeline/pipeline.go:1169-1197`,
   untouched) copies it to the review and the trace. Test: `TestNVIDIAKeepsAnNVIDIAAnswerWithNoReceipt`. **Holds only
   while the base is NVIDIA's own API — see C1.**
3. **Every call goes through the guardrail.** Read from code, not from a new test: `internal/review/open.go:85` builds
   `generator.NewLive(settings, nil)` with no mode switch, and `pipeline.guard` (`pipeline.go:1119-1129`) wraps any
   `generator.Generator` unless `GuardOff`; neither consults the generator mode. The diff touches neither path. No
   nvidia-mode end-to-end guardrail test exists (reach stated in Remainder).
4. **The key is never printed.** Load refuses a malformed key with an empty value field
   (`internal/generator/settings.go:206-207`); an empty key is `nvidia: ████████████████████ is empty`, naming the
   variable only (`settings.go:256-257`); every failure reason is a fixed string or a model name
   (`live.go`, `httpReason`), and no response body is ever put into an error. The key is in the redactor
   (`internal/redact/redact.go:26`). Tests: `TestNVIDIAKeyIsNeverInAnError` (401/403/429/500 bodies echoing the key, a
   refused answer, a refused setting), `TestNVIDIAModeNamesAMissingKey` (livecheck). **The redactor entry itself has no
   test — C2.**
5. **The key: refused when wrong, down when empty.** Wrong format → `ConfigError` (`settings.go:206`); rejected by the
   server → `nvidia: unauthorized` on 401 or 403, not retried (`live.go:255-259`, `rejectsKey` `live.go:294-296`);
   empty → `service.Down{nvidia}` and `generator.Absent{Service: "nvidia"}` (`open.go:79`). Tests:
   `TestNVIDIANamesARejectedKeyAndARateLimit` (posts == 1), `TestNVIDIAEmptyKeyIsDownAndNamed`. Mutants m3 (403 not
   rejected), m6 (429 not named) killed. Each mode reads only its own secret (`settings.go:198-202`); mutant m9 (the key
   falling back to the Nightshift token) killed by `TestNVIDIAEmptyKeyIsDownAndNamed` and
   `TestNVIDIAModeNamesAMissingKey`.
6. **Nightshift mode unchanged.** `finish` keeps its receipt rule byte-for-byte except `down` → `l.down`
   (`live.go:169-189`); `authorize` still sends `meterless-app`/`meterless-scope` (`live.go:283-290`); `rejectsKey`
   adds 403 only under `ModeNVIDIA`. The parser reads the chunk `model` field into a separate `chunkModel` that only
   `finishHosted` reads (`stream.go:79-81`). Measured by my transport probe (no port needed), PROBE-3, all PASS at tip:
   no-receipt answer refused as `nightshift: answer carries no receipt…` (3a), receipt `openai/gpt-4o` beside a chunk
   claiming `nvidia/claimed` refused by the receipt (3b), 403 stays `nightshift: http 403` (3c), bad JSON is
   `nightshift: bad stream` (3d) and `nvidia: bad stream` in nvidia mode (3e), `████████████████████` alone leaves
   nightshift mode down on `███████████████████████ is empty` (3f).

**Build / vet / test** (`go build ./... && go vet ./... && go test -count=1 ./...`, this host, sandboxed): build and vet
clean. 31 packages `ok`, 6 `FAIL` (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`,
`internal/query`, `internal/review`). `grep -c 'bind: operation not permitted'` over the run = **6**: every failure is
`httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`, i.e. this sandbox refusing
a loopback listener, not the code. Re-run with the listener tests skipped (`-skip …`): `internal/generator`,
`internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review` → `ok`; `cmd/livecheck`'s non-listener
tests (incl. `TestNVIDIAModeNamesAMissingKey`) → PASS. 31 + 6 = 37 packages, matching the master's count; the 6 are
green here only on their non-listener subset.

### Differential

What the reading says when the property is FALSE, per instrument:

- Hosted NVIDIA check off (m1): red, 3 tests. Discriminates.
- Nightshift NVIDIA check off (m2): the suite's killers (`TestLiveRefusesANonNVIDIARoutedModel`,
  `TestLiveRefusesAnAnswerWithoutAReceipt`, …) are httptest-based and could not run here → the suite reads **green**
  in this sandbox in both worlds, so my suite reading is UNMEASURED for m2. My PROBE-3 kills m2 (3a and 3b red). The
  master's "red in both modes" is therefore confirmed for nvidia mode by the suite and for nightshift mode only by my
  probe, not by the suite as I could run it.
- `████████████████████` removed from `redact.Names` (m11): **green** across `internal/redact`, `internal/pipeline`,
  `internal/review`, `cmd/livecheck` (runnable set). `grep -rn "generator.EnvKey"` finds no test that sets the key and
  asserts its redaction; `redact_test.go:74-78` iterates `Names` itself, so it cannot see an entry go missing. The
  listener tests do not reference the key either. A FALSE world reads the same as a TRUE one → C2.
- Hosted mode reading a Nightshift receipt object (m7), last chunk model winning (m8), livecheck printing a fixed
  `nightshift` mode (m13), `open.go` dropping `Service` from `Absent` (m12), 403 rejected in both modes (m4): all
  **green** on the suite → O1–O5. m4 is killed by my PROBE-3c only.

### Unplug

Mutants via `go test -overlay` (no worktree file edited), 14 in all:

| Mutant | Location | Result |
|---|---|---|
| m1 hosted NVIDIA check off | `live.go:209` | killed (3 tests) |
| m2 nightshift NVIDIA check off | `live.go:177` | survives runnable suite; killed by PROBE-3 |
| m3 403 not a rejected key | `live.go:295` | killed |
| m4 403 rejected in nightshift too | `live.go:295` | survives suite; killed by PROBE-3c |
| m5 Nightshift headers sent to NVIDIA | `live.go:285` | killed |
| m6 429 not named | `live.go:319` | killed |
| m7 hosted mode prefers a receipt model | `live.go:203` | **survives** |
| m8 last chunk model wins | `stream.go:79` | **survives** |
| m9 key falls back to the Nightshift token | `settings.go:202` | killed (generator + livecheck) |
| m10 hosted default base = Nightshift's | `settings.go:200` | killed |
| m11 key out of the redactor | `redact.go:26` | **survives** (all runnable packages) |
| m12 `Absent` loses its service | `open.go:79` | **survives** (review, non-listener) |
| m13 livecheck prints a fixed mode | `cmd/livecheck/main.go:135` | **survives** |
| m14 every failure named nightshift | `settings.go:268` | killed |

### Closure

The closure members — the values an outside party writes — are the hosted response (status, body, every chunk's
`model` and any `meterless` object) and the environment (`PROVEN_GENERATOR_MODE`, `_KEY`, `_BASE_URL`, `_MODEL`).

- Response status/body: bodies are discarded (≤4 KiB) and never reach an error; 401/403/429/other are fixed reasons.
- Chunk `model`: becomes the recorded model; it passes through the pipeline's redactor before it is kept
  (`pipeline.go:1180`) and through livecheck's (`cmd/livecheck/main.go:81`). A key echoed there is redacted only via
  `redact.go:26` (C2).
- A `meterless` object in a hosted stream: parsed but ignored by `finishHosted` (correct; untested → O1).
- `PROVEN_GENERATOR_BASE_URL` in nvidia mode: **C1**. Measured, PROBE-1: mode `nvidia`, key set,
  `PROVEN_GENERATOR_BASE_URL=http://127.0.0.1:8088` (the address `docs/LAB.md` tells a Nightshift user to set), served a
  stream whose chunk says `nvidia/claimed` and whose Nightshift receipt says `openai/gpt-4o`:
  `host=127.0.0.1:8088 keySent=true text="ok" err=<nil> use={Model:nvidia/claimed ReceiptID:no receipt}`. The NVIDIA key
  is sent to the router's address over plain http, and the router's chunk `model` field — which ADR-0019 says is not a
  receipt — satisfies the `nvidia/*` requirement. `finishHosted`'s own justification (`live.go:195-197`, "Proven calls
  NVIDIA directly and no router stands between") holds only for the default base.

### Remainder

- **Reach of point 3 (guardrail):** established by reading `open.go:85` and `pipeline.go:1119-1129`, with the
  pre-existing guardrail tests (`internal/pipeline/guardrail_test.go`, mostly listener-based, not run here). No test
  drives a nvidia-mode generator through the guard. Since the diff touches neither path, I hold it, not as measured.
- **Behaviour change in nightshift mode (O6):** `parseStream` now returns `errBadStream` instead of a
  `service.Down`, and `attempt` (`live.go:264-274`) checks the deadline before naming it. A non-JSON event arriving after
  this call's deadline is now `timeout` where it was `bad stream`. Neither is wrong; it is a change the order did not ask
  for, so it is named.
- **Deviations from the order, named:** (a) the order puts scratch under `██████████████` in the worktree, but that
  path is not git-ignored and `rm` is outside this leg's permission class, so leaving it would break "leave the worktree
  as you found it". Probe, mutants and overlays live in this run's artifact directory instead:
  `████████████████████████████████████████████████████████████████████████████████████████████████████████████████████`
  (`probe/review_probe_test.go`, `mut/spec.json`, `mut/*.go.txt`, `mut/*.json`). (b) The order spells the status H1
  `# REVIEW — the-generator-calls-nvidia-hosted-review`; the engine's channel grammar requires the reviewed slug,
  `# REVIEW — the-generator-calls-nvidia-hosted`, so the engine's form is used.
- **Declined to judge (R11):** ADR-0031 points 4 (`proven scan` runs live) and 5 (lab `.env`) — not in this commit;
  whether `defaultNVIDIAModel` and `DefaultSkepticModel` are live ids at NVIDIA (the code itself says UNRUN,
  `settings.go:52-58`); the generated `docs/CONFIGURATION.md` text beyond `cmd/settings-doc`/`settingspage` passing.

### Profile

**security** (R9: configuration, a credential, the trust placed in a provider's answer), with **docs** for
`docs/LAB.md` and the settings page.
S1 trust boundaries: Proven → hosted API (key out, answer in); environment → Proven (mode, key, base). The answer's
`model` field is now a trusted statement of identity — the new trust, and C1 is where it crosses a boundary it was not
designed for. S3: each new check unplugged once (Unplug table). S4: no dependency change (`go.mod`/`go.sum` untouched
in `--stat`); no secret in the diff — `hostedKey` (`nvidia_test.go:20`) is a labelled stand-in; no invisible or
bidirectional characters seen in the diff read. D4: `docs/LAB.md`'s steps were not executed (no real key — Unrun).

### Focus

1. *Can a hosted answer name a model it was not answered by?* With the default base: only if NVIDIA's own API
   misstates its `model` field, which is the stated, ruled trust (master's ruling 2026-10-24, ADR-0031). With a set
   base: **yes** — PROBE-1 (C1).
2. *Can it skip the guardrail?* No path found: the guard is mode-agnostic and unchanged (Ground 3, Remainder).
3. *The key:* in the redactor yes (`redact.go:26`, unguarded — C2); refused when malformed, `unauthorized` on 401/403,
   down and named when empty; absent from every error measured (`TestNVIDIAKeyIsNeverInAnError`) and from livecheck's
   output (`TestNVIDIAModeNamesAMissingKey`).
4. *Nightshift unchanged:* yes, measured by PROBE-3 (3a–3d, 3f) at tip; one incidental timing change (O6).
5. *Build/vet/test:* Ground, last paragraph.

### Conditions

- **C1 · `internal/generator/settings.go:209-216` + `live.go:192-219` · blocking-before-merge, owed by the author.**
  In nvidia mode a set `PROVEN_GENERATOR_BASE_URL` is taken over unchanged: the NVIDIA key goes to that address (plain
  http allowed) and the chunk `model` field it returns is trusted as the answering model. `docs/LAB.md` tells Nightshift
  users to set exactly that variable, so a user switching modes per the same page meets it. Trigger and evidence:
  PROBE-1 (Closure). Remedy (the master picks one): give nvidia mode its own base setting or ignore
  `PROVEN_GENERATOR_BASE_URL` there; or refuse a non-https or non-default base in nvidia mode by name; or, at minimum,
  name the base host on livecheck's generator line and in the trace's model event and say in `docs/LAB.md` and the
  settings page that the `model` field is trusted only from NVIDIA's own address. Proof expected: a test with the base
  set to another address that fails before the remedy and passes after it.
- **C2 · `internal/redact/redact.go:26` · blocking-before-merge, owed by the author.** The key's redactor entry has no
  test; mutant m11 survives every runnable package. Remedy: a test in which a hosted answer echoes
  `████████████████████` (in its `model` field) and the trace, the review's model list and livecheck's line carry
  `[redacted: ████████████████████]`; DoD names the mutated field (the key's entry in `redact.Names`) and shows the test
  red with it removed.
- **Owed, non-blocking (R6):**
  O1 `live.go:202-205` — a test that a `meterless` object in a hosted stream is ignored (m7 survives).
  O2 `stream.go:78-81` — a test that a later chunk's `model` cannot replace the first (m8 survives; PROBE-2 shows the
  behaviour is right today: `nvidia/a` then `openai/b` → `nvidia/a`).
  O3 `cmd/livecheck/main.go:135` — a ready-line test in nvidia mode (`mode nvidia, … receipt no receipt`); m13 survives.
  O4 `internal/review/open.go:79` — a review-level test that an empty NVIDIA key names `nvidia` both in failed services
  and on each refused call (m12 survives).
  O5 `live.go:295` — a nightshift-mode test that 403 stays `http 403` (m4 survives the suite).
  O6 `live.go:264-274` — the bad-stream/timeout ordering change in nightshift mode (Remainder): accept or restore, by
  the master's ruling.

### Unrun

- **Listener-based tests (every test that calls `httptest.NewServer`)** in `cmd/livecheck`, `internal/generator`
  (`live_test.go`), `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`: refused by this
  sandbox, quoted: `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. Counted 6
  panics, one per package; each package's remaining tests were re-run with `-skip`. These include the suite's whole
  Nightshift regression set and `TestAllServicesReady`; my transport probe stands in for part of it (PROBE-3), not all.
- **A live call with a real NVIDIA key** — no key, no network (order's boundary). The default model ids and the hosted
  API's real `model` field shape are therefore unconfirmed (the code says so too).
- **`docs/LAB.md` steps 1-4** executed by hand — need a real key.
- **The owner's second, independent NVIDIA review through Nightshift** — owed; that route is down today per the order.

## The review status as filed

## REVIEW — the-generator-calls-nvidia-hosted
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Reviewed 1a45a80 on claude-opus-5-5: nvidia mode keeps the nvidia/* check, "no receipt", the guard and Nightshift's receipt rule (14 mutants, transport probes).
  C1: a set PROVEN_GENERATOR_BASE_URL carries into nvidia mode — the key goes there and the router's chunk model is trusted (probe-measured).
  C2: the key's redactor entry has no test (mutant survives). Six owed tests/notes O1-O6.
  Unrun: httptest suites (sandbox refuses bind), a live key call, the owed NVIDIA-through-Nightshift review.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T03:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| private credential or variable name | 6 |
