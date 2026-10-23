#!/bin/sh
# Start the lab if it is down, then wait until proven's review is on the pull request.
# If the lab is up and the runner is not, exit 1 and say so. Do not start the runner then.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
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

base="http://${GITEA_INTERNAL_HOST}:${GITEA_CONTAINER_HTTP_PORT}/api/v1/repos/${LAB_ADMIN_USER}/${LAB_DEMO_REPO}"
deadline=$(($(date +%s) + E2E_WAIT_SECONDS))
while [ "$(date +%s)" -lt "$deadline" ]; do
  pulls=$(compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 15 "$base/pulls?state=open" || true)
  num=$(printf '%s' "$pulls" | sed -n 's/.*"number":\([0-9][0-9]*\).*/\1/p' | head -n 1)
  if [ -n "$num" ]; then
    comments=$(compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 15 "$base/issues/$num/comments" || true)
    if printf '%s' "$comments" | grep -q 'proven-review'; then
      echo "e2e: proven review is on the pull request"
      printf '%s\n' "$comments"
      exit 0
    fi
  fi
  sleep "$E2E_POLL_SECONDS"
done

echo "e2e: no proven review on the pull request after ${E2E_WAIT_SECONDS}s" >&2
echo "e2e: open the pull request's Actions tab in the lab, or run: make logs SERVICE=runner" >&2
exit 1
