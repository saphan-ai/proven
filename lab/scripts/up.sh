#!/bin/sh
# Build the images, start the lab, create the admin when it is missing, seed the demo, start the runner.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
. "$ROOT/lab/scripts/banner.sh"

# Keys for a live review (docs/LAB.md, "Go live"): lab/.env is read here, and the seed stores each non-empty one as
# an Actions secret. Nothing is printed from it. It is read first so the refusals below see its settings too.
load_lab_env || exit 1
# LAB_BIND is retired: the lab publishes on 127.0.0.1 only. A LAB_BIND in the environment is refused by name.
refuse_retired_bind
# The start page carries the lab's password and is published on 127.0.0.1 in lab/compose.yaml. A LAB_PAGE_BIND that
# names another address is refused by name. Both refusals come ahead of the Docker check, so a refused bind is named
# on a host with no Docker as well, and before any container starts.
if [ "${LAB_PAGE_BIND+set}" = set ]; then
  lab_require_loopback LAB_PAGE_BIND "$LAB_PAGE_BIND" || exit 1
fi

# The flow (lab/scripts/flow.sh): the settings are validated, the lab's ports are checked free, and the questions are
# asked on a terminal for what no setting answered. With no terminal and no setting it is a sample lab and nothing is
# asked. It comes before any container, so a refusal here has started nothing.
. "$ROOT/lab/scripts/flow.sh"
lab_flow
need_docker

# What a person is told while this runs (lab/scripts/ui.sh): the plan now, before the first Docker call, then one status
# line after each phase. The EXIT trap is the only one of this script: it names the phase that failed, then removes the
# temporary files of the admin-user check.
users=
err=
trap 'ui_exit $?; rm -f "$users" "$err"' EXIT
ui_plan
ui_begin 1 "Build Proven's image" "the build output above"

# A password the judge supplies is checked before any container starts. Gitea
# refuses one shorter than 8 characters, and a line break would split it.
if [ -n "${LAB_ADMIN_PASSWORD:-}" ]; then
  case $LAB_ADMIN_PASSWORD in
    *'
'*) die "up: LAB_ADMIN_PASSWORD holds a line break" ;;
  esac
  if [ "${#LAB_ADMIN_PASSWORD}" -lt 8 ]; then
    die "up: LAB_ADMIN_PASSWORD is shorter than 8 characters, which Gitea refuses"
  fi
fi

if [ -z "${GITEA_RUN_USER:-}" ]; then
  die "up: GITEA_RUN_USER is not set"
fi

if ! grep -q "tcp://127.0.0.1:${DIND_PORT}" "$ROOT/lab/runner/config.yaml"; then
  die "up: lab/runner/config.yaml does not send jobs to tcp://127.0.0.1:${DIND_PORT}"
fi

sh "$ROOT/lab/scripts/image.sh"
sh "$ROOT/lab/scripts/seed-image.sh"

# The documentation site is one more image, built from the repository (docs-site/Dockerfile). It is a pleasant
# thing to have and not what the lab is for: a build that fails is named and the lab goes on without it.
docs_ready=1
if ! compose build docs; then
  docs_ready=0
  echo "up: the documentation image did not build; the lab goes on without it (the page's docs link will not open)" >&2
fi
ui_done "Proven's image is built"

ui_begin 2 "Start Gitea, Grafana and the lab's Docker" "make logs SERVICE=gitea, or SERVICE=lgtm, or SERVICE=dind"
compose up -d gitea dind lgtm

# Path C: Zetaris on this machine, when asked (make up ZETARIS=local). It needs the lab's network, so it starts here.
# A platform that does not start is named and the lab goes on, as the documentation image does.
if [ "${ZETARIS:-}" = "local" ]; then
  if sh "$ROOT/lab/scripts/zetaris.sh"; then
    # The lab's reviews reach it by its name on the lab network: no hand edit of lab/.env for the address.
    if [ -z "${ZETARIS_REST_URL:-}" ]; then
      ZETARIS_REST_URL="http://${ZETARIS_LAB_ALIAS}:${ZETARIS_API_PORT}${ZETARIS_REST_PATH}"
      export ZETARIS_REST_URL
    fi
  else
    echo "up: ZETARIS=local did not complete (see above); the lab goes on without it" >&2
  fi
fi

internal="http://${GITEA_INTERNAL_HOST}:${GITEA_CONTAINER_HTTP_PORT}"

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

# ADR-0045: the dashboard is up before the pull request's review starts, so the review is exported to it.
# A Grafana that does not answer is named, and the lab goes on: the review is still posted.
grafana_internal="http://${LGTM_INTERNAL_HOST}:${LGTM_GRAFANA_CONTAINER_PORT}"
grafana_ready=0
waited=0
while [ "$waited" -lt "$LGTM_WAIT_SECONDS" ]; do
  if compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 3 "$grafana_internal/api/health" >/dev/null 2>&1; then
    grafana_ready=1
    break
  fi
  waited=$((waited + 5))
  sleep 5
done
if [ "$grafana_ready" -ne 1 ]; then
  echo "up: Grafana did not answer within ${LGTM_WAIT_SECONDS}s; the dashboard may be empty at first (make logs SERVICE=lgtm)" >&2
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

# The Prover starts each test in PROVEN_TEST_IMAGE on this inner daemon. The
# lab measured "No such image: golang:1.26" when nothing had pulled it. The
# pull is by digest when TEST_IMAGE_DIGEST is set, then the image is tagged
# with the name the Prover uses.
if [ -n "${TEST_IMAGE_DIGEST:-}" ]; then
  test_repo=${PROVEN_TEST_IMAGE%%:*}
  docker exec "$dind_id" docker pull "${test_repo}@${TEST_IMAGE_DIGEST}"
  docker exec "$dind_id" docker tag "${test_repo}@${TEST_IMAGE_DIGEST}" "$PROVEN_TEST_IMAGE"
