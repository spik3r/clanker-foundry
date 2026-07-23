#!/usr/bin/env bash
# Symlink this pack into supported agent locations with GNU Stow.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v stow >/dev/null 2>&1; then
  printf '%s\n' 'GNU Stow is required. Install it, then run scripts/stow.sh.' >&2
  exit 1
fi

stow_package() {
  local target="$1"
  local package="$2"

  mkdir -p "$target"
  printf 'Stowing %s into %s...\n' "$package" "$target"
  stow --no-folding --dir "$repo_dir" --target "$target" --restow "$package"
}

stow_package "$HOME/.agents/skills" skills
stow_package "$HOME/.claude/skills" skills
stow_package "${XDG_CONFIG_HOME:-$HOME/.config}/agents" global-agents

"$repo_dir/global-agents/wire-global-agents.sh"
printf '%s\n' 'Stowing completed.'
