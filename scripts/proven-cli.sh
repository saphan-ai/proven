#!/bin/sh
# One customer-side build/login/probe path for the supported CLI transports.
set -eu
operation=${1:-}
backend=${2:-}
case "$backend" in claude|copilot|codex|cursor|qwen) ;; *)
  echo 'Usage: sh scripts/proven-cli.sh build|login|check claude|copilot|codex|cursor|qwen [model]' >&2
  exit 2 ;;
esac
image=${PROVEN_CLI_IMAGE:-proven-$backend:local}
profile_volume=${PROVEN_CLI_HOME_VOLUME:-proven-$backend-home}
model=${3:-cli-default}
[ "$model" != cli-default ] || case "$backend" in claude) model=sonnet ;; copilot) model=auto ;; esac
root=$(cd "$(dirname "$0")/.." && pwd)
case "$operation" in
  build)
    if [ "$backend" = cursor ]; then
      docker build -f "$root/examples/cli/Dockerfile.cursor" -t "$image" "$root"
    else
      docker build -f "$root/examples/cli/Dockerfile" --target "$backend" -t "$image" "$root"
    fi
    echo "Built $image in your environment. Scan it before deployment; see docs/CLI_IMAGES.md."
    ;;
  login)
    docker image inspect "$image" >/dev/null
    docker volume create "$profile_volume" >/dev/null
    case "$backend" in
      claude) set -- env CLAUDE_CONFIG_DIR=/opt/proven/.config/proven/claude claude auth login ;;
      copilot) set -- sh -c 'export COPILOT_HOME="$HOME/.config/proven/copilot"; copilot login --device-code && copilot config disableAllHooks true' ;;
      codex) set -- env CODEX_HOME=/opt/proven/.config/proven/codex codex login --device-auth ;;
      cursor) set -- env CURSOR_CONFIG_DIR=/opt/proven/.config/proven/cursor AGENT_CLI_CREDENTIAL_STORE=file NO_OPEN_BROWSER=1 agent login ;;
      qwen) set -- qwen --auth-type qwen-oauth ;;
    esac
    docker run --rm -it --cap-drop ALL --security-opt no-new-privileges \
      --mount "type=volume,src=$profile_volume,dst=/opt/proven" --entrypoint /usr/bin/env "$image" "$@"
    ;;
  check)
    docker volume inspect "$profile_volume" >/dev/null
    # Exactly one short model request. No repository is mounted or reviewed.
    docker run --rm -i --cap-drop ALL --security-opt no-new-privileges \
      --mount "type=volume,src=$profile_volume,dst=/opt/proven" \
      -e PROVEN_GENERATOR_MODE=command -e "PROVEN_GENERATOR_MODEL=$model" \
      -e "PROVEN_GENERATOR_COMMAND=[\"proven-$backend\",\"{model}\"]" \
      -e PROVEN_GENERATOR_TIMEOUT=60s --entrypoint timeout "$image" \
      --kill-after=2s 65s proven model check
    ;;
  *) echo 'Choose build, login or check.' >&2; exit 2 ;;
esac
