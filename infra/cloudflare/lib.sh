# lib.sh: the shared part of up.sh, down.sh and verify.sh. POSIX sh; sourced,
# never run. Needs curl and jq.
#
# The API token is read once, from CLOUDFLARE_API_TOKEN or from the file the
# settings name (token_file). It is never printed, never put on a command line
# and never written to a file: curl reads it from its configuration on stdin.

umask 077

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

say() { printf '%s\n' "$*"; }

need() {
	for need_c in "$@"; do
		command -v "$need_c" >/dev/null 2>&1 || die "$need_c is needed and is not on the path"
	done
}

WORK=$(mktemp -d) || die "cannot make a temporary directory"
trap 'rm -rf "$WORK"' EXIT HUP INT TERM

TAG="managed-by:proven-judges-gateway"
PLANNED="(planned)"

# ---- settings: key=value lines, parsed, never sourced ----

load_settings() {
	[ -f "$1" ] || die "the settings file $1 does not exist; copy infra/cloudflare/settings.example"
	S_account_id= S_zone= S_hostname= S_gateway_port=8089 S_name=proven-judges-gateway
	S_judges= S_session_duration=24h S_runner_token_duration=336h S_team_name=
	S_state_dir= S_runner_secret_file= S_gateway_env_file= S_tunnel_token_file= S_token_file=
	S_gateway_bin= S_gateway_settings_file= S_identities_expire=
	S_judge_per_minute=20 S_judge_per_day=500 S_runner_per_minute=20 S_runner_per_day=500
	while IFS= read -r line || [ -n "$line" ]; do
		line=$(printf '%s' "$line" | sed -e 's/^[ \t]*//' -e 's/[ \t]*$//')
		case $line in '' | '#'*) continue ;; esac
		key=${line%%=*}
		[ "$key" != "$line" ] || die "settings: a line without '=': $key"
		val=${line#*=}
		key=$(printf '%s' "$key" | tr -d ' \t')
		# strip spaces around the value, then one pair of quotes
		val=$(printf '%s' "$val" | sed -e 's/^[ \t]*//' -e 's/[ \t]*$//')
		case $val in
		\"*\") val=${val#\"} val=${val%\"} ;;
		\'*\') val=${val#\'} val=${val%\'} ;;
		esac
		case $key in
		account_id | zone | hostname | gateway_port | name | judges | session_duration | runner_token_duration | team_name | state_dir | runner_secret_file | gateway_env_file | tunnel_token_file | token_file | gateway_bin | gateway_settings_file | identities_expire | judge_per_minute | judge_per_day | runner_per_minute | runner_per_day)
			eval "S_$key=\$val"
			;;
		*) die "settings: unknown key $key" ;;
		esac
	done <"$1"

	case $S_account_id in '' | *[!0-9a-f]*) die "settings: account_id is 32 lower-case hex digits" ;; esac
	[ ${#S_account_id} -eq 32 ] || die "settings: account_id is 32 lower-case hex digits"
	case $S_zone in '' | *[!a-z0-9.-]* | .* | *.) die "settings: zone is a domain name, lower case" ;; esac
	case $S_hostname in '' | *[!a-z0-9.-]*) die "settings: hostname is a host name, lower case" ;; esac
	case $S_hostname in "$S_zone" | *."$S_zone") ;; *) die "settings: hostname must be the zone or inside it" ;; esac
	case $S_gateway_port in '' | *[!0-9]*) die "settings: gateway_port is a port number" ;; esac
	[ "$S_gateway_port" -ge 1 ] && [ "$S_gateway_port" -le 65535 ] || die "settings: gateway_port is a port number"
	case $S_name in '' | *[!a-z0-9-]*) die "settings: name is lower-case letters, digits and hyphens" ;; esac
	for dur in "$S_session_duration" "$S_runner_token_duration"; do
		case $dur in '' | *[!0-9h]* | h* | *h?*) die "settings: a duration is a number of hours such as 24h" ;; esac
	done
	[ -n "$S_judges" ] || die "settings: judges is empty; an Access policy with nobody on it admits nobody"
	for e in $(printf '%s' "$S_judges" | tr ',' ' '); do
		case $e in *[!A-Za-z0-9._%+@*-]*) die "settings: judges holds an entry that is not an e-mail or *@domain" ;; esac
		case $e in *@*@*) die "settings: judges holds an entry with two @" ;; ?*@?*.?*) ;; *) die "settings: judges holds an entry that is not an e-mail or *@domain" ;; esac
	done
	state_default=${XDG_STATE_HOME:-${HOME:-.}/.local/state}/proven-judges-gateway
	[ -n "$S_state_dir" ] || S_state_dir=$state_default
	[ -n "$S_runner_secret_file" ] || S_runner_secret_file=$S_state_dir/runner-service-token.env
	[ -n "$S_gateway_env_file" ] || S_gateway_env_file=$S_state_dir/gateway-access.env
	[ -n "$S_tunnel_token_file" ] || S_tunnel_token_file=$S_state_dir/tunnel.token
	STATE=$S_state_dir/state.env
	ACC=/accounts/$S_account_id
}

