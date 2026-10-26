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
reviewdog=${REVIEWDOG_BIN:-reviewdog}

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
reporter=
if [ "$comment" = true ]; then
  token=${INPUT_GITHUB_TOKEN:-}
  [ -n "$token" ] || fail "comment is true but github-token is empty; set github-token to a token that may write pull requests, and give the job pull-requests: write."
  # Gitea Actions sets GITHUB_SERVER_URL to its own address and GITEA_ACTIONS=true;
  # GitHub (and GitHub Enterprise Server) set no GITEA_ACTIONS.
  if [ "${GITEA_ACTIONS:-}" = true ]; then reporter=gitea-pr-review; else reporter=github-pr-review; fi
  command -v "$reviewdog" >/dev/null 2>&1 || fail "comment is true but reviewdog '$reviewdog' was not found in the image; rebuild action/Dockerfile."
fi

# Credentials never reach the scan: Proven reads the code under review, and that
# code must not see a token, whether comment is true or false. The posting token
# was copied into $token above (an unexported shell variable) and goes only to reviewdog.
unset INPUT_GITHUB_TOKEN GITHUB_TOKEN ACTIONS_RUNTIME_TOKEN ACTIONS_ID_TOKEN_REQUEST_TOKEN ACTIONS_ID_TOKEN_REQUEST_URL
# The runner also passes each input as INPUT_<NAME> with its hyphen kept, a name sh cannot unset.
scrub="env -u INPUT_GITHUB-TOKEN"

# reviewdog posts to a forge address of its own: GITHUB_API_URL (api.github.com when
# unset) or, on Gitea, GITEA_ADDRESS. With PROVEN_NETWORK=local that address is judged
# by the same rule as every other, before anything runs (ADR-0034): Proven refuses it by
# name when it is outside the organisation's network, and the step stops with its exit code.
if [ -n "$reporter" ]; then
  if [ "$reporter" = gitea-pr-review ]; then
    forge_setting=GITEA_ADDRESS
    forge_address=${GITHUB_SERVER_URL:-}
  else
    forge_setting=GITHUB_API_URL
    forge_address=${GITHUB_API_URL:-https://api.github.com}
  fi
  forge_rc=0
  $scrub "$proven" check-address "$forge_setting" "$forge_address" >/dev/null || forge_rc=$?
  [ "$forge_rc" -eq 0 ] || exit "$forge_rc"
fi
# The judged address must be reviewdog's only route (reviewdog 0.21.2, cmd/reviewdog/main.go
# and service/*/diff.go). GITHUB_API outranks GITHUB_API_URL; a proxy variable sends the
# connection to the proxy; a missing commit makes reviewdog run `git fetch` from the
# repository's web address, another host; and REVIEWDOG_TOKEN selects its hosted service.
# None of those addresses was judged, so with PROVEN_NETWORK=local reviewdog starts without
# them, and a run that needs the fetch fails and says so instead of fetching.
route_scrub=
if [ "$(printf %s "${PROVEN_NETWORK:-}" | tr -d ' \t')" = local ]; then
  route_scrub="env -u GITHUB_API -u HTTPS_PROXY -u https_proxy -u HTTP_PROXY -u http_proxy -u ALL_PROXY -u all_proxy -u REVIEWDOG_TOKEN REVIEWDOG_SKIP_GIT_FETCH=true REVIEWDOG_SKIP_DOGHOUSE=true"
fi

# The runner owns the checkout and this container runs as root, so git would
# refuse it as "dubious ownership" when Proven reads the pull request's base.
git config --global --add safe.directory "$workspace" 2>/dev/null || true

# A file left by an earlier run must not be counted as this run's.
rm -f "$workspace/$json_file" "$workspace/$sarif_file" "$workspace/$text_file"
scratch=$(mktemp -d)
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

# comment: post Proven's SARIF on the pull request through reviewdog (ADR-0020 §7).
# A review that could not run wrote no SARIF, and there is nothing to post then.
if [ "$comment" = true ] && [ -f "$workspace/$sarif_file" ]; then
  postrc=0
  (
    cd "$workspace"
    # reviewdog 0.21.2 reads the Gitea server from GITEA_ADDRESS and refuses
    # without it (cmd/reviewdog/main.go, giteaService); on GitHub and GitHub
    # Enterprise Server it reads the runner's own GITHUB_API_URL.
    if [ "$reporter" = gitea-pr-review ]; then export GITEA_ADDRESS="${GITHUB_SERVER_URL:-}"; fi
    REVIEWDOG_GITHUB_API_TOKEN=$token REVIEWDOG_GITEA_API_TOKEN=$token \
      $route_scrub "$reviewdog" -f=sarif -name=proven "-reporter=$reporter" -fail-level=none <"$sarif_file"
  ) || postrc=$?
  if [ "$postrc" -ne 0 ]; then
    echo "proven action: posting the review on the pull request failed: reviewdog ($reporter) exited $postrc. Check that github-token may write pull requests (pull-requests: write) and that this run is a pull request." >&2
    [ "$rc" -ne 0 ] || rc=2
  fi
fi

rm -rf "$scratch"
exit "$rc"
