---
id: CF-process-diagram-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a BPMN-style process-diagram generator

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff and of the rendered
pictures. This is the first profile of step 4b in the
[generator pipeline plan](../docs/generator-pipeline-plan.md), requested by the
repository owner on 12 November 2026.

Scope: a new Go module `modeling/diagram/` with the `diagramgen` command and
the `process-diagram/v1` profile. A model lists lanes, start and end events,
tasks, exclusive and parallel gateways and sequence flows, with no coordinates.
The generator emits an inert SVG, BPMN 2.0 XML with diagram coordinates, the
canonical model and the contract manifest, using a fixed layout algorithm. Also
three example models with committed output, a profile document, an authoring
schema, a `diagram-model-go` CI workflow with its entry in the workflow
inventory test, a registry entry in `modeling/gencontract/generators.json`,
this task record and new immutable Project Memory files.

Exclusions: class, sequence and dashboard profiles; reading BPMN; any claim of
BPMN conformance beyond the listed emitted elements; the lifecycle process
view; other generators, the harness code and the contract text.

Checks: gofmt, `go vet` and `go test -race` in `modeling/diagram` and
`modeling/gencontract`, the workflow inventory test, the repository validator,
whitespace check and committed Project Memory validation.
