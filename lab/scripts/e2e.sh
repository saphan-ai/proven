#!/bin/sh
# Start the lab if it is down, then wait for the pull request's CURRENT head
# commit to have a workflow run, and require that run to have succeeded and to
# have posted proven's review. An older review comment, left on a kept volume,
# is not proof: it must be created at or after the run's start.
# If the lab is up and the runner is not, exit 1 and say so. Do not start the runner then.
#
# Gitea 1.22.6 (lab/versions.env) serves GET /repos/{owner}/{repo}/actions/tasks.
# Each entry is one workflow run: head_sha, run_number, status, run_started_at.
# Gitea 1.22.6 has no /actions/runs endpoint; that one arrives in a later release.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
. "$here/lib.sh"
. "$here/gitea-api.sh"
need_docker

if ! service_running gitea; then
  sh "$ROOT/lab/scripts/up.sh"
fi

if ! service_running runner; then
  echo "e2e: runner is not running, so the workflow cannot post a review" >&2
  exit 1
fi

if ! compose exec -T runner sh -c 'test -s /data/.runner'; then
  echo "e2e: runner is not registered with Gitea, so the workflow cannot post a review" >&2
  exit 1
fi

gitea_login
base=$(gitea_repo_api)

# newest_run TASKS_JSON SHA — print "run_number<TAB>status<TAB>started_at" of the
# run with the highest run_number among those whose head_sha is SHA.
newest_run() {
  printf '%s' "$1" | awk -v sha="$2" '
    function str(obj, key) {
      if (match(obj, "\"" key "\":\"[^\"]*\"")) {
        return substr(obj, RSTART + length(key) + 4, RLENGTH - length(key) - 5)
      }
      return ""
    }
    function num(obj, key) {
      if (match(obj, "\"" key "\":[0-9]+")) {
        return substr(obj, RSTART + length(key) + 3, RLENGTH - length(key) - 3) + 0
      }
      return -1
    }
    { buf = buf $0 }
    END {
      n = split(buf, objs, /\},\{/)
      best = -1
      for (i = 1; i <= n; i++) {
        if (str(objs[i], "head_sha") != sha) continue
        r = num(objs[i], "run_number")
        if (r > best) {
          best = r
          status = str(objs[i], "status")
          started = str(objs[i], "run_started_at")
          # A run that has not started carries the zero time. Its creation time is the bound then.
          if (started == "" || substr(started, 1, 4) == "0001") started = str(objs[i], "created_at")
        }
      }
      if (best >= 0) printf "%d\t%s\t%s\n", best, status, started
    }'
}

# review_since COMMENTS_JSON SINCE — exit 0 when a comment that starts with
# proven's marker was created at or after SINCE. The marker may be JSON-escaped.
review_since() {
  printf '%s' "$1" | awk -v since="$2" '
    { buf = buf $0 }
    END {
      rest = buf
      while (match(rest, /"body":"(<|\\u003c)!-- proven-review --/)) {
        rest = substr(rest, RSTART + RLENGTH)
        if (match(rest, /"created_at":"[^"]*"/)) {
          ts = substr(rest, RSTART + 14, RLENGTH - 15)
          if ((ts "") >= (since "")) exit 0
        }
      }
      exit 1
    }'
}

utc() {
  case $1 in
    ????-??-??T??:??:??Z) return 0 ;;
    *) return 1 ;;
  esac
}

deadline=$(($(date +%s) + E2E_WAIT_SECONDS))
last="no workflow run has started for the pull request's head commit"
while [ "$(date +%s)" -lt "$deadline" ]; do
  pulls=$(gitea_curl "$base/pulls?state=open" 2>/dev/null || true)
  num=$(printf '%s' "$pulls" | sed -n 's/.*"number":\([0-9][0-9]*\).*/\1/p' | head -n 1)
  sha=
  if [ -n "$num" ]; then
    pr=$(gitea_curl "$base/pulls/$num" 2>/dev/null || true)
    sha=$(printf '%s' "$pr" | sed -n 's/.*"head":{[^}]*"sha":"\([0-9a-f]\{40\}\)".*/\1/p' | head -n 1)
  fi
  if [ -n "$sha" ]; then
    tasks=$(gitea_curl "$base/actions/tasks?limit=50" 2>/dev/null || true)
    run=$(newest_run "$tasks" "$sha")
    if [ -n "$run" ]; then
      run_no=$(printf '%s' "$run" | cut -f 1)
      status=$(printf '%s' "$run" | cut -f 2)
      started=$(printf '%s' "$run" | cut -f 3)
      last="workflow run #$run_no for head commit $sha is $status"
      case $status in
        failure | cancelled | skipped)
          echo "e2e: workflow run #$run_no for pull request #$num (head commit $sha) finished with status $status" >&2
          echo "e2e: an older proven review comment, if one is there, is not a review of this commit" >&2
          echo "e2e: read the run in the lab's Actions tab, or run: make logs SERVICE=runner" >&2
          exit 1
          ;;
        success)
          if ! utc "$started"; then
            echo "e2e: run #$run_no started at '$started', which is not a UTC time; cannot order it against the comment" >&2
            exit 1
          fi
          comments=$(gitea_curl "$base/issues/$num/comments" 2>/dev/null || true)
          if review_since "$comments" "$started"; then
            echo "e2e: proven review is on the pull request (workflow run #$run_no succeeded; head commit $sha)"
            printf '%s\n' "$comments"
            exit 0
          fi
          last="workflow run #$run_no succeeded, but no proven review comment was created at or after $started"
          ;;
      esac
    fi
  fi
  sleep "$E2E_POLL_SECONDS"
done

echo "e2e: no proven review from the current run after ${E2E_WAIT_SECONDS}s: $last" >&2
echo "e2e: open the pull request's Actions tab in the lab, or run: make logs SERVICE=runner" >&2
exit 1
