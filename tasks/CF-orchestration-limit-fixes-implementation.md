---
id: CF-orchestration-limit-fixes
status: in-progress
owner: Foundry limit fixes
repositories: [spik3r/clanker-foundry]
---

# Align admitted workflow limits with execution

## Implementation claim

The user authorized fixing the two reproduced fresh-review findings on branch
`codex/deterministic-orchestration-plan`, starting at `001f9aa`. The checkout,
staged diff and untracked paths were clean. The connected open-PR search returned
PR 5 about Project Memory CI; this bounded runtime fix does not change that helper
or its CI lane. Existing task records are under review; this follow-up owns only
the following exact paths and its new immutable Project Memory source/events:

- `runtime/src/contract.mjs`
- `runtime/contracts/v1.schema.json`
- `runtime/src/state.mjs`
- `runtime/src/runner.mjs`
- `runtime/src/adapters.mjs`
- `runtime/test/contract.test.mjs`
- `runtime/test/state.test.mjs`
- `runtime/test/runner.test.mjs`
- `runtime/test/adapters.test.mjs`
- `runtime/test/store.test.mjs`
- `runtime/README.md`
- `docs/deterministic-orchestration-rollout.md`
- `tasks/CF-orchestration-limit-fixes-implementation.md`

Support the existing dependency contract through the runner and text adapter.
Resolve and bound immutable inputs before claiming or acknowledging a launch.
Align history capacity with the full admitted stage/retry/rework envelope, retain
existing event replay, and make capacity exhaustion explicit for extra manual
recovery cycles. Preserve old sources, events, claims and unrelated work. No live
provider call, namespace activation, publication, merge or deployment is included.

## Validation and next action

Add regressions for 65 dependencies, maximum subject counts, byte-limit failure
before dispatch, and workflows crossing the old 128-attempt history cap. Verify
existing replay and actor/cleanup protections. Run Node 24 runtime tests, pack
validation, whitespace checks and the committed full-diff evidence gate. Next:
register this reviewed source and record a genuine start before implementation.
