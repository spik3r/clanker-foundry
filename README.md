# Clanker Foundry

Reusable skills, instructions and tools for AI-assisted engineering, from
investigating a change to building it and reviewing the evidence.

## Four core parts

### 🧩 Model-driven pipeline

Turn a reviewed design into supported code and diagrams, then check the result
against independent expectations.

[Explore the pipeline →](docs/model-driven-pipeline.md) · [Diagram starters](templates/diagrams/README.md) · [Document templates](templates/README.md)

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

## Try a model review without installing skills

From a Foundry checkout, run the bundled lifecycle example with Go 1.27.3,
Bash and Python 3. The replay uses a disposable copy of the model.

```bash
cd modeling/lifecycle
review_dir="$(mktemp -d)"
go run ./cmd/foundryreview --tutorial > "$review_dir/unrun.html"
./revisionreview/acceptance-html.sh > "$review_dir/replay.html"
printf 'Open the reports in %s\n' "$review_dir"
```

Open the HTML files in your browser. In `unrun.html`, the expectations are
declared but execution is **not run**. In `replay.html`, compare the same
reviewed expectations with what each revision actually returned. **Matching**
means the observations agreed for the supplied fixture, not that the
requirements are correct or the change is approved.

[Follow the full lifecycle walkthrough](modeling/lifecycle/revisionreview/README.md).

## Apply it to your project

1. **Model only what needs to change.** Keep one source of truth for the selected
   behaviour. Generate the view or code from it and check the output for drift.
2. **Write acceptance separately.** Keep expected outcomes in a reviewed fixture,
   not a file generated from the same model. Exercise the actual entry point; a
   deliberately broken implementation should fail even when generation is current.
3. **Bring both to review.** Keep model validation, generated artefact checks and
   behavioural acceptance distinct. A consumer adapter can emit
   `foundry-review-evidence/v1` for `foundryreview`; mark missing execution
   **not run** and evidence for a different revision **stale**. The renderer
   presents observations; it does not independently authenticate them.

- [Consumer gates and Mement0rq article pilot](docs/consumer-pipeline-gates.md):
  pin the generator, check generated diagrams, then verify real browser outcomes.
- [Kai’s homepage and blog link](https://blog.kaitait.com/dev/thoughts/2026/11/28/i-only-wanted-to-add-a-link.html):
  the model owns just the introduction and two links. Hiding the blog link in
  CSS left the generated HTML current but broke the independent browser test.
- [Go model-builder example](examples/model-builder-go/README.md): see a scoped
  example, its review packet and the recorded evidence.
- [Generation commands](modeling/foundrygen/README.md): see how to plan, generate
  and check declared outputs with a pinned Foundry revision.

## Install agent skills

The recommended install uses Git, Bash and GNU Stow. Keep the clone in a permanent
location: installed assets are symlinks back to it. Review the
[installation guide](docs/getting-started.md) and script before running them;
the install links assets into both `~/.agents/` and `~/.claude/` and wires global
instructions for supported clients.

```bash
mkdir -p "$HOME/code"
git clone https://github.com/spik3r/clanker-foundry.git "$HOME/code/clanker-foundry"
cd "$HOME/code/clanker-foundry"
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
