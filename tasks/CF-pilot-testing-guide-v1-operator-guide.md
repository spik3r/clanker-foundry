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

## Observed local result

Added a testing guide and README entry point without changing executable source or
existing generated artifacts. Its five Bash blocks ran in order on this checkout
with Go 1.27.3, Python 3 and an outside-checkout private workspace: the fixed
rehearsals, non-executing preflight and separately invoked candidate replay all
returned the documented observations. An earlier full `/tmp` caused a rehearsal
build failure; the guide now supports an explicit outside-checkout storage parent
and sets temporary and cache paths. An earlier draft expected the wrong successful
proposal cleanup label; independent review caught it and the corrected full run
passed. Neither unsuccessful attempt is counted as a pass.

A separate read-only reviewer inspected the final corrections, checked all five
blocks with `bash -n`, and ran the two summary readers on independently observed
rehearsal receipts. Repository validation and Markdown relative links passed;
optional `skills-ref` was unavailable. The source checkouts and receipts concern
fixed invented examples, not human approval, visual review or deployment.

The reviewer separately reran focused serial Go 1.27.3 race tests for tutorial,
rehearsal, merge rehearsal, proposal replay and both handoff CLI/package pairs:
all seven packages passed with workspace-backed temporary storage and cache. A
prior run failed when `/tmp` filled and is not counted as a pass. The reviewer
observed an unchanged lifecycle module and empty temporary directory afterward.
The initial start event accidentally supplied all three file names as one
comma-separated value; the next checkpoint records the exact three paths before
completion rather than editing immutable history.
