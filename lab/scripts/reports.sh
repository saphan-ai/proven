#!/bin/sh
# Build the report page of every demo pull request: the review Proven writes with `proven scan --md`, set in the way
# GitHub shows a Markdown comment, served at <the start page>/report/N.html (lab/page/nginx.conf). make up runs it after the
# seed; make reports runs it again.
#
# The review is Proven's sample review of that pull request's change: the same recorded answers the workflow's job uses in
# sample mode, so the page says what the comment says. In a live lab the page says it is the sample review and that the live
# review is the comment (lab/seed/report.sh). The run takes no network and no key: Proven's image, with no network, the
# change's fixture, SAMPLE_MODE=true, and the settings the workflow gives its review (the one automated fix, the special
# reviewer), so the page names the reviewers the comment names. Nothing here passes a credential: the lab's keys are not read.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

rows=$(mktemp)
trap 'rm -f "$rows"' EXIT
compose run --rm --no-deps -T --entrypoint cat seed /state/prs.tsv </dev/null >"$rows" 2>/dev/null || true
if [ ! -s "$rows" ]; then
  die "reports: the lab has stored no pull requests yet. NEXT — run make up."
fi

# The two Meterless features the lab passes to a review ride in sample mode too, so the page matches the comment.
features=
if [ -n "${PROVEN_MARKOVIAN:-}" ]; then
  features="-e PROVEN_MARKOVIAN"
fi

tab=$(printf '\t')
done_count=0
failed=0
# The rows are read on fd 3: the commands inside the loop read standard input and would drain it.
while IFS="$tab" read -r num branch fixture verdict <&3; do
  [ -n "$num" ] || continue
  md=$(mktemp)
  # $features is a fixed word or empty, never a value: the flag names a variable, it does not carry one.
  # shellcheck disable=SC2086
  if docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges:true \
    -e SAMPLE_MODE=true -e FIXTURE="$fixture" -e SCOUT_INTENT="${PROVEN_LAB_SCOUT_INTENT:-}" -e SPECIAL_REVIEWER="${LAB_SPECIAL_REVIEWER:-}" $features \
    --entrypoint sh "$PROVEN_IMAGE:latest" -c 'mkdir -p /tmp/proven-run && cd /tmp/proven-run && d="/opt/proven/fixture/$FIXTURE" && { [ -d "$d" ] || d=/opt/proven/fixture/discount; } && echo autofix: on > /tmp/proven-run/proven.yaml && { [ -z "$SPECIAL_REVIEWER" ] || printf "special_reviewer: %s\n" "$SPECIAL_REVIEWER" >> /tmp/proven-run/proven.yaml; } && { [ -z "$SCOUT_INTENT" ] || printf "scout_intent: %s\n" "$SCOUT_INTENT" >> /tmp/proven-run/proven.yaml; } && export PROVEN_CONFIG=/tmp/proven-run/proven.yaml && proven scan --md "$d"' \
    </dev/null >"$md" 2>/dev/null &&
    compose run --rm --no-deps -T --entrypoint sh seed -c 'exec sh /usr/local/bin/report.sh render "$1"' sh "$num" <"$md" >/dev/null; then
    echo "reports: #$num $branch: the report page is built"
    done_count=$((done_count + 1))
  else
    echo "reports: #$num $branch: the report page was not built. The pull request's review is the comment in Gitea." >&2
    failed=1
  fi
  rm -f "$md"
done 3<"$rows"

echo "reports: $done_count report pages built"
exit "$failed"
