#!/bin/sh
# Require the demo to be whole (needs the running lab; it is the live half of the controls):
#   - every pull request in lab/seed/prs.tsv exists and carries Proven's comment within E2E_WAIT_SECONDS;
#   - a pull request the manifest calls clean carries the "clean:" line and no finding;
#   - a pull request the manifest calls defect carries at least one finding;
#   - the trace store holds at least one proven.review trace per pull request.
# It exits 0 only when all four hold, and names each pull request that does not.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
. "$ROOT/lab/scripts/gitea-api.sh"
need_docker

if ! service_running gitea; then
  die "demo-check: the lab is not running; run make up"
fi
gitea_login
base=$(gitea_repo_api)
manifest="$ROOT/lab/seed/prs.tsv"
tab=$(printf '\t')
fail=0
total=0

grep -v '^#' "$manifest" | grep -v '^$' >/tmp/demo-check-rows.$$ || true
trap 'rm -f /tmp/demo-check-rows.$$' EXIT

# The rows are read on fd 3: gitea_curl runs docker compose run -T, which drains the loop's stdin.
rows=$(wc -l </tmp/demo-check-rows.$$ | tr -d ' ')
started=$(date +%s)
while IFS="$tab" read -r branch fixture class verdict title says <&3; do
  total=$((total + 1))
  num=$(gitea_curl "$base/pulls/$LAB_DEMO_BASE/$branch" 2>/dev/null | grep -o '"number":[0-9]*' | head -n 1 | cut -d: -f2 || true)
  if [ -z "$num" ]; then
    echo "demo-check: [$total/$rows] $branch ($title): no pull request" >&2
    fail=1
    continue
  fi
  waited=0
  body=
  announced=0
  while :; do
    body=$(gitea_curl "$base/issues/$num/comments" 2>/dev/null || true)
    if printf '%s' "$body" | grep -q 'proven-review'; then
      break
    fi
    # A person waiting sees which pull request it is, and for how long the whole wait may last.
    if [ "$announced" -eq 0 ]; then
      echo "demo-check: [$total/$rows] #$num $branch: waiting for Proven's review ($title)"
      announced=1
    fi
    # Two bounds: this pull request's own (E2E_WAIT_SECONDS) and the whole wait's (DEMO_WAIT_SECONDS).
    if [ "$waited" -ge "$E2E_WAIT_SECONDS" ] || [ $(($(date +%s) - started)) -ge "$DEMO_WAIT_SECONDS" ]; then
      break
    fi
    waited=$((waited + E2E_POLL_SECONDS))
    sleep "$E2E_POLL_SECONDS"
  done
  if ! printf '%s' "$body" | grep -q 'proven-review'; then
    echo "demo-check: [$total/$rows] #$num $branch: no review comment after $(($(date +%s) - started))s of waiting in all" >&2
    fail=1
    continue
  fi
  has_clean=0
  has_finding=0
  if printf '%s' "$body" | grep -qF 'clean: Proven kept no finding'; then
    has_clean=1
  fi
  if printf '%s' "$body" | grep -qF '\nfinding '; then
    has_finding=1
  fi
  case $verdict in
    clean)
      if [ "$has_clean" -ne 1 ] || [ "$has_finding" -ne 0 ]; then
        echo "demo-check: [$total/$rows] #$num $branch: the manifest says clean; the comment has clean=$has_clean findings=$has_finding" >&2
        fail=1
        continue
      fi
      ;;
    defect)
      if [ "$has_finding" -ne 1 ] || [ "$has_clean" -ne 0 ]; then
        echo "demo-check: [$total/$rows] #$num $branch: the manifest says defect; the comment has clean=$has_clean findings=$has_finding" >&2
        fail=1
        continue
      fi
      ;;
    *)
      echo "demo-check: $branch: the manifest's verdict is $verdict, not clean or defect" >&2
      fail=1
      continue
      ;;
  esac
  echo "demo-check: [$total/$rows] #$num $branch: $verdict, as the manifest says"
done 3</tmp/demo-check-rows.$$

# One proven.review trace per pull request, at least. The trace carries no pull request number, so the count
# is the instrument: fewer traces than pull requests is a review whose export did not arrive.
grafana="http://${LGTM_INTERNAL_HOST}:${LGTM_GRAFANA_CONTAINER_PORT}"
lab_curl() {
  compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 20 "$@"
}
uid=$(lab_curl "$grafana/api/datasources/name/Tempo" 2>/dev/null | sed -n 's/.*"uid":"\([^"]*\)".*/\1/p' | head -n 1 || true)
if [ -z "$uid" ]; then
  echo "demo-check: Grafana has no data source named Tempo, or did not answer" >&2
  exit 1
fi
# The export can trail the comment by a moment, so the count is read again for up to TRACE_WAIT_SECONDS.
trace_waited=0
while :; do
  found=$(lab_curl --get --data-urlencode 'q={ name = "proven.review" }' --data-urlencode 'limit=100' \
    "$grafana/api/datasources/proxy/uid/$uid/api/search" 2>/dev/null || true)
  traces=$(printf '%s' "$found" | grep -o '"traceID":"[0-9a-f]*"' | sort -u | wc -l | tr -d ' ')
  if [ "$traces" -ge "$total" ] || [ "$trace_waited" -ge "$TRACE_WAIT_SECONDS" ]; then
    break
  fi
  trace_waited=$((trace_waited + E2E_POLL_SECONDS))
  sleep "$E2E_POLL_SECONDS"
done
echo "demo-check: $traces proven.review traces for $total pull requests"
if [ "$traces" -lt "$total" ]; then
  echo "demo-check: fewer traces than pull requests: a review's export did not arrive (make logs SERVICE=lgtm)" >&2
  fail=1
fi

exit "$fail"
