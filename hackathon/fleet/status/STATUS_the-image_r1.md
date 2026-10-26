# STATUS — the-image · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: b4769f8cba1b3a8f9b7191fb4a71061e49d6926c [the-image] scripts: check the image against the binary
- updated: 2026-10-24T18:36:43Z

## Summary

Proven's image prints the same version and the same sample review as the binary from its commit; scripts/image-check.sh and make image-check prove it and print the size (97 MB). The seat denied docker, so the master did the round on runner-cursor; the Dockerfile needed no change.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-image
- state: stop-2-ready
- abstract: Proven's image prints the same version and the same sample review as the binary from its commit; scripts/image-check.sh and make image-check prove it and print the size (97 MB). The seat denied docker, so the master did the round on runner-cursor; the Dockerfile needed no change.
- last commit: b4769f8cba1b3a8f9b7191fb4a71061e49d6926c [the-image] scripts: check the image against the binary
- next / blocked-on: named review (codex + claude-code + an extra Nemotron review)
- note: master-was-wrong: —. One call of docker info was refused before the program started (permission class █████████████████████████████████████, allowed form none, tool-command-not-in-class). Did not hunt. HEAD is 2b8fd4a34666b6fe5406bf7e1af21932251f2582, the order base, and this leg made no commit. testing-bar: no metric is declared in the order, so the bar is UNRUN and no number was invented. Image build, four-way comparison, scripts/image-check.sh, Makefile image-check, docs section, negative control, go vet and go test are UNRUN because the order stops at this refusal. remark: the refusal names the permission class, not the docker daemon, so it cannot say whether the engine is up. remark: updated is the session clock 2026-10-24T16:12:00Z because date is denied on this seat, so the stamp is not a measured clock. · master-round 2026-10-25 on runner-cursor (docker 29.9.1, go 1.26.10): image from a clean git archive vs the binary — same version line, same sample review (trace and fixture paths normalised), size 97073149 bytes; make image-check in a git checkout ok; negative control: only the image\'s fixture changed by one line → FAIL "the image\'s sample review exited 2". A first run from a macOS tar was falsely green (AppleDouble ._ files broke both reviews identically); the script now requires both reviews to succeed.
- updated: 2026-10-25T02:30:00Z
- artifacts:
  - SHOUT_docker-denied.md
- evidence:
  - SHOUT_docker-denied.md

## REDACTED

| what | times |
|---|---|
| policy binding name | 1 |
