#!/bin/sh
# Negative control for `make e2e`: an old proven review comment is on the pull
# request, the runner is alive, and the CURRENT run fails. e2e must exit non-zero.
#
# It needs a lab that already passed `make e2e`, so an old review comment exists.
# It commits a workflow that fails on purpose to the pull request's head branch,
# runs e2e, and always puts the original workflow back. Putting it back starts
# one more run, which succeeds; `make e2e` passes again once that run is done.
# Exit 0: e2e refused the failed run and named it. Exit 1: it did not.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
. "$here/lib.sh"
. "$here/gitea-api.sh"
need_docker

if ! service_running gitea || ! service_running runner; then
  die "e2e-failed-run: the lab is not running; run make e2e first"
fi
gitea_login
base=$(gitea_repo_api)
file_url="$base/contents/.github/workflows/proven.yml"

pulls=$(gitea_curl "$base/pulls?state=open")
num=$(printf '%s' "$pulls" | sed -n 's/.*"number":\([0-9][0-9]*\).*/\1/p' | head -n 1)
if [ -z "$num" ]; then
  die "e2e-failed-run: no open pull request; run make up"
fi
comments=$(gitea_curl "$base/issues/$num/comments")
if ! printf '%s' "$comments" | grep -q 'proven-review'; then
  die "e2e-failed-run: no earlier proven review comment on pull request #$num; run make e2e first"
fi

# The workflow file as the head branch holds it now. Gitea returns it in base64.
current=$(gitea_curl "$file_url?ref=$LAB_DEMO_HEAD")
original=$(printf '%s' "$current" | json_first content | sed 's/\\n//g')
if [ -z "$original" ]; then
  die "e2e-failed-run: could not read the workflow from branch $LAB_DEMO_HEAD"
fi

# put_workflow BASE64 MESSAGE — commit BASE64 as the workflow on the head branch.
put_workflow() {
  file_sha=$(gitea_curl "$file_url?ref=$LAB_DEMO_HEAD" | json_first sha)
  printf '{"message":"%s","branch":"%s","sha":"%s","content":"%s"}' \
    "$2" "$LAB_DEMO_HEAD" "$file_sha" "$1" |
    gitea_curl -X PUT -H 'Content-Type: application/json' --data-binary @- "$file_url" >/dev/null
}

restore() {
  if put_workflow "$original" "Restore the proven workflow after the negative control"; then
    echo "e2e-failed-run: the original workflow is back; a new run starts, wait for it before make e2e"
  else
    echo "e2e-failed-run: could not restore the workflow on branch $LAB_DEMO_HEAD; run make clean && make up" >&2
  fi
}
trap restore EXIT

failing=$(printf '%s\n' \
  'name: proven' \
  'on:' \
  '  pull_request:' \
  'defaults:' \
  '  run:' \
  '    shell: sh' \
  'jobs:' \
  '  review:' \
  "    runs-on: $RUNNER_LABEL" \
  '    steps:' \
  '      - name: Fail on purpose (lab negative control)' \
  '        run: |' \
  '          echo "negative control: this run fails on purpose" >&2' \
  '          exit 1' |
  base64 | tr -d '\n')

put_workflow "$failing" "Negative control: make the current run fail"
echo "e2e-failed-run: pull request #$num has an old review comment and a workflow that fails; running e2e"

rc=0
out=$(sh "$here/e2e.sh" 2>&1) || rc=$?
printf '%s\n' "$out"
if [ "$rc" -eq 0 ]; then
  echo "e2e-failed-run: FAIL: e2e exited 0 although the current run fails" >&2
  exit 1
fi
if ! printf '%s' "$out" | grep -q 'finished with status failure'; then
  echo "e2e-failed-run: FAIL: e2e exited $rc but did not name the failed run" >&2
  exit 1
fi
echo "e2e-failed-run: ok: e2e exited $rc and named the failed run"
