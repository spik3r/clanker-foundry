---
id: CF-orchestration-limit-fixes
status: review
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

Implemented shared input limits, preclaim input checks and history capacity for
all admitted automatic retries/reworks. The full maximum workflow completes and
replays 356 attempts. Actual SQLite regressions deliver 65 and 256 dependencies,
retain oversized work without acknowledgement or uncertainty, and let later work
and cleanup proceed with a dispatch limit of one.

Node v24.22.2 passed all 160 runtime tests with no skips. Pack validation and
whitespace checks passed; optional skills-ref is not installed. Independent review
rechecked the fixes and its fairness follow-up, passing 49 runner/store tests and
reporting no remaining material findings. No live provider or Docker rerun was
needed for these coordinator-only changes. Next: local owner review; record staged
completion fingerprints and validate the committed full diff against origin/main.
No publication, namespace activation, merge or deployment is included.
