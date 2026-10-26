#!/bin/sh
# up.sh: put Cloudflare Access in front of the judges' gateway, as code.
#
#   sh up.sh plan   SETTINGS    print what would change; change nothing
#   sh up.sh apply  SETTINGS    create what is missing, update what differs
#   sh up.sh rotate-runner SETTINGS
#                               issue the GitHub runner's service token a new secret
#
# Every object is found by its name first; only what is missing is created, so a
# second run changes nothing. The objects, in order: the named tunnel and its
# ingress to 127.0.0.1:<gateway_port>, the DNS CNAME for the one host, the
# e-mail one-time-PIN identity provider (only if the account has none), the
# Access application for that host, the runner's service token, and the two
# policies: the judges' allowlist (e-mail one-time PIN) and the runner's
# service-auth policy. It then writes the application's AUD tag and the team
# domain into the gateway's environment file and adds the matching identity rows
# to the gateway's settings file.
#
# The API token comes from CLOUDFLARE_API_TOKEN or the token_file setting and is
# never printed. The runner's secret goes once to a mode 0600 file, never to the
# screen or a log. Needs sh, curl and jq.
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck source=lib.sh
. "$HERE/lib.sh"

usage() {
	printf 'usage: sh up.sh plan|apply|rotate-runner SETTINGS\n' >&2
	exit 2
}
[ $# -eq 2 ] || usage
ROTATE=
case $1 in
plan) MODE=plan ;;
apply) MODE=apply ;;
rotate-runner)
	MODE=apply
	ROTATE=1
	;;
*) usage ;;
esac

need curl jq
load_settings "$2"
load_token
[ "$MODE" = plan ] || state_init

RUNNER_NAME=$S_name-runner
JUDGES_POLICY="$S_name judges"
RUNNER_POLICY="$S_name runner"

# ---- the token must be live before anything else ----

CF_QUIET=1
if cf GET /user/tokens/verify && [ "$(jq -r '.result.status // empty' "$RESP")" = active ]; then
	CF_QUIET=
elif cf GET "$ACC/tokens/verify" && [ "$(jq -r '.result.status // empty' "$RESP")" = active ]; then
	CF_QUIET=
else
	CF_QUIET=
	cf GET /user/tokens/verify || true
	die "the API token is not active for this account"
fi
say "ok:   the API token is active"

# ---- the zone: it must exist; this script never creates one ----

