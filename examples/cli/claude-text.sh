#!/bin/sh
# Customer-side adapter. Install Claude Code yourself; this file ships no CLI.
set -eu
umask 077
export CLAUDE_CONFIG_DIR="$HOME/.config/proven/claude"
export DISABLE_AUTOUPDATER=1
mkdir -p "$CLAUDE_CONFIG_DIR"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"
claude --print --model "$1" --output-format text --no-session-persistence \
  --safe-mode --tools '' --strict-mcp-config --mcp-config '{"mcpServers":{}}'
