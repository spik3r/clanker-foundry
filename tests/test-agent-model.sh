#!/usr/bin/env bash
# Isolated regression tests for scripts/agent-model.sh.
#
# Cases copy the resolver and tracked profiles into temporary fixture repositories.
# Mapping error cases use invented files. Real HOME and local overrides are never
# read, and no model provider is called.
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
export CODEX_HOME="$test_root/codex"
export TMPDIR="$test_root/tmp"
mkdir -p "$HOME" "$XDG_CONFIG_HOME" "$CODEX_HOME" "$TMPDIR"

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
  bash "$fixture/scripts/agent-model.sh" "$@" >"$test_root/out" 2>"$test_root/err" || status=$?
  out="$(cat "$test_root/out")"
  err="$(cat "$test_root/err")"
}

# expect_model <name> <expected> <client> <agent>
expect_model() {
  local name="$1" expected="$2"
  shift 2
  run "$@"
  printf '%s\n' "$expected" > "$test_root/expected"
  if [[ $status -ne 0 || "$out" != "$expected" || -s "$test_root/err" ]] ||
      ! cmp -s "$test_root/expected" "$test_root/out"; then
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
  if [[ -s "$test_root/out" ]]; then
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
# A valid local override does not hide a broken shared file.
printf 'claude.fast=invented-a\nclaude.fast=invented-b\n' > "$shared"
printf 'claude.fast=invented-local\n' > "$local_map"
expect_error 'shared duplicate with local override' 1 'agents/model-map.conf:2: duplicate key' claude fixture-fast
rm -f "$shared" "$local_map"
expect_error 'missing shared file' 1 'Missing model map: agents/model-map.conf' claude fixture-fast

# Reject NUL bytes before an awk implementation can truncate or hide a record.
# The fixture bytes are written directly, never held in shell variables. Check
# comments and EOF too, so a text parser cannot silently skip malformed bytes.
for location in shared local; do
  reset_maps
  printf 'claude.fast=invented-shared\n' > "$shared"
  target="$shared"
  label=agents/model-map.conf
  if [[ "$location" == local ]]; then
    target="$local_map"
    label=agents/model-map.local.conf
  fi
  while IFS='|' read -r name content line; do
    printf '%b' "$content" > "$target"
    expect_error "$location NUL $name" 1 "$label:$line: NUL byte is not allowed" claude fixture-fast
  done <<'NUL_CASES'
before value|claude.fast=\000invented-a\n|1
inside value|claude.fast=invented-\000a\n|1
after value|claude.fast=invented-a\000\n|1
before key|\000claude.fast=invented-a\n|1
inside key|claude.fa\000st=invented-a\n|1
before duplicate|claude.fast=invented-a\n\000claude.fast=invented-b\n|2
before malformed record|claude.fast=invented-a\n\000invalid record\n|2
blank-looking record|claude.fast=invented-a\n\000\n|2
inside comment|# note\000hidden\nclaude.fast=invented-a\n|1
trailing comment|claude.fast=invented-a\n# note\000hidden\n|2
EOF record|claude.fast=invented-a\n\000|2
EOF value|claude.fast=invented-a\000|1
after blank lines|\n\nclaude.fast=invented-a\n\000invalid\n|4
multiple bytes|# note\000hidden\nclaude.fast=invented-a\000\n|1
NUL_CASES
done

# A valid local override must not hide malformed bytes in the shared file.
printf 'claude.fast=invented-shared\n\000invalid\n' > "$shared"
printf 'claude.fast=invented-local\n' > "$local_map"
expect_error 'shared NUL with local override' 1 'agents/model-map.conf:2: NUL byte is not allowed' claude fixture-fast
rm -f "$local_map"

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

# Lookup keys are literal data too. AWK -v must not reinterpret an unsupported
# tier's octal escapes as the spelling of a different, configured tier.
for location in shared local; do
  reset_maps
  printf 'claude.fast=invented-shared-fast\nclaude.balanced=invented-shared-balanced\n' > "$shared"
  if [[ "$location" == local ]]; then
    printf 'claude.fast=invented-local-fast\nclaude.balanced=invented-local-balanced\n' > "$local_map"
  fi
  for tier in '\146ast' 'ba\154anced' missing; do
    # shellcheck disable=SC2016
    printf 'Suggested model tier: `%s`\n' "$tier" > "$fixture/agents/fixture-escaped.md"
    expect_error "$location literal tier $tier" 1 "No model mapping for claude.$tier" claude fixture-escaped
  done
done

# Model values retain their literal backslashes in both files.
reset_maps
printf '%s\n' 'claude.fast=invented-\146model' > "$shared"
expect_model 'shared backslash-bearing value' 'invented-\146model' claude fixture-fast
printf '%s\n' 'claude.fast=invented-local-\154model' > "$local_map"
expect_model 'local backslash-bearing override' 'invented-local-\154model' claude fixture-fast
rm -f "$local_map"

# Unknown or missing selections.
printf 'claude.fast=invented-a\n' > "$shared"
expect_error 'missing tier mapping' 1 'No model mapping for codex.fast' codex fixture-fast
expect_error 'unknown client' 2 'Usage:' gemini fixture-fast
expect_error 'unknown agent' 2 'Unknown agent profile: no-such-agent' claude no-such-agent
expect_error 'profile without tier' 1 'No model tier in' claude fixture-none
run claude
[[ $status -eq 2 && -z "$out" ]] || fail "missing argument: expected usage exit 2, got $status"

# Profile selection takes a single nonempty basename, never a path. These files
# are invented fixtures; the parent-path target stays inside the disposable repo.
mkdir -p "$fixture/agents/nested"
# shellcheck disable=SC2016
printf 'Suggested model tier: `fast`\n' > "$fixture/outside.md"
# shellcheck disable=SC2016
printf 'Suggested model tier: `fast`\n' > "$fixture/agents/nested/profile.md"
# shellcheck disable=SC2016
printf 'Suggested model tier: `fast`\n' > "$fixture/agents/.md"
for selector in ../outside ./fixture-fast nested/profile ..//outside '' /absolute; do
  expect_error "path-shaped or empty profile '$selector'" 2 'Usage:' claude "$selector"
done

# Do not introduce a new naming grammar for existing flat profile basenames.
for name in fixture-fast fixture_name fixture.name 'fixture space'; do
  # shellcheck disable=SC2016
  printf 'Suggested model tier: `fast`\n' > "$fixture/agents/$name.md"
  expect_model "flat profile basename '$name'" invented-a claude "$name"
done

# Only a terminal CR from a CRLF line ending is ignored; tier grammar is unchanged.
while IFS='|' read -r name content; do
  for ending in '\n' '\r\n'; do
    printf '%b%b' "$content" "$ending" > "$fixture/agents/fixture-invalid.md"
    expect_error "profile $name ending $ending" 1 'No model tier in' claude fixture-invalid
  done
done <<'PROFILE_CASES'
unquoted|Suggested model tier: fast
single quoted|Suggested model tier: 'fast'
double quoted|Suggested model tier: "fast"
unclosed|Suggested model tier: `fast
extra backtick|Suggested model tier: ``fast`
suffix|Suggested model tier: `fast` suffix
trailing space|Suggested model tier: `fast`\040
leading space| Suggested model tier: `fast`
extra prefix space|Suggested model tier:  `fast`
PROFILE_CASES
# shellcheck disable=SC2016
printf 'Suggested model tier: `%s`\r\r\n' fast > "$fixture/agents/fixture-invalid.md"
expect_error 'profile two terminal CR bytes' 1 'No model tier in' claude fixture-invalid
# shellcheck disable=SC2016
printf 'Suggested model tier: `fa\rst`\n' > "$fixture/agents/fixture-invalid.md"
expect_error 'profile internal CR is not stripped' 1 $'No model mapping for claude.fa\rst' claude fixture-invalid

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

# Exercise both line endings using the actual shipped declarations. Normalize only
# the copied line ending, so this also works when the test checkout itself is CRLF.
representations="$test_root/representations"
for ending in lf crlf; do
  mkdir -p "$representations/$ending/scripts" "$representations/$ending/agents"
  cp "$resolver" "$representations/$ending/scripts/agent-model.sh"
  cp "$repo_dir/agents/model-map.conf" "$representations/$ending/agents/"
done
for agent in explorer-researcher architect-planner reviewer-verifier builder-developer; do
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"
    printf '%s\n' "$line" >> "$representations/lf/agents/$agent.md"
    printf '%s\r\n' "$line" >> "$representations/crlf/agents/$agent.md"
  done < "$repo_dir/agents/$agent.md"
done
for ending in lf crlf; do
  fixture="$representations/$ending"
  while read -r client agent expected; do
    expect_model "$ending default $client $agent" "$expected" "$client" "$agent"
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
done

# CRLF profile support must not alter mapping overrides, fallback or CRLF refusal.
fixture="$representations/crlf"
shared="$fixture/agents/model-map.conf"
local_map="$fixture/agents/model-map.local.conf"
printf 'claude.fast=invented-shared\nclaude.balanced=invented-shared-balanced\n' > "$shared"
printf 'claude.fast=invented-local\n' > "$local_map"
expect_model 'CRLF profile local overrides shared' invented-local claude explorer-researcher
expect_model 'CRLF profile partial local falls back' invented-shared-balanced claude architect-planner
rm -f "$local_map"
expect_model 'CRLF profile no local file' invented-shared claude explorer-researcher
printf 'claude.fast=invented-local\r\n' > "$local_map"
expect_error 'CRLF local mapping with CRLF profile' 1 'agents/model-map.local.conf:1: expected' claude explorer-researcher
rm -f "$local_map"
printf 'claude.fast=invented-shared\r\n' > "$shared"
expect_error 'CRLF shared mapping with CRLF profile' 1 'agents/model-map.conf:1: expected' claude explorer-researcher

if [[ $failures -gt 0 ]]; then
  printf '%d agent model resolver check(s) failed\n' "$failures" >&2
  exit 1
fi
printf '%s\n' 'PASS  agent model resolver tests'
