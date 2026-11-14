---
id: CF-bpmn-pool-layout-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Inset BPMN lanes and keep process flows inside the pool

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff and of the rendered
pictures. Rendering the generated BPMN with bpmn-js, recorded in
`CF-bpmn-import-check-v1`, showed three layout problems: the pool label drawn
over the lane labels, a loop line outside the pool, and two edges into
different end events sharing a run.

Scope, under `modeling/diagram/`: a pool box that contains the channel zones,
lanes inset 30 units for the pool label band in both SVG and BPMN, gutter slot
ordering and channel fallback so flows into different targets never share a
collinear run, tests reading both outputs on the examples and on generated
random models, the profile text, `diagramgen` 1.2.2 and regenerated examples.
Also this task record and new immutable Project Memory files.

Exclusions: other profiles' layout, the contract, the harness and other
generators.

Checks: gofmt, `go vet` and `go test -race` in `modeling/diagram` and
`modeling/gencontract`, the node schema tests, the bpmn-moddle check, the
repository validator, whitespace check and committed Project Memory validation.
