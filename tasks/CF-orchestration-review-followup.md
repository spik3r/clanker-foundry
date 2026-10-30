---
id: CF-orchestration-review
status: review
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
The task was pinned and a genuine start recorded before edits. Checkpoints extend
the exact implementation scope before each additional affected component.

## Local result and validation

A fresh-context Sol reviewer found seven reproducible runtime defects. Corrected
prototype-sensitive validation, fake-worker cleanup ownership, acknowledged-launch
recovery, combined replay size, growing stored-record size, effective check bounds
and generated evidence references. Added regression tests for each. A bounded Luna
builder fixed fake ownership; the lead integrated the remaining corrections.
The independent reviewer rechecked every reproduction and eight recovery scenarios.
See the [review report](../docs/deterministic-orchestration-review.md).

The previously failing invalid UTF-8 filename fixture now constructs a Git tree
without an invalid host filename. Its rejection and deletion assertions remain;
no tests were skipped and production Project Memory code is unchanged.

Observed final checks on Node v24.22.2:

- Combined runtime and lifecycle suites: 227/227 passed, no skips (84 runtime,
  143 Project Memory/workflow tests).
- Pack validation and installer regressions passed.
- Dependency audit: zero known vulnerabilities in the pinned runtime dependencies.
- Unstaged whitespace check passed; staged and committed-evidence checks run in
  the required local completion sequence.
- Optional `skills-ref` is unavailable. Remote CI has not run on this local branch.

Editing is finished; this claim reserves no additional work. Next: commit the
reviewed fixes and immutable completion evidence locally, verify committed
coverage, and deliver the summary. Live provider integration, writable-host
isolation, optional evidence automation and an approved namespace cutover remain
separate decisions; none was silently enabled by this review.
