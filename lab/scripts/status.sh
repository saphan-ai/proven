#!/bin/sh
# Containers, Gitea health, runner registration, the latest review on the pull request, the lab URL.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

lab_url="http://${LAB_HOST}:${GITEA_HTTP_PORT}/"
echo "lab: $lab_url"
if service_running lgtm; then
  echo "dashboard: http://${LAB_HOST}:${LGTM_GRAFANA_PORT}/d/${LGTM_DASHBOARD_UID}"
else
  echo "dashboard: not running"
fi
echo
echo "containers:"
compose ps
echo
echo "gitea health:"
internal="http://${GITEA_INTERNAL_HOST}:${GITEA_CONTAINER_HTTP_PORT}"
if service_running gitea; then
  compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 10 "$internal/api/healthz" || echo "gitea: health check failed"
else
  echo "gitea: not running"
fi
echo
echo "runner:"
if ! service_running runner; then
  echo "runner: not running"
elif compose exec -T runner sh -c 'test -s /data/.runner'; then
  echo "runner: registered"
else
  echo "runner: running, not registered"
fi
echo
echo "pull request:"
base="http://${GITEA_INTERNAL_HOST}:${GITEA_CONTAINER_HTTP_PORT}/api/v1/repos/${LAB_ADMIN_USER}/${LAB_DEMO_REPO}"
if service_running gitea; then
  pulls=$(compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 15 "$base/pulls?state=open" || true)
  num=$(printf '%s' "$pulls" | sed -n 's/.*"number":\([0-9][0-9]*\).*/\1/p' | head -n 1)
  if [ -z "$num" ]; then
    echo "pull request: none open"
  else
    echo "pull request: $num"
    comments=$(compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 15 "$base/issues/$num/comments" || true)
    if printf '%s' "$comments" | grep -q 'proven-review'; then
      echo "review: present"
    else
      echo "review: not on the pull request yet"
    fi
  fi
else
  echo "pull request: gitea is not running"
fi
