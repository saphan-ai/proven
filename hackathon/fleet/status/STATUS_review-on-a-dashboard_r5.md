# STATUS — review-on-a-dashboard · round 5

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 9 min
- last commit: 7229e9c7cd9803bac08983c710a8e90364ebc04c [review-on-a-dashboard] merge: absorb main b17d30f
- updated: 2026-10-25T05:07:39Z

## Summary

Round 3. Every error text that becomes a span status now goes through the review's redactor (root and agent statuses). A failed-call control was red at 140c8f4, green at the tip, and red again when the fix is unplugged.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — review-on-a-dashboard
- state: stop-2-ready
- abstract: Round 3. Every error text that becomes a span status now goes through the review's redactor (root and agent statuses). A failed-call control was red at 140c8f4, green at the tip, and red again when the fix is unplugged.
  The two OWED items are done too: an on/off result parity test, and the config doc names the invalid-setting refusal. go vet exits 0; go test ./... has one red I did not cause (pipeline TestMaterializedTreesAreReadableByOthers, mode 664).
  Docker, make and the dashboard were not run.
- last commit: 7229e9c7cd9803bac08983c710a8e90364ebc04c [review-on-a-dashboard] merge: absorb main b17d30f
- done: BLOCKING finding ADDRESSED, both OWED ADDRESSED
  - absorb (master, 2026-10-25): fa28eb3 merges main 41c1194 (the-image landed); one conflict, the Makefile .PHONY line, resolved to keep both image-check and trace; go vet and go test ./... ok, make -n image-check trace ok, check-links ok
  - history rebuilt (master, 2026-10-25): each absorb takes main's file as main had it; the two telemetry edits that rode inside absorbs are their own commits (test: sample fix scenario reads OTEL_*; pipeline: begin feeds the trace to telemetry); the final tree is identical to fa28eb3
  - absorb (master, 2026-10-25): 7229e9c merges main b17d30f; two import conflicts (pipeline.go, review.go) resolved, the unused query import dropped; go vet and go test ./... ok
  - master, 2026-10-25: after the absorb the sample review outlasts 180 s; c540b3b sets E2E_WAIT_SECONDS=600; at c540b3b on runner-cursor: up 0, e2e 0, trace 0, failed-run 0, runner stopped 2, down/clean/check-links 0
  - master, 2026-10-25: absorbed main db9f7c2 at 67313d9: main opens the trace in begin() after its outside-the-target check, so the trace redactor and span-event observer attach there; open.go keeps both sides; settingspage imports merged; build, vet, all tests, check-links green
  - `T.SetRedactor` and `T.errorText` in internal/telemetry/telemetry.go; Review.End and Stage.End use errorText; pipeline.Build hands its redactor to the telemetry.
  - Red at 140c8f4: `go test ./internal/review -run TestAFailed` printed "span agent planner status description holds the secret" for both new tests (capture in ██████████████████████████████████████████). Green at the tip.
  - Unplug control: with the SetRedactor line in pipeline.Build commented out, both tests fail again; restored before commit.
  - TestAFailedReviewKeepsTheSecretOutOfTheExportedStatus: plain error from the planner, review fails; root and planner status hold the marker, no secret in any name, status, attribute or event.
  - TestAFailedServiceKeepsTheSecretOutOfTheAgentStatus: service.Down reason with the secret, review degraded; planner status holds the marker.
  - TestTelemetryDoesNotChangeWhatAReviewDecides: whole Result equal with exporter off and on, TracePath cleared.
  - docs/CONFIGURATION.md regenerated from internal/settingspage/catalog.go with `go run ./cmd/settings-doc`.
- next / blocked-on: none
- artifacts:
  - internal/telemetry/telemetry.go
  - internal/pipeline/pipeline.go
  - internal/review/telemetry_status_test.go
  - internal/settingspage/catalog.go
  - docs/CONFIGURATION.md
- evidence:
  - Paths where error text becomes telemetry: root status (Review.End) and agent status (Stage.End) were raw, both fixed. The model call status is the fixed string "model call failed" and its error.type is the fixed "generator". Span events are trace-file events, already redacted by the writer. No span uses RecordError. No other path found.
  - `go vet ./...` exit 0.
  - `go test ./internal/review ./internal/telemetry ./internal/settingspage -count=1` exit 0.
  - `go test ./...`: one FAIL, internal/pipeline TestMaterializedTreesAreReadableByOthers (temp go.mod mode 664). The reviewer saw the same failure; my pipeline change is one line in Build. I did not baseline it separately.
  - Attribution: `git log --format='%an <%ae>' 140c8f4..HEAD` prints Marcin Marzec only; 0 Co-authored-by trailers; no merge commit. No local main ref here, so the check ran against 140c8f4.
  - Not run: make up, make e2e, make trace, make down, make clean, make check-links, dashboard panels (no Docker on this seat; the master runs them). Race detector not run.
- note: master-was-wrong: — (the reviewer's claim held, lines 219 and 282 set the raw status; the model-call span already used a fixed text, so only two sites leaked). NOT ADDRESSED: none. I redacted the descriptions rather than replacing them with codes, which keeps the text useful on the timeline.
- updated: 2026-10-24T15:05:00Z

CLOSE-OK · review-on-a-dashboard · commit 4d50977ee605164841f31037909bcea5566103a7 · branch fix/review-on-a-dashboard · worktree █████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
