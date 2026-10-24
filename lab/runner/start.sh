#!/bin/sh
# Register once, then run the daemon. The registration token is written by the seed.
set -eu

if ! command -v act_runner >/dev/null 2>&1; then
  echo "runner: act_runner is not on PATH in this image" >&2
  exit 1
fi

cd /data

if [ ! -s /data/.runner ]; then
  i=0
  while [ ! -s /reg/token ]; do
    i=$((i + 1))
    if [ "$i" -gt 150 ]; then
      echo "runner: registration token did not appear" >&2
      exit 1
    fi
    echo "runner: waiting for the registration token"
    sleep 2
  done
  token=$(cat /reg/token)
  act_runner register \
    --no-interactive \
    --config /config.yaml \
    --instance "${GITEA_INSTANCE_URL}" \
    --token "$token" \
    --name "${GITEA_RUNNER_NAME}" \
    --labels "${GITEA_RUNNER_LABELS}"
fi

exec act_runner daemon --config /config.yaml
