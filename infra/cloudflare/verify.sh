#!/bin/sh
# verify.sh: look at the live host from outside, as two callers would.
#
#   sh verify.sh SETTINGS
#
# 1. a call with no credential must be stopped by Access (a redirect to the
#    sign-in page, or 401 or 403), never answered by the gateway;
# 2. a call with the runner's service token must pass Access and reach the
#    gateway, which refuses an empty evidence request by name (HTTP 400,
#    not_a_template_request). That one answer shows Access, the tunnel, the
#    gateway and the runner's identity row all work.
#
# The runner's secret is read from the file up.sh wrote and goes to curl on its
# standard input, never on a command line. It uses no Cloudflare API token.
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck source=lib.sh
. "$HERE/lib.sh"

[ $# -eq 1 ] || {
	printf 'usage: sh verify.sh SETTINGS\n' >&2
	exit 2
}
need curl jq
load_settings "$1"
BASE=${VERIFY_BASE:-https://$S_hostname}
case $BASE in
https://* | http://127.0.0.1:* | http://localhost:*) ;;
*) die "VERIFY_BASE must be https, or http on this machine" ;;
esac
FAILED=0

# call [CLIENT_ID CLIENT_SECRET]: POST an empty evidence request; the answer's
# status is in $WORK/code and its body in $WORK/body.
call() {
	{
		printf 'silent\nmax-time = 30\nrequest = "POST"\nurl = "%s/v1/evidence/query"\n' "$BASE"
		printf 'header = "Content-Type: application/json"\ndata = "{}"\n'
		if [ $# -eq 2 ]; then
			printf 'header = "CF-Access-Client-Id: %s"\nheader = "CF-Access-Client-Secret: %s"\n' "$1" "$2"
		fi
		printf 'output = "%s/body"\nwrite-out = "%%{http_code}"\n' "$WORK"
	} | curl -K - >"$WORK/code" 2>/dev/null || printf '000' >"$WORK/code"
}

call
code=$(cat "$WORK/code")
case $code in
301 | 302 | 303 | 307 | 308 | 401 | 403) say "ok:   an anonymous call is stopped by Access (HTTP $code)" ;;
*)
	say "FAIL: an anonymous call got HTTP $code; Access is not in front of the gateway"
	FAILED=1
	;;
esac

[ -r "$S_runner_secret_file" ] || die "the runner's secret file $S_runner_secret_file is not readable; run up.sh first"
cid=$(sed -n 's/^CF_ACCESS_CLIENT_ID=//p' "$S_runner_secret_file")
csec=$(sed -n 's/^CF_ACCESS_CLIENT_SECRET=//p' "$S_runner_secret_file")
[ -n "$cid" ] && [ -n "$csec" ] || die "the runner's secret file does not hold both values"
call "$cid" "$csec"
code=$(cat "$WORK/code")
if [ "$code" = 400 ] && jq -e '.refused == "not_a_template_request"' "$WORK/body" >/dev/null 2>&1; then
	say "ok:   the runner's service token passes Access and the gateway names the empty request"
else
	say "FAIL: the runner's call got HTTP $code (wanted 400 not_a_template_request)"
	[ "$code" != 403 ] || say "      403 usually means Access refused the token, or the gateway has no identity row for it"
	FAILED=1
fi
[ "$FAILED" -eq 0 ] || exit 1
say "done: verified"
