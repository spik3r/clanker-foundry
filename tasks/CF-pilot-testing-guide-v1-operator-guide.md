---
id: CF-pilot-testing-guide-v1
status: in-progress
owner: Lifecycle pilot testing guide
repositories: [spik3r/clanker-foundry]
---

# Document how to test the lifecycle pilot

## Implementation claim

The coordinating owner requested a repository guide to try the current small
model-driven lifecycle pilot and understand its observable results. This
single-writer documentation lane is based on reviewed main
`81c6647b00c56cdb8341a160e4fc06faf4f8e55d`. The claim is a local proposal, not
a published reservation. The coordinating owner approved this scope after the
current open-PR and task/Project Memory checks; the existing pipeline introduction
is historical and the open process-view lane is separate.

Exact scope: new `docs/model-driven-pilot-testing.md`, one discoverability link
in `README.md`, this task record and new immutable Project Memory source/event
files for this task. Do not change the compiler, generated artifacts, runtime,
existing tutorial, fixed rehearsals, other task records or prior evidence. The
held process-view artifact and other contributors' work remain out of scope.

Provide executable commands and expected observations grounded in the merged
implementation. Separate fixed invented rehearsals from authoring a real design;
review preparation from separately authorized execution; generated structural
checks from independent behavioral expectations; and matching results from human
approval or production readiness. State local toolchain and trusted-host limits,
including that isolated temporary copies are not a security sandbox. Link to the
existing detailed references instead of replacing them.

## Validation and handoff

Run the documented local commands that the environment and authorization permit,
report actual observations and limits, and run repository and committed-diff
Project Memory validation. Obtain a separate review of the guide where available.
Normal branch publication and a draft PR are authorized; merge, deployment,
worker activation and formal human approval are not part of this claim.