cfd GET "/zones?name=$(uri "$S_zone")&account.id=$S_account_id"
ZONE_ID=$(jq -r --arg n "$S_zone" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
[ -n "$ZONE_ID" ] || die "the zone $S_zone is not on this account; add the domain to Cloudflare first (this script does not create zones)"
say "ok:   zone $S_zone"

# ---- the runner's service token (also the whole of rotate-runner) ----

ST_ID= ST_CLIENT=
find_runner_token() {
	cfd GET "$ACC/access/service_tokens?per_page=100"
	ST_ID=$(jq -r --arg n "$RUNNER_NAME" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
	ST_CLIENT=$(jq -r --arg n "$RUNNER_NAME" '[.result[]? | select(.name == $n) | .client_id][0] // empty' "$RESP")
}

write_runner_secret() {
	write_private "$S_runner_secret_file" \
		"CF_ACCESS_CLIENT_ID=$(jq -r --arg c "$ST_CLIENT" '.result.client_id // $c' "$RESP")" \
		"CF_ACCESS_CLIENT_SECRET=$(jq -r '.result.client_secret' "$RESP")"
	say "ok:   the runner's client id and secret are in $S_runner_secret_file (mode 0600); move them into the repository secrets, then delete the file"
}

if [ -n "$ROTATE" ]; then
	find_runner_token
	[ -n "$ST_ID" ] || die "there is no service token named $RUNNER_NAME; run apply first"
	cfd POST "$ACC/access/service_tokens/$ST_ID/rotate"
	jq -e '.result.client_secret' "$RESP" >/dev/null 2>&1 || die "Cloudflare answered the rotation without a new secret"
	write_runner_secret
	exit 0
fi

# ---- the Zero Trust organisation: the team domain ----

TEAM=
CF_QUIET=1
if cf GET "$ACC/access/organizations"; then TEAM=$(jq -r '.result.auth_domain // empty' "$RESP"); fi
CF_QUIET=
if [ -z "$TEAM" ]; then
	[ -n "$S_team_name" ] || die "the account has no Zero Trust organisation; set team_name in the settings and run again, or create it once in the dashboard"
	if act "create the Zero Trust organisation $S_team_name"; then
		jq -n --arg n "$S_team_name" '{name: $n, auth_domain: ($n + ".cloudflareaccess.com")}' >"$(bodyfile org)"
		cfd POST "$ACC/access/organizations" "$(bodyfile org)"
		TEAM=$(jq -r '.result.auth_domain' "$RESP")
	else
		TEAM=$S_team_name.cloudflareaccess.com
	fi
else
	say "ok:   team domain $TEAM"
fi

# ---- the tunnel ----

cfd GET "$ACC/cfd_tunnel?name=$(uri "$S_name")&is_deleted=false"
T_ID=$(jq -r --arg n "$S_name" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
T_NEW=
if [ -n "$T_ID" ]; then
	say "ok:   tunnel $S_name exists"
elif act "create the tunnel $S_name"; then
	jq -n --arg n "$S_name" '{name: $n, config_src: "cloudflare"}' >"$(bodyfile tunnel)"
	cfd POST "$ACC/cfd_tunnel" "$(bodyfile tunnel)"
	T_ID=$(jq -r '.result.id' "$RESP")
	record tunnel "$T_ID"
	T_NEW=1
else
	T_ID=$PLANNED
fi

jq -n --arg h "$S_hostname" --arg s "http://127.0.0.1:$S_gateway_port" \
	'{config: {ingress: [{hostname: $h, service: $s}, {service: "http_status:404"}]}}' >"$(bodyfile ingress)"
WANT_INGRESS=$(jq -S -c '.config.ingress | map({hostname, service})' "$(bodyfile ingress)")
HAVE_INGRESS=
if [ "$T_ID" != "$PLANNED" ]; then
	cfd GET "$ACC/cfd_tunnel/$T_ID/configurations"
	HAVE_INGRESS=$(jq -S -c '(.result.config.ingress // []) | map({hostname, service})' "$RESP")
fi
if [ "$HAVE_INGRESS" = "$WANT_INGRESS" ]; then
	say "ok:   ingress $S_hostname -> 127.0.0.1:$S_gateway_port"
elif act "set the ingress of the tunnel: $S_hostname -> 127.0.0.1:$S_gateway_port, everything else 404"; then
	cfd PUT "$ACC/cfd_tunnel/$T_ID/configurations" "$(bodyfile ingress)"
fi

# ---- the DNS record: one CNAME for the one host ----

WANT_TARGET=$T_ID.cfargotunnel.com
cfd GET "/zones/$ZONE_ID/dns_records?name=$(uri "$S_hostname")&per_page=100"
DNS_ID=$(jq -r --arg n "$S_hostname" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
if [ -n "$DNS_ID" ]; then
	DNS_TYPE=$(jq -r '.result[0].type' "$RESP")
	DNS_COMMENT=$(jq -r '.result[0].comment // ""' "$RESP")
	DNS_CONTENT=$(jq -r '.result[0].content' "$RESP")
	DNS_PROXIED=$(jq -r '.result[0].proxied' "$RESP")
	[ "$DNS_TYPE" = CNAME ] || die "$S_hostname already has a $DNS_TYPE record that this script did not make; remove it or choose another hostname"
	[ "$DNS_COMMENT" = "$TAG" ] || die "$S_hostname already has a CNAME that this script did not make (its comment is not $TAG); remove it or choose another hostname"
	if [ "$DNS_CONTENT" = "$WANT_TARGET" ] && [ "$DNS_PROXIED" = true ]; then
		say "ok:   DNS $S_hostname -> the tunnel"
	elif act "point the DNS record $S_hostname at the tunnel"; then
		jq -n --arg t "$WANT_TARGET" '{content: $t, proxied: true}' >"$(bodyfile dnspatch)"
		cfd PATCH "/zones/$ZONE_ID/dns_records/$DNS_ID" "$(bodyfile dnspatch)"
	fi
elif act "create the DNS CNAME $S_hostname -> the tunnel"; then
	jq -n --arg n "$S_hostname" --arg t "$WANT_TARGET" --arg c "$TAG" \
		'{type: "CNAME", name: $n, content: $t, proxied: true, ttl: 1, comment: $c}' >"$(bodyfile dns)"
	cfd POST "/zones/$ZONE_ID/dns_records" "$(bodyfile dns)"
	record dns "$(jq -r '.result.id' "$RESP")"
fi

# ---- the e-mail one-time-PIN identity provider ----

cfd GET "$ACC/access/identity_providers"
IDP_ID=$(jq -r '[.result[]? | select(.type == "onetimepin") | .id][0] // empty' "$RESP")
if [ -n "$IDP_ID" ]; then
	say "ok:   the account has an e-mail one-time-PIN identity provider"
elif act "create the e-mail one-time-PIN identity provider"; then
	jq -n '{name: "One-time PIN", type: "onetimepin", config: {}}' >"$(bodyfile idp)"
	cfd POST "$ACC/access/identity_providers" "$(bodyfile idp)"
	IDP_ID=$(jq -r '.result.id' "$RESP")
	record idp "$IDP_ID"
else
	IDP_ID=$PLANNED
fi

# ---- the Access application ----

jq -n --arg n "$S_name" --arg d "$S_hostname" --arg s "$S_session_duration" --arg i "$IDP_ID" \
	'{name: $n, domain: $d, type: "self_hosted", session_duration: $s, allowed_idps: [$i],
	  auto_redirect_to_identity: false, app_launcher_visible: false}' >"$(bodyfile app)"
cfd GET "$ACC/access/apps?per_page=100"
APP_ID=$(jq -r --arg n "$S_name" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
AUD=
if [ -n "$APP_ID" ]; then
	AUD=$(jq -r --arg n "$S_name" '[.result[] | select(.name == $n) | .aud][0] // empty' "$RESP")
	HAVE_APP=$(jq -S -c --arg n "$S_name" '[.result[] | select(.name == $n)][0] | {domain, session_duration, allowed_idps: ((.allowed_idps // []) | sort)}' "$RESP")
	WANT_APP=$(jq -S -c '{domain, session_duration, allowed_idps: (.allowed_idps | sort)}' "$(bodyfile app)")
	if [ "$HAVE_APP" = "$WANT_APP" ]; then
		say "ok:   Access application $S_name for $S_hostname"
	elif act "update the Access application $S_name"; then
		cfd PUT "$ACC/access/apps/$APP_ID" "$(bodyfile app)"
	fi
elif act "create the Access application $S_name for $S_hostname"; then
	cfd POST "$ACC/access/apps" "$(bodyfile app)"
	APP_ID=$(jq -r '.result.id' "$RESP")
	AUD=$(jq -r '.result.aud' "$RESP")
	record app "$APP_ID"
else
	APP_ID=$PLANNED
fi

# ---- the runner's service token ----

find_runner_token
if [ -n "$ST_ID" ]; then
	say "ok:   service token $RUNNER_NAME exists"
	if [ ! -f "$S_runner_secret_file" ]; then
		say "note: its secret was shown once when it was made and is not on this machine; run 'sh up.sh rotate-runner' to issue a new one"
	fi
elif act "create the service token $RUNNER_NAME (valid $S_runner_token_duration)"; then
	jq -n --arg n "$RUNNER_NAME" --arg d "$S_runner_token_duration" '{name: $n, duration: $d}' >"$(bodyfile st)"
	cfd POST "$ACC/access/service_tokens" "$(bodyfile st)"
	ST_ID=$(jq -r '.result.id' "$RESP")
	ST_CLIENT=$(jq -r '.result.client_id' "$RESP")
	record service_token "$ST_ID"
	write_runner_secret
else
	ST_ID=$PLANNED
fi

# ---- the two policies ----

ensure_policy() { # name decision precedence include-json
	ep_name=$1 ep_decision=$2 ep_prec=$3 ep_include=$4
	jq -n --arg n "$ep_name" --arg d "$ep_decision" --argjson p "$ep_prec" --argjson i "$ep_include" \
		'{name: $n, decision: $d, precedence: $p, include: $i}' >"$(bodyfile policy)"
	ep_id=
	if [ "$APP_ID" != "$PLANNED" ]; then
		cfd GET "$ACC/access/apps/$APP_ID/policies"
		ep_id=$(jq -r --arg n "$ep_name" '[.result[]? | select(.name == $n) | .id][0] // empty' "$RESP")
	fi
	if [ -n "$ep_id" ]; then
		ep_have=$(jq -S -c --arg n "$ep_name" '[.result[] | select(.name == $n)][0] | {decision, include}' "$RESP")
		ep_want=$(jq -S -c '{decision, include}' "$(bodyfile policy)")
		if [ "$ep_have" = "$ep_want" ]; then
			say "ok:   policy $ep_name"
		elif act "update the policy $ep_name"; then
			cfd PUT "$ACC/access/apps/$APP_ID/policies/$ep_id" "$(bodyfile policy)"
		fi
	elif act "create the policy $ep_name ($ep_decision)"; then
		cfd POST "$ACC/access/apps/$APP_ID/policies" "$(bodyfile policy)"
		record policy "$(jq -r '.result.id' "$RESP")"
	fi
}

JUDGES_INCLUDE=$(printf '%s' "$S_judges" | tr ',' '\n' | jq -R -s -c '
	split("\n") | map(gsub("^\\s+|\\s+$"; "") | ascii_downcase) | map(select(length > 0))
	| map(if startswith("*@") then {email_domain: {domain: .[2:]}} else {email: {email: .}} end)')
ensure_policy "$JUDGES_POLICY" allow 1 "$JUDGES_INCLUDE"
RUNNER_INCLUDE=$(jq -n -c --arg t "$ST_ID" '[{service_token: {token_id: $t}}]')
ensure_policy "$RUNNER_POLICY" non_identity 2 "$RUNNER_INCLUDE"

# ---- the gateway's side: the environment file and the identity rows ----

WANT_ENV="JUDGES_GATEWAY_AUTH=access
JUDGES_GATEWAY_ACCESS_TEAM=$TEAM
JUDGES_GATEWAY_ACCESS_AUD=$AUD"
if [ -f "$S_gateway_env_file" ] && [ "$(cat "$S_gateway_env_file")" = "$WANT_ENV" ]; then
	say "ok:   the gateway's Access settings in $S_gateway_env_file"
elif act "write the AUD tag and team domain to $S_gateway_env_file"; then
	write_private "$S_gateway_env_file" "$WANT_ENV"
fi
if [ ! -f "$S_tunnel_token_file" ] || [ -n "$T_NEW" ]; then
	if act "write the connector's run token to $S_tunnel_token_file (mode 0600)"; then
		cfd GET "$ACC/cfd_tunnel/$T_ID/token"
		write_private "$S_tunnel_token_file" "$(jq -r '.result' "$RESP")"
	fi
fi

sync_identity() { # match per-minute per-day
	if [ -z "$S_gateway_bin" ] || [ -z "$S_gateway_settings_file" ]; then return 0; fi
	[ -n "$S_identities_expire" ] || die "settings: identities_expire (an RFC 3339 time) is needed when gateway_bin is set"
	if [ -f "$S_gateway_settings_file" ] &&
		JUDGES_GATEWAY_TOKENS_FILE=$S_gateway_settings_file "$S_gateway_bin" identity list | awk 'NR > 1 {print $1}' | grep -qxF "$1"; then
		say "ok:   gateway identity $1"
	elif act "add the gateway identity $1"; then
		JUDGES_GATEWAY_TOKENS_FILE=$S_gateway_settings_file "$S_gateway_bin" identity add \
			--match "$1" --expires "$S_identities_expire" --per-minute "$2" --per-day "$3" >/dev/null
	fi
}
for j in $(printf '%s' "$S_judges" | tr ',' ' ' | tr 'A-Z' 'a-z'); do
	sync_identity "$j" "$S_judge_per_minute" "$S_judge_per_day"
done
if [ -n "$ST_CLIENT" ]; then
	sync_identity "$ST_CLIENT" "$S_runner_per_minute" "$S_runner_per_day"
fi

say "done: $MODE finished"
if [ -n "$AUD" ]; then say "      team $TEAM, application AUD $AUD, host $S_hostname"; fi
