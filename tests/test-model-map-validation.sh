#!/usr/bin/env bash
# Isolated map-format checks, with an optional unchanged baseline validator.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
validator="${1:-$repo_dir/scripts/validate.sh}"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT
export HOME="$test_root/home" XDG_CONFIG_HOME="$test_root/config" CODEX_HOME="$test_root/codex"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_COUNT=0
export GIT_CEILING_DIRECTORIES="$test_root"
unset GIT_CONFIG_PARAMETERS GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE
unset GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES
mkdir -p "$HOME" "$XDG_CONFIG_HOME" "$CODEX_HOME" "$test_root/bin"

fixture="$test_root/repo"
mkdir -p "$fixture/scripts" "$fixture/tests" "$fixture/global-agents" \
  "$fixture/agents" "$fixture/skills/fixture-skill"
cp "$validator" "$fixture/scripts/validate.sh"
printf '# Fixture\n' > "$fixture/AGENTS.md"
printf '# Fixture\n' > "$fixture/README.md"
printf '@AGENTS.md\n' > "$fixture/CLAUDE.md"
# shellcheck disable=SC2016
printf 'Read `AGENTS.md`.\n' > "$fixture/GEMINI.md"
printf '%s\n' '---' 'name: fixture-skill' 'description: Invented map validation fixture.' \
  '---' '# Fixture' > "$fixture/skills/fixture-skill/SKILL.md"
printf '# Fixture\n' > "$fixture/skills/fixture-skill/README.md"
# Backticks are literal Markdown in the fixture profile.
# shellcheck disable=SC2016
printf '%s\n' '---' 'name: fixture' 'description: Invented profile.' '---' \
  'Suggested model tier: `fast`' > "$fixture/agents/fixture.md"
printf '#!/usr/bin/env bash\n' > "$fixture/tests/syntax.sh"
printf '#!/usr/bin/env bash\n' > "$fixture/global-agents/syntax.sh"
printf 'claude.fast=PRIVATE_MODEL_SHARED\ncodex.fast=PRIVATE_MODEL_SHARED\n' > "$test_root/shared-base"
shared="$fixture/agents/model-map.conf"
local_map="$fixture/agents/model-map.local.conf"

# Simulate only regular map-reader failures. Other awk calls use the real tool.
# Partial output must be discarded when the reader exits nonzero.
REAL_AWK="$(command -v awk)"
export REAL_AWK
cat > "$test_root/bin/awk" <<'AWK'
#!/usr/bin/env bash
set -euo pipefail
for argument in "$@"; do
  if [ -n "$MAP_READER_MODE" ] && [ "$argument" = "$MAP_READER_FILE" ]; then
    printf 'called\n' >> "$MAP_READER_TRACE"
    if [ "$MAP_READER_MODE" = partial ]; then printf 'PRIVATE_MODEL_PARTIAL\n'; fi
    exit 73
  fi
done
exec "$REAL_AWK" "$@"
AWK
chmod +x "$test_root/bin/awk"
export PATH="$test_root/bin:$PATH"
export MAP_READER_TRACE="$test_root/reader-calls"
export MAP_READER_MODE='' MAP_READER_FILE=''

reset_maps() {
  cp "$test_root/shared-base" "$shared"
  rm -rf "$local_map"
  export MAP_READER_MODE='' MAP_READER_FILE=''
  : > "$MAP_READER_TRACE"
}

checks=0
failures=0
expect() {
  local label="$1" expected="$2" diagnostic="${3:-}" status=0 grep_status
  checks=$((checks + 1))
  (cd "$fixture" && bash scripts/validate.sh) > "$test_root/out" 2> "$test_root/err" || status=$?
  if [ "$status" -ne "$expected" ]; then
    printf 'FAIL  %s: expected exit %s, got %s\n' "$label" "$expected" "$status" >&2
    failures=$((failures + 1))
  elif [ "$expected" -eq 1 ]; then
    if ! grep -qF "$diagnostic" "$test_root/err"; then
      printf 'FAIL  %s: missing expected diagnostic\n' "$label" >&2
      failures=$((failures + 1))
    fi
  elif [ -s "$test_root/err" ] || ! grep -qxF 'PASS  no problems found' "$test_root/out"; then
    printf 'FAIL  %s: expected clean validator success\n' "$label" >&2
    failures=$((failures + 1))
  fi
  if grep -qF PRIVATE_MODEL_ "$test_root/out" "$test_root/err"; then
    printf 'FAIL  %s: diagnostic disclosed a model value\n' "$label" >&2
    failures=$((failures + 1))
  else
    grep_status=$?
    if [ "$grep_status" -ne 1 ]; then
      printf 'FAIL  %s: diagnostic redaction check failed\n' "$label" >&2
      failures=$((failures + 1))
    fi
  fi
}

reset_maps
expect 'shared defaults and missing local map' 0
printf 'claude.fast=PRIVATE_MODEL_LOCAL\n' > "$local_map"
expect 'local override repeats a shared key' 0
printf 'codex.flagship=PRIVATE_MODEL_UNUSED\n' > "$local_map"
expect 'local map can omit the selected key' 0

reset_maps
printf 'claude.fast=PRIVATE_MODEL_SHARED\ncodex.fast=PRIVATE_MODEL_SHARED' > "$shared"
printf 'claude.fast=PRIVATE_MODEL_LOCAL' > "$local_map"
expect 'valid maps without final newline' 0
printf '# Comment with café text\n\nclaude.fast=!\ncodex.fast=~\n' > "$shared"
printf '\t# ignored comment\n\nclaude.balanced=PRIVATE_MODEL_LOCAL\n' > "$local_map"
expect 'ASCII value boundaries and non-ASCII comments' 0

