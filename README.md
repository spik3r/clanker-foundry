# Clanker Foundry

Reusable skills, instructions and tools for AI-assisted engineering, from
investigating a change to building it and reviewing the evidence.

## Four core parts

### 🧩 Model-driven pipeline

Turn a reviewed design into supported code and diagrams, then check the result
against independent expectations.

[Explore the pipeline →](docs/model-driven-pipeline.md)

### ⚙️ Orchestration

Coordinate typed tasks and supervised workers with an optional single-host runtime.

[Explore orchestration →](runtime/README.md)

### 🛠️ Skills

Give coding agents reusable procedures for investigating, building and reviewing
changes.

[Browse the skills →](docs/skills.md)

### 📖 Project Memory

Record which task, file versions and declared checks belong to each repository
change.

[Explore Project Memory →](skills/project-memory/README.md)

**Start with skills; add the other parts when needed.** These are separate tools,
not an automatically connected system: modeling supports bounded examples,
orchestration requires explicit rollout, and Project Memory records evidence
rather than granting ownership or proving that tests ran.

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

The installation guide also covers project-scoped or single-skill use, conflicts
and validation. Existing files can stop installation;
inspect conflicts rather than blindly adopting them.

For a reusable team baseline without global installation, use the
[project-local setup](docs/project-shared-base.md): pin `.agent-baseline`, keep a
short project overlay and link only the skills you select.

## Project instructions and direction

A project's `AGENTS.md` holds its standing facts, rules and commands; skills hold
repeatable procedures, with tool-specific pointers rather than copied instructions.

See the [current direction and consumer gates](docs/current-direction.md) for
priorities, remaining questions and the boundary between demonstrations and real use.
