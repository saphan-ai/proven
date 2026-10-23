#!/bin/sh
# Build the images, start the lab, create the admin once, seed the demo, start the runner.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

if ! grep -q "tcp://127.0.0.1:${DIND_PORT}" "$ROOT/lab/runner/config.yaml"; then
  die "up: lab/runner/config.yaml does not send jobs to tcp://127.0.0.1:${DIND_PORT}"
fi

sh "$ROOT/lab/scripts/image.sh"
sh "$ROOT/lab/scripts/seed-image.sh"

compose up -d gitea dind

internal="http://${GITEA_INTERNAL_HOST}:${GITEA_CONTAINER_HTTP_PORT}"
lab_url="http://${LAB_BIND}:${GITEA_HTTP_PORT}/"

i=0
ready=0
while [ "$i" -lt 40 ]; do
  if compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 3 "$internal/api/healthz" >/dev/null 2>&1; then
    ready=1
    break
  fi
  i=$((i + 1))
  sleep 3
done
if [ "$ready" -ne 1 ]; then
  die "up: Gitea did not become healthy"
fi

dind_id=$(compose ps -q dind)
if [ -z "$dind_id" ]; then
  die "up: the dind container is not running"
fi
i=0
dockerd=0
while [ "$i" -lt 30 ]; do
  if docker exec "$dind_id" docker info >/dev/null 2>&1; then
    dockerd=1
    break
  fi
  i=$((i + 1))
  sleep 2
done
if [ "$dockerd" -ne 1 ]; then
  die "up: the lab's Docker daemon did not answer"
fi

docker save "$PROVEN_IMAGE:latest" | docker exec -i "$dind_id" docker load
docker exec "$dind_id" docker pull "docker:${DOCKER_VERSION}-cli"

pass=$(compose run --rm --no-deps -T --user 0:0 --entrypoint cat seed /state/password 2>/dev/null || true)
if [ -z "$pass" ]; then
  pass=$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')
  printf '%s' "$pass" | compose run --rm --no-deps -T --user 0:0 --entrypoint sh seed -c \
    'mkdir -p /state /reg && cat > /state/password && chmod 600 /state/password && chown proven:proven /state /reg /state/password && chmod 700 /state /reg'
  echo "lab admin user: $LAB_ADMIN_USER"
  echo "lab admin password: $pass"
  echo "lab url: $lab_url"
  echo "This password is printed once. make clean removes it with the lab volume."
else
  echo "lab already installed; admin password was printed at first start"
  echo "lab url: $lab_url"
fi

i=0
listed=0
while [ "$i" -lt 20 ]; do
  if compose exec -T gitea gitea admin user list >"$ROOT/.saphan/work/judges-lab-makefile/gitea-users.txt" 2>/dev/null; then
    listed=1
    break
  fi
  i=$((i + 1))
  sleep 2
done
if [ "$listed" -ne 1 ]; then
  die "up: Gitea did not accept an admin command"
fi
if ! grep -q "$LAB_ADMIN_USER" "$ROOT/.saphan/work/judges-lab-makefile/gitea-users.txt"; then
  compose exec -T gitea gitea admin user create \
    --admin \
    --username "$LAB_ADMIN_USER" \
    --password "$pass" \
    --email "$LAB_ADMIN_EMAIL" \
    --must-change-password=false
fi
rm -f "$ROOT/.saphan/work/judges-lab-makefile/gitea-users.txt"

LAB_ADMIN_PASSWORD=$pass compose run --rm -T seed
compose up -d runner
echo "lab: $lab_url"
