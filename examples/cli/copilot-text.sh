#!/bin/sh
# Customer-side adapter. Login uses a dedicated profile, not a PR posting token.
set -eu
umask 077
export COPILOT_HOME="$HOME/.config/proven/copilot"
mkdir -p "$COPILOT_HOME"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"
copilot --silent --output-format text --no-auto-update --no-ask-user \
  --no-custom-instructions --disable-builtin-mcps --disable-client-telemetry \
  --available-tools --model "$1" --deny-tool 'shell,read,write,url,memory'
