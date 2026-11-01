#!/usr/bin/env bash
# Isolated regression tests for scripts/agent-model.sh.
#
# Every case runs a copy of the resolver inside a temporary fixture repository
# with invented mapping files. The real checkout, HOME and any local override are
# never read, and no model provider is called.
#
# Usage: tests/test-agent-model.sh [resolver-path]
# Pass another resolver path to run the same cases against it, for example a
# baseline copy from main.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
resolver="${1:-$repo_dir/scripts/agent-model.sh}"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

export HOME="$test_root/home"
export XDG_CONFIG_HOME="$test_root/config"
mkdir -p "$HOME" "$XDG_CONFIG_HOME"

fixture="$test_root/repo"
mkdir -p "$fixture/scripts" "$fixture/agents"
cp "$resolver" "$fixture/scripts/agent-model.sh"
for tier in fast balanced flagship; do
  # Write literal Markdown backticks, not shell command substitution.
  # shellcheck disable=SC2016
  printf 'Suggested model tier: `%s`\n' "$tier" > "$fixture/agents/fixture-$tier.md"
done
printf 'No tier here.\n' > "$fixture/agents/fixture-none.md"
shared="$fixture/agents/model-map.conf"
local_map="$fixture/agents/model-map.local.conf"

failures=0
fail() {
  printf 'FAIL  %s\n' "$1" >&2
  failures=$((failures + 1))
}

reset_maps() {
  rm -f "$shared" "$local_map"
}

# run <client> <agent>: sets status, out and err from one resolver run.
run() {
  status=0
  out="$(bash "$fixture/scripts/agent-model.sh" "$@" 2>"$test_root/err")" || status=$?
  err="$(cat "$test_root/err")"
}

# expect_model <name> <expected> <client> <agent>
expect_model() {
  local name="$1" expected="$2"
  shift 2
  run "$@"
  if [[ $status -ne 0 || "$out" != "$expected" || -n "$err" ]]; then
    fail "$name: expected '$expected' and exit 0, got exit $status, stdout '$out', stderr '$err'"
  fi
}

# expect_error <name> <exit> <stderr-substring> <client> <agent>
# A failure must print nothing on stdout and must not echo any invented value.
expect_error() {
  local name="$1" expected_status="$2" message="$3"
  shift 3
  run "$@"
  if [[ $status -ne $expected_status ]]; then
    fail "$name: expected exit $expected_status, got $status (stdout '$out')"
  fi
  if [[ -n "$out" ]]; then
    fail "$name: stdout must be empty on error, got '$out'"
  fi
  if [[ "$err" != *"$message"* ]]; then
    fail "$name: stderr lacks '$message': '$err'"
  fi
  if [[ "$err" == *invented-* ]]; then
    fail "$name: stderr exposes a mapping value: '$err'"
  fi
}

# Duplicate keys in the shared file, in both orders.
reset_maps
printf 'claude.fast=invented-a\nclaude.fast=invented-b\n' > "$shared"
expect_error 'shared duplicate a,b' 1 'agents/model-map.conf:2: duplicate key claude.fast' claude fixture-fast
printf 'claude.fast=invented-b\nclaude.fast=invented-a\n' > "$shared"
expect_error 'shared duplicate b,a' 1 'agents/model-map.conf:2: duplicate key claude.fast' claude fixture-fast

# Identical duplicates are still ambiguous configuration.
printf 'claude.fast=invented-a\nclaude.fast=invented-a\n' > "$shared"
expect_error 'shared identical duplicate' 1 'duplicate key claude.fast' claude fixture-fast

# A duplicate of a key other than the one requested still fails.
printf 'claude.fast=invented-a\ncodex.fast=invented-b\ncodex.fast=invented-c\n' > "$shared"
expect_error 'shared unrelated duplicate' 1 'duplicate key codex.fast' claude fixture-fast

# Duplicate keys in the local file, in both orders, and identical values.
reset_maps
printf 'claude.fast=invented-shared\n' > "$shared"
printf 'claude.fast=invented-a\nclaude.fast=invented-b\n' > "$local_map"
expect_error 'local duplicate a,b' 1 'agents/model-map.local.conf:2: duplicate key claude.fast' claude fixture-fast
printf 'claude.fast=invented-b\nclaude.fast=invented-a\n' > "$local_map"
expect_error 'local duplicate b,a' 1 'agents/model-map.local.conf:2: duplicate key claude.fast' claude fixture-fast
printf 'claude.fast=invented-a\nclaude.fast=invented-a\n' > "$local_map"
expect_error 'local identical duplicate' 1 'duplicate key claude.fast' claude fixture-fast
printf 'codex.fast=invented-a\n# note\ncodex.fast=invented-b\n' > "$local_map"
expect_error 'local duplicate of other key' 1 'agents/model-map.local.conf:3: duplicate key codex.fast' claude fixture-fast

