# Kai's Agent Skills

A small, modular engineering skill pack using the open Agent Skills `SKILL.md` format,
plus the templates, workflows, and checklists that support it.

## Repository layout

| Path | Contents |
|---|---|
| `skills/` | The skill pack (see below) |
| `templates/` | Fill-in starting points: `prompts/`, `architecture-review/`, `adr/`, `pr/`, `project/` |
| `workflows/` | Repeatable procedures, e.g. `bug-investigation.md` |
| `checklists/` | Verification lists: `security.md`, `code-review.md` |
| `docs/` | Framework and external-skill guidance |
| `scripts/stow.sh` | Stows skills and global instructions into supported agent locations |
| `global-agents/` | Machine-wide global `AGENTS.md` example and setup guide |
| `AGENTS.md` | Repo instructions and conventions — the source of truth for agents (`CLAUDE.md` and `GEMINI.md` point at it) |

## Included skills

Short summaries below; each skill directory has its own `README.md` with detail
on modes, behaviour, output, and example invocations.

| Skill | Purpose | Default behaviour |
|---|---|---|
| [`engineer`](skills/engineer/README.md) | End-to-end investigation, planning, implementation and validation | May edit in implementation or repair mode |
| [`review-change`](skills/review-change/README.md) | Independent PR, branch, commit or working-tree review | Read-only |
| [`investigate-codebase`](skills/investigate-codebase/README.md) | Trace an unfamiliar code path or subsystem | Read-only |
| [`architecture-review`](skills/architecture-review/README.md) | Assess boundaries, coupling, data, reliability, security and evolution | Read-only unless asked for a plan |
| [`performance-investigation`](skills/performance-investigation/README.md) | Measure and isolate performance bottlenecks | Evidence-first; edits only for scoped experiments |
| [`safe-refactor`](skills/safe-refactor/README.md) | Improve structure while preserving observable behaviour | Small edits with continuous validation |
| [`adr`](skills/adr/README.md) | Create or update an Architecture Decision Record | Writes documentation only |
| [`checkpoint`](skills/checkpoint/README.md) | Create, refresh, or resume durable handoffs (`checkpoints/`) across sessions and agents | Writes documentation only |
| [`context-offload`](skills/context-offload/README.md) | Create a compact, replaceable `context.md` handoff for one active task | Writes documentation only |
| [`braindump-distiller`](skills/braindump-distiller/README.md) | Turn unstructured ideas into a phased plan or interactive checklist | Read-only unless asked to save a plan |

## Suggested usage

Examples:

```text
Use engineer in implementation mode to add ...
Use engineer in validation mode on my current changes.
Use review-change to review this branch against main. Do not edit files.
Use investigate-codebase to trace how refunds flow from the API to Stripe.
Use architecture-review on the authentication and authorisation boundaries.
Use performance-investigation to diagnose the slow portfolio endpoint.
Use safe-refactor to split this service without changing behaviour.
Use adr to record our decision to use managed identity.
Use checkpoint to write a handoff before I switch models.
```

`templates/prompts/` holds fill-in-the-blank versions of the most common requests.

## Installation

### Quick start (recommended)

Use GNU Stow to create every personal skill symlink and wire the shared global agent
instructions. Keep this repository in a permanent location; do not run Stow from a
temporary clone.

1. Install GNU Stow with your system package manager. On macOS with Homebrew:

   ```bash
   brew install stow
   ```

2. Clone this repository, then enter it:

   ```bash
   git clone <repository-url> ~/code/clanker-foundry
   cd ~/code/clanker-foundry
   ```

3. Review [scripts/stow.sh](scripts/stow.sh), then run it:

   ```bash
   scripts/stow.sh
   ```

4. Start a new agent session. The script has symlinked this pack into
   `~/.agents/skills` for Codex and `~/.claude/skills` for Claude Code, stowed the
   canonical global instructions, and run the existing wiring for Codex, Claude Code,
   opencode, and Gemini CLI.

The script creates missing target directories and uses `stow --restow`, so it is safe
to run again after pulling updates. It may stop if an existing non-Stow file conflicts
with a target; inspect and resolve the conflict rather than using Stow's `--adopt`
option blindly.

### Project-scoped use

The Stow script installs personal skills. For a skill that belongs only to one project,
use that project's supported skill directory (`.agents/skills/` or `.claude/skills/`)
and keep it under that project's version control.

### One-off manual link

For a temporary or single-skill installation, run these commands from this repository:

```bash
# Add one skill to Codex.
ln -s "$PWD/skills/engineer" "$HOME/.agents/skills/engineer"

# Remove that symlink later. `unlink` will not remove a directory.
unlink "$HOME/.agents/skills/engineer"
```

Do not manually replace a link managed by Stow; use `scripts/stow.sh` after changing
the pack instead.

## Recommended setup

Install these globally:

- `engineer`
- `review-change`
- `investigate-codebase`
- `safe-refactor`
- `adr`
- `checkpoint`

Consider keeping these global but invoke them manually when needed:

- `architecture-review`
- `performance-investigation`

Keep stack-specific workflows, commands and architecture rules in project-level skills or `AGENTS.md`.

For a machine-wide behaviour base shared by every project, see `global-agents/`.

## External skills

See [useful external skills](docs/external-skills.md) for the official OpenAI curated
`security-threat-model`, Caveman, and complementary skills from
[mattpocock/skills](https://github.com/mattpocock/skills).

## Design notes

- Each skill is self-contained. No skill depends on another skill being installed.
- Review and investigation skills default to read-only.
- The descriptions include explicit trigger boundaries so agents are less likely to invoke the wrong workflow.
- The pack avoids scripts and tool allowlists to remain portable across compatible clients.
- The skills complement `AGENTS.md`: global facts and standing rules belong in `AGENTS.md`; repeatable procedures belong in skills and workflows; reusable output shapes belong in templates; verification lists belong in checklists.

## Validation

The folders follow the Agent Skills open format:

- directory name matches the `name`;
- names use lowercase letters and hyphens;
- each `SKILL.md` contains `name` and `description` frontmatter;
- all skills remain well below the recommended 500-line limit.

You can also validate them with the `skills-ref` validator when installed:

```bash
skills-ref validate ./skills/engineer
```
