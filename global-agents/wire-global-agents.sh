#!/usr/bin/env bash
# Wire each agent tool's user-level file to the canonical global AGENTS.md.
# Run once after saving the canonical file; safe to re-run.
set -euo pipefail

CANON="${XDG_CONFIG_HOME:-$HOME/.config}/agents/AGENTS.md"
mkdir -p "$(dirname "$CANON")"
[ -f "$CANON" ] || { echo "Create $CANON first (copy the example from this directory)." >&2; exit 1; }

# Codex — reads ~/.codex/AGENTS.md natively
mkdir -p "$HOME/.codex"
ln -sf "$CANON" "$HOME/.codex/AGENTS.md"

# Claude Code — import (not symlink) so Claude-only memory can still live below it
mkdir -p "$HOME/.claude"
touch "$HOME/.claude/CLAUDE.md"
grep -qxF "@$CANON" "$HOME/.claude/CLAUDE.md" || echo "@$CANON" >> "$HOME/.claude/CLAUDE.md"

# opencode — reads ~/.config/opencode/AGENTS.md as its global layer
mkdir -p "$HOME/.config/opencode"
ln -sf "$CANON" "$HOME/.config/opencode/AGENTS.md"

# Gemini CLI — reads ~/.gemini/GEMINI.md
mkdir -p "$HOME/.gemini"
ln -sf "$CANON" "$HOME/.gemini/GEMINI.md"

echo "Wired: Codex, opencode, Gemini symlinked; Claude import added -> $CANON"
