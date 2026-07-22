# Global AGENTS.md — one base for every project

One instruction file that applies to every repo, so Claude Code, Codex, opencode,
Gemini CLI, and any other agent start from the same rails. Projects add their own
`AGENTS.md` on top; nothing global gets copied per-repo.

This directory holds:

- `AGENTS.md` — the example global base, ready to install
- `wire-global-agents.sh` — wires each tool to the canonical copy
- this README — why, where, and how to verify

## Why a global base

Every project needs the same baseline: correctness over cleverness, verify before
claiming success, no personal data in prompts, small reviewable diffs. Repeating that
in each repo's `AGENTS.md` means N copies that drift. Keep one canonical file at the
user level and let each project's file layer on top — one source of truth, the rest
are pointers.

What belongs in the global base: tool-agnostic behaviour and writing rules that are
true everywhere. What stays project-local: stack, versions, commands, and
repo-specific conventions. When the two conflict, the project wins (the file says so
up top).

## Where to keep it

There is no global location every tool agrees on. The `AGENTS.md` convention is
settled at the *project* level (an `AGENTS.md` at repo root, plus optional per-folder
files, closest wins), but the user-level path differs per tool:

| Tool | User-level file it actually reads |
|---|---|
| Codex | `~/.codex/AGENTS.md` |
| Claude Code | `~/.claude/CLAUDE.md` (note: *not* AGENTS.md) |
| opencode | `~/.config/opencode/AGENTS.md` |
| Gemini CLI | `~/.gemini/GEMINI.md` |
| GitLab Duo | `$XDG_CONFIG_HOME/gitlab/duo/AGENTS.md` |
| Cursor | `.cursor/rules/` (project-level only) |

So keep **one canonical file** and point every tool at it. Store it at the proposed
XDG path, aligned with where the convention is heading:

```text
$XDG_CONFIG_HOME/agents/AGENTS.md    # usually ~/.config/agents/AGENTS.md
```

Nothing reads that path natively yet, so wire each tool to it:

- **Codex** reads `~/.codex/AGENTS.md` as its global layer, then merges repo files on
  top. Symlink it to the canonical file.
- **Claude Code** reads `~/.claude/CLAUDE.md` as global user memory and supports
  `@path` imports (absolute paths, up to 4 hops). Add one import line pointing at the
  canonical file — use an import rather than a symlink so Claude-only memory can still
  live below it. Note `~` is not expanded in imports; the path must be absolute.
- **opencode** reads `~/.config/opencode/AGENTS.md` as its global layer. Symlink it.
- **Gemini CLI** reads `~/.gemini/GEMINI.md`. Symlink it to the canonical file.
- **Other tools** take the same treatment: a symlink where the tool reads a fixed
  file, or a one-line pointer where it does not.

Per project: a repo's own `AGENTS.md` is picked up automatically and layers on top
(Codex and opencode read it natively; for Claude Code a project `CLAUDE.md` containing
`@AGENTS.md` does the same; Gemini reads `GEMINI.md`, so point that at `AGENTS.md`
too). Keep project files short — stack, commands, local conventions — and let them
lean on the global base for everything generic. See
`templates/project/project-agents.md` for a starter.

## Setup, start to finish

1. Copy `AGENTS.md` from this directory to `~/.config/agents/AGENTS.md` — the
   canonical copy. Edit it to taste before wiring; every tool inherits it.
2. Run `./wire-global-agents.sh` to symlink Codex, opencode, and Gemini to it and add
   the Claude import.
3. In each project, keep a short `AGENTS.md` for stack, commands, and conventions; it
   overrides the global base where they differ.
4. Verify each tool picked it up — see below.

On a single-tool machine (e.g. a Claude-only work laptop), run just that tool's part
of the script; the rest is only needed where that tool is used.

## Verifying it loaded

**Codex** — `codex --print-instructions` dumps the fully merged text (global + repo +
cwd) to stdout, so you see exactly what the model got.

**Claude Code** — run `/memory` inside a session: it lists the loaded memory files,
their paths, and load order. Because the base comes in by `@` import, it shows up as
part of the expanded `~/.claude/CLAUDE.md` — if it is missing there, the import path
is wrong (use an absolute path; `~` is not expanded). `/context` shows how much of the
window it consumes. At the file level:

```bash
cat ~/.claude/CLAUDE.md            # should contain the @<absolute path> import line
cat ~/.config/agents/AGENTS.md     # the canonical file the import resolves to
```

**opencode / Gemini CLI** — the check is the symlink itself:

```bash
ls -l ~/.config/opencode/AGENTS.md ~/.gemini/GEMINI.md   # both point at the canonical file
```

Then start a session in any repo and confirm behaviour that only the global file would
produce (e.g. the writing rules).

## Updating

Edit the canonical file only. Symlinks pick changes up immediately; Claude's import
reads the file fresh each session. Never edit the per-tool copies — that is how drift
starts.
