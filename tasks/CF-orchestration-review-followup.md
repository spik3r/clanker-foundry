---
id: CF-orchestration-review
status: in-progress
owner: Foundry review follow-up
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Review the orchestration pilot with fresh context and fix findings

## Implementation claim

The user requested a fresh-context review of the whole implementation, fixes for
actionable findings or outstanding work, and a summary. Owner/integration lead:
Foundry review follow-up on `codex/deterministic-orchestration-plan`, reviewed
implementation base `220afbf`. A clean checkout and the connected open-PR search
showed no competing local changes or open PRs before this claim. This is a local
claim, not a published shared reservation.

Bounded write scope, only where a reproduced review finding requires it:
- `runtime/src/contract.mjs`, `runtime/src/state.mjs`, `runtime/src/store.mjs`
- `runtime/src/runner.mjs`, `runtime/src/adapters.mjs`
- `runtime/src/import-export.mjs`, `runtime/src/cli.mjs`
- `runtime/contracts/v1.schema.json`
- `runtime/package.json`, `runtime/package-lock.json`
- `runtime/test/contract.test.mjs`, `runtime/test/state.test.mjs`
- `runtime/test/store.test.mjs`, `runtime/test/runner.test.mjs`
- `runtime/test/adapters.test.mjs`, `runtime/test/import-export.test.mjs`
- `runtime/test/cli.test.mjs`, `runtime/test/fixtures.mjs`
- `runtime/README.md`, `.github/workflows/orchestration.yml`
- `docs/adr/0001-deterministic-runtime.md`
- `docs/deterministic-orchestration-plan.md`
- `docs/deterministic-orchestration-review.md`
- `tests/project-memory.test.mjs`, only the invalid UTF-8 filename fixture
- This task record and new immutable source/event metadata.

Independent reviewers remain read-only. The lead implements fixes and owns
evidence, integration and local commits. Preserve unrelated work and existing
memory records. Do not add live provider calls, unattended writable workers,
external tracker migration, publication, merge or deployment.

## Acceptance

- A reviewer without conversation history checks source, schemas, tests, docs,
  CI, model/plan coverage and outstanding work against the actual branch.
- Fix reproduced correctness/recovery/validation issues and add relevant regressions.
- Resolve the known platform-dependent fixture without skipping its rejection check.
- Distinguish defects from planned integrations requiring credentials/host decisions.
- Record actual checks, remaining limits and a concrete follow-up summary.

## Validation and next action

Use Node 24 for runtime and lifecycle tests. Run focused regressions, then the
runtime suite and both Project Memory suites, pack validation, installer checks,
whitespace and committed-diff evidence validation. Re-review affected behavior.
Next: pin this task, record a genuine start, collect fresh findings and implement
only supported corrections before final review and local commit.
