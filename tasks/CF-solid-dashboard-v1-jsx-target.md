---
id: CF-solid-dashboard-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a Solid component target to the dashboard profile

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is the first
step of the second-consumer trial in the
[generator pipeline plan](../docs/generator-pipeline-plan.md). The repository
owner approved the trial on 15 November 2026: generated SolidJS components
for the heisentick research pages.

Scope: an optional `target` member (`html` default, or `solid`) of
`dashboard/v1` in `modeling/diagram/`. With `solid`, `diagramgen` emits
`Dashboard.jsx` and `dashboard.css`, reading host CSS custom properties, with
chart geometry shared with the HTML output. Models without the member keep
their digest and bytes. Also: `diagramgen` 1.3.0 with regenerated examples,
one new example, the schema, docs, the registry, this task record and new
immutable Project Memory files.

Exclusions: the heisentick application itself, other profiles' layouts, the
harness code and the contract text.
