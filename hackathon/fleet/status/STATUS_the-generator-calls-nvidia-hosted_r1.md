# STATUS — the-generator-calls-nvidia-hosted · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 1a45a80 [the-generator-calls-nvidia-hosted] generator: add nvidia mode
- updated: 2026-10-23T20:58:03Z

## Summary

The generator has a second provider, NVIDIA's hosted API, chosen by PROVEN_GENERATOR_MODE=nightshift|nvidia (default nightshift). In nvidia mode the key is ████████████████████ (in the redactor), the answering model is the response's model field recorded with "no receipt", and the nvidia/* requirement is checked against it; livecheck names the mode and a missing or rejected key; CONFIGURATION.md is regenerated and docs/LAB.md is new.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-generator-calls-nvidia-hosted
- state: stop-2-ready
- abstract: The generator has a second provider, NVIDIA's hosted API, chosen by PROVEN_GENERATOR_MODE=nightshift|nvidia (default nightshift). In nvidia mode the key is ████████████████████ (in the redactor), the answering model is the response's model field recorded with "no receipt", and the nvidia/* requirement is checked against it; livecheck names the mode and a missing or rejected key; CONFIGURATION.md is regenerated and docs/LAB.md is new.
- last commit: 1a45a80 [the-generator-calls-nvidia-hosted] generator: add nvidia mode
- done: nvidia mode built, tested and documented in one commit
  - mode, key, defaults, wrong mode refused, empty key = nvidia down and named: internal/generator/settings.go; tests TestLoadNVIDIAModeDefaults, TestLoadRefusesAWrongModeAndDefaultsToNightshift, TestNVIDIAEmptyKeyIsDownAndNamed pass
  - nvidia/* answer kept, non-nvidia refused where required, no model refused, "no receipt" recorded, no Nightshift headers sent, 401/403/429 named: internal/generator/nvidia_test.go, 8 tests pass
  - key never in an error; key in the redactor: redact.Names holds EnvKey, TestFromEnvReadsEveryNamedVariable iterates Names
  - livecheck checks the mode in use and names a missing key: TestNVIDIAModeNamesAMissingKey passes
  - negative control (law #27): `if false &&` before the nvidiaModel check in finishHosted made TestNVIDIARefusesANonNVIDIAAnswerWhereRequired red (text "ready" err nil); restored, green
- evidence:
  - master-was-wrong: nothing refuted. One correction to the order's wording: it asks for an httptest server, but this sandbox cannot bind a port (listen tcp6 [::1]:0: operation not permitted), so the new tests serve recorded responses through the http client's transport. Same recorded responses, no port.
  - go build ./... exit 0; go vet ./... exit 0; gofmt -l . empty
  - go test -count=1 ./... : ok for every package except the 6 named in note, FAIL there with the port panic (same as base)
  - go test -skip '^(TestLive|TestClosedPort)' ./internal/generator ./internal/redact ./internal/settingspage ./cmd/settings-doc : ok x4
  - go run ./cmd/settings-doc exit 0 (docs/CONFIGURATION.md regenerated)
- next / blocked-on: master runs the full suite where a port can be bound (the 6 red packages below) and, if wanted, one live call with a real key
- note: UNRUN here: every httptest.NewServer test (the sandbox denies binding a port) and any call with a real key; the new tests use a transport, not a port. Baseline on main 0b954a9 (E10) already red in 6 packages for the port denial only: cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review. The first panic aborts a package, so its other tests are UNRUN too. After my change the same 6 are red for the same panic and no other package is red. Run instead: generator/redact/settingspage/settings-doc with -skip '^(TestLive|TestClosedPort)' green; livecheck TestNVIDIAMode green. Judgement calls: (1) default hosted model nvidia/llama-3.3-nemotron-super-49b-v1.5 is from my knowledge of NVIDIA's published list, not confirmed by a live call; the Skeptic keeps the existing default; (2) in nvidia mode 403 is also read as a rejected key; (3) a receipt of "no receipt" is carried as Use.ReceiptID = generator.NoReceipt with FromReceipt false, no new trace field; (4) livecheck's generator line now begins "mode <mode>, ", so two existing expectations in cmd/livecheck/main_test.go were edited; (5) HTTP 429 reason is now "http 429, rate limited". The commit shows a packed-refs.lock permission message but landed. No Co-authored-by line on the commit. Session recording: none made by this leg (headless, no recorder available to it).
- artifacts:
  - internal/generator/live.go
  - internal/generator/settings.go
  - internal/generator/nvidia_test.go
  - docs/CONFIGURATION.md
  - docs/LAB.md
- updated: 2026-10-24T00:00:00Z

CLOSE-OK · the-generator-calls-nvidia-hosted · commit 1a45a80 · branch fix/the-generator-calls-nvidia-hosted · worktree █████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private credential or variable name | 2 |
