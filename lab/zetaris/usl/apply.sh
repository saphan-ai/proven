#!/bin/sh
# apply.sh — create or replace Proven's semantic layer in Zetaris, run its gate, and print the settings to use.
# docs/ZETARIS_USL.md is the runbook. Run it as the Zetaris administrator: registration is refused for Proven's
# own key (ADR-0023, point 2).
#
# Reads the Zetaris address, the numeric organisation id and the key from the environment, with the names
# docs/CONFIGURATION.md lists. Never from an argument or a file. It prints none of the three.
# DRY=1 prints every statement it would send and sends nothing.
#
# Exit codes: 0 applied and the gate passed (or DRY) · 2 a setting is missing or malformed ·
#             3 Zetaris refused a call · 4 the gate returned rows · 5 the gate's answer was not exactly one count (count.py).
# The key and the organisation id are removed from every refusal and transport error before it is printed.
set -eu

here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
layer="$here/layer.sql"
gate="$here/gate.sql"

# The settings to point Proven at the layer once it is applied. One place, so the output and the docs agree.
settings='PROVEN_OSV_TABLE=px_usl.px_package
PROVEN_KEV_VIEW=px_usl.px_kev_cve
PROVEN_EPSS_VIEW=px_usl.px_epss_score
PROVEN_OSV_ID_VIEW=px_usl.px_advisory'

die() { code=$1; shift; printf '%s\n' "$*" >&2; exit "$code"; }

# statements prints one line per statement of a definitions file: "N <sql>", or "T <sql>" when the refusal of that
# statement is tolerated. A whole-line comment is dropped; the "tolerate:" comment marks the next statement.
statements() {
  awk '
    /^[ \t]*--/ { if ($0 ~ /^[ \t]*-- tolerate:/) tol = 1; next }
    /^[ \t]*$/ { next }
    {
      line = $0; sub(/^[ \t]+/, "", line)
      buf = (buf == "" ? line : buf " " line)
      if (line ~ /;[ \t]*$/) {
        sub(/;[ \t]*$/, "", buf)
        print (tol ? "T " : "N ") buf
        buf = ""; tol = 0
      }
    }
    END { if (buf != "") { print "unterminated statement" > "/dev/stderr"; exit 1 } }
  ' "$1"
}

# gate_sql prints the gate's SELECT as one line, and gate_rule its name.
gate_sql() { grep -v '^[ 	]*--' "$gate" | tr '\n' ' ' | sed 's/[ 	]*$//'; }
gate_rule() { sed -n 's/^-- rule:[ 	]*//p' "$gate" | head -n 1; }

rule=$(gate_rule)
[ -n "$rule" ] || die 2 "apply.sh: gate.sql names no rule (its first comment is '-- rule: <name>')"
rows_sql=$(gate_sql)
count_sql="SELECT COUNT(*) AS violations FROM ($rows_sql) gate_rows"

if [ -n "${DRY:-}" ]; then
  n=0
  statements "$layer" | while IFS= read -r line; do
    n=$((n + 1))
    printf 'DRY statement %s: %s\n' "$n" "${line#? }"
  done
  printf 'DRY gate %s: %s\n' "$rule" "$count_sql"
  printf '%s\n' 'DRY: nothing was sent. The settings below are printed after a real apply passes the gate:'
  printf '%s\n' "$settings"
  exit 0
fi

: "${ZETARIS_REST_URL:=}" "${ZETARIS_ORG_ID:=}" "${ZETARIS_API_KEY:=}"
missing=
[ -n "$ZETARIS_REST_URL" ] || missing="$missing ZETARIS_REST_URL"
[ -n "$ZETARIS_ORG_ID" ] || missing="$missing ZETARIS_ORG_ID"
[ -n "$ZETARIS_API_KEY" ] || missing="$missing ZETARIS_API_KEY"
[ -z "$missing" ] || die 2 "apply.sh: not set in the environment:$missing. Nothing was sent."
case $ZETARIS_REST_URL in
  http://*|https://*) ;;
  *) die 2 "apply.sh: the Zetaris address is not an http or https URL. Nothing was sent." ;;
esac
case $ZETARIS_REST_URL in
  *[!A-Za-z0-9._~:/%@-]*) die 2 "apply.sh: the Zetaris address holds a character outside a plain URL. Nothing was sent." ;;
esac
case $ZETARIS_ORG_ID in
  *[!0-9]*) die 2 "apply.sh: the organisation id is not numeric. Nothing was sent." ;;
esac
# The key is written into a curl config on a pipe, inside double quotes: refuse a byte that would end the line or the quote.
case $ZETARIS_API_KEY in
  *[!A-Za-z0-9._~+/=:-]*) die 2 "apply.sh: the Zetaris key holds a character outside letters, digits and . _ ~ + / = : -. Nothing was sent." ;;
esac
command -v python3 > /dev/null 2>&1 || die 2 "apply.sh: python3 is not on the path; it reads the gate's answer. Nothing was sent."
base=${ZETARIS_REST_URL%/}

work=$(mktemp -d "${TMPDIR:-/tmp}/zetaris-usl.XXXXXX")
trap 'rm -rf "$work"' EXIT INT TERM
chmod 700 "$work"

