#!/bin/sh
# Containers, Gitea health, runner registration, the review on each demo pull request, and the lab's login.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
. "$ROOT/lab/scripts/banner.sh"
need_docker

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
echo "pull requests:"
base="http://${GITEA_INTERNAL_HOST}:${GITEA_CONTAINER_HTTP_PORT}/api/v1/repos/${LAB_ADMIN_USER}/${LAB_DEMO_REPO}"
if service_running gitea; then
  pulls=$(compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 15 "$base/pulls?state=open&limit=50" || true)
  nums=$(printf '%s' "$pulls" | grep -o '"number":[0-9]*' | cut -d: -f2 | sort -n || true)
  if [ -z "$nums" ]; then
    echo "pull requests: none open"
  else
    for num in $nums; do
      comments=$(compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 15 "$base/issues/$num/comments" || true)
      if printf '%s' "$comments" | grep -q 'proven-review'; then
        echo "pull request $num: review present"
      else
        echo "pull request $num: review not on it yet"
      fi
    done
  fi
else
  echo "pull requests: gitea is not running"
fi

# The login, in the form make up ends with.
pass=
if service_running gitea; then
  pass=$(lab_read_password)
fi
echo
if [ -n "$pass" ]; then
  lab_banner "$pass"
else
  echo "login: the lab is not installed; run make up"
fi
