#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$repo_dir/tests/test-wire-preflight.sh"
bash "$repo_dir/tests/test-stow-preflight.sh"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

export HOME="$test_root/home"
export XDG_CONFIG_HOME="$test_root/config"
export CODEX_HOME="$test_root/codex"
mkdir -p "$XDG_CONFIG_HOME/agents" "$HOME/.claude"
cp "$repo_dir/global-agents/AGENTS.md" "$XDG_CONFIG_HOME/agents/AGENTS.md"
printf 'existing Claude memory' > "$HOME/.claude/CLAUDE.md"

"$repo_dir/global-agents/wire-global-agents.sh"
test "$(readlink "$CODEX_HOME/AGENTS.md")" = "$XDG_CONFIG_HOME/agents/AGENTS.md"
test "$(readlink "$XDG_CONFIG_HOME/opencode/AGENTS.md")" = "$XDG_CONFIG_HOME/agents/AGENTS.md"
test "$(readlink "$HOME/.gemini/GEMINI.md")" = "$XDG_CONFIG_HOME/agents/AGENTS.md"
grep -qxF "@${XDG_CONFIG_HOME}/agents/AGENTS.md" "$HOME/.claude/CLAUDE.md"
test "$(tail -c 1 "$HOME/.claude/CLAUDE.md")" = ""

before="$(cat "$HOME/.claude/CLAUDE.md")"
"$repo_dir/global-agents/wire-global-agents.sh"
test "$(cat "$HOME/.claude/CLAUDE.md")" = "$before"

rm "$CODEX_HOME/AGENTS.md"
ln -s "$test_root/missing" "$CODEX_HOME/AGENTS.md"
if "$repo_dir/global-agents/wire-global-agents.sh" --check 2>/dev/null; then
  printf '%s\n' 'dangling symlink was not rejected' >&2
  exit 1
fi

# A Claude symlink must never be treated as an appendable file.
rm -rf "$CODEX_HOME/AGENTS.md"
rm -f "$HOME/.claude/CLAUDE.md"
ln -s "$XDG_CONFIG_HOME/agents/AGENTS.md" "$HOME/.claude/CLAUDE.md"
if "$repo_dir/global-agents/wire-global-agents.sh" --check 2>/dev/null; then
  printf '%s\n' 'Claude symlink was not rejected' >&2
  exit 1
fi

# Stow must perform a clean install and remain safe to rerun.
rm -f "$HOME/.claude/CLAUDE.md"
rm -rf "$test_root/codex" "$HOME/.agents" "$HOME/.claude" "$HOME/.gemini" "$XDG_CONFIG_HOME/opencode"
mkdir -p "$HOME/.claude"
printf 'claude memory\n' > "$HOME/.claude/CLAUDE.md"
"$repo_dir/scripts/stow.sh" >/dev/null
canonical_before="$(cat "$XDG_CONFIG_HOME/agents/AGENTS.md")"
"$repo_dir/scripts/stow.sh" >/dev/null
test "$(cat "$XDG_CONFIG_HOME/agents/AGENTS.md")" = "$canonical_before"
test ! -e "$XDG_CONFIG_HOME/agents/README.md"
test ! -e "$HOME/.agents/agents/model-map.conf"

# A regular package target conflict stops before replacing the file.
conflict_root="$(mktemp -d "$test_root/conflict.XXXXXX")"
mkdir -p "$conflict_root/home/.agents/skills"
printf 'keep me\n' > "$conflict_root/home/.agents/skills/engineer"
if HOME="$conflict_root/home" XDG_CONFIG_HOME="$conflict_root/config" CODEX_HOME="$conflict_root/codex" \
  "$repo_dir/scripts/stow.sh" >/dev/null 2>&1; then
  printf '%s\n' 'package conflict was not rejected' >&2
  exit 1
fi
grep -qxF 'keep me' "$conflict_root/home/.agents/skills/engineer"

# Validator must inspect every shell file and include valid untracked Markdown.
# Run this check in a copied repository so fixtures never touch the worktree.
fixture_repo="$test_root/validator-repo"
mkdir -p "$fixture_repo"
(cd "$repo_dir" && tar --exclude=.git -cf - .) | (cd "$fixture_repo" && tar -xf -)
printf 'codex.fast=local-model\n' > "$fixture_repo/agents/model-map.local.conf"
fixture_home="$test_root/fixture-home"
fixture_config="$test_root/fixture-config"
mkdir -p "$fixture_home" "$fixture_config"
HOME="$fixture_home" XDG_CONFIG_HOME="$fixture_config" CODEX_HOME="$test_root/fixture-codex" \
  "$fixture_repo/scripts/stow.sh" >/dev/null
