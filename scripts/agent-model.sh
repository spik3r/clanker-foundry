#!/usr/bin/env bash
# Print the concrete model for a portable agent profile and client.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
map_file="$repo_dir/agents/model-map.conf"

usage() {
  printf '%s\n' "Usage: scripts/agent-model.sh <claude|codex> <agent-name>" >&2
  exit 2
}

[[ $# -eq 2 ]] || usage
client="$1"
agent="$2"
profile="$repo_dir/agents/$agent.md"

case "$client" in
  claude|codex) ;;
  *) usage ;;
esac

[[ -f "$profile" ]] || {
  printf 'Unknown agent profile: %s\n' "$agent" >&2
  exit 2
}

tier="$(sed -n 's/^Suggested model tier: `\([^`]*\)`$/\1/p' "$profile")"
[[ -n "$tier" ]] || {
  printf 'No model tier in %s\n' "$profile" >&2
  exit 1
}

model="$(awk -F= -v key="$client.$tier" '$1 == key { print $2; exit }' "$map_file")"
[[ -n "$model" ]] || {
  printf 'No model mapping for %s.%s\n' "$client" "$tier" >&2
  exit 1
}

printf '%s\n' "$model"