reset_maps
marker="$test_root/must-not-run"
# shellcheck disable=SC2016
printf 'claude.fast=$(touch${IFS}%s)\ncodex.fast=PRIVATE_MODEL_SHARED\n' "$marker" > "$shared"
expect 'shell-shaped model remains literal data' 0
test ! -e "$marker"

reset_maps
printf 'claude.fast=PRIVATE_MODEL_LINKED\n' > "$test_root/local-target"
ln -s "$test_root/local-target" "$local_map"
expect 'regular local symlink remains accepted' 0

for file in shared local; do
  for key in claude.fast codex.flagship; do
    for value in same different; do
      reset_maps
      target="$shared"
      label=agents/model-map.conf
      if [ "$file" = local ]; then target="$local_map"; label=agents/model-map.local.conf; fi
      if [ "$file" = shared ]; then
        printf 'claude.fast=PRIVATE_MODEL_FIRST\ncodex.fast=PRIVATE_MODEL_SHARED\n' > "$target"
        if [ "$key" = codex.flagship ]; then printf '%s=PRIVATE_MODEL_FIRST\n' "$key" >> "$target"; fi
      else
        printf '%s=PRIVATE_MODEL_FIRST\n' "$key" > "$target"
      fi
      if [ "$value" = same ]; then model=PRIVATE_MODEL_FIRST; else model=PRIVATE_MODEL_SECOND; fi
      printf '%s=%s\n' "$key" "$model" >> "$target"
      expect "$file duplicate $key $value value" 1 "$label:"
      if ! grep -qF "duplicate key $key (first defined on line " "$test_root/err"; then
        printf 'FAIL  %s duplicate lacks key and first-line diagnostic\n' "$file" >&2
        failures=$((failures + 1))
      fi
    done
  done
done

for file in shared local; do
  for kind in empty comments spaces tab nonascii equal blank-value missing-equal unsupported-key crlf; do
    reset_maps
    target="$shared"
    label=agents/model-map.conf
    if [ "$file" = local ]; then target="$local_map"; label=agents/model-map.local.conf; fi
    diagnostic='invalid or empty model mapping'
    case "$kind" in
      empty) : > "$target"; diagnostic='no model mappings' ;;
      comments) printf '# no mappings\n\n' > "$target"; diagnostic='no model mappings' ;;
      spaces) printf 'claude.fast=PRIVATE_MODEL two\n' > "$target" ;;
      tab) printf 'claude.fast=PRIVATE_MODEL\tvalue\n' > "$target" ;;
      nonascii) printf 'claude.fast=PRIVATE_MODEL_café\n' > "$target" ;;
      equal) printf 'claude.fast=PRIVATE_MODEL=value\n' > "$target" ;;
      blank-value) printf 'claude.fast=\n' > "$target" ;;
      missing-equal) printf 'claude.fast\n' > "$target" ;;
      unsupported-key) printf 'other.fast=PRIVATE_MODEL_VALUE\n' > "$target" ;;
      crlf) printf 'claude.fast=PRIVATE_MODEL_VALUE\r\n' > "$target" ;;
    esac
    expect "$file $kind" 1 "$label:"
    if ! grep -qF "$diagnostic" "$test_root/err"; then
      printf 'FAIL  %s %s lacks expected format diagnostic\n' "$file" "$kind" >&2
      failures=$((failures + 1))
    fi
  done
done

for file in shared local; do
  for position in record comment eof; do
    reset_maps
    target="$shared"
    label=agents/model-map.conf
    if [ "$file" = local ]; then target="$local_map"; label=agents/model-map.local.conf; fi
    case "$position" in
      record) printf 'claude.fast=PRIVATE_MODEL_\0VALUE\n' > "$target" ;;
      comment) printf '# comment\0hidden\nclaude.fast=PRIVATE_MODEL_VALUE\n' > "$target" ;;
      eof) printf 'claude.fast=PRIVATE_MODEL_VALUE\n\0' > "$target" ;;
    esac
    expect "$file NUL in $position" 1 "$label:"
    if ! grep -qF 'NUL byte is not allowed' "$test_root/err"; then
      printf 'FAIL  %s NUL in %s lacks NUL diagnostic\n' "$file" "$position" >&2
      failures=$((failures + 1))
    fi
  done
done

for kind in directory dangling; do
  reset_maps
  if [ "$kind" = directory ]; then mkdir "$local_map"; else ln -s "$test_root/missing" "$local_map"; fi
  expect "local $kind" 1 'agents/model-map.local.conf: missing or not a regular file'
done

for file in shared local; do
  for mode in silent partial; do
    reset_maps
    printf 'claude.fast=PRIVATE_MODEL_LOCAL\n' > "$local_map"
    export MAP_READER_MODE="$mode"
    if [ "$file" = shared ]; then
      export MAP_READER_FILE=agents/model-map.conf
    else
      export MAP_READER_FILE=agents/model-map.local.conf
    fi
    expect "$file reader $mode failure" 1 "$MAP_READER_FILE: could not read model mappings"
    grep -qxF called "$MAP_READER_TRACE"
  done
done

if [ "$failures" -ne 0 ]; then
  printf 'FAILED  %s assertions across %s model-map cases\n' "$failures" "$checks" >&2
  exit 1
fi
printf 'PASS  %s model-map validation cases\n' "$checks"