test ! -e "$fixture_home/.agents/agents/model-map.local.conf"

(cd "$fixture_repo" && git init -q && git config user.email test@example.invalid && \
  git config user.name 'Installer Test' && git add -A && git commit -qm fixture)
invalid_shell="$fixture_repo/tests/validator-invalid.sh"
untracked_doc="$fixture_repo/.validator-untracked.md"
cleanup_validation() {
  rm -f "$invalid_shell" "$untracked_doc"
}
trap 'cleanup_validation; rm -rf "$test_root"' EXIT
printf '%s\n' 'if (' > "$invalid_shell"
printf '%s\n' '[valid](README.md)' > "$untracked_doc"
if (cd "$fixture_repo" && bash scripts/validate.sh) >/dev/null 2>&1; then
  printf '%s\n' 'validator accepted invalid later shell file' >&2
  exit 1
fi
rm "$invalid_shell"
printf '%s\n' '[broken](missing-file.md)' > "$untracked_doc"
if (cd "$fixture_repo" && bash scripts/validate.sh) >/dev/null 2>&1; then
  printf '%s\n' 'validator accepted broken untracked Markdown link' >&2
  exit 1
fi
printf '%s\n' '[valid](README.md)' > "$untracked_doc"
(cd "$fixture_repo" && bash scripts/validate.sh) >/dev/null
# URL path escapes name filesystem bytes, not a literal percent spelling.
link_failures=0
expect_link() {
  local expected="$1" target="$2" status=0
  printf '[guide](%s)\n' "$target" > "$untracked_doc"
  (cd "$fixture_repo" && bash scripts/validate.sh) > "$test_root/link-check.log" 2>&1 || status=$?
  if [ "$status" -ne "$expected" ]; then
    printf 'link %s: expected %s, got %s\n' "$target" "$expected" "$status" >&2
    link_failures=$((link_failures + 1))
  fi
}
printf '# Target\n' > "$fixture_repo/file name.md"
printf '# Target\n' > "$fixture_repo/hash#name.md"
printf '# Target\n' > "$fixture_repo/percent%name.md"
printf '# Target\n' > "$fixture_repo/double%20name.md"
printf '# Target\n' > "$fixture_repo/café.md"
printf '# Target\n' > "$fixture_repo/plus+name.md"
printf '# Target\n' > "$fixture_repo/back\name.md"
printf '# Target\n' > "$fixture_repo/"$'ending\n'
printf '# Target\n' > "$fixture_repo/bad%2G.md"
printf '# Target\n' > "$fixture_repo/bad%.md"
printf '# Decoy\n' > "$fixture_repo/missing%20name.md"
expect_link 0 'file%20name.md'
expect_link 1 'missing%20name.md'
expect_link 1 'absent%20name.md'
expect_link 0 'hash%23name.md#fragment'
expect_link 0 'percent%25name.md'
expect_link 0 'double%2520name.md'
expect_link 0 'caf%C3%A9.md'
expect_link 0 'caf%c3%a9.md'
expect_link 0 'plus+name.md'
expect_link 0 'back\name.md'
expect_link 0 'ending%0A'
expect_link 1 'README.md%00suffix'
expect_link 1 'README.md%00'
expect_link 0 'bad%2G.md'
expect_link 0 'bad%.md'
expect_link 0 'README.md#ordinary-fragment'
expect_link 0 'https://example.invalid/absent%20name.md'
expect_link 0 '#anchor'
# Raw queries are URL metadata; encoded question marks remain filename bytes.
printf '# Target\n' > "$fixture_repo/question?name.md"
printf '# Decoy\n' > "$fixture_repo/query-missing.md?plain=1"
expect_link 0 'README.md?plain=1'
expect_link 1 'query-missing.md?plain=1'
expect_link 1 'query-absent.md?plain=1'
expect_link 0 'README.md?plain=1#fragment'
expect_link 0 'question%3Fname.md'
expect_link 0 'question%3Fname.md?plain=1'
expect_link 0 '?plain=1'
expect_link 0 'README.md?ignored=%00%2G%'
expect_link 0 '?ignored=%00%2G%'
[ "$link_failures" -eq 0 ] || exit 1
rm "$untracked_doc"

bash "$repo_dir/tests/test-agent-model.sh"

printf '%s\n' 'PASS  installer wiring regression tests'
