# STATUS — corpus-pv-016 · round 2

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 3 min
- last commit: ffd3e35c8a96f4b31a0a12b08290b750ea8ab8ba test/corpus: name the corpus by its path after the ADR-0030 move (master)
- updated: 2026-10-24T02:52:26Z

## Summary

Codex GPT-6; base and change both printed backup_failed delivered=0, while the change page promises delivery. Quote drift now exits 1; macOS remains unrun.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-016
- state: stop-2-ready
- abstract: Codex GPT-6; base and change both printed backup_failed delivered=0, while the change page promises delivery. Quote drift now exits 1; macOS remains unrun.
- last commit: ffd3e35c8a96f4b31a0a12b08290b750ea8ab8ba test/corpus: name the corpus by its path after the ADR-0030 move (master)
- done: C1, C2 and C3 addressed; the case and manifest record the Linux run.
  - master, 2026-10-24: the branch was carried onto proven's rewritten main (no AI trailer, no seat author; case files identical) and the master replaced the retired tests/corpus path with test/corpus in the case (ffd3e35); check.sh re-run on macm5 matches CASE.md §8; one codex review of this tip accepted (owner: one review where one settles it)
  - bash tests/corpus/cases/PV-016/check.sh printed 24 claims and exited 0.
  - A one-quote mutation in change/docs/alerts/chat-setup.md exited 0 before the fix and 1 after it; the page was restored.
  - git diff --stat fe146e0..HEAD lists only PV-016 outside expected/ and tests/corpus/manifest.json.
- next / blocked-on: Run bash tests/corpus/cases/PV-016/check.sh on macOS with Go 1.26 and record its output and exit code before claiming cross-platform execution.
- artifacts:
  - tests/corpus/cases/PV-016/CASE.md
  - tests/corpus/cases/PV-016/check.sh
  - tests/corpus/manifest.json
- evidence:
  - go version printed go1.26.5 linux/amd64.
  - git diff --check exited 0.
  - git log -1 reported 02cdcfff9d896298a11216fb838c7eae0ff5dea0.
  - macOS check: not run; this seat runs Linux.
- note: Master-was-wrong: the round 1 instruction to make manifest status a string conflicted with the object used by every other case; restored the PV-016 object with measured dates. No expected/ file was read. The run did not expose a routed model identifier beyond the session identity above.
- updated: 2026-10-23T01:51:37Z

## REDACTED

none
