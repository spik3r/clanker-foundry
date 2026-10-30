---
id: CF-orchestration-preparation-cancel
status: in-progress
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

Run Node 24 focused and complete runtime tests, the existing Project Memory suites,
pack/installer/whitespace checks and committed evidence validation. Have the
independent readiness reviewer recheck the race. Next: register the pinned claim,
record a genuine start, implement and validate. Push only after review and required
checks pass; no merge or deployment is authorized.