else
  echo "up: TEST_IMAGE_DIGEST is not set; pulling $PROVEN_TEST_IMAGE by tag" >&2
  docker exec "$dind_id" docker pull "$PROVEN_TEST_IMAGE"
  pulled=$(docker exec "$dind_id" docker image inspect --format '{{index .RepoDigests 0}}' "$PROVEN_TEST_IMAGE" 2>/dev/null || true)
  echo "up: the tag resolved to ${pulled:-an unknown digest}. Set TEST_IMAGE_DIGEST in lab/versions.env to pin it." >&2
fi

pass=$(compose run --rm --no-deps -T --entrypoint cat seed /state/password 2>/dev/null || true)
if [ -z "$pass" ]; then
  if [ -n "${LAB_ADMIN_PASSWORD:-}" ]; then
    pass=$LAB_ADMIN_PASSWORD
    supplied=1
  else
    pass=$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')
    supplied=0
  fi
  printf '%s' "$pass" | compose run --rm --no-deps -T --entrypoint sh seed -c \
    'umask 077 && cat > /state/password'
  # The password goes into no log. On a terminal the banner at the end shows it, and make password shows it again.
  # Anywhere else it is kept in a 0600 file and only the path is printed (deliver_admin_password in lib.sh).
  if [ "$supplied" -eq 1 ]; then
    echo "up: lab admin password: the value of LAB_ADMIN_PASSWORD, not printed"
  elif stdout_is_terminal; then
    echo "up: generated the lab admin password; it is in the lab volume until make clean"
  else
    deliver_admin_password "$pass"
  fi
else
  echo "up: the lab is already installed; the volume is kept; repairing the admin user, the runner and the pull requests"
  if [ -n "${LAB_ADMIN_PASSWORD:-}" ] && [ "$LAB_ADMIN_PASSWORD" != "$pass" ]; then
    echo "up: LAB_ADMIN_PASSWORD differs from the installed lab's password and is ignored. make clean, then make up, installs it." >&2
  fi
fi

# The image refuses `gitea` as root. `compose exec` is root unless a user is
# named, so every CLI call uses GITEA_RUN_USER. A kept volume is not a reason
# to stop: the calls below create whatever is still missing.
gitea_cli() {
  compose exec -T -u "$GITEA_RUN_USER" gitea gitea "$@"
}

# The list is read once. A clean clone has no private tree beside the
# repository, so the file is temporary and this script removes it on exit.
users=$(mktemp)
err=$(mktemp)
i=0
listed=0
while [ "$i" -lt 20 ]; do
  if gitea_cli admin user list >"$users" 2>"$err"; then
    listed=1
    break
  fi
  i=$((i + 1))
  sleep 2
done
if [ "$listed" -ne 1 ]; then
  cat "$err" >&2
  die "up: Gitea did not accept an admin command"
fi
if ! grep -q "$LAB_ADMIN_USER" "$users"; then
  gitea_cli admin user create \
    --admin \
    --username "$LAB_ADMIN_USER" \
    --password "$pass" \
    --email "$LAB_ADMIN_EMAIL" \
    --must-change-password=false
fi
rm -f "$users" "$err"
ui_done "Gitea, Grafana and the lab's Docker are up"

ui_begin 3 "Create the demo repository and its pull requests" "make logs SERVICE=seed, or SERVICE=runner"
# The seed reads the password from /state/password, written above, and every credential from /state/secrets, written
# here at mode 0600 over standard input. Neither is passed in. The credentials file is removed when the seed exits,
# and removed again here whatever the seed's status was, so a seed that failed midway leaves none.
lab_write_secrets || die "up: the credentials were not handed to the lab. Nothing was seeded."
seed_status=0
compose run --rm -T seed </dev/null || seed_status=$?
lab_remove_secrets
if [ "$seed_status" -ne 0 ]; then
  echo "up: the seed failed (status $seed_status); the credentials file was removed." >&2
  exit "$seed_status"
fi
compose up -d runner

# The start page is rendered by the seed into the page volume, so the service starts after it.
compose up -d page
if [ "$docs_ready" -eq 1 ]; then
  compose up -d docs
fi
ui_done "The demo repository and its $(ui_pr_count) pull requests are created, and the CI runner is running"

# Proven reviews each pull request in the runner. The banner appears when the dashboards have data, so this waits for
# every review (demo-check.sh, with a line per pull request and DEMO_WAIT_SECONDS as the bound). A review that does not
# arrive is named in red and the lab is still shown: it is up, and make demo-check waits again.
wait_status=0
ui_begin 4 "Wait for Proven to review every pull request" "make logs SERVICE=runner, then make demo-check"
if [ "${PROVEN_LAB_WAIT:-yes}" = no ]; then
  ui_done "Not waiting for the reviews (PROVEN_LAB_WAIT=no); make demo-check waits for them"
else
  sh "$ROOT/lab/scripts/demo-check.sh" || wait_status=$?
  if [ "$wait_status" -eq 0 ]; then
    ui_done "Proven reviewed every pull request"
  else
    ui_fail "the lab is up, but not every review is there yet"
  fi
fi

# The URL, the user and the password are the last lines of every run, the first and the ones after it. A password
# the judge supplied in LAB_ADMIN_PASSWORD is the judge's own and is not printed: the line names the setting.
shown=$pass
if [ -n "${LAB_ADMIN_PASSWORD:-}" ] && [ "$LAB_ADMIN_PASSWORD" = "$pass" ]; then
  shown="the value of LAB_ADMIN_PASSWORD"
fi
echo
lab_banner "$shown"
exit "$wait_status"
