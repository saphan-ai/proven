#!/bin/sh
set -eu
umask 077
export CURSOR_CONFIG_DIR="$HOME/.config/proven/cursor"
export AGENT_CLI_CREDENTIAL_STORE=file
export CURSOR_AGENT_DISABLE_DEBUG_LOG=1
mkdir -p "$CURSOR_CONFIG_DIR"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir "$work/.cursor"
# This project is disposable and contains no target files. A persistent home
# is dedicated to login; do not mount a personal development profile here.
cat >"$work/.cursor/cli.json" <<'JSON'
{"permissions":{"allow":[],"deny":["Shell(*)","Read(**)","Read(/**)","Write(**)","Write(/**)"]}}
JSON
cd "$work"
if [ "$1" = cli-default ]; then set --; else set -- --model "$1"; fi
agent --print --mode ask --output-format text --disable-auto-update "$@"
