#!/bin/sh
# The GitHub Action's entry point (ADR-0020 §3). action.yml passes every input as
# INPUT_<NAME> with hyphens as underscores; this maps each to a Proven flag, runs
# `proven scan`, writes the JSON and SARIF files into the workspace and every output
# to $GITHUB_OUTPUT. The exit code is Proven's (ADR-0020 §2): `strict` decides whether
# findings fail the step. PROVEN_BIN names the binary (default: proven on PATH).
set -eu

fail() {
  echo "proven action: $1" >&2
  exit 2
}

bool() {
  case "$2" in
    true | false) ;;
    *) fail "input $1 is '$2'; it takes true or false. Set $1 to true or false." ;;
  esac
}

# The script's own directory, read before the scan changes the working directory:
# post-comment.sh sits beside it.
here=$(cd "$(dirname "$0")" && pwd)
proven=${PROVEN_BIN:-proven}
workspace=${GITHUB_WORKSPACE:-}
[ -n "$workspace" ] || fail "GITHUB_WORKSPACE is not set; run this script from a GitHub Actions step."
[ -d "$workspace" ] || fail "GITHUB_WORKSPACE '$workspace' is not a directory."
[ -n "${GITHUB_OUTPUT:-}" ] || fail "GITHUB_OUTPUT is not set; run this script from a GitHub Actions step."

target=${INPUT_TARGET:-.}
config=${INPUT_CONFIG:-}
output_format=${INPUT_OUTPUT_FORMAT:-json/sarif}
severity=${INPUT_SEVERITY:-INFO}
include=${INPUT_INCLUDE:-}
exclude=${INPUT_EXCLUDE:-}
timeout=${INPUT_TIMEOUT:-}
strict=${INPUT_STRICT:-true}
baseline=${INPUT_BASELINE_COMMIT:-}
sample=${INPUT_SAMPLE_MODE:-false}
comment=${INPUT_COMMENT:-false}

bool strict "$strict"
bool sample-mode "$sample"
bool comment "$comment"
case "$output_format" in
  json/sarif | json | sarif | text) ;;
  *) fail "input output-format is '$output_format'; it takes json/sarif, json, sarif or text." ;;
esac

