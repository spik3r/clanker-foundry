# Global AGENTS.md — one base for every project

One instruction file that applies to every repo, so Claude Code, Codex, opencode,
Gemini CLI, and any other agent start from the same rails. Projects add their own
`AGENTS.md` on top; nothing global gets copied per-repo.

This directory holds:

- `AGENTS.md` — the tracked global base and source of truth
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

Stow the tracked file to the proposed XDG path, then point every tool at that link:

```text
$XDG_CONFIG_HOME/agents/AGENTS.md    # usually ~/.config/agents/AGENTS.md
```

Nothing reads that path natively yet, so wire each tool to it:

- **Codex** reads `~/.codex/AGENTS.md` as its global layer, then merges repo files on
  top. Symlink it to the canonical file.
- **Claude Code** reads `~/.claude/CLAUDE.md` as global user memory and supports
  `@path` imports (relative or absolute paths, up to 4 hops). The wiring script uses
  an absolute import so Claude-only memory can still live below it. Current Claude
  documentation also supports home-relative imports; see
  [memory imports](https://code.claude.com/docs/en/memory).
- **opencode** reads `~/.config/opencode/AGENTS.md` as its global layer. Symlink it.
- **Gemini CLI** reads `~/.gemini/GEMINI.md`. Symlink it to the canonical file.
- **Other tools** take the same treatment: a symlink where the tool reads a fixed
  file, or a one-line pointer where it does not.

Per project: a repo's own `AGENTS.md` is picked up automatically and layers on top
(Codex and opencode read it natively; for Claude Code a project `CLAUDE.md` containing
`@AGENTS.md` imports the project file; Gemini supports `@./AGENTS.md` in
`GEMINI.md`). A prose pointer is not an automatic import. Verify the effective
context in each client rather than assuming that a reference loaded another file.
Keep project files short — stack, commands, local conventions — and let them
lean on the global base for everything generic. See
`templates/project/project-agents.md` for a starter.

## Setup, start to finish

1. From the repository root, run `scripts/stow.sh`. It links this directory's
   `AGENTS.md` to `~/.config/agents/AGENTS.md`, then wires each tool to that link.
2. Edit `global-agents/AGENTS.md`, commit the change, and pull it on other machines.
   The links pick up the change without another copy step.
3. In each project, keep a short `AGENTS.md` for stack, commands, and conventions; it
   overrides the global base where they differ.
4. Verify each tool picked it up — see below.

The installer wires all supported clients; it has no single-client flag. For a
project-local choice without global changes, use the shared-base setup below.

## Verifying it loaded

**Codex** — start a new session in the project after installation. Check the global
link with `ls -l "${CODEX_HOME:-$HOME/.codex}/AGENTS.md"`, then ask Codex to report
which instruction files it loaded. The CLI has no supported `--print-instructions`
option; see the [official AGENTS.md guidance](https://learn.chatgpt.com/docs/agent-configuration/agents-md).

**Claude Code** — run `/memory` inside a session: it lists the loaded memory files,
their paths, and load order. Because the base comes in by `@` import, it shows up as
part of the expanded `~/.claude/CLAUDE.md`. If it is missing, inspect the import
path and client configuration. `/context` shows how much of the window it consumes.
At the file level:

```bash
cat ~/.claude/CLAUDE.md            # should contain the @<absolute path> import line
cat ~/.config/agents/AGENTS.md     # the canonical file the import resolves to
```

**opencode / Gemini CLI** — inspect the symlink first:

```bash
ls -l ~/.config/opencode/AGENTS.md ~/.gemini/GEMINI.md   # both point at the canonical file
```

Then check the effective context in a fresh session. Gemini CLI provides
`/memory show` and `/memory reload`; see its
[context documentation](https://geminicli.com/docs/cli/gemini-md/). A link existing
on disk does not prove that the client loaded it.

## Updating

Edit the tracked `global-agents/AGENTS.md` file only. Symlinks pick changes up
immediately; Claude's import reads it fresh each session. Commit and pull that file to
share changes. Never edit the per-tool links.

## Pinning a project baseline

For a reviewed project pin instead of personal installation, follow the
[shared-base setup guide](../docs/project-shared-base.md). It uses a Git submodule
at `.agent-baseline`, a project-owned overlay and selected relative skill links.
The guide covers existing instructions, clean clones, CI, updates, customization
and client discovery checks. No global files are changed.

Use `.agent-baseline/global-agents/AGENTS.md` as the reusable base. Foundry's root
`AGENTS.md` describes work on this repository and must not become a consumer's
instruction file. The project setup helper and `--check` never advance the pin.
