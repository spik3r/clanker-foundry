#!/usr/bin/env bash
# Print the concrete model for a portable agent profile and client.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
map_file="$repo_dir/agents/model-map.conf"
local_map_file="$repo_dir/agents/model-map.local.conf"

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

tier="$(awk '
  /^Suggested model tier: / {
    tier = $0
    sub(/^Suggested model tier: /, "", tier)
    tick = sprintf("%c", 96)
    if (tier ~ "^" tick "[^" tick "]+" tick "$") {
      gsub(tick, "", tier)
      print tier
      exit
    }
  }
' "$profile")"
[[ -n "$tier" ]] || {
  printf 'No model tier in %s\n' "$profile" >&2
  exit 1
}

# Read a mapping file as data and print the value for key, if present. The whole
# file is checked first, so a duplicate or malformed entry fails even when it is
# not the requested key. Messages name the file and line but never echo values.
read_map() {
  local prefix nul_line
  # BSD awk treats NUL as a string terminator, so reject it before parsing.
  # Bash 3.2 read with an empty delimiter detects the byte without storing it.
  if LC_ALL=C IFS= read -r -d '' prefix < "$2"; then
    nul_line="$(printf '%s' "$prefix" | LC_ALL=C wc -l)"
    printf '%s:%d: NUL byte is not allowed\n' "$3" "$((nul_line + 1))" >&2
    return 1
  fi
  LC_ALL=C awk -v key="$1" -v label="$3" '
    function bad(message) {
      printf "%s:%d: %s\n", label, NR, message > "/dev/stderr"
      failed = 1
      exit 1
    }
    /^[[:space:]]*(#|$)/ { next }
    {
      eq = index($0, "=")
      name = substr($0, 1, eq - 1)
      value = substr($0, eq + 1)
      if (eq == 0 || name !~ /^[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$/ ||
          value !~ /^[[:graph:]]+$/ || index(value, "=")) {
        bad("expected <client>.<tier>=<model>")
      }
      if (name in first_line) {
        bad("duplicate key " name " (first defined on line " first_line[name] ")")
      }
      first_line[name] = NR
      entries++
      if (name == key) found = value
    }
    END {
      if (failed) exit 1
      if (entries == 0) {
        printf "%s: no model mappings\n", label > "/dev/stderr"
        exit 1
      }
      if (found != "") print found
    }
  ' "$2"
}

key="$client.$tier"
[[ -f "$map_file" ]] || {
  printf 'Missing model map: agents/model-map.conf\n' >&2
  exit 1
}
shared_model="$(read_map "$key" "$map_file" agents/model-map.conf)" || exit 1

# A local file that exists must be valid; only a missing key falls back.
model=""
if [[ -e "$local_map_file" || -L "$local_map_file" ]]; then
  [[ -f "$local_map_file" ]] || {
    printf 'Not a regular file: agents/model-map.local.conf\n' >&2
    exit 1
  }
  model="$(read_map "$key" "$local_map_file" agents/model-map.local.conf)" || exit 1
fi
[[ -n "$model" ]] || model="$shared_model"
[[ -n "$model" ]] || {
  printf 'No model mapping for %s.%s\n' "$client" "$tier" >&2
  exit 1
}

printf '%s\n' "$model"