# Paths are taken from the workspace; the review itself runs in a scratch directory,
# because Proven writes its trace and state under the directory it runs in.
case "$target" in
  /*) ;;
  *) target=$workspace/$target ;;
esac
case "$config" in
  '' | /*) ;;
  *) config=$workspace/$config ;;
esac

# Outputs are workspace-relative: a later step runs on the runner, where the
# workspace is not at the path this container sees.
json_file=proven-results.json
sarif_file=proven-results.sarif
text_file=proven-results.txt

case "$output_format" in
  json) results_file=$json_file ;;
  text) results_file=$text_file ;;
  *) results_file=$sarif_file ;;
esac

set --
set -- "$@" "--severity=$severity"
set -- "$@" "--json-output=$workspace/$json_file" "--sarif-output=$workspace/$sarif_file"
[ -z "$config" ] || set -- "$@" "--config=$config"
[ -z "$timeout" ] || set -- "$@" "--timeout=$timeout"
[ -z "$baseline" ] || set -- "$@" "--baseline-commit=$baseline"
[ "$sample" != true ] || set -- "$@" --sample
if [ "$strict" = true ]; then set -- "$@" --error; else set -- "$@" --no-error; fi
case "$output_format" in
  json) set -- "$@" --json ;;
  sarif) set -- "$@" --sarif ;;
  text) set -- "$@" --text "--text-output=$workspace/$text_file" ;;
  *) set -- "$@" --text ;;
esac

# include and exclude are comma-separated globs; each becomes one flag.
# set -f: a glob is for Proven to match against changed paths, never for this shell
# to expand against the checkout it happens to run in.
oldifs=$IFS
IFS=,
set -f
for glob in $include; do
  [ -z "$glob" ] || set -- "$@" "--include=$glob"
done
for glob in $exclude; do
  [ -z "$glob" ] || set -- "$@" "--exclude=$glob"
done
set +f
IFS=$oldifs

# The write token is read only when comment is true, and before anything runs;
# below, it is removed from the environment Proven runs in.
if [ "$comment" = true ]; then
  token=${INPUT_GITHUB_TOKEN:-}
  [ -n "$token" ] || fail "comment is true but github-token is empty; set github-token to a token that may write pull requests, and give the job pull-requests: write."
fi

# Credentials never reach the scan: Proven reads the code under review, and that
# code must not see a token, whether comment is true or false. The posting token
# was copied into $token above (an unexported shell variable) and goes only to the
# summary comment poster (post-comment.sh).
unset INPUT_GITHUB_TOKEN GITHUB_TOKEN ACTIONS_RUNTIME_TOKEN ACTIONS_ID_TOKEN_REQUEST_TOKEN ACTIONS_ID_TOKEN_REQUEST_URL
# The runner also passes each input as INPUT_<NAME> with its hyphen kept, a name sh cannot unset.
scrub="env -u INPUT_GITHUB-TOKEN"

# The runner owns the checkout and this container runs as root, so git would
# refuse it as "dubious ownership" when Proven reads the pull request's base.
git config --global --add safe.directory "$workspace" 2>/dev/null || true

# A file left by an earlier run must not be counted as this run's.
rm -f "$workspace/$json_file" "$workspace/$sarif_file" "$workspace/$text_file"
scratch=$(mktemp -d)
# The summary comment is the Markdown writer's output, kept in the scratch directory
# with the trace: it is the comment's body, not a result file of the workspace.
[ "$comment" != true ] || set -- "$@" "--md-output=$scratch/proven-comment.md"
cd "$scratch"
rc=0
$scrub "$proven" scan "$@" "$target" || rc=$?

# Outputs are written whatever the exit code, so a later step can read them.
# A review that could not run (2) or refused its configuration (7) wrote no JSON.
findings=0
unresolved=0
if [ -f "$workspace/$json_file" ]; then
  findings=$(grep -o '"unresolved": *\(true\|false\)' "$workspace/$json_file" | wc -l | tr -d ' ')
  unresolved=$(grep -o '"unresolved": *true' "$workspace/$json_file" | wc -l | tr -d ' ')
fi
{
  echo "results-file=$results_file"
  echo "json-file=$json_file"
  echo "sarif-file=$sarif_file"
  echo "findings-count=$findings"
  echo "unresolved-count=$unresolved"
} >>"$GITHUB_OUTPUT"

# comment: the Action writes ONE comment per pull request (the owner's order of
# 2026-10-25): the summary, edited in place on every run (post-comment.sh). It writes
# no line comments. A review that could not run wrote no Markdown, and there is nothing
# to post then. The reviewing run, the head commit and the pull request come from the
# event; a run that is not a pull request has none to post on.
if [ "$comment" = true ] && [ -f "$scratch/proven-comment.md" ]; then
  . "$here/post-comment.sh"
  event=${GITHUB_EVENT_PATH:-}
  pr=
  head=
  if [ -f "$event" ]; then
    pr=$(awk '{ buf = buf $0 } END { if (match(buf, /"number" *: *[0-9]+/)) { s = substr(buf, RSTART, RLENGTH); sub(/^"number" *: */, "", s); print s } }' "$event")
    head=$(awk '{ buf = buf $0 } END { if (match(buf, /"head" *: *\{/)) { rest = substr(buf, RSTART); if (match(rest, /"sha" *: *"[0-9a-f]+"/)) { s = substr(rest, RSTART, RLENGTH); sub(/^"sha" *: *"/, "", s); sub(/"$/, "", s); print s } } }' "$event")
  fi
  [ -n "$head" ] || head=${GITHUB_SHA:-}
  if [ -n "$pr" ] && [ -n "${GITHUB_API_URL:-}" ] && [ -n "${GITHUB_REPOSITORY:-}" ] && [ -n "$head" ]; then
    sumrc=0
    pc_post "$GITHUB_API_URL" "$GITHUB_REPOSITORY" "$pr" "${GITHUB_RUN_NUMBER:-0}" "$head" "$token" "$scratch/proven-comment.md" || sumrc=$?
    if [ "$sumrc" -ne 0 ]; then
      echo "proven action: posting the summary comment on the pull request failed. Check that github-token may write pull requests (pull-requests: write); on Gitea set PROVEN_COMMENT_ACCOUNT if the token cannot read /user." >&2
      [ "$rc" -ne 0 ] || rc=2
    fi
  else
    echo "proven action: no pull request, API address, repository or head commit in this run's environment; the summary comment was not posted." >&2
  fi
fi

rm -rf "$scratch"
exit "$rc"
