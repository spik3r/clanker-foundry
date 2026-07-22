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

### Codex

Personal skills are read from:

```bash
~/.agents/skills
```

Copy or symlink the skill folders:

```bash
mkdir -p ~/.agents/skills
cp -R skills/* ~/.agents/skills/
```

Or keep this pack in one location and symlink each folder:

```bash
for skill in skills/*/; do
  ln -sfn "$PWD/${skill%/}" "$HOME/.agents/skills/$(basename "$skill")"
done
```

For repository-scoped Codex skills, place them under:

```text
<repo>/.agents/skills/
```

### Claude Code

Personal skills are read from:

```bash
~/.claude/skills
```

Copy or symlink the same folders:

```bash
mkdir -p ~/.claude/skills
for skill in skills/*/; do
  ln -sfn "$PWD/${skill%/}" "$HOME/.claude/skills/$(basename "$skill")"
done
```

For repository-scoped Claude Code skills, place them under:

```text
<repo>/.claude/skills/
```

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
