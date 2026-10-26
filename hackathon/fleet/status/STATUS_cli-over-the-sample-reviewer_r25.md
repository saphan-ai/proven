# STATUS — cli-over-the-sample-reviewer · round 25

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: c9243c18791a11684e5e74657eaf42bcdd7c268c [cli-over-the-sample-reviewer] absorb main e109e53 (ADR-0030 layout move)
- updated: 2026-10-23T13:40:30Z

## Summary

An empty boolean flag exits 7 and names true, false. The ci help names a missing head. A sample scan's connect count is 0, and one deliberate dial makes it 1.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-over-the-sample-reviewer
- state: stop-2-ready
- abstract: An empty boolean flag exits 7 and names true, false. The ci help names a missing head. A sample scan's connect count is 0, and one deliberate dial makes it 1.
- last commit: c9243c18791a11684e5e74657eaf42bcdd7c268c [cli-over-the-sample-reviewer] absorb main e109e53 (ADR-0030 layout move)
- done: Round 14 pays both reviews of round 13. Each claim names the command that proves it.
  - MASTER, after the leg (2026-10-23 ~13:40Z): 7db55dd moves the public status file to hackathon/fleet/status/ (ADR-0030) and c9243c1 absorbs main e109e53 (the layout move) into the branch. The leg's tip 26ec7b9 is unchanged inside it. On c9243c1, macm5: `go vet ./...` exit 0, `go test -count=1 ./...` exit 0, 34 packages ok, 0 FAIL, `scripts/check-links.sh` exit 0; `git diff --stat main c9243c1 -- AGENTS.md internal/review` empty.
  - Empty boolean flags exit 7. Before the fix, `go run ███████████████████████████████████████████████████████` printed EXIT 2 for `scan --sample=` (`invalid boolean value "" for -sample: strconv.ParseBool: parsing "": invalid syntax`) and the same for `scan --error=`. After the fix the same command printed EXIT 7 and `configuration refused: --sample="": value must be true or false; allowed: true, false`, and the same shape for `--error=`.
  - With the refusal check reverted to `bad != ""`, `go test -count=1 -v ./internal/cli -run 'TestRefusedBoolFlagEmptyValue|TestBadValuesAreRefused'` failed: `refusedBoolFlag() = "", "", false`, and `--sample=` / `--error=` were `code 2, want 7`. Restored. `TestBadValuesAreRefused` now has an empty-value row for every boolean flag, and `TestRefusedBoolFlagEmptyValue` calls `refusedBoolFlag`.
  - Help names the missing head. With that sentence removed, `go test -count=1 -v ./internal/cli -run 'TestHelpAndVersion$'` failed: `help missing "no head commit"`. Restored. The sentence names GITHUB_SHA and GITHUB_EVENT_PATH.
  - Sample mode opens no IP connection. The instrument is the child's connect syscall, not http.DefaultTransport. `go test -count=1 -v ./internal/cli -run TestSampleModeOpensNoConnection` passed (network 0). With one `net.Dial("tcp", "127.0.0.1:17991")` in the sample path, and the test listening there, the same test failed: `exit 0 network 1 unix 0 other 0 unread 0`. The dial was removed. A probe before that test printed `quiet connects=0`, `dial connects=1`, and a custom `http.Transport` `connects=1`.
  - `git log --oneline c57abe4..HEAD` is 4f5276d, 1529571, 26ec7b9. `git diff --stat 90d6951..HEAD -- AGENTS.md internal/review` printed nothing.
  - `go vet ./...` exit 0. `gofmt -l` on the five edited files printed nothing. `go test -count=1 ./...` exit 0: 31 packages ok, 8 with no test files, 0 FAIL. `go test -count=1 -race ./internal/cli` ok. Go is go1.26.8 linux/amd64; go.mod says `go 1.26` and has no toolchain line.
- next / blocked-on: none
- updated: 2026-10-23T12:48:41Z
- artifacts:
  - internal/cli/config.go
  - internal/cli/config_test.go
  - internal/cli/connect_linux_test.go
  - internal/cli/run.go
  - internal/cli/run_test.go
- evidence:
  - Baseline before edits: `go test -count=1 ./internal/cli` ok, 11.232s, on c57abe40036b1ef03372060ea223f9072ee6eb53.
  - Reproduction through cli.Run, before the fix: scan --sample= and scan --error= each exited 2 with the flag package's one-dash ParseBool text.
  - Reverted-fix red, then restored: empty --sample and --error exit 2; refusedBoolFlag returns false for Set("").
  - Help sentence removed, then restored: TestHelpAndVersion fails on a missing "no head commit".
  - Network seam connected, then removed: TestSampleModeOpensNoConnection reports network 1, then passes with the dial gone.
  - No Makefile and no .github workflow. The acceptance commands are go vet and go test, above.
  - testing-bar: no metric, command, or baseline is declared. Nothing to ratchet. No bar-weakening edit. The checks run are project checks (go vet, gofmt) and a suite check (go test).
  - Per-test pass, fail, and skip counts UNRUN. `go test` without -v does not print them. No package printed FAIL.
- note: master-was-wrong: —. ruling: the instrument is the connect syscall of a child — CLONE_NEWUSER returned permission denied and CLONE_NEWNET returned operation not permitted, and this count is 0 with no dial and 1 for net.Dial and for an http.Transport that is not DefaultTransport — cost if wrong: a send that never calls connect is invisible. ruling: an AF_UNIX connect is not a failure — it is not a network connection, and the sample scan's unix count was 0 — cost if wrong: a local socket would pass. ruling: the test is linux/amd64 only — the syscall number and the ioctl sizes are that architecture — cost if wrong: another OS does not compile this test.

```owner-block-v1
version: 1
stream: cli-over-the-sample-reviewer
elements:
  - kind: none
    abstract: Nothing on your hand. This round is master-verify-only.
```

CLOSE-OK · cli-over-the-sample-reviewer · commit 26ec7b92f2e6d46432d8a8375f55f62d184f8416 · branch fix/cli-over-the-sample-reviewer · worktree ████████████████████████████████████████████████████████████████

remark: the connect count sees the connect syscall on linux/amd64, including a dial that does not use http.DefaultTransport; a sendto that never calls connect is not seen, and on that path the count stays 0 whether or not bytes are sent.
remark: an AF_UNIX connect is counted apart and was 0 for the sample scan; a failed read of the peer address fails the test (unread) rather than reporting 0.
remark: CLONE_NEWUSER was permission denied and CLONE_NEWNET was operation not permitted on this seat, so the instrument is not a private network namespace.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
