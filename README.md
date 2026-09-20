# Kai's Agent Skills

A small, modular engineering skill pack using the open Agent Skills `SKILL.md` format,
plus the templates, workflows, and checklists that support it.

## Repository layout

| Path | Contents |
|---|---|
| `skills/` | The skill pack (see below) |
| `agents/` | Ready-to-launch subagent profiles: explorer-researcher, architect-planner, builder-developer, reviewer-verifier |
| `templates/` | Fill-in starting points: `prompts/`, `architecture-review/`, `adr/`, `pr/`, `project/` |
| `workflows/` | Repeatable procedures, e.g. `bug-investigation.md` |
| `checklists/` | Verification lists: `security.md`, `code-review.md` |
| `docs/` | Framework and external-skill guidance |
| `scripts/stow.sh` | Stows skills, subagents, workflows, templates, checklists, and global instructions |
| `global-agents/` | Tracked machine-wide `AGENTS.md`, setup guide, and tool wiring |
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
| [`knowledge-note`](skills/knowledge-note/README.md) | Create or update connected notes in a knowledge base or vault | Edits only the named vault |
| [`orchestrator`](skills/orchestrator/README.md) | Coordinate scoped subagents through research, planning, building and independent review | Delegates only when the task benefits from orchestration |

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
Use orchestrator to coordinate this feature across research, planning, implementation and review.
```

`templates/prompts/` holds fill-in-the-blank versions of the most common requests.
It includes an [`explain-code`](templates/prompts/explain-code.md) prompt for tracing
unfamiliar code without editing it. [`workflows/feature-delivery.md`](workflows/feature-delivery.md)
shows how to run the four agent profiles as a feature delivery pipeline.
For reusable subagent profiles, see [agents/](agents/README.md):
explorer-researcher, architect-planner, builder-developer, and independent
reviewer-verifier.

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

4. Start a new agent session. For both `~/.agents/` and `~/.claude/`, the script has
   installed `skills/`, `agents/`, `workflows/`, `templates/`, and `checklists/` as
   symlinks to this repository. It has also stowed the tracked global instructions to
   `~/.config/agents/AGENTS.md` and wired Codex, Claude Code, opencode, and Gemini CLI
   to that canonical link.

   The subagent profiles use portable model tiers. Resolve their current client-specific
   model through `scripts/agent-model.sh <claude|codex> <agent-name>`; see the
   [subagent model mapping](agents/README.md#model-selection).

The script creates missing target directories and uses `stow --restow`, so it is safe
to run again after pulling updates. Repository-only `docs/` and `scripts/` are not
installed. The script may stop if an existing non-Stow file conflicts with a target;
inspect and resolve the conflict rather than using Stow's `--adopt` option blindly.

Older installs may have a copied `~/.config/agents/AGENTS.md`. If it matches the tracked
file, the script moves it to `AGENTS.md.pre-stow` before creating the link. If it differs,
the script stops so you can merge those local changes into `global-agents/AGENTS.md`.
Commit and pull changes to the tracked file to share them across machines.

### Project-scoped use

The Stow script installs personal agent assets. For an asset that belongs only to one
project, use the matching project directory, such as `.agents/skills/`,
`.agents/workflows/`, or the equivalent `.claude/` path, and keep it under that
project's version control.

### One-off manual link

For a temporary or single-skill installation, run these commands from this repository:

```bash
# Add one skill to Codex.
ln -s "$PWD/skills/engineer" "$HOME/.agents/skills/engineer"

# Remove that symlink later. `unlink` will not remove a directory.
unlink "$HOME/.agents/skills/engineer"

# Add one Claude Code subagent profile.
ln -s "$PWD/agents/explorer-researcher.md" "$HOME/.claude/agents/explorer-researcher.md"

# Remove that profile later.
unlink "$HOME/.claude/agents/explorer-researcher.md"
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
- `orchestrator`

Keep stack-specific workflows, commands and architecture rules in project-level skills or `AGENTS.md`.

For a machine-wide behaviour base shared by every project, see
[`global-agents/`](global-agents/README.md).

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

Run the validator after any change:

```bash
scripts/validate.sh
```

It checks that the folders follow the Agent Skills open format:

- directory name matches the `name`;
- names use lowercase letters and hyphens;
- each `SKILL.md` contains `name` and `description` frontmatter;
- all skills remain well below the recommended 500-line limit;
- each skill ships a `README.md`, model mappings are nonempty, shell scripts parse, and
  every relative Markdown link in tracked or unignored untracked Markdown resolves on disk.
  Inline links are checked; anchors, reference-style links, and paths in inline code are
  outside this check.

CI runs the same script on every push and pull request. When the `skills-ref` validator
is installed, both the script and CI additionally run it per skill:

```bash
skills-ref validate ./skills/engineer
```

Installer regression tests run with:

```bash
bash tests/test-install.sh
```

## Pin a project baseline

For team repositories, pin a reviewed baseline with the [submodule recipe in the global
agent guide](global-agents/README.md#pinning-a-project-baseline). A missing pinned
baseline is a setup error. Never commit secrets or local model overrides.
