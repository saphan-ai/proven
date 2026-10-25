#!/bin/sh
# Ask the dashboard's trace store for the latest review's trace (ADR-0045).
# The query is TraceQL, sent to Tempo through Grafana's data source proxy, from
# inside the lab network. It prints the query, the trace id, the span names and
# the reviewer models the spans name. It exits 0 only when the trace holds the
# review's root span, the planner, the skeptic and the judge, and both reviewers
# (default and special). It exits 1 when the store holds no such trace within
# TRACE_WAIT_SECONDS, or the trace lacks one of those.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

if ! service_running lgtm; then
  die "trace: the dashboard stack is not running; run make up"
fi

grafana="http://${LGTM_INTERNAL_HOST}:${LGTM_GRAFANA_CONTAINER_PORT}"
# curl runs in the seed image, which is on the lab network and has curl.
lab_curl() {
  compose run --rm --no-deps -T --entrypoint curl seed -fsS --max-time 20 "$@"
}

uid=$(lab_curl "$grafana/api/datasources/name/Tempo" 2>/dev/null | sed -n 's/.*"uid":"\([^"]*\)".*/\1/p' | head -n 1 || true)
if [ -z "$uid" ]; then
  die "trace: Grafana has no data source named Tempo, or did not answer"
fi
proxy="$grafana/api/datasources/proxy/uid/$uid"

query='{ name = "proven.review" }'
echo "trace: query: $query"

i=0
id=
while [ "$i" -lt "$TRACE_WAIT_SECONDS" ]; do
  found=$(lab_curl --get --data-urlencode "q=$query" --data-urlencode "limit=1" "$proxy/api/search" 2>/dev/null || true)
  id=$(printf '%s' "$found" | sed -n 's/.*"traceID":"\([0-9a-f]*\)".*/\1/p' | head -n 1)
  if [ -n "$id" ]; then
    break
  fi
  i=$((i + 5))
  sleep 5
done
if [ -z "$id" ]; then
  echo "trace: the store holds no trace named proven.review after ${TRACE_WAIT_SECONDS}s" >&2
  exit 1
fi
echo "trace: trace id: $id"

body=$(lab_curl "$proxy/api/traces/$id")
names=$(printf '%s' "$body" | grep -o '"name":"[^"]*"' | sort -u || true)
reviewers=$(printf '%s' "$body" | tr '{' '\n' | grep -A1 '"key":"proven.reviewer"' | grep -o '"stringValue":"[^"]*"' | sort | uniq -c || true)
echo "trace: spans:"
printf '%s\n' "$names" | sed 's/^"name":"\(.*\)"$/  \1/'
echo "trace: reviewers named by the spans (count, reviewer):"
printf '%s\n' "$reviewers" | sed 's/"stringValue":"\(.*\)"/\1/; s/^/  /'

fail=0
for want in 'proven.review' 'agent planner' 'agent skeptic' 'agent judge'; do
  if ! printf '%s\n' "$names" | grep -q "\"name\":\"$want\""; then
    echo "trace: no span named $want" >&2
    fail=1
  fi
done
for want in default special; do
  if ! printf '%s\n' "$reviewers" | grep -q "\"stringValue\":\"$want\""; then
    echo "trace: no span names the $want reviewer" >&2
    fail=1
  fi
done
exit "$fail"
