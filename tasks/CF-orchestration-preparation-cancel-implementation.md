---
id: CF-orchestration-preparation-cancel
status: review
owner: Foundry cancellation fix
repositories: [spik3r/clanker-foundry]
---

# Stop cancelled attempts before provider invocation

## Implementation claim

The user requested a fresh readiness review and conditional branch push. The
review found a cancellation race while the isolated adapter awaits workspace
reads. This bounded fix is required before pushing. Base `c8b0224`, branch
`codex/deterministic-orchestration-plan`; staged, unstaged and untracked paths
were clean. Remote main was refreshed and is already included. Existing ownership
records and immutable evidence remain unchanged. Exact scope:

- `runtime/src/isolation.mjs`
- `runtime/test/isolation.test.mjs`
- `tasks/CF-orchestration-preparation-cancel-implementation.md`
- New immutable Project Memory source and event files for this task.

Reject a revoked attempt after input reads and immediately before its queued
provider call. Preserve unconfirmed cleanup while preparation remains pending;
after it settles without invocation, allow genuine cleanup confirmation. Add
regressions for delayed reads and cancellation between preparation and invocation.
Do not change Docker configuration, credentials, namespace authority or promotion.

## Validation and next action

Added abort checks after workspace reads and inside the queued provider callback.
Delayed-read and queued-callback regressions prove zero provider invocations and
zero Docker plan applications after cancellation. Cleanup stays unconfirmed while
reads are pending and becomes confirmed once preparation has stopped.

Node v24.22.2 passed all 308 runtime and Project Memory/workflow tests, without
skips. Pack, installer and whitespace checks passed; optional skills-ref is not
installed. The independent reviewer reran four isolated-adapter tests and the
original delayed-read reproduction, reporting no remaining readiness blockers.
No live provider or Docker call was needed for these cancellation guards.
Next: validate the committed full-diff evidence and push the reviewed branch only
if all gates pass. The user authorized that conditional push; no merge or
deployment is included.
