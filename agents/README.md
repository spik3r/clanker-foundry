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

## Model selection

Profiles specify a portable tier, not a vendor model name: `fast`, `balanced`, or
`flagship`. [`model-map.conf`](model-map.conf) maps those tiers to each client. The
current defaults are Claude `haiku` / `sonnet` / `opus` and Codex
`gpt-5.6-terra` / `gpt-5.6-terra` / `gpt-5.6-sol`.

Resolve the preferred concrete model before launching an agent:

```bash
scripts/agent-model.sh claude explorer-researcher
# haiku

scripts/agent-model.sh codex builder-developer
# gpt-5.6-sol
```

The profiles themselves do not force a host model. Claude Code therefore inherits its
normal session model when it loads a profile; select the resolved model in the launch
or orchestration surface. Codex orchestration likewise selects the resolved model when
it starts the subagent. Edit `model-map.conf` to match the model aliases available to
your account. For a machine-only change, create untracked `model-map.local.conf`; it
overrides matching shared entries and stays out of Git and Stow. Run
`scripts/agent-model.sh` to inspect the selected value.

Pass a concise task brief when launching a profile: outcome, scope, repository or
sources, constraints, and the preceding stage's handoff. The profiles define each
agent's role and output; the task brief provides the task-specific facts.
