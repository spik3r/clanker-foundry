# AGENTS.md — clanker-foundry

This file is the single source of truth for agent instructions in this repository.
Every other tool-specific file (`CLAUDE.md`, `GEMINI.md`) is a pointer to this file.
Edit this file; never edit the pointers.

## What this repo is

A personal library of agent assets: skills in the open Agent Skills `SKILL.md` format,
plus the templates, workflows, and checklists those skills draw on, plus a global
`AGENTS.md` base and the wiring to share it across tools. There is no build, no test
suite, and no runtime code. Changes are documentation; validation is reading
carefully and checking every link and path resolves.

## Layout

- `skills/` — the skill pack. One directory per skill, each holding a `SKILL.md`
  (the procedure the agent follows) and a `README.md` (what the skill is, when to
  use it, and how it behaves, for humans picking skills). Install by symlinking a
  skill directory into `~/.agents/skills/` (Codex), `~/.claude/skills/` (Claude
  Code), or a repo's `.agents/skills/`.
- `templates/` — fill-in-the-blank starting points:
  - `templates/prompts/` — prompt templates for common agent requests
  - `templates/architecture-review/` — report structure for architecture reviews
  - `templates/adr/` — Architecture Decision Record template
  - `templates/pr/` — pull request description template
  - `templates/project/` — starter `AGENTS.md`, `README.md`, and checklist for new repos
- `workflows/` — named, repeatable procedures (e.g. `bug-investigation.md`). A
  workflow is a process to follow; a skill is when and why to follow it.
- `checklists/` — verification lists (`security.md`, `code-review.md`) meant to be
  walked item by item, with evidence, not skimmed.
- `docs/` — framework and external-skill guidance.
- `scripts/stow.sh` — GNU Stow installation script for the pack and global instructions.
- `global-agents/` — the machine-wide global `AGENTS.md` example, a setup guide, and
  `wire-global-agents.sh` to point Codex, Claude Code, opencode, and Gemini CLI at
  one canonical file.
- `CLAUDE.md`, `GEMINI.md` — pointers to this file. See "One source of truth" below.

## Conventions

Writing follows the rules in `global-agents/AGENTS.md`. In short: plain words, no
filler or achievement language, describe behaviour, separate facts from assumptions
from opinions.

Skills follow the Agent Skills open format:

- directory name matches the `name` in frontmatter;
- names use lowercase letters and hyphens;
- `SKILL.md` frontmatter carries `name` and `description`;
- the description states trigger boundaries — when to use the skill and when not to;
- keep each `SKILL.md` well under 500 lines;
- skills are self-contained: no skill depends on another being installed.

Division of responsibility:

- standing facts and rules belong in `AGENTS.md` files;
- repeatable procedures belong in `workflows/` and `skills/`;
- reusable output shapes belong in `templates/`;
- verification lists belong in `checklists/`.

Review and investigation skills default to read-only. Do not add edit behaviour to
them without changing their descriptions to match.

## One source of truth for agent instructions

`AGENTS.md` is the settled convention for project-level agent instructions and is
read natively by Codex, opencode, and a growing set of tools. Tools that read a
different filename get a pointer, not a copy:

- `CLAUDE.md` — one line: `@AGENTS.md`. Claude Code expands the import.
- `GEMINI.md` — a one-line pointer telling the agent to read `AGENTS.md`.
- Codex, opencode — no file needed; they read `AGENTS.md` directly.

The same pattern applies one level up: `global-agents/` holds one canonical global
file, and each tool's user-level path symlinks or imports to it. Copies drift;
pointers do not. When adding a new tool to this repo, add a pointer file or a
symlink — never a duplicate of this file's content.

## Validation

After changing a skill:

- confirm the frontmatter still parses and `name` still matches the directory;
- check every path referenced from `README.md`, this file, and the skills resolves;
- if `skills-ref` is installed, run `skills-ref validate ./skills/<name>`.

After moving or renaming anything, grep the repo for the old path — README install
snippets, skill cross-references, and template links go stale silently.
