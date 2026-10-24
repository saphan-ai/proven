#!/bin/sh
# Offline check of the e2e parsers against statuses Gitea 1.22.6 served on a
# live lab (23 Oct 2026). No Docker, no network. Run: make check-e2e-parse
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
. "$here/e2e-parse.sh"
ROOT=$(CDPATH= cd -- "$here/../.." && pwd)

fail=0
expect() {
  if [ "$2" != "$3" ]; then
    echo "e2e-parse-check: FAIL: $1: got '$2', want '$3'" >&2
    fail=1
  fi
}

ctx=$(workflow_context "$ROOT/lab/workflow/proven.yml")
expect "workflow context" "$ctx" "proven / review (pull_request)"

# Real shape: run 2's success (id 8) is newer by id than run 3's pending (id 7).
statuses='[{"id":8,"status":"success","target_url":"/proven/discount/actions/runs/2/jobs/0","description":"Successful in 2s","context":"proven / review (pull_request)","created_at":"2026-10-23T21:14:22Z","creator":{"id":1,"login":"lab"}},{"id":7,"status":"pending","target_url":"/proven/discount/actions/runs/3/jobs/0","description":"Has started running","context":"proven / review (pull_request)","created_at":"2026-10-23T21:14:21Z","creator":{"id":1,"login":"lab"}},{"id":4,"status":"success","target_url":"/proven/discount/actions/runs/1/jobs/0","description":"Successful in 2s","context":"proven / review (pull_request)","created_at":"2026-10-23T21:14:19Z","creator":{"id":1,"login":"lab"}}]'
tab=$(printf '\t')
expect "highest run is pending" "$(newest_run "$statuses" "$ctx")" "3${tab}pending${tab}2026-10-23T21:14:21Z"
expect "other context ignored" "$(newest_run "$statuses" "other / job (push)")" ""

done3='[{"id":9,"status":"failure","target_url":"/p/d/actions/runs/3/jobs/0","context":"proven / review (pull_request)","created_at":"2026-10-23T21:15:00Z"},{"id":7,"status":"pending","target_url":"/p/d/actions/runs/3/jobs/0","context":"proven / review (pull_request)","created_at":"2026-10-23T21:14:21Z"}]'
expect "run 3 failed, start is its pending" "$(newest_run "$done3" "$ctx")" "3${tab}failure${tab}2026-10-23T21:14:21Z"
expect "no statuses" "$(newest_run '[]' "$ctx")" ""

comments='[{"id":1,"body":"<!-- proven-review -->\nfinding F1","created_at":"2026-10-23T21:14:20Z"}]'
if review_since "$comments" "2026-10-23T21:14:21Z"; then
  echo "e2e-parse-check: FAIL: a comment older than the run counted" >&2
  fail=1
fi
review_since "$comments" "2026-10-23T21:14:20Z" || { echo "e2e-parse-check: FAIL: a comment at the run start did not count" >&2; fail=1; }

[ "$fail" -eq 0 ] && echo "e2e-parse-check: ok"
exit "$fail"
