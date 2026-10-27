#!/bin/sh
# Build the report page of every demo pull request: the review Proven posted on it, set in the way GitHub shows a Markdown
# comment, served at <the start page>/report/N.html (lab/page/nginx.conf). make up runs it once the reviews are in; make
# reports runs it again.
#
# The page is that pull request's own comment, read back from Gitea (lab/seed/report.sh comment): in sample mode and in a
# live lab alike it says what the comment says, a live finding or a failed service included. No review is made here: a
# comment that is not posted yet is waited for (LAB_REPORT_WAIT_SECONDS, 120 by default), then named, and its Report link
# keeps saying the report is not ready. The read is anonymous (the demo repository is public): nothing here passes a
# credential, and the lab's keys are not read.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

rows=$(mktemp)
trap 'rm -f "$rows"' EXIT
compose run --rm --no-deps -T --entrypoint cat seed /state/prs.tsv </dev/null >"$rows" 2>/dev/null || true
if [ ! -s "$rows" ]; then
  die "reports: the lab has stored no pull requests yet. NEXT — run make up."
fi

wait_for=${LAB_REPORT_WAIT_SECONDS:-120}
case $wait_for in '' | *[!0-9]*) die "reports: LAB_REPORT_WAIT_SECONDS is not a number of seconds" ;; esac
deadline=$(($(date +%s) + wait_for))
tab=$(printf '\t')
done_count=0
failed=0
# The rows are read on fd 3: the commands inside the loop read standard input and would drain it.
while IFS="$tab" read -r num branch fixture verdict <&3; do
  [ -n "$num" ] || continue
  md=$(mktemp)
  status=0
  while :; do
    status=0
    compose run --rm --no-deps -T --entrypoint sh seed -c 'exec sh /usr/local/bin/report.sh comment "$1"' sh "$num" \
      </dev/null >"$md" 2>/dev/null || status=$?
    [ "$status" -eq 3 ] && [ "$(date +%s)" -lt "$deadline" ] || break
    sleep 5
  done
  if [ "$status" -eq 0 ] &&
    compose run --rm --no-deps -T --entrypoint sh seed -c 'exec sh /usr/local/bin/report.sh render "$1"' sh "$num" <"$md" >/dev/null; then
    echo "reports: #$num $branch: the report page is built from its comment"
    done_count=$((done_count + 1))
  elif [ "$status" -eq 3 ]; then
    echo "reports: #$num $branch: Proven has not posted its review yet; make reports builds the page once it has." >&2
    failed=1
  else
    echo "reports: #$num $branch: the report page was not built. The pull request's review is the comment in Gitea." >&2
    failed=1
  fi
  rm -f "$md"
done 3<"$rows"

echo "reports: $done_count report pages built"
exit "$failed"
