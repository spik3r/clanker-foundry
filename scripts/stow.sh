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
stow_package "$HOME/.agents/agents" agents
stow_package "$HOME/.claude/agents" agents

# Global instructions are a real, user-owned file, not a Stow symlink, so
# machine-local edits survive and only AGENTS.md lands in the config directory
# (not this package's README or wiring script). Seed it from the example on the
# first run; never overwrite an existing file. wire-global-agents.sh then points
# each tool at it.
canon="${XDG_CONFIG_HOME:-$HOME/.config}/agents/AGENTS.md"
if [ -e "$canon" ]; then
  printf 'Global instructions already present at %s; leaving as-is.\n' "$canon"
else
  mkdir -p "$(dirname "$canon")"
  cp "$repo_dir/global-agents/AGENTS.md" "$canon"
  printf 'Seeded global instructions at %s from the example.\n' "$canon"
fi

"$repo_dir/global-agents/wire-global-agents.sh"
printf '%s\n' 'Stowing completed.'
