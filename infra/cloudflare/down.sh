#!/bin/sh
# down.sh: take away what up.sh made, and only that, in reverse order.
#
#   sh down.sh plan   SETTINGS [--by-name]   print what would be deleted
#   sh down.sh apply  SETTINGS [--by-name]   delete it
#   sh down.sh runner SETTINGS [--by-name]   delete only the runner's service token and policy
#
# An object is deleted only when its name matches the settings AND up.sh
# recorded its id in the state file when it created it. An object that was there
# before up.sh ran is not in the state file, so it stays. A DNS record is also
# deleted only if its comment carries the tag up.sh gives it. When the state
# file is lost, --by-name accepts the name (and the tag) alone.
#
# The gateway's own identity rows are not touched: revoke them with
# 'judges-gateway identity revoke'. The API token is read as up.sh reads it and is
# never printed.
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck source=lib.sh
. "$HERE/lib.sh"

usage() {
	printf 'usage: sh down.sh plan|apply|runner SETTINGS [--by-name]\n' >&2
	exit 2
}
[ $# -ge 2 ] && [ $# -le 3 ] || usage
ONLY_RUNNER=
case $1 in
plan) MODE=plan ;;
apply) MODE=apply ;;
runner)
	MODE=apply
	ONLY_RUNNER=1
	;;
*) usage ;;
esac
BY_NAME=
if [ $# -eq 3 ]; then
	[ "$3" = --by-name ] || usage
	BY_NAME=1
fi

need curl jq
load_settings "$2"
load_token
RUNNER_NAME=$S_name-runner
JUDGES_POLICY="$S_name judges"
RUNNER_POLICY="$S_name runner"

if [ ! -f "$STATE" ] && [ -z "$BY_NAME" ]; then
	die "there is no state file at $STATE, so nothing is known to be made by up.sh; use --by-name to accept the names alone"
fi

# mine KIND ID: true when up.sh made this object.
mine() { [ -n "$BY_NAME" ] || recorded "$1" "$2"; }

remove() { # description path
	if act "delete $1"; then cfd DELETE "$2"; fi
}

# ---- the policies of the application ----

cfd GET "$ACC/access/apps?per_page=100"
APP_ID=$(jq -r --arg n "$S_name" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")

if [ -n "$APP_ID" ]; then
	cfd GET "$ACC/access/apps/$APP_ID/policies"
	cp_json=$(cat "$RESP")
	for pname in "$RUNNER_POLICY" "$JUDGES_POLICY"; do
		[ -z "$ONLY_RUNNER" ] || [ "$pname" = "$RUNNER_POLICY" ] || continue
		pid=$(printf '%s' "$cp_json" | jq -r --arg n "$pname" '[.result[]? | select(.name == $n) | .id][0] // empty')
		if [ -z "$pid" ]; then
			say "ok:   no policy $pname"
		elif mine policy "$pid"; then
			remove "the policy $pname" "$ACC/access/apps/$APP_ID/policies/$pid"
			[ "$MODE" = plan ] || unrecord policy "$pid"
		else
			say "keep: policy $pname was not made by up.sh"
		fi
	done
else
	say "ok:   no Access application $S_name"
fi

# ---- the runner's service token ----

cfd GET "$ACC/access/service_tokens?per_page=100"
ST_ID=$(jq -r --arg n "$RUNNER_NAME" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
if [ -z "$ST_ID" ]; then
	say "ok:   no service token $RUNNER_NAME"
elif mine service_token "$ST_ID"; then
	remove "the service token $RUNNER_NAME" "$ACC/access/service_tokens/$ST_ID"
	if [ "$MODE" = apply ]; then
		unrecord service_token "$ST_ID"
		rm -f "$S_runner_secret_file"
	fi
else
	say "keep: service token $RUNNER_NAME was not made by up.sh"
fi

if [ -n "$ONLY_RUNNER" ]; then
	say "done: $MODE finished (the runner only)"
	exit 0
fi

# ---- the application ----

if [ -n "$APP_ID" ]; then
	if mine app "$APP_ID"; then
		remove "the Access application $S_name" "$ACC/access/apps/$APP_ID"
		[ "$MODE" = plan ] || unrecord app "$APP_ID"
	else
		say "keep: Access application $S_name was not made by up.sh"
	fi
fi

# ---- the DNS record ----

cfd GET "/zones?name=$(uri "$S_zone")&account.id=$S_account_id"
ZONE_ID=$(jq -r --arg n "$S_zone" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
if [ -n "$ZONE_ID" ]; then
	cfd GET "/zones/$ZONE_ID/dns_records?name=$(uri "$S_hostname")&per_page=100"
	DNS_ID=$(jq -r --arg n "$S_hostname" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
	DNS_COMMENT=$(jq -r '.result[0].comment // ""' "$RESP")
	if [ -z "$DNS_ID" ]; then
		say "ok:   no DNS record $S_hostname"
	elif [ "$DNS_COMMENT" = "$TAG" ] && mine dns "$DNS_ID"; then
		remove "the DNS record $S_hostname" "/zones/$ZONE_ID/dns_records/$DNS_ID"
		[ "$MODE" = plan ] || unrecord dns "$DNS_ID"
	else
		say "keep: DNS record $S_hostname was not made by up.sh (state or tag differs)"
	fi
else
	say "keep: the zone $S_zone is not on this account; its DNS record is untouched"
fi

# ---- the tunnel ----

cfd GET "$ACC/cfd_tunnel?name=$(uri "$S_name")&is_deleted=false"
T_ID=$(jq -r --arg n "$S_name" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
if [ -z "$T_ID" ]; then
	say "ok:   no tunnel $S_name"
elif mine tunnel "$T_ID"; then
	if [ "$MODE" = apply ]; then
		# A tunnel with live connections cannot be deleted: clear them first.
		CF_QUIET=1
		cf DELETE "$ACC/cfd_tunnel/$T_ID/connections" || true
		CF_QUIET=
	fi
	remove "the tunnel $S_name" "$ACC/cfd_tunnel/$T_ID"
	if [ "$MODE" = apply ]; then
		unrecord tunnel "$T_ID"
		rm -f "$S_tunnel_token_file"
	fi
else
	say "keep: tunnel $S_name was not made by up.sh"
fi

# ---- the identity provider: only if up.sh made it ----

cfd GET "$ACC/access/identity_providers"
for idp in $(jq -r '.result[]? | select(.type == "onetimepin") | .id' "$RESP"); do
	if recorded idp "$idp"; then
		remove "the e-mail one-time-PIN identity provider" "$ACC/access/identity_providers/$idp"
		[ "$MODE" = plan ] || unrecord idp "$idp"
	else
		say "keep: the e-mail one-time-PIN identity provider was not made by up.sh"
	fi
done

if [ "$MODE" = apply ]; then rm -f "$S_gateway_env_file"; fi
say "done: $MODE finished"
