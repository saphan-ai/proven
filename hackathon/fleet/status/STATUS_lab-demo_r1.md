# STATUS — lab-demo · round 1

- state: ready-for-review
- executor: a Claude Code agent (an executor seat; not a Cursor agent, as the order named the backend)
- model: Sonnet 5.5 (`claude-sonnet-5-5`)
- last commit: see the branch tip; the commits are `8e1ac33`, `6c6a98a`, `03c8c66`, `0e6fc99`
- updated: 2026-10-25 UTC

## Summary
`make up` now ends with the owner's login block and prints nothing after it. The lab seeds seven pull requests from the sample sets (two clean, five that each show a different class of defect), a start page on port 3002, and the documentation site on port 3003. The dashboard shows 24 hours and has a value in every panel for those reviews, on paper. `make up` itself was not run: the seat has no Docker.

## Acceptance, item by item
1. **The banner, exactly the owner's block, first line "Start here: …", nothing after it, URLs from settings.** `sh lab/scripts/banner-check.sh --no-tty-ok` exit 0. It compares the redirected banner with the owner's block line for line, changes eight settings and requires each line to follow, and requires nothing after the Grafana line. **The terminal case was not run:** `script(1)` and Python's `pty` both refused to make a pseudo-terminal on this seat. Negative control: changing `Username:` to `User:` in `banner.sh` turned the check red; reverted.
2. **No password in any log.** The redirected banner carries the file's place and the check requires the test password to be absent. `go test ./lab` renders the page offline with a password of `p&ss|w\rd<b>"q"` and requires: the password is HTML-escaped inside the page's one `<code id="p">`, it is not in the command's output, and it is in no committed file of the lab.
3. **Seven demo pull requests from existing sample sets, distinct classes, two clean.** `go test ./lab` (exit 0): every row of `lab/seed/prs.tsv` names a sample set with `base`, `change` and `fixture.json`, the image's Dockerfile copies it, the branch is `demo-<set>`, and no two rows share a class. A second test breaks six fields one at a time and requires the check to name each. One sample set is new: `testdata/fixture/discount-clean`, the correct version of the discount change, with an empty Planner answer; its prompt hash is the one the sample reader named.
4. **Clean pull requests are reviewed too.** The workflow adds one line, `clean: Proven kept no finding on this change.`, above a review that keeps none; the review lines below it name the reviewers and the failed services. Simulated here with the real binary and the workflow's own commands (a throwaway script, not committed): the clean sets give a comment with the clean line and no finding line, the defective sets the reverse, and each comment is valid JSON. **Not run in a real lab.**
5. **Telemetry arrives.** Measured here: one sample review of each fixture, with `OTEL_EXPORTER_OTLP_ENDPOINT` set to a local recording listener, posts to `/v1/traces` and `/v1/metrics` and exits 0. So the binary exports; why the owner's lab did not show it was **not found**. The lab seeded the new workflow into a repository whose pull request was already open and did not review it again. The seed now commits the current workflow to every branch, which starts that review. The nested container's name lookup of `lgtm` from the inner daemon is unmeasured.
6. **The dashboard has data in every panel.** `go test ./internal/telemetry` (exit 0) holds the dashboard to the instruments the code makes. Changes: 24 hours instead of 30 minutes; panel 1 asks for a time series by name (`range`, `format: time_series`), a fix for "Data does not have a time field" that is **unverified** because Grafana was not run; the two panels that stay empty when nothing differs or nothing was dropped show a `none` row with zero. `proven.reviewer.disagreements` is never emitted in sample mode (measured on five fixtures with the lab's `proven.yaml`), so that panel's zero is the true count.
7. **Docs.** `docs/LAB.md` steps 2–6 and a new demo section, `docs/CONFIGURATION.md`, ADR-0026 point 7 and ADR-0017 point 5 amended in place, `docs/SAMPLE_MODE.md` for the new sample set.
8. **The start page and the documentation service.** `lab/page/lab.html.tmpl` is the owner's design with placeholders; the seed renders it into a volume; nginx (`nginxinc/nginx-unprivileged`, read-only filesystem, all capabilities dropped, no new privileges, `127.0.0.1` only) serves it at `/lab`. The documentation image is `docs-site/Dockerfile` as a service on port 3003, also loopback only. `go test ./lab` checks that every published port binds to `LAB_BIND`, the template's placeholders are all filled, every link has `target="_blank" rel="noopener"`, and the template names no address.

## Evidence
- `go vet ./...` exit 0. `go test ./...`: every package `ok` except `internal/cli/sarifschema` `TestSampleSARIFMatchesTheSchema`, which fails with "mkdir /…/.cache/proven: permission denied" (the seat cannot write its cache directory; not touched by this change).
- `sh -n` on all 29 lab scripts: exit 0. `gofmt -l lab`: nothing.
- `scripts/check-links.sh`: **exit 1**, two links in `docs-site/generated/settings.md` (`OSV_TABLE.md`, `POLICY.md`). That file is generated, tracked and untouched by this change. The check was not run on a clean base, so "already red before" rests on the file being untouched.
- The seed's own rewrite of the workflow was run on the real file: the `docker run` line carries the OTLP variables, the fixture variable and the second reviewer, and the post goes to the lab's address.
- Fifteen sample sets were run; the clean ones are `unknown-author` and the new `discount-clean`. `refuted` and `unsafe-prompt` print no finding too, but the first is a botched change whose claim the Skeptic refuted and the second stops at the guardrail, so neither is shown as "clean".

## Not run
- `make up`, `make clean && make up`: no Docker on the seat.
- The nginx image tag and the documentation image build: never pulled or built.
- The dashboard in Grafana, `make demo-check` (the live half), `make trace`.
- The terminal case of `banner-check.sh`.
- The documentation site's own tests (`npm test`) over the changed pages.

## Blocked on / next
Nothing blocks. Next: the master runs `make clean && make up` and `make demo-check` on a Docker host and looks at panel 1.

## REDACTED
Nothing was blacked out: no credentials, host names, user names, absolute paths or private-record names appear above (0 items).
