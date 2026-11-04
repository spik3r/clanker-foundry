# Clanker Foundry

Clanker Foundry is a small, opinionated engineering framework for AI coding
agents: reusable skills, instructions and workflows that help people investigate,
build and review changes with evidence. Start with the skill pack on its own, or
explore the optional tools for task history, modeling and orchestration.

## Get started

The recommended install uses Git, Bash and GNU Stow. Keep the clone in a permanent
location: installed assets are symlinks back to it. Review the
[installation guide](docs/getting-started.md) and script before running them;
the install links assets into both `~/.agents/` and `~/.claude/` and wires global
instructions for supported clients.

```bash
git clone https://github.com/spik3r/clanker-foundry.git ~/code/clanker-foundry
cd ~/code/clanker-foundry
cat scripts/stow.sh global-agents/wire-global-agents.sh  # review before installing
scripts/stow.sh
```

Start a new agent session, then try a scoped request such as:

> Use `investigate-codebase` to trace how this request is handled. Do not edit files.

The [getting-started guide](docs/getting-started.md) also covers project-scoped or
single-skill use, conflicts and validation. Existing files can stop installation;
inspect conflicts rather than blindly adopting them.

## Find your way around

### Skills and working practices

[Skills](docs/skills.md) are self-contained procedures in the open `SKILL.md`
format. For example, `engineer` supports implementation, while `review-change`
and `investigate-codebase` default to read-only. The catalog has the full list,
examples and suggested setup. [Workflows](workflows/) describe repeatable processes;
[templates](templates/) and [checklists](checklists/) support their outputs and checks.

[Agent profiles](agents/README.md) outline explorer, planner, builder and independent
reviewer roles. The [feature-delivery workflow](workflows/feature-delivery.md) shows
how they can work together; the host still chooses and launches the models.

### Start with AGENTS.md

A project's `AGENTS.md` holds its standing facts, rules and commands; skills hold
repeatable procedures. This repository's [AGENTS.md](AGENTS.md) is its source of
truth for agent instructions, with tool-specific pointers rather than copies.
For a shared machine-wide base and a team baseline, see the
[global agent guide](global-agents/README.md).

### Tasks and Project Memory

The [task workflow guide](docs/task-workflows.md) explains ownership and handoffs
using your existing tracker or task files. Optional
[Project Memory](skills/project-memory/README.md) records pinned task sources and
immutable local implementation events, linking completion declarations to staged
content. It is evidence, not a claim lock, proof a test ran, or a replacement for
the canonical task store. Foundry uses it for its own changes; consumers need not.
The [format contract](skills/project-memory/FORMAT.md) covers compatibility and
validation limits.

### Model-driven pipeline

What if a reviewed design could produce code and views that are checked against
independent expectations? People and AI can clarify intent and draft a structured
model; deterministic tools validate and translate the supported parts. The model
is the design, not the AI. Repeatability does not prove correctness, so review and
independent checks matter.

The [plain-language pipeline guide](docs/model-driven-pipeline.md) explains the
motivation and future goal. To try what exists today, follow the
[pilot testing guide](docs/model-driven-pilot-testing.md): a bounded task-lifecycle
example, not a general app generator. It is separate from the runtime below.

### Orchestration

The optional [orchestration runtime](runtime/README.md) coordinates typed tasks,
guarded transitions and supervised workers on a single host. It has explicit
ownership, authority and isolation limits; existing task claims do not transfer
automatically. Start with the runtime’s fake-worker demo and read the
[rollout guide](docs/deterministic-orchestration-rollout.md) before enabling dispatch.
Neither the runtime nor the model-driven example is required to use the skill pack.
