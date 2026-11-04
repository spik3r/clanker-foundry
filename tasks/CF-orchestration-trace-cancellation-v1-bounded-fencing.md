---
id: CF-orchestration-trace-cancellation-v1
status: in-progress
owner: Foundry cancellation trace contributor
repositories: [spik3r/clanker-foundry]
---

# Trace owner cancellation, retained cleanup and retired fencing

## Implementation claim

Owner: Foundry cancellation trace contributor. Base: main at
6eb8049f7238d847caefe3450920ed7c599f9b9f. Branch:
codex/orchestration-trace-cancellation-v1. Remote main was verified at this exact
revision; the live open-PR inventory was empty and remote branches showed no
cancellation trace lane. The prior completion trace is done and reserves no work.
This local claim does not publish ownership or override a new competing owner.

Write only runtime/model-trace-v1/check.mjs, manifest.json, README.md,
views/state.dot, views/sequence.md, testdata/cancellation-oracle.json,
runtime/test/model-trace-v1.test.mjs, this task and its new immutable Project
Memory source/events. Preserve all existing oracle bytes and runtime producers.
Preserve recovery-conformance, builder packet, preparation and admission work.
No store, runner, provider, dispatcher, evaluator implementation, other-system
test, publication, merge or deployment is authorized.

## Bounded acceptance

Freeze separately authored literal observations from static source and contract
reading before implementation or replay. Cover owner cancellation without an
attempt, active cancellation retaining resources and pending cleanup while
retiring the fence, matching retired-token cleanup, unauthorized cancellation,
stale cleanup token and post-cancellation submission. Compare full execution,
exact event and inert effects or exact refusal code/message and unchanged inputs.
Keep remaining commands/states explicitly unmapped; replace the old CANCEL
coverage mutation with a genuinely unmapped command. Add behavioral controls
that bypass source pins and detect cancellation/cleanup/fencing discrepancies.

Run focused and full runtime tests, trace checker, repository/pack/installer
checks, whitespace and committed Project Memory evidence validation with the
existing toolchain and locked dependencies. Obtain independent exact-source
review; contributor separation supplies no human approval or universal proof.
