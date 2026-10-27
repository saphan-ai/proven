#!/bin/sh
# Name which live services are ready, for the setup the workflow will review with. Reads lab/.env when that file exists.
#
# The mode, the guardrail's mode and key, and whether Zetaris is on are not decided here. They are decided by the lines
# of lab/workflow/proven.yml between "derive-begin" and "derive-end", and this script runs those same lines on the same
# settings. So a key the workflow reuses (one NVIDIA key serves the guardrail too) is reused here, and a service the
# workflow runs without (Zetaris, when no address and key are set) is not asked for.
#
# LIVECHECK_BIN names the checker to run in place of ./bin/livecheck. It is a seam for the tests, not for use by hand.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

load_lab_env || exit 1

# With ZETARIS=local and no address of your own, the check runs from this host, so it asks the loopback port the
# platform is published on (lab/zetaris/override.yml), not the name the lab's reviews use on the lab network.
if [ "${ZETARIS:-}" = "local" ] && [ -z "${ZETARIS_REST_URL:-}" ]; then
  ZETARIS_REST_URL="http://127.0.0.1:${ZETARIS_API_HOST_PORT}${ZETARIS_REST_PATH}"
  export ZETARIS_REST_URL
fi

workflow=$ROOT/lab/workflow/proven.yml
derive=$(sed -n '/# derive-begin/,/# derive-end/p' "$workflow" | sed 's/^[[:space:]]*//')
if [ -z "$derive" ]; then
  die "livecheck: lab/workflow/proven.yml has no derive-begin ... derive-end block, so the setup the workflow will use cannot be read. NEXT — restore the block; make check-golive holds it."
fi

# The workflow reads every one of these as a stored secret that may be empty. Here an unset one is empty too.
for name in PROVEN_GENERATOR_MODE PROVEN_GENERATOR_KEY PROVEN_NVIDIA_BASE_URL PROVEN_GUARDRAIL_MODE PROVEN_GUARDRAIL_KEY \
  PROVEN_GENERATOR_BASE_URL METERLESS_RUNTIME_TOKEN ZETARIS_REST_URL ZETARIS_API_KEY ZETARIS_ORG_ID \
  JUDGES_GATEWAY_URL JUDGES_GATEWAY_TOKEN JUDGES_GATEWAY_ACCESS_TOKEN; do
  eval ": \${$name:=}"
done
GATEWAY_ZETARIS_PATH=$JUDGES_GATEWAY_ZETARIS_PATH
GATEWAY_ZETARIS_ORG_ID=$JUDGES_GATEWAY_ZETARIS_ORG_ID
eval "$derive"

echo "livecheck: the workflow will run: $state"
if [ "$sample" = true ]; then
  echo "livecheck: no key, token or gateway is set, so the workflow runs a SAMPLE review and no live service is asked. NEXT — set a key in lab/.env (docs/LAB.md, \"Go live\"), then run make livecheck again."
  exit 1
fi

if [ -n "${LIVECHECK_BIN:-}" ]; then
  checker=$LIVECHECK_BIN
else
  sh "$ROOT/lab/scripts/build.sh"
  checker=$ROOT/bin/livecheck
fi

rc=0
out=$("$checker") || rc=$?
printf '%s\n' "$out"

# Zetaris is optional in the workflow: with no address and key it runs without. Its query line then does not fail
# this check, and the line says why. Any other service that is not ready still does.
if [ "$rc" -eq 1 ] && [ "$zetaris" = off ]; then
  others=$(printf '%s\n' "$out" | grep ': not ready' | grep -vc '^query: ' || true)
  if [ "$others" -eq 0 ]; then
    echo "remark: the query line is not needed in this setup: Zetaris is off in the workflow (no ZETARIS_REST_URL, ZETARIS_API_KEY and ZETARIS_ORG_ID together), so the review runs without it."
    rc=0
  fi
fi
exit "$rc"
