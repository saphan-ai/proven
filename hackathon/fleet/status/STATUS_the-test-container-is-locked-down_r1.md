# STATUS — the-test-container-is-locked-down · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 07b901d [the-test-container-is-locked-down] testrun: lock down container, env
- updated: 2026-10-23T11:49:12Z

## Summary

The container runs with --user 65534:65534, --pids-limit 256, --cap-drop ALL, no-new-privileges, --read-only and bounded tmpfs scratch; limits come from PROVEN_TEST_USER, PROVEN_TEST_PIDS and PROVEN_TEST_TMP_SIZE (docs regenerated). The runtime process and every git call get an explicit minimal environment (new internal/childenv); both negative controls went red and were restored. UNRUN: a real container. Full go test is red only where tests bind a local listener, which this sandbox refuses.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-test-container-is-locked-down
- state: stop-2-ready
- abstract: The container runs with --user 65534:65534, --pids-limit 256, --cap-drop ALL, no-new-privileges, --read-only and bounded tmpfs scratch; limits come from PROVEN_TEST_USER, PROVEN_TEST_PIDS and PROVEN_TEST_TMP_SIZE (docs regenerated). The runtime process and every git call get an explicit minimal environment (new internal/childenv); both negative controls went red and were restored. UNRUN: a real container. Full go test is red only where tests bind a local listener, which this sandbox refuses.
- last commit: 07b901d [the-test-container-is-locked-down] testrun: lock down container, env
- done: container hardening flags, minimal child environment and three settings, each proven by a test
  - argument test: internal/testrun/container_hardening_test.go TestArgvHardensTheContainer and TestArgvScratchIsBoundedTmpfs, red at base (`██████████████████████████████████████████████████████████████`), green after
  - environment tests: TestRuntimeProcessEnvironmentOmitsSecrets (testrun) and TestGitEnvironmentOmitsSecrets (change), red at base, green after
  - settings: TestLoadHardeningSettingsAreReadAndRefused refuses root, 0, -1 and a zero scratch size by variable name
  - negative controls (law #27): removing --cap-drop turned TestArgvHardensTheContainer red; os.Environ() in commandOutput turned TestRuntimeProcessEnvironmentOmitsSecrets red; os.Environ() appended in gitCommand turned TestGitEnvironmentOmitsSecrets red; all restored (`negative-controls.txt` beside the red-at-base file)
- next / blocked-on: none
- artifacts:
  - internal/childenv/childenv.go
  - internal/testrun/container.go
  - internal/testrun/settings.go
  - internal/change/between.go
  - docs/CONFIGURATION.md
- note: master-was-wrong: the order places the git calls at container.go ~432; they are in internal/change/between.go, and container.go held only the runtime call. I fixed both through the shared internal/childenv package. Commands run: go build ./... exit 0; go vet ./... exit 0; gofmt -l internal cmd exit 0, no output; go run ./cmd/settings-doc exit 0; go test -count=1 on testrun, change, childenv, settingspage and settings-doc exit 0, all ok. Full go test -count=1 ./... exit 1 (`full.txt`): FAIL in cmd/livecheck, generator, guardrail, pipeline, query and review, each a panic "httptest: failed to listen on a port: bind: operation not permitted", the sandbox refusing a listener; those packages are untouched here and tests after the first panic in each did not run (UNRUN, named). The existing contract test TestContractW1bT3 took the --tmpfs value for a host path, so I made its helper skip a --tmpfs value, which is a path inside the container. The scratch mounts /tmp and /.cache are my reading of what a Go test needs under an identity with no home, UNRUN against a real container; a master with a runtime should confirm one sample test passes. The runtime inherits only PATH, HOME, DOCKER_HOST, DOCKER_CONTEXT, DOCKER_CONFIG and XDG_RUNTIME_DIR. No session recording was made on this headless leg. The commit has no Co-authored-by line. A packed-refs.lock permission error followed the commit, which landed.
- updated: 2026-10-23T12:30:00Z

CLOSE-OK · the-test-container-is-locked-down · commit 07b901d · branch fix/the-test-container-is-locked-down · worktree █████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
