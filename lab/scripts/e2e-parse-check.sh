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

# Comments in the shape Gitea 1.22.6 serves: the user object precedes the body,
# and < and > may be written as < and >.
sha=0123456789abcdef0123456789abcdef01234567
other=fedcba9876543210fedcba9876543210fedcba98
mk() { # mk ID LOGIN MARKER-TEXT TIME
  printf '{"id":%s,"html_url":"http://lab/proven/discount/pulls/1#issuecomment-%s","user":{"id":1,"login":"%s","login_name":"","full_name":""},"original_author":"","body":"%s\\nfinding F1","created_at":"%s"}' "$1" "$1" "$2" "$3" "$4"
}
good="<!-- proven-review run=3 sha=$sha -->"

# accept: run 3, this head, posted by the workflow's account.
c=$(mk 1 proven-bot "$good" 2026-10-24T04:31:02Z)
review_for "[$c]" 3 "$sha" proven-bot >/dev/null || { echo "e2e-parse-check: FAIL: the run's own review was refused" >&2; fail=1; }
c=$(mk 1 proven-bot "\\u003c!-- proven-review run=3 sha=$sha --\\u003e" 2026-10-24T04:31:02Z)
review_for "[$c]" 3 "$sha" proven-bot >/dev/null || { echo "e2e-parse-check: FAIL: the review with an escaped marker was refused" >&2; fail=1; }

# refuse: a marker by another account, with run and commit right.
c=$(mk 1 unrelated "$good" 2026-10-24T04:31:02Z)
out=$(review_for "[$c]" 3 "$sha" proven-bot) && { echo "e2e-parse-check: FAIL: an unrelated account's marker was accepted" >&2; fail=1; }
case $out in *"by unrelated, not by proven-bot"*) ;; *) echo "e2e-parse-check: FAIL: the ignored author was not named: '$out'" >&2; fail=1 ;; esac

# refuse: the lab admin (the account a judge signs in with) posts the exact
# marker after a successful run that posted nothing. Same run, same commit.
c=$(mk 1 proven "$good" 2026-10-24T04:31:02Z)
out=$(review_for "[$c]" 3 "$sha" proven-bot) && { echo "e2e-parse-check: FAIL: the admin's exact marker was accepted as the workflow's review" >&2; fail=1; }
case $out in *"by proven, not by proven-bot"*) ;; *) echo "e2e-parse-check: FAIL: the ignored admin was not named: '$out'" >&2; fail=1 ;; esac

# refuse: a successful run that posted nothing (no comments, or only a human's).
review_for '[]' 3 "$sha" proven-bot >/dev/null && { echo "e2e-parse-check: FAIL: no comment was accepted" >&2; fail=1; }
c=$(mk 1 proven-bot "looks good to me" 2026-10-24T04:31:02Z)
review_for "[$c]" 3 "$sha" proven-bot >/dev/null && { echo "e2e-parse-check: FAIL: a comment without the marker was accepted" >&2; fail=1; }

# refuse: an older run's review, posted in the same second as run 4's first status.
c=$(mk 1 proven-bot "<!-- proven-review run=3 sha=$sha -->" 2026-10-24T04:31:01Z)
review_for "[$c]" 4 "$sha" proven-bot >/dev/null && { echo "e2e-parse-check: FAIL: an older run's review from the same second was accepted" >&2; fail=1; }

# refuse: a review of another head commit, by the right account and run.
c=$(mk 1 proven-bot "<!-- proven-review run=3 sha=$other -->" 2026-10-24T04:31:02Z)
review_for "[$c]" 3 "$sha" proven-bot >/dev/null && { echo "e2e-parse-check: FAIL: a review of another head commit was accepted" >&2; fail=1; }

# a refused marker does not hide a later good one; the author is read per comment.
c1=$(mk 1 unrelated "$good" 2026-10-24T04:31:02Z)
c2=$(mk 2 proven-bot "$good" 2026-10-24T04:31:03Z)
review_for "[$c1,$c2]" 3 "$sha" proven-bot >/dev/null || { echo "e2e-parse-check: FAIL: the run's review after an unrelated one was refused" >&2; fail=1; }
review_for "[$c2,$c1]" 3 "$sha" proven-bot >/dev/null || { echo "e2e-parse-check: FAIL: the run's review before an unrelated one was refused" >&2; fail=1; }

[ "$fail" -eq 0 ] && echo "e2e-parse-check: ok"
exit "$fail"
