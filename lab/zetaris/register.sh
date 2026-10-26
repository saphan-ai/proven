#!/bin/sh
# Register the evidence definitions (lab/zetaris/evidence.sql) in the Zetaris platform and grant them to Proven's own
# user. Runs in the `zetaris-register` service of the lab (profile zetaris), as the judge's Zetaris administrator.
# The administrator's key is read from standard input (make zetaris pipes it in), never from the environment, never printed, never written. A statement that
# the platform answers with anything but 200 stops the run and is named with the platform's own message.
set -eu

: "${ZETARIS_REGISTER_URL:?ZETARIS_REGISTER_URL is not set}"
: "${ZETARIS_ORG_ID:?register: ZETARIS_ORG_ID is not set. NEXT: put the numeric organisation id in lab/.env (docs/LAB.md, path C).}"
ZETARIS_ADMIN_API_KEY=$(cat)
[ -n "$ZETARIS_ADMIN_API_KEY" ] || { echo "register: no administrator key arrived on standard input. NEXT: put ZETARIS_ADMIN_API_KEY in lab/.env and run make zetaris (docs/LAB.md, path C). It is used here and stored nowhere." >&2; exit 1; }
: "${ZETARIS_PROVEN_USER:?register: ZETARIS_PROVEN_USER is not set. NEXT: put the name of the non-admin user you made for Proven in lab/.env.}"

sql_file=${ZETARIS_EVIDENCE_SQL:-/zetaris/evidence.sql}

# send STATEMENT: one statement to /sql-editor/sqls/run. The statement is escaped for a JSON string.
send() {
  esc=$(printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g')
  id=$(cat /proc/sys/kernel/random/uuid)
  body=$(printf '{"queryId":"%s","sql":"%s","source":"SqlEditor","limit":100}' "$id" "$esc")
  code=$(curl -sS --max-time 60 -o /tmp/answer.json -w '%{http_code}' -X POST \
    -H "Authorization: Bearer $ZETARIS_ADMIN_API_KEY" \
    -H "X-Org-ID: $ZETARIS_ORG_ID" \
    -H "X-Request-ID: $(cat /proc/sys/kernel/random/uuid)" \
    -H 'Content-Type: application/json' \
    -d "$body" "$ZETARIS_REGISTER_URL/sql-editor/sqls/run")
  if [ "$code" != "200" ]; then
    echo "register: the platform answered HTTP $code to: $(printf '%s' "$1" | cut -c1-80)" >&2
    head -c 300 /tmp/answer.json >&2
    echo >&2
    return 1
  fi
  echo "register: ok: $(printf '%s' "$1" | cut -c1-60)"
}

# A definition that already exists is answered with an error on a second run; a kept platform is not a reason to stop
# for the statements that only make it, so those are named and passed over. The grants must succeed.
block=""
while IFS= read -r line || [ -n "$line" ]; do
  case $line in
    '--'*) continue ;;
    '')
      if [ -n "$block" ]; then
        send "$block" || true
        block=""
      fi
      ;;
    *) block="${block:+$block }$line" ;;
  esac
done <"$sql_file"
if [ -n "$block" ]; then
  send "$block" || true
fi

send "ASSIGN LIGHTNING DATABASE PX_EVIDENCE_DB TO USER '$ZETARIS_PROVEN_USER'" || exit 1
send "GRANT SELECT ON LIGHTNING DATABASE NAMESPACE PX_EVIDENCE_DB TO USER '$ZETARIS_PROVEN_USER'" || exit 1
send "GRANT SELECT ON VIEW NAMESPACE px_evidence TO USER '$ZETARIS_PROVEN_USER'" || exit 1
echo "register: the evidence definitions are in place and granted to $ZETARIS_PROVEN_USER"
