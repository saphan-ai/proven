#!/bin/sh
# Start the lab if it is down, then wait for the pull request's CURRENT head
# commit to have a workflow run, and require that run to have succeeded and to
# have posted proven's review. An older review comment, left on a kept volume,
# is not proof: the comment's marker must name this run's number and this head
# commit, and the comment must be by the workflow's own account (LAB_BOT_USER), never the admin
# a judge signs in with: a person's comment cannot pass for the workflow's.
# If the lab is up and the runner is not, exit 1 and say so. Do not start the runner then.
#
# Gitea 1.22.6 (lab/versions.env) has no Actions runs, tasks, jobs or workflows
# route (measured: HTTP 404). It serves the head commit's statuses:
# GET /repos/{o}/{r}/commits/{sha}/statuses. The workflow's job posts one status
# per change, with the run number in target_url (/actions/runs/N/). The seed can
# start more than one run on one head, so the run is the highest N, not the
# newest id. The parsers are in e2e-parse.sh and are checked offline by
# make check-e2e-parse.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
. "$here/lib.sh"
. "$here/gitea-api.sh"
. "$here/e2e-parse.sh"
need_docker

if ! service_running gitea; then
  # No question at a terminal: this check is run by a script. A setting (in the environment or lab/.env) still decides the mode.
  load_lab_env || exit 1
  PROVEN_LAB_MODE=${PROVEN_LAB_MODE:-sample}
  export PROVEN_LAB_MODE
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

# The context Gitea gives the statuses of the workflow's job, read from the
# workflow the lab seeds, not typed a second time.
context=$(workflow_context "$ROOT/lab/workflow/proven.yml")
if [ -z "$context" ]; then
  die "e2e: could not read the workflow name and job from lab/workflow/proven.yml"
fi

deadline=$(($(date +%s) + E2E_WAIT_SECONDS))
last="no workflow run has started for the pull request's head commit"
while [ "$(date +%s)" -lt "$deadline" ]; do
  # The pull request of $LAB_DEMO_HEAD, the branch e2e-failed-run breaks; the lab opens seven, so never "the first".
  num=$(gitea_curl "$base/pulls/$LAB_DEMO_BASE/$LAB_DEMO_HEAD" 2>/dev/null | grep -o '"number":[0-9]*' | head -n 1 | cut -d: -f2 || true)
  sha=
  if [ -n "$num" ]; then
    pr=$(gitea_curl "$base/pulls/$num" 2>/dev/null || true)
    sha=$(printf '%s' "$pr" | sed -n 's/.*"head":{[^}]*"sha":"\([0-9a-f]\{40\}\)".*/\1/p' | head -n 1)
  fi
  if [ -n "$sha" ]; then
    statuses=$(gitea_curl "$base/commits/$sha/statuses?limit=50" 2>/dev/null || true)
    run=$(newest_run "$statuses" "$context")
    if [ -z "$run" ]; then
      last="no status '$context' on head commit $sha yet"
    else
      run_no=$(printf '%s' "$run" | cut -f 1)
      status=$(printf '%s' "$run" | cut -f 2)
      last="workflow run #$run_no for head commit $sha is $status"
      case $status in
        failure | error)
          echo "e2e: workflow run #$run_no for pull request #$num (head commit $sha) finished with status $status" >&2
          echo "e2e: an older proven review comment, if one is there, is not a review of this commit" >&2
          echo "e2e: read the run in the lab's Actions tab, or run: make logs SERVICE=runner" >&2
          exit 1
          ;;
        success)
          comments=$(gitea_curl "$base/issues/$num/comments" 2>/dev/null || true)
          # The newest run's own review, or a review by another run of this same head commit that succeeded: a second
          # run of one head posts nothing (lab/workflow/proven.yml), so the review is the first run's.
          notes=
          for review_run in $(success_runs "$statuses" "$context"); do
            if run_notes=$(review_for "$comments" "$review_run" "$sha" "$LAB_BOT_USER"); then
              echo "e2e: proven review is on the pull request (workflow run #$review_run succeeded; head commit $sha; the newest run is #$run_no)"
              printf '%s\n' "$comments"
              exit 0
            fi
            notes="$notes$run_notes
"
          done
          last="workflow run #$run_no succeeded, but no proven review by $LAB_BOT_USER names a succeeded run and head commit $sha"
          notes=$(printf '%s' "$notes" | sort -u | sed '/^$/d')
          if [ -n "$notes" ]; then
            last="$last ($(printf '%s' "$notes" | tr '\n' ';'))"
          fi
          ;;
      esac
    fi
  fi
  sleep "$E2E_POLL_SECONDS"
done

echo "e2e: no proven review from the current run after ${E2E_WAIT_SECONDS}s: $last" >&2
echo "e2e: open the pull request's Actions tab in the lab, or run: make logs SERVICE=runner" >&2
exit 1
