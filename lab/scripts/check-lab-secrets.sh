#!/bin/sh
# The live control for round 5 (docs/LAB.md, "How a credential travels"). It needs Docker and a lab that `make up`
# started with a sentinel in each credential's name; run it with the SAME environment, so the Compose render sees them:
#
#   for n in $(sed -n 's/^LAB_SECRET_KEYS=//p' lab/versions.env); do export "$n=sentinel-$n-0123456789"; done
#   PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=no PROVEN_LAB_OWN_SETUP=no make up   (the keys are refused upstream: that is fine)
#   sh lab/scripts/check-lab-secrets.sh
#
# It checks that no sentinel is in `docker compose config`, in `docker inspect` of every container of the project
# (the exited seed runs too), or in any container's process arguments, and that every sentinel's NAME is an Actions
# secret of the demo repository. Then the transfer file: it must not exist in the state volume. It prints names, never
# values. Exit 0 only when every line says ok.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
. "$ROOT/lab/scripts/gitea-api.sh"

need_docker
bad=0
say() { echo "check-lab-secrets: $1"; }

for name in $LAB_SECRET_KEYS; do
  eval "val=\${$name:-}"
  if [ "$val" != "sentinel-$name-0123456789" ]; then
    die "check-lab-secrets: $name is not set to sentinel-$name-0123456789 in this shell. NEXT — export it as the header of this file says, then make up, then this check."
  fi
done

cfg=$(compose config 2>&1) || die "check-lab-secrets: docker compose config failed"
ids=$(docker ps -aq --filter "label=com.docker.compose.project=$COMPOSE_PROJECT_NAME")
insp=""
for id in $ids; do
  insp="$insp$(docker inspect "$id")"
done
procs=""
for id in $(docker ps -q --filter "label=com.docker.compose.project=$COMPOSE_PROJECT_NAME"); do
  procs="$procs$(docker top "$id" -eo args 2>/dev/null || true)"
done

for name in $LAB_SECRET_KEYS; do
  s="sentinel-$name-0123456789"
  for where in cfg insp procs; do
    eval "text=\$$where"
    if printf '%s' "$text" | grep -qF "$s"; then
      say "RED: the value of $name is in $where"
      bad=1
    fi
  done
done
[ "$bad" -eq 0 ] && say "ok: no sentinel in docker compose config, docker inspect (every container of the project) or process arguments"

list=$(gitea_curl "$(gitea_repo_api)/actions/secrets" </dev/null) || die "check-lab-secrets: could not list the Actions secrets"
for name in $LAB_SECRET_KEYS; do
  if printf '%s' "$list" | grep -q "\"name\":\"$name\""; then
    say "ok: $name is an Actions secret of the demo repository"
  else
    say "RED: $name is not an Actions secret of the demo repository"
    bad=1
  fi
done

if compose run --rm --no-deps -T --entrypoint test seed -e /state/secrets </dev/null 2>/dev/null; then
  say "RED: /state/secrets still exists in the state volume"
  bad=1
else
  say "ok: /state/secrets is gone from the state volume"
fi
exit "$bad"
