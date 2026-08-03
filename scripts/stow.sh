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

stow_shared_asset() {
  local package="$1"

  stow_package "$HOME/.agents/$package" "$package"
  stow_package "$HOME/.claude/$package" "$package"
}

# Install every reusable runtime asset for both clients. Repository documentation,
# setup scripts, and global instruction wiring remain repo-local.
for package in skills agents workflows templates checklists; do
  stow_shared_asset "$package"
done

# Stow the tracked global instructions as the canonical user-level file. Migrate
# the old copied file only when it still matches the repository; never hide local
# edits. The package ignore file keeps its README and wiring script repo-local.
global_target="${XDG_CONFIG_HOME:-$HOME/.config}/agents"
canon="$global_target/AGENTS.md"
if [ -e "$canon" ] && [ ! -L "$canon" ]; then
  if ! cmp -s "$repo_dir/global-agents/AGENTS.md" "$canon"; then
    printf 'Cannot stow global instructions: %s contains local changes.\n' "$canon" >&2
    printf '%s\n' 'Merge those changes into global-agents/AGENTS.md, then move the existing file and rerun.' >&2
    exit 1
  fi

  backup="$canon.pre-stow"
  if [ -e "$backup" ] || [ -L "$backup" ]; then
    printf 'Cannot migrate global instructions: backup already exists at %s.\n' "$backup" >&2
    exit 1
  fi
  mv "$canon" "$backup"
  printf 'Moved the matching legacy copy to %s.\n' "$backup"
fi

stow_package "$global_target" global-agents
"$repo_dir/global-agents/wire-global-agents.sh"
printf '%s\n' 'Stowing completed.'
