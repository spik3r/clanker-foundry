#!/usr/bin/env bash
# Isolated filename regressions for the pack validator, with no installer calls.
# Usage: bash tests/test-validate.sh [validator-path]
# The optional path lets these same expectations check an unchanged baseline.
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
mkdir -p "$HOME" "$XDG_CONFIG_HOME" "$CODEX_HOME" "$test_root/empty-template"

fixture="$test_root/repo"
mkdir -p "$fixture/scripts" "$fixture/tests" "$fixture/global-agents" \
  "$fixture/agents" "$fixture/skills/fixture-skill" "$fixture/docs"
cp "$validator" "$fixture/scripts/validate.sh"
printf '# Fixture\n' > "$fixture/AGENTS.md"
printf '# Fixture\n' > "$fixture/README.md"
printf '@AGENTS.md\n' > "$fixture/CLAUDE.md"
# shellcheck disable=SC2016
printf 'Read `AGENTS.md`.\n' > "$fixture/GEMINI.md"
printf 'claude.fast=fixture-model\ncodex.fast=fixture-model\n' > "$fixture/agents/model-map.conf"
printf '%s\n' '---' 'name: fixture-skill' 'description: Use for the invented validation fixture.' \
  '---' '# Fixture skill' > "$fixture/skills/fixture-skill/SKILL.md"
printf '# Fixture skill\n' > "$fixture/skills/fixture-skill/README.md"
printf '#!/usr/bin/env bash\n' > "$fixture/tests/syntax.sh"
printf '#!/usr/bin/env bash\n' > "$fixture/global-agents/syntax.sh"
printf 'ignored/\n' > "$fixture/.gitignore"
git -C "$fixture" init -q --template="$test_root/empty-template"
git -C "$fixture" add -- scripts tests global-agents agents skills \
  AGENTS.md README.md CLAUDE.md GEMINI.md .gitignore

checks=0
failures=0
expect() {
  local label="$1" expected="$2" display="${3:-}" status=0
  checks=$((checks + 1))
  bash "$fixture/scripts/validate.sh" > "$test_root/out" 2> "$test_root/err" || status=$?
  : > "$test_root/expected-err"
  if [ "$expected" -eq 1 ]; then
    printf 'FAIL  %s: broken link -> fixture-missing.txt\nFAILED  1 problem(s)\n' \
      "$display" > "$test_root/expected-err"
  fi
  if [ "$status" -ne "$expected" ] || ! cmp -s "$test_root/expected-err" "$test_root/err"; then
    printf 'FAIL  %s: expected exit %s and the exact link diagnostic; got exit %s\n' \
      "$label" "$expected" "$status" >&2
    cat "$test_root/err" >&2
    failures=$((failures + 1))
  elif [ "$expected" -eq 0 ] && ! grep -qxF 'PASS  no problems found' "$test_root/out"; then
    printf 'FAIL  %s: missing validator success summary\n' "$label" >&2
    failures=$((failures + 1))
  fi
}

# Each file is checked with a real sibling target, then with a missing one. The
# targets use existing Markdown syntax; only source filenames vary. Exact stderr
# comparisons also require literal filename diagnostics, including newlines.
check_document() {
  local label="$1" name="$2" tracked="$3" broken_status="$4" prefix="$5"
  local file="$fixture/$name" parent
  parent="${file%/*}"
  mkdir -p "$parent"
  printf 'Existing sibling target.\n' > "$parent/fixture-target.txt"
  printf '[Guide](fixture-target.txt)\n' > "$file"
  if [ "$tracked" = true ]; then git -C "$fixture" add -f -- "$name"; fi
  expect "$label valid" 0
  printf '[Guide](fixture-missing.txt)\n' > "$file"
  expect "$label broken" "$broken_status" "$prefix$name"
  if [ "$tracked" = true ]; then git -C "$fixture" rm -q --cached -f -- "$name"; fi
  rm -f "$file" "$parent/fixture-target.txt"
}

names=(
  'docs/plain.md'
  'docs/café.md'
  'docs/quote"name.md'
  'docs/back\slash.md'
  $'docs/tab\tname.md'
  $'docs/line\nname.md'
  $'docs/parent\n/guide.md'
  '-root.md'
)
expect 'clean Git fixture' 0
for quoting in true false; do
  git -C "$fixture" config --local core.quotePath "$quoting"
  for tracked in true false; do
    number=0
    for name in "${names[@]}"; do
      number=$((number + 1))
      check_document "Git quotePath=$quoting tracked=$tracked filename=$number" "$name" "$tracked" 1 ''
    done
    # An ignore pattern excludes only untracked files, not tracked documents.
    broken_status=0
    if [ "$tracked" = true ]; then broken_status=1; fi
    check_document "Git ignored quotePath=$quoting tracked=$tracked" \
      'ignored/café.md' "$tracked" "$broken_status" ''
  done
  [ "$(git -C "$fixture" config --local --get core.quotePath)" = "$quoting" ]
done

# Without a Git repository, retain the filesystem fallback and its .git exclusion.
# The ceiling prevents fixture discovery from falling through to a parent checkout.
mv "$fixture/.git" "$test_root/saved-git"
mkdir "$fixture/.git"
printf '[Hidden](fixture-missing.txt)\n' > "$fixture/.git/hidden.md"
if git -C "$fixture" rev-parse --show-toplevel >/dev/null 2>&1; then
  printf 'FAIL  fallback fixture unexpectedly belongs to a Git repository\n' >&2
  exit 1
fi
expect 'non-Git fixture excludes .git' 0
number=0
for name in "${names[@]}"; do
  number=$((number + 1))
  check_document "non-Git filename=$number" "$name" false 1 './'
done
# The filesystem fallback deliberately does not interpret Git ignore rules.
check_document 'non-Git ignored directory' 'ignored/café.md' false 1 './'

if [ "$failures" -ne 0 ]; then
  printf 'FAILED  %s of %s pack validator path checks\n' "$failures" "$checks" >&2
  exit 1
fi
printf 'PASS  %s pack validator path checks\n' "$checks"