# ---- the token: from the environment or a file, never printed ----

load_token() {
	CF_TOKEN=${CLOUDFLARE_API_TOKEN:-}
	if [ -z "$CF_TOKEN" ] && [ -n "$S_token_file" ]; then
		[ -r "$S_token_file" ] || die "the token_file setting names a file that cannot be read"
		CF_TOKEN=$(sed -n '1{s/[[:space:]]//g;p;q}' "$S_token_file")
	fi
	[ -n "$CF_TOKEN" ] || die "no API token: set CLOUDFLARE_API_TOKEN, or token_file in the settings"
	case $CF_TOKEN in *[!A-Za-z0-9_.~-]*) die "the API token holds characters outside the usual set; the value is not printed" ;; esac
	API_BASE=${CLOUDFLARE_API_BASE:-https://api.cloudflare.com/client/v4}
	case $API_BASE in
	https://* | http://127.0.0.1:* | http://localhost:*) ;;
	*) die "CLOUDFLARE_API_BASE must be https, or http on this machine" ;;
	esac
}

# uri percent-encodes one value for a query string.
uri() { jq -rn --arg v "$1" '$v|@uri'; }

# cf METHOD PATH [BODYFILE]: one API call. The answer is in $RESP. Returns 0
# when Cloudflare says success. On failure it prints every error by its code and
# message (never the token) unless CF_QUIET is set.
CF_QUIET=
RESP=$WORK/resp.json
cf() {
	cf_method=$1 cf_path=$2 cf_body=${3:-}
	: >"$RESP"
	cf_rc=0
	{
		printf 'silent\nshow-error\nmax-time = 60\n'
		printf 'request = "%s"\nurl = "%s%s"\n' "$cf_method" "$API_BASE" "$cf_path"
		printf 'header = "Authorization: Bearer %s"\n' "$CF_TOKEN"
		printf 'header = "Content-Type: application/json"\n'
		printf 'output = "%s"\nwrite-out = "%%{http_code}"\n' "$RESP"
		if [ -n "$cf_body" ]; then printf 'data-binary = "@%s"\n' "$cf_body"; fi
	} | curl -K - >"$WORK/code" 2>"$WORK/curl.err" || cf_rc=$?
	if [ "$cf_rc" -ne 0 ]; then
		[ -n "$CF_QUIET" ] || printf 'error: %s %s: the request failed (curl exit %s)\n' "$cf_method" "$cf_path" "$cf_rc" >&2
		return 1
	fi
	if jq -e '.success == true' "$RESP" >/dev/null 2>&1; then return 0; fi
	if [ -z "$CF_QUIET" ]; then
		if jq -e '.errors|length>0' "$RESP" >/dev/null 2>&1; then
			printf 'error: %s %s: HTTP %s\n' "$cf_method" "$cf_path" "$(cat "$WORK/code")" >&2
			jq -r '.errors[] | "  Cloudflare error \(.code): \(.message)"' "$RESP" >&2
		else
			printf 'error: %s %s: HTTP %s with no Cloudflare answer\n' "$cf_method" "$cf_path" "$(cat "$WORK/code")" >&2
		fi
	fi
	return 1
}

# cfd is cf that ends the run when the call fails.
cfd() { cf "$@" || die "stopped at: $1 $2"; }

# body NAME: the JSON body is in $WORK/NAME.json, built by the caller from jq.
bodyfile() { printf '%s/%s.json' "$WORK" "$1"; }

# ---- what a run does: plan prints, apply acts ----

MODE=plan
# act "text": prints what happens; returns 0 when the change is to be made now.
act() {
	if [ "$MODE" = plan ]; then
		say "plan: would $1"
		return 1
	fi
	say "do:   $1"
	return 0
}

# ---- the state file: the ids of objects up.sh created, and only those ----

state_init() {
	mkdir -p "$S_state_dir" || die "cannot make the state directory"
	chmod 700 "$S_state_dir" 2>/dev/null || true
	[ -f "$STATE" ] || : >"$STATE"
}

record() {
	grep -qxF "$1=$2" "$STATE" 2>/dev/null || printf '%s=%s\n' "$1" "$2" >>"$STATE"
}

unrecord() {
	[ -f "$STATE" ] || return 0
	grep -vxF "$1=$2" "$STATE" >"$WORK/state.new" || true
	cat "$WORK/state.new" >"$STATE"
}

recorded() { [ -f "$STATE" ] && grep -qxF "$1=$2" "$STATE"; }

# write_secret FILE TEXT...: writes a file at mode 0600 through a temporary name.
write_private() {
	wp_file=$1
	shift
	wp_dir=$(dirname "$wp_file")
	mkdir -p "$wp_dir" || die "cannot make the directory of $wp_file"
	printf '%s\n' "$@" >"$wp_file.new" || die "cannot write $wp_file"
	chmod 600 "$wp_file.new"
	mv "$wp_file.new" "$wp_file"
}
