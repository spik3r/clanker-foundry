---
id: CF-orchestration-rollout
status: in-progress
owner: Foundry managed execution
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Integrate verified hosts, evidence gates and namespace activation

## Implementation claim

The user requested continuing orchestration and finishing the outstanding provider,
writable-host, evidence-bridge and namespace-cutover work. Owner/integration lead:
Foundry managed execution, branch `codex/deterministic-orchestration-plan`, base
`aa89562`. The checkout was clean and the connected open-PR search returned no
open PRs. This is a bounded local claim, not a published shared reservation.

Exact implementation scope:
- `runtime/src/contract.mjs`
- `runtime/contracts/v1.schema.json`
- `runtime/src/state.mjs`
- `runtime/src/runner.mjs`
- `runtime/src/cli.mjs`
- `runtime/test/state.test.mjs`
- `runtime/test/runner.test.mjs`
- `runtime/test/cli.test.mjs`
- `runtime/package.json`
- `runtime/README.md`
- `AGENTS.md`
- `README.md`
- `docs/deterministic-orchestration-plan.md`
- `docs/deterministic-orchestration-review.md`
- `docs/deterministic-orchestration-rollout.md`
- `docs/adr/0002-managed-execution.md`
- `runtime/config/namespace-cf-auto.json`
- `.github/workflows/orchestration.yml`
- This task record and new immutable Project Memory source/event files.

Builders use disjoint paths; the lead owns lifecycle records, integration and
local commits. Preserve existing tasks, active claims and all immutable evidence.
Live verification uses only invented public-safe payloads and owned temporary
workspaces. Credentials remain in the existing provider login or host environment;
never copy them into repository files, logs or model inputs. No publication,
merge, external deployment or silent transfer of existing CF claims is included.

## Acceptance and dependencies

Implement the corresponding remaining architecture boundary with strict validation,
revision/identity checks, deterministic failure/recovery and regression tests.
Reuse the current single-host runtime and offline Project Memory helper.
Managed writes require verified isolation and exact authorized scope; provider
and coordinator credentials stay outside the writable host. Evidence starts
must precede edits; completion requires actual staged fingerprints and checks.
Namespace activation needs a recorded reviewed source and drained prior authority;
new namespaces preserve historical CF records. Missing external requirements fail
closed and are reported rather than counted as successful live validation.

## Validation and next action

Run focused Node 24 tests, relevant full runtime/lifecycle suites, pack and
installer checks, whitespace and committed-diff evidence validation. Independently
review the final implementation. Record actual live-provider and host-isolation
observations separately from fake transport tests. Next: register this pinned
source, record a genuine start, then implement the assigned slice.
