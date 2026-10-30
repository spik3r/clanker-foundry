---
id: CF-orchestration-contracts
status: in-progress
owner: Foundry runtime implementation
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Define strict runtime contracts and pilot decisions

## Implementation claim

The user authorized orchestrated implementation using Luna or appropriate
subagents. Owner and integration lead: Foundry runtime implementation on branch
`codex/deterministic-orchestration-plan`, base `d879f1f`. This bounded local
claim is not a published shared reservation. The checkout was clean and the
connected open-PR search returned no open PRs before implementation.

Exact write scope:
- `runtime/package.json`
- `runtime/package-lock.json`
- `runtime/contracts/v1.schema.json`
- `runtime/src/contract.mjs`
- `runtime/test/contract.test.mjs`
- `runtime/test/fixtures.mjs`
- `docs/adr/0001-deterministic-runtime.md`
- `.gitignore`
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
claim live provider execution from simulated tests. Record actual Node version;
the current host provides Node 26 rather than the documented Node 24 prerequisite.
Next: register pinned source and record a genuine start before implementation.