uuid() {
  hex=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
  v=$(printf '%s' "$hex" | cut -c17-17)
  case $v in 0|4|8|c) v=8 ;; 1|5|9|d) v=9 ;; 2|6|a|e) v=a ;; *) v=b ;; esac
  printf '%s-%s-4%s-%s%s-%s' "$(printf '%s' "$hex" | cut -c1-8)" "$(printf '%s' "$hex" | cut -c9-12)" \
    "$(printf '%s' "$hex" | cut -c14-16)" "$v" "$(printf '%s' "$hex" | cut -c18-20)" "$(printf '%s' "$hex" | cut -c21-32)"
}

# redact prints stdin with every occurrence of the key and of the organisation id removed, and only printable bytes.
# Every diagnostic goes through it: a remote error may echo either one.
redact() {
  KEYTEXT=$ZETARIS_API_KEY ORGTEXT=$ZETARIS_ORG_ID awk '
    function scrub(line, needle, mark,   i) {
      while (needle != "" && (i = index(line, needle)) > 0) line = substr(line, 1, i - 1) mark substr(line, i + length(needle))
      return line
    }
    { print scrub(scrub($0, ENVIRON["KEYTEXT"], "[key]"), ENVIRON["ORGTEXT"], "[org]") }
  ' | tr -cd '[:print:]\n'
}

# send posts one statement to the SQL editor and leaves the status in $work/status and the body in $work/body.
# The key goes to curl on a pipe as a config, never on the command line.
send() {
  esc=$(printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g')
  printf '{"queryId":"%s","sql":"%s","source":"SqlEditor","limit":100}' "$(uuid)" "$esc" > "$work/request"
  printf 'url = "%s/sql-editor/sqls/run"\nheader = "Authorization: Bearer %s"\nheader = "X-Org-ID: %s"\nheader = "X-Request-ID: %s"\nheader = "Content-Type: application/json"\n' \
    "$base" "$ZETARIS_API_KEY" "$ZETARIS_ORG_ID" "$(uuid)" |
    curl -sS -K - -X POST --max-time 120 --data-binary "@$work/request" -o "$work/body" -w '%{http_code}' > "$work/status" 2> "$work/err" || {
      printf 'apply.sh: the call to Zetaris failed: %s\n' "$(redact < "$work/err")" >&2
      return 1
    }
}

n=0
statements "$layer" > "$work/statements"
total=$(wc -l < "$work/statements" | tr -d ' ')
while IFS= read -r line; do
  n=$((n + 1))
  kind=${line%% *}
  sql=${line#? }
  send "$sql" || die 3 "apply.sh: statement $n of $total was not delivered. The layer may be half applied; run apply.sh again."
  status=$(cat "$work/status")
  if [ "$status" != 200 ]; then
    if [ "$kind" = T ] && redact < "$work/body" | grep -qi 'exist'; then
      printf 'statement %s of %s: refused as expected (%s), going on\n' "$n" "$total" "$status"
      continue
    fi
    printf 'apply.sh: Zetaris refused statement %s of %s (HTTP %s): %s\n' "$n" "$total" "$status" "$(printf '%s' "$sql" | cut -c1-80)" >&2
    redact < "$work/body" | head -c 2000 >&2
    printf '\n' >&2
    die 3 "apply.sh: stopped. The layer may be half applied; fix the refusal and run apply.sh again."
  fi
  printf 'statement %s of %s: ok\n' "$n" "$total"
done < "$work/statements"

# The gate. An answer that cannot be read is a refusal: no answer is never a pass.
send "$count_sql" || die 3 "apply.sh: the gate query was not delivered. Proven is not pointed at the layer."
status=$(cat "$work/status")
if [ "$status" != 200 ]; then
  printf 'apply.sh: Zetaris refused the gate query (HTTP %s):\n' "$status" >&2
  redact < "$work/body" | head -c 2000 >&2
  printf '\n' >&2
  die 3 "apply.sh: Proven is not pointed at the layer."
fi
count=$(python3 -I "$here/count.py" "$work/body" violations 2> "$work/shape") ||
  die 5 "apply.sh: the answer of gate $rule could not be read ($(redact < "$work/shape")), so the gate does not pass. Proven is not pointed at the layer."
if [ "$count" -gt 0 ]; then
  printf 'apply.sh: REFUSED. Gate %s returned %s violating row(s). Proven is not pointed at the layer.\n' "$rule" "$count" >&2
  if send "$rows_sql LIMIT 20" && [ "$(cat "$work/status")" = 200 ]; then
    printf 'The first rows (at most 20):\n' >&2
    redact < "$work/body" | head -c 4000 >&2
    printf '\n' >&2
  fi
  printf '%s\n' "What to do: docs/ZETARIS_USL.md, \"When the gate refuses\"." >&2
  exit 4
fi
printf 'gate %s: 0 violating rows\n' "$rule"
printf '%s\n' 'The layer is applied. Give Proven these settings, and the administrator grants Proven read on the container (docs/ZETARIS_USL.md):'
printf '%s\n' "$settings"
