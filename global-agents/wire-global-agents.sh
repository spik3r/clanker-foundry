#!/usr/bin/env bash
# Wire each agent tool's user-level file to the canonical global AGENTS.md.
# Run once after saving the canonical file; safe to re-run.
set -euo pipefail

CANON="${XDG_CONFIG_HOME:-$HOME/.config}/agents/AGENTS.md"
codex_dir="${CODEX_HOME:-$HOME/.codex}"
claude_dir="$HOME/.claude"
opencode_dir="${XDG_CONFIG_HOME:-$HOME/.config}/opencode"
gemini_dir="$HOME/.gemini"

check_link_target() {
  local target="$1"
  local kind="$2"
  if [ -L "$target" ]; then
    [ "$(readlink "$target")" = "$CANON" ] || {
      printf 'Cannot wire %s: existing symlink points elsewhere.\n' "$target" >&2
      return 1
    }
  elif [ -e "$target" ]; then
    printf 'Cannot wire %s: existing %s would be overwritten.\n' "$target" "$kind" >&2
    return 1
  fi
}

check_conflicts() {
  check_link_target "$codex_dir/AGENTS.md" "file"
  check_link_target "$opencode_dir/AGENTS.md" "file"
  check_link_target "$gemini_dir/GEMINI.md" "file"

  local claude="$claude_dir/CLAUDE.md"
  if [ -L "$claude" ] || [ -d "$claude" ]; then
    printf 'Cannot wire %s: existing symlink or directory cannot receive an import.\n' "$claude" >&2
    return 1
  fi
}

if [ "${1:-}" = "--check" ]; then
  check_conflicts
  exit 0
fi

mkdir -p "$(dirname "$CANON")"
[ -f "$CANON" ] || { echo "Create $CANON first (copy the example from this directory)." >&2; exit 1; }
check_conflicts

# Codex — reads ~/.codex/AGENTS.md natively
mkdir -p "$codex_dir"
[ -L "$codex_dir/AGENTS.md" ] || ln -s "$CANON" "$codex_dir/AGENTS.md"

# Claude Code — import (not symlink) so Claude-only memory can still live below it
mkdir -p "$claude_dir"
touch "$claude_dir/CLAUDE.md"
if ! grep -qxF "@$CANON" "$claude_dir/CLAUDE.md"; then
  if [ -s "$claude_dir/CLAUDE.md" ] && [ "$(tail -c 1 "$claude_dir/CLAUDE.md")" != "" ]; then
    printf '\n' >> "$claude_dir/CLAUDE.md"
  fi
  printf '@%s\n' "$CANON" >> "$claude_dir/CLAUDE.md"
fi

# opencode — reads ~/.config/opencode/AGENTS.md as its global layer
mkdir -p "$opencode_dir"
[ -L "$opencode_dir/AGENTS.md" ] || ln -s "$CANON" "$opencode_dir/AGENTS.md"

# Gemini CLI — reads ~/.gemini/GEMINI.md
mkdir -p "$gemini_dir"
[ -L "$gemini_dir/GEMINI.md" ] || ln -s "$CANON" "$gemini_dir/GEMINI.md"

echo "Wired: Codex, opencode, Gemini symlinked; Claude import added -> $CANON"
