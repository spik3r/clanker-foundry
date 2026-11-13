---
id: CF-uml-diagrams-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add UML-style sequence and class diagram profiles to diagramgen

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by
supervised subagents and coordinator review of the diff and of the rendered
pictures. These are the third and fourth profiles of step 4b in the
[generator pipeline plan](../docs/generator-pipeline-plan.md), requested by the
repository owner on 12 November 2026.

Scope: profiles `sequence-diagram/v1` and `class-diagram/v1` in
`modeling/diagram/`. A sequence model lists participants, ordered messages,
notes and non-nested frames. A class model lists classes, interfaces and enums
with members, and typed relations. Neither carries coordinates. `diagramgen`
emits an inert SVG, the canonical model and the manifest using fixed layout
algorithms. Also four example models with committed output, two profile
documents, two authoring schemas with generator-agreement tests, registry
entries, and small shared SVG helper packages. `diagramgen` becomes 1.2.0 and
the committed process and dashboard examples are regenerated for the new
version. Also this task record and new immutable Project Memory files.

Exclusions: any claim of UML conformance beyond the listed notation, other
UML diagram kinds, the harness code, the contract text and other generators.

Checks: gofmt, `go vet` and `go test -race` in `modeling/diagram` and
`modeling/gencontract`, the node schema tests, the workflow inventory test, the
repository validator, whitespace check and committed Project Memory validation.
