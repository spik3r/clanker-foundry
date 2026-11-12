---
id: CF-dashboard-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a static dashboard profile to diagramgen

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff and of the rendered
pages. This is the second profile of step 4b in the
[generator pipeline plan](../docs/generator-pipeline-plan.md), requested by the
repository owner on 12 November 2026.

Scope: profile `dashboard/v1` in `modeling/diagram/`. A model holds inline
datasets and stat, bar, line and table widgets in a row layout. `diagramgen`
emits one self-contained HTML file with inline styles and SVG charts, no script
and no external reference, plus the canonical model and manifest. Numbers are
parsed and canonicalised as decimal text and chart geometry uses integer
arithmetic. Also three example models with committed output, a profile
document, an authoring schema with a generator-agreement test, a registry
entry, and a change to the registry consistency test so one generator may
serve several profiles. `diagramgen` becomes 1.1.0, the shared model size limit
becomes 512 KiB, and the committed process-diagram examples are regenerated for
the new version. Also this task record and new immutable Project Memory files.

Exclusions: live data, computed values, interactivity, class and sequence
diagram profiles, the harness code, the contract text and other generators.

Checks: gofmt, `go vet` and `go test -race` in `modeling/diagram` and
`modeling/gencontract`, the node schema tests, the workflow inventory test, the
repository validator, whitespace check and committed Project Memory validation.
