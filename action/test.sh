#!/bin/sh
# Tests the Action without GitHub: action.yml against ADR-0020 §3, and
# entrypoint.sh against the sample fixture with a temporary GITHUB_OUTPUT and
# GITHUB_WORKSPACE. Run from anywhere:  sh action/test.sh
# PROVEN_BIN names a built binary; without it the script builds one.
set -u

here=$(cd "$(dirname "$0")" && pwd)
root=$(dirname "$here")
fixture=$root/testdata/fixture/discount
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if [ -z "${PROVEN_BIN:-}" ]; then
  (cd "$root" && go build -o "$tmp/proven" ./cmd/proven) || { echo "FAIL build: go build ./cmd/proven" >&2; exit 2; }
  PROVEN_BIN=$tmp/proven
fi
export PROVEN_BIN

pass=0
failed=0
check() {
  if [ "$2" = "$3" ]; then
    pass=$((pass + 1))
    echo "ok   $1"
  else
    failed=$((failed + 1))
    echo "FAIL $1: want '$2', got '$3'"
  fi
}

# --- action.yml: the names of ADR-0020 §3, and nothing else -------------------
names() { # section file: the keys one indent under a top-level section
  awk -v s="$1:" '
    /^[a-z]/ { in_s = ($0 == s) ; next }
    in_s && /^  [a-z-]+:/ { sub(/^  /, ""); sub(/:.*/, ""); print }
  ' "$2" | sort | tr '\n' ' '
}
check "action.yml inputs" \
  "baseline-commit comment config exclude github-token include output-format sample-mode severity strict target timeout " \
  "$(names inputs "$root/action.yml")"
check "action.yml outputs" \
  "findings-count json-file results-file sarif-file unresolved-count " \
  "$(names outputs "$root/action.yml")"
check "action.yml runs docker" "1" "$(grep -c '^  using: docker$' "$root/action.yml")"
check "action.yml strict default" "1" "$(awk '/^  strict:/ {f=1} f && /default: "true"/ {print; exit}' "$root/action.yml" | wc -l | tr -d ' ')"

# --- entrypoint.sh ------------------------------------------------------------
# run NAME VAR=value...: a fresh workspace, a fresh GITHUB_OUTPUT; sets $rc, $ws, $out.
run() {
  name=$1
  shift
  ws=$tmp/$name/ws
  out=$tmp/$name/out
  mkdir -p "$ws"
  : >"$out"
  # RUNDIR is the script's working directory: the runner starts the container in the checkout.
  (
    cd "${RUNDIR:-.}" &&
      env -i PATH="$PATH" HOME="$tmp" TMPDIR="$tmp" PROVEN_BIN="$PROVEN_BIN" \
        GITHUB_WORKSPACE="$ws" GITHUB_OUTPUT="$out" \
        INPUT_TARGET="$fixture" INPUT_SAMPLE_MODE=true "$@" \
        sh "${ENTRYPOINT:-$here/entrypoint.sh}" >"$tmp/$name/stdout" 2>"$tmp/$name/stderr"
  )
  rc=$?
}
output() { sed -n "s/^$1=//p" "$out"; }

run default
check "default: exit code (strict, findings)" 1 "$rc"
check "default: results-file" proven-results.sarif "$(output results-file)"
check "default: json-file" proven-results.json "$(output json-file)"
check "default: sarif-file" proven-results.sarif "$(output sarif-file)"
check "default: findings-count" 4 "$(output findings-count)"
check "default: unresolved-count" 0 "$(output unresolved-count)"
check "default: json file written in the workspace" 1 "$(test -s "$ws/proven-results.json" && echo 1)"
check "default: sarif file is SARIF 2.1.0" 1 "$(grep -c '"version": *"2.1.0"' "$ws/proven-results.sarif")"
check "default: workspace holds only the results" "proven-results.json proven-results.sarif " "$(ls "$ws" | tr '\n' ' ')"
check "default: stdout is the text review" 1 "$(grep -c 'finding F1' "$tmp/default/stdout")"

run lenient INPUT_STRICT=false
check "strict=false: exit code" 0 "$rc"
check "strict=false: findings-count" 4 "$(output findings-count)"

run asjson INPUT_OUTPUT_FORMAT=json INPUT_STRICT=false
check "output-format=json: results-file" proven-results.json "$(output results-file)"
check "output-format=json: stdout is JSON" 1 "$(grep -c '"findings": \[' "$tmp/asjson/stdout")"