# The same key once in each file is a legitimate override, not a duplicate.
reset_maps
printf 'claude.fast=invented-shared\nclaude.balanced=invented-shared-balanced\n' > "$shared"
printf 'claude.fast=invented-local\n' > "$local_map"
expect_model 'local overrides shared' invented-local claude fixture-fast

# A local file that lacks the key falls back to the shared value.
expect_model 'partial local falls back' invented-shared-balanced claude fixture-balanced

# A missing local file uses the shared file.
rm -f "$local_map"
expect_model 'no local file' invented-shared claude fixture-fast

# Empty, comment-only and malformed local files fail instead of falling back.
: > "$local_map"
expect_error 'empty local file' 1 'agents/model-map.local.conf: no model mappings' claude fixture-fast
printf '# only a comment\n\n' > "$local_map"
expect_error 'comment-only local file' 1 'agents/model-map.local.conf: no model mappings' claude fixture-fast
for line in 'claude.fast' 'claude.fast=' '=invented-a' 'claude.fast = invented-a' \
  'claude.fast=invented-a # note' 'Claude.fast=invented-a' 'claudefast=invented-a' \
  'claude.fast=invented-a=b' 'export claude.fast=invented-a'; do
  printf 'codex.fast=invented-ok\n%s\n' "$line" > "$local_map"
  expect_error "malformed local line '$line'" 1 'agents/model-map.local.conf:2: expected <client>.<tier>=<model>' claude fixture-fast
done
printf 'claude.fast=invented-a\r\n' > "$local_map"
expect_error 'CRLF local line' 1 'agents/model-map.local.conf:1: expected' claude fixture-fast
rm -f "$local_map"
mkdir "$local_map"
expect_error 'local path is a directory' 1 'Not a regular file: agents/model-map.local.conf' claude fixture-fast
rmdir "$local_map"
ln -s "$test_root/missing-target" "$local_map"
expect_error 'dangling local symlink' 1 'Not a regular file: agents/model-map.local.conf' claude fixture-fast
rm -f "$local_map"

# Empty and malformed shared files fail.
: > "$shared"
expect_error 'empty shared file' 1 'agents/model-map.conf: no model mappings' claude fixture-fast
printf 'claude.fast invented-a\n' > "$shared"
expect_error 'malformed shared line' 1 'agents/model-map.conf:1: expected' claude fixture-fast
rm -f "$shared"
expect_error 'missing shared file' 1 'Missing model map: agents/model-map.conf' claude fixture-fast

# Comments, indented comments and blank lines are ignored.
printf '# header\n\n   # indented\n\t\nclaude.fast=invented-a\n\n# trailer\n' > "$shared"
expect_model 'comments and blank lines' invented-a claude fixture-fast

# Mapping files are data. Shell syntax in a value is printed literally and a
# line of shell is rejected; neither runs.
marker="$test_root/executed"
# The fixture must hold unexpanded shell syntax.
# shellcheck disable=SC2016
printf 'claude.fast=$(touch${IFS}%s)\n' "$marker" > "$shared"
expect_model 'shell syntax value stays literal' "\$(touch\${IFS}$marker)" claude fixture-fast
printf 'claude.fast=invented-a\ntouch %s\n' "$marker" > "$shared"
expect_error 'shell line rejected' 1 'agents/model-map.conf:2: expected' claude fixture-fast
[[ ! -e "$marker" ]] || fail 'mapping content was executed'

# Unknown or missing selections.
printf 'claude.fast=invented-a\n' > "$shared"
expect_error 'missing tier mapping' 1 'No model mapping for codex.fast' codex fixture-fast
expect_error 'unknown client' 2 'Usage:' gemini fixture-fast
expect_error 'unknown agent' 2 'Unknown agent profile: no-such-agent' claude no-such-agent
expect_error 'profile without tier' 1 'No model tier in' claude fixture-none
run claude
[[ $status -eq 2 && -z "$out" ]] || fail "missing argument: expected usage exit 2, got $status"

# The tracked defaults still resolve to the documented models.
defaults="$test_root/defaults"
mkdir -p "$defaults/scripts" "$defaults/agents"
cp "$resolver" "$defaults/scripts/agent-model.sh"
cp "$repo_dir"/agents/*-*.md "$repo_dir/agents/model-map.conf" "$defaults/agents/"
while read -r client agent expected; do
  status=0
  out="$(bash "$defaults/scripts/agent-model.sh" "$client" "$agent" 2>&1)" || status=$?
  [[ $status -eq 0 && "$out" == "$expected" ]] ||
    fail "default $client $agent: expected '$expected', got exit $status '$out'"
done <<'EOF'
claude explorer-researcher haiku
claude architect-planner sonnet
claude reviewer-verifier sonnet
claude builder-developer opus
codex explorer-researcher gpt-5.6-terra
codex architect-planner gpt-5.6-terra
codex reviewer-verifier gpt-5.6-terra
codex builder-developer gpt-5.6-sol
EOF

if [[ $failures -gt 0 ]]; then
  printf '%d agent model resolver check(s) failed\n' "$failures" >&2
  exit 1
fi
printf '%s\n' 'PASS  agent model resolver tests'
