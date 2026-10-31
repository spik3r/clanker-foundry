---
id: CF-orchestration-astra-review
status: in-progress
owner: Astra PR review
repositories: [spik3r/clanker-foundry]
---

# Repair cancellation and never-started container cleanup

## Implementation claim

Astra owns this bounded review follow-up for PR 9 on
`codex/deterministic-orchestration-plan`, reviewed head `61f7b27757b8` and
main `7a03b095eff5`. The current PR has no assignee or requested reviewer;
its prior implementation evidence frontiers are complete and no active
implementation claim overlaps these files. The repository owner authorized
review, fixing unowned findings, publication and merge after verification.
This local claim does not transfer any runtime namespace or existing task.

Exact scope:
- `runtime/src/adapters.mjs`
- `runtime/test/adapters.test.mjs`
- `runtime/src/isolation.mjs`
- `runtime/test/isolation.test.mjs`
- `tasks/CF-orchestration-astra-review-fixes.md`
- New immutable Project Memory source and event files for this task.

Reject cancellation immediately before an injected text callback is invoked,
including the queued-microtask boundary. Preserve honest cleanup while an
already-running callback has not settled. Recognize an inspected, owned Docker
container that never started as nonrunning and removable; do not treat an
uninspected or running container as stopped. Add deterministic regressions for
both failures and independently recheck the final diff.

No provider calls, namespace activation, operating database changes, deployment,
credential changes, claim migration, history rewriting or branch-protection
changes are included. The lead owns staging, lifecycle evidence and commits.

## Validation and next action

Before publication: run targeted regressions, the complete Node 24 runtime and
Project Memory suites, the pack validator, whitespace checks and the committed
full-diff evidence gate. Run installer and actual isolation checks where the host
supports them; otherwise distinguish host limits from final-head CI results.
Recheck the remote head before pushing and use normal merge controls only after
reviewing the exact final revision.