run assarif INPUT_OUTPUT_FORMAT=sarif INPUT_STRICT=false
check "output-format=sarif: stdout is SARIF" 1 "$(grep -c '"version": *"2.1.0"' "$tmp/assarif/stdout")"

run astext INPUT_OUTPUT_FORMAT=text INPUT_STRICT=false
check "output-format=text: results-file" proven-results.txt "$(output results-file)"
check "output-format=text: text file has the review" 1 "$(grep -c 'finding F1' "$ws/proven-results.txt")"

run severe INPUT_SEVERITY=ERROR INPUT_STRICT=false
check "severity=ERROR: every fixture finding is ERROR, so all 4 stay" 4 "$(output findings-count)"

# The fixture cannot show a flag whose effect it does not exercise, so a stub
# binary records the command line each input produces.
cat >"$tmp/stub.sh" <<'EOF'
#!/bin/sh
for a in "$@"; do printf '%s\n' "$a"; done >"$STUB_ARGS"
for a in "$@"; do
  case "$a" in --json-output=*) printf '{"findings": [\n{"unresolved": true},\n{"unresolved": false},\n{"unresolved": true}\n]}\n' >"${a#--json-output=}" ;; esac
done
exit "${STUB_RC:-0}"
EOF
chmod +x "$tmp/stub.sh"
stubrun() {
  name=$1
  shift
  STUB_ARGS=$tmp/$name.args
  export STUB_ARGS
  run "$name" PROVEN_BIN="$tmp/stub.sh" STUB_ARGS="$STUB_ARGS" "$@"
  args=$(tr '\n' ' ' <"$STUB_ARGS")
}
contains() { # 1 when the recorded command line has this substring anywhere
  case "$args" in
    *"$1"*) echo 1 ;;
    *) echo 0 ;;
  esac
}
hasarg() { case " $args" in *" $1 "*) echo 1 ;; *) echo 0 ;; esac }

stubrun mapped INPUT_CONFIG=p.yaml INPUT_SEVERITY=WARNING INPUT_INCLUDE='a/*,b/*' INPUT_EXCLUDE='c/*' \
  INPUT_TIMEOUT=30 INPUT_BASELINE_COMMIT=abc123 INPUT_STRICT=true
check "maps config" 1 "$(hasarg "--config=$ws/p.yaml")"
check "maps severity" 1 "$(hasarg --severity=WARNING)"
check "maps include (first)" 1 "$(hasarg '--include=a/*')"
check "maps include (second)" 1 "$(hasarg '--include=b/*')"
check "maps exclude" 1 "$(hasarg '--exclude=c/*')"
check "maps timeout" 1 "$(hasarg --timeout=30)"
check "maps baseline-commit" 1 "$(hasarg --baseline-commit=abc123)"
check "maps sample-mode=true" 1 "$(hasarg --sample)"
check "maps strict=true" 1 "$(hasarg --error)"
check "counts findings from the JSON" 3 "$(output findings-count)"
check "counts unresolved from the JSON" 2 "$(output unresolved-count)"
check "the target is the last argument" "$fixture" "$(tail -n 1 "$STUB_ARGS")"

stubrun unmapped INPUT_SAMPLE_MODE=false INPUT_STRICT=false
check "maps strict=false" 1 "$(hasarg --no-error)"
check "sample-mode=false adds no --sample" 0 "$(hasarg --sample)"
check "empty baseline adds no --baseline-commit" 0 "$(contains --baseline-commit)"
check "empty timeout adds no --timeout" 0 "$(contains --timeout)"

stubrun failed STUB_RC=2 INPUT_STRICT=true
check "Proven's exit code 2 is the step's" 2 "$rc"

run badformat INPUT_OUTPUT_FORMAT=xml
check "output-format=xml: refused, exit 2" 2 "$rc"
check "output-format=xml: nothing written to GITHUB_OUTPUT" 0 "$(wc -c <"$out" | tr -d ' ')"

run badbool INPUT_STRICT=maybe
check "strict=maybe: refused, exit 2" 2 "$rc"

run noconfig INPUT_CONFIG=does-not-exist.yaml
check "missing config: Proven's refusal is the step's exit code (7)" 7 "$rc"
check "missing config: counts are 0" "0 0" "$(output findings-count) $(output unresolved-count)"

