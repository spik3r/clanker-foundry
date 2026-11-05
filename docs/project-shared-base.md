# Reuse a pinned project baseline

A project can keep a reviewed Foundry revision in a Git submodule at
`.agent-baseline`, a short project-owned `AGENTS.md`, and links to selected skills.
The project commits the pin and links, so collaborators get the same files without
installing anything in their home directories. Requirements: Git, Bash 3.2 or later,
and a filesystem with working symlinks (tested on macOS; CI is configured for Linux).

Use the reusable base at `.agent-baseline/global-agents/AGENTS.md`. Never install
Foundry's root `AGENTS.md` as your project's instructions: it governs maintenance
of this pack, including its own task store and validation rules. Project Memory,
modeling and orchestration remain separate, optional choices.

## First setup

Run from the root of your consumer Git repository. Review the upstream revision
and scripts before running code from them. Choose a full commit ID that includes
`scripts/project-setup.sh`; do not substitute a moving branch for the review.

```bash
reviewed_commit='<full-reviewed-commit-id>'
git submodule add https://github.com/spik3r/clanker-foundry.git .agent-baseline
git -C .agent-baseline checkout --detach "$reviewed_commit"
git add .gitmodules .agent-baseline
bash .agent-baseline/scripts/project-setup.sh engineer
```

The helper creates only project-local files:

| Path | Purpose |
|---|---|
| `AGENTS.md` | Editable project facts, commands and rules; explicit instruction to read the reusable base |
| `CLAUDE.md`, `GEMINI.md` | Imports of the reusable base and local `AGENTS.md` |
| `.agents/foundry-skills` | Selected skill names, one per line; an empty file selects none |
| `.agents/skills/engineer` | Relative link for clients that discover this skill location |
| `.claude/skills/engineer` | Relative link for Claude Code |

Pass more names as separate arguments, for example `engineer review-change`.
With no names on the first run, setup selects no skills. Later runs read the
selection file; they never infer a new selection from the contents of the pack.

Fill in the TODOs in `AGENTS.md`. Review and commit named paths, preserving any
unrelated work already in the project:

```bash
bash .agent-baseline/scripts/project-setup.sh --check
git add .gitmodules .agent-baseline AGENTS.md CLAUDE.md GEMINI.md \
  .agents/foundry-skills .agents/skills/engineer .claude/skills/engineer
git diff --cached --check
git diff --cached
git commit -m 'Pin shared agent baseline and select engineer skill'
```

The helper does not commit, fetch, install global files, run Stow, adopt task
tracking, or replace existing files. It requires the submodule checkout to match
the staged gitlink and `.gitmodules` to match its staged file, with one baseline
path and URL. The URL remains a review decision, not an authenticity check. Local
baseline changes, invalid names, missing skills, symlink parents and conflicting
destinations fail before any wiring is created. Do not run concurrent setup writers.
Filesystem errors can still interrupt creation; inspect and rerun after fixing them.

## Existing project instructions

Setup preserves existing instructions. It stops if an existing `AGENTS.md` lacks
this exact standalone line; add it manually after reviewing the existing rules:

```text
Read `.agent-baseline/global-agents/AGENTS.md` before applying the project rules below.
```

Existing regular `CLAUDE.md` and `GEMINI.md` must each contain these standalone
imports. Other content is preserved:

```text
@.agent-baseline/global-agents/AGENTS.md
@AGENTS.md
```

Reconcile overlapping instructions manually; the helper checks lines, not their
meaning. It refuses a root `AGENTS.override.md` because that can shadow `AGENTS.md`
in Codex. It also refuses symlinked instruction files or destination directories.
Keep project-specific rules in the root overlay and custom skills in separately
named directories, not inside the submodule. No changes to Foundry's own files are
needed to describe a consumer's stack or policies.

## Clean clones and CI

Use `git clone --recurse-submodules <your-project-url>`, or initialize an existing
clone explicitly before running the checker:

```bash
git submodule update --init -- .agent-baseline
bash .agent-baseline/scripts/project-setup.sh --check
```

CI must initialize this submodule too (for `actions/checkout`, set `submodules:
true`). Relative links survive a different checkout location. A missing baseline
is a setup error; the helper cannot be executed from an uninitialized submodule.
Do not silently fall back to an unrelated global installation.

`--check` is read-only: it verifies the staged pin, submodule configuration, clean
baseline, required instruction lines, selection and links. It does not initialize
anything, repair missing links or prove a client loaded the files. Run it after
cloning, changing the selection or reviewing a baseline update.

## Updates and customizations

Review upstream changes, including changes to instructions and selected skills.
Then stage the new pin in a normal project pull request:

```bash
new_reviewed_commit='<full-reviewed-commit-id>'
git -C .agent-baseline fetch origin
git -C .agent-baseline checkout --detach "$new_reviewed_commit"
git add .agent-baseline
bash .agent-baseline/scripts/project-setup.sh --check
git diff --cached --submodule=log
```

Commit the gitlink update and have collaborators run the initialization command
again after pulling. Setup never advances the pin. The project overlay is copied
only on first setup, so later template improvements require manual review and merge.
To roll back, check out the previous reviewed baseline commit, stage the gitlink,
run `--check`, and commit the rollback. Reconcile selections and their links
explicitly if that older revision lacks a currently selected skill.

To add a skill, add its name to `.agents/foundry-skills`, then rerun setup without
arguments and review both new links. To remove one, delete its selection line and
explicitly `unlink` its reviewed `.agents/skills/<name>` and
`.claude/skills/<name>` links. Setup and check reject leftover Foundry links for
unselected skills; they never delete files on your behalf. Commit the selection
and both link changes together. Unrelated custom skills remain untouched.

Other reusable assets stay available through the same pin: reference
`.agent-baseline/workflows/`, `.agent-baseline/templates/` and
`.agent-baseline/checklists/` as needed. Create project-authored outputs in your own
directories. The `.agent-baseline/agents/` profiles are reference material until
you explicitly configure the client's supported agent mechanism; this helper does
not install or launch subagents. References to these directories do not imply
automatic client discovery.

## Confirm client loading

A prose path in `AGENTS.md` is an instruction to read another file, not a parser
import. Codex users should explicitly ask the agent to read the reusable base and
confirm its loaded instruction chain; root overrides, nested rules and instruction
size limits can affect the result. Claude Code and Gemini CLI support `@` imports,
but the effective context still depends on client configuration and version.

Start a fresh session. In Claude Code use `/memory`; in Gemini CLI use
`/memory show` (or `/memory reload` after changes). Inspect the selected skill catalog too.
Codex supports symlinked skill folders in `.agents/skills`; other clients may need
their own discovery configuration. Do not assume `.agents/skills` is universal or
that a project skill automatically replaces a globally installed skill of the same
name. Review duplicates in the actual client before relying on the project pin.

Official references: [Codex instructions](https://learn.chatgpt.com/docs/agent-configuration/agents-md),
[Codex skills](https://learn.chatgpt.com/docs/build-skills),
[Claude Code memory](https://code.claude.com/docs/en/memory), and
[Gemini CLI context](https://geminicli.com/docs/cli/gemini-md/).
See also [context portability](agent-context-portability.md).

## Project pin versus personal installation

`scripts/stow.sh` is a separate personal choice: it links assets into home-directory
client locations and wires global instructions. Those links follow that checkout
as it changes. This project setup needs neither Stow nor a global installation; its
reviewed gitlink controls the team's base. Using both can introduce duplicate skills
or different global defaults, so inspect client context rather than assuming identical
behavior. Do not run Stow from a project's `.agent-baseline` as part of this recipe.
