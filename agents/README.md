# Sample subagents

Reusable subagent profiles for a deliberate four-stage delivery pipeline:

1. `explorer-researcher` — bounded, read-only discovery and distilled evidence.
2. `architect-planner` — read-only planning before implementation.
3. `builder-developer` — implementation and validation from an approved plan.
4. `reviewer-verifier` — independent, read-only audit after the build.

The files use the Markdown frontmatter convention recognised by Claude Code's personal
agent directory. `scripts/stow.sh` links them into `~/.claude/agents/`. It also links
them into `~/.agents/agents/` as a stable, tool-neutral profile library for launchers
that use that location; configure other tools to load that directory when supported.

Pass a concise task brief when launching a profile: outcome, scope, repository or
sources, constraints, and the preceding stage's handoff. The profiles define each
agent's role and output; the task brief provides the task-specific facts.