# A glob is Proven's to match: a path in the working directory that fits it must not
# turn it into other words. The runner starts the container in the checkout.
mkdir -p "$tmp/cwd/vendor/a" "$tmp/cwd/vendor/b"
RUNDIR=$tmp/cwd stubrun globs INPUT_EXCLUDE='vendor/*' INPUT_INCLUDE='*,vendor/*'
check "exclude glob reaches Proven verbatim" 1 "$(hasarg '--exclude=vendor/*')"
check "exclude glob is not expanded against the working directory" 0 "$(contains '--exclude=vendor/a')"
check "include glob '*' reaches Proven verbatim" 1 "$(hasarg '--include=*')"

# comment: reviewdog posts Proven's SARIF. A stub records its arguments, its
# stdin, and the token and server it was given.
cat >"$tmp/reviewdog.sh" <<'EOF'
#!/bin/sh
for a in "$@"; do printf '%s\n' "$a"; done >"$RD_ARGS"
cat >"$RD_STDIN"
printf '%s|%s\n' "${REVIEWDOG_GITHUB_API_TOKEN:-}" "${REVIEWDOG_GITEA_API_TOKEN:-}" >"$RD_TOKEN"
exit "${RD_RC:-0}"
EOF
chmod +x "$tmp/reviewdog.sh"
rdrun() {
  name=$1
  shift
  RD_ARGS=$tmp/$name.rdargs
  RD_STDIN=$tmp/$name.rdstdin
  RD_TOKEN=$tmp/$name.rdtoken
  rm -f "$RD_ARGS" "$RD_STDIN" "$RD_TOKEN"
  run "$name" REVIEWDOG_BIN="$tmp/reviewdog.sh" RD_ARGS="$RD_ARGS" RD_STDIN="$RD_STDIN" RD_TOKEN="$RD_TOKEN" "$@"
  args=
  [ ! -f "$RD_ARGS" ] || args=$(tr '\n' ' ' <"$RD_ARGS")
}

rdrun postgh INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false
check "comment=true: exit 0" 0 "$rc"
check "comment=true: reviewdog reads SARIF" 1 "$(hasarg -f=sarif)"
check "comment=true: reporter on GitHub" 1 "$(hasarg -reporter=github-pr-review)"
check "comment=true: reviewdog is given Proven's SARIF, byte for byte" 1 "$(cmp -s "$ws/proven-results.sarif" "$RD_STDIN" && echo 1)"
check "comment=true: the token reaches reviewdog" "tok-1|tok-1" "$(cat "$RD_TOKEN")"

rdrun postgitea INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false GITEA_ACTIONS=true GITHUB_SERVER_URL=https://gitea.example
check "comment=true on Gitea: reporter" 1 "$(hasarg -reporter=gitea-pr-review)"

rdrun poststrict INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=true
check "comment=true, strict, findings: still posted" 1 "$(hasarg -reporter=github-pr-review)"
check "comment=true, strict, findings: exit is Proven's 1" 1 "$rc"

rdrun postfail INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false RD_RC=1
check "comment=true, reviewdog fails: the step fails (2)" 2 "$rc"
check "comment=true, reviewdog fails: the message names the post" 1 "$(grep -c 'posting the review on the pull request failed' "$tmp/postfail/stderr")"
check "comment=true, reviewdog fails: outputs still written" 4 "$(output findings-count)"

rdrun notoken INPUT_COMMENT=true INPUT_GITHUB_TOKEN=
check "comment=true, no token: refused, exit 2" 2 "$rc"
check "comment=true, no token: names github-token" 1 "$(grep -c 'github-token is empty' "$tmp/notoken/stderr")"
check "comment=true, no token: Proven did not run" 0 "$(test -e "$tmp/notoken/ws/proven-results.json" && echo 1 || echo 0)"

rdrun nocomment INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false
check "comment=false: reviewdog not run" 0 "$(test -e "$RD_ARGS" && echo 1 || echo 0)"

rdrun nocommentnotoken INPUT_GITHUB_TOKEN= INPUT_STRICT=false
check "comment=false: no token needed" 0 "$rc"

rdrun nofail INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_CONFIG=does-not-exist.yaml
check "comment=true, Proven refused (7): nothing posted, exit 7" "7 0" "$rc $(test -e "$RD_ARGS" && echo 1 || echo 0)"

echo "passed $pass, failed $failed"
[ "$failed" -eq 0 ]
