# STATUS — the-reviewer-guards-hold · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 304584c47959cedd85a4f876a93d9c7fed897fbd [the-reviewer-guards-hold] absorb evidence-through-zetaris before landing
- updated: 2026-10-22T12:13:40Z

## Summary

Sample adapters cannot dial, so netguard and the test that only checked an armed client are gone. Every non-test Go file under internal/ and cmd/ is checked for banned imports, with an exempt list. The Judge drops a test that fails on the change and does not pass on the base. change.Read returns an error that names a symlink.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-reviewer-guards-hold
- state: stop-2-ready
- abstract: Sample adapters cannot dial, so netguard and the test that only checked an armed client are gone. Every non-test Go file under internal/ and cmd/ is checked for banned imports, with an exempt list. The Judge drops a test that fails on the change and does not pass on the base. change.Read returns an error that names a symlink.
- last commit: 304584c47959cedd85a4f876a93d9c7fed897fbd [the-reviewer-guards-hold] absorb evidence-through-zetaris before landing
- done: the four guards hold, and each named test fails when that guard is removed
  - import ban, mutated field the import list of internal/candidate/candidate.go, added import _ "net". go test -count=1 -run TestInternalAndCmdDoNotImportBannedPackages ./internal/pipeline exit 1, imports_test.go:81 internal/candidate/candidate.go imports net. Restored with git checkout. Tree clean after.
  - Judge base-pass, mutated field counts() in internal/agent/judge/judge.go, ignored PassesOnBase. go test -count=1 -run TestDropsATestThatDoesNotPassOnTheBase ./internal/agent/judge exit 1, Dropped false, DropReason kept. Restored with git checkout.
  - symlink, mutated field the ModeSymlink check in internal/change/read.go, removed so ReadFile follows the link. go test -count=1 -run TestReadRefusesASymlink ./internal/change exit 1, both base and change, read x; the symlink target was returned. Restored with git checkout.
  - netguard was deleted, so negative control (d) was not run. Nothing else executes Request.Command, so nothing was rerouted.
  - go build ./... && go vet ./... && go test -count=1 ./... exit 0. Build and vet printed nothing. Test summary below.
  - go run ./cmd/proven scan --sample testdata/fixture/discount exit 0. Printed F1 and F2. Did not print F3. Trace verdict for F3 is dropped: no evidence.
- next / blocked-on: none
- artifacts:
  - internal/change/read.go
  - internal/change/read_test.go
  - internal/agent/judge/judge_test.go
  - internal/pipeline/pipeline.go
  - internal/pipeline/pipeline_test.go
  - internal/pipeline/imports_test.go
  - █████████████████████████████████████
- evidence:
  - baseline before any edit, go test -count=1 ./... exit 0, at edc8dc99fc5cb6760048834e3bf8e332369142c6. Every package ok, including internal/netguard 0.002s. No package failed.
  - tip go test -count=1 ./... exit 0. ok cmd/proven 0.169s; answer 0.002s; judge 0.002s; planner 0.003s; change 0.061s; fixture 0.003s; generator 0.213s; pipeline 1.039s; query 0.002s; review 1.432s; testrun 12.142s; trace 0.024s. investigator, prover, skeptic, writer, candidate, limit, service have no test files. netguard is no longer a package.
  - tip go test -count=1 -json ./... exit 0. Test-level pass 142, fail 0, skip 1. The skip is internal/testrun TestContractW1bT10, t.Skip PROVEN_TEST_DOCKER=1. That file was not edited. remark: the plain go test summary prints ok for a package that skipped a test, so the skip count is from the json run, not from the summary lines.
  - go version go1.26.8. go.mod says go 1.26 and has no toolchain line. No Makefile and no .github workflow in this tree. The acceptance command is the one the order names. Race detector UNRUN: this diff adds no shared state and no goroutine. Testing bar: no metric, command or baseline is declared, so none was invented.
  - symlink reproduction before the fix, go run of the worktree program that points change/x at /etc/hosts, exit 0. files=1 hosts_len=570 path=x change_len=570 equals_hosts=true. The same tree failed TestReadRefusesASymlink before the check existed.
  - co-authored count, git log --format=%B edc8dc99fc5cb6760048834e3bf8e332369142c6..HEAD | grep -ci co-authored, printed 0. grep exit 1 because there was no match.
  - scan stdout. finding F1 ERROR Discount is applied twice. evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true. finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list. evidence query local-sample snapshot=sample-2026-10-22. row evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true. failed services: none. No finding F3 line.
  - trace logs/trace.jsonl F3 verdict text is dropped: no evidence. The scan created that file. It is untracked. rm is not in this role's allow list, so the file is still there and is not part of any commit.
- note: master-was-wrong: M2's second clause only. M1 held. M3 held. No session recording was in this worktree (ADR-0024). ruling: netguard deleted because sample adapters cannot dial. ruling: a symlink is an error that names the path. ruling: an exemption is one file and one import. ruling: Request.Command stays in the container runner. ruling: os.StartProcess is flagged as a selector on package os, not a dot import.
  - M1 held, read before any edit. pipeline.Build armed a netguard client in sample mode. adapters() never passed it to generator.NewSample, query.OpenSample or testrun.NewSample. review.Open on the sample path does not call generator.NewLive. Client() was only called from Arm().
  - M2 first clause held. Request.Command is executed only by internal/testrun/container.go, which runs the container runtime. Sample, Live and Off do not execute it. M2 second clause did not hold. Other exec.Command calls on this tree are git in internal/change/between.go, git in internal/change/between_test.go, git in internal/review/selection_test.go, and sh in internal/testrun/container_test.go (an exit-code helper). None of those runs Request.Command. remark: the search was exec.Command, CommandContext, StartProcess, LookPath and cmd.Run. A call through a name those patterns do not spell would not have shown.
  - M3 held. Before the fix, change/x -> /etc/hosts was returned equal to /etc/hosts, 570 bytes. The permanent test uses a file outside the fixture, not /etc/hosts, so it does not depend on that host file.
  - Session recording: none published and none handed to this leg. The only worktree file whose name matches record is fleet/decisions/ADR-0024. This spawn was not given a run id or a recording path. Publishing is not this order.
  - Cost if the netguard ruling is wrong: a sample path that dials without importing net, net/http, os/exec, syscall or plugin has no refusing client. The import walk is the control that remains.
  - Cost if the symlink ruling is wrong: a fixture that intentionally uses a symlink fails the review instead of being skipped. Read has no trace, so a skip would look like a missing file.
  - Cost if the exemption ruling is wrong: a file that must hold a second banned import fails this test until the exempt list names it. A whole-file exemption would have hidden that import.
  - Cost if the command ruling is wrong: an executor the search missed still runs a model command on the host.
  - Cost if the StartProcess ruling is wrong: import . "os" then a bare StartProcess call is not flagged. The selector os.StartProcess, including an import alias, is flagged.
- updated: 2026-10-22T11:39:30Z

CLOSE-OK · the-reviewer-guards-hold · commit 0ccb911b17662b057999211face54f8dde8223ed · branch fix/the-reviewer-guards-hold · worktree ████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
