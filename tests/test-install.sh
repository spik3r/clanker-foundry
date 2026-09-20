#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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
rm "$untracked_doc"

printf '%s\n' 'PASS  installer wiring regression tests'
