---
id: CF-orchestration-store
status: review
owner: Foundry runtime implementation
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Implement transactional coordination and outbox

## Implementation claim

The user authorized orchestrated implementation using Luna or appropriate
subagents. Owner and integration lead: Foundry runtime implementation on branch
`codex/deterministic-orchestration-plan`, base `d879f1f`. This bounded local
claim is not a published shared reservation. The checkout was clean and the
connected open-PR search returned no open PRs before implementation.

Exact write scope:
- `runtime/src/store.mjs`
- `runtime/test/store.test.mjs`
- This task record and new immutable Project Memory source/event files.

Workers own disjoint implementation paths; the lead owns task/evidence records,
dependency setup, integration and local commits. No worker may stage, commit,
publish, deploy, change an existing task owner or rewrite memory history.

## Acceptance and dependencies

Follow the saved deterministic orchestration plan. Use an opt-in single-host
Node/SQLite pilot in this repository, versioned schemas, strict admission,
deterministic transitions, fenced attempts, transactional reservations/receipts,
and restart-safe outbox dispatch. Contracts precede core/store integration;
the pilot depends on their tested APIs. Preserve standalone skills, canonical
Markdown task inputs and offline Project Memory behavior. Live task migration,
automatic writable workers, publication, merge and deployment are excluded.

## Validation and next action

Run focused runtime tests, pack validation, installer regressions when available,
whitespace checks, independent review and committed evidence validation. Do not
claim live provider execution from simulated tests. At claim time the host exposed
Node 26; a Node 24 runtime was later obtained for final validation and completion.
The source was pinned and a genuine start recorded before implementation.

## Local result and observed validation

Added SQLite atomic state/events/receipts/outbox/reservations, immutable artifacts and subjects, recorded-input replay, permanent globally unique attempt identities, pinned dependency outputs and safe unknown-outcome reconciliation. A Sol builder implemented the slice.

The final runtime suite passed 73/73 tests on Node v24.22.2. After making test
temporary paths portable, the affected runner/CLI tests passed 12/12. Pack
validation, installer regression tests and unstaged whitespace checks passed.
Optional `skills-ref` is unavailable. Existing Project Memory regressions ran on
Node 24: 142/143 passed; the unchanged invalid UTF-8 filename fixture failed with
`EPERM` before exercising helper logic, matching the earlier baseline failure.
The full existing suite is therefore not claimed as passing.

Independent Sol review reproduced assignment/dispatch, artifact retention,
revocation, candidate-validation and demo-isolation defects during development.
The builders fixed them and targeted rechecks found no remaining reproduced
P1/P2 findings. These tests include real separate-process SQLite races and local
filesystem fixtures; model responses use injected callbacks, not a live provider.

Implementation editing is finished; this claim reserves no further work.
Next: lead stages the final scope, records completion, commits locally and checks
committed-diff evidence; owner reviews the pilot before any live provider setup
or rollout. No existing task ownership was migrated. No publication, merge,
deployment or unattended writable worker was enabled.
