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
oldifs=$IFS
IFS=,
for glob in $include; do
  [ -z "$glob" ] || set -- "$@" "--include=$glob"
done
for glob in $exclude; do
  [ -z "$glob" ] || set -- "$@" "--exclude=$glob"
done
IFS=$oldifs

if [ "$comment" = true ]; then
  echo "proven action: comment is true; the pull request comment is posted by a separate step, not by this one." >&2
fi

# A file left by an earlier run must not be counted as this run's.
rm -f "$workspace/$json_file" "$workspace/$sarif_file" "$workspace/$text_file"
scratch=$(mktemp -d)
cd "$scratch"
rc=0
"$proven" scan "$@" "$target" || rc=$?

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

rm -rf "$scratch"
exit "$rc"
