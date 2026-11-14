---
id: CF-bpmn-import-check-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Check generated BPMN with a BPMN implementation

## Implementation claim

Owner: Claude pipeline coordinator, with delegated work by a supervised
subagent and coordinator review. The
[generator pipeline plan](../docs/generator-pipeline-plan.md) lists importing
the generated BPMN XML into a BPMN tool as never checked.

Scope: an opt-in script `modeling/diagram/verify/bpmn-moddle.mjs` that loads
generated `diagram.bpmn` files with the bpmn-moddle library and fails on any
warning, unresolved reference or round-trip warning; its README; one line in
the diagram README; corrected statements in the process profile and the plan
about what has been checked. No generator code changes and no dependency is
added to any package file. Also this task record and new immutable Project
Memory files.

Exclusions: OMG XSD validation, which is not reachable from the environment
used, and any rendering or layout change.

Checks: the script on the three committed examples and on a deliberately
broken file; gofmt, `go vet` and `go test -race` in `modeling/diagram`; the
repository validator, whitespace check and committed Project Memory validation.
