---
id: CF-research-page-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a research-page profile to diagramgen

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff and output. This is
the second step of the heisentick second-consumer trial in the
[generator pipeline plan](../docs/generator-pipeline-plan.md). The repository
owner chose this design on 20 November 2026. It replaces a static dashboard
because the target pages are interactive.

Scope: profile `research-page/v1` in `modeling/diagram/`. The model holds a
research page's outline and fixed text and names hand-written slots.
`diagramgen` emits a Solid component that reaches host components through
one adapter module, plus `slots.json`. Shared JSX and text helpers move to
internal packages without changing dashboard bytes. Also: diagramgen 1.4.0
with regenerated examples, one example, the schema with its agreement test,
docs, the registry, this task record and new immutable Project Memory files.

Exclusions: the heisentick application itself, interactive page logic, the
harness code and the contract text.
