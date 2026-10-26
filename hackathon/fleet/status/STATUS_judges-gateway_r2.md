# STATUS — judges-gateway · round 2

- state: ready-for-review
- executor: claude-code
- model: `claude-sonnet-5-5`
- last commit: the tip of branch `fix/judges-gateway` (`[judges-gateway] docs: ADR-0051 and the runbook for log and reply bounds`)
- updated: 2026-10-26, UTC

## Summary
Round 2 repairs the two BLOCKING findings of the review of round 1. A call the gateway cannot log is now observable and
stops the gateway from serving more calls until the sink works. The model answer is bounded in bytes.

## The two findings
1. **A failed call-log or alarm write was silent.** Decision (b) of the order, recorded as ADR-0051 decision 9: the
   gateway serves the call that finds the sink broken (the line holds the status and bytes, so it cannot be written
   first), holds the line, writes one `JUDGES-GATEWAY-FAULT` line on standard error, turns `GET /healthz` to 503, and
   refuses every later call with `503 log_unavailable` before any upstream is called. The next call or probe writes the
   held lines; one `JUDGES-GATEWAY-RECOVERED` line states how many were lost. The alarm sink has the same treatment.
2. **The model answer had no byte bound.** New setting `JUDGES_GATEWAY_MAX_RESPONSE_BYTES` (default 4 MiB, allowed 1 KiB
   to 64 MiB). A larger declared `Content-Length` is refused with `502 response_over_limit` before any byte. An
   undeclared stream is relayed up to the bound and then the connection is aborted, so the judge's read fails; the log
   line has status 200, `bytes` as relayed, `refused: response_over_limit` and a reason. ADR-0051 decision 10 and the
   runbook say what the owner sees and does.

## Controls (each names the field it mutates)
- A log writer that returns an error: no silent 200 (fault line, probe 503, next call `log_unavailable`, upstream count
  unchanged), recovery delivers the held line. Mutation check: write error ignored in the sink → `TestFailedLogWriteIsNeverSilent`
  and `TestFailedAlarmWriteIsNeverSilent` failed; restored.
- An alarm writer that returns an error: the same, at the 80% call; the held alarm is delivered once on recovery.
- An answer over the cap, with and without `Content-Length`: the judge receives at most the cap, the end is a read error
  or a named refusal, the log line says so. Mutation check: bound removed → both tests failed; restored.
- An answer of exactly the cap, with and without `Content-Length`, is relayed whole.
- Every round-1 control passes: `go test -race -count=3 ./internal/gateway ./cmd/judges-gateway` → ok.
- `go vet ./...` → exit 0; `gofmt -l cmd internal` → empty.

## Not run
- `scripts/check-links.sh` exits 1 for the same two links in `docs-site/generated/settings.md` as in round 1; no link
  added here is broken.
- If standard error also fails, nothing can name the fault; only the probe says it. No test closes standard error.
- A partial write (a write that stores some bytes and then fails) is not handled: the whole line is retried.
- Nothing on Cloudflare, no real Nightshift or Zetaris.
